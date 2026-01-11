package notify

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// NotifyService 通知服务（用于API层调用）
type NotifyService struct {
	manager *NotifyManager
}

// NewNotifyService 创建通知服务
func NewNotifyService() *NotifyService {
	return &NotifyService{
		manager: NewNotifyManager(),
	}
}

// LoadConfigsFromDB 从数据库配置加载提供者
// configs 是从数据库读取的配置列表
func (s *NotifyService) LoadConfigsFromDB(configs []ConfigItem) error {
	return s.manager.LoadConfigs(configs)
}

// SendTaskNotification 发送任务完成通知
func (s *NotifyService) SendTaskNotification(ctx context.Context, result *NotifyResult) error {
	return s.manager.Send(ctx, result)
}

// TaskCompleteInfo 任务完成信息（用于从Redis或数据库获取）
type TaskCompleteInfo struct {
	TaskId      string    `json:"taskId"`
	TaskName    string    `json:"taskName"`
	Status      string    `json:"status"`
	AssetCount  int       `json:"assetCount"`
	VulCount    int       `json:"vulCount"`
	StartTime   time.Time `json:"startTime"`
	EndTime     time.Time `json:"endTime"`
	WorkspaceId string    `json:"workspaceId"`
}

// BuildNotifyResult 从任务完成信息构建通知结果
func BuildNotifyResult(info *TaskCompleteInfo) *NotifyResult {
	duration := ""
	if !info.StartTime.IsZero() && !info.EndTime.IsZero() {
		d := info.EndTime.Sub(info.StartTime)
		if d.Hours() >= 1 {
			duration = d.Round(time.Minute).String()
		} else if d.Minutes() >= 1 {
			duration = d.Round(time.Second).String()
		} else {
			duration = d.Round(time.Millisecond).String()
		}
	}

	return &NotifyResult{
		TaskId:      info.TaskId,
		TaskName:    info.TaskName,
		Status:      info.Status,
		AssetCount:  info.AssetCount,
		VulCount:    info.VulCount,
		Duration:    duration,
		StartTime:   info.StartTime,
		EndTime:     info.EndTime,
		WorkspaceId: info.WorkspaceId,
	}
}

// SendNotificationAsync 异步发送通知（不阻塞主流程）
func SendNotificationAsync(ctx context.Context, configs []ConfigItem, result *NotifyResult) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logx.Errorf("SendNotificationAsync panic: %v", r)
			}
		}()

		// 创建独立的context，不受父context取消影响
		notifyCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		manager := NewNotifyManager()
		if err := manager.LoadConfigs(configs); err != nil {
			logx.Errorf("Load notify configs failed: %v", err)
			return
		}

		if err := manager.Send(notifyCtx, result); err != nil {
			logx.Errorf("Send notification failed: %v", err)
		} else {
			logx.Infof("Task notification sent: taskId=%s, status=%s", result.TaskId, result.Status)
		}
	}()
}

// ParseConfigJSON 解析配置JSON
func ParseConfigJSON(configJSON string) (map[string]interface{}, error) {
	var config map[string]interface{}
	err := json.Unmarshal([]byte(configJSON), &config)
	return config, err
}
