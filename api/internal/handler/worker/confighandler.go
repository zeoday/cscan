package worker

import (
	"encoding/json"
	"net/http"
	"strings"

	"cscan/api/internal/svc"
	"cscan/model"
	"cscan/pkg/response"
	"cscan/rpc/task/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// ==================== Templates Config Types ====================

// WorkerTemplatesReq 模板获取请求
type WorkerTemplatesReq struct {
	// 按标签获取
	Tags       []string `json:"tags,omitempty"`
	Severities []string `json:"severities,omitempty"`
	// 按ID获取
	NucleiTemplateIds []string `json:"nucleiTemplateIds,omitempty"`
	CustomPocIds      []string `json:"customPocIds,omitempty"`
}

// WorkerTemplatesResp 模板获取响应
type WorkerTemplatesResp struct {
	Code      int      `json:"code"`
	Msg       string   `json:"msg"`
	Success   bool     `json:"success"`
	Templates []string `json:"templates"`
	Count     int32    `json:"count"`
}

// ==================== Fingerprints Config Types ====================

// WorkerFingerprintsReq 指纹获取请求
type WorkerFingerprintsReq struct {
	EnabledOnly bool `json:"enabledOnly"`
}

// WorkerFingerprintDocument 指纹文档
type WorkerFingerprintDocument struct {
	Id        string            `json:"id"`
	Name      string            `json:"name"`
	Category  string            `json:"category"`
	Rule      string            `json:"rule"`
	Source    string            `json:"source"`
	Headers   map[string]string `json:"headers"`
	Cookies   map[string]string `json:"cookies"`
	Html      []string          `json:"html"`
	Scripts   []string          `json:"scripts"`
	ScriptSrc []string          `json:"scriptSrc"`
	Meta      map[string]string `json:"meta"`
	Css       []string          `json:"css"`
	Url       []string          `json:"url"`
	IsBuiltin bool              `json:"isBuiltin"`
	Enabled   bool              `json:"enabled"`
}

// WorkerFingerprintsResp 指纹获取响应
type WorkerFingerprintsResp struct {
	Code         int                         `json:"code"`
	Msg          string                      `json:"msg"`
	Success      bool                        `json:"success"`
	Fingerprints []WorkerFingerprintDocument `json:"fingerprints"`
	Count        int32                       `json:"count"`
}

// ==================== Subfinder Config Types ====================

// WorkerSubfinderReq Subfinder配置获取请求
type WorkerSubfinderReq struct {
	WorkspaceId string `json:"workspaceId"`
}

// WorkerSubfinderProvider Subfinder数据源
type WorkerSubfinderProvider struct {
	Id          string   `json:"id"`
	Provider    string   `json:"provider"`
	Keys        []string `json:"keys"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
}

// WorkerSubfinderResp Subfinder配置获取响应
type WorkerSubfinderResp struct {
	Code      int                       `json:"code"`
	Msg       string                    `json:"msg"`
	Success   bool                      `json:"success"`
	Providers []WorkerSubfinderProvider `json:"providers"`
	Count     int32                     `json:"count"`
}

// ==================== HttpService Config Types ====================

// WorkerHttpServiceReq HTTP服务映射获取请求
type WorkerHttpServiceReq struct {
	EnabledOnly bool `json:"enabledOnly"`
}

// WorkerHttpServiceMapping HTTP服务映射
type WorkerHttpServiceMapping struct {
	Id          string `json:"id"`
	ServiceName string `json:"serviceName"`
	IsHttp      bool   `json:"isHttp"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// WorkerHttpServiceResp HTTP服务映射获取响应
type WorkerHttpServiceResp struct {
	Code     int                        `json:"code"`
	Msg      string                     `json:"msg"`
	Success  bool                       `json:"success"`
	Mappings []WorkerHttpServiceMapping `json:"mappings"`
	Count    int32                      `json:"count"`
}

// ==================== Active Fingerprints Config Types ====================

// WorkerActiveFingerprintsReq 主动指纹获取请求
type WorkerActiveFingerprintsReq struct {
	EnabledOnly bool `json:"enabledOnly"`
}

// WorkerActiveFingerprintDocument 主动指纹文档
type WorkerActiveFingerprintDocument struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`        // 应用名称（用于关联被动指纹）
	Paths       []string `json:"paths"`       // 主动探测路径列表
	Description string   `json:"description"`
	Enabled     bool     `json:"enabled"`
	// 关联的被动指纹规则（用于匹配响应）
	Rule      string            `json:"rule,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
	Cookies   map[string]string `json:"cookies,omitempty"`
	Html      []string          `json:"html,omitempty"`
	Scripts   []string          `json:"scripts,omitempty"`
	ScriptSrc []string          `json:"scriptSrc,omitempty"`
	Meta      map[string]string `json:"meta,omitempty"`
	Css       []string          `json:"css,omitempty"`
	Url       []string          `json:"url,omitempty"`
}

// WorkerActiveFingerprintsResp 主动指纹获取响应
type WorkerActiveFingerprintsResp struct {
	Code         int                               `json:"code"`
	Msg          string                            `json:"msg"`
	Success      bool                              `json:"success"`
	Fingerprints []WorkerActiveFingerprintDocument `json:"fingerprints"`
	Count        int32                             `json:"count"`
}

// ==================== Templates Handler ====================

// WorkerConfigTemplatesHandler 模板配置获取接口
// POST /api/v1/worker/config/templates
func WorkerConfigTemplatesHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req WorkerTemplatesReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, &WorkerTemplatesResp{Code: 400, Msg: "参数解析失败"})
			return
		}

		var templates []string
		var count int32

		// 优先按ID获取
		if len(req.NucleiTemplateIds) > 0 || len(req.CustomPocIds) > 0 {
			rpcReq := &pb.GetTemplatesByIdsReq{
				NucleiTemplateIds: req.NucleiTemplateIds,
				CustomPocIds:      req.CustomPocIds,
			}
			rpcResp, err := svcCtx.TaskRpcClient.GetTemplatesByIds(r.Context(), rpcReq)
			if err != nil {
				logx.Errorf("[WorkerConfigTemplates] RPC GetTemplatesByIds error: %v", err)
				response.Error(w, err)
				return
			}
			templates = rpcResp.Templates
			count = rpcResp.Count
		} else {
			// 按标签获取
			rpcReq := &pb.GetTemplatesByTagsReq{
				Tags:       req.Tags,
				Severities: req.Severities,
			}
			rpcResp, err := svcCtx.TaskRpcClient.GetTemplatesByTags(r.Context(), rpcReq)
			if err != nil {
				logx.Errorf("[WorkerConfigTemplates] RPC GetTemplatesByTags error: %v", err)
				response.Error(w, err)
				return
			}
			templates = rpcResp.Templates
			count = rpcResp.Count
		}

		httpx.OkJson(w, &WorkerTemplatesResp{
			Code:      0,
			Msg:       "success",
			Success:   true,
			Templates: templates,
			Count:     count,
		})
	}
}

// ==================== Fingerprints Handler ====================

// WorkerConfigFingerprintsHandler 指纹配置获取接口
// POST /api/v1/worker/config/fingerprints
func WorkerConfigFingerprintsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req WorkerFingerprintsReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, &WorkerFingerprintsResp{Code: 400, Msg: "参数解析失败"})
			return
		}

		rpcReq := &pb.GetCustomFingerprintsReq{
			EnabledOnly: req.EnabledOnly,
		}

		rpcResp, err := svcCtx.TaskRpcClient.GetCustomFingerprints(r.Context(), rpcReq)
		if err != nil {
			logx.Errorf("[WorkerConfigFingerprints] RPC GetCustomFingerprints error: %v", err)
			response.Error(w, err)
			return
		}

		// 转换指纹数据
		fingerprints := make([]WorkerFingerprintDocument, 0, len(rpcResp.Fingerprints))
		for _, fp := range rpcResp.Fingerprints {
			fingerprints = append(fingerprints, WorkerFingerprintDocument{
				Id:        fp.Id,
				Name:      fp.Name,
				Category:  fp.Category,
				Rule:      fp.Rule,
				Source:    fp.Source,
				Headers:   fp.Headers,
				Cookies:   fp.Cookies,
				Html:      fp.Html,
				Scripts:   fp.Scripts,
				ScriptSrc: fp.ScriptSrc,
				Meta:      fp.Meta,
				Css:       fp.Css,
				Url:       fp.Url,
				IsBuiltin: fp.IsBuiltin,
				Enabled:   fp.Enabled,
			})
		}

		httpx.OkJson(w, &WorkerFingerprintsResp{
			Code:         0,
			Msg:          "success",
			Success:      true,
			Fingerprints: fingerprints,
			Count:        rpcResp.Count,
		})
	}
}

// ==================== Subfinder Handler ====================

// WorkerConfigSubfinderHandler Subfinder配置获取接口
// POST /api/v1/worker/config/subfinder
func WorkerConfigSubfinderHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req WorkerSubfinderReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, &WorkerSubfinderResp{Code: 400, Msg: "参数解析失败"})
			return
		}

		rpcReq := &pb.GetSubfinderProvidersReq{
			WorkspaceId: req.WorkspaceId,
		}

		rpcResp, err := svcCtx.TaskRpcClient.GetSubfinderProviders(r.Context(), rpcReq)
		if err != nil {
			logx.Errorf("[WorkerConfigSubfinder] RPC GetSubfinderProviders error: %v", err)
			response.Error(w, err)
			return
		}

		// 转换数据源数据
		providers := make([]WorkerSubfinderProvider, 0, len(rpcResp.Providers))
		for _, p := range rpcResp.Providers {
			providers = append(providers, WorkerSubfinderProvider{
				Id:          p.Id,
				Provider:    p.Provider,
				Keys:        p.Keys,
				Status:      p.Status,
				Description: p.Description,
			})
		}

		httpx.OkJson(w, &WorkerSubfinderResp{
			Code:      0,
			Msg:       "success",
			Success:   true,
			Providers: providers,
			Count:     rpcResp.Count,
		})
	}
}

// ==================== HttpService Handler ====================

// WorkerConfigHttpServiceHandler HTTP服务映射获取接口
// POST /api/v1/worker/config/httpservice
func WorkerConfigHttpServiceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req WorkerHttpServiceReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, &WorkerHttpServiceResp{Code: 400, Msg: "参数解析失败"})
			return
		}

		rpcReq := &pb.GetHttpServiceMappingsReq{
			EnabledOnly: req.EnabledOnly,
		}

		rpcResp, err := svcCtx.TaskRpcClient.GetHttpServiceMappings(r.Context(), rpcReq)
		if err != nil {
			logx.Errorf("[WorkerConfigHttpService] RPC GetHttpServiceMappings error: %v", err)
			response.Error(w, err)
			return
		}

		// 转换映射数据
		mappings := make([]WorkerHttpServiceMapping, 0, len(rpcResp.Mappings))
		for _, m := range rpcResp.Mappings {
			mappings = append(mappings, WorkerHttpServiceMapping{
				Id:          m.Id,
				ServiceName: m.ServiceName,
				IsHttp:      m.IsHttp,
				Description: m.Description,
				Enabled:     m.Enabled,
			})
		}

		httpx.OkJson(w, &WorkerHttpServiceResp{
			Code:     0,
			Msg:      "success",
			Success:  true,
			Mappings: mappings,
			Count:    rpcResp.Count,
		})
	}
}

// ==================== Active Fingerprints Handler ====================

// WorkerConfigActiveFingerprintsHandler 主动指纹配置获取接口
// POST /api/v1/worker/config/activefingerprints
func WorkerConfigActiveFingerprintsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req WorkerActiveFingerprintsReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, &WorkerActiveFingerprintsResp{Code: 400, Msg: "参数解析失败"})
			return
		}

		ctx := r.Context()
		var docs []WorkerActiveFingerprintDocument

		// 获取主动指纹
		var activeFingerprints []model.ActiveFingerprint
		var err error
		if req.EnabledOnly {
			activeFingerprints, err = svcCtx.ActiveFingerprintModel.FindEnabled(ctx)
		} else {
			activeFingerprints, err = svcCtx.ActiveFingerprintModel.FindAll(ctx)
		}
		if err != nil {
			logx.Errorf("[WorkerConfigActiveFingerprints] Find error: %v", err)
			response.Error(w, err)
			return
		}

		// 收集所有主动指纹的名称，用于批量查询关联的被动指纹
		names := make([]string, 0, len(activeFingerprints))
		for _, fp := range activeFingerprints {
			names = append(names, fp.Name)
		}

		// 批量获取关联的被动指纹规则
		passiveFpMap := make(map[string]*model.Fingerprint)
		if len(names) > 0 && svcCtx.FingerprintModel != nil {
			passiveFingerprints, err := svcCtx.FingerprintModel.FindByNames(ctx, names)
			if err != nil {
				logx.Infof("[WorkerConfigActiveFingerprints] FindByNames error: %v", err)
				// 不返回错误，继续处理（主动指纹仍然可以返回，只是没有匹配规则）
			} else {
				for _, pf := range passiveFingerprints {
					// 使用小写名称作为key，支持不区分大小写匹配
					passiveFpMap[strings.ToLower(pf.Name)] = pf
				}
				logx.Debugf("[WorkerConfigActiveFingerprints] Found %d passive fingerprints for %d active fingerprints", len(passiveFingerprints), len(names))
			}
		}

		// 构建返回数据，包含关联的被动指纹规则
		for _, fp := range activeFingerprints {
			doc := WorkerActiveFingerprintDocument{
				Id:          fp.Id.Hex(),
				Name:        fp.Name,
				Paths:       fp.Paths,
				Description: fp.Description,
				Enabled:     fp.Enabled,
			}

			// 查找关联的被动指纹，复制其匹配规则（使用小写名称匹配）
			if passiveFp, ok := passiveFpMap[strings.ToLower(fp.Name)]; ok {
				doc.Rule = passiveFp.Rule
				doc.Headers = passiveFp.Headers
				doc.Cookies = passiveFp.Cookies
				doc.Html = passiveFp.HTML
				doc.Scripts = passiveFp.Scripts
				doc.ScriptSrc = passiveFp.ScriptSrc
				doc.Meta = passiveFp.Meta
				doc.Css = passiveFp.CSS
				doc.Url = passiveFp.URL
			}

			docs = append(docs, doc)
		}

		httpx.OkJson(w, &WorkerActiveFingerprintsResp{
			Code:         0,
			Msg:          "success",
			Success:      true,
			Fingerprints: docs,
			Count:        int32(len(docs)),
		})
	}
}

// ==================== POC Config Types ====================

// WorkerPocReq POC获取请求
type WorkerPocReq struct {
	PocId   string `json:"pocId"`
	PocType string `json:"pocType"` // nuclei, custom
}

// WorkerPocResp POC获取响应
type WorkerPocResp struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Success bool   `json:"success"`
	Content string `json:"content"`
	PocId   string `json:"pocId"`
	PocType string `json:"pocType"`
}

// ==================== POC Handler ====================

// WorkerConfigPocHandler POC配置获取接口
// POST /api/v1/worker/config/poc
func WorkerConfigPocHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req WorkerPocReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.OkJson(w, &WorkerPocResp{Code: 400, Msg: "参数解析失败"})
			return
		}

		if req.PocId == "" {
			httpx.OkJson(w, &WorkerPocResp{Code: 400, Msg: "pocId不能为空"})
			return
		}

		rpcReq := &pb.GetPocByIdReq{
			PocId:   req.PocId,
			PocType: req.PocType,
		}

		rpcResp, err := svcCtx.TaskRpcClient.GetPocById(r.Context(), rpcReq)
		if err != nil {
			logx.Errorf("[WorkerConfigPoc] RPC GetPocById error: %v", err)
			response.Error(w, err)
			return
		}

		if !rpcResp.Success {
			httpx.OkJson(w, &WorkerPocResp{
				Code:    500,
				Msg:     rpcResp.Message,
				Success: false,
			})
			return
		}

		httpx.OkJson(w, &WorkerPocResp{
			Code:    0,
			Msg:     "success",
			Success: true,
			Content: rpcResp.Content,
			PocId:   rpcResp.PocId,
			PocType: rpcResp.PocType,
		})
	}
}
