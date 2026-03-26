package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("CINERESERVE_PRIMARY.ENV", "test")
	t.Setenv("CINERESERVE_SERVER.PORT", "8080")
	t.Setenv("CINERESERVE_SERVER.READ_TIMEOUT", "15")
	t.Setenv("CINERESERVE_SERVER.WRITE_TIMEOUT", "15")
	t.Setenv("CINERESERVE_SERVER.IDLE_TIMEOUT", "30")
	t.Setenv("CINERESERVE_SERVER.CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	t.Setenv("CINERESERVE_DATABASE.HOST", "localhost")
	t.Setenv("CINERESERVE_DATABASE.PORT", "5432")
	t.Setenv("CINERESERVE_DATABASE.USER", "postgres")
	t.Setenv("CINERESERVE_DATABASE.PASSWORD", "postgres")
	t.Setenv("CINERESERVE_DATABASE.NAME", "movie_reservation")
	t.Setenv("CINERESERVE_DATABASE.SSL_MODE", "disable")
	t.Setenv("CINERESERVE_DATABASE.MAX_OPEN_CONNS", "10")
	t.Setenv("CINERESERVE_DATABASE.MAX_IDLE_CONNS", "5")
	t.Setenv("CINERESERVE_DATABASE.CONN_MAX_LIFETIME", "300")
	t.Setenv("CINERESERVE_DATABASE.CONN_MAX_IDLE_TIME", "60")
	t.Setenv("CINERESERVE_REDIS.ADDRESS", "localhost:6379")
	t.Setenv("CINERESERVE_AUTH.SECRET_KEY", "secret")
	t.Setenv("CINERESERVE_INTEGRATION.RESEND_API_KEY", "re_test")
	t.Setenv("CINERESERVE_APP.BASE_URL", "http://localhost:3000")
	t.Setenv("CINERESERVE_APP.NAME", "CineReserve")
	t.Setenv("CINERESERVE_AZURE.STORAGE_ACCOUNT_NAME", "storage")
	t.Setenv("CINERESERVE_AZURE.STORAGE_CONTAINER_NAME", "movie-posters")
	t.Setenv("CINERESERVE_AZURE.STORAGE_QUEUE_NAME", "ticket-emails")
	t.Setenv("CINERESERVE_AZURE.STORAGE_CONNECTION_STRING", "UseDevelopmentStorage=true")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "test", cfg.Primary.Env)
	require.Equal(t, []string{"http://localhost:3000", "http://localhost:5173"}, cfg.Server.CORSAllowedOrigins)
	require.Equal(t, "CineReserve", cfg.App.Name)
}

func TestSplitCSV(t *testing.T) {
	require.Equal(t, []string{"a", "b", "c"}, splitCSV("a, b, c"))
	require.Empty(t, splitCSV(""))
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
