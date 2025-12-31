package taskservicelogic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubfinderProvidersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSubfinderProvidersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubfinderProvidersLogic {
	return &GetSubfinderProvidersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取Subfinder数据源配置
func (l *GetSubfinderProvidersLogic) GetSubfinderProviders(in *pb.GetSubfinderProvidersReq) (*pb.GetSubfinderProvidersResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetSubfinderProvidersResp{}, nil
}
