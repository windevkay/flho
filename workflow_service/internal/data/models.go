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
	Runs       RunModelInterface
	Workflows  WorkflowModelInterface
}

func GetModels(client *mongo.Client, dbName string) Models {
	return Models{
		Identities: NewIdentityModel(client, dbName),
		Runs:       NewRunModel(client, dbName),
		Workflows:  NewWorkflowModel(client, dbName),
	}
}

func GetMockModels() Models {
	return Models{}
}
