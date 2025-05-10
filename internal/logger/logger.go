package logger

import (
	"log/slog"
	"os"
)

// New creates a new Zap logger with the specified log level and output file.
// If outputFile is empty or "stdout", logs will be written to standard output.
// If outputFile is "stderr", logs will be written to standard error.
// Otherwise, logs will be written to the specified file.
func New(config Config) (*slog.Logger, error) {
	var programLevel = new(slog.LevelVar) // Info by default
	if err := programLevel.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, err
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: programLevel,
	}))

	if config.Format == "text" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: programLevel,
		}))
	}

	return logger, nil
}
