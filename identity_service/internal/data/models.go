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
	Identities IdentityModelInterface
	Tokens     TokenModelInterface
}

func GetModels(db *sql.DB) Models {
	return Models{
		Identities: IdentityModel{DB: db},
		Tokens:     TokenModel{DB: db},
	}
}

func GetMockModels() Models {
	return Models{}
}
