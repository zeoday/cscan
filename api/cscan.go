package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cscan/api/internal/config"
	"cscan/api/internal/handler"
	"cscan/api/internal/svc"
	"cscan/model"
	"cscan/scheduler"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/cscan.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	svcCtx := svc.NewServiceContext(c)

	// 创建HTTP服务器
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, svcCtx)

	// 创建任务调度器服务
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Redis.Host,
		Password: c.Redis.Pass,
	})
	schedulerSvc := scheduler.NewSchedulerService(rdb, svcCtx.SyncMethods)
	go schedulerSvc.Start()

	// 启动定时任务执行消息订阅
	go startCronExecuteSubscriber(svcCtx, schedulerSvc.GetScheduler())

	fmt.Printf("Starting API server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// CronExecuteMessage 定时任务执行消息
type CronExecuteMessage struct {
	CronTaskId  string `json:"cronTaskId"`
	WorkspaceId string `json:"workspaceId"`
	MainTaskId  string `json:"mainTaskId"`
	TaskName    string `json:"taskName"`
	Target      string `json:"target"`
	Config      string `json:"config"`
}

// startCronExecuteSubscriber 启动定时任务执行消息订阅
func startCronExecuteSubscriber(svcCtx *svc.ServiceContext, sched *scheduler.Scheduler) {
	ctx := context.Background()
	pubsub := svcCtx.RedisClient.Subscribe(ctx, "cscan:cron:execute")
	defer pubsub.Close()

	logx.Info("Cron execute subscriber started")

	ch := pubsub.Channel()
	for msg := range ch {
		var execMsg CronExecuteMessage
		if err := json.Unmarshal([]byte(msg.Payload), &execMsg); err != nil {
			logx.Errorf("Failed to parse cron execute message: %v", err)
			continue
		}

		logx.Infof("Received cron execute message: cronTaskId=%s, taskName=%s", execMsg.CronTaskId, execMsg.TaskName)

		// 创建新的 MainTask 并推送到队列
		if err := createAndPushCronTask(ctx, svcCtx, sched, &execMsg); err != nil {
			logx.Errorf("Failed to create cron task: %v", err)
		}
	}
}

// createAndPushCronTask 创建定时任务的 MainTask 并推送到队列
func createAndPushCronTask(ctx context.Context, svcCtx *svc.ServiceContext, sched *scheduler.Scheduler, msg *CronExecuteMessage) error {
	workspaceId := msg.WorkspaceId
	if workspaceId == "" {
		workspaceId = "default"
	}

	// 解析任务配置
	var taskConfig map[string]interface{}
	if err := json.Unmarshal([]byte(msg.Config), &taskConfig); err != nil {
		return fmt.Errorf("failed to parse task config: %v", err)
	}

	// 生成新的任务ID
	newTaskId := uuid.New().String()

	// 创建新的 MainTask
	taskModel := svcCtx.GetMainTaskModel(workspaceId)
	newTask := &model.MainTask{
		TaskId:      newTaskId,
		Name:        fmt.Sprintf("%s (定时)", msg.TaskName),
		Target:      msg.Target,
		Config:      msg.Config,
		Status:      model.TaskStatusCreated,
		IsCron:      true,
		CronRule:    msg.CronTaskId,
	}

	if err := taskModel.Insert(ctx, newTask); err != nil {
		return fmt.Errorf("failed to insert main task: %v", err)
	}

	logx.Infof("Created cron main task: taskId=%s, name=%s", newTaskId, newTask.Name)

	// 计算子任务数量（基于目标数量和启用的模块数）
	targets := strings.Split(msg.Target, "\n")
	var validTargets []string
	for _, t := range targets {
		t = strings.TrimSpace(t)
		if t != "" {
			validTargets = append(validTargets, t)
		}
	}

	// 计算启用的模块数
	enabledModules := 0
	if ps, ok := taskConfig["portscan"].(map[string]interface{}); ok {
		if enable, _ := ps["enable"].(bool); enable {
			enabledModules++
		}
	}
	if ds, ok := taskConfig["domainscan"].(map[string]interface{}); ok {
		if enable, _ := ds["enable"].(bool); enable {
			enabledModules++
		}
	}
	if fp, ok := taskConfig["fingerprint"].(map[string]interface{}); ok {
		if enable, _ := fp["enable"].(bool); enable {
			enabledModules++
		}
	}
	if poc, ok := taskConfig["pocscan"].(map[string]interface{}); ok {
		if enable, _ := poc["enable"].(bool); enable {
			enabledModules++
		}
	}
	if dir, ok := taskConfig["dirscan"].(map[string]interface{}); ok {
		if enable, _ := dir["enable"].(bool); enable {
			enabledModules++
		}
	}
	if enabledModules == 0 {
		enabledModules = 1
	}

	// 分批处理目标
	batchSize := 50
	if bs, ok := taskConfig["batchSize"].(float64); ok && bs > 0 {
		batchSize = int(bs)
	}

	var batches []string
	for i := 0; i < len(validTargets); i += batchSize {
		end := i + batchSize
		if end > len(validTargets) {
			end = len(validTargets)
		}
		batches = append(batches, strings.Join(validTargets[i:end], "\n"))
	}
	if len(batches) == 0 {
		batches = []string{msg.Target}
	}

	subTaskCount := len(batches) * enabledModules

	// 更新任务状态为 STARTED
	now := time.Now()
	taskModel.Update(ctx, newTask.Id.Hex(), map[string]interface{}{
		"status":         model.TaskStatusStarted,
		"sub_task_count": subTaskCount,
		"sub_task_done":  0,
		"start_time":     now,
	})

	// 保存主任务信息到 Redis
	taskInfoKey := "cscan:task:info:" + newTaskId
	taskInfoData, _ := json.Marshal(map[string]interface{}{
		"workspaceId":    workspaceId,
		"mainTaskId":     newTask.Id.Hex(),
		"subTaskCount":   subTaskCount,
		"batchCount":     len(batches),
		"enabledModules": enabledModules,
	})
	svcCtx.RedisClient.Set(ctx, taskInfoKey, taskInfoData, 24*time.Hour)

	// 从配置中获取指定的 Worker 列表
	var workers []string
	if w, ok := taskConfig["workers"].([]interface{}); ok {
		for _, v := range w {
			if s, ok := v.(string); ok {
				workers = append(workers, s)
			}
		}
	}

	// 为每个批次创建子任务并推送到队列
	for i, batch := range batches {
		// 复制配置并替换目标
		subConfig := make(map[string]interface{})
		for k, v := range taskConfig {
			subConfig[k] = v
		}
		subConfig["target"] = batch
		subConfig["subTaskIndex"] = i
		subConfig["subTaskTotal"] = len(batches)
		subConfigBytes, _ := json.Marshal(subConfig)

		// 生成子任务ID
		subTaskId := newTaskId
		if len(batches) > 1 {
			subTaskId = newTaskId + "-" + strconv.Itoa(i)
		}

		schedTask := &scheduler.TaskInfo{
			TaskId:      subTaskId,
			MainTaskId:  newTask.Id.Hex(),
			WorkspaceId: workspaceId,
			TaskName:    newTask.Name,
			Config:      string(subConfigBytes),
			Priority:    0,
			Workers:     workers,
		}

		logx.Infof("Pushing cron sub-task %d/%d: taskId=%s, targets=%d", i+1, len(batches), subTaskId, len(strings.Split(batch, "\n")))

		if err := sched.PushTask(ctx, schedTask); err != nil {
			logx.Errorf("Failed to push cron task to queue: %v", err)
			continue
		}

		// 保存子任务信息到 Redis（多批次时）
		if len(batches) > 1 {
			subTaskInfoKey := "cscan:task:info:" + subTaskId
			subTaskInfoData, _ := json.Marshal(map[string]interface{}{
				"workspaceId":  workspaceId,
				"mainTaskId":   newTask.Id.Hex(),
				"subTaskCount": subTaskCount,
			})
			svcCtx.RedisClient.Set(ctx, subTaskInfoKey, subTaskInfoData, 24*time.Hour)
		}
	}

	logx.Infof("Cron task created and pushed: taskId=%s, batches=%d, subTaskCount=%d", newTaskId, len(batches), subTaskCount)
	return nil
}
