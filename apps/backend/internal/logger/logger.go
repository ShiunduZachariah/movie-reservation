package logger

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

func New(env string) *zerolog.Logger {
	level := zerolog.InfoLevel
	if strings.EqualFold(env, "development") {
		level = zerolog.DebugLevel
	}

	zerolog.SetGlobalLevel(level)
	output := os.Stdout
	logger := zerolog.New(output).With().Timestamp().Logger()

	if strings.EqualFold(env, "development") {
		console := zerolog.ConsoleWriter{Out: output}
		logger = zerolog.New(console).With().Timestamp().Logger()
	}

	return &logger
}
