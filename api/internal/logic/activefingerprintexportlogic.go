package logic

import (
	"context"
	"sort"
	"strings"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"

	"gopkg.in/yaml.v3"
)

type ActiveFingerprintExportLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActiveFingerprintExportLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActiveFingerprintExportLogic {
	return &ActiveFingerprintExportLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActiveFingerprintExportLogic) ActiveFingerprintExport() (*types.ActiveFingerprintExportResp, error) {
	// 获取所有主动指纹
	docs, err := l.svcCtx.ActiveFingerprintModel.FindAll(l.ctx)
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return &types.ActiveFingerprintExportResp{
			Code:    0,
			Msg:     "没有可导出的指纹",
			Content: "",
		}, nil
	}

	// 转换为dir.yaml格式
	dirYaml := make(map[string][]string)
	for _, doc := range docs {
		dirYaml[doc.Name] = doc.Paths
	}

	// 按名称排序输出
	names := make([]string, 0, len(dirYaml))
	for name := range dirYaml {
		names = append(names, name)
	}
	sort.Strings(names)

	// 手动构建YAML以保持格式
	var sb strings.Builder
	for _, name := range names {
		paths := dirYaml[name]
		sb.WriteString(name + ":\n")
		for _, path := range paths {
			sb.WriteString("  - \"" + path + "\"\n")
		}
	}

	// 如果手动构建失败，使用yaml库
	content := sb.String()
	if content == "" {
		yamlBytes, err := yaml.Marshal(dirYaml)
		if err != nil {
			return nil, err
		}
		content = string(yamlBytes)
	}

	return &types.ActiveFingerprintExportResp{
		Code:    0,
		Msg:     "导出成功",
		Content: content,
	}, nil
}
