package tests

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

type accessToken struct {
	User    string
	Refresh string
}

func decodeAccess(token string) (accessToken, error) {
	c, err := decodeToken([]byte(accessSecret), token)
	if err != nil {
		return accessToken{}, err
	}
	return accessToken{
		User:    c["usr"].(string),
		Refresh: c["ref"].(string),
	}, nil
}

type refreshToken struct {
	User string
}

func decodeRefresh(token string) (refreshToken, error) {
	c, err := decodeToken([]byte(refreshSecret), token)
	if err != nil {
		return refreshToken{}, err
	}
	return refreshToken{User: c["usr"].(string)}, nil
}

func decodeToken(secret []byte, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return secret, nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, err
	}
	return nil, err
}
