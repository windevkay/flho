package data

type MockWorkflowModel struct{}

func (w MockWorkflowModel) InsertWithTx(workflow *Workflow) error {
	return nil
}

func (w MockWorkflowModel) GetIdentityId(uuid string) (int64, error) {
	return 1, nil
}

func (w MockWorkflowModel) Get(id int64) (*Workflow, error) {
	workflow := &Workflow{
		Name: "mock_workflow",
	}

	return workflow, nil
}

func (w MockWorkflowModel) GetAll(organizationId int64, filters Filters) ([]*Workflow, Metadata, error) {
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
