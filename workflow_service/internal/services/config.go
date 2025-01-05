package services

import (
	"log/slog"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"github.com/windevkay/flho/workflow_service/internal/data"
	"github.com/windevkay/flho/workflow_service/internal/rpc"
)

type ServiceConfig struct {
	Models    data.Models
	Rpclients rpc.Clients
	Channel   *amqp091.Channel
	Wg        *sync.WaitGroup
	Logger    *slog.Logger
}

func (s *ServiceConfig) Register(m data.Models, r rpc.Clients, c *amqp091.Channel, w *sync.WaitGroup, l *slog.Logger) {
	s.Models = m
	s.Rpclients = r
	s.Channel = c
	s.Wg = w
	s.Logger = l
}
