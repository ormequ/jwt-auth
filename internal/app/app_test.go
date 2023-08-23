package app

import (
	"context"
	"encoding/base64"
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

const userIDNotFound = "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
const userIDDefault = "f47ac10b-58cc-4372-a567-0e02b2c3d479"

var accessSecret = []byte("test-access-secret")

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

func repoGetTokenByID(t *testing.T, hash string, exp time.Time) Repo {
	r := mocks.NewRepo(t)
	r.
		On("GetTokenByID", mock.Anything, mock.AnythingOfType("string")).
		Return(func(_ context.Context, userID string) (entities.RefreshToken, error) {
			if userID == userIDNotFound {
				return entities.RefreshToken{}, ErrNotFound
			}
			return entities.RefreshToken{UserID: userID, Hash: hash, Expires: exp}, nil
		})

	return r
}

func repoGetTokenByIDCreateOrUpdate(t *testing.T, hash string, exp time.Time) Repo {
	r := mocks.NewRepo(t)
	r.
		On("GetTokenByID", mock.Anything, mock.AnythingOfType("string")).
		Return(func(_ context.Context, userID string) (entities.RefreshToken, error) {
			return entities.RefreshToken{UserID: userID, Hash: hash, Expires: exp}, nil
		})
	r.
		On("CreateOrUpdate", mock.Anything, mock.AnythingOfType("entities.RefreshToken")).
		Return(nil)

	return r
}

func repoCreateOrUpdate(t *testing.T) Repo {
	r := mocks.NewRepo(t)
	r.
		On("CreateOrUpdate", mock.Anything, mock.AnythingOfType("entities.RefreshToken")).
		Return(nil)

	return r
}

func TestApp_GeneratePair(t *testing.T) {
	type fields struct {
		repo           Repo
		hasher         Hasher
		accessSecret   []byte
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
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:    ctx,
				userID: userIDDefault,
			},
			want: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				p, _ := i.(entities.JWTPair)
				claims, err := decodeToken(accessSecret, p.Access)
				return assert.NoError(t, err) && assert.Equal(t, userIDDefault, claims["sub"].(string))
			},
			wantErr: nil,
		},
		{
			name: "expired access token",
			fields: fields{
				repo:           repoCreateOrUpdate(t),
				hasher:         hasherGenerate(t),
				accessSecret:   accessSecret,
				accessExpires:  -time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:    ctx,
				userID: userIDDefault,
			},
			want: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				p, _ := i.(entities.JWTPair)
				_, err := decodeToken(accessSecret, p.Access)
				return assert.ErrorIs(t, err, jwt.ErrTokenExpired)
			},
			wantErr: nil,
		},
		{
			name: "incorrect user id",
			fields: fields{
				repo:           nil,
				hasher:         nil,
				accessSecret:   accessSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:    ctx,
				userID: "non-uuid-string",
			},
			want: nil,
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrInvalidUserID)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := App{
				repo:           tt.fields.repo,
				hasher:         tt.fields.hasher,
				accessSecret:   tt.fields.accessSecret,
				accessExpires:  tt.fields.accessExpires,
				refreshExpires: tt.fields.refreshExpires,
			}
			got, err := a.GeneratePair(tt.args.ctx, tt.args.userID)
			if tt.wantErr != nil && tt.wantErr(t, err, fmt.Sprintf("GeneratePair(%v, %v)", tt.args.ctx, tt.args.userID)) {
				return
			} else if tt.wantErr == nil {
				require.NoError(t, err)
			}
			if tt.want != nil {
				require.True(t, tt.want(t, got, fmt.Sprintf("GeneratePair(%v, %v)", tt.args.ctx, tt.args.userID)))
			}
		})
	}
}

func generateAccess(userID string, exp time.Time) string {
	access := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": userID,
		"exp": exp.Unix(),
	})
	res, _ := access.SignedString(accessSecret)
	return res
}

func TestApp_Refresh(t *testing.T) {
	type fields struct {
		repo           Repo
		hasher         Hasher
		accessSecret   []byte
		accessExpires  time.Duration
		refreshExpires time.Duration
	}
	type args struct {
		ctx     context.Context
		access  string
		refresh string
	}
	accCorrect := generateAccess(userIDDefault, time.Now().UTC().Add(time.Minute))
	accNotFound := generateAccess(userIDNotFound, time.Now().UTC().Add(time.Minute))
	accExpired := generateAccess(userIDDefault, time.Now().UTC().Add(-time.Minute))
	refresh := base64.StdEncoding.EncodeToString([]byte(randomToken()))
	refreshHash := refresh + "-hash"
	now := time.Now().UTC()
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
				repo:           repoGetTokenByIDCreateOrUpdate(t, refreshHash, now.Add(time.Minute)),
				hasher:         hasherCompareGenerate(t),
				accessSecret:   accessSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:     ctx,
				access:  accCorrect,
				refresh: refresh,
			},
			want: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				p, _ := i.(entities.JWTPair)
				claims, err := decodeToken(accessSecret, p.Access)
				return assert.NoError(t, err) && assert.Equal(t, userIDDefault, claims["sub"].(string))
			},
			wantErr: nil,
		},
		{
			name: "another token in database (hasher.Compare returns error)",
			fields: fields{
				repo:           repoGetTokenByID(t, refreshHash, now.Add(time.Minute)),
				hasher:         hasherCompare(t),
				accessSecret:   accessSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:     ctx,
				access:  accCorrect,
				refresh: refresh,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, ErrPermissionDenied, err)
			},
		},
		{
			name: "user not found",
			fields: fields{
				repo:           repoGetTokenByID(t, refreshHash, now.Add(time.Minute)),
				hasher:         nil,
				accessSecret:   accessSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:     ctx,
				access:  accNotFound,
				refresh: refresh,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, ErrNotFound, err)
			},
		},
		{
			name: "incorrect access",
			fields: fields{
				repo:           nil,
				hasher:         nil,
				accessSecret:   accessSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:     ctx,
				access:  "adsfasfasfd",
				refresh: refresh,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, ErrIncorrectToken, err)
			},
		},
		{
			name: "refresh token expired",
			fields: fields{
				repo:           repoGetTokenByID(t, refreshHash, now.Add(-time.Minute)),
				hasher:         nil,
				accessSecret:   accessSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:     ctx,
				access:  accCorrect,
				refresh: refresh,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, ErrExpired, err)
			},
		},
		{
			name: "access token expired",
			fields: fields{
				repo:           repoGetTokenByIDCreateOrUpdate(t, refreshHash, now.Add(time.Minute)),
				hasher:         hasherCompareGenerate(t),
				accessSecret:   accessSecret,
				accessExpires:  time.Minute,
				refreshExpires: time.Minute,
			},
			args: args{
				ctx:     ctx,
				access:  accExpired,
				refresh: base64.StdEncoding.EncodeToString([]byte(randomToken())),
			},
			want: func(t assert.TestingT, i interface{}, i2 ...interface{}) bool {
				p, _ := i.(entities.JWTPair)
				claims, err := decodeToken(accessSecret, p.Access)
				return assert.NoError(t, err) && assert.Equal(t, userIDDefault, claims["sub"].(string))
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
				accessExpires:  tt.fields.accessExpires,
				refreshExpires: tt.fields.refreshExpires,
			}
			got, err := a.Refresh(tt.args.ctx, tt.args.access, tt.args.refresh)
			if tt.wantErr != nil && tt.wantErr(t, err, fmt.Sprintf("Refresh(%v, %v)", tt.args.ctx, tt.args.refresh)) {
				return
			} else if tt.wantErr == nil {
				require.NoError(t, err)
			}
			require.True(t, tt.want(t, got, fmt.Sprintf("Refresh(%v, %v)", tt.args.ctx, tt.args.refresh)))
		})
	}
}
