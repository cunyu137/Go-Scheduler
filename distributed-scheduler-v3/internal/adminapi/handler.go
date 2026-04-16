package adminapi

import (
	"net/http"
	"strconv"

	"distributed-scheduler-v3/internal/repository"
	"distributed-scheduler-v3/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type LeaderView interface {
	IsLeader() bool
}

type Handler struct {
	taskService     *service.TaskService
	taskRepo        *repository.TaskRepository
	instanceRepo    *repository.TaskInstanceRepository
	workerRepo      *repository.WorkerRepository
	callbackService *service.CallbackService
	leaderView      LeaderView
	logger          *logrus.Logger
}

func NewHandler(taskService *service.TaskService, taskRepo *repository.TaskRepository, instanceRepo *repository.TaskInstanceRepository, workerRepo *repository.WorkerRepository, callbackService *service.CallbackService, leaderView LeaderView, logger *logrus.Logger) *Handler {
	return &Handler{taskService: taskService, taskRepo: taskRepo, instanceRepo: instanceRepo, workerRepo: workerRepo, callbackService: callbackService, leaderView: leaderView, logger: logger}
}

func (h *Handler) Health(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func (h *Handler) Leader(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"is_leader": h.leaderView.IsLeader()})
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
	c.JSON(http.StatusOK, task)
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
	c.JSON(http.StatusOK, task)
}

func (h *Handler) ListTasks(c *gin.Context) {
	items, err := h.taskRepo.List(parseInt(c.Query("limit"), 50), parseInt(c.Query("offset"), 0))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) ListTaskInstances(c *gin.Context) {
	items, err := h.instanceRepo.ListByTaskID(int64(parseInt(c.Query("task_id"), 0)), parseInt(c.Query("limit"), 50), parseInt(c.Query("offset"), 0))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) ListWorkers(c *gin.Context) {
	items, err := h.workerRepo.List(parseInt(c.Query("limit"), 100), parseInt(c.Query("offset"), 0))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

type registerWorkerRequest struct {
	WorkerID string `json:"worker_id" binding:"required"`
	Address  string `json:"address" binding:"required"`
}

func (h *Handler) RegisterWorker(c *gin.Context) {
	var req registerWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.workerRepo.Upsert(req.WorkerID, req.Address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"accepted": true})
}

func (h *Handler) WorkerHeartbeat(c *gin.Context) {
	var req registerWorkerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.workerRepo.Heartbeat(req.WorkerID, req.Address); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) TaskCallback(c *gin.Context) {
	var req service.CallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.callbackService.Handle(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func parseInt(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
