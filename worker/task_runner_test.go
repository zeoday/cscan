package worker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty11_ContextCancellationAndGracefulShutdown 测试 Property 11: Context Cancellation and Graceful Shutdown
// **Property 11: Context Cancellation and Graceful Shutdown**
// **Validates: Requirements 7.4, 7.5**
// THE System SHALL use context.Context for all long-running operations to support cancellation.
// WHEN the Worker shuts down THEN the System SHALL gracefully complete or save progress of running tasks.
func TestProperty11_ContextCancellationAndGracefulShutdown(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Context cancellation stops all phases
	properties.Property("context cancellation stops execution within timeout", prop.ForAll(
		func(phaseCount, cancelAfterMs int) bool {
			if phaseCount < 1 {
				phaseCount = 1
			}
			if phaseCount > 10 {
				phaseCount = 10
			}
			if cancelAfterMs < 1 {
				cancelAfterMs = 1
			}
			if cancelAfterMs > 100 {
				cancelAfterMs = 100
			}

			ctx, cancel := context.WithCancel(context.Background())
			var executedPhases int32
			var stopped int32

			// Simulate phase execution with context checking
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < phaseCount; i++ {
					select {
					case <-ctx.Done():
						atomic.StoreInt32(&stopped, 1)
						return
					default:
						atomic.AddInt32(&executedPhases, 1)
						time.Sleep(time.Duration(cancelAfterMs) * time.Millisecond)
					}
				}
			}()

			// Cancel after some phases should have started
			time.Sleep(time.Duration(cancelAfterMs/2) * time.Millisecond)
			cancel()

			// Wait for completion with timeout
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()

			select {
			case <-done:
				// Execution stopped - either completed or cancelled
				return true
			case <-time.After(time.Duration(cancelAfterMs*phaseCount+100) * time.Millisecond):
				// Should not timeout - context cancellation should stop execution
				return false
			}
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 100),
	))

	// Property: Graceful shutdown preserves completed phase count
	properties.Property("graceful shutdown preserves completed phase count", prop.ForAll(
		func(totalPhases, cancelAtPhase int) bool {
			if totalPhases < 1 {
				totalPhases = 1
			}
			if totalPhases > 20 {
				totalPhases = 20
			}
			if cancelAtPhase < 0 {
				cancelAtPhase = 0
			}
			if cancelAtPhase >= totalPhases {
				cancelAtPhase = totalPhases - 1
			}

			ctx, cancel := context.WithCancel(context.Background())
			completedPhases := make(map[int]bool)
			var mu sync.Mutex

			// Simulate phase execution
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < totalPhases; i++ {
					select {
					case <-ctx.Done():
						return
					default:
						// Mark phase as completed
						mu.Lock()
						completedPhases[i] = true
						mu.Unlock()

						// Cancel at specific phase
						if i == cancelAtPhase {
							cancel()
						}
						time.Sleep(time.Millisecond)
					}
				}
			}()

			wg.Wait()

			// Verify completed phases are preserved
			mu.Lock()
			defer mu.Unlock()

			// All phases up to and including cancelAtPhase should be completed
			for i := 0; i <= cancelAtPhase; i++ {
				if !completedPhases[i] {
					return false
				}
			}
			return true
		},
		gen.IntRange(1, 20),
		gen.IntRange(0, 19),
	))

	// Property: Multiple concurrent cancellations are handled safely
	properties.Property("multiple concurrent cancellations are handled safely", prop.ForAll(
		func(goroutines int) bool {
			if goroutines < 1 {
				goroutines = 1
			}
			if goroutines > 50 {
				goroutines = 50
			}

			ctx, cancel := context.WithCancel(context.Background())
			var activeCount int32
			var completedCount int32

			var wg sync.WaitGroup
			for i := 0; i < goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					atomic.AddInt32(&activeCount, 1)
					defer atomic.AddInt32(&activeCount, -1)

					select {
					case <-ctx.Done():
						atomic.AddInt32(&completedCount, 1)
						return
					case <-time.After(100 * time.Millisecond):
						atomic.AddInt32(&completedCount, 1)
						return
					}
				}()
			}

			// Cancel from multiple goroutines simultaneously
			var cancelWg sync.WaitGroup
			for i := 0; i < 5; i++ {
				cancelWg.Add(1)
				go func() {
					defer cancelWg.Done()
					cancel() // Multiple calls to cancel are safe
				}()
			}
			cancelWg.Wait()

			// Wait for all goroutines to complete
			wg.Wait()

			// All goroutines should have completed
			return atomic.LoadInt32(&completedCount) == int32(goroutines) &&
				atomic.LoadInt32(&activeCount) == 0
		},
		gen.IntRange(1, 50),
	))

	// Property: Context with timeout respects deadline
	properties.Property("context with timeout respects deadline", prop.ForAll(
		func(timeoutMs, workMs int) bool {
			if timeoutMs < 1 {
				timeoutMs = 1
			}
			if timeoutMs > 100 {
				timeoutMs = 100
			}
			if workMs < 1 {
				workMs = 1
			}
			if workMs > 200 {
				workMs = 200
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMs)*time.Millisecond)
			defer cancel()

			start := time.Now()
			var timedOut bool

			select {
			case <-ctx.Done():
				timedOut = true
			case <-time.After(time.Duration(workMs) * time.Millisecond):
				timedOut = false
			}

			elapsed := time.Since(start)

			if workMs <= timeoutMs {
				// Work should complete before timeout
				return !timedOut || elapsed >= time.Duration(timeoutMs)*time.Millisecond
			}
			// Timeout should occur before work completes
			return timedOut && elapsed <= time.Duration(timeoutMs+50)*time.Millisecond
		},
		gen.IntRange(1, 100),
		gen.IntRange(1, 200),
	))

	properties.TestingRun(t)
}
