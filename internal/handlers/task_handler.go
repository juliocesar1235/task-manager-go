package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"task-manager/internal/models"
	"task-manager/internal/store"

	"github.com/google/uuid"
)

type TaskHandler struct {
	store store.TaskStore
}

func NewTaskHandler(store store.TaskStore) *TaskHandler {
	return &TaskHandler{
		store: store,
	}
}

type createTaskRequest struct {
	Title  string `json:"title"`
	Status string `json:"status"`
}

type updateTaskRequest struct {
	Title  *string `json:"title"`
	Status *string `json:"status"`
}

var validStatuses = []string{"IN_PROGRESS", "PENDING", "COMPLETED", "ERROR"}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var req createTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == "" {
		RespondError(w, http.StatusBadRequest, "title is required")
		return
	}

	if req.Status != "" {
		if !isValidStatus(req.Status) {
			RespondError(w, http.StatusBadRequest, "invalid status, must be one of the supported ones")
			return
		}
	}

	if req.Status == "" {
		req.Status = "IN_PROGRESS"
	}

	task := &models.Task{
		ID:     uuid.NewString(),
		Title:  req.Title,
		Status: req.Status,
	}

	if err := h.store.Create(task); err != nil {
		RespondError(w, http.StatusInternalServerError, "failed to create task, internal error")
		return
	}

	RespondJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("id")

	task, err := h.store.Get(taskId)
	if errors.Is(err, store.ErrTaskNotFound) {
		RespondError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve task, internal error")
		return
	}

	RespondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.store.GetAll()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "failed to retrieve tasks, internal error")
		return
	}
	RespondJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("id")
	if taskId == "" {
		RespondError(w, http.StatusBadRequest, "task id is required")
		return
	}
	task, err := h.store.Get(taskId)
	if errors.Is(err, store.ErrTaskNotFound) {
		RespondError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "failed to retrieve task")
		return
	}

	defer r.Body.Close()
	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title != nil {
		if *req.Title == "" {
			RespondError(w, http.StatusBadRequest, "title cannot be empty")
			return
		}
		task.Title = *req.Title
	}

	if req.Status != nil {
		if !isValidStatus(*req.Status) {
			RespondError(w, http.StatusBadRequest, "invalid status, must be one of the supported ones")
			return
		}
		task.Status = *req.Status
	}

	err = h.store.Update(taskId, task)
	if errors.Is(err, store.ErrTaskNotFound) {
		RespondError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	RespondJSON(w, http.StatusOK, task)

}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("id")
	err := h.store.Delete(taskId)
	if errors.Is(err, store.ErrTaskNotFound) {
		RespondError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "failed to delete task, internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func isValidStatus(tStatus string) bool {
	return slices.Contains(validStatuses, tStatus)
}
