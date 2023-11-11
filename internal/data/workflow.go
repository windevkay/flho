package data

import "time"

type Workflow struct {
	ID              int64     `json:"-"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	UniqueID        string    `json:"uniqueId"`
	Name            string    `json:"name"`
	States          []string  `json:"states"`
	StartState      string    `json:"startState"`
	EndState        string    `json:"endState"`
	IsTimed         bool      `json:"isTimed"`
	Timeout         time.Time `json:"timeout,omitempty"`
	CallbackWebhook string    `json:"webhook,omitempty"`
	Alert           bool      `json:"alert"`
	AlertEmail      string    `json:"alertEmail,omitempty"`
	AlertWebhook    string    `json:"alertWebhook,omitempty"`
	Active          bool      `json:"active"`
	Version         int32     `json:"version"`
}
