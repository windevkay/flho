package data

type MockWorkflowModel struct{}

func (w MockWorkflowModel) Insert(workflow *Workflow) error {
	return nil
}

func (w MockWorkflowModel) Get(id int64) (*Workflow, error) {
	workflow := &Workflow{
		Name: "mock_workflow",
	}

	return workflow, nil
}

func (w MockWorkflowModel) Update(workflow *Workflow) error {
	return nil
}

func (w MockWorkflowModel) Delete(id int64) error {
	return nil
}
