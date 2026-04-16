package workerapi

import (
	"net/http"

	"distributed-scheduler-v2/internal/workerrunner"

	"github.com/gin-gonic/gin"
)

type Handler struct{ runner *workerrunner.Runner }

func NewHandler(runner *workerrunner.Runner) *Handler { return &Handler{runner: runner} }
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "ok", "worker_id": h.runner.WorkerID()})
}

type executeRequest struct {
	InstanceID     int64  `json:"instance_id" binding:"required"`
	HandlerName    string `json:"handler_name" binding:"required"`
	Payload        string `json:"payload"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	RetryCount     int    `json:"retry_count"`
	MaxRetry       int    `json:"max_retry"`
	IdempotentKey  string `json:"idempotent_key" binding:"required"`
}

func (h *Handler) Execute(c *gin.Context) {
	var req executeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.runner.ExecuteAsync(workerrunner.Job{InstanceID: req.InstanceID, HandlerName: req.HandlerName, Payload: req.Payload, TimeoutSeconds: req.TimeoutSeconds, RetryCount: req.RetryCount, MaxRetry: req.MaxRetry, IdempotentKey: req.IdempotentKey})
	c.JSON(http.StatusOK, gin.H{"message": "accepted"})
}
