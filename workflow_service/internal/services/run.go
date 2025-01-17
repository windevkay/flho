package services

import (
	"context"
	"errors"
	"time"

	"github.com/windevkay/flho/workflow_service/internal/data"
	"github.com/windevkay/flhoutils/helpers"
)

type RunService struct {
	*ServiceConfig
	countdowns map[string]context.CancelFunc
}

func (rs *RunService) StartRun(workflowUniqueId string, step int) (string, error) {
	workflow, err := rs.Models.Workflows.Get(workflowUniqueId)
	if err != nil {
		rs.Logger.Error(err.Error())
		return "", err
	}

	state, validStep := workflow.GetStateByStep(step)
	if !validStep {
		return "", errors.New("invalid step provided")
	}

	run := &data.Run{
		WorkflowId: workflow.ID,
		UniqueID:   helpers.GenerateUniqueId(12),
		Step:       step,
	}

	err = rs.Models.Runs.Insert(run)
	if err != nil {
		return "", err
	}

	rs.analyzeStateRetry(state, run.UniqueID)

	return run.UniqueID, nil
}

func (rs *RunService) UpdateRun(workflowUniqueId string, runUniqueId string, step int) error {
	workflow, err := rs.Models.Workflows.Get(workflowUniqueId)
	if err != nil {
		rs.Logger.Error(err.Error())
		return err
	}

	state, validStep := workflow.GetStateByStep(step)
	if !validStep {
		return errors.New("invalid step provided")
	}

	run, err := rs.Models.Runs.Get(runUniqueId)
	if err != nil {
		return err
	}

	run.Step = step

	err = rs.Models.Runs.Update(run)
	if err != nil {
		return err
	}

	if cancelExisting, exists := rs.countdowns[runUniqueId]; exists {
		cancelExisting()
	}

	rs.analyzeStateRetry(state, runUniqueId)

	return nil
}

func (rs *RunService) analyzeStateRetry(state data.State, identifier string) {
	if state.Retry {
		helpers.RunInBackground(func() {
			ctx, cancel := context.WithCancel(context.Background())
			rs.initiateCountdown(ctx, time.Duration(state.RetryAfter))

			rs.countdowns[identifier] = cancel
		}, rs.Wg)
	}
}

func (rs *RunService) initiateCountdown(ctx context.Context, duration time.Duration) {
	select {
	case <-time.After(duration):
		rs.Logger.Info("Countdown completed after", "duration", duration.String())
	case <-ctx.Done():
		rs.Logger.Info("Countdown cancelled")
		return
	}
}
