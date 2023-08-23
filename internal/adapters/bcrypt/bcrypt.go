package bcrypt

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"jwt-auth/internal/app"
)

type BCrypt struct {
	cost int
}

func (b BCrypt) Generate(ctx context.Context, token string) (string, error) {
	const fn = "bcrypt.Generate"

	if ctx.Err() != nil {
		return "", fmt.Errorf("fn=%s err='%v'", fn, ctx.Err())
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(token), b.cost)
	if err != nil {
		return "", fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	return string(bytes), nil
}

func (b BCrypt) Compare(ctx context.Context, hash string, token string) error {
	const fn = "bcrypt.Compare"

	if ctx.Err() != nil {
		return fmt.Errorf("fn=%s err='%v'", fn, ctx.Err())
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return app.ErrPermissionDenied
	} else if err != nil {
		return fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	return nil
}

func New(cost int) BCrypt {
	return BCrypt{cost: cost}
}
