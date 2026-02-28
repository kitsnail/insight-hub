package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/kitsnail/insight-hub/internal/model"
	"github.com/kitsnail/insight-hub/internal/repository"
)

// Handler API Handler (v2 兼容层)
type Handler struct {
	itemRepo repository.ItemRepositoryV3
	tagRepo  repository.TagRepository
	taskRepo repository.TaskRepository
}

// NewHandler 创建 Handler
func NewHandler(itemRepo repository.ItemRepositoryV3, tagRepo repository.TagRepository, taskRepo repository.TaskRepository) *Handler {
	return &Handler{
		itemRepo: itemRepo,
		tagRepo:  tagRepo,
		taskRepo: taskRepo,
	}
}

// RegisterRoutes 注册路由 (v2 兼容)
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Items (v2 兼容路径)
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

	// Tasks
	mux.HandleFunc("GET /api/tasks", h.ListTasks)
	mux.HandleFunc("POST /api/tasks", h.CreateTask)
	mux.HandleFunc("GET /api/tasks/{id}", h.GetTask)
	mux.HandleFunc("PUT /api/tasks/{id}", h.UpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", h.DeleteTask)

	// Health
	mux.HandleFunc("GET /api/health", h.Health)
}

// ========== Items (v2 兼容) ==========

// CreateItem 创建 Item
func (h *Handler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req model.ItemCreate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// 验证
	if req.Type == "" {
		req.Type = model.TypeInsight
	}
	if req.SourceSystem == "" {
		req.SourceSystem = "manual"
	}
	if req.Title == "" {
		h.error(w, http.StatusBadRequest, "title is required")
		return
	}

	item, err := h.itemRepo.Create(r.Context(), &req)
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

	item, err := h.itemRepo.GetByID(r.Context(), id)
	if err != nil {
		h.error(w, http.StatusNotFound, "item not found")
		return
	}

	h.json(w, http.StatusOK, item)
}

// ListItems 列出 Items
func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	query := &model.ItemQuery{
		Type:   r.URL.Query().Get("type"),
		Status: r.URL.Query().Get("status"),
		Limit:  parseInt(r.URL.Query().Get("page_size"), 20),
	}

	if query.Status == "" {
		query.Status = "active"
	}

	result, err := h.itemRepo.List(r.Context(), query)
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

// DeleteItem 删除 Item
func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
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

// SearchItems 搜索 Items
func (h *Handler) SearchItems(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.error(w, http.StatusBadRequest, "missing query")
		return
	}

	limit := parseInt(r.URL.Query().Get("limit"), 20)

	items, err := h.itemRepo.Search(r.Context(), query, limit)
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
	var tag model.Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if tag.Name == "" {
		h.error(w, http.StatusBadRequest, "name is required")
		return
	}

	if err := h.tagRepo.Create(r.Context(), &tag); err != nil {
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

	tag, err := h.tagRepo.GetByID(r.Context(), id)
	if err != nil {
		h.error(w, http.StatusNotFound, "tag not found")
		return
	}

	h.json(w, http.StatusOK, tag)
}

// ListTags 列出 Tags
func (h *Handler) ListTags(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")

	tags, err := h.tagRepo.List(r.Context(), category)
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

	var tag model.Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tag.ID = id
	if err := h.tagRepo.Update(r.Context(), &tag); err != nil {
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

	if err := h.tagRepo.Delete(r.Context(), id); err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== Tasks ==========

// CreateTask 创建 Task
func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var task model.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if task.Title == "" {
		h.error(w, http.StatusBadRequest, "title is required")
		return
	}

	if err := h.taskRepo.Create(r.Context(), &task); err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusCreated, task)
}

// GetTask 获取 Task
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "missing id")
		return
	}

	task, err := h.taskRepo.GetByID(r.Context(), id)
	if err != nil {
		h.error(w, http.StatusNotFound, "task not found")
		return
	}

	h.json(w, http.StatusOK, task)
}

// ListTasks 列出 Tasks
func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	status := model.TaskStatus(r.URL.Query().Get("status"))
	limit := parseInt(r.URL.Query().Get("page_size"), 20)
	offset := 0

	tasks, total, err := h.taskRepo.List(r.Context(), status, limit, offset)
	if err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, map[string]interface{}{
		"tasks": tasks,
		"total": total,
	})
}

// UpdateTask 更新 Task
func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "missing id")
		return
	}

	var task model.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		h.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task.ID = id
	if err := h.taskRepo.Update(r.Context(), &task); err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.json(w, http.StatusOK, task)
}

// DeleteTask 删除 Task
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		h.error(w, http.StatusBadRequest, "missing id")
		return
	}

	if err := h.taskRepo.Delete(r.Context(), id); err != nil {
		h.error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ========== Health ==========

// Health 健康检查
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	h.json(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
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
