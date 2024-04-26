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
	Workflows WorkflowModelInterface
	Users     UserModelInterface
}

func GetModels(db *sql.DB) Models {
	return Models{
		Workflows: WorkflowModel{DB: db},
		Users:     UserModel{DB: db},
	}
}

func GetMockModels() Models {
	return Models{
		Workflows: MockWorkflowModel{},
	}
}
