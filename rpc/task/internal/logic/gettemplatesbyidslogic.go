package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetTemplatesByIdsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetTemplatesByIdsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTemplatesByIdsLogic {
	return &GetTemplatesByIdsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据ID列表批量获取模板内容
func (l *GetTemplatesByIdsLogic) GetTemplatesByIds(in *pb.GetTemplatesByIdsReq) (*pb.GetTemplatesByIdsResp, error) {
	var templates []string

	// 获取Nuclei默认模板
	if len(in.NucleiTemplateIds) > 0 {
		nucleiTemplates, err := l.svcCtx.NucleiTemplateModel.FindByIds(l.ctx, in.NucleiTemplateIds)
		if err != nil {
			l.Logger.Errorf("FindByIds for nuclei templates failed: %v", err)
			return &pb.GetTemplatesByIdsResp{
				Success: false,
				Message: "获取Nuclei模板失败: " + err.Error(),
			}, nil
		}
		for _, t := range nucleiTemplates {
			if t.Content != "" && t.Enabled {
				templates = append(templates, t.Content)
			}
		}
	}

	// 获取自定义POC
	if len(in.CustomPocIds) > 0 {
		customPocs, err := l.svcCtx.CustomPocModel.FindByIds(l.ctx, in.CustomPocIds)
		if err != nil {
			l.Logger.Errorf("FindByIds for custom pocs failed: %v", err)
			return &pb.GetTemplatesByIdsResp{
				Success: false,
				Message: "获取自定义POC失败: " + err.Error(),
			}, nil
		}
		for _, p := range customPocs {
			if p.Content != "" && p.Enabled {
				templates = append(templates, p.Content)
			}
		}
	}

	return &pb.GetTemplatesByIdsResp{
		Success:   true,
		Message:   "success",
		Templates: templates,
		Count:     int32(len(templates)),
	}, nil
}
