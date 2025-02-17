package data

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Identities IdentityModelInterface
	Tokens     TokenModelInterface
}

func GetModels(client *mongo.Client, dbName string) Models {
	return Models{
		Identities: NewIdentityModel(client, dbName),
		Tokens:     NewTokenModel(client, dbName),
	}
}
