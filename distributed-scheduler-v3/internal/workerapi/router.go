package workerapi

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()
	r.GET("/health", h.Health)
	r.POST("/internal/execute", h.Execute)
	return r
}
