package api

import (
	"net/http"
	"strconv"

	"distributed-scheduler-v1/internal/model"
	"distributed-scheduler-v1/internal/repository"
	"distributed-scheduler-v1/internal/service"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	taskService  *service.TaskService
	taskRepo     *repository.TaskRepository
	instanceRepo *repository.TaskInstanceRepository
}

func NewTaskHandler(taskService *service.TaskService, taskRepo *repository.TaskRepository, instanceRepo *repository.TaskInstanceRepository) *TaskHandler {
	return &TaskHandler{taskService: taskService, taskRepo: taskRepo, instanceRepo: instanceRepo}
}

func (h *TaskHandler) CreateDelayTask(c *gin.Context) {
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

func (h *TaskHandler) CreateCronTask(c *gin.Context) {
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

func (h *TaskHandler) ListTasks(c *gin.Context) {
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

func (h *TaskHandler) ListTaskInstances(c *gin.Context) {
	taskID, err := strconv.ParseInt(c.Query("task_id"), 10, 64)
	if err != nil || taskID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task_id"})
		return
	}
	page := parseIntDefault(c.Query("page"), 1)
	pageSize := parseIntDefault(c.Query("page_size"), 10)
	instances, total, err := h.instanceRepo.ListByTaskID(taskID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": instances, "total": total, "page": page, "page_size": pageSize})
}

func (h *TaskHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok", "status": model.TaskStatusActive})
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
