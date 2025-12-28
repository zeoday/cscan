package logic

import (
	"context"

	"cscan/rpc/task/internal/svc"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetWorkerConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetWorkerConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetWorkerConfigLogic {
	return &GetWorkerConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取Worker配置
func (l *GetWorkerConfigLogic) GetWorkerConfig(in *pb.GetWorkerConfigReq) (*pb.GetWorkerConfigResp, error) {
	workerName := in.WorkerName

	// 从Redis获取Worker配置
	configKey := "cscan:worker:config:" + workerName
	config, err := l.svcCtx.RedisClient.Get(l.ctx, configKey).Result()
	if err != nil {
		// 如果没有特定配置，返回空配置
		l.Logger.Infof("GetWorkerConfig: no config found for worker %s", workerName)
		return &pb.GetWorkerConfigResp{
			Config: "{}",
		}, nil
	}

	return &pb.GetWorkerConfigResp{
		Config: config,
	}, nil
}
