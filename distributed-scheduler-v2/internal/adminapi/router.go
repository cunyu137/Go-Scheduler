package adminapi

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()
	r.GET("/health", h.Health)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/tasks/delay", h.CreateDelayTask)
		v1.POST("/tasks/cron", h.CreateCronTask)
		v1.GET("/tasks", h.ListTasks)
		v1.GET("/task-instances", h.ListTaskInstances)
		v1.GET("/workers", h.ListWorkers)
	}

	internal := r.Group("/internal")
	{
		internal.POST("/workers/register", h.RegisterWorker)
		internal.POST("/workers/heartbeat", h.Heartbeat)
		internal.POST("/instances/:id/start", h.MarkStarted)
		internal.POST("/instances/:id/finish", h.MarkFinished)
	}
	return r
}
