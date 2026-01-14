# AGENTS.md - CSCAN Codebase Guide

This guide helps agentic coding agents understand and work with the CSCAN codebase.

## Project Overview

CSCAN is an enterprise-grade distributed network asset scanning platform built with Go-Zero (backend) and Vue3 (frontend).
- **Backend**: Go 1.24+ with go-zero framework, MongoDB 6, Redis 7
- **Frontend**: Vue 3.4 + Element Plus + Vite + Pinia
- **Architecture**: API → RPC → Worker nodes with distributed task scheduling

## Build & Test Commands

### Backend (Go)
```bash
# Run services locally
docker-compose -f docker-compose.dev.yaml up -d  # Start dependencies (MongoDB, Redis)
go run api/cscan.go -f api/etc/cscan.yaml         # Start API server
go run rpc/task/task.go -f rpc/task/etc/task.yaml  # Start RPC server
go run cmd/worker/main.go -k <key> -s http://localhost:8888  # Start Worker

# Run tests
go test ./...                           # Run all tests
go test -v ./scheduler/...              # Run tests in specific package with verbose output
go test -run TestPushAndPopTask ./scheduler  # Run single test
go test -count=1 ./scheduler/...        # Run tests once (disable caching)

# Build
go build -o bin/api ./api
go build -o bin/worker ./cmd/worker
```

### Frontend (Vue3)
```bash
cd web
npm install     # Install dependencies
npm run dev     # Start development server
npm run build   # Build for production
npm run preview # Preview production build
```

### Docker
```bash
docker-compose up -d                    # Production deployment
./cscan.sh (Linux/macOS) or .\cscan.bat (Windows)  # Interactive management
```

## Code Style Guidelines

### Go Code

#### Imports
```go
import (
    "context"          // 1. Standard library
    "encoding/json"
    "time"

    "github.com/google/uuid"         // 2. Third-party packages
    "go.mongodb.org/mongo-driver/bson"

    "cscan/model"                   // 3. Local packages (no go.mod prefix)
    "cscan/api/internal/svc"
)
```

#### Naming Conventions
- **Files**: `snake_case.go` (e.g., `tasklogic.go`, `scanner.go`)
- **Packages**: lowercase, single word (e.g., `model`, `scheduler`, `scanner`)
- **Structs**: PascalCase (e.g., `MainTask`, `TaskInfo`, `ScanConfig`)
- **Interfaces**: Simple nouns, may include `-er` suffix (e.g., `Scanner`, `ScannerOptions`)
- **Functions**: PascalCase for exported (e.g., `NewTask`, `GetTask`), camelCase for unexported
- **Constants**: PascalCase for exported (e.g., `TaskStatusCreated`) or UPPER_SNAKE_CASE
- **Variables**: camelCase (e.g., `taskId`, `workspaceId`)
- **JSON tags**: camelCase (e.g., `json:"taskId"`)
- **BSON tags**: snake_case (e.g., `bson:"task_id"`)

#### Struct Definitions
```go
type MainTask struct {
    Id          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    TaskId      string             `bson:"task_id" json:"taskId"`
    Name        string             `bson:"name" json:"name"`
    Status      string             `bson:"status" json:"status"`
    CreateTime  time.Time          `bson:"create_time" json:"createTime"`
    UpdateTime  time.Time          `bson:"update_time" json:"updateTime"`
}
```

#### Error Handling
- Return errors as last return value: `(result, error)`
- Use custom error types from `api/internal/svc/errors.go`:
  ```go
  return ErrValidation.New("Invalid input")
  return ErrNotFound.Newf("%s not found", resource)
  return ErrInternal.New("Database connection failed")
  ```
- Wrap context with errors: `return fmt.Errorf("failed to insert task: %w", err)`
- Use go-zero's xerr package for business errors

#### Logging
```go
import "github.com/zeromicro/go-zero/core/logx"

// Structured logging with context
l.Logger.Infof("Task created: taskId=%s, workspaceId=%s", taskId, workspaceId)
l.Logger.Errorf("Failed to update task: %v", err)
l.Logger.WithContext(ctx).Info("Processing request")
```

#### Go-Zero Patterns
- Logic layer receives context and ServiceContext:
  ```go
  type MainTaskCreateLogic struct {
      logx.Logger
      ctx    context.Context
      svcCtx *svc.ServiceContext
  }

  func NewMainTaskCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MainTaskCreateLogic {
      return &MainTaskCreateLogic{
          Logger: logx.WithContext(ctx),
          ctx:    ctx,
          svcCtx: svcCtx,
      }
  }
  ```
- Use ServiceContext to access models, Redis, and scheduler

#### MongoDB Queries
```go
import "go.mongodb.org/mongo-driver/bson"

filter := bson.M{"status": TaskStatusCreated}
opts := options.Find().SetSort(bson.D{{Key: "create_time", Value: -1}})
cursor, err := coll.Find(ctx, filter, opts)
```

#### Testing
- Use `setupTestScheduler(t)` pattern for tests requiring Redis:
  ```go
  func setupTestScheduler(t *testing.T) (*Scheduler, func()) {
      mr, err := miniredis.Run()  // Use miniredis for in-memory Redis
      if err != nil {
          t.Fatalf("Failed to start miniredis: %v", err)
      }
      rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
      scheduler := NewScheduler(rdb)
      cleanup := func() { rdb.Close(); mr.Close() }
      return scheduler, cleanup
  }
  ```
- Table-driven tests for multiple cases
- Property-based testing with gopter for concurrent/edge cases
- Run single test: `go test -run TestPushAndPopTask -v ./scheduler`

### Frontend (Vue3)

#### Component Structure
```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const tableData = ref([])

onMounted(async () => {
  await fetchData()
})
</script>

<template>
  <el-table :data="tableData">
    <!-- content -->
  </el-table>
</template>

<style lang="scss" scoped>
/* styles */
</style>
```

#### State Management
- Use Pinia for global state
- Store files in `web/src/store/`

#### API Calls
- Use axios with interceptors in `web/src/utils/request.js`
- API endpoints in `web/src/api/`

## Project Structure

```
cscan/
├── api/              # HTTP API layer (go-zero)
│   └── internal/
│       ├── logic/     # Business logic (one file per handler)
│       ├── svc/       # Service context, dependencies
│       └── types/     # Request/response types
├── rpc/              # RPC layer (go-zero)
│   ├── task/         # Task service definition
│   └── pb/           # Protocol buffers
├── worker/           # Worker node implementation
├── scanner/          # Scanning engines (naabu, nuclei, subfinder, etc.)
├── scheduler/        # Task scheduling and queue management
├── model/            # MongoDB models and operations
├── pkg/              # Shared utilities
├── onlineapi/        # FOFA/Hunter/Quake integrations
├── web/              # Vue3 frontend
└── docker/           # Docker configurations
```

## Key Dependencies

- **go-zero**: Microservice framework (rest/rpc)
- **MongoDB**: Primary database with collections per workspace
- **Redis**: Task queue, caching, pub/sub for real-time updates
- **ProjectDiscovery**: naabu, nuclei, subfinder, httpx, dnsx
- **Chromedp**: Web screenshots
- **Element Plus**: Vue3 UI component library
- **Pinia**: State management
- **Vite**: Build tool

## Common Patterns

### Task Creation & Execution
1. Create task in MongoDB with status `CREATED`
2. Split targets into batches using `scheduler.NewTargetSplitter(batchSize)`
3. Push sub-tasks to Redis queue via `scheduler.PushTaskBatch()`
4. Workers pop tasks and execute scans
5. Update task progress in real-time via Redis pub/sub
6. Mark task as `SUCCESS`/`FAILURE` when complete

### Workspace Multi-Tenancy
- Each workspace has separate MongoDB collections (`{workspaceId}_maintask`, `{workspaceId}_assets`)
- Use `svcCtx.GetMainTaskModel(workspaceId)` to get workspace-specific model
- Default workspace: `"default"`

### Atomic Operations
- Use `FindOneAndUpdate` with filters for atomic updates
- Example: `IncrSubTaskDoneAtomic()` prevents race conditions in counter increments

## Development Notes

- Use Chinese comments in business logic files (user-facing)
- Use English comments in interfaces and library code
- Always validate inputs before processing
- Use context with timeout for external API calls
- Log errors with context for debugging
- Clean up Redis data when tasks are deleted
- Use property-based testing for concurrent operations
