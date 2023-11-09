package data

import "time"

type Workflow struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	Version   int32     `json:"version"`
}
