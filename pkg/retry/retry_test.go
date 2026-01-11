package retry

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cscan/pkg/xerr"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// ==================== Unit Tests ====================

// TestDoSuccess 测试成功执行
func TestDoSuccess(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	err := Do(ctx, func() error {
		callCount++
		return nil
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1", callCount)
	}
}

// TestDoRetryableError 测试可重试错误
func TestDoRetryableError(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	config := Config{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
	}

	err := DoWithConfig(ctx, config, func() error {
		callCount++
		return xerr.NewNetworkError("host", 80, "connect", errors.New("refused"))
	})

	if err == nil {
		t.Error("Do() should return error after max retries")
	}
	// 初始尝试 + 3次重试 = 4次
	if callCount != 4 {
		t.Errorf("callCount = %d, want 4", callCount)
	}
	if !strings.Contains(err.Error(), "max retries") {
		t.Errorf("error should mention max retries: %v", err)
	}
}

// TestDoNonRetryableError 测试不可重试错误
func TestDoNonRetryableError(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	config := Config{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
	}

	configErr := xerr.NewConfigError("field", "value", "invalid")
	err := DoWithConfig(ctx, config, func() error {
		callCount++
		return configErr
	})

	if err != configErr {
		t.Errorf("Do() should return original error for non-retryable: %v", err)
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (no retries for non-retryable)", callCount)
	}
}

// TestDoContextCancellation 测试 context 取消
func TestDoContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	callCount := 0

	config := Config{
		MaxRetries:     10,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
		Multiplier:     2.0,
	}

	// 在第一次失败后取消 context
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := DoWithConfig(ctx, config, func() error {
		callCount++
		return xerr.NewNetworkError("host", 80, "connect", errors.New("refused"))
	})

	if err == nil {
		t.Error("Do() should return error when context is cancelled")
	}
	if !strings.Contains(err.Error(), "context cancelled") {
		t.Errorf("error should mention context cancelled: %v", err)
	}
}

// TestDoSuccessAfterRetries 测试重试后成功
func TestDoSuccessAfterRetries(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	config := Config{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
	}

	err := DoWithConfig(ctx, config, func() error {
		callCount++
		if callCount < 3 {
			return xerr.NewNetworkError("host", 80, "connect", errors.New("refused"))
		}
		return nil
	})

	if err != nil {
		t.Errorf("Do() error = %v, want nil", err)
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

// TestDoWithResult 测试带结果的重试
func TestDoWithResult(t *testing.T) {
	ctx := context.Background()

	config := Config{
		MaxRetries:     3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     2.0,
	}

	callCount := 0
	result := DoWithResult(ctx, config, func() error {
		callCount++
		if callCount < 2 {
			return xerr.NewNetworkError("host", 80, "connect", errors.New("refused"))
		}
		return nil
	})

	if result.Err != nil {
		t.Errorf("DoWithResult() error = %v, want nil", result.Err)
	}
	if result.Attempts != 2 {
		t.Errorf("Attempts = %d, want 2", result.Attempts)
	}
}

// TestWithMaxRetries 测试配置辅助函数
func TestWithMaxRetries(t *testing.T) {
	config := WithMaxRetries(5)
	if config.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", config.MaxRetries)
	}
	// 其他字段应该是默认值
	if config.InitialBackoff != DefaultConfig.InitialBackoff {
		t.Errorf("InitialBackoff = %v, want %v", config.InitialBackoff, DefaultConfig.InitialBackoff)
	}
}

// TestWithBackoff 测试退避配置辅助函数
func TestWithBackoff(t *testing.T) {
	config := WithBackoff(100*time.Millisecond, 5*time.Second, 1.5)
	if config.InitialBackoff != 100*time.Millisecond {
		t.Errorf("InitialBackoff = %v, want 100ms", config.InitialBackoff)
	}
	if config.MaxBackoff != 5*time.Second {
		t.Errorf("MaxBackoff = %v, want 5s", config.MaxBackoff)
	}
	if config.Multiplier != 1.5 {
		t.Errorf("Multiplier = %v, want 1.5", config.Multiplier)
	}
}

// TestNewConfig 测试创建自定义配置
func TestNewConfig(t *testing.T) {
	config := NewConfig(5, 50*time.Millisecond, 2*time.Second, 3.0)
	if config.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", config.MaxRetries)
	}
	if config.InitialBackoff != 50*time.Millisecond {
		t.Errorf("InitialBackoff = %v, want 50ms", config.InitialBackoff)
	}
	if config.MaxBackoff != 2*time.Second {
		t.Errorf("MaxBackoff = %v, want 2s", config.MaxBackoff)
	}
	if config.Multiplier != 3.0 {
		t.Errorf("Multiplier = %v, want 3.0", config.Multiplier)
	}
}

// ==================== Property Tests ====================

// TestProperty4_TaskTimeoutAndRetry 测试 Property 4: Task Timeout and Retry
// **Property 4: Task Timeout and Retry**
// **Validates: Requirements 3.4**
// For any task that times out, the scheduler SHALL retry with exponential backoff,
// and the retry count SHALL not exceed the configured maximum.
func TestProperty4_TaskTimeoutAndRetry(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Retry count never exceeds MaxRetries + 1 (initial attempt + retries)
	properties.Property("retry count never exceeds MaxRetries + 1", prop.ForAll(
		func(maxRetries int) bool {
			// Limit to reasonable range
			if maxRetries < 0 {
				maxRetries = 0
			}
			if maxRetries > 10 {
				maxRetries = maxRetries % 10
			}

			ctx := context.Background()
			var callCount int32

			config := Config{
				MaxRetries:     maxRetries,
				InitialBackoff: 1 * time.Millisecond,
				MaxBackoff:     10 * time.Millisecond,
				Multiplier:     2.0,
			}

			_ = DoWithConfig(ctx, config, func() error {
				atomic.AddInt32(&callCount, 1)
				return xerr.NewNetworkError("host", 80, "connect", errors.New("timeout"))
			})

			// Total attempts should be maxRetries + 1 (initial attempt)
			expectedAttempts := int32(maxRetries + 1)
			return atomic.LoadInt32(&callCount) == expectedAttempts
		},
		gen.IntRange(0, 10),
	))

	// Property: Non-retryable errors are not retried
	properties.Property("non-retryable errors are not retried", prop.ForAll(
		func(maxRetries int) bool {
			if maxRetries < 1 {
				maxRetries = 1
			}
			if maxRetries > 10 {
				maxRetries = maxRetries % 10
			}

			ctx := context.Background()
			var callCount int32

			config := Config{
				MaxRetries:     maxRetries,
				InitialBackoff: 1 * time.Millisecond,
				MaxBackoff:     10 * time.Millisecond,
				Multiplier:     2.0,
			}

			_ = DoWithConfig(ctx, config, func() error {
				atomic.AddInt32(&callCount, 1)
				return xerr.NewConfigError("field", "value", "invalid")
			})

			// Should only be called once for non-retryable errors
			return atomic.LoadInt32(&callCount) == 1
		},
		gen.IntRange(1, 10),
	))

	// Property: Successful operation stops retrying
	properties.Property("successful operation stops retrying", prop.ForAll(
		func(maxRetries, successAfter int) bool {
			if maxRetries < 1 {
				maxRetries = 1
			}
			if maxRetries > 10 {
				maxRetries = maxRetries % 10
			}
			if successAfter < 1 {
				successAfter = 1
			}
			// Ensure successAfter is within retry range
			if successAfter > maxRetries+1 {
				successAfter = (successAfter % maxRetries) + 1
			}

			ctx := context.Background()
			var callCount int32

			config := Config{
				MaxRetries:     maxRetries,
				InitialBackoff: 1 * time.Millisecond,
				MaxBackoff:     10 * time.Millisecond,
				Multiplier:     2.0,
			}

			err := DoWithConfig(ctx, config, func() error {
				count := atomic.AddInt32(&callCount, 1)
				if int(count) < successAfter {
					return xerr.NewNetworkError("host", 80, "connect", errors.New("timeout"))
				}
				return nil
			})

			// Should succeed and stop at successAfter attempts
			return err == nil && atomic.LoadInt32(&callCount) == int32(successAfter)
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 10),
	))

	// Property: Context cancellation stops retrying
	properties.Property("context cancellation stops retrying", prop.ForAll(
		func(maxRetries int) bool {
			if maxRetries < 2 {
				maxRetries = 2
			}
			if maxRetries > 10 {
				maxRetries = maxRetries % 10
			}

			ctx, cancel := context.WithCancel(context.Background())
			var callCount int32

			config := Config{
				MaxRetries:     maxRetries,
				InitialBackoff: 50 * time.Millisecond,
				MaxBackoff:     100 * time.Millisecond,
				Multiplier:     2.0,
			}

			// Cancel after first attempt
			go func() {
				time.Sleep(10 * time.Millisecond)
				cancel()
			}()

			err := DoWithConfig(ctx, config, func() error {
				atomic.AddInt32(&callCount, 1)
				return xerr.NewNetworkError("host", 80, "connect", errors.New("timeout"))
			})

			// Should have error and fewer attempts than max
			count := atomic.LoadInt32(&callCount)
			return err != nil && count < int32(maxRetries+1)
		},
		gen.IntRange(2, 10),
	))

	// Property: DoWithResult returns correct attempt count
	properties.Property("DoWithResult returns correct attempt count", prop.ForAll(
		func(maxRetries, failCount int) bool {
			if maxRetries < 0 {
				maxRetries = 0
			}
			if maxRetries > 10 {
				maxRetries = maxRetries % 10
			}
			if failCount < 0 {
				failCount = 0
			}

			ctx := context.Background()
			var callCount int32

			config := Config{
				MaxRetries:     maxRetries,
				InitialBackoff: 1 * time.Millisecond,
				MaxBackoff:     10 * time.Millisecond,
				Multiplier:     2.0,
			}

			result := DoWithResult(ctx, config, func() error {
				count := atomic.AddInt32(&callCount, 1)
				if int(count) <= failCount {
					return xerr.NewNetworkError("host", 80, "connect", errors.New("timeout"))
				}
				return nil
			})

			// Verify attempt count matches actual calls
			return result.Attempts == int(atomic.LoadInt32(&callCount))
		},
		gen.IntRange(0, 10),
		gen.IntRange(0, 15),
	))

	properties.TestingRun(t)
}
