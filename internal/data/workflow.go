package data

import (
	"slices"
	"time"

	"github.com/windevkay/flho/internal/validator"
)

type Workflow struct {
	ID              int64     `json:"-"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	UniqueID        string    `json:"uniqueId"`
	Name            string    `json:"name"`
	States          []string  `json:"states"`
	StartState      string    `json:"startState"`
	EndState        string    `json:"endState"`
	CallbackWebhook string    `json:"webhook,omitempty"`
	IsTimed         bool      `json:"isTimed"`
	Timeout         Timeout   `json:"timeout,omitempty"`
	AlertWebhook    string    `json:"alertWebhook,omitempty"`
	Active          bool      `json:"active"`
	Version         int32     `json:"version"`
}

func ValidateWorkflow(v *validator.Validator, w *Workflow) {
	v.Check(w.Name != "", "name", "must be provided")

	v.Check(len(w.States) >= 2, "states", "must have atleast 2 values")
	v.Check(validator.Unique(w.States), "states", "must not contain duplicate values")

	v.Check(w.StartState != "", "startState", "must be provided")
	v.Check(slices.Contains(w.States, w.StartState), "startState", "must be part of the states list")

	v.Check(w.EndState != "", "endState", "must be provided")
	v.Check(slices.Contains(w.States, w.EndState), "endState", "must be part of the states list")

	if w.IsTimed {
		v.Check(w.AlertWebhook != "", "alertWebhook", "must be provided")
		v.Check(w.Timeout != 0, "timeout", "must be provided")
		v.Check(w.Timeout > 0, "timeout", "must be a non-negative value")
	}
}
