package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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

	migrationsPath, err := resolveMigrationsPath()
	if err != nil {
		return err
	}

	migrationsFS := os.DirFS(migrationsPath)
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

func resolveMigrationsPath() (string, error) {
	candidates := []string{
		filepath.Join("apps", "backend", "internal", "database", "migrations"),
		filepath.Join("internal", "database", "migrations"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("resolve migrations path: could not find migrations directory")
}
