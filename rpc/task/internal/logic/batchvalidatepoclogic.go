package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchValidatePocLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchValidatePocLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchValidatePocLogic {
	return &BatchValidatePocLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量POC验证 - 此功能由Worker本地执行，RPC仅作为备用接口
func (l *BatchValidatePocLogic) BatchValidatePoc(in *pb.BatchValidatePocReq) (*pb.BatchValidatePocResp, error) {
	// 批量POC验证通常由Worker本地执行
	// 此RPC接口作为备用，返回未实现提示
	return &pb.BatchValidatePocResp{
		Success: false,
		Message: "批量POC验证请使用Worker本地执行",
	}, nil
}
