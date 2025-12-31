package taskservicelogic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type NewTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewNewTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *NewTaskLogic {
	return &NewTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建执行器任务
func (l *NewTaskLogic) NewTask(in *pb.NewTaskReq) (*pb.NewTaskResp, error) {
	// todo: add your logic here and delete this line

	return &pb.NewTaskResp{}, nil
}
