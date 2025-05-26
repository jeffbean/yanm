package logger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         Config
		expectError string
	}{
		{
			name: "valid level info, json format",
			cfg: Config{
				Level:  "info",
				Format: "json",
			},
		},
		{
			name: "valid level debug, text format",
			cfg: Config{
				Level:  "debug",
				Format: "text",
			},
		},
		{
			name: "valid level warn, json format",
			cfg: Config{
				Level:  "warn",
				Format: "json",
			},
		},
		{
			name: "valid level error, text format",
			cfg: Config{
				Level:  "error",
				Format: "text",
			},
		},
		{
			name: "invalid level string",
			cfg: Config{
				Level:  "invalid_level_string",
				Format: "json",
			},
			expectError: "slog: level string \"invalid_level_string\": unknown name",
		},
		{
			name: "empty level string (expect error)",
			cfg: Config{
				Level:  "", // slog.LevelVar.UnmarshalText returns error for empty string
				Format: "json",
			},
			expectError: "slog: level string \"\": unknown name",
		},
		{
			name: "empty format (defaults to json)",
			cfg: Config{
				Level:  "info",
				Format: "", // Should default to JSON
			},
		},
		{
			name: "output file stderr (current code uses stdout)",
			cfg: Config{
				Level:  "info",
				Format: "json",
			},
		},
		{
			name: "output file to a specific file (current code uses stdout)",
			cfg: Config{
				Level:  "info",
				Format: "text",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger, err := New(tc.cfg)

			if tc.expectError != "" {
				require.EqualError(t, err, tc.expectError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, logger)
			logger.Info("Test log entry", "testCase", tc.name)
		})
		// Note: Verifying the actual handler type (JSON/Text) or the exact level
		// from the slog.Logger instance is non-trivial without inspecting unexported fields
		// or capturing output. These tests primarily ensure the New function behaves as expected
		// regarding error returns and successful logger instantiation.
	}
}
