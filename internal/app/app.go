package app

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"jwt-auth/internal/entities"
	"jwt-auth/internal/logger"
	"log/slog"
	"math/rand"
	"time"
)

//go:generate go run github.com/vektra/mockery/v2@v2.32.4 --name=Repo
type Repo interface {
	CreateOrUpdate(ctx context.Context, token entities.RefreshToken) error
	GetToken(ctx context.Context, userID string) (entities.RefreshToken, error)
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

func randomString() string {
	const sym = "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz()[]{}+-*/=_.,:;!?@#$%^&"
	const minLen = 10
	const maxLen = 72

	res := make([]byte, rand.Intn(maxLen-minLen)+minLen)
	symB := []byte(sym)
	for i := range res {
		res[i] = symB[rand.Intn(len(sym))]
	}
	return string(res)
}

func (a App) generateRefresh(ctx context.Context) (token string, hash string, err error) {
	for !errors.Is(err, ErrNotFound) {
		token = randomString()
		hash, err = a.hasher.Generate(ctx, token)
		if err != nil {
			return
		}
		_, err = a.repo.GetToken(ctx, hash)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return
		}
	}
	err = nil
	return
}

func (a App) GeneratePair(ctx context.Context, id string) (entities.JWTPair, error) {
	const fn = "app.GeneratePair"

	log := logger.Log(ctx).With(
		slog.String("fn", fn),
		slog.String("id", id),
	)

	log.Debug("generating refresh token")
	rawRefresh, hashRefresh, err := a.generateRefresh(ctx)
	log.Debug("generated", slog.String("token", rawRefresh), slog.String("hash", hashRefresh))

	if err != nil {
		return entities.JWTPair{}, err
	}
	b64Refresh := base64.StdEncoding.EncodeToString([]byte(rawRefresh))
	now := time.Now().UTC()
	refresh := entities.NewRefresh(id, hashRefresh, now.Add(a.refreshExpires))

	log.Debug("creating or updating refresh token")
	err = a.repo.CreateOrUpdate(ctx, refresh)
	if err != nil {
		return entities.JWTPair{}, err
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"ref": b64Refresh,
		"usr": id,
		"exp": now.Add(a.accessExpires),
	})
	log.Debug("signing access token")
	strAccess, err := access.SignedString(a.accessSecret)
	if err != nil {
		return entities.JWTPair{}, fmt.Errorf("fn=%s err='%v'", fn, err)
	}

	return entities.NewPair(strAccess, b64Refresh), nil
}

func (a App) Refresh(ctx context.Context, userID string, b64Refresh string) (entities.JWTPair, error) {
	const fn = "app.Refresh"

	log := logger.Log(ctx).With(slog.String("fn", fn))

	log.Debug("decoding", slog.String("encoded", b64Refresh))
	rawRefresh, err := base64.StdEncoding.DecodeString(b64Refresh)
	if err != nil {
		return entities.JWTPair{}, fmt.Errorf("fn=%s err='%v'", fn, err)
	}

	log = logger.Log(ctx).With(slog.String("token", string(rawRefresh)))
	log.Debug("getting token")
	refresh, err := a.repo.GetToken(ctx, userID)
	if err != nil {
		return entities.JWTPair{}, err
	}
	log.Debug("comparing")
	err = a.hasher.Compare(ctx, refresh.Token, string(rawRefresh))
	if err != nil {
		return entities.JWTPair{}, err
	}
	if refresh.Expires.Before(time.Now().UTC()) {
		return entities.JWTPair{}, ErrExpired
	}
	log.Debug("generating new pair")
	return a.GeneratePair(ctx, refresh.UserID)
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
