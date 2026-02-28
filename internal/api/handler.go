package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kitsnail/insight-hub/internal/model"
	"github.com/kitsnail/insight-hub/internal/service"
)

// Handler API Handler
type Handler struct {
	itemSvc *service.ItemService
	tagSvc  *service.TagService
}

// NewHandler 创建 Handler
func NewHandler(itemSvc *service.ItemService, tagSvc *service.TagService) *Handler {
	return &Handler{
		itemSvc: itemSvc,
		tagSvc:  tagSvc,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Items
	mux.HandleFunc("GET /api/items", h.ListItems)
	mux.HandleFunc("POST /api/items", h.CreateItem)
	mux.HandleFunc("GET /api/items/{id}", h.GetItem)
	mux.HandleFunc("PUT /api/items/{id}", h.UpdateItem)
	mux.HandleFunc("DELETE /api/items/{id}", h.DeleteItem)
	mux.HandleFunc("GET /api/items/search", h.SearchItems)

	// Tags
	mux.HandleFunc("GET /api/tags", h.ListTags)
	mux.HandleFunc("POST /api/tags", h.CreateTag)
	mux.HandleFunc("GET /api/tags/{id}", h.GetTag)
	mux.HandleFunc("PUT /api/tags/{id}", h.UpdateTag)
	mux.HandleFunc("DELETE /api/tags/{id}", h.DeleteTag)
}

// ========== Items ==========

// CreateItem 创建 Item
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var input service.CreateItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 验证
	if input.Type == "" {
		input.Type = model.ItemTypeNote
	}
	if input.Source == "" {
		input.Source = "manual"
	}

	item, err := h.itemSvc.CreateItem(input)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusCreated, item)
}

// GetItem 获取 Item
func (h *Handler) GetItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "missing id")
		return
	}

	item, err := h.itemSvc.GetItem(id)
	if err != nil {
		h.error(w, http.StatusNotFound, "item not found")
		return
	}

	h.json(w, http.StatusOK, item)
}

// ListItems 列出 Items
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	input := service.ListItemsInput{
		Page:     parseInt(r.URL.Query().Get("page"), 1),
		PageSize: parseInt(r.URL.Query().Get("page_size"), 20),
		Type:     model.ItemType(r.URL.Query().Get("type")),
		Status:   model.ItemStatus(r.URL.Query().Get("status")),
	}

	// 默认不显示已删除
	if input.Status == "" {
		input.Status = model.ItemStatusActive
	}

	result, err := h.itemSvc.ListItems(input)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, result)
}

// UpdateItem 更新 Item
func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "missing id")
		return
	}

	var input service.UpdateItemInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	item, err := h.itemSvc.UpdateItem(id, input)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, item)
}

// DeleteItem 删除 Item
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "missing id")
		return
	}

	if err := h.itemSvc.DeleteItem(id); err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SearchItems 搜索 Items
func (h *Handler) SearchItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.error(w, http.StatusBadRequest, "missing query")
		return
	}

	limit := parseInt(r.URL.Query().Get("limit"), 20)

	items, err := h.itemSvc.SearchItems(query, limit)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"query": query,
	})
}

// ========== Tags ==========

// CreateTag 创建 Tag
func (h *Handler) CreateTag(w http.ResponseWriter, r *http.Request) {
	var input service.CreateTagInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.Name == "" {
		h.error(w, http.StatusBadRequest, "name is required")
		return
	}

	tag, err := h.tagSvc.CreateTag(input)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusCreated, tag)
}

// GetTag 获取 Tag
func (h *Handler) GetTag(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.error(w, http.StatusBadRequest, "invalid id")
		return
	}

	tag, err := h.tagSvc.GetTag(id)
	if err != nil {
		h.error(w, http.StatusNotFound, "tag not found")
		return
	}

	h.json(w, http.StatusOK, tag)
}

// ListTags 列出 Tags
func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	tags, err := h.tagSvc.ListTagsWithCount(category)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, tags)
}

// UpdateTag 更新 Tag
func (h *Handler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.error(w, http.StatusBadRequest, "invalid id")
		return
	}

	var input service.UpdateTagInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tag, err := h.tagSvc.UpdateTag(id, input)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, tag)
}

// DeleteTag 删除 Tag
func (h *Handler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.tagSvc.DeleteTag(id); err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== Helpers ==========

func (h *Handler) json(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) error(w http.ResponseWriter, status int, message string) {
	h.json(w, status, map[string]string{"error": message})
}

func parseInt(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}
