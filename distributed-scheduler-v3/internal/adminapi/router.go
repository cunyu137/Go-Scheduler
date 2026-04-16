package adminapi

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()
	r.GET("/health", h.Health)
	r.GET("/api/v1/leader", h.Leader)
	r.POST("/api/v1/tasks/delay", h.CreateDelayTask)
	r.POST("/api/v1/tasks/cron", h.CreateCronTask)
	r.GET("/api/v1/tasks", h.ListTasks)
	r.GET("/api/v1/task-instances", h.ListTaskInstances)
	r.GET("/api/v1/workers", h.ListWorkers)

	internal := r.Group("/internal")
	internal.POST("/workers/register", h.RegisterWorker)
	internal.POST("/workers/heartbeat", h.WorkerHeartbeat)
	internal.POST("/tasks/callback", h.TaskCallback)
	return r
}
