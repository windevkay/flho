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

func (w MockWorkflowModel) GetAll(name string, states []string, filters Filters) ([]*Workflow, Metadata, error) {
	var workflows []*Workflow

	workflow := &Workflow{
		Name: "mock_workflow",
	}
	workflows = append(workflows, workflow)

	return workflows, Metadata{}, nil
}

func (w MockWorkflowModel) Update(workflow *Workflow) error {
	return nil
}

func (w MockWorkflowModel) Delete(id int64) error {
	return nil
}
