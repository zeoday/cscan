package logic

import (
	"context"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"

	"go.mongodb.org/mongo-driver/bson"
)

type ActiveFingerprintClearLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActiveFingerprintClearLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActiveFingerprintClearLogic {
	return &ActiveFingerprintClearLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActiveFingerprintClearLogic) ActiveFingerprintClear() (*types.ActiveFingerprintClearResp, error) {
	// 先获取总数
	count, _ := l.svcCtx.ActiveFingerprintModel.Count(l.ctx, bson.M{})

	// 清空所有
	err := l.svcCtx.ActiveFingerprintModel.DeleteAll(l.ctx)
	if err != nil {
		return nil, err
	}

	return &types.ActiveFingerprintClearResp{
		Code:    0,
		Msg:     "清空成功",
		Deleted: int(count),
	}, nil
}
