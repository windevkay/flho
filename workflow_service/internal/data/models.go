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
	Workflows  WorkflowModelInterface
	Identities IdentityModelInterface
}

func GetModels(client *mongo.Client, dbName string) Models {
	return Models{
		Workflows:  NewWorkflowModel(client, dbName),
		Identities: NewIdentityModel(client, dbName),
	}
}

func GetMockModels() Models {
	return Models{}
}
