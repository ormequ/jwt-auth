package app

import "context"

type Repo interface {
	CreateOrUpdate(ctx context.Context, id string, hash string) error
	GetUserID(ctx context.Context, token string) (string, error)
}

type Hasher interface {
	Generate(ctx context.Context, token string) (string, error)
}

type App struct {
	repo   Repo
	hasher Hasher
}

func (a App) GeneratePair(ctx context.Context, id string) (JWTPair, error) {
	return JWTPair{}, nil
}

func (a App) Refresh(ctx context.Context, token string) (JWTPair, error) {
	return JWTPair{}, nil
}

func New(repo Repo, hasher Hasher) App {
	return App{
		repo:   repo,
		hasher: hasher,
	}
}
