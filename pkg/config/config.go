package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Tracing  TracingConfig  `mapstructure:"tracing"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	GRPCPort int `mapstructure:"grpc_port"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// TracingConfig holds tracing configuration
type TracingConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	ServiceName string `mapstructure:"service_name"`
	Endpoint    string `mapstructure:"endpoint"`
}

// Load loads configuration from file and environment
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.grpc_port", 9090)
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "postgres")
	v.SetDefault("database.dbname", "slips")
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("tracing.enabled", true)
	v.SetDefault("tracing.service_name", "slips-core")
	v.SetDefault("tracing.endpoint", "localhost:4317")

	// Read from config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Override with environment variables
	v.SetEnvPrefix("SLIPS")
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// DatabaseURL returns the database connection string
// WARNING: This contains the password in plaintext. Never log or expose this value.
// Use SafeDatabaseURL() for logging purposes.
func (c *DatabaseConfig) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}

// SafeDatabaseURL returns a sanitized database connection string for logging
func (c *DatabaseConfig) SafeDatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:***@%s:%d/%s?sslmode=%s",
		c.User, c.Host, c.Port, c.DBName, c.SSLMode,
	)
}
