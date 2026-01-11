package xerr

import (
	"context"
	"errors"
	"fmt"
)

// ScanError 扫描错误
// 用于封装扫描过程中发生的错误，包含扫描器名称、目标和阶段信息
type ScanError struct {
	Scanner string // 扫描器名称
	Target  string // 扫描目标
	Phase   string // 扫描阶段
	Cause   error  // 原始错误
}

// Error 实现 error 接口
func (e *ScanError) Error() string {
	return fmt.Sprintf("scan error [%s] target=%s phase=%s: %v",
		e.Scanner, e.Target, e.Phase, e.Cause)
}

// Unwrap 实现 errors.Unwrap 接口，支持错误链
func (e *ScanError) Unwrap() error {
	return e.Cause
}

// NewScanError 创建扫描错误
func NewScanError(scanner, target, phase string, cause error) *ScanError {
	return &ScanError{
		Scanner: scanner,
		Target:  target,
		Phase:   phase,
		Cause:   cause,
	}
}

// ConfigError 配置错误
// 用于封装配置验证过程中发生的错误
type ConfigError struct {
	Field   string      // 字段名称
	Value   interface{} // 字段值
	Message string      // 错误消息
}

// Error 实现 error 接口
func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error: field=%s value=%v msg=%s",
		e.Field, e.Value, e.Message)
}

// NewConfigError 创建配置错误
func NewConfigError(field string, value interface{}, message string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// NetworkError 网络错误
// 用于封装网络操作过程中发生的错误
type NetworkError struct {
	Host  string // 主机地址
	Port  int    // 端口号
	Op    string // 操作类型（如 connect, read, write）
	Cause error  // 原始错误
}

// Error 实现 error 接口
func (e *NetworkError) Error() string {
	return fmt.Sprintf("network error: %s %s:%d: %v",
		e.Op, e.Host, e.Port, e.Cause)
}

// Unwrap 实现 errors.Unwrap 接口，支持错误链
func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// NewNetworkError 创建网络错误
func NewNetworkError(host string, port int, op string, cause error) *NetworkError {
	return &NetworkError{
		Host:  host,
		Port:  port,
		Op:    op,
		Cause: cause,
	}
}

// IsRetryable 判断错误是否可重试
// 网络错误和超时错误通常可以重试
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// 网络错误可重试
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return true
	}

	// 超时错误可重试
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// 配置错误不可重试
	var configErr *ConfigError
	if errors.As(err, &configErr) {
		return false
	}

	return false
}

// IsScanError 判断是否为扫描错误
func IsScanError(err error) bool {
	var scanErr *ScanError
	return errors.As(err, &scanErr)
}

// IsConfigError 判断是否为配置错误
func IsConfigError(err error) bool {
	var configErr *ConfigError
	return errors.As(err, &configErr)
}

// IsNetworkError 判断是否为网络错误
func IsNetworkError(err error) bool {
	var netErr *NetworkError
	return errors.As(err, &netErr)
}

// GetScanError 从错误链中提取 ScanError
func GetScanError(err error) *ScanError {
	var scanErr *ScanError
	if errors.As(err, &scanErr) {
		return scanErr
	}
	return nil
}

// GetConfigError 从错误链中提取 ConfigError
func GetConfigError(err error) *ConfigError {
	var configErr *ConfigError
	if errors.As(err, &configErr) {
		return configErr
	}
	return nil
}

// GetNetworkError 从错误链中提取 NetworkError
func GetNetworkError(err error) *NetworkError {
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return netErr
	}
	return nil
}
