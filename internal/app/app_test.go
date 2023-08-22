package app

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"jwt-auth/internal/app/mocks"
	"jwt-auth/internal/entities"
	"log/slog"
	"testing"
	"time"
)

var accessSecret = []byte("test-access-secret")
var refreshSecret = []byte("test-refresh-secret")

//nolint:all
var ctx = context.WithValue(context.Background(), "log", slog.Default())

func hasherGenerate(t *testing.T) Hasher {
	h := mocks.NewHasher(t)
	h.
		On("Generate", mock.Anything, mock.AnythingOfType("string")).
		Return(func(_ context.Context, token string) (string, error) {
			return token + "-hash", nil
		})
	return h
}

func hasherCompare(t *testing.T) Hasher {
	h := mocks.NewHasher(t)
	h.
		On("Compare", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(ErrPermissionDenied)
	return h
}

func hasherCompareGenerate(t *testing.T) Hasher {
	h := mocks.NewHasher(t)
	h.
		On("Compare", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(nil)
	h.
		On("Generate", mock.Anything, mock.AnythingOfType("string")).
		Return(func(_ context.Context, token string) (string, error) {
			return token + "-hash", nil
		})
	return h
}
func repoGetTokenByID(t *testing.T, refreshToken string) Repo {
	r := mocks.NewRepo(t)
	r.
		On("GetTokenByID", mock.Anything, mock.AnythingOfType("string")).
		Return(func(_ context.Context, userID string) (string, error) {
			if userID == "not-found" {
				return "", ErrNotFound
			}
			return refreshToken, nil
		})

	return r
}

func repoGetTokenByIDCreateOrUpdate(t *testing.T, refreshToken string) Repo {
	r := mocks.NewRepo(t)
	r.
		On("GetTokenByID", mock.Anything, mock.AnythingOfType("string")).
		Return(refreshToken, nil)
	r.
		On("CreateOrUpdate", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(nil)

	return r
}

func repoCreateOrUpdate(t *testing.T) Repo {
	r := mocks.NewRepo(t)
	r.
		On("CreateOrUpdate", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Return(nil)

	return r
}

func TestApp_GeneratePair(t *testing.T) {
	type fields struct {
		repo           Repo
		hasher         Hasher
		accessSecret   []byte
		refreshSecret  []byte
		accessExpires  time.Duration
		refreshExpires time.Duration
	}
	type args struct {
		ctx    context.Context
		userID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.ValueAssertionFunc
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "correct generating",
			fields: fields{
				repo:           repoCreateOrUpdate(t),
				hasher:         hasherGenerate(t),
				accessSecret:   accessSecret,
				refreshSecret:  refreshSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:    ctx,
				userID: "test-id",
			},
			want: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				p, _ := i.(entities.JWTPair)
				claims, err := decodeToken(accessSecret, p.Access)
				return assert.NoError(t, err) &&
					assert.Equal(t, "test-id", claims["usr"].(string)) &&
					assert.Equal(t, getSignature(p.Refresh), claims["ref"].(string))
			},
			wantErr: nil,
		},
		{
			name: "expired access token",
			fields: fields{
				repo:           repoCreateOrUpdate(t),
				hasher:         hasherGenerate(t),
				accessSecret:   accessSecret,
				refreshSecret:  refreshSecret,
				accessExpires:  -time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:    ctx,
				userID: "test-id",
			},
			want: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				p, _ := i.(entities.JWTPair)
				_, err := decodeToken(accessSecret, p.Access)
				return assert.ErrorIs(t, err, jwt.ErrTokenExpired)
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := App{
				repo:           tt.fields.repo,
				hasher:         tt.fields.hasher,
				accessSecret:   tt.fields.accessSecret,
				refreshSecret:  tt.fields.refreshSecret,
				accessExpires:  tt.fields.accessExpires,
				refreshExpires: tt.fields.refreshExpires,
			}
			got, err := a.GeneratePair(tt.args.ctx, tt.args.userID)
			if tt.wantErr != nil && tt.wantErr(t, err, fmt.Sprintf("GeneratePair(%v, %v)", tt.args.ctx, tt.args.userID)) {
				return
			} else if tt.wantErr == nil {
				require.NoError(t, err)
			}
			require.True(t, tt.want(t, got, fmt.Sprintf("GeneratePair(%v, %v)", tt.args.ctx, tt.args.userID)))
		})
	}
}

func generateRefresh(userID string, exp time.Time) string {
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"usr": userID,
		"exp": exp.Unix(),
	})
	res, _ := refresh.SignedString(refreshSecret)
	return res
}

func TestApp_Refresh(t *testing.T) {
	type fields struct {
		repo           Repo
		hasher         Hasher
		accessSecret   []byte
		refreshSecret  []byte
		accessExpires  time.Duration
		refreshExpires time.Duration
	}
	type args struct {
		ctx     context.Context
		refresh string
	}
	refCorrect := generateRefresh("correct-user", time.Now().UTC().Add(time.Minute))
	refNotFound := generateRefresh("not-found", time.Now().UTC().Add(time.Minute))
	refExpired := generateRefresh("expired", time.Now().UTC().Add(-time.Minute))
	fmt.Println(refExpired)
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.ValueAssertionFunc
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "correct refreshing",
			fields: fields{
				repo:           repoGetTokenByIDCreateOrUpdate(t, refCorrect),
				hasher:         hasherCompareGenerate(t),
				accessSecret:   accessSecret,
				refreshSecret:  refreshSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:     ctx,
				refresh: refCorrect,
			},
			want: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				p, _ := i.(entities.JWTPair)
				claims, err := decodeToken(accessSecret, p.Access)
				return assert.NoError(t, err) &&
					assert.Equal(t, "correct-user", claims["usr"].(string)) &&
					assert.Equal(t, getSignature(p.Refresh), claims["ref"].(string))
			},
			wantErr: nil,
		},
		{
			name: "another token in database (hasher.Compare returns error)",
			fields: fields{
				repo:           repoGetTokenByID(t, refCorrect),
				hasher:         hasherCompare(t),
				accessSecret:   accessSecret,
				refreshSecret:  refreshSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:     ctx,
				refresh: refCorrect,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, ErrPermissionDenied, err)
			},
		},
		{
			name: "user not found",
			fields: fields{
				repo:           repoGetTokenByID(t, refNotFound),
				hasher:         nil,
				accessSecret:   accessSecret,
				refreshSecret:  refreshSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:     ctx,
				refresh: refNotFound,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, ErrNotFound, err)
			},
		},
		{
			name: "token expired",
			fields: fields{
				repo:           nil,
				hasher:         nil,
				accessSecret:   accessSecret,
				refreshSecret:  refreshSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:     ctx,
				refresh: refExpired,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, ErrExpired, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := App{
				repo:           tt.fields.repo,
				hasher:         tt.fields.hasher,
				accessSecret:   tt.fields.accessSecret,
				refreshSecret:  tt.fields.refreshSecret,
				accessExpires:  tt.fields.accessExpires,
				refreshExpires: tt.fields.refreshExpires,
			}
			got, err := a.Refresh(tt.args.ctx, tt.args.refresh)
			if tt.wantErr != nil && tt.wantErr(t, err, fmt.Sprintf("Refresh(%v, %v)", tt.args.ctx, tt.args.refresh)) {
				return
			} else if tt.wantErr == nil {
				require.NoError(t, err)
			}
			require.True(t, tt.want(t, got, fmt.Sprintf("Refresh(%v, %v)", tt.args.ctx, tt.args.refresh)))
		})
	}
}
