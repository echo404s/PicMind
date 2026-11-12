package router

import (
	"net/http"

	"github.com/echo404s/PicMind/backend/vision/handler"
	"github.com/gin-gonic/gin"
)

// Register mounts vision HTTP routes for api-gateway.
func Register(r *gin.Engine, _ *handler.Handlers) {
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
}
