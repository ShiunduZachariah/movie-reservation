package job

import (
	"context"
	"fmt"
	"net"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/config"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/email"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

type Service struct {
	client *asynq.Client
	server *asynq.Server
	logger *zerolog.Logger
}

func New(cfg *config.Config, emailClient *email.Client, logger *zerolog.Logger) *Service {
	redisConn := asynq.RedisClientOpt{Addr: cfg.Redis.Address}
	svc := &Service{
		client: asynq.NewClient(redisConn),
		server: asynq.NewServer(redisConn, asynq.Config{Concurrency: 5}),
		logger: logger,
	}

	mux := asynq.NewServeMux()
	handlers := NewHandlers(emailClient, logger)
	mux.HandleFunc(TypeTicketConfirmation, handlers.HandleTicketConfirmation)

	go func() {
		if err := svc.server.Run(mux); err != nil {
			if _, ok := err.(net.Error); ok {
				logger.Warn().Err(err).Msg("job server not started")
				return
			}
			logger.Error().Err(err).Msg("job server stopped")
		}
	}()

	return svc
}

func (s *Service) Enqueue(ctx context.Context, task *asynq.Task) error {
	if _, err := s.client.EnqueueContext(ctx, task); err != nil {
		return fmt.Errorf("enqueue job: %w", err)
	}
	return nil
}

func (s *Service) Close() {
	if s == nil {
		return
	}
	if s.client != nil {
		if err := s.client.Close(); err != nil {
			s.logger.Warn().Err(err).Msg("close asynq client")
		}
	}
	if s.server != nil {
		s.server.Shutdown()
	}
}
