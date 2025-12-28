package logic

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RequestResourceLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRequestResourceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RequestResourceLogic {
	return &RequestResourceLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 请求资源文件
func (l *RequestResourceLogic) RequestResource(in *pb.RequestResourceReq) (*pb.RequestResourceResp, error) {
	category := in.Category
	name := in.Name

	if category == "" || name == "" {
		return &pb.RequestResourceResp{
			Path: "",
			Hash: "",
			Data: nil,
		}, nil
	}

	// 构建资源文件路径
	// 资源文件存放在 resources/<category>/<name>
	resourcePath := filepath.Join("resources", category, name)

	// 检查文件是否存在
	if _, err := os.Stat(resourcePath); os.IsNotExist(err) {
		l.Logger.Infof("RequestResource: resource not found: %s", resourcePath)
		return &pb.RequestResourceResp{
			Path: "",
			Hash: "",
			Data: nil,
		}, nil
	}

	// 读取文件内容
	data, err := os.ReadFile(resourcePath)
	if err != nil {
		l.Logger.Errorf("RequestResource: failed to read resource: %v", err)
		return &pb.RequestResourceResp{
			Path: "",
			Hash: "",
			Data: nil,
		}, nil
	}

	// 计算MD5哈希
	hash := md5.Sum(data)
	hashStr := hex.EncodeToString(hash[:])

	return &pb.RequestResourceResp{
		Path: resourcePath,
		Hash: hashStr,
		Data: data,
	}, nil
}
