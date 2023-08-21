package bcrypt

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"jwt-auth/internal/app"
	"testing"
)

func TestBCrypt_Generate(t *testing.T) {
	type fields struct {
		Cost int
	}
	type args struct {
		ctx   context.Context
		token string
	}

	bgCtx := context.Background()
	canceledCtx, cancel := context.WithCancel(bgCtx)
	cancel()

	tests := [...]struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "correct generation",
			fields: fields{Cost: 10},
			args: args{
				ctx:   bgCtx,
				token: "qwe123!!~....утф😊",
			},
			want: hash("qwe123!!~....утф😊", 10),
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
		{
			name:   "canceled context",
			fields: fields{Cost: 10},
			args: args{
				ctx:   canceledCtx,
				token: "qwe123!!~....утф😊",
			},
			want: "",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BCrypt{
				cost: tt.fields.Cost,
			}
			got, err := b.Generate(tt.args.ctx, tt.args.token)
			if !tt.wantErr(t, err, fmt.Sprintf("Generate(%v, %v)", tt.args.ctx, tt.args.token)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Generate(%v, %v)", tt.args.ctx, tt.args.token)
		})
	}
}

func hash(s string, cost int) string {
	res, _ := bcrypt.GenerateFromPassword([]byte(s), cost)
	return string(res)
}

func TestBCrypt_Compare(t *testing.T) {
	type fields struct {
		Cost int
	}
	type args struct {
		ctx   context.Context
		hash  string
		token string
	}

	bgCtx := context.Background()
	canceledCtx, cancel := context.WithCancel(bgCtx)
	cancel()

	tests := [...]struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:   "correct comparison",
			fields: fields{Cost: 10},
			args: args{
				ctx:   bgCtx,
				hash:  hash("qwe123!!~....утф😊", 10),
				token: "qwe123!!~....утф😊",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
		{
			name:   "incorrect token",
			fields: fields{Cost: 10},
			args: args{
				ctx:   bgCtx,
				hash:  hash("qwe123!!~....утф😊", 10),
				token: "qwe123!!~....ут😊ф",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, app.ErrPermissionDenied)
			},
		},
		{
			name:   "context canceled",
			fields: fields{Cost: 10},
			args: args{
				ctx:   canceledCtx,
				hash:  hash("qwe123!!~....утф😊", 10),
				token: "qwe123!!~....утф😊",
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return assert.Error(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BCrypt{
				cost: tt.fields.Cost,
			}
			tt.wantErr(t, b.Compare(tt.args.ctx, tt.args.hash, tt.args.token), fmt.Sprintf("Compare(%v, %v, %v)", tt.args.ctx, tt.args.hash, tt.args.token))
		})
	}
}
