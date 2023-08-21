package httpserver

import (
	"github.com/gin-gonic/gin"
	"jwt-auth/internal/app"
	"net/http"
)

func SetRoutes(r gin.IRouter, a app.App) {
	r.Any("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/generate", generatePair(a))
	r.POST("/refresh", refreshPair(a))
}
