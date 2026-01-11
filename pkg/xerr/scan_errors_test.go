package xerr

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// ==================== Unit Tests ====================

// TestScanErrorBasic 测试 ScanError 基本功能
func TestScanErrorBasic(t *testing.T) {
	cause := errors.New("connection refused")
	err := NewScanError("naabu", "192.168.1.1", "port_scan", cause)

	if err.Scanner != "naabu" {
		t.Errorf("Scanner = %s, want naabu", err.Scanner)
	}
	if err.Target != "192.168.1.1" {
		t.Errorf("Target = %s, want 192.168.1.1", err.Target)
	}
	if err.Phase != "port_scan" {
		t.Errorf("Phase = %s, want port_scan", err.Phase)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

// TestScanErrorString 测试 ScanError 字符串格式
func TestScanErrorString(t *testing.T) {
	cause := errors.New("timeout")
	err := NewScanError("nuclei", "example.com", "vuln_scan", cause)

	errStr := err.Error()
	if !strings.Contains(errStr, "nuclei") {
		t.Errorf("Error string should contain scanner name: %s", errStr)
	}
	if !strings.Contains(errStr, "example.com") {
		t.Errorf("Error string should contain target: %s", errStr)
	}
	if !strings.Contains(errStr, "vuln_scan") {
		t.Errorf("Error string should contain phase: %s", errStr)
	}
	if !strings.Contains(errStr, "timeout") {
		t.Errorf("Error string should contain cause: %s", errStr)
	}
}

// TestScanErrorUnwrap 测试 ScanError 错误链
func TestScanErrorUnwrap(t *testing.T) {
	cause := errors.New("original error")
	err := NewScanError("httpx", "test.com", "fingerprint", cause)

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

// TestConfigErrorBasic 测试 ConfigError 基本功能
func TestConfigErrorBasic(t *testing.T) {
	err := NewConfigError("ports", "", "ports cannot be empty")

	if err.Field != "ports" {
		t.Errorf("Field = %s, want ports", err.Field)
	}
	if err.Value != "" {
		t.Errorf("Value = %v, want empty string", err.Value)
	}
	if err.Message != "ports cannot be empty" {
		t.Errorf("Message = %s, want 'ports cannot be empty'", err.Message)
	}
}

// TestConfigErrorString 测试 ConfigError 字符串格式
func TestConfigErrorString(t *testing.T) {
	err := NewConfigError("rate", -1, "rate must be non-negative")

	errStr := err.Error()
	if !strings.Contains(errStr, "rate") {
		t.Errorf("Error string should contain field name: %s", errStr)
	}
	if !strings.Contains(errStr, "-1") {
		t.Errorf("Error string should contain value: %s", errStr)
	}
	if !strings.Contains(errStr, "rate must be non-negative") {
		t.Errorf("Error string should contain message: %s", errStr)
	}
}

// TestNetworkErrorBasic 测试 NetworkError 基本功能
func TestNetworkErrorBasic(t *testing.T) {
	cause := errors.New("connection refused")
	err := NewNetworkError("192.168.1.1", 80, "connect", cause)

	if err.Host != "192.168.1.1" {
		t.Errorf("Host = %s, want 192.168.1.1", err.Host)
	}
	if err.Port != 80 {
		t.Errorf("Port = %d, want 80", err.Port)
	}
	if err.Op != "connect" {
		t.Errorf("Op = %s, want connect", err.Op)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

// TestNetworkErrorString 测试 NetworkError 字符串格式
func TestNetworkErrorString(t *testing.T) {
	cause := errors.New("timeout")
	err := NewNetworkError("example.com", 443, "read", cause)

	errStr := err.Error()
	if !strings.Contains(errStr, "example.com") {
		t.Errorf("Error string should contain host: %s", errStr)
	}
	if !strings.Contains(errStr, "443") {
		t.Errorf("Error string should contain port: %s", errStr)
	}
	if !strings.Contains(errStr, "read") {
		t.Errorf("Error string should contain operation: %s", errStr)
	}
	if !strings.Contains(errStr, "timeout") {
		t.Errorf("Error string should contain cause: %s", errStr)
	}
}

// TestNetworkErrorUnwrap 测试 NetworkError 错误链
func TestNetworkErrorUnwrap(t *testing.T) {
	cause := errors.New("original error")
	err := NewNetworkError("test.com", 8080, "write", cause)

	unwrapped := errors.Unwrap(err)
	if unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}
}

// TestIsRetryable 测试 IsRetryable 函数
func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"network error", NewNetworkError("host", 80, "connect", errors.New("refused")), true},
		{"config error", NewConfigError("field", "value", "invalid"), false},
		{"context deadline exceeded", context.DeadlineExceeded, true},
		{"wrapped network error", fmt.Errorf("wrapped: %w", NewNetworkError("host", 80, "connect", errors.New("refused"))), true},
		{"wrapped config error", fmt.Errorf("wrapped: %w", NewConfigError("field", "value", "invalid")), false},
		{"generic error", errors.New("some error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryable(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryable(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

// TestErrorTypeCheckers 测试错误类型检查函数
func TestErrorTypeCheckers(t *testing.T) {
	scanErr := NewScanError("scanner", "target", "phase", errors.New("cause"))
	configErr := NewConfigError("field", "value", "message")
	netErr := NewNetworkError("host", 80, "op", errors.New("cause"))
	genericErr := errors.New("generic")

	// Test IsScanError
	if !IsScanError(scanErr) {
		t.Error("IsScanError should return true for ScanError")
	}
	if IsScanError(configErr) {
		t.Error("IsScanError should return false for ConfigError")
	}
	if IsScanError(genericErr) {
		t.Error("IsScanError should return false for generic error")
	}

	// Test IsConfigError
	if !IsConfigError(configErr) {
		t.Error("IsConfigError should return true for ConfigError")
	}
	if IsConfigError(scanErr) {
		t.Error("IsConfigError should return false for ScanError")
	}

	// Test IsNetworkError
	if !IsNetworkError(netErr) {
		t.Error("IsNetworkError should return true for NetworkError")
	}
	if IsNetworkError(scanErr) {
		t.Error("IsNetworkError should return false for ScanError")
	}
}

// TestGetErrorFunctions 测试错误提取函数
func TestGetErrorFunctions(t *testing.T) {
	scanErr := NewScanError("scanner", "target", "phase", errors.New("cause"))
	configErr := NewConfigError("field", "value", "message")
	netErr := NewNetworkError("host", 80, "op", errors.New("cause"))

	// Test GetScanError
	if GetScanError(scanErr) != scanErr {
		t.Error("GetScanError should return the ScanError")
	}
	if GetScanError(configErr) != nil {
		t.Error("GetScanError should return nil for non-ScanError")
	}

	// Test GetConfigError
	if GetConfigError(configErr) != configErr {
		t.Error("GetConfigError should return the ConfigError")
	}
	if GetConfigError(scanErr) != nil {
		t.Error("GetConfigError should return nil for non-ConfigError")
	}

	// Test GetNetworkError
	if GetNetworkError(netErr) != netErr {
		t.Error("GetNetworkError should return the NetworkError")
	}
	if GetNetworkError(scanErr) != nil {
		t.Error("GetNetworkError should return nil for non-NetworkError")
	}
}

// TestWrappedErrorExtraction 测试从包装错误中提取
func TestWrappedErrorExtraction(t *testing.T) {
	scanErr := NewScanError("scanner", "target", "phase", errors.New("cause"))
	wrapped := fmt.Errorf("outer: %w", fmt.Errorf("inner: %w", scanErr))

	extracted := GetScanError(wrapped)
	if extracted == nil {
		t.Error("GetScanError should extract from wrapped error")
	}
	if extracted.Scanner != "scanner" {
		t.Errorf("Extracted Scanner = %s, want scanner", extracted.Scanner)
	}
}

// ==================== Property Tests ====================

// TestProperty1_ErrorContextPreservation 测试 Property 1: Error Context Preservation
// **Property 1: Error Context Preservation**
// **Validates: Requirements 1.4, 5.2**
// For any error returned by a scanner or worker, unwrapping the error chain
// SHALL reveal the original cause with full context (scanner name, target, phase).
func TestProperty1_ErrorContextPreservation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: ScanError preserves all context through error chain
	properties.Property("ScanError preserves scanner, target, phase, and cause", prop.ForAll(
		func(scanner, target, phase, causeMsg string) bool {
			cause := errors.New(causeMsg)
			err := NewScanError(scanner, target, phase, cause)

			// Verify context is preserved in error message
			errStr := err.Error()
			if !strings.Contains(errStr, scanner) {
				return false
			}
			if !strings.Contains(errStr, target) {
				return false
			}
			if !strings.Contains(errStr, phase) {
				return false
			}
			if !strings.Contains(errStr, causeMsg) {
				return false
			}

			// Verify unwrap reveals original cause
			unwrapped := errors.Unwrap(err)
			if unwrapped == nil || unwrapped.Error() != causeMsg {
				return false
			}

			// Verify errors.Is works with cause
			if !errors.Is(err, cause) {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	// Property: NetworkError preserves all context through error chain
	properties.Property("NetworkError preserves host, port, op, and cause", prop.ForAll(
		func(host string, port int, op, causeMsg string) bool {
			// Ensure port is valid
			if port < 0 || port > 65535 {
				port = port % 65536
				if port < 0 {
					port = -port
				}
			}

			cause := errors.New(causeMsg)
			err := NewNetworkError(host, port, op, cause)

			// Verify context is preserved in error message
			errStr := err.Error()
			if !strings.Contains(errStr, host) {
				return false
			}
			if !strings.Contains(errStr, fmt.Sprintf("%d", port)) {
				return false
			}
			if !strings.Contains(errStr, op) {
				return false
			}
			if !strings.Contains(errStr, causeMsg) {
				return false
			}

			// Verify unwrap reveals original cause
			unwrapped := errors.Unwrap(err)
			if unwrapped == nil || unwrapped.Error() != causeMsg {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.IntRange(0, 65535),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	// Property: ConfigError preserves all context
	properties.Property("ConfigError preserves field, value, and message", prop.ForAll(
		func(field, message string, value int) bool {
			err := NewConfigError(field, value, message)

			// Verify context is preserved in error message
			errStr := err.Error()
			if !strings.Contains(errStr, field) {
				return false
			}
			if !strings.Contains(errStr, fmt.Sprintf("%d", value)) {
				return false
			}
			if !strings.Contains(errStr, message) {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.Int(),
	))

	// Property: Wrapped errors preserve context through errors.As
	properties.Property("Wrapped ScanError can be extracted with errors.As", prop.ForAll(
		func(scanner, target, phase, causeMsg string, wrapLevels int) bool {
			// Limit wrap levels to reasonable range
			if wrapLevels < 0 {
				wrapLevels = -wrapLevels
			}
			wrapLevels = wrapLevels % 5

			cause := errors.New(causeMsg)
			var err error = NewScanError(scanner, target, phase, cause)

			// Wrap the error multiple times
			for i := 0; i < wrapLevels; i++ {
				err = fmt.Errorf("wrap level %d: %w", i, err)
			}

			// Verify we can extract the original ScanError
			var scanErr *ScanError
			if !errors.As(err, &scanErr) {
				return false
			}

			// Verify context is preserved
			if scanErr.Scanner != scanner {
				return false
			}
			if scanErr.Target != target {
				return false
			}
			if scanErr.Phase != phase {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.IntRange(0, 10),
	))

	properties.TestingRun(t)
}
