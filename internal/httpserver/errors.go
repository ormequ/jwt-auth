package httpserver

import (
	"errors"
	"github.com/gin-gonic/gin"
	"jwt-auth/internal/app"
	"jwt-auth/internal/logger"
	"log/slog"
	"net/http"
)

var (
	ErrInternal     = errors.New("internal server error")
	ErrEmptyID      = errors.New("user id is required")
	ErrEmptyRefresh = errors.New("refresh token is required")
)

func hideError(err error) (int, error) {
	if errors.Is(err, app.ErrNotFound) {
		return http.StatusNotFound, err
	}
	if errors.Is(err, app.ErrPermissionDenied) || errors.Is(err, app.ErrExpired) {
		return http.StatusForbidden, err
	}
	return http.StatusInternalServerError, ErrInternal
}

func handleError(c *gin.Context, err error) {
	code, hidden := hideError(err)
	if errors.Is(hidden, ErrInternal) {
		logger.Log(c).Error("internal server error", slog.String("error", err.Error()))
	}
	c.JSON(code, errorResponse(hidden))
}
