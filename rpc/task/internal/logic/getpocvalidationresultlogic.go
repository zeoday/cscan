package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPocValidationResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPocValidationResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPocValidationResultLogic {
	return &GetPocValidationResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询POC验证结果
func (l *GetPocValidationResultLogic) GetPocValidationResult(in *pb.GetPocValidationResultReq) (*pb.GetPocValidationResultResp, error) {
	taskId := in.TaskId
	if taskId == "" {
		return &pb.GetPocValidationResultResp{
			Success: false,
			Message: "TaskId不能为空",
		}, nil
	}

	// 从Redis获取验证结果
	resultKey := "cscan:poc:validation:" + taskId
	_, err := l.svcCtx.RedisClient.Get(l.ctx, resultKey).Result()
	if err != nil {
		return &pb.GetPocValidationResultResp{
			Success: false,
			Message: "未找到验证结果",
			Status:  "NOT_FOUND",
		}, nil
	}

	// 返回基本状态（实际结果需要根据具体存储格式解析）
	return &pb.GetPocValidationResultResp{
		Success: true,
		Message: "success",
		Status:  "COMPLETED",
	}, nil
}
