package httpserver

import (
	"errors"
	"github.com/gin-gonic/gin"
	"jwt-auth/internal/app"
	"net/http"
)

var ErrInternal = errors.New("internal server error")

func hideError(err error) (int, gin.H) {
	if errors.Is(err, app.ErrNotFound) {
		return http.StatusNotFound, errorResponse(err)
	}
	if errors.Is(err, app.ErrPermissionDenied) {
		return http.StatusForbidden, errorResponse(err)
	}
	return http.StatusInternalServerError, errorResponse(ErrInternal)
}
