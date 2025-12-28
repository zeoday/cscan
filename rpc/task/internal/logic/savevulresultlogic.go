package logic

import (
	"context"

	"cscan/model"
	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveVulResultLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveVulResultLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveVulResultLogic {
	return &SaveVulResultLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存漏洞结果
func (l *SaveVulResultLogic) SaveVulResult(in *pb.SaveVulResultReq) (*pb.SaveVulResultResp, error) {
	if len(in.Vuls) == 0 {
		return &pb.SaveVulResultResp{
			Success: true,
			Message: "No vulnerabilities to save",
			Total:   0,
		}, nil
	}

	workspaceId := in.WorkspaceId
	if workspaceId == "" {
		workspaceId = "default"
	}

	vulModel := l.svcCtx.GetVulModel(workspaceId)
	var savedCount int32

	for _, pbVul := range in.Vuls {
		vul := &model.Vul{
			Authority: pbVul.Authority,
			Host:      pbVul.Host,
			Port:      int(pbVul.Port),
			Url:       pbVul.Url,
			PocFile:   pbVul.PocFile,
			Source:    pbVul.Source,
			Severity:  pbVul.Severity,
			Extra:     pbVul.Extra,
			Result:    pbVul.Result,
			TaskId:    in.MainTaskId,
		}

		// 漏洞知识库关联字段
		if pbVul.CvssScore != nil {
			vul.CvssScore = *pbVul.CvssScore
		}
		if pbVul.CveId != nil {
			vul.CveId = *pbVul.CveId
		}
		if pbVul.CweId != nil {
			vul.CweId = *pbVul.CweId
		}
		if pbVul.Remediation != nil {
			vul.Remediation = *pbVul.Remediation
		}
		if len(pbVul.References) > 0 {
			vul.References = pbVul.References
		}

		// 证据链字段
		if pbVul.MatcherName != nil {
			vul.MatcherName = *pbVul.MatcherName
		}
		if len(pbVul.ExtractedResults) > 0 {
			vul.ExtractedResults = pbVul.ExtractedResults
		}
		if pbVul.CurlCommand != nil {
			vul.CurlCommand = *pbVul.CurlCommand
		}
		if pbVul.Request != nil {
			vul.Request = *pbVul.Request
		}
		if pbVul.Response != nil {
			vul.Response = *pbVul.Response
		}
		if pbVul.ResponseTruncated != nil {
			vul.ResponseTruncated = *pbVul.ResponseTruncated
		}

		// 使用Upsert避免重复
		if err := vulModel.Upsert(l.ctx, vul); err != nil {
			l.Logger.Errorf("SaveVulResult: failed to upsert vul: %v", err)
			continue
		}
		savedCount++
	}

	l.Logger.Infof("SaveVulResult: saved %d vulnerabilities", savedCount)

	return &pb.SaveVulResultResp{
		Success: true,
		Message: "Vulnerabilities saved successfully",
		Total:   savedCount,
	}, nil
}
