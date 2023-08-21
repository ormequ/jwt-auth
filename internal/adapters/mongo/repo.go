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
	"jwt-auth/internal/entities"
)

type Repo struct {
	tokens *mongo.Collection
}

func (r Repo) CreateOrUpdate(ctx context.Context, token entities.RefreshToken) error {
	const fn = "mongo.CreateOrUpdate"
	exp := primitive.NewDateTimeFromTime(token.Expires)
	update := bson.D{{"$set", bson.D{{"token", token.Token}, {"expires", exp}}}}
	opts := options.Update().SetUpsert(true)
	_, err := r.tokens.UpdateByID(ctx, token.UserID, update, opts)
	if err != nil {
		return fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	return nil
}

type token struct {
	Hash    string             `json:"token" bson:"token"`
	Expires primitive.DateTime `json:"expires" bson:"expires"`
}

func (r Repo) GetToken(ctx context.Context, userID string) (entities.RefreshToken, error) {
	const fn = "mongo.GetToken"
	res := r.tokens.FindOne(ctx, bson.M{"_id": userID})
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return entities.RefreshToken{}, app.ErrNotFound
	}
	if err := res.Err(); err != nil {
		return entities.RefreshToken{}, fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	tok := token{}
	if err := res.Decode(&tok); err != nil {
		return entities.RefreshToken{}, fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	return entities.NewRefresh(userID, tok.Hash, tok.Expires.Time()), nil
}

func New(db *mongo.Database) Repo {
	return Repo{tokens: db.Collection("tokens")}
}
