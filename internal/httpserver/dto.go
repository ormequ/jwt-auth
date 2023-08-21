package httpserver

import (
	"github.com/gin-gonic/gin"
	"jwt-auth/internal/app"
)

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type generateRequest struct {
	ID string `json:"id" binding:"required"`
}

type JWTPairResponse struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	AccessToken  string `json:"access_token" binding:"required"`
}

func jwtPairToResponse(pair app.JWTPair) JWTPairResponse {
	return JWTPairResponse{
		RefreshToken: pair.Refresh,
		AccessToken:  pair.Access,
	}
}

func errorResponse(err error) gin.H {
	return gin.H{
		"data":  nil,
		"error": err,
	}
}

func jwtSuccessResponse(pair app.JWTPair) gin.H {
	return gin.H{
		"data":  jwtPairToResponse(pair),
		"error": nil,
	}
}
