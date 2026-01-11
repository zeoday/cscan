package logic

import (
	"context"
	"time"

	"cscan/pkg/notify"
	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

type IncrSubTaskDoneLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrSubTaskDoneLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrSubTaskDoneLogic {
	return &IncrSubTaskDoneLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 递增子任务完成数（模块级别）
func (l *IncrSubTaskDoneLogic) IncrSubTaskDone(in *pb.IncrSubTaskDoneReq) (*pb.IncrSubTaskDoneResp, error) {
	l.Logger.Infof("IncrSubTaskDone: taskId=%s, mainTaskId=%s, phase=%s", in.TaskId, in.MainTaskId, in.Phase)

	if in.WorkspaceId == "" || in.MainTaskId == "" {
		return &pb.IncrSubTaskDoneResp{
			Success: false,
			Message: "workspaceId or mainTaskId is empty",
		}, nil
	}

	// 获取任务模型
	taskModel := l.svcCtx.GetMainTaskModel(in.WorkspaceId)

	// 递增 sub_task_done
	if err := taskModel.IncrSubTaskDone(l.ctx, in.MainTaskId); err != nil {
		l.Logger.Errorf("IncrSubTaskDone: failed to incr, mainTaskId=%s, error=%v", in.MainTaskId, err)
		return &pb.IncrSubTaskDoneResp{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// 获取最新的任务状态
	task, err := taskModel.FindById(l.ctx, in.MainTaskId)
	if err != nil {
		l.Logger.Errorf("IncrSubTaskDone: failed to find task, mainTaskId=%s, error=%v", in.MainTaskId, err)
		return &pb.IncrSubTaskDoneResp{
			Success: true,
			Message: "incremented but failed to get task status",
		}, nil
	}

	allDone := task.SubTaskDone >= task.SubTaskCount
	l.Logger.Infof("IncrSubTaskDone: mainTaskId=%s, phase=%s, done=%d, total=%d, allDone=%v",
		in.MainTaskId, in.Phase, task.SubTaskDone, task.SubTaskCount, allDone)

	// 计算进度百分比
	progress := 0
	if task.SubTaskCount > 0 {
		progress = task.SubTaskDone * 100 / task.SubTaskCount
	}

	// 更新进度到数据库
	update := bson.M{
		"progress":      progress,
		"current_phase": in.Phase,
	}

	// 如果全部完成，更新状态
	if allDone {
		update["status"] = "SUCCESS"
		update["progress"] = 100
		update["end_time"] = time.Now()
	}

	if err := taskModel.Update(l.ctx, in.MainTaskId, update); err != nil {
		l.Logger.Errorf("IncrSubTaskDone: failed to update progress, mainTaskId=%s, error=%v", in.MainTaskId, err)
	}

	// 如果全部完成，发送通知
	if allDone {
		l.sendTaskNotification(in.WorkspaceId, in.MainTaskId, "SUCCESS")
	}

	return &pb.IncrSubTaskDoneResp{
		Success:      true,
		Message:      "ok",
		SubTaskDone:  int32(task.SubTaskDone),
		SubTaskCount: int32(task.SubTaskCount),
		AllDone:      allDone,
	}, nil
}

// sendTaskNotification 发送任务完成通知
func (l *IncrSubTaskDoneLogic) sendTaskNotification(workspaceId, mainTaskId, status string) {
	// 获取任务详情
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)
	task, err := taskModel.FindById(l.ctx, mainTaskId)
	if err != nil {
		l.Logger.Errorf("sendTaskNotification: failed to get task, mainTaskId=%s, error=%v", mainTaskId, err)
		return
	}

	// 获取资产和漏洞统计
	assetModel := l.svcCtx.GetAssetModel(workspaceId)
	vulModel := l.svcCtx.GetVulModel(workspaceId)

	assetCount, _ := assetModel.CountByTaskId(l.ctx, mainTaskId)
	vulCount, _ := vulModel.CountByTaskId(l.ctx, mainTaskId)

	// 获取启用的通知配置
	configs, err := l.svcCtx.NotifyConfigModel.FindEnabled(l.ctx)
	if err != nil {
		l.Logger.Errorf("sendTaskNotification: failed to get notify configs, error=%v", err)
		return
	}

	if len(configs) == 0 {
		l.Logger.Infof("sendTaskNotification: no enabled notify configs")
		return
	}

	// 构建通知配置列表
	var configItems []notify.ConfigItem
	for _, c := range configs {
		configItems = append(configItems, notify.ConfigItem{
			Provider:        c.Provider,
			Config:          c.Config,
			MessageTemplate: c.MessageTemplate,
		})
	}

	// 构建通知结果
	result := &notify.NotifyResult{
		TaskId:      mainTaskId,
		TaskName:    task.Name,
		Status:      status,
		AssetCount:  int(assetCount),
		VulCount:    int(vulCount),
		WorkspaceId: workspaceId,
	}

	// 设置时间（处理指针类型）
	if task.StartTime != nil {
		result.StartTime = *task.StartTime
	}
	if task.EndTime != nil {
		result.EndTime = *task.EndTime
	}

	// 计算耗时
	if task.StartTime != nil && task.EndTime != nil {
		d := task.EndTime.Sub(*task.StartTime)
		if d.Hours() >= 1 {
			result.Duration = d.Round(time.Minute).String()
		} else if d.Minutes() >= 1 {
			result.Duration = d.Round(time.Second).String()
		} else {
			result.Duration = d.Round(time.Millisecond).String()
		}
	}

	// 异步发送通知
	notify.SendNotificationAsync(l.ctx, configItems, result)
	l.Logger.Infof("sendTaskNotification: notification queued for task %s, status=%s", mainTaskId, status)
}
