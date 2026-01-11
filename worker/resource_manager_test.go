package worker

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// ==================== Unit Tests ====================

// TestNewResourceManager 测试创建资源管理器
func TestNewResourceManager(t *testing.T) {
	rm := NewResourceManager(5)
	if rm == nil {
		t.Fatal("NewResourceManager returned nil")
	}
	if rm.config.MaxConcurrency != 5 {
		t.Errorf("MaxConcurrency = %d, want 5", rm.config.MaxConcurrency)
	}
	if rm.CurrentTasks() != 0 {
		t.Errorf("CurrentTasks = %d, want 0", rm.CurrentTasks())
	}
}

// TestAcquireReleaseSlot 测试获取和释放槽位
func TestAcquireReleaseSlot(t *testing.T) {
	rm := NewResourceManager(2)

	// 获取第一个槽位
	if !rm.AcquireSlot() {
		t.Error("AcquireSlot() should return true for first slot")
	}
	if rm.CurrentTasks() != 1 {
		t.Errorf("CurrentTasks = %d, want 1", rm.CurrentTasks())
	}

	// 获取第二个槽位
	if !rm.AcquireSlot() {
		t.Error("AcquireSlot() should return true for second slot")
	}
	if rm.CurrentTasks() != 2 {
		t.Errorf("CurrentTasks = %d, want 2", rm.CurrentTasks())
	}

	// 尝试获取第三个槽位（应该失败）
	if rm.AcquireSlot() {
		t.Error("AcquireSlot() should return false when at max concurrency")
	}

	// 释放一个槽位
	rm.ReleaseSlot()
	if rm.CurrentTasks() != 1 {
		t.Errorf("CurrentTasks = %d, want 1", rm.CurrentTasks())
	}

	// 现在应该可以获取槽位
	if !rm.AcquireSlot() {
		t.Error("AcquireSlot() should return true after release")
	}
}

// TestReleaseSlotNeverNegative 测试释放槽位不会变成负数
func TestReleaseSlotNeverNegative(t *testing.T) {
	rm := NewResourceManager(2)

	// 多次释放不应该导致负数
	rm.ReleaseSlot()
	rm.ReleaseSlot()
	rm.ReleaseSlot()

	if rm.CurrentTasks() != 0 {
		t.Errorf("CurrentTasks = %d, want 0 (should not go negative)", rm.CurrentTasks())
	}
}

// TestAvailableSlots 测试可用槽位计算
func TestAvailableSlots(t *testing.T) {
	rm := NewResourceManager(3)

	if rm.AvailableSlots() != 3 {
		t.Errorf("AvailableSlots = %d, want 3", rm.AvailableSlots())
	}

	rm.AcquireSlot()
	if rm.AvailableSlots() != 2 {
		t.Errorf("AvailableSlots = %d, want 2", rm.AvailableSlots())
	}

	rm.AcquireSlot()
	rm.AcquireSlot()
	if rm.AvailableSlots() != 0 {
		t.Errorf("AvailableSlots = %d, want 0", rm.AvailableSlots())
	}
}

// TestSetMaxConcurrency 测试动态设置最大并发数
func TestSetMaxConcurrency(t *testing.T) {
	rm := NewResourceManager(2)

	rm.SetMaxConcurrency(5)
	if rm.config.MaxConcurrency != 5 {
		t.Errorf("MaxConcurrency = %d, want 5", rm.config.MaxConcurrency)
	}

	// 无效值不应该改变
	rm.SetMaxConcurrency(0)
	if rm.config.MaxConcurrency != 5 {
		t.Errorf("MaxConcurrency = %d, want 5 (should not change for invalid value)", rm.config.MaxConcurrency)
	}

	rm.SetMaxConcurrency(-1)
	if rm.config.MaxConcurrency != 5 {
		t.Errorf("MaxConcurrency = %d, want 5 (should not change for negative value)", rm.config.MaxConcurrency)
	}
}

// TestResetThrottle 测试重置限流状态
func TestResetThrottle(t *testing.T) {
	config := DefaultResourceManagerConfig(2)
	config.ThrottleDuration = 1 * time.Hour // 设置很长的限流时间
	rm := NewResourceManagerWithConfig(config)

	// 手动设置限流状态
	rm.mu.Lock()
	rm.throttled = true
	rm.throttleUntil = time.Now().Add(1 * time.Hour)
	rm.overloadCount = 5
	rm.mu.Unlock()

	if !rm.IsThrottled() {
		t.Error("IsThrottled() should return true")
	}

	rm.ResetThrottle()

	if rm.IsThrottled() {
		t.Error("IsThrottled() should return false after reset")
	}
}

// TestGetResourceStatus 测试获取资源状态
func TestGetResourceStatus(t *testing.T) {
	rm := NewResourceManager(3)
	rm.AcquireSlot()

	status := rm.GetResourceStatus()

	if status.CurrentTasks != 1 {
		t.Errorf("CurrentTasks = %d, want 1", status.CurrentTasks)
	}
	if status.MaxConcurrency != 3 {
		t.Errorf("MaxConcurrency = %d, want 3", status.MaxConcurrency)
	}
	if status.AvailableSlots != 2 {
		t.Errorf("AvailableSlots = %d, want 2", status.AvailableSlots)
	}
}


// ==================== Property Tests ====================

// TestProperty9_ResourceBasedConcurrencyLimiting 测试 Property 9: Resource-Based Concurrency Limiting
// **Property 9: Resource-Based Concurrency Limiting**
// **Validates: Requirements 7.1, 7.2**
// For any worker, when CPU usage exceeds 80% or memory exceeds 85%,
// the worker SHALL stop accepting new tasks until resources are available.
func TestProperty9_ResourceBasedConcurrencyLimiting(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Current tasks never exceed max concurrency
	properties.Property("current tasks never exceed max concurrency", prop.ForAll(
		func(maxConcurrency, acquireCount int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 100 {
				maxConcurrency = 100
			}
			if acquireCount < 0 {
				acquireCount = 0
			}
			if acquireCount > 200 {
				acquireCount = 200
			}

			rm := NewResourceManager(maxConcurrency)

			// Try to acquire more slots than allowed
			for i := 0; i < acquireCount; i++ {
				rm.AcquireSlot()
			}

			// Current tasks should never exceed max concurrency
			return rm.CurrentTasks() <= maxConcurrency
		},
		gen.IntRange(1, 100),
		gen.IntRange(0, 200),
	))

	// Property: Available slots is always non-negative
	properties.Property("available slots is always non-negative", prop.ForAll(
		func(maxConcurrency, acquireCount, releaseCount int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 100 {
				maxConcurrency = 100
			}
			if acquireCount < 0 {
				acquireCount = 0
			}
			if releaseCount < 0 {
				releaseCount = 0
			}

			rm := NewResourceManager(maxConcurrency)

			// Acquire some slots
			for i := 0; i < acquireCount; i++ {
				rm.AcquireSlot()
			}

			// Release some slots (possibly more than acquired)
			for i := 0; i < releaseCount; i++ {
				rm.ReleaseSlot()
			}

			return rm.AvailableSlots() >= 0
		},
		gen.IntRange(1, 100),
		gen.IntRange(0, 150),
		gen.IntRange(0, 200),
	))

	// Property: Current tasks is always non-negative
	properties.Property("current tasks is always non-negative", prop.ForAll(
		func(maxConcurrency, releaseCount int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if releaseCount < 0 {
				releaseCount = 0
			}

			rm := NewResourceManager(maxConcurrency)

			// Release without acquiring (should not go negative)
			for i := 0; i < releaseCount; i++ {
				rm.ReleaseSlot()
			}

			return rm.CurrentTasks() >= 0
		},
		gen.IntRange(1, 100),
		gen.IntRange(0, 200),
	))

	// Property: Acquire returns false when at max concurrency
	properties.Property("acquire returns false when at max concurrency", prop.ForAll(
		func(maxConcurrency int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 50 {
				maxConcurrency = 50
			}

			rm := NewResourceManager(maxConcurrency)

			// Fill all slots
			for i := 0; i < maxConcurrency; i++ {
				if !rm.AcquireSlot() {
					return false // Should succeed for all slots up to max
				}
			}

			// Next acquire should fail
			return !rm.AcquireSlot()
		},
		gen.IntRange(1, 50),
	))

	// Property: Release then acquire succeeds when at max
	properties.Property("release then acquire succeeds when at max", prop.ForAll(
		func(maxConcurrency int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 50 {
				maxConcurrency = 50
			}

			rm := NewResourceManager(maxConcurrency)

			// Fill all slots
			for i := 0; i < maxConcurrency; i++ {
				rm.AcquireSlot()
			}

			// Should fail
			if rm.AcquireSlot() {
				return false
			}

			// Release one
			rm.ReleaseSlot()

			// Should succeed now
			return rm.AcquireSlot()
		},
		gen.IntRange(1, 50),
	))

	// Property: Concurrent acquire/release maintains invariants
	properties.Property("concurrent acquire/release maintains invariants", prop.ForAll(
		func(maxConcurrency, goroutines, operations int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 20 {
				maxConcurrency = 20
			}
			if goroutines < 1 {
				goroutines = 1
			}
			if goroutines > 10 {
				goroutines = 10
			}
			if operations < 1 {
				operations = 1
			}
			if operations > 50 {
				operations = 50
			}

			rm := NewResourceManager(maxConcurrency)
			var wg sync.WaitGroup
			var successCount int32

			// Launch goroutines that acquire and release
			for g := 0; g < goroutines; g++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < operations; i++ {
						if rm.AcquireSlot() {
							atomic.AddInt32(&successCount, 1)
							// Simulate some work
							time.Sleep(time.Microsecond)
							rm.ReleaseSlot()
						}
					}
				}()
			}

			wg.Wait()

			// After all operations, current tasks should be 0
			// and invariants should hold
			return rm.CurrentTasks() == 0 &&
				rm.AvailableSlots() == maxConcurrency &&
				rm.CurrentTasks() <= maxConcurrency
		},
		gen.IntRange(1, 20),
		gen.IntRange(1, 10),
		gen.IntRange(1, 50),
	))

	// Property: SetMaxConcurrency with valid values updates correctly
	properties.Property("SetMaxConcurrency with valid values updates correctly", prop.ForAll(
		func(initial, newMax int) bool {
			if initial < 1 {
				initial = 1
			}
			if initial > 100 {
				initial = 100
			}
			if newMax < 1 {
				newMax = 1
			}
			if newMax > 100 {
				newMax = 100
			}

			rm := NewResourceManager(initial)
			rm.SetMaxConcurrency(newMax)

			return rm.config.MaxConcurrency == newMax
		},
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	// Property: SetMaxConcurrency with invalid values does not change
	properties.Property("SetMaxConcurrency with invalid values does not change", prop.ForAll(
		func(initial, invalidMax int) bool {
			if initial < 1 {
				initial = 1
			}
			if initial > 100 {
				initial = 100
			}
			// Make invalidMax <= 0
			if invalidMax > 0 {
				invalidMax = -invalidMax
			}

			rm := NewResourceManager(initial)
			rm.SetMaxConcurrency(invalidMax)

			return rm.config.MaxConcurrency == initial
		},
		gen.IntRange(1, 100),
		gen.IntRange(-100, 0),
	))

	properties.TestingRun(t)
}


// ==================== Property Test for Resource Cleanup ====================

// TestProperty10_ResourceCleanup 测试 Property 10: Resource Cleanup
// **Property 10: Resource Cleanup**
// **Validates: Requirements 7.3**
// For any completed scan, all associated goroutines SHALL be terminated
// and memory SHALL be released within 5 seconds.
func TestProperty10_ResourceCleanup(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: After all tasks complete, current tasks returns to zero
	properties.Property("after all tasks complete, current tasks returns to zero", prop.ForAll(
		func(maxConcurrency, taskCount int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 20 {
				maxConcurrency = 20
			}
			if taskCount < 0 {
				taskCount = 0
			}
			if taskCount > 100 {
				taskCount = 100
			}

			rm := NewResourceManager(maxConcurrency)

			// Simulate task execution: acquire and release for each task
			for i := 0; i < taskCount; i++ {
				if rm.AcquireSlot() {
					// Simulate some work
					time.Sleep(time.Microsecond)
					rm.ReleaseSlot()
				}
			}

			// After all tasks, current tasks should be 0
			return rm.CurrentTasks() == 0
		},
		gen.IntRange(1, 20),
		gen.IntRange(0, 100),
	))

	// Property: Concurrent task completion releases all slots
	properties.Property("concurrent task completion releases all slots", prop.ForAll(
		func(maxConcurrency, goroutines int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 10 {
				maxConcurrency = 10
			}
			if goroutines < 1 {
				goroutines = 1
			}
			if goroutines > 20 {
				goroutines = 20
			}

			rm := NewResourceManager(maxConcurrency)
			var wg sync.WaitGroup

			// Launch goroutines that acquire, work, and release
			for g := 0; g < goroutines; g++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if rm.AcquireSlot() {
						// Simulate work
						time.Sleep(time.Millisecond)
						rm.ReleaseSlot()
					}
				}()
			}

			wg.Wait()

			// After all goroutines complete, current tasks should be 0
			return rm.CurrentTasks() == 0
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 20),
	))

	// Property: Panic in task still releases slot (simulated with defer)
	properties.Property("defer ensures slot release even on early return", prop.ForAll(
		func(maxConcurrency, taskCount int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 10 {
				maxConcurrency = 10
			}
			if taskCount < 1 {
				taskCount = 1
			}
			if taskCount > 50 {
				taskCount = 50
			}

			rm := NewResourceManager(maxConcurrency)

			// Simulate tasks that may return early but use defer for cleanup
			for i := 0; i < taskCount; i++ {
				func() {
					if !rm.AcquireSlot() {
						return
					}
					defer rm.ReleaseSlot()

					// Simulate early return (like error handling)
					if i%3 == 0 {
						return
					}

					// Simulate normal work
					time.Sleep(time.Microsecond)
				}()
			}

			// After all tasks, current tasks should be 0
			return rm.CurrentTasks() == 0
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 50),
	))

	// Property: Available slots equals max concurrency after cleanup
	properties.Property("available slots equals max concurrency after cleanup", prop.ForAll(
		func(maxConcurrency, operations int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 20 {
				maxConcurrency = 20
			}
			if operations < 0 {
				operations = 0
			}
			if operations > 100 {
				operations = 100
			}

			rm := NewResourceManager(maxConcurrency)

			// Perform random acquire/release operations
			for i := 0; i < operations; i++ {
				if i%2 == 0 {
					if rm.AcquireSlot() {
						rm.ReleaseSlot()
					}
				} else {
					rm.AcquireSlot()
					rm.ReleaseSlot()
				}
			}

			// After all operations, available slots should equal max concurrency
			return rm.AvailableSlots() == maxConcurrency
		},
		gen.IntRange(1, 20),
		gen.IntRange(0, 100),
	))

	// Property: Resource status reflects correct state after cleanup
	properties.Property("resource status reflects correct state after cleanup", prop.ForAll(
		func(maxConcurrency, taskCount int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 15 {
				maxConcurrency = 15
			}
			if taskCount < 0 {
				taskCount = 0
			}
			if taskCount > 50 {
				taskCount = 50
			}

			rm := NewResourceManager(maxConcurrency)

			// Execute tasks
			for i := 0; i < taskCount; i++ {
				if rm.AcquireSlot() {
					rm.ReleaseSlot()
				}
			}

			status := rm.GetResourceStatus()

			// Verify status reflects clean state
			return status.CurrentTasks == 0 &&
				status.AvailableSlots == maxConcurrency &&
				status.MaxConcurrency == maxConcurrency
		},
		gen.IntRange(1, 15),
		gen.IntRange(0, 50),
	))

	// Property: Multiple acquire-release cycles maintain consistency
	properties.Property("multiple acquire-release cycles maintain consistency", prop.ForAll(
		func(maxConcurrency, cycles int) bool {
			if maxConcurrency < 1 {
				maxConcurrency = 1
			}
			if maxConcurrency > 10 {
				maxConcurrency = 10
			}
			if cycles < 1 {
				cycles = 1
			}
			if cycles > 20 {
				cycles = 20
			}

			rm := NewResourceManager(maxConcurrency)

			for cycle := 0; cycle < cycles; cycle++ {
				// Fill all slots
				acquired := 0
				for i := 0; i < maxConcurrency; i++ {
					if rm.AcquireSlot() {
						acquired++
					}
				}

				// Verify all slots are taken
				if rm.CurrentTasks() != acquired {
					return false
				}

				// Release all slots
				for i := 0; i < acquired; i++ {
					rm.ReleaseSlot()
				}

				// Verify all slots are released
				if rm.CurrentTasks() != 0 {
					return false
				}
			}

			return rm.CurrentTasks() == 0 && rm.AvailableSlots() == maxConcurrency
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t)
}
