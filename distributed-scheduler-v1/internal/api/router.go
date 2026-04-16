package api

import "github.com/gin-gonic/gin"

func NewRouter(taskHandler *TaskHandler) *gin.Engine {
	r := gin.Default()
	r.GET("/health", taskHandler.Health)
	v1 := r.Group("/api/v1")
	{
		v1.POST("/tasks/delay", taskHandler.CreateDelayTask)
		v1.POST("/tasks/cron", taskHandler.CreateCronTask)
		v1.GET("/tasks", taskHandler.ListTasks)
		v1.GET("/task-instances", taskHandler.ListTaskInstances)
	}
	return r
}
