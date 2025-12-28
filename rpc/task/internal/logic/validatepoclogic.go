package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

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

// POC验证 - 此功能由Worker本地执行，RPC仅作为备用接口
func (l *ValidatePocLogic) ValidatePoc(in *pb.ValidatePocReq) (*pb.ValidatePocResp, error) {
	// POC验证通常由Worker本地执行
	// 此RPC接口作为备用，返回未实现提示
	return &pb.ValidatePocResp{
		Success: false,
		Message: "POC验证请使用Worker本地执行",
		Matched: false,
	}, nil
}
