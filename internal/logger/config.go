package logger

// Config represents the logger configuration
type Config struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}
