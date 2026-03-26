package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/config"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/database"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/logger"
	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/server"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.Primary.Env)
	ctx := context.Background()

	if len(os.Args) > 2 && os.Args[1] == "migrate" {
		if err := database.Migrate(ctx, cfg.Database, log, os.Args[2]); err != nil {
			log.Fatal().Err(err).Msg("migration failed")
		}
		return
	}

	srv, err := server.New(ctx, cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create server")
	}

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal().Err(err).Msg("server stopped")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("shutdown failed")
	}
}
