package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/config"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/database"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/handler"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/blob"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/email"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/lib/job"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/repository"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/router"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Server struct {
	Config     *config.Config
	Logger     *zerolog.Logger
	Echo       *echo.Echo
	DB         *database.Database
	Redis      *redis.Client
	Jobs       *job.Service
	Blob       *blob.Client
	Queue      *azqueue.QueueClient
	HTTPServer *http.Server
}

func New(ctx context.Context, cfg *config.Config, logger *zerolog.Logger) (*Server, error) {
	db, err := database.New(ctx, cfg.Database, logger)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Address})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Warn().Err(err).Msg("redis unavailable at startup")
	}

	emailClient := email.New(cfg.Integration.ResendAPIKey)
	jobService := job.New(cfg, emailClient, logger)

	var blobClient *blob.Client
	if cfg.Azure.StorageConnectionString != "" {
		blobClient, err = blob.New(cfg.Azure.StorageConnectionString, cfg.Azure.StorageContainerName, cfg.Azure.StorageAccountName)
		if err != nil {
			logger.Warn().Err(err).Msg("blob client not configured")
		}
	}

	var queueClient *azqueue.QueueClient
	if cfg.Azure.StorageConnectionString != "" && cfg.Azure.StorageQueueName != "" {
		queueClient, err = azqueue.NewQueueClientFromConnectionString(cfg.Azure.StorageConnectionString, cfg.Azure.StorageQueueName, nil)
		if err != nil {
			logger.Warn().Err(err).Msg("queue client not configured")
		}
	}

	repos := repository.New(db.Pool)
	services := service.New(cfg, repos, db.Pool, jobService, emailClient, blobClient, queueClient, logger)
	handlers := handler.New(services)
	echoServer := router.New(cfg, handlers, services)

	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      echoServer,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	return &Server{
		Config:     cfg,
		Logger:     logger,
		Echo:       echoServer,
		DB:         db,
		Redis:      redisClient,
		Jobs:       jobService,
		Blob:       blobClient,
		Queue:      queueClient,
		HTTPServer: httpServer,
	}, nil
}

func (s *Server) Start() error {
	s.Logger.Info().Str("port", s.Config.Server.Port).Msg("starting server")
	if err := s.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.HTTPServer != nil {
		if err := s.HTTPServer.Shutdown(ctx); err != nil {
			return err
		}
	}
	if s.Jobs != nil {
		s.Jobs.Close()
	}
	if s.Redis != nil {
		if err := s.Redis.Close(); err != nil {
			s.Logger.Warn().Err(err).Msg("close redis client")
		}
	}
	if s.DB != nil {
		s.DB.Close()
	}
	return nil
}
