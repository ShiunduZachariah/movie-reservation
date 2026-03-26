package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv/autoload"
	envprovider "github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

const envPrefix = "CINERESERVE_"

type Config struct {
	Primary     PrimaryConfig     `koanf:"primary" validate:"required"`
	Server      ServerConfig      `koanf:"server" validate:"required"`
	Database    DatabaseConfig    `koanf:"database" validate:"required"`
	Auth        AuthConfig        `koanf:"auth" validate:"required"`
	Redis       RedisConfig       `koanf:"redis" validate:"required"`
	Integration IntegrationConfig `koanf:"integration" validate:"required"`
	App         AppConfig         `koanf:"app" validate:"required"`
	Azure       AzureConfig       `koanf:"azure" validate:"required"`
}

type PrimaryConfig struct {
	Env string `koanf:"env" validate:"required"`
}

type ServerConfig struct {
	Port               string   `koanf:"port" validate:"required"`
	ReadTimeout        int      `koanf:"read_timeout" validate:"required"`
	WriteTimeout       int      `koanf:"write_timeout" validate:"required"`
	IdleTimeout        int      `koanf:"idle_timeout" validate:"required"`
	CORSAllowedOrigins []string `koanf:"cors_allowed_origins" validate:"required,min=1"`
}

type DatabaseConfig struct {
	Host            string `koanf:"host" validate:"required"`
	Port            int    `koanf:"port" validate:"required"`
	User            string `koanf:"user" validate:"required"`
	Password        string `koanf:"password"`
	Name            string `koanf:"name" validate:"required"`
	SSLMode         string `koanf:"ssl_mode" validate:"required"`
	MaxOpenConns    int32  `koanf:"max_open_conns" validate:"required"`
	MaxIdleConns    int32  `koanf:"max_idle_conns" validate:"required"`
	ConnMaxLifetime int    `koanf:"conn_max_lifetime" validate:"required"`
	ConnMaxIdleTime int    `koanf:"conn_max_idle_time" validate:"required"`
}

type AuthConfig struct {
	SecretKey string `koanf:"secret_key" validate:"required"`
}

type RedisConfig struct {
	Address string `koanf:"address" validate:"required"`
}

type IntegrationConfig struct {
	ResendAPIKey string `koanf:"resend_api_key" validate:"required"`
	ResendFrom   string `koanf:"resend_from"`
}

type AppConfig struct {
	BaseURL string `koanf:"base_url" validate:"required,url"`
	Name    string `koanf:"name" validate:"required"`
}

type AzureConfig struct {
	StorageAccountName      string `koanf:"storage_account_name" validate:"required"`
	StorageContainerName    string `koanf:"storage_container_name" validate:"required"`
	StorageQueueName        string `koanf:"storage_queue_name" validate:"required"`
	StorageConnectionString string `koanf:"storage_connection_string" validate:"required"`
}

func Load() (*Config, error) {
	k := koanf.New(".")
	if err := k.Load(envprovider.Provider(envPrefix, ".", envKeyMapper), nil); err != nil {
		return nil, fmt.Errorf("load env config: %w", err)
	}

	cfg := &Config{}
	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if len(cfg.Server.CORSAllowedOrigins) == 1 && strings.Contains(cfg.Server.CORSAllowedOrigins[0], ",") {
		cfg.Server.CORSAllowedOrigins = splitCSV(cfg.Server.CORSAllowedOrigins[0])
	}

	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func envKeyMapper(key string) string {
	return strings.ToLower(strings.TrimPrefix(key, envPrefix))
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	return cfg
}
