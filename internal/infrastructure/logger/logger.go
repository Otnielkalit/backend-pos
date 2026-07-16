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

	// Ensure logs directory exists
	_ = os.MkdirAll("logs", 0755)

	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil && env != "test" {
		log.Warn().Err(err).Msg("failed to open log file, falling back to stdout only")
	}

	if env == "development" {
		// Pretty-print for local development on stdout
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
		}

		var multi zerolog.LevelWriter
		if logFile != nil {
			// Write JSON to file for Promtail, and pretty to console
			multi = zerolog.MultiLevelWriter(consoleWriter, logFile)
		} else {
			multi = zerolog.MultiLevelWriter(consoleWriter)
		}

		log.Logger = zerolog.New(multi).With().
			Timestamp().
			Str("service", appName).
			Str("env", env).
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
