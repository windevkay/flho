package data

import (
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
	"github.com/windevkay/flho/internal/validator"
)

type CircuitBreakerStatus string

const (
	HALFOPEN CircuitBreakerStatus = "HALFOPEN"
	CLOSED   CircuitBreakerStatus = "CLOSED"
	OPEN     CircuitBreakerStatus = "OPEN"
)

type Workflow struct {
	ID                          int64                `json:"-"`
	CreatedAt                   time.Time            `json:"created_at"`
	UpdatedAt                   time.Time            `json:"updated_at,omitempty"`
	UniqueID                    string               `json:"uniqueId"`
	Name                        string               `json:"name"`
	States                      []string             `json:"states"`
	StartState                  string               `json:"startState"`
	EndState                    string               `json:"endState"`
	CallbackWebhook             string               `json:"webhook,omitempty"`
	Retry                       bool                 `json:"retry"`
	RetryAfter                  Timeout              `json:"retryAfter,omitempty"`
	RetryURL                    string               `json:"retryUrl,omitempty"`
	CircuitBreaker              bool                 `json:"circuitBreaker"`
	CircuitBreakerStatus        CircuitBreakerStatus `json:"-"`
	CircuitBreakerFailureCount  int32                `json:"circuitBreakerFailureCount,omitempty"`
	CircuitBreakerOpenTimeout   Timeout              `json:"circuitBreakerOpenTimeout,omitempty"`
	CircuitBreakerHalfOpenCount int32                `json:"circuitBreakerHalfOpenCount,omitempty"`
	Active                      bool                 `json:"active"`
	Version                     int32                `json:"version"`
}

func ValidateWorkflow(v *validator.Validator, w *Workflow) {
	v.Check(w.Name != "", "name", "must be provided")

	v.Check(len(w.States) >= 2, "states", "must have atleast 2 values")
	v.Check(validator.Unique(w.States), "states", "must not contain duplicate values")

	v.Check(w.StartState != "", "startState", "must be provided")
	v.Check(slices.Contains(w.States, w.StartState), "startState", "must be part of the states list")

	v.Check(w.EndState != "", "endState", "must be provided")
	v.Check(slices.Contains(w.States, w.EndState), "endState", "must be part of the states list")

	if w.Retry {
		v.Check(w.RetryURL != "", "retryUrl", "must be provided")
		v.Check(w.RetryAfter != 0, "retryAfter", "must be provided")
		v.Check(w.RetryAfter > 0, "retryAfter", "must be a non-negative value")
	}

	if w.CircuitBreaker {
		v.Check(w.CircuitBreakerFailureCount != 0, "circuitBreakerFailureCount", "must be provided")
		v.Check(w.CircuitBreakerFailureCount > 0, "circuitBreakerFailureCount", "must be a non-negative value")
		v.Check(w.CircuitBreakerOpenTimeout != 0, "circuitBreakerOpenTimeout", "must be provided")
		v.Check(w.CircuitBreakerOpenTimeout > 0, "circuitBreakerOpenTimeout", "must be a non-negative value")
		v.Check(w.CircuitBreakerHalfOpenCount != 0, "circuitBreakerHalfOpenCount", "must be provided")
		v.Check(w.CircuitBreakerHalfOpenCount > 0, "circuitBreakerHalfOpenCount", "must be a non-negative value")
	}
}

type WorkflowModel struct {
	DB *sql.DB
}

func (w WorkflowModel) Insert(workflow *Workflow) error {
	query := `INSERT INTO workflows (uniqueid, name, states, startstate, endstate, callbackwebhook, retry, retryafter, retryurl, circuitbreaker, circuitbreakerstatus, circuitbreakerfailurecount, circuitbreakeropentimeout, circuitbreakerhalfopencount, active)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
				RETURNING id, created_at, uniqueid, version`

	args := []any{workflow.UniqueID, workflow.Name, pq.Array(workflow.States), workflow.StartState, workflow.EndState, workflow.CallbackWebhook, workflow.Retry, workflow.RetryAfter, workflow.RetryURL, workflow.CircuitBreaker, workflow.CircuitBreakerStatus, workflow.CircuitBreakerFailureCount, workflow.CircuitBreakerOpenTimeout, workflow.CircuitBreakerHalfOpenCount, workflow.Active}

	return w.DB.QueryRow(query, args...).Scan(&workflow.ID, &workflow.CreatedAt, &workflow.UniqueID, &workflow.Version)
}

func (w WorkflowModel) Get(id int64) (*Workflow, error) {
	return nil, nil
}

func (w WorkflowModel) Update(workflow *Workflow) error {
	return nil
}

func (w WorkflowModel) Delete(id int64) error {
	return nil
}
