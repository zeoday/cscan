package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ValidateFingerprintLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewValidateFingerprintLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ValidateFingerprintLogic {
	return &ValidateFingerprintLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 验证指纹匹配 - 此功能由Worker本地执行，RPC仅作为备用接口
func (l *ValidateFingerprintLogic) ValidateFingerprint(in *pb.ValidateFingerprintReq) (*pb.ValidateFingerprintResp, error) {
	// 指纹验证通常由Worker本地执行
	// 此RPC接口作为备用，返回未实现提示
	return &pb.ValidateFingerprintResp{
		Success: false,
		Message: "指纹验证请使用Worker本地执行",
		Matched: false,
	}, nil
}
