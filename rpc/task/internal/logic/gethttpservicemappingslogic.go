package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHttpServiceMappingsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHttpServiceMappingsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHttpServiceMappingsLogic {
	return &GetHttpServiceMappingsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取HTTP服务映射
func (l *GetHttpServiceMappingsLogic) GetHttpServiceMappings(in *pb.GetHttpServiceMappingsReq) (*pb.GetHttpServiceMappingsResp, error) {
	// 获取所有启用的HTTP服务映射
	mappings, err := l.svcCtx.HttpServiceMappingModel.FindEnabled(l.ctx)
	if err != nil {
		l.Logger.Errorf("FindEnabled for http service mappings failed: %v", err)
		return &pb.GetHttpServiceMappingsResp{
			Success: false,
			Message: "获取HTTP服务映射失败: " + err.Error(),
		}, nil
	}

	// 转换为protobuf格式
	var pbMappings []*pb.HttpServiceMappingDocument
	for _, m := range mappings {
		pbMappings = append(pbMappings, &pb.HttpServiceMappingDocument{
			Id:          m.Id.Hex(),
			ServiceName: m.ServiceName,
			IsHttp:      m.IsHttp,
			Description: m.Description,
			Enabled:     m.Enabled,
		})
	}

	return &pb.GetHttpServiceMappingsResp{
		Success:  true,
		Message:  "success",
		Mappings: pbMappings,
		Count:    int32(len(pbMappings)),
	}, nil
}
