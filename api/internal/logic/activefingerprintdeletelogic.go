package logic

import (
	"context"
	"errors"

	"cscan/api/internal/svc"
	"cscan/api/internal/types"
)

type ActiveFingerprintDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewActiveFingerprintDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActiveFingerprintDeleteLogic {
	return &ActiveFingerprintDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ActiveFingerprintDeleteLogic) ActiveFingerprintDelete(req *types.ActiveFingerprintDeleteReq) (*types.BaseResp, error) {
	if req.Id == "" {
		return nil, errors.New("ID不能为空")
	}

	err := l.svcCtx.ActiveFingerprintModel.Delete(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.BaseResp{
		Code: 0,
		Msg:  "删除成功",
	}, nil
}
