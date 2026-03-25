package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"task-manager/internal/middleware"
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

	// Default Status if empty
	if req.Status == "" {
		req.Status = "IN_PROGRESS"
	}

	status := models.TaskStatus(req.Status)
	if !status.IsValid() {
		RespondError(w, http.StatusBadRequest, "invalid status")
		return
	}

	task := &models.Task{
		ID:     uuid.NewString(),
		Title:  req.Title,
		Status: status,
	}

	if err := h.store.Create(task); err != nil {
		slog.Error("failed to create task", "error", err, "request_id", middleware.GetRequestID(r))
		RespondError(w, http.StatusInternalServerError, "failed to create task, internal error")
		return
	}

	RespondJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("id")

	if _, err := uuid.Parse(taskId); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid task id format")
		return
	}

	task, err := h.store.Get(taskId)
	if errors.Is(err, store.ErrTaskNotFound) {
		RespondError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		slog.Error("failed to get task", "error", err, "request_id", middleware.GetRequestID(r))
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve task, internal error")
		return
	}

	RespondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.store.GetAll()
	if err != nil {
		slog.Error("failed to get tasks", "error", err, "request_id", middleware.GetRequestID(r))
		RespondError(w, http.StatusInternalServerError, "failed to retrieve tasks, internal error")
		return
	}
	RespondJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("id")

	if _, err := uuid.Parse(taskId); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid task id format")
		return
	}

	defer r.Body.Close()
	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Title == nil && req.Status == nil {
		RespondError(w, http.StatusBadRequest, "at least one field must be provided")
		return
	}

	var taskToUpdate models.UpdateTaskInput

	if req.Title != nil {
		if *req.Title == "" {
			RespondError(w, http.StatusBadRequest, "title cannot be empty")
			return
		}
		taskToUpdate.Title = req.Title
	}

	if req.Status != nil {
		status := models.TaskStatus(*req.Status)
		if !status.IsValid() {
			RespondError(w, http.StatusBadRequest, "invalid status, must be one of the supported ones")
			return
		}
		taskToUpdate.Status = &status
	}

	updatedTask, err := h.store.Update(taskId, &taskToUpdate)
	if errors.Is(err, store.ErrTaskNotFound) {
		RespondError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		slog.Error("failed to update task", "error", err, "request_id", middleware.GetRequestID(r))
		RespondError(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	RespondJSON(w, http.StatusOK, updatedTask)

}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskId := r.PathValue("id")

	if _, err := uuid.Parse(taskId); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid task id format")
		return
	}
	err := h.store.Delete(taskId)
	if errors.Is(err, store.ErrTaskNotFound) {
		RespondError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		slog.Error("failed to delete task", "error", err, "request_id", middleware.GetRequestID(r))
		RespondError(w, http.StatusInternalServerError, "failed to delete task, internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
