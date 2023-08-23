package tests

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func decodeAccess(token string) (string, error) {
	c, err := decodeToken([]byte(accessSecret), token)
	if err != nil {
		return "", err
	}
	return c["sub"].(string), nil
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

func encodeToken(userID string, exp time.Duration, secret []byte) string {
	access := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().UTC().Add(exp).Unix(),
	})
	res, _ := access.SignedString(secret)
	return res
}
