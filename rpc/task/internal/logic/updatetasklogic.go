package logic

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"
	"cscan/scheduler"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
)

type UpdateTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTaskLogic {
	return &UpdateTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新任务状态
func (l *UpdateTaskLogic) UpdateTask(in *pb.UpdateTaskReq) (*pb.UpdateTaskResp, error) {
	taskId := in.TaskId
	state := in.State

	l.Logger.Infof("UpdateTask: taskId=%s, state=%s", taskId, state)

	// 从处理中集合移除
	processingKey := "cscan:task:processing"
	l.svcCtx.RedisClient.SRem(l.ctx, processingKey, taskId)

	// 更新任务状态到Redis
	statusKey := "cscan:task:status:" + taskId
	statusData := map[string]interface{}{
		"taskId": taskId,
		"state":  state,
		"worker": in.Worker,
		"result": in.Result,
	}
	statusJson, _ := json.Marshal(statusData)
	l.svcCtx.RedisClient.Set(l.ctx, statusKey, statusJson, 0)

	// 如果任务完成或失败，添加到完成集合
	if state == "SUCCESS" || state == "FAILURE" || state == "COMPLETED" {
		completedKey := "cscan:task:completed"
		taskInfo := scheduler.TaskInfo{
			TaskId: taskId,
		}
		taskJson, _ := json.Marshal(taskInfo)
		l.svcCtx.RedisClient.SAdd(l.ctx, completedKey, string(taskJson))
	}

	// 更新数据库中的任务状态（包括开始时间、结束时间、进度）
	l.updateTaskInDB(taskId, state, in.Result)

	return &pb.UpdateTaskResp{
		Success: true,
		Message: "Task status updated",
	}, nil
}

// updateTaskInDB 更新数据库中的任务状态
func (l *UpdateTaskLogic) updateTaskInDB(taskId, state, result string) {
	// 从Redis获取任务信息（workspaceId）
	taskInfoKey := "cscan:task:info:" + taskId
	taskInfoData, err := l.svcCtx.RedisClient.Get(l.ctx, taskInfoKey).Result()
	if err != nil {
		l.Logger.Errorf("UpdateTask: failed to get task info from Redis, taskId=%s, error=%v", taskId, err)
		return
	}

	var taskInfo map[string]interface{}
	if err := json.Unmarshal([]byte(taskInfoData), &taskInfo); err != nil {
		l.Logger.Errorf("UpdateTask: failed to parse task info, taskId=%s, error=%v", taskId, err)
		return
	}

	workspaceId, _ := taskInfo["workspaceId"].(string)
	mainTaskId, _ := taskInfo["mainTaskId"].(string) // MongoDB ObjectID (Hex)
	subTaskCount := 1
	if count, ok := taskInfo["subTaskCount"].(float64); ok {
		subTaskCount = int(count)
	}
	if workspaceId == "" {
		l.Logger.Errorf("UpdateTask: workspaceId is empty, taskId=%s", taskId)
		return
	}

	// 获取任务模型
	taskModel := l.svcCtx.GetMainTaskModel(workspaceId)
	now := time.Now()

	// 构建更新字段
	update := bson.M{
		"status": state,
	}

	// 判断是否是子任务：
	// 子任务ID格式: {parentTaskId}-{index}，其中 parentTaskId 是 UUID
	// 单任务ID格式: {taskId}，就是 UUID 本身
	// 判断方法：检查 taskId 最后一个 "-" 后面是否是纯数字
	isSubTask := false
	if subTaskCount > 1 {
		lastDash := strings.LastIndex(taskId, "-")
		if lastDash > 0 && lastDash < len(taskId)-1 {
			suffix := taskId[lastDash+1:]
			// 检查后缀是否全是数字
			isNumber := true
			for _, c := range suffix {
				if c < '0' || c > '9' {
					isNumber = false
					break
				}
			}
			isSubTask = isNumber
		}
	}

	l.Logger.Infof("UpdateTask: taskId=%s, mainTaskId=%s, subTaskCount=%d, isSubTask=%v", taskId, mainTaskId, subTaskCount, isSubTask)

	// 根据状态设置不同字段
	switch state {
	case "STARTED":
		// 任务开始时设置开始时间
		update["start_time"] = now
		update["progress"] = 10 // 开始时进度设为10%
	case "SUCCESS", "COMPLETED":
		if isSubTask {
			// 子任务完成，递增 sub_task_done
			// 使用 $inc 操作符，mainTaskId 是 MongoDB ObjectID
			if err := taskModel.IncrSubTaskDone(l.ctx, mainTaskId); err != nil {
				l.Logger.Errorf("UpdateTask: failed to incr sub_task_done, mainTaskId=%s, error=%v", mainTaskId, err)
			} else {
				l.Logger.Infof("UpdateTask: sub_task_done incremented, mainTaskId=%s, taskId=%s", mainTaskId, taskId)
			}
			// 检查是否所有子任务都完成了
			// 使用 FindById 而不是 FindByTaskId，因为 mainTaskId 是 MongoDB ObjectID
			task, err := taskModel.FindById(l.ctx, mainTaskId)
			if err == nil && task != nil && task.SubTaskDone+1 >= task.SubTaskCount {
				// 所有子任务完成，更新主任务状态
				update["end_time"] = now
				update["progress"] = 100
				update["result"] = result
				l.Logger.Infof("UpdateTask: all sub-tasks completed, mainTaskId=%s, done=%d, total=%d", mainTaskId, task.SubTaskDone+1, task.SubTaskCount)
			} else {
				// 还有子任务未完成，只更新 sub_task_done，不更新主任务状态
				if err != nil {
					l.Logger.Errorf("UpdateTask: failed to find task by id, mainTaskId=%s, error=%v", mainTaskId, err)
				} else if task != nil {
					l.Logger.Infof("UpdateTask: sub-task completed, mainTaskId=%s, done=%d, total=%d", mainTaskId, task.SubTaskDone+1, task.SubTaskCount)
				}
				return
			}
		} else {
			// 单任务或最后一个子任务完成
			update["end_time"] = now
			update["progress"] = 100
			update["result"] = result
		}
	case "FAILURE":
		// 任务失败时设置结束时间
		update["end_time"] = now
		update["result"] = result
	case "STOPPED":
		// 任务停止时设置结束时间
		update["end_time"] = now
		update["result"] = "任务已停止"
	}

	// 更新数据库，mainTaskId 是 MongoDB ObjectID
	if mainTaskId != "" {
		if err := taskModel.Update(l.ctx, mainTaskId, update); err != nil {
			l.Logger.Errorf("UpdateTask: failed to update task in DB, mainTaskId=%s, error=%v", mainTaskId, err)
		} else {
			l.Logger.Infof("UpdateTask: task updated in DB, mainTaskId=%s, state=%s", mainTaskId, state)
		}
	}
}
