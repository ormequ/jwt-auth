package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"jwt-auth/internal/entities"
	"jwt-auth/internal/logger"
	"log/slog"
	"strings"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.32.4 --name=Repo
type Repo interface {
	CreateOrUpdate(ctx context.Context, userID string, hash string) error
	GetTokenByID(ctx context.Context, userID string) (string, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.32.4 --name=Hasher
type Hasher interface {
	Generate(ctx context.Context, token string) (string, error)
	Compare(ctx context.Context, hash string, token string) error
}

type App struct {
	repo           Repo
	hasher         Hasher
	accessSecret   []byte
	refreshSecret  []byte
	accessExpires  time.Duration
	refreshExpires time.Duration
}

func getSignature(token string) string {
	return strings.Split(token, ".")[2]
}

func (a App) GeneratePair(ctx context.Context, userID string) (entities.JWTPair, error) {
	const fn = "app.GeneratePair"

	log := logger.Log(ctx).With(
		slog.String("fn", fn),
		slog.String("userID", userID),
	)

	log.Debug("generating refresh token")
	now := time.Now().UTC()
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"usr": userID,
		"exp": now.Add(a.refreshExpires).Unix(),
	})
	strRefresh, err := refresh.SignedString(a.refreshSecret)
	if err != nil {
		return entities.JWTPair{}, fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	signRefresh := getSignature(strRefresh)
	hashRefresh, err := a.hasher.Generate(ctx, signRefresh)
	if err != nil {
		return entities.JWTPair{}, fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	log.Debug("generated", slog.String("token", strRefresh), slog.String("hash", hashRefresh))

	err = a.repo.CreateOrUpdate(ctx, userID, hashRefresh)
	if err != nil {
		return entities.JWTPair{}, err
	}

	log.Debug("generating access token")
	access := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"ref": signRefresh,
		"usr": userID,
		"exp": now.Add(a.accessExpires).Unix(),
	})
	strAccess, err := access.SignedString(a.accessSecret)
	if err != nil {
		return entities.JWTPair{}, fmt.Errorf("fn=%s err='%v'", fn, err)
	}

	return entities.NewPair(strAccess, strRefresh), nil
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

func (a App) Refresh(ctx context.Context, refresh string) (entities.JWTPair, error) {
	const fn = "app.Refresh"

	log := logger.Log(ctx).With(slog.String("fn", fn))
	log.Debug("decoding", slog.String("token", refresh))
	claims, err := decodeToken(a.refreshSecret, refresh)
	if errors.Is(err, jwt.ErrTokenExpired) {
		return entities.JWTPair{}, ErrExpired
	}
	userID := claims["usr"].(string)
	if err != nil {
		return entities.JWTPair{}, fmt.Errorf("fn=%s err='%v'", fn, err)
	}

	hash, err := a.repo.GetTokenByID(ctx, userID)
	if err != nil {
		return entities.JWTPair{}, err
	}
	log.Debug("comparing")
	err = a.hasher.Compare(ctx, hash, getSignature(refresh))
	if err != nil {
		return entities.JWTPair{}, err
	}
	log.Debug("generating new pair")
	return a.GeneratePair(ctx, userID)
}

func New(repo Repo, hasher Hasher, accessSecret string, refreshSecret string, accessExpires time.Duration, refreshExpires time.Duration) App {
	return App{
		repo:           repo,
		hasher:         hasher,
		accessSecret:   []byte(accessSecret),
		refreshSecret:  []byte(refreshSecret),
		accessExpires:  accessExpires,
		refreshExpires: refreshExpires,
	}
}
