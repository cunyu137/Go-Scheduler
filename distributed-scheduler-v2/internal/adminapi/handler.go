package adminapi

import (
	"net/http"
	"strconv"

	"distributed-scheduler-v2/internal/model"
	"distributed-scheduler-v2/internal/repository"
	"distributed-scheduler-v2/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	taskService     *service.TaskService
	taskRepo        *repository.TaskRepository
	instanceRepo    *repository.TaskInstanceRepository
	workerRepo      *repository.WorkerRepository
	callbackService *service.CallbackService
	logger          *logrus.Logger
}

func NewHandler(taskService *service.TaskService, taskRepo *repository.TaskRepository, instanceRepo *repository.TaskInstanceRepository, workerRepo *repository.WorkerRepository, callbackService *service.CallbackService, logger *logrus.Logger) *Handler {
	return &Handler{taskService: taskService, taskRepo: taskRepo, instanceRepo: instanceRepo, workerRepo: workerRepo, callbackService: callbackService, logger: logger}
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok", "status": model.TaskStatusActive})
}

func (h *Handler) CreateDelayTask(c *gin.Context) {
	var req service.CreateDelayTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.taskService.CreateDelayTask(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": task})
}

func (h *Handler) CreateCronTask(c *gin.Context) {
	var req service.CreateCronTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.taskService.CreateCronTask(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok", "data": task})
}

func (h *Handler) ListTasks(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 10)
	status := parseIntDefault(c.Query("status"), 0)
	taskType := c.Query("task_type")
	tasks, total, err := h.taskRepo.List(page, pageSize, taskType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": tasks, "total": total, "page": page, "page_size": pageSize})
}

func (h *Handler) ListTaskInstances(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Query("task_id"), 10, 64)
	if err != nil || taskID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task_id"})
		return
	}
	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 10)
	items, total, err := h.instanceRepo.ListByTaskID(taskID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total, "page": page, "page_size": pageSize})
}

func (h *Handler) ListWorkers(c *gin.Context) {
	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 20)
	items, total, err := h.workerRepo.List(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total, "page": page, "page_size": pageSize})
}

type workerRegisterRequest struct {
	ID      string `json:"id" binding:"required"`
	Address string `json:"address" binding:"required"`
}

func (h *Handler) RegisterWorker(c *gin.Context) {
	var req workerRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.workerRepo.UpsertHeartbeat(req.ID, req.Address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}
func (h *Handler) Heartbeat(c *gin.Context) { h.RegisterWorker(c) }

type finishRequest struct {
	WorkerID   string `json:"worker_id" binding:"required"`
	Success    bool   `json:"success"`
	TimedOut   bool   `json:"timed_out"`
	Message    string `json:"message"`
	RetryCount int    `json:"retry_count"`
	MaxRetry   int    `json:"max_retry"`
}

type startRequest struct {
	WorkerID string `json:"worker_id" binding:"required"`
}

func (h *Handler) MarkStarted(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}
	var req startRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.callbackService.MarkStarted(id, req.WorkerID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (h *Handler) MarkFinished(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid instance id"})
		return
	}
	var req finishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.callbackService.Finish(id, req.WorkerID, req.Success, req.TimedOut, req.Message, req.RetryCount, req.MaxRetry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil || v <= 0 {
		return def
	}
	return v
}
