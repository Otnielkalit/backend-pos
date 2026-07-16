package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init configures the global zerolog logger.
// In development, output is human-readable (ConsoleWriter).
// In production, output is structured JSON — consumed by Loki via Promtail.
//
// Standard fields written on every log entry:
//
//	service, env — for filtering in Grafana dashboards
func Init(appName, env string) {
	zerolog.TimeFieldFormat = time.RFC3339

	if env == "development" {
		// Pretty-print for local development
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}).With().
			Str("service", appName).
			Logger()
		return
	}

	// Production: structured JSON output
	log.Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", appName).
		Str("env", env).
		Logger()
}

// Get returns the global zerolog logger.
// Use this to get a logger instance in infrastructure code (e.g., middleware).
// In feature code, prefer passing a logger via context using WithContext/FromContext.
func Get() zerolog.Logger {
	return log.Logger
}

// WithRequestID returns a child logger with request_id field attached.
// Called by the HTTP logger middleware per request.
func WithRequestID(requestID string) zerolog.Logger {
	return log.Logger.With().Str("request_id", requestID).Logger()
}
