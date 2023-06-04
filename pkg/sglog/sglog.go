package sglog

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

type connectionIDType int

const (
	ConnectionIDKey connectionIDType = iota
)

var logger zerolog.Logger

func init() {
	logLevel, err := zerolog.ParseLevel(os.Getenv("LOG"))

	if err != nil {
		logLevel = zerolog.ErrorLevel
	}

	logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		Level(logLevel)
}

// WithRqId returns a context which knows its request ID
func WithRqId(ctx context.Context, rqId string) context.Context {
	return context.WithValue(ctx, ConnectionIDKey, rqId)
}

func Logger(ctx context.Context) zerolog.Logger {
	newLogger := logger
	if ctx != nil {
		if ctxRqId, ok := ctx.Value(ConnectionIDKey).(string); ok {
			return logger.With().Timestamp().Str("S.Name", ctxRqId).Logger()
		}
	}
	return newLogger.With().Timestamp().Logger()
}
