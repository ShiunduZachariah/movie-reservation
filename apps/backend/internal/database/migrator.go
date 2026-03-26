package database

import (
	"context"
	"fmt"
	"os"

	"github.com/ShiunduZachariah/movie-reservation/apps/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/rs/zerolog"
)

func Migrate(ctx context.Context, cfg config.DatabaseConfig, logger *zerolog.Logger, direction string) error {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("open migration pool: %w", err)
	}
	defer pool.Close()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration conn: %w", err)
	}
	defer conn.Release()

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), "schema_version")
	if err != nil {
		return fmt.Errorf("new migrator: %w", err)
	}

	migrationsFS := os.DirFS("internal/database/migrations")
	if err := migrator.LoadMigrations(migrationsFS); err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	switch direction {
	case "up":
		err = migrator.Migrate(ctx)
	case "down":
		err = migrator.MigrateTo(ctx, 0)
	default:
		return fmt.Errorf("unsupported migration direction: %s", direction)
	}
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	logger.Info().Str("direction", direction).Msg("migrations completed")
	return nil
}
