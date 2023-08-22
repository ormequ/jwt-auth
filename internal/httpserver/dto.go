package httpserver

import (
	"github.com/gin-gonic/gin"
	"jwt-auth/internal/entities"
)

type RefreshRequest struct {
	RefreshToken string `json:"token" binding:"required"`
}

type GenerateRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type JWTPairResponse struct {
	Refresh string `json:"refresh_token" binding:"required"`
	Access  string `json:"access_token" binding:"required"`
}

func jwtPairToResponse(pair entities.JWTPair) JWTPairResponse {
	return JWTPairResponse{
		Refresh: pair.Refresh,
		Access:  pair.Access,
	}
}

func errorResponse(err error) gin.H {
	return gin.H{
		"data":  nil,
		"error": err.Error(),
	}
}

func jwtSuccessResponse(pair entities.JWTPair) gin.H {
	return gin.H{
		"data":  jwtPairToResponse(pair),
		"error": nil,
	}
}
