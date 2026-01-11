package subdomain

import (
	"net/http"

	"cscan/api/internal/logic"
	"cscan/api/internal/svc"
	"cscan/api/internal/types"
	"cscan/pkg/response"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// SubdomainDictListHandler 子域名字典列表
func SubdomainDictListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubdomainDictListReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewSubdomainDictListLogic(r.Context(), svcCtx)
		resp, err := l.SubdomainDictList(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// SubdomainDictSaveHandler 保存子域名字典
func SubdomainDictSaveHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubdomainDictSaveReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewSubdomainDictSaveLogic(r.Context(), svcCtx)
		resp, err := l.SubdomainDictSave(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// SubdomainDictDeleteHandler 删除子域名字典
func SubdomainDictDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SubdomainDictDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			response.ParamError(w, err.Error())
			return
		}

		l := logic.NewSubdomainDictDeleteLogic(r.Context(), svcCtx)
		resp, err := l.SubdomainDictDelete(&req)
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// SubdomainDictClearHandler 清空子域名字典
func SubdomainDictClearHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewSubdomainDictClearLogic(r.Context(), svcCtx)
		resp, err := l.SubdomainDictClear()
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}

// SubdomainDictEnabledListHandler 获取启用的子域名字典列表（用于任务创建时选择）
func SubdomainDictEnabledListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewSubdomainDictEnabledListLogic(r.Context(), svcCtx)
		resp, err := l.SubdomainDictEnabledList()
		if err != nil {
			response.Error(w, err)
			return
		}
		httpx.OkJson(w, resp)
	}
}
