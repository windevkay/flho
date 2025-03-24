package services

import (
	"log/slog"
	"sync"

	"github.com/windevkay/flho/internal/data"
)

type ValidationErr struct {
	Err    error
	Fields map[string]string
}

func (c *ValidationErr) Error() string { return "validation error" }

type RunInBackgroundFunc func(f func(), wg *sync.WaitGroup)

type ServiceConfig struct {
	Background RunInBackgroundFunc
	Logger     *slog.Logger
	Models     data.Models
	Wg         *sync.WaitGroup
}

func (s *ServiceConfig) Register(models data.Models, wg *sync.WaitGroup, logger *slog.Logger, bg RunInBackgroundFunc) {
	s.Models = models
	s.Wg = wg
	s.Logger = logger
	s.Background = bg
}
