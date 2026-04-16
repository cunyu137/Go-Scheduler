package workerapi

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()
	r.GET("/health", h.Health)
	internal := r.Group("/internal")
	{
		internal.POST("/execute", h.Execute)
	}
	return r
}
