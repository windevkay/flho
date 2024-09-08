package data

import (
	"context"
	"database/sql"
	"time"
)

type State struct {
	ID         int64      `json:"-"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	WorkflowId int64      `json:"workflowId"`
	Name       string     `json:"name"`
	RetryUrl   string     `json:"retryUrl"`
	RetryAfter Timeout    `json:"retryAfter"`
}

type StateModelInterface interface {
	Insert(state *State, workflowId int64) error
}

type StateModel struct {
	DB *sql.DB
}

func (s StateModel) Insert(state *State, workflowId int64) error {
	query := `INSERT INTO states (workflow_id, name, retryurl, retryafter)
				VALUES ($1, $2, $3, $4)
				RETURNING created_at, workflow_id`

	args := []any{workflowId, state.Name, state.RetryUrl, state.RetryAfter}

	// db operations have 3 seconds max to resolve
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return s.DB.QueryRowContext(ctx, query, args...).Scan(&state.CreatedAt, &state.WorkflowId)
}
