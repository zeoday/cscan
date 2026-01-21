package worker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cscan/api/internal/svc"
	"cscan/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestRescanFlowIntegration tests the complete rescan flow:
// 1. First scan creates initial results
// 2. Second scan archives old results and saves new ones
// 3. Historical data is queryable after rescan
//
// **Validates: Requirements 3.1, 3.4, 3.5**
func TestRescanFlowIntegration(t *testing.T) {
	// Setup MongoDB connection
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err, "Failed to connect to MongoDB")
	defer client.Disconnect(ctx)

	db := client.Database("cscan_test")
	workspaceId := "test_workspace_rescan"

	// Clean up before test
	cleanupRescanTest(t, ctx, db, workspaceId)
	defer cleanupRescanTest(t, ctx, db, workspaceId)

	// Create service context
	svcCtx := &svc.ServiceContext{
		MongoDB: db,
	}

	// Test data
	testAuthority := "example.com:8080"
	testHost := "example.com"
	testPort := 8080
	mainTaskId := "test-task-rescan-" + time.Now().Format("20060102150405")

	// ==================== FIRST SCAN ====================
	t.Run("FirstScan_CreatesInitialResults", func(t *testing.T) {
		// Prepare first scan request
		firstScanReq := WorkerDirScanResultReq{
			WorkspaceId: workspaceId,
			MainTaskId:  mainTaskId,
			Results: []WorkerDirScanResultDocument{
				{
					Authority:     testAuthority,
					Host:          testHost,
					Port:          testPort,
					URL:           "http://example.com:8080/admin",
					Path:          "/admin",
					StatusCode:    200,
					ContentLength: 1024,
					Title:         "Admin Panel",
				},
				{
					Authority:     testAuthority,
					Host:          testHost,
					Port:          testPort,
					URL:           "http://example.com:8080/login",
					Path:          "/login",
					StatusCode:    200,
					ContentLength: 512,
					Title:         "Login Page",
				},
			},
		}

		// Call the handler
		reqBody, _ := json.Marshal(firstScanReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/worker/task/dirscan", strings.NewReader(string(reqBody)))
		w := httptest.NewRecorder()

		handler := WorkerDirScanResultHandler(svcCtx)
		handler(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code, "First scan should succeed")

		var resp WorkerDirScanResultResp
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err, "Failed to parse response")
		assert.True(t, resp.Success, "First scan should be successful")
		assert.Equal(t, int64(2), resp.Total, "Should save 2 directory scan results")

		// Verify results are in database
		dirScanModel := model.NewDirScanResultModel(db)
		filter := bson.M{
			"workspace_id": workspaceId,
			"host":         testHost,
			"port":         testPort,
		}
		count, err := dirScanModel.CountByFilter(ctx, filter)
		require.NoError(t, err, "Failed to count directory scan results")
		assert.Equal(t, int64(2), count, "Should have 2 directory scan results in database")

		// Verify scan_time is set
		results, err := dirScanModel.FindByFilter(ctx, filter, 1, 10)
		require.NoError(t, err, "Failed to fetch directory scan results")
		require.Len(t, results, 2, "Should have 2 results")
		for _, result := range results {
			assert.False(t, result.ScanTime.IsZero(), "ScanTime should be set")
			assert.Equal(t, int64(1), result.Version, "Version should be 1 for first scan")
		}

		// Verify no historical data exists yet
		historyModel := model.NewScanResultHistoryModel(db, workspaceId)
		historyCount, err := historyModel.Count(ctx, bson.M{
			"host": testHost,
			"port": testPort,
		})
		require.NoError(t, err, "Failed to count history")
		assert.Equal(t, int64(0), historyCount, "Should have no historical data after first scan")
	})

	// Wait a bit to ensure different timestamps
	time.Sleep(100 * time.Millisecond)

	// ==================== SECOND SCAN (RESCAN) ====================
	t.Run("SecondScan_ArchivesOldAndSavesNew", func(t *testing.T) {
		// Prepare second scan request with different results
		secondScanReq := WorkerDirScanResultReq{
			WorkspaceId: workspaceId,
			MainTaskId:  mainTaskId,
			Results: []WorkerDirScanResultDocument{
				{
					Authority:     testAuthority,
					Host:          testHost,
					Port:          testPort,
					URL:           "http://example.com:8080/admin",
					Path:          "/admin",
					StatusCode:    403, // Changed from 200
					ContentLength: 1024,
					Title:         "Admin Panel - Forbidden",
				},
				{
					Authority:     testAuthority,
					Host:          testHost,
					Port:          testPort,
					URL:           "http://example.com:8080/api",
					Path:          "/api",
					StatusCode:    200,
					ContentLength: 256,
					Title:         "API Endpoint", // New path
				},
			},
		}

		// Call the handler
		reqBody, _ := json.Marshal(secondScanReq)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/worker/task/dirscan", strings.NewReader(string(reqBody)))
		w := httptest.NewRecorder()

		handler := WorkerDirScanResultHandler(svcCtx)
		handler(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code, "Second scan should succeed")

		var resp WorkerDirScanResultResp
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err, "Failed to parse response")
		assert.True(t, resp.Success, "Second scan should be successful")
		assert.Equal(t, int64(2), resp.Total, "Should save 2 directory scan results")

		// Verify current results are updated
		dirScanModel := model.NewDirScanResultModel(db)
		filter := bson.M{
			"workspace_id": workspaceId,
			"host":         testHost,
			"port":         testPort,
		}
		count, err := dirScanModel.CountByFilter(ctx, filter)
		require.NoError(t, err, "Failed to count directory scan results")
		assert.Equal(t, int64(2), count, "Should still have 2 directory scan results in database")

		// Verify new results have updated data
		results, err := dirScanModel.FindByFilter(ctx, filter, 1, 10)
		require.NoError(t, err, "Failed to fetch directory scan results")
		require.Len(t, results, 2, "Should have 2 results")

		// Check that we have the new paths
		paths := make(map[string]bool)
		for _, result := range results {
			paths[result.Path] = true
			assert.False(t, result.ScanTime.IsZero(), "ScanTime should be set")
			assert.Equal(t, int64(1), result.Version, "Version should be 1 for new scan")
		}
		assert.True(t, paths["/admin"], "Should have /admin path")
		assert.True(t, paths["/api"], "Should have /api path (new)")
		assert.False(t, paths["/login"], "Should not have /login path (removed)")

		// Verify historical data was created
		historyModel := model.NewScanResultHistoryModel(db, workspaceId)
		historyCount, err := historyModel.Count(ctx, bson.M{
			"host": testHost,
			"port": testPort,
		})
		require.NoError(t, err, "Failed to count history")
		assert.Equal(t, int64(1), historyCount, "Should have 1 historical record after rescan")
	})

	// ==================== QUERY HISTORICAL DATA ====================
	t.Run("HistoricalData_IsQueryable", func(t *testing.T) {
		// Query historical data using HistoryService
		historyService := svc.NewHistoryService(db)
		historyReq := &svc.GetResultHistoryReq{
			WorkspaceId: workspaceId,
			Authority:   testAuthority,
			Host:        testHost,
			Port:        testPort,
			StartTime:   time.Now().Add(-1 * time.Hour),
			EndTime:     time.Now().Add(1 * time.Hour),
		}

		historyResp, err := historyService.GetResultHistory(ctx, historyReq)
		require.NoError(t, err, "Failed to get result history")
		require.NotNil(t, historyResp, "History response should not be nil")
		require.Len(t, historyResp.Versions, 1, "Should have 1 historical version")

		// Verify historical version contains the old data
		version := historyResp.Versions[0]
		assert.False(t, version.ScanTimestamp.IsZero(), "Historical version should have scan timestamp")
		assert.Equal(t, int64(2), version.DirScanCount, "Historical version should have 2 dir scan results")

		// Query the actual historical scan result
		historyModel := model.NewScanResultHistoryModel(db, workspaceId)
		historyResults, err := historyModel.FindByAuthority(ctx, workspaceId, testAuthority, testHost, testPort, 10)
		require.NoError(t, err, "Failed to fetch historical results")
		require.Len(t, historyResults, 1, "Should have 1 historical result")

		historicalRecord := historyResults[0]
		assert.Equal(t, testHost, historicalRecord.Host, "Historical record should have correct host")
		assert.Equal(t, testPort, historicalRecord.Port, "Historical record should have correct port")
		assert.Len(t, historicalRecord.DirScanResults, 2, "Historical record should contain 2 dir scan results")

		// Verify the old paths are in historical data
		historicalPaths := make(map[string]bool)
		for _, result := range historicalRecord.DirScanResults {
			historicalPaths[result.Path] = true
		}
		assert.True(t, historicalPaths["/admin"], "Historical data should have /admin path")
		assert.True(t, historicalPaths["/login"], "Historical data should have /login path")
		assert.False(t, historicalPaths["/api"], "Historical data should not have /api path (added in second scan)")
	})

	// ==================== VERIFY MOST RECENT RESULTS ====================
	t.Run("MostRecentResults_AreReturnedByDefault", func(t *testing.T) {
		// Use ScanResultService to get current results
		scanResultService := svc.NewScanResultService(db)
		dirScanReq := &svc.GetDirScanResultsReq{
			WorkspaceId: workspaceId,
			Authority:   testAuthority,
			Host:        testHost,
			Port:        testPort,
			Limit:       100,
			Offset:      0,
		}

		dirScanResp, err := scanResultService.GetDirScanResults(ctx, dirScanReq)
		require.NoError(t, err, "Failed to get directory scan results")
		require.NotNil(t, dirScanResp, "Directory scan response should not be nil")
		assert.Equal(t, int64(2), dirScanResp.Total, "Should have 2 current directory scan results")
		assert.Len(t, dirScanResp.Results, 2, "Should return 2 results")

		// Verify these are the most recent results (from second scan)
		paths := make(map[string]int)
		for _, result := range dirScanResp.Results {
			paths[result.Path] = result.StatusCode
		}
		assert.Equal(t, 403, paths["/admin"], "Should have updated status code for /admin")
		assert.Equal(t, 200, paths["/api"], "Should have new /api path")
		_, hasLogin := paths["/login"]
		assert.False(t, hasLogin, "Should not have /login path (removed in second scan)")
	})
}

// cleanupRescanTest cleans up test data
func cleanupRescanTest(t *testing.T, ctx context.Context, db *mongo.Database, workspaceId string) {
	// Clean up directory scan results
	dirScanModel := model.NewDirScanResultModel(db)
	_, err := dirScanModel.DeleteByFilter(ctx, bson.M{"workspace_id": workspaceId})
	if err != nil {
		t.Logf("Warning: Failed to clean up directory scan results: %v", err)
	}

	// Clean up historical data
	historyModel := model.NewScanResultHistoryModel(db, workspaceId)
	_, err = historyModel.Clear(ctx, workspaceId)
	if err != nil {
		t.Logf("Warning: Failed to clean up historical data: %v", err)
	}

	// Clean up assets
	assetModel := model.NewAssetModel(db, workspaceId)
	_, err = assetModel.DeleteByFilter(ctx, bson.M{})
	if err != nil {
		t.Logf("Warning: Failed to clean up assets: %v", err)
	}
}
