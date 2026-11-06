package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Register mounts gateway routes under the given group
// (e.g. main uses r.Group("/gateway") → /gateway/v1/ping).
func Register(r *gin.RouterGroup) {
	v1 := r.Group("/v1")
	{
		v1.GET("/ping", ping)
	}
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}
