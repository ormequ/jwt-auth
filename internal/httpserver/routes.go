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
	r.POST("/tokens/:user_id", generatePair(a))
	// Метод PUT, т.к. запрос изменяет только существующие записи
	r.PUT("/tokens/:user_id", refreshPair(a))
}
