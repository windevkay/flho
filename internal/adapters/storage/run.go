package storage

import (
	"github.com/google/uuid"
	"github.com/windevkay/flho/internal/domain"
)

type Store interface {
	Get(key string) any
	Set(key string, value any)
}

type RunRepo struct {
	store Store
}

func NewRunRepo(store Store) *RunRepo {
	return &RunRepo{
		store: store,
	}
}

func (r *RunRepo) SaveRun(run *domain.Run) error {
	r.store.Set(run.UniqueID.String(), run)
	return nil
}

func (r *RunRepo) UpdateRun(runID uuid.UUID, step int) error {
	run := r.store.Get(runID.String()).(domain.Run)
	run.Step = step
	r.store.Set(runID.String(), run)

	return nil
}
