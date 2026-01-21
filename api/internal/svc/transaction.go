package svc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// TransactionManager 事务管理器
type TransactionManager struct {
	client *mongo.Client
}

// NewTransactionManager 创建事务管理器
func NewTransactionManager(client *mongo.Client) *TransactionManager {
	return &TransactionManager{client: client}
}

// TransactionOptions 事务选项
type TransactionOptions struct {
	MaxRetries int
	Timeout    time.Duration
}

// DefaultTransactionOptions 默认事务选项
func DefaultTransactionOptions() TransactionOptions {
	return TransactionOptions{
		MaxRetries: 3,
		Timeout:    30 * time.Second,
	}
}

// ExecuteInTransaction 在事务中执行操作
func (tm *TransactionManager) ExecuteInTransaction(
	ctx context.Context,
	fn func(sessCtx mongo.SessionContext) error,
) error {
	return tm.ExecuteInTransactionWithOptions(ctx, DefaultTransactionOptions(), fn)
}

// ExecuteInTransactionWithOptions 在事务中执行操作（带选项）
func (tm *TransactionManager) ExecuteInTransactionWithOptions(
	ctx context.Context,
	opts TransactionOptions,
	fn func(sessCtx mongo.SessionContext) error,
) error {
	// 设置超时
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// 创建会话
	session, err := tm.client.StartSession()
	if err != nil {
		return fmt.Errorf("start session: %w", err)
	}
	defer session.EndSession(ctx)

	// 事务选项
	txnOpts := options.Transaction().
		SetReadConcern(readconcern.Snapshot()).
		SetWriteConcern(writeconcern.Majority())

	// 执行事务
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	}, txnOpts)

	return err
}

// ExecuteWithRetry 带重试的事务执行
func (tm *TransactionManager) ExecuteWithRetry(
	ctx context.Context,
	maxRetries int,
	fn func(sessCtx mongo.SessionContext) error,
) error {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		err := tm.ExecuteInTransaction(ctx, fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// 检查是否是可重试的错误
		if !isRetryableTransactionError(err) {
			return err
		}

		// 指数退避
		backoff := time.Duration(1<<uint(i)) * 100 * time.Millisecond
		if backoff > 5*time.Second {
			backoff = 5 * time.Second
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
	}

	return fmt.Errorf("transaction failed after %d retries: %w", maxRetries, lastErr)
}

// isRetryableTransactionError 判断是否是可重试的事务错误
func isRetryableTransactionError(err error) bool {
	if err == nil {
		return false
	}

	// MongoDB 事务冲突和网络错误可重试
	if cmdErr, ok := err.(mongo.CommandError); ok {
		switch cmdErr.Code {
		case 112: // WriteConflict
			return true
		case 251: // TransactionAborted
			return true
		case 11600: // InterruptedAtShutdown
			return true
		case 11601: // Interrupted
			return true
		case 11602: // InterruptedDueToReplStateChange
			return true
		}
	}

	return false
}

// BulkOperationResult 批量操作结果
type BulkOperationResult struct {
	InsertedCount int64
	ModifiedCount int64
	DeletedCount  int64
	UpsertedCount int64
	Errors        []error
}

// BulkOperationBuilder 批量操作构建器
type BulkOperationBuilder struct {
	tm         *TransactionManager
	operations []func(sessCtx mongo.SessionContext) error
	mu         sync.Mutex
}

// NewBulkOperationBuilder 创建批量操作构建器
func (tm *TransactionManager) NewBulkOperationBuilder() *BulkOperationBuilder {
	return &BulkOperationBuilder{
		tm:         tm,
		operations: make([]func(sessCtx mongo.SessionContext) error, 0),
	}
}

// Add 添加操作
func (b *BulkOperationBuilder) Add(op func(sessCtx mongo.SessionContext) error) *BulkOperationBuilder {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.operations = append(b.operations, op)
	return b
}

// Execute 执行所有操作
func (b *BulkOperationBuilder) Execute(ctx context.Context) error {
	return b.tm.ExecuteInTransaction(ctx, func(sessCtx mongo.SessionContext) error {
		for _, op := range b.operations {
			if err := op(sessCtx); err != nil {
				return err
			}
		}
		return nil
	})
}

// ExecuteWithRetry 带重试执行所有操作
func (b *BulkOperationBuilder) ExecuteWithRetry(ctx context.Context, maxRetries int) error {
	return b.tm.ExecuteWithRetry(ctx, maxRetries, func(sessCtx mongo.SessionContext) error {
		for _, op := range b.operations {
			if err := op(sessCtx); err != nil {
				return err
			}
		}
		return nil
	})
}

// Clear 清空操作
func (b *BulkOperationBuilder) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.operations = b.operations[:0]
}

// Count 获取操作数量
func (b *BulkOperationBuilder) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.operations)
}
