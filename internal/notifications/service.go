package notifications

import (
	"context"
	"log/slog"
)

type Service struct {
	logger *slog.Logger
}

func NewService(logger *slog.Logger) Service {
	return Service{logger: logger}
}

func (s Service) Publish(_ context.Context, event string, payload map[string]any) {
	s.logger.Info("notification_event", slog.String("event", event), slog.Any("payload", payload))
}

