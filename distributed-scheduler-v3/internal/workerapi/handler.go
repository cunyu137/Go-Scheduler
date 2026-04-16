package workerapi

import (
	"net/http"

	"distributed-scheduler-v3/internal/adminclient"
	"distributed-scheduler-v3/internal/model"
	"distributed-scheduler-v3/internal/workerrunner"

	"github.com/gin-gonic/gin"
)

type Handler struct{ runner *workerrunner.Runner }

func NewHandler(runner *workerrunner.Runner) *Handler { return &Handler{runner: runner} }

func (h *Handler) Health(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }

func (h *Handler) Execute(c *gin.Context) {
	var req adminclient.DispatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"accepted": false, "message": err.Error()})
		return
	}
	if req.HandlerName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"accepted": false, "message": "handler_name required"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"accepted": true, "message": "accepted"})
	go h.runner.Run(req.InstanceID, req.HandlerName, req.Payload, req.TimeoutSeconds, req.IdempotentKey)
	_ = model.InstanceStatusSuccess
}
