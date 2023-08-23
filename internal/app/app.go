package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"jwt-auth/internal/entities"
	"jwt-auth/internal/logger"
	"log/slog"
	"math/big"
	"regexp"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.32.4 --name=Repo
type Repo interface {
	CreateOrUpdate(ctx context.Context, token entities.RefreshToken) error
	GetTokenByID(ctx context.Context, userID string) (entities.RefreshToken, error)
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
	accessExpires  time.Duration
	refreshExpires time.Duration
}

func randomToken() string {
	const minLen = 10
	const maxLen = 72
	l, _ := rand.Int(rand.Reader, big.NewInt((maxLen-minLen)+minLen))
	b := make([]byte, l.Int64())
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return string(b)
}

func isValidUUID(input string) bool {
	uuidPattern := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	match, _ := regexp.MatchString(uuidPattern, input)
	return match
}

func (a App) GeneratePair(ctx context.Context, userID string) (entities.JWTPair, error) {
	const fn = "app.GeneratePair"

	log := logger.Log(ctx).With(
		slog.String("fn", fn),
		slog.String("userID", userID),
	)

	log.Debug("validating user ID")
	valid := isValidUUID(userID)
	if !valid {
		return entities.JWTPair{}, ErrInvalidUserID
	}

	log.Debug("generating refresh token")
	now := time.Now().UTC()
	refresh := randomToken()
	hashRefresh, err := a.hasher.Generate(ctx, refresh)
	if err != nil {
		return entities.JWTPair{}, err
	}
	log.Debug("generated", slog.String("token", refresh), slog.String("hash", hashRefresh))

	err = a.repo.CreateOrUpdate(ctx, entities.NewRefresh(userID, hashRefresh, now.Add(a.refreshExpires)))
	if err != nil {
		return entities.JWTPair{}, err
	}

	log.Debug("generating access token")
	access := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": userID,
		"exp": now.Add(a.accessExpires).Unix(),
	})
	strAccess, err := access.SignedString(a.accessSecret)
	if err != nil {
		return entities.JWTPair{}, fmt.Errorf("fn=%s err='%v'", fn, err)
	}

	b64Refresh := base64.StdEncoding.EncodeToString([]byte(refresh))
	return entities.NewPair(strAccess, b64Refresh), nil
}

func decodeToken(secret []byte, tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrIncorrectToken
		}

		return secret, nil
	})

	if token == nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return claims, err
	}
	return nil, err
}

func (a App) Refresh(ctx context.Context, access string, b64Refresh string) (entities.JWTPair, error) {
	const fn = "app.Refresh"

	log := logger.Log(ctx).With(slog.String("fn", fn))
	log.Debug("decoding", slog.String("refresh_in_base64", b64Refresh), slog.String("access", access))
	refresh, err := base64.StdEncoding.DecodeString(b64Refresh)
	if err != nil {
		return entities.JWTPair{}, ErrIncorrectToken
	}

	claims, err := decodeToken(a.accessSecret, access)
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return entities.JWTPair{}, ErrIncorrectToken
		}
		return entities.JWTPair{}, err
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return entities.JWTPair{}, ErrIncorrectToken
	}

	token, err := a.repo.GetTokenByID(ctx, userID)
	if err != nil {
		return entities.JWTPair{}, err
	}
	if token.Expires.Before(time.Now().UTC()) {
		return entities.JWTPair{}, ErrExpired
	}

	log.Debug("comparing")
	err = a.hasher.Compare(ctx, token.Hash, string(refresh))
	if err != nil {
		return entities.JWTPair{}, err
	}
	log.Debug("generating new pair")
	return a.GeneratePair(ctx, userID)
}

func New(repo Repo, hasher Hasher, accessSecret string, accessExpires time.Duration, refreshExpires time.Duration) App {
	return App{
		repo:           repo,
		hasher:         hasher,
		accessSecret:   []byte(accessSecret),
		accessExpires:  accessExpires,
		refreshExpires: refreshExpires,
	}
}
