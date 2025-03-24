package data

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrRecordNotFound   = errors.New("record not found")
	ErrEditConflict     = errors.New("edit conflict")
	ErrValidationFailed = errors.New("validation failed")
	ErrDuplicateEmail   = errors.New("duplicate email")
)

type Models struct {
	Identities IdentityModelInterface
	Tokens     TokenModelInterface
	Runs       RunModelInterface
	Workflows  WorkflowModelInterface
}

func GetModels(client *mongo.Client, dbName string) Models {
	return Models{
		Identities: NewIdentityModel(client, dbName),
		Tokens:     NewTokenModel(client, dbName),
		Runs:       NewRunModel(client, dbName),
		Workflows:  NewWorkflowModel(client, dbName),
	}
}
