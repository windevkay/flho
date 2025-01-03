package data

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Identity struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"created_at"`
	UUID      string             `bson:"uuid"`
}

type IdentityModelInterface interface {
	GetIdentityId(uuid string) (primitive.ObjectID, error)
}

type IdentityModel struct {
	Collection *mongo.Collection
}

func NewIdentityModel(client *mongo.Client, dbName string) IdentityModel {
	collection := client.Database(dbName).Collection("identities")
	return IdentityModel{
		Collection: collection,
	}
}

func (i IdentityModel) GetIdentityId(uuid string) (primitive.ObjectID, error) {
	var id primitive.ObjectID

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := i.Collection.FindOne(ctx, bson.M{"uuid": uuid}).Decode(&id)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return id, nil
}
