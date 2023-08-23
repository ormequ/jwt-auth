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

	// Метод POST, т.к. запрос предполагает возможность добавления в БД запись
	r.POST("/generate", generatePair(a))
	// Метод PUT, т.к. запрос изменяет только существующие записи
	r.PUT("/refresh", refreshPair(a))
}
