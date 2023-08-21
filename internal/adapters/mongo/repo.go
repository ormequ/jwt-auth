package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jwt-auth/internal/app"
)

type Repo struct {
	tokens *mongo.Collection
}

func (r Repo) CreateOrUpdate(ctx context.Context, id string, hash string) error {
	const fn = "mongo.CreateOrUpdate"
	update := bson.D{{"$set", bson.D{{"token", hash}}}}
	opts := options.Update().SetUpsert(true)
	_, err := r.tokens.UpdateByID(ctx, id, update, opts)
	return fmt.Errorf("fn=%s err='%v'", fn, err)
}

type user struct {
	ID string `json:"_id" bson:"_id"`
}

func (r Repo) GetUserID(ctx context.Context, token string) (string, error) {
	const fn = "mongo.GetToken"
	res := r.tokens.FindOne(ctx, bson.M{"token": token})
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return "", app.ErrNotFound
	}
	if err := res.Err(); err != nil {
		return "", fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	u := user{}
	err := res.Decode(u)
	return u.ID, fmt.Errorf("fn=%s err='%v'", fn, err)
}

func New(db *mongo.Database) Repo {
	return Repo{tokens: db.Collection("tokens")}
}
