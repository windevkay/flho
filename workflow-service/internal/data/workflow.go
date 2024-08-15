package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/windevkay/flhoutils/validator"
)

type Workflow struct {
	ID             int64      `json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
	OrganizationId int64      `json:"organizationId"`
	UniqueID       string     `json:"uniqueId"`
	Name           string     `json:"name"`
	States         []State    `json:"states"`
	//StartState string     `json:"startState"`
	//EndState   string     `json:"endState"`
	//RetryWebhook string     `json:"retryWebhook,omitempty"`
	//RetryAfter   Timeout    `json:"retryAfter,omitempty"`
	Active  bool  `json:"active"`
	Version int32 `json:"version"`
}

func ValidateWorkflow(v *validator.Validator, w *Workflow) {
	v.Check(w.Name != "", "name", "must be provided")

	v.Check(len(w.States) >= 2, "states", "must have atleast 2 values")
	//v.Check(validator.Unique(w.States), "states", "must not contain duplicate values")

	// v.Check(w.StartState != "", "startState", "must be provided")
	// v.Check(slices.Contains(w.States, w.StartState), "startState", "must be part of the states list")

	// v.Check(w.EndState != "", "endState", "must be provided")
	// v.Check(slices.Contains(w.States, w.EndState), "endState", "must be part of the states list")
}

type WorkflowModelInterface interface {
	InsertWithTx(workflow *Workflow) error
	Get(id int64) (*Workflow, error)
	GetAll(name string, states []string, filters Filters) ([]*Workflow, Metadata, error)
	Update(workflow *Workflow) error
	Delete(id int64) error
}

type WorkflowModel struct {
	DB *sql.DB
}

func (w WorkflowModel) InsertWithTx(workflow *Workflow) error {
	query := `INSERT INTO workflows (organizationid, uniqueid, name, active)
				VALUES ($1, $2, $3, $4)
				RETURNING id, created_at, updated_at, version`

	args := []any{workflow.OrganizationId, workflow.UniqueID, workflow.Name, workflow.Active}

	// db operations have 3 seconds max to resolve
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Start a transaction using the context
	tx, err := w.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Ensure the transaction is rolled back in case of an error
	defer tx.Rollback()

	err = tx.QueryRow(query, args...).Scan(&workflow.ID, &workflow.CreatedAt, &workflow.UpdatedAt, &workflow.Version)
	if err != nil {
		return err
	}

	// insert states for workflow
	for _, state := range workflow.States {
		query := `INSERT INTO states (workflowid, name, retryurl, retryafter)
                    VALUES ($1, $2, $3, $4)
                    RETURNING created_at`

		args := []any{workflow.ID, state.Name, state.RetryUrl, state.RetryAfter}

		err := tx.QueryRow(query, args...).Scan(&state.CreatedAt)
		if err != nil {
			return err
		}
	}

	// commit the transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (w WorkflowModel) Get(id int64) (*Workflow, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
        SELECT 
            w.id, w.uniqueid, w.name, w.active, w.created_at, w.updated_at, w.version,
            s.id, s.name, s.retryurl, s.retryafter, s.created_at
        FROM 
            workflows w
        LEFT JOIN 
            states s ON w.id = s.workflowid
        WHERE 
            w.id = $1`

	rows, err := w.DB.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflow Workflow
	workflow.States = []State{}

	// Iterate over the rows
	for rows.Next() {
		var state State
		err := rows.Scan(
			&workflow.ID,
			&workflow.UniqueID,
			&workflow.Name,
			&workflow.Active,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
			&workflow.Version,
			&state.ID,
			&state.Name,
			&state.RetryUrl,
			&state.RetryAfter,
			&state.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		workflow.States = append(workflow.States, state)
	}

	// Check for errors from iterating over rows
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &workflow, nil
}

func (w WorkflowModel) GetAll(name string, states []string, filters Filters) ([]*Workflow, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, updated_at, uniqueid, name, states, startstate, endstate, retrywebhook, retryafter, active, version 
		FROM workflows
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (states @> $2 OR $2 = '{}') 
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{name, pq.Array(states), filters.limit(), filters.offest()}

	rows, err := w.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	var workflows []*Workflow

	for rows.Next() {
		var workflow Workflow

		err := rows.Scan(
			&totalRecords, // scanned count
			&workflow.ID,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
			&workflow.UniqueID,
			&workflow.Name,
			pq.Array(&workflow.States),
			&workflow.StartState,
			&workflow.EndState,
			&workflow.RetryWebhook,
			&workflow.RetryAfter,
			&workflow.Active,
			&workflow.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		workflows = append(workflows, &workflow)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return workflows, metadata, nil
}

func (w WorkflowModel) Update(workflow *Workflow) error {
	query := `
		UPDATE workflows
		SET updated_at = NOW(), name = $1, states = $2, startstate = $3, endstate = $4, retrywebhook = $5, retryafter = $6, version = version + 1
		WHERE id = $7 AND version = $8
		RETURNING version`

	args := []any{
		workflow.Name,
		pq.Array(workflow.States),
		workflow.StartState,
		workflow.EndState,
		workflow.RetryWebhook,
		workflow.RetryAfter,
		workflow.ID,
		workflow.Version,
	}

	// db operations have 3 seconds max to resolve
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := w.DB.QueryRowContext(ctx, query, args...).Scan(&workflow.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (w WorkflowModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM workflows WHERE id = $1`

	// db operations have 3 seconds max to resolve
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := w.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsDeleted == 0 {
		return ErrRecordNotFound
	}

	return nil
}
