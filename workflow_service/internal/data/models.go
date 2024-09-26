package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Workflows  WorkflowModelInterface
	States     StateModelInterface
	Identities IdentityModelInterface
}

func GetModels(db *sql.DB) Models {
	return Models{
		Workflows:  WorkflowModel{DB: db},
		States:     StateModel{DB: db},
		Identities: IdentityModel{DB: db},
	}
}

func GetMockModels() Models {
	return Models{
		Workflows: MockWorkflowModel{},
	}
}
