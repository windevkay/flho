package services

import (
	"errors"

	"github.com/windevkay/flho/workflow_service/internal/data"
	"github.com/windevkay/flhoutils/helpers"
)

type RunService struct {
	*ServiceConfig
}

func (rs *RunService) StartRun(workflowUniqueId string, step int) (string, error) {
	workflow, err := rs.Models.Workflows.Get(workflowUniqueId)
	if err != nil {
		rs.Logger.Error(err.Error())
		return "", err
	}

	if validStep := workflow.HasStateStep(step); !validStep {
		return "", errors.New("invalid step provided")
	}

	run := &data.Run{
		UniqueID: helpers.GenerateUniqueId(12),
		Step:     step,
	}

	err = rs.Models.Runs.Insert(run)
	if err != nil {
		return "", err
	}

	return run.UniqueID, nil
}

func (rs *RunService) UpdateRun(runUniqueId string, step int) error {
	run, err := rs.Models.Runs.Get(runUniqueId)
	if err != nil {
		return err
	}

	run.Step = step

	return rs.Models.Runs.Update(run)
}
