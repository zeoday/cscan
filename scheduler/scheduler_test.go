package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/redis/go-redis/v9"
)

// setupTestScheduler 创建测试用的调度器（使用miniredis）
func setupTestScheduler(t *testing.T) (*Scheduler, func()) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	scheduler := NewScheduler(rdb)

	cleanup := func() {
		rdb.Close()
		mr.Close()
	}

	return scheduler, cleanup
}

// ==================== Unit Tests ====================

// TestNewScheduler 测试创建调度器
func TestNewScheduler(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	if scheduler == nil {
		t.Error("NewScheduler should return non-nil scheduler")
	}
	if scheduler.metrics == nil {
		t.Error("Scheduler should have metrics initialized")
	}
	if scheduler.queueKey != "cscan:task:queue" {
		t.Errorf("queueKey = %s, want cscan:task:queue", scheduler.queueKey)
	}
}

// TestPushAndPopTask 测试推送和弹出任务
func TestPushAndPopTask(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	ctx := context.Background()

	task := &TaskInfo{
		TaskId:      "test-task-1",
		MainTaskId:  "main-1",
		WorkspaceId: "ws-1",
		TaskName:    "test",
		Priority:    5,
	}

	// 推送任务
	err := scheduler.PushTask(ctx, task)
	if err != nil {
		t.Fatalf("PushTask failed: %v", err)
	}

	// 弹出任务
	popped, err := scheduler.PopTask(ctx)
	if err != nil {
		t.Fatalf("PopTask failed: %v", err)
	}
	if popped == nil {
		t.Fatal("PopTask should return task")
	}
	if popped.TaskId != task.TaskId {
		t.Errorf("TaskId = %s, want %s", popped.TaskId, task.TaskId)
	}
}

// TestPriorityOrdering 测试优先级排序
func TestPriorityOrdering(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	ctx := context.Background()

	// 推送不同优先级的任务（优先级越高，数值越大）
	tasks := []*TaskInfo{
		{TaskId: "low", Priority: 1},
		{TaskId: "high", Priority: 10},
		{TaskId: "medium", Priority: 5},
	}

	for _, task := range tasks {
		if err := scheduler.PushTask(ctx, task); err != nil {
			t.Fatalf("PushTask failed: %v", err)
		}
		// 添加小延迟确保时间戳不同
		time.Sleep(10 * time.Millisecond)
	}

	// 弹出任务，应该按优先级从高到低
	expectedOrder := []string{"high", "medium", "low"}
	for i, expected := range expectedOrder {
		popped, err := scheduler.PopTask(ctx)
		if err != nil {
			t.Fatalf("PopTask %d failed: %v", i, err)
		}
		if popped == nil {
			t.Fatalf("PopTask %d returned nil", i)
		}
		if popped.TaskId != expected {
			t.Errorf("PopTask %d: TaskId = %s, want %s", i, popped.TaskId, expected)
		}
	}
}

// TestEmptyQueue 测试空队列
func TestEmptyQueue(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	ctx := context.Background()

	task, err := scheduler.PopTask(ctx)
	if err != nil {
		t.Fatalf("PopTask on empty queue failed: %v", err)
	}
	if task != nil {
		t.Error("PopTask on empty queue should return nil")
	}
}

// TestMetricsRecording 测试性能指标记录
func TestMetricsRecording(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	ctx := context.Background()

	// 推送任务
	task := &TaskInfo{TaskId: "metrics-test", Priority: 1}
	scheduler.PushTask(ctx, task)

	// 弹出任务
	scheduler.PopTask(ctx)

	// 检查指标
	stats := scheduler.metrics.GetStats()
	if stats["pushCount"].(int64) != 1 {
		t.Errorf("pushCount = %d, want 1", stats["pushCount"])
	}
	if stats["popCount"].(int64) != 1 {
		t.Errorf("popCount = %d, want 1", stats["popCount"])
	}
}

// TestWorkerSpecificQueue 测试Worker专属队列
func TestWorkerSpecificQueue(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	ctx := context.Background()

	// 推送到特定Worker的队列
	task := &TaskInfo{
		TaskId:   "worker-task",
		Priority: 5,
		Workers:  []string{"worker-1"},
	}
	if err := scheduler.PushTask(ctx, task); err != nil {
		t.Fatalf("PushTask failed: %v", err)
	}

	// 从公共队列弹出应该为空
	publicTask, _ := scheduler.PopTask(ctx)
	if publicTask != nil {
		t.Error("Public queue should be empty")
	}

	// 从Worker专属队列弹出
	workerTask, err := scheduler.PopTaskForWorker(ctx, "worker-1")
	if err != nil {
		t.Fatalf("PopTaskForWorker failed: %v", err)
	}
	if workerTask == nil {
		t.Fatal("Worker queue should have task")
	}
	if workerTask.TaskId != task.TaskId {
		t.Errorf("TaskId = %s, want %s", workerTask.TaskId, task.TaskId)
	}
}

// TestBatchPush 测试批量推送
func TestBatchPush(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	ctx := context.Background()

	tasks := []*TaskInfo{
		{TaskId: "batch-1", Priority: 1},
		{TaskId: "batch-2", Priority: 2},
		{TaskId: "batch-3", Priority: 3},
	}

	if err := scheduler.PushTaskBatch(ctx, tasks); err != nil {
		t.Fatalf("PushTaskBatch failed: %v", err)
	}

	// 验证队列长度
	length, err := scheduler.GetQueueLength(ctx)
	if err != nil {
		t.Fatalf("GetQueueLength failed: %v", err)
	}
	if length != 3 {
		t.Errorf("Queue length = %d, want 3", length)
	}
}

// TestPeekTask 测试查看任务（不移除）
func TestPeekTask(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	ctx := context.Background()

	task := &TaskInfo{TaskId: "peek-test", Priority: 5}
	scheduler.PushTask(ctx, task)

	// Peek不应该移除任务
	peeked, err := scheduler.PeekTask(ctx)
	if err != nil {
		t.Fatalf("PeekTask failed: %v", err)
	}
	if peeked == nil || peeked.TaskId != task.TaskId {
		t.Error("PeekTask should return the task")
	}

	// 队列长度应该仍然是1
	length, _ := scheduler.GetQueueLength(ctx)
	if length != 1 {
		t.Errorf("Queue length after peek = %d, want 1", length)
	}
}

// ==================== Property Tests ====================

// TestProperty2_PriorityQueueOrdering 测试 Property 2: Priority Queue Ordering
// **Property 2: Priority Queue Ordering**
// **Validates: Requirements 3.1, 3.2**
// For any set of tasks with different priorities, the scheduler SHALL always
// return the highest priority task first when a worker requests a task.
func TestProperty2_PriorityQueueOrdering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Tasks are always dequeued in priority order (highest first)
	properties.Property("Higher priority tasks are dequeued first", prop.ForAll(
		func(priorities []int) bool {
			if len(priorities) < 2 {
				return true // Skip trivial cases
			}

			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx := context.Background()

			// Push tasks with given priorities
			for i, priority := range priorities {
				task := &TaskInfo{
					TaskId:   string(rune('A' + i)),
					Priority: priority,
				}
				if err := scheduler.PushTask(ctx, task); err != nil {
					return false
				}
				// Small delay to ensure different timestamps
				time.Sleep(time.Millisecond)
			}

			// Pop all tasks and verify order
			var prevPriority *int
			for range priorities {
				task, err := scheduler.PopTask(ctx)
				if err != nil || task == nil {
					return false
				}

				// Each task should have priority <= previous (descending order)
				if prevPriority != nil && task.Priority > *prevPriority {
					return false
				}
				prevPriority = &task.Priority
			}

			return true
		},
		gen.SliceOfN(10, gen.IntRange(0, 100)),
	))

	// Property: Same priority tasks maintain FIFO order
	properties.Property("Same priority tasks maintain FIFO order", prop.ForAll(
		func(count int) bool {
			if count < 2 || count > 20 {
				return true
			}

			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx := context.Background()

			// Push tasks with same priority
			samePriority := 5
			for i := 0; i < count; i++ {
				task := &TaskInfo{
					TaskId:   string(rune('A' + i)),
					Priority: samePriority,
				}
				if err := scheduler.PushTask(ctx, task); err != nil {
					return false
				}
				time.Sleep(time.Millisecond)
			}

			// Pop all tasks - should be in FIFO order
			for i := 0; i < count; i++ {
				task, err := scheduler.PopTask(ctx)
				if err != nil || task == nil {
					return false
				}
				expected := string(rune('A' + i))
				if task.TaskId != expected {
					return false
				}
			}

			return true
		},
		gen.IntRange(2, 20),
	))

	// Property: High priority task added later still comes first
	properties.Property("High priority task added later comes first", prop.ForAll(
		func(lowPriority, highPriority int) bool {
			if highPriority <= lowPriority {
				return true // Skip invalid cases
			}

			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx := context.Background()

			// Push low priority task first
			lowTask := &TaskInfo{TaskId: "low", Priority: lowPriority}
			scheduler.PushTask(ctx, lowTask)
			time.Sleep(10 * time.Millisecond)

			// Push high priority task later
			highTask := &TaskInfo{TaskId: "high", Priority: highPriority}
			scheduler.PushTask(ctx, highTask)

			// High priority task should come first
			first, _ := scheduler.PopTask(ctx)
			if first == nil || first.TaskId != "high" {
				return false
			}

			second, _ := scheduler.PopTask(ctx)
			if second == nil || second.TaskId != "low" {
				return false
			}

			return true
		},
		gen.IntRange(0, 50),
		gen.IntRange(51, 100),
	))

	// Property: Batch push maintains priority order
	properties.Property("Batch push maintains priority order", prop.ForAll(
		func(priorities []int) bool {
			if len(priorities) < 2 {
				return true
			}

			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx := context.Background()

			// Create tasks with given priorities
			tasks := make([]*TaskInfo, len(priorities))
			for i, p := range priorities {
				tasks[i] = &TaskInfo{
					TaskId:   string(rune('A' + i)),
					Priority: p,
				}
			}

			// Batch push
			if err := scheduler.PushTaskBatch(ctx, tasks); err != nil {
				return false
			}

			// Pop and verify priority order
			var prevPriority *int
			for range priorities {
				task, err := scheduler.PopTask(ctx)
				if err != nil || task == nil {
					return false
				}
				if prevPriority != nil && task.Priority > *prevPriority {
					return false
				}
				prevPriority = &task.Priority
			}

			return true
		},
		gen.SliceOfN(10, gen.IntRange(0, 100)),
	))

	// Property: PopTask returns within reasonable time (< 100ms as per requirement)
	properties.Property("PopTask returns within 100ms", prop.ForAll(
		func(taskCount int) bool {
			if taskCount < 1 || taskCount > 50 {
				return true
			}

			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx := context.Background()

			// Push some tasks
			for i := 0; i < taskCount; i++ {
				task := &TaskInfo{TaskId: string(rune('A' + i)), Priority: i}
				scheduler.PushTask(ctx, task)
			}

			// Measure pop time
			start := time.Now()
			_, err := scheduler.PopTask(ctx)
			elapsed := time.Since(start)

			if err != nil {
				return false
			}

			// Should complete within 100ms
			return elapsed < 100*time.Millisecond
		},
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t)
}


// ==================== Cancellation Tests ====================

// TestCancelSignalBasic 测试基本的取消信号功能
func TestCancelSignalBasic(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	ctx := context.Background()

	// 发送取消信号
	taskId := "cancel-test-1"
	action := "STOP"
	if err := scheduler.SendCancelSignal(ctx, taskId, action); err != nil {
		t.Fatalf("SendCancelSignal failed: %v", err)
	}

	// 检查取消信号
	signal, err := scheduler.CheckCancelSignal(ctx, taskId)
	if err != nil {
		t.Fatalf("CheckCancelSignal failed: %v", err)
	}
	if signal == nil {
		t.Fatal("CheckCancelSignal should return signal")
	}
	if signal.TaskId != taskId {
		t.Errorf("TaskId = %s, want %s", signal.TaskId, taskId)
	}
	if signal.Action != action {
		t.Errorf("Action = %s, want %s", signal.Action, action)
	}
}

// TestClearCancelSignal 测试清除取消信号
func TestClearCancelSignal(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	ctx := context.Background()

	taskId := "clear-test-1"
	scheduler.SendCancelSignal(ctx, taskId, "STOP")

	// 清除信号
	if err := scheduler.ClearCancelSignal(ctx, taskId); err != nil {
		t.Fatalf("ClearCancelSignal failed: %v", err)
	}

	// 验证已清除
	signal, _ := scheduler.CheckCancelSignal(ctx, taskId)
	if signal != nil {
		t.Error("Signal should be cleared")
	}
}

// TestPublishCancelSignal 测试发布取消信号（Pub/Sub）
func TestPublishCancelSignal(t *testing.T) {
	scheduler, cleanup := setupTestScheduler(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 订阅取消信号
	signalChan := scheduler.SubscribeCancelSignals(ctx)

	// 等待订阅建立
	time.Sleep(100 * time.Millisecond)

	// 发布取消信号
	taskId := "pubsub-test-1"
	action := "STOP"
	if err := scheduler.PublishCancelSignal(ctx, taskId, action); err != nil {
		t.Fatalf("PublishCancelSignal failed: %v", err)
	}

	// 等待接收信号
	select {
	case signal := <-signalChan:
		if signal.TaskId != taskId {
			t.Errorf("TaskId = %s, want %s", signal.TaskId, taskId)
		}
		if signal.Action != action {
			t.Errorf("Action = %s, want %s", signal.Action, action)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for cancel signal")
	}
}

// TestProperty5_CancellationPropagation 测试 Property 5: Cancellation Propagation
// **Property 5: Cancellation Propagation**
// **Validates: Requirements 3.5**
// For any cancelled task, the cancellation signal SHALL reach the executing
// worker within 1 second.
func TestProperty5_CancellationPropagation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Cancel signal is stored and retrievable
	properties.Property("Cancel signal is stored and retrievable", prop.ForAll(
		func(taskId string, action string) bool {
			if taskId == "" {
				return true
			}

			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx := context.Background()

			// Send cancel signal
			if err := scheduler.SendCancelSignal(ctx, taskId, action); err != nil {
				return false
			}

			// Check signal is retrievable
			signal, err := scheduler.CheckCancelSignal(ctx, taskId)
			if err != nil {
				return false
			}
			if signal == nil {
				return false
			}

			return signal.TaskId == taskId && signal.Action == action
		},
		gen.AlphaString(),
		gen.OneConstOf("STOP", "PAUSE"),
	))

	// Property: Cancel signal delivery via Pub/Sub is within 1 second
	properties.Property("Cancel signal delivery via Pub/Sub is within 1 second", prop.ForAll(
		func(taskIndex int) bool {
			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Subscribe to cancel signals
			signalChan := scheduler.SubscribeCancelSignals(ctx)

			// Wait for subscription to be established
			time.Sleep(50 * time.Millisecond)

			taskId := "task-" + string(rune('A'+taskIndex%26))
			action := "STOP"

			// Record send time
			sendTime := time.Now()

			// Publish cancel signal
			if err := scheduler.PublishCancelSignal(ctx, taskId, action); err != nil {
				return false
			}

			// Wait for signal with 1 second timeout
			select {
			case signal := <-signalChan:
				receiveTime := time.Now()
				latency := receiveTime.Sub(sendTime)

				// Verify signal content
				if signal.TaskId != taskId || signal.Action != action {
					return false
				}

				// Verify latency is within 1 second
				return latency < 1*time.Second

			case <-time.After(1 * time.Second):
				// Timeout - signal not received within 1 second
				return false
			}
		},
		gen.IntRange(0, 100),
	))

	// Property: Multiple cancel signals are all delivered
	properties.Property("Multiple cancel signals are all delivered", prop.ForAll(
		func(signalCount int) bool {
			if signalCount < 1 || signalCount > 10 {
				return true
			}

			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			// Subscribe to cancel signals
			signalChan := scheduler.SubscribeCancelSignals(ctx)

			// Wait for subscription
			time.Sleep(50 * time.Millisecond)

			// Send multiple cancel signals
			taskIds := make(map[string]bool)
			for i := 0; i < signalCount; i++ {
				taskId := "multi-" + string(rune('A'+i))
				taskIds[taskId] = false
				scheduler.PublishCancelSignal(ctx, taskId, "STOP")
				time.Sleep(10 * time.Millisecond) // Small delay between signals
			}

			// Receive all signals within timeout
			timeout := time.After(2 * time.Second)
			received := 0
			for received < signalCount {
				select {
				case signal := <-signalChan:
					if _, exists := taskIds[signal.TaskId]; exists {
						taskIds[signal.TaskId] = true
						received++
					}
				case <-timeout:
					// Check how many were received
					return received == signalCount
				}
			}

			// Verify all signals were received
			for _, wasReceived := range taskIds {
				if !wasReceived {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	// Property: Cancel signal can be cleared
	properties.Property("Cancel signal can be cleared", prop.ForAll(
		func(taskId string) bool {
			if taskId == "" {
				return true
			}

			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx := context.Background()

			// Send and then clear signal
			scheduler.SendCancelSignal(ctx, taskId, "STOP")
			scheduler.ClearCancelSignal(ctx, taskId)

			// Signal should be gone
			signal, _ := scheduler.CheckCancelSignal(ctx, taskId)
			return signal == nil
		},
		gen.AlphaString(),
	))

	// Property: Cancel signal contains valid timestamp
	properties.Property("Cancel signal contains valid timestamp", prop.ForAll(
		func(taskId string) bool {
			if taskId == "" {
				return true
			}

			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx := context.Background()

			beforeSend := time.Now()
			scheduler.SendCancelSignal(ctx, taskId, "STOP")
			afterSend := time.Now()

			signal, err := scheduler.CheckCancelSignal(ctx, taskId)
			if err != nil || signal == nil {
				return false
			}

			// Timestamp should be between before and after send
			return !signal.Timestamp.Before(beforeSend) && !signal.Timestamp.After(afterSend)
		},
		gen.AlphaString(),
	))

	// Property: Different actions are preserved correctly
	properties.Property("Different actions are preserved correctly", prop.ForAll(
		func(action string) bool {
			scheduler, cleanup := setupTestScheduler(t)
			defer cleanup()

			ctx := context.Background()
			taskId := "action-test"

			scheduler.SendCancelSignal(ctx, taskId, action)

			signal, err := scheduler.CheckCancelSignal(ctx, taskId)
			if err != nil || signal == nil {
				return false
			}

			return signal.Action == action
		},
		gen.OneConstOf("STOP", "PAUSE", "RESUME", "CANCEL"),
	))

	properties.TestingRun(t)
}
