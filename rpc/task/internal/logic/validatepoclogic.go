package logic

import (
	"context"
	"encoding/json"
	"time"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"
	"cscan/scheduler"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type ValidatePocLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidatePocLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidatePocLogic {
	return &ValidatePocLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// POC验证 - 创建验证任务并推送到队列，由Worker执行
func (l *ValidatePocLogic) ValidatePoc(in *pb.ValidatePocReq) (*pb.ValidatePocResp, error) {
	// 生成任务ID
	taskId := uuid.New().String()

	// 获取workspaceId，如果未指定则使用default
	workspaceId := in.WorkspaceId
	if workspaceId == "" {
		workspaceId = "default"
	}

	// 判断是批量模式还是单目标模式
	taskType := "poc_validate"
	var targetUrls []string
	if in.BatchMode && len(in.Urls) > 0 {
		taskType = "poc_batch_validate"
		targetUrls = in.Urls
	} else if in.Url != "" {
		targetUrls = []string{in.Url}
	}

	// 构建任务配置
	taskConfig := map[string]interface{}{
		"taskType":    taskType,
		"urls":        targetUrls,
		"pocId":       in.PocId,
		"pocType":     in.PocType,
		"timeout":     in.Timeout,
		"workspaceId": workspaceId,
		"batchMode":   in.BatchMode,
	}
	// 兼容单目标模式
	if len(targetUrls) == 1 {
		taskConfig["url"] = targetUrls[0]
	}
	configBytes, _ := json.Marshal(taskConfig)

	// 创建任务信息
	taskName := "POC验证"
	if in.BatchMode {
		taskName = "POC批量扫描"
	}
	task := &scheduler.TaskInfo{
		TaskId:      taskId,
		MainTaskId:  taskId,
		WorkspaceId: workspaceId,
		TaskName:    taskName,
		Config:      string(configBytes),
		Priority:    2, // 高优先级
	}

	// 推送任务到队列（使用 Sorted Set，时间戳作为分数实现 FIFO）
	taskJson, _ := json.Marshal(task)
	queueKey := "cscan:task:queue"
	score := float64(time.Now().UnixNano())
	err := l.svcCtx.RedisClient.ZAdd(l.ctx, queueKey, redis.Z{
		Score:  score,
		Member: taskJson,
	}).Err()
	if err != nil {
		// 如果是类型错误，尝试删除旧 key 后重试
		if err.Error() == "WRONGTYPE Operation against a key holding the wrong kind of value" {
			l.svcCtx.RedisClient.Del(l.ctx, queueKey)
			err = l.svcCtx.RedisClient.ZAdd(l.ctx, queueKey, redis.Z{
				Score:  score,
				Member: taskJson,
			}).Err()
		}
		if err != nil {
			l.Logger.Errorf("ValidatePoc: failed to push task to queue, error=%v", err)
			return &pb.ValidatePocResp{
				Success: false,
				Message: "任务入队失败: " + err.Error(),
				Matched: false,
			}, nil
		}
	}

	// 保存任务信息到Redis（用于结果查询）
	taskInfoKey := "cscan:task:info:" + taskId
	taskInfoData, _ := json.Marshal(map[string]interface{}{
		"workspaceId": workspaceId,
		"mainTaskId":  taskId,
		"taskType":    taskType,
		"urls":        targetUrls,
		"pocId":       in.PocId,
		"pocType":     in.PocType,
		"batchMode":   in.BatchMode,
		"createTime":  time.Now().Local().Format("2006-01-02 15:04:05"),
	})
	l.svcCtx.RedisClient.Set(l.ctx, taskInfoKey, taskInfoData, 24*time.Hour)

	l.Logger.Infof("ValidatePoc: task created, taskId=%s, targets=%d, pocId=%s, workspaceId=%s, batchMode=%v", 
		taskId, len(targetUrls), in.PocId, workspaceId, in.BatchMode)

	return &pb.ValidatePocResp{
		Success: true,
		Message: "POC验证任务已下发",
		Matched: false,
		TaskId:  taskId,
	}, nil
}
