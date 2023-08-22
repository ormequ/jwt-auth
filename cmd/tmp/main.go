package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Замените на свой секретный ключ
var secretKey = []byte("your-secret-key")

func generateTokenString(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func main() {
	// Создание токена с пользовательскими утверждениями
	claims := jwt.MapClaims{
		"usr": "1234567890123456789012345678901234567890",
		"exp": time.Now().Add(time.Hour * 1).Unix(), // Срок действия 1 час
	}
	tokenString, err := generateTokenString(&claims)
	if err != nil {
		fmt.Println("Ошибка создания JWT токена:", err)
		return
	}
	fmt.Println("JWT токен:", tokenString)
}
