package data

import (
	"context"
	"database/sql"
	"time"
)

type IdentityModelInterface interface {
	GetIdentityId(uuid string) (int64, error)
}

type IdentityModel struct {
	DB *sql.DB
}

func (i IdentityModel) GetIdentityId(uuid string) (int64, error) {
	query := `SELECT id FROM workflow_identity_identities WHERE uuid = $1`

	var id int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := i.DB.QueryRowContext(ctx, query, uuid).Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}
