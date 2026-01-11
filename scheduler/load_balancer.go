package scheduler

import (
	"context"
	"encoding/json"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// LoadBalancer 负载均衡器
// 根据Worker负载情况分发任务
type LoadBalancer struct {
	rdb           *redis.Client
	workerLoadKey string
	mu            sync.RWMutex
	cache         map[string]*WorkerLoad // 本地缓存
	cacheExpiry   time.Time
	cacheTTL      time.Duration
}

// LoadBalancerConfig 负载均衡器配置
type LoadBalancerConfig struct {
	CacheTTL          time.Duration // 缓存过期时间
	HeartbeatTimeout  time.Duration // 心跳超时时间
	CPUThreshold      float64       // CPU阈值
	MemThreshold      float64       // 内存阈值
	TaskLoadWeight    float64       // 任务负载权重
	CPUWeight         float64       // CPU权重
	MemWeight         float64       // 内存权重
}

// DefaultLoadBalancerConfig 默认配置
func DefaultLoadBalancerConfig() *LoadBalancerConfig {
	return &LoadBalancerConfig{
		CacheTTL:         5 * time.Second,
		HeartbeatTimeout: 30 * time.Second,
		CPUThreshold:     90.0,
		MemThreshold:     90.0,
		TaskLoadWeight:   0.5,
		CPUWeight:        0.3,
		MemWeight:        0.2,
	}
}

// NewLoadBalancer 创建负载均衡器
func NewLoadBalancer(rdb *redis.Client) *LoadBalancer {
	return &LoadBalancer{
		rdb:           rdb,
		workerLoadKey: "cscan:worker:load",
		cache:         make(map[string]*WorkerLoad),
		cacheTTL:      5 * time.Second,
	}
}

// UpdateWorkerLoad 更新Worker负载信息
func (lb *LoadBalancer) UpdateWorkerLoad(ctx context.Context, load *WorkerLoad) error {
	load.LastHeartbeat = time.Now()
	data, err := json.Marshal(load)
	if err != nil {
		return err
	}

	// 更新Redis
	if err := lb.rdb.HSet(ctx, lb.workerLoadKey, load.WorkerName, data).Err(); err != nil {
		return err
	}

	// 更新本地缓存
	lb.mu.Lock()
	lb.cache[load.WorkerName] = load
	lb.mu.Unlock()

	return nil
}

// GetWorkerLoads 获取所有Worker负载信息
func (lb *LoadBalancer) GetWorkerLoads(ctx context.Context) ([]*WorkerLoad, error) {
	// 检查缓存是否有效
	lb.mu.RLock()
	if time.Now().Before(lb.cacheExpiry) && len(lb.cache) > 0 {
		loads := make([]*WorkerLoad, 0, len(lb.cache))
		for _, load := range lb.cache {
			loads = append(loads, load)
		}
		lb.mu.RUnlock()
		return loads, nil
	}
	lb.mu.RUnlock()

	// 从Redis获取
	data, err := lb.rdb.HGetAll(ctx, lb.workerLoadKey).Result()
	if err != nil {
		return nil, err
	}

	loads := make([]*WorkerLoad, 0, len(data))
	newCache := make(map[string]*WorkerLoad)

	for _, v := range data {
		var load WorkerLoad
		if err := json.Unmarshal([]byte(v), &load); err != nil {
			continue
		}
		loads = append(loads, &load)
		newCache[load.WorkerName] = &load
	}

	// 更新缓存
	lb.mu.Lock()
	lb.cache = newCache
	lb.cacheExpiry = time.Now().Add(lb.cacheTTL)
	lb.mu.Unlock()

	return loads, nil
}

// GetAvailableWorkers 获取可用的Worker列表（按负载排序）
func (lb *LoadBalancer) GetAvailableWorkers(ctx context.Context, config *LoadBalancerConfig) ([]*WorkerLoad, error) {
	if config == nil {
		config = DefaultLoadBalancerConfig()
	}

	loads, err := lb.GetWorkerLoads(ctx)
	if err != nil {
		return nil, err
	}

	// 过滤可用的Worker
	available := make([]*WorkerLoad, 0)
	now := time.Now()

	for _, load := range loads {
		// 检查心跳
		if now.Sub(load.LastHeartbeat) > config.HeartbeatTimeout {
			continue
		}
		// 检查任务槽位
		if load.CurrentTasks >= load.MaxConcurrency {
			continue
		}
		// 检查CPU
		if load.CPUPercent > config.CPUThreshold {
			continue
		}
		// 检查内存
		if load.MemPercent > config.MemThreshold {
			continue
		}
		available = append(available, load)
	}

	// 按负载分数排序（升序，负载低的在前）
	sort.Slice(available, func(i, j int) bool {
		scoreI := lb.calculateLoadScore(available[i], config)
		scoreJ := lb.calculateLoadScore(available[j], config)
		return scoreI < scoreJ
	})

	return available, nil
}

// calculateLoadScore 计算负载分数（越低越好）
func (lb *LoadBalancer) calculateLoadScore(load *WorkerLoad, config *LoadBalancerConfig) float64 {
	if load.MaxConcurrency == 0 {
		return 100.0
	}

	taskLoad := float64(load.CurrentTasks) / float64(load.MaxConcurrency) * 100
	return taskLoad*config.TaskLoadWeight +
		load.CPUPercent*config.CPUWeight +
		load.MemPercent*config.MemWeight
}

// SelectBestWorker 选择最佳Worker
func (lb *LoadBalancer) SelectBestWorker(ctx context.Context) (*WorkerLoad, error) {
	workers, err := lb.GetAvailableWorkers(ctx, nil)
	if err != nil {
		return nil, err
	}
	if len(workers) == 0 {
		return nil, nil
	}
	return workers[0], nil
}

// SelectWorkersForTask 为任务选择多个Worker（用于分布式任务）
func (lb *LoadBalancer) SelectWorkersForTask(ctx context.Context, count int) ([]*WorkerLoad, error) {
	workers, err := lb.GetAvailableWorkers(ctx, nil)
	if err != nil {
		return nil, err
	}

	if len(workers) < count {
		return workers, nil
	}
	return workers[:count], nil
}

// DistributeTask 分发任务到最佳Worker
// 返回选中的Worker名称
func (lb *LoadBalancer) DistributeTask(ctx context.Context, scheduler *Scheduler, task *TaskInfo) (string, error) {
	// 如果任务已指定Worker，直接使用
	if len(task.Workers) > 0 {
		return task.Workers[0], scheduler.PushTask(ctx, task)
	}

	// 选择最佳Worker
	worker, err := lb.SelectBestWorker(ctx)
	if err != nil {
		return "", err
	}

	// 如果没有可用Worker，推送到公共队列
	if worker == nil {
		return "", scheduler.PushTask(ctx, task)
	}

	// 推送到选中Worker的专属队列
	task.Workers = []string{worker.WorkerName}
	if err := scheduler.PushTask(ctx, task); err != nil {
		return "", err
	}

	return worker.WorkerName, nil
}

// DistributeTaskBatch 批量分发任务（负载均衡）
func (lb *LoadBalancer) DistributeTaskBatch(ctx context.Context, scheduler *Scheduler, tasks []*TaskInfo) error {
	if len(tasks) == 0 {
		return nil
	}

	// 获取可用Worker
	workers, err := lb.GetAvailableWorkers(ctx, nil)
	if err != nil {
		// 如果获取失败，推送到公共队列
		return scheduler.PushTaskBatch(ctx, tasks)
	}

	if len(workers) == 0 {
		// 没有可用Worker，推送到公共队列
		return scheduler.PushTaskBatch(ctx, tasks)
	}

	// 按负载均衡分配任务
	workerIndex := 0
	for _, task := range tasks {
		// 如果任务已指定Worker，跳过
		if len(task.Workers) > 0 {
			continue
		}

		// 轮询分配给可用Worker
		task.Workers = []string{workers[workerIndex].WorkerName}
		workerIndex = (workerIndex + 1) % len(workers)
	}

	return scheduler.PushTaskBatch(ctx, tasks)
}

// RemoveWorker 移除Worker（下线时调用）
func (lb *LoadBalancer) RemoveWorker(ctx context.Context, workerName string) error {
	// 从Redis删除
	if err := lb.rdb.HDel(ctx, lb.workerLoadKey, workerName).Err(); err != nil {
		return err
	}

	// 从缓存删除
	lb.mu.Lock()
	delete(lb.cache, workerName)
	lb.mu.Unlock()

	return nil
}

// GetWorkerStats 获取Worker统计信息
func (lb *LoadBalancer) GetWorkerStats(ctx context.Context) map[string]interface{} {
	loads, _ := lb.GetWorkerLoads(ctx)

	totalWorkers := len(loads)
	availableWorkers := 0
	totalCapacity := 0
	usedCapacity := 0
	avgCPU := 0.0
	avgMem := 0.0

	for _, load := range loads {
		if load.IsAvailable() {
			availableWorkers++
		}
		totalCapacity += load.MaxConcurrency
		usedCapacity += load.CurrentTasks
		avgCPU += load.CPUPercent
		avgMem += load.MemPercent
	}

	if totalWorkers > 0 {
		avgCPU /= float64(totalWorkers)
		avgMem /= float64(totalWorkers)
	}

	return map[string]interface{}{
		"totalWorkers":     totalWorkers,
		"availableWorkers": availableWorkers,
		"totalCapacity":    totalCapacity,
		"usedCapacity":     usedCapacity,
		"avgCPUPercent":    avgCPU,
		"avgMemPercent":    avgMem,
	}
}
