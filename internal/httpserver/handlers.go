package httpserver

import (
	"github.com/gin-gonic/gin"
	"jwt-auth/internal/app"
	"net/http"
)

func generatePair(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req generateRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse(ErrEmptyRefresh))
			return
		}
		pair, err := a.GeneratePair(c, req.UserID)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, jwtSuccessResponse(pair))
	}
}

func refreshPair(a app.App) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req refreshRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, errorResponse(ErrEmptyRefresh))
			return
		}
		pair, err := a.Refresh(c, req.RefreshToken)
		if err != nil {
			handleError(c, err)
			return
		}
		c.JSON(http.StatusOK, jwtSuccessResponse(pair))
	}
}
