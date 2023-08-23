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
	update := bson.D{
		primitive.E{
			Key: "$set",
			Value: bson.D{
				primitive.E{Key: "hash", Value: token.Hash},
				primitive.E{Key: "expires", Value: primitive.NewDateTimeFromTime(token.Expires)},
			},
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.tokens.UpdateByID(ctx, token.UserID, update, opts)
	if err != nil {
		return fmt.Errorf("fn=%s err='%v'", fn, err)
	}
	return nil
}

type token struct {
	Hash    string             `json:"hash" bson:"hash"`
	Expires primitive.DateTime `json:"expires" bson:"expires"`
}

func (r Repo) GetTokenByID(ctx context.Context, userID string) (entities.RefreshToken, error) {
	const fn = "mongo.GetTokenByID"
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
