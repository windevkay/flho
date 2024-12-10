package queue

type externalServiceData struct {
	name   string
	events []struct{ entityName string }
}

// list of other exchanges this service is interested in setting up
var (
	externalExchanges = []externalServiceData{
		{
			name: "identity_service_exchange",
			events: []struct{ entityName string }{
				{
					entityName: "identity",
				},
			},
		},
	}
)
