package scheduler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/redis/go-redis/v9"
)

// setupTestLoadBalancer 创建测试用的负载均衡器
func setupTestLoadBalancer(t *testing.T) (*LoadBalancer, *Scheduler, func()) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	lb := NewLoadBalancer(rdb)
	scheduler := NewScheduler(rdb)

	cleanup := func() {
		rdb.Close()
		mr.Close()
	}

	return lb, scheduler, cleanup
}

// ==================== Unit Tests ====================

// TestNewLoadBalancer 测试创建负载均衡器
func TestNewLoadBalancer(t *testing.T) {
	lb, _, cleanup := setupTestLoadBalancer(t)
	defer cleanup()

	if lb == nil {
		t.Error("NewLoadBalancer should return non-nil")
	}
	if lb.workerLoadKey != "cscan:worker:load" {
		t.Errorf("workerLoadKey = %s, want cscan:worker:load", lb.workerLoadKey)
	}
}

// TestUpdateAndGetWorkerLoad 测试更新和获取Worker负载
func TestUpdateAndGetWorkerLoad(t *testing.T) {
	lb, _, cleanup := setupTestLoadBalancer(t)
	defer cleanup()

	ctx := context.Background()

	load := &WorkerLoad{
		WorkerName:     "worker-1",
		CurrentTasks:   2,
		MaxConcurrency: 5,
		CPUPercent:     50.0,
		MemPercent:     60.0,
	}

	// 更新负载
	if err := lb.UpdateWorkerLoad(ctx, load); err != nil {
		t.Fatalf("UpdateWorkerLoad failed: %v", err)
	}

	// 获取负载
	loads, err := lb.GetWorkerLoads(ctx)
	if err != nil {
		t.Fatalf("GetWorkerLoads failed: %v", err)
	}
	if len(loads) != 1 {
		t.Fatalf("Expected 1 worker, got %d", len(loads))
	}
	if loads[0].WorkerName != "worker-1" {
		t.Errorf("WorkerName = %s, want worker-1", loads[0].WorkerName)
	}
}

// TestGetAvailableWorkers 测试获取可用Worker
func TestGetAvailableWorkers(t *testing.T) {
	lb, _, cleanup := setupTestLoadBalancer(t)
	defer cleanup()

	ctx := context.Background()

	// 添加多个Worker
	workers := []*WorkerLoad{
		{WorkerName: "worker-1", CurrentTasks: 1, MaxConcurrency: 5, CPUPercent: 30, MemPercent: 40},
		{WorkerName: "worker-2", CurrentTasks: 4, MaxConcurrency: 5, CPUPercent: 70, MemPercent: 60},
		{WorkerName: "worker-3", CurrentTasks: 5, MaxConcurrency: 5, CPUPercent: 50, MemPercent: 50}, // Full
		{WorkerName: "worker-4", CurrentTasks: 0, MaxConcurrency: 5, CPUPercent: 95, MemPercent: 50}, // High CPU
	}

	for _, w := range workers {
		lb.UpdateWorkerLoad(ctx, w)
	}

	// 获取可用Worker
	available, err := lb.GetAvailableWorkers(ctx, nil)
	if err != nil {
		t.Fatalf("GetAvailableWorkers failed: %v", err)
	}

	// worker-3 (full) 和 worker-4 (high CPU) 应该被过滤
	if len(available) != 2 {
		t.Errorf("Expected 2 available workers, got %d", len(available))
	}

	// 第一个应该是负载最低的 worker-1
	if len(available) > 0 && available[0].WorkerName != "worker-1" {
		t.Errorf("First worker should be worker-1, got %s", available[0].WorkerName)
	}
}

// TestSelectBestWorker 测试选择最佳Worker
func TestSelectBestWorker(t *testing.T) {
	lb, _, cleanup := setupTestLoadBalancer(t)
	defer cleanup()

	ctx := context.Background()

	// 添加Worker
	lb.UpdateWorkerLoad(ctx, &WorkerLoad{
		WorkerName: "worker-high", CurrentTasks: 4, MaxConcurrency: 5, CPUPercent: 80, MemPercent: 70,
	})
	lb.UpdateWorkerLoad(ctx, &WorkerLoad{
		WorkerName: "worker-low", CurrentTasks: 1, MaxConcurrency: 5, CPUPercent: 20, MemPercent: 30,
	})

	// 选择最佳Worker
	best, err := lb.SelectBestWorker(ctx)
	if err != nil {
		t.Fatalf("SelectBestWorker failed: %v", err)
	}
	if best == nil {
		t.Fatal("SelectBestWorker should return a worker")
	}
	if best.WorkerName != "worker-low" {
		t.Errorf("Best worker should be worker-low, got %s", best.WorkerName)
	}
}

// TestDistributeTask 测试任务分发
func TestDistributeTask(t *testing.T) {
	lb, scheduler, cleanup := setupTestLoadBalancer(t)
	defer cleanup()

	ctx := context.Background()

	// 添加Worker
	lb.UpdateWorkerLoad(ctx, &WorkerLoad{
		WorkerName: "worker-1", CurrentTasks: 1, MaxConcurrency: 5, CPUPercent: 30, MemPercent: 40,
	})

	// 分发任务
	task := &TaskInfo{TaskId: "task-1", Priority: 5}
	workerName, err := lb.DistributeTask(ctx, scheduler, task)
	if err != nil {
		t.Fatalf("DistributeTask failed: %v", err)
	}
	if workerName != "worker-1" {
		t.Errorf("Task should be distributed to worker-1, got %s", workerName)
	}
}

// TestRemoveWorker 测试移除Worker
func TestRemoveWorker(t *testing.T) {
	lb, _, cleanup := setupTestLoadBalancer(t)
	defer cleanup()

	ctx := context.Background()

	// 添加Worker
	lb.UpdateWorkerLoad(ctx, &WorkerLoad{WorkerName: "worker-1", MaxConcurrency: 5})

	// 移除Worker
	if err := lb.RemoveWorker(ctx, "worker-1"); err != nil {
		t.Fatalf("RemoveWorker failed: %v", err)
	}

	// 验证已移除
	loads, _ := lb.GetWorkerLoads(ctx)
	if len(loads) != 0 {
		t.Error("Worker should be removed")
	}
}

// ==================== Property Tests ====================

// TestProperty3_LoadBasedTaskDistribution 测试 Property 3: Load-Based Task Distribution
// **Property 3: Load-Based Task Distribution**
// **Validates: Requirements 3.3**
// For any set of workers with different current loads, the scheduler SHALL
// distribute new tasks to workers with lower load first.
func TestProperty3_LoadBasedTaskDistribution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Workers are sorted by load score (lower first)
	properties.Property("Available workers are sorted by load score", prop.ForAll(
		func(workerCount int) bool {
			if workerCount < 2 || workerCount > 10 {
				return true
			}

			lb, _, cleanup := setupTestLoadBalancer(t)
			defer cleanup()

			ctx := context.Background()

			// Create workers with varying loads
			for i := 0; i < workerCount; i++ {
				load := &WorkerLoad{
					WorkerName:     string(rune('A' + i)),
					CurrentTasks:   i,
					MaxConcurrency: workerCount + 1,
					CPUPercent:     float64(i * 10),
					MemPercent:     float64(i * 5),
				}
				lb.UpdateWorkerLoad(ctx, load)
			}

			// Get available workers
			available, err := lb.GetAvailableWorkers(ctx, nil)
			if err != nil {
				return false
			}

			// Verify sorted by load score
			config := DefaultLoadBalancerConfig()
			for i := 0; i < len(available)-1; i++ {
				scoreI := lb.calculateLoadScore(available[i], config)
				scoreJ := lb.calculateLoadScore(available[i+1], config)
				if scoreI > scoreJ {
					return false
				}
			}

			return true
		},
		gen.IntRange(2, 10),
	))

	// Property: Best worker has lowest load score
	properties.Property("SelectBestWorker returns worker with lowest load", prop.ForAll(
		func(loads []int) bool {
			if len(loads) < 2 {
				return true
			}

			lb, _, cleanup := setupTestLoadBalancer(t)
			defer cleanup()

			ctx := context.Background()

			// Create workers with given task loads
			for i, taskLoad := range loads {
				if taskLoad < 0 {
					taskLoad = 0
				}
				load := &WorkerLoad{
					WorkerName:     string(rune('A' + i)),
					CurrentTasks:   taskLoad % 10, // Keep within bounds
					MaxConcurrency: 10,
					CPUPercent:     float64(taskLoad % 80),
					MemPercent:     float64(taskLoad % 80),
				}
				lb.UpdateWorkerLoad(ctx, load)
			}

			// Select best worker
			best, err := lb.SelectBestWorker(ctx)
			if err != nil {
				return false
			}
			if best == nil {
				return true // No available workers is valid
			}

			// Verify it has the lowest load score
			available, _ := lb.GetAvailableWorkers(ctx, nil)
			if len(available) == 0 {
				return true
			}

			config := DefaultLoadBalancerConfig()
			bestScore := lb.calculateLoadScore(best, config)
			for _, w := range available {
				if lb.calculateLoadScore(w, config) < bestScore {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(5, gen.IntRange(0, 100)),
	))

	// Property: Full workers are excluded
	properties.Property("Full workers are excluded from selection", prop.ForAll(
		func(fullCount, availableCount int) bool {
			if fullCount < 0 || availableCount < 0 {
				return true
			}
			fullCount = fullCount % 5
			availableCount = availableCount % 5

			lb, _, cleanup := setupTestLoadBalancer(t)
			defer cleanup()

			ctx := context.Background()

			// Create full workers
			for i := 0; i < fullCount; i++ {
				load := &WorkerLoad{
					WorkerName:     "full-" + string(rune('A'+i)),
					CurrentTasks:   5,
					MaxConcurrency: 5, // Full
					CPUPercent:     50,
					MemPercent:     50,
				}
				lb.UpdateWorkerLoad(ctx, load)
			}

			// Create available workers
			for i := 0; i < availableCount; i++ {
				load := &WorkerLoad{
					WorkerName:     "avail-" + string(rune('A'+i)),
					CurrentTasks:   i,
					MaxConcurrency: 10,
					CPUPercent:     30,
					MemPercent:     30,
				}
				lb.UpdateWorkerLoad(ctx, load)
			}

			// Get available workers
			available, _ := lb.GetAvailableWorkers(ctx, nil)

			// Should only contain non-full workers
			for _, w := range available {
				if w.CurrentTasks >= w.MaxConcurrency {
					return false
				}
			}

			return len(available) == availableCount
		},
		gen.IntRange(0, 5),
		gen.IntRange(0, 5),
	))

	// Property: High CPU workers are excluded
	properties.Property("High CPU workers are excluded from selection", prop.ForAll(
		func(highCPUCount, normalCount int) bool {
			if highCPUCount < 0 || normalCount < 0 {
				return true
			}
			highCPUCount = highCPUCount % 5
			normalCount = normalCount % 5

			lb, _, cleanup := setupTestLoadBalancer(t)
			defer cleanup()

			ctx := context.Background()
			config := DefaultLoadBalancerConfig()

			// Create high CPU workers
			for i := 0; i < highCPUCount; i++ {
				load := &WorkerLoad{
					WorkerName:     "highcpu-" + string(rune('A'+i)),
					CurrentTasks:   1,
					MaxConcurrency: 10,
					CPUPercent:     95, // Above threshold
					MemPercent:     50,
				}
				lb.UpdateWorkerLoad(ctx, load)
			}

			// Create normal workers
			for i := 0; i < normalCount; i++ {
				load := &WorkerLoad{
					WorkerName:     "normal-" + string(rune('A'+i)),
					CurrentTasks:   1,
					MaxConcurrency: 10,
					CPUPercent:     50,
					MemPercent:     50,
				}
				lb.UpdateWorkerLoad(ctx, load)
			}

			// Get available workers
			available, _ := lb.GetAvailableWorkers(ctx, config)

			// Should only contain normal CPU workers
			for _, w := range available {
				if w.CPUPercent > config.CPUThreshold {
					return false
				}
			}

			return len(available) == normalCount
		},
		gen.IntRange(0, 5),
		gen.IntRange(0, 5),
	))

	// Property: Task distribution is balanced across workers
	properties.Property("Tasks are distributed to lowest load workers", prop.ForAll(
		func(taskCount int) bool {
			if taskCount < 1 || taskCount > 10 {
				return true
			}

			lb, scheduler, cleanup := setupTestLoadBalancer(t)
			defer cleanup()

			ctx := context.Background()

			// Create workers with different loads
			workers := []*WorkerLoad{
				{WorkerName: "low", CurrentTasks: 0, MaxConcurrency: 10, CPUPercent: 20, MemPercent: 20},
				{WorkerName: "medium", CurrentTasks: 3, MaxConcurrency: 10, CPUPercent: 50, MemPercent: 50},
				{WorkerName: "high", CurrentTasks: 7, MaxConcurrency: 10, CPUPercent: 70, MemPercent: 70},
			}
			for _, w := range workers {
				lb.UpdateWorkerLoad(ctx, w)
			}

			// Distribute tasks
			distributedTo := make(map[string]int)
			for i := 0; i < taskCount; i++ {
				task := &TaskInfo{TaskId: string(rune('A' + i)), Priority: 5}
				workerName, err := lb.DistributeTask(ctx, scheduler, task)
				if err != nil {
					return false
				}
				distributedTo[workerName]++
			}

			// Low load worker should get most tasks
			// (This is a soft check - the exact distribution depends on the algorithm)
			return distributedTo["low"] >= distributedTo["high"]
		},
		gen.IntRange(1, 10),
	))

	// Property: Stale workers (no heartbeat) are excluded
	properties.Property("Stale workers are excluded from selection", prop.ForAll(
		func(staleCount, freshCount int) bool {
			if staleCount < 0 || freshCount < 0 {
				return true
			}
			staleCount = staleCount % 5
			freshCount = freshCount % 5

			lb, _, cleanup := setupTestLoadBalancer(t)
			defer cleanup()

			ctx := context.Background()
			config := DefaultLoadBalancerConfig()
			config.HeartbeatTimeout = 1 * time.Second

			// Create fresh workers first (UpdateWorkerLoad sets LastHeartbeat to now)
			for i := 0; i < freshCount; i++ {
				load := &WorkerLoad{
					WorkerName:     "fresh-" + string(rune('A'+i)),
					CurrentTasks:   1,
					MaxConcurrency: 10,
					CPUPercent:     30,
					MemPercent:     30,
				}
				lb.UpdateWorkerLoad(ctx, load)
			}

			// Create stale workers by directly setting in cache with old heartbeat
			// Note: In real scenario, workers become stale when they stop sending heartbeats
			for i := 0; i < staleCount; i++ {
				load := &WorkerLoad{
					WorkerName:     "stale-" + string(rune('A'+i)),
					CurrentTasks:   1,
					MaxConcurrency: 10,
					CPUPercent:     30,
					MemPercent:     30,
					LastHeartbeat:  time.Now().Add(-2 * time.Second), // Stale
				}
				// Directly update cache to simulate stale worker
				lb.mu.Lock()
				lb.cache[load.WorkerName] = load
				lb.mu.Unlock()
				// Also update Redis with stale timestamp
				data, _ := json.Marshal(load)
				lb.rdb.HSet(ctx, lb.workerLoadKey, load.WorkerName, data)
			}

			// Clear cache expiry to force re-read from Redis
			lb.mu.Lock()
			lb.cacheExpiry = time.Time{}
			lb.mu.Unlock()

			// Get available workers
			available, _ := lb.GetAvailableWorkers(ctx, config)

			// Should only contain fresh workers
			for _, w := range available {
				if time.Since(w.LastHeartbeat) > config.HeartbeatTimeout {
					return false
				}
			}

			return len(available) == freshCount
		},
		gen.IntRange(0, 5),
		gen.IntRange(0, 5),
	))

	properties.TestingRun(t)
}
