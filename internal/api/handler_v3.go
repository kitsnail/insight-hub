package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/kitsnail/insight-hub/internal/model"
	"github.com/kitsnail/insight-hub/internal/repository"
)

// HandlerV3 API Handler v3
type HandlerV3 struct {
	itemRepo repository.ItemRepositoryV3
}

// NewHandlerV3 创建 Handler v3
func NewHandlerV3(itemRepo repository.ItemRepositoryV3) *HandlerV3 {
	return &HandlerV3{
		itemRepo: itemRepo,
	}
}

// RegisterRoutes 注册路由
func (h *HandlerV3) RegisterRoutes(mux *http.ServeMux) {
	// Items CRUD
	mux.HandleFunc("POST /api/v1/items", h.CreateItem)
	mux.HandleFunc("GET /api/v1/items", h.ListItems)
	mux.HandleFunc("GET /api/v1/items/{id}", h.GetItem)
	mux.HandleFunc("PUT /api/v1/items/{id}", h.UpdateItem)
	mux.HandleFunc("DELETE /api/v1/items/{id}", h.DeleteItem)

	// 去重检查
	mux.HandleFunc("POST /api/v1/items/check-duplicate", h.CheckDuplicate)

	// 批量操作
	mux.HandleFunc("POST /api/v1/items/batch", h.BatchCreateItems)
	mux.HandleFunc("POST /api/v1/items/batch-delete", h.BatchDeleteItems)
	mux.HandleFunc("POST /api/v1/items/batch-update-status", h.BatchUpdateStatus)

	// Stats
	mux.HandleFunc("GET /api/v1/stats", h.GetStats)
	mux.HandleFunc("GET /api/v1/stats/by-type", h.GetStatsByType)
	mux.HandleFunc("GET /api/v1/stats/by-source", h.GetStatsBySource)

	// Health
	mux.HandleFunc("GET /api/v1/health", h.Health)
}

// ========== Items CRUD ==========

// CreateItem 创建 Item
// POST /api/v1/items
func (h *HandlerV3) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req model.ItemCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 验证必填字段
	if req.Type == "" {
		h.error(w, http.StatusBadRequest, "type is required")
		return
	}
	if req.SourceSystem == "" {
		h.error(w, http.StatusBadRequest, "source_system is required")
		return
	}
	if req.Title == "" {
		h.error(w, http.StatusBadRequest, "title is required")
		return
	}

	// 自动生成 ID（如果未提供）
	if req.ID == "" {
		req.ID = req.SourceSystem + ":" + uuid.New().String()
	}

	// 设置默认状态
	if req.Status == "" {
		req.Status = model.ItemStatusActive
	}

	item, err := h.itemRepo.Create(r.Context(), &req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusCreated, item)
}

// GetItem 获取 Item
// GET /api/v1/items/{id}
func (h *HandlerV3) GetItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "missing id")
		return
	}

	item, err := h.itemRepo.GetByID(r.Context(), id)
	if err != nil {
		h.error(w, http.StatusNotFound, "item not found")
		return
	}

	h.json(w, http.StatusOK, item)
}

// ListItems 列出 Items
// GET /api/v1/items?type=memory&source_system=openclaw&limit=100&offset=0
func (h *HandlerV3) ListItems(w http.ResponseWriter, r *http.Request) {
	var query model.ItemQuery

	// 解析查询参数
	query.Type = r.URL.Query().Get("type")
	query.Category = r.URL.Query().Get("category")
	query.SourceSystem = r.URL.Query().Get("source_system")
	query.SourceAgent = r.URL.Query().Get("source_agent")
	query.Status = r.URL.Query().Get("status")
	query.DateFrom = r.URL.Query().Get("date_from")
	query.DateTo = r.URL.Query().Get("date_to")
	query.Search = r.URL.Query().Get("search")
	query.Tags = r.URL.Query().Get("tags")
	query.SortBy = r.URL.Query().Get("sort_by")

	// 解析数值
	query.Limit = parseInt(r.URL.Query().Get("limit"), 100)
	query.Offset = parseInt(r.URL.Query().Get("offset"), 0)

	// 解析排序方向
	query.SortDesc = r.URL.Query().Get("sort_desc") != "false"

	result, err := h.itemRepo.List(r.Context(), &query)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, result)
}

// UpdateItem 更新 Item
// PUT /api/v1/items/{id}
func (h *HandlerV3) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "missing id")
		return
	}

	var req model.ItemUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.itemRepo.Update(r.Context(), id, &req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, item)
}

// DeleteItem 删除 Item（软删除）
// DELETE /api/v1/items/{id}
func (h *HandlerV3) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "missing id")
		return
	}

	if err := h.itemRepo.Delete(r.Context(), id); err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== 去重检查 ==========

// CheckDuplicate 检查重复项
// POST /api/v1/items/check-duplicate
func (h *HandlerV3) CheckDuplicate(w http.ResponseWriter, r *http.Request) {
	var req model.DedupCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 验证必填字段
	if req.SourceSystem == "" {
		h.error(w, http.StatusBadRequest, "source_system is required")
		return
	}

	resp, err := h.itemRepo.CheckDuplicate(r.Context(), &req)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, resp)
}

// ========== 批量操作 ==========

// BatchCreateItems 批量创建 Items
// POST /api/v1/items/batch
func (h *HandlerV3) BatchCreateItems(w http.ResponseWriter, r *http.Request) {
	var req model.BatchCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Items) == 0 {
		h.error(w, http.StatusBadRequest, "items array is empty")
		return
	}

	if len(req.Items) > 1000 {
		h.error(w, http.StatusBadRequest, "maximum 1000 items per batch")
		return
	}

	// 验证每个 item
	for i, item := range req.Items {
		if item.Type == "" {
			h.error(w, http.StatusBadRequest, fmt.Sprintf("items[%d].type is required", i))
			return
		}
		if item.SourceSystem == "" {
			h.error(w, http.StatusBadRequest, fmt.Sprintf("items[%d].source_system is required", i))
			return
		}
		if item.Title == "" {
			h.error(w, http.StatusBadRequest, fmt.Sprintf("items[%d].title is required", i))
			return
		}
		// 自动生成 ID
		if item.ID == "" {
			req.Items[i].ID = item.SourceSystem + ":" + uuid.New().String()
		}
		// 设置默认状态
		if req.Items[i].Status == "" {
			req.Items[i].Status = model.ItemStatusActive
		}
	}

	// 转换为指针数组
	items := make([]*model.ItemCreate, len(req.Items))
	for i := range req.Items {
		items[i] = &req.Items[i]
	}

	// 批量创建
	created, err := h.itemRepo.BatchCreate(r.Context(), items)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 构建响应
	resp := &model.BatchCreateResponse{
		Total:       len(req.Items),
		Succeeded:   len(created),
		FailedCount: len(req.Items) - len(created),
		Created:     make([]model.Item, len(created)),
		Failed:      []model.BatchCreateError{},
	}

	for i, item := range created {
		resp.Created[i] = *item
	}

	h.json(w, http.StatusCreated, resp)
}

// ========== Stats ==========

// GetStats 获取统计概览
// GET /api/v1/stats
func (h *HandlerV3) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.itemRepo.Stats(r.Context())
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, stats)
}

// GetStatsByType 按类型统计
// GET /api/v1/stats/by-type
func (h *HandlerV3) GetStatsByType(w http.ResponseWriter, r *http.Request) {
	counts, err := h.itemRepo.CountByType(r.Context())
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, map[string]interface{}{
		"by_type": counts,
	})
}

// GetStatsBySource 按来源统计
// GET /api/v1/stats/by-source
func (h *HandlerV3) GetStatsBySource(w http.ResponseWriter, r *http.Request) {
	counts, err := h.itemRepo.CountBySource(r.Context())
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, map[string]interface{}{
		"by_source": counts,
	})
}

// ========== 批量操作 ==========

// BatchDeleteItems 批量删除
// POST /api/v1/items/batch-delete
func (h *HandlerV3) BatchDeleteItems(w http.ResponseWriter, r *http.Request) {
	var req model.BatchDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		h.error(w, http.StatusBadRequest, "ids array is empty")
		return
	}

	if len(req.IDs) > 100 {
		h.error(w, http.StatusBadRequest, "maximum 100 items per batch")
		return
	}

	resp, err := h.itemRepo.BatchDelete(r.Context(), req.IDs)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, resp)
}

// BatchUpdateStatus 批量更新状态
// POST /api/v1/items/batch-update-status
func (h *HandlerV3) BatchUpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req model.BatchUpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.IDs) == 0 {
		h.error(w, http.StatusBadRequest, "ids array is empty")
		return
	}

	if len(req.IDs) > 100 {
		h.error(w, http.StatusBadRequest, "maximum 100 items per batch")
		return
	}

	if req.Status == "" {
		h.error(w, http.StatusBadRequest, "status is required")
		return
	}

	resp, err := h.itemRepo.BatchUpdateStatus(r.Context(), req.IDs, req.Status)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, resp)
}

// ========== Health ==========

// Health 健康检查
func (h *HandlerV3) Health(w http.ResponseWriter, r *http.Request) {
	h.json(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"version": "v3",
	})
}

// ========== Helpers ==========

func (h *HandlerV3) json(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *HandlerV3) error(w http.ResponseWriter, status int, message string) {
	h.json(w, status, map[string]string{"error": message})
}
