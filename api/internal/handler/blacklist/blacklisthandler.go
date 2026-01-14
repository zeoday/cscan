package blacklist

import (
	"net/http"

	"cscan/api/internal/logic"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// BlacklistConfigGetHandler 获取黑名单配置
func BlacklistConfigGetHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewBlacklistLogic(r.Context(), svcCtx)
		resp, err := l.GetBlacklistConfig()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// BlacklistConfigSaveHandler 保存黑名单配置
func BlacklistConfigSaveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.BlacklistConfigSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewBlacklistLogic(r.Context(), svcCtx)
		resp, err := l.SaveBlacklistConfig(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// BlacklistRulesHandler 获取黑名单规则列表（供Worker调用）
func BlacklistRulesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewBlacklistLogic(r.Context(), svcCtx)
		resp, err := l.GetBlacklistRules()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
