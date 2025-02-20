package services

import (
	"log/slog"
	"sync"

	"github.com/rabbitmq/amqp091-go"
	"github.com/windevkay/flho/identity_service/internal/data"
	"github.com/windevkay/flho/identity_service/internal/rpc"
)

type SendMessageToQueueFunc func(ch *amqp091.Channel, data interface{}, entity string, action string) error
type RunInBackgroundFunc func(f func(), wg *sync.WaitGroup)

type ServiceConfig struct {
	Background RunInBackgroundFunc
	Channel    *amqp091.Channel
	Logger     *slog.Logger
	Message    SendMessageToQueueFunc
	Models     data.Models
	Rpclients  rpc.Clients
	Wg         *sync.WaitGroup
}

func (s *ServiceConfig) Register(models data.Models, rpc rpc.Clients, ch *amqp091.Channel, wg *sync.WaitGroup, logger *slog.Logger, bg RunInBackgroundFunc, message SendMessageToQueueFunc) {
	s.Models = models
	s.Rpclients = rpc
	s.Channel = ch
	s.Wg = wg
	s.Logger = logger
	s.Background = bg
	s.Message = message
}
