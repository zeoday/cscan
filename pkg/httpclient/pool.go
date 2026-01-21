package httpclient

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	// DefaultClient 默认全局连接池
	DefaultClient *http.Client

	// HighConcurrencyClient 高并发场景连接池
	HighConcurrencyClient *http.Client

	// LongLivedClient 长连接场景连接池
	LongLivedClient *http.Client

	// InsecureClient 跳过TLS验证的客户端
	InsecureClient *http.Client

	initOnce sync.Once
)

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
	IdleConnTimeout     time.Duration
	Timeout             time.Duration
	SkipVerify          bool
	KeepAlive           time.Duration
	TLSHandshakeTimeout time.Duration
}

// DefaultPoolConfig 默认连接池配置
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		MaxConnsPerHost:     50,
		IdleConnTimeout:     90 * time.Second,
		Timeout:             30 * time.Second,
		SkipVerify:          false,
		KeepAlive:           30 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}

// HighConcurrencyPoolConfig 高并发连接池配置
func HighConcurrencyPoolConfig() PoolConfig {
	return PoolConfig{
		MaxIdleConns:        500,
		MaxIdleConnsPerHost: 50,
		MaxConnsPerHost:     100,
		IdleConnTimeout:     60 * time.Second,
		Timeout:             15 * time.Second,
		SkipVerify:          true,
		KeepAlive:           30 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}
}

// LongLivedPoolConfig 长连接池配置
func LongLivedPoolConfig() PoolConfig {
	return PoolConfig{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 5,
		MaxConnsPerHost:     20,
		IdleConnTimeout:     300 * time.Second,
		Timeout:             120 * time.Second,
		SkipVerify:          false,
		KeepAlive:           60 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}

// Init 初始化全局客户端
func Init() {
	initOnce.Do(func() {
		DefaultClient = NewPooledClient(DefaultPoolConfig())
		HighConcurrencyClient = NewPooledClient(HighConcurrencyPoolConfig())
		LongLivedClient = NewPooledClient(LongLivedPoolConfig())

		insecureCfg := DefaultPoolConfig()
		insecureCfg.SkipVerify = true
		InsecureClient = NewPooledClient(insecureCfg)
	})
}

// NewPooledClient 创建带连接池的HTTP客户端
func NewPooledClient(cfg PoolConfig) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: cfg.KeepAlive,
		}).DialContext,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		TLSHandshakeTimeout: cfg.TLSHandshakeTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SkipVerify,
		},
		ForceAttemptHTTP2:     true,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: cfg.Timeout,
		DisableKeepAlives:     false,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}
}

// NewPooledClientWithTransport 创建带自定义Transport的客户端
func NewPooledClientWithTransport(cfg PoolConfig, customTransport func(*http.Transport)) *http.Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: cfg.KeepAlive,
		}).DialContext,
		MaxIdleConns:        cfg.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.MaxConnsPerHost,
		IdleConnTimeout:     cfg.IdleConnTimeout,
		TLSHandshakeTimeout: cfg.TLSHandshakeTimeout,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SkipVerify,
		},
		ForceAttemptHTTP2:     true,
		ExpectContinueTimeout: 1 * time.Second,
	}

	if customTransport != nil {
		customTransport(transport)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
	}
}

// Get 使用默认客户端发送GET请求
func Get(url string) (*http.Response, error) {
	Init()
	return DefaultClient.Get(url)
}

// GetWithTimeout 使用指定超时发送GET请求
func GetWithTimeout(url string, timeout time.Duration) (*http.Response, error) {
	cfg := DefaultPoolConfig()
	cfg.Timeout = timeout
	client := NewPooledClient(cfg)
	return client.Get(url)
}

// GetInsecure 跳过TLS验证发送GET请求
func GetInsecure(url string) (*http.Response, error) {
	Init()
	return InsecureClient.Get(url)
}

// Do 使用默认客户端执行请求
func Do(req *http.Request) (*http.Response, error) {
	Init()
	return DefaultClient.Do(req)
}

// DoWithClient 使用指定客户端执行请求
func DoWithClient(client *http.Client, req *http.Request) (*http.Response, error) {
	if client == nil {
		Init()
		client = DefaultClient
	}
	return client.Do(req)
}

// CloseIdleConnections 关闭所有空闲连接
func CloseIdleConnections() {
	if DefaultClient != nil {
		if t, ok := DefaultClient.Transport.(*http.Transport); ok {
			t.CloseIdleConnections()
		}
	}
	if HighConcurrencyClient != nil {
		if t, ok := HighConcurrencyClient.Transport.(*http.Transport); ok {
			t.CloseIdleConnections()
		}
	}
	if LongLivedClient != nil {
		if t, ok := LongLivedClient.Transport.(*http.Transport); ok {
			t.CloseIdleConnections()
		}
	}
	if InsecureClient != nil {
		if t, ok := InsecureClient.Transport.(*http.Transport); ok {
			t.CloseIdleConnections()
		}
	}
}

// PoolStats 连接池统计信息
type PoolStats struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	MaxConnsPerHost     int
}

// GetPoolStats 获取连接池统计
func GetPoolStats(client *http.Client) *PoolStats {
	if client == nil {
		return nil
	}
	t, ok := client.Transport.(*http.Transport)
	if !ok {
		return nil
	}
	return &PoolStats{
		MaxIdleConns:        t.MaxIdleConns,
		MaxIdleConnsPerHost: t.MaxIdleConnsPerHost,
		MaxConnsPerHost:     t.MaxConnsPerHost,
	}
}
