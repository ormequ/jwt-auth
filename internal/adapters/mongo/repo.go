package mongo

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"jwt-auth/internal/app"
)

type Repo struct {
	tokens *mongo.Collection
}

func (r Repo) CreateOrUpdate(ctx context.Context, userID string, hash string) error {
	const fn = "mongo.CreateOrUpdate"
	update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "hash", Value: hash}}}}
	opts := options.Update().SetUpsert(true)
	_, err := r.tokens.UpdateByID(ctx, userID, update, opts)
	if err != nil {
		return fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	return nil
}

type token struct {
	Hash string `json:"hash" bson:"hash"`
}

func (r Repo) GetTokenByID(ctx context.Context, userID string) (string, error) {
	const fn = "mongo.GetTokenByID"
	res := r.tokens.FindOne(ctx, bson.M{"_id": userID})
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return "", app.ErrNotFound
	}
	if err := res.Err(); err != nil {
		return "", fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	tok := token{}
	if err := res.Decode(&tok); err != nil {
		return "", fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	return tok.Hash, nil
}

func New(db *mongo.Database) Repo {
	return Repo{tokens: db.Collection("tokens")}
}
