package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Tracing  TracingConfig  `mapstructure:"tracing"`
	Auth     AuthConfig     `mapstructure:"auth"`
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

// AuthConfig holds authentication configuration
type AuthConfig struct {
	IdentraGRPCEndpoint string      `mapstructure:"identra_grpc_endpoint"`
	ExpectedIssuer      string      `mapstructure:"expected_issuer"`
	OAuth               OAuthConfig `mapstructure:"oauth"`
}

// OAuthConfig holds OAuth-specific configuration
type OAuthConfig struct {
	Provider    string `mapstructure:"provider"`
	RedirectURL string `mapstructure:"redirect_url"`
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
	v.SetDefault("auth.identra_grpc_endpoint", "localhost:8080")
	v.SetDefault("auth.expected_issuer", "identra")

	// Read from config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Override with environment variables
	// SLIPS_DATABASE_PASSWORD maps to database.password
	v.SetEnvPrefix("SLIPS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Explicitly bind nested config keys to environment variables
	// This is required for viper to properly handle nested structures
	_ = v.BindEnv("database.password")
	_ = v.BindEnv("database.host")
	_ = v.BindEnv("database.port")
	_ = v.BindEnv("database.user")
	_ = v.BindEnv("database.dbname")
	_ = v.BindEnv("database.sslmode")
	_ = v.BindEnv("auth.identra_grpc_endpoint")
	_ = v.BindEnv("auth.expected_issuer")
	_ = v.BindEnv("auth.oauth.provider")
	_ = v.BindEnv("auth.oauth.redirect_url")
	_ = v.BindEnv("server.grpc_port")
	_ = v.BindEnv("tracing.enabled")
	_ = v.BindEnv("tracing.service_name")
	_ = v.BindEnv("tracing.endpoint")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Log configuration (excluding sensitive data)
	log.Printf("[CONFIG] GRPC Port: %d", cfg.Server.GRPCPort)
	log.Printf("[CONFIG] Database Host: %s:%d", cfg.Database.Host, cfg.Database.Port)
	log.Printf("[CONFIG] Database Name: %s", cfg.Database.DBName)
	log.Printf("[CONFIG] Tracing Enabled: %t", cfg.Tracing.Enabled)
	log.Printf("[CONFIG] Auth Identra Endpoint: %s", cfg.Auth.IdentraGRPCEndpoint)
	log.Printf("[CONFIG] Auth Expected Issuer: %s", cfg.Auth.ExpectedIssuer)
	log.Printf("[CONFIG] OAuth Provider: %s", cfg.Auth.OAuth.Provider)
	log.Printf("[CONFIG] OAuth Redirect URL: %s", cfg.Auth.OAuth.RedirectURL)

	// Also log environment variable status for OAuth redirect URL
	if envVal := os.Getenv("SLIPS_AUTH_OAUTH_REDIRECT_URL"); envVal != "" {
		log.Printf("[CONFIG] Environment variable SLIPS_AUTH_OAUTH_REDIRECT_URL is set to: %s", envVal)
	} else {
		log.Printf("[CONFIG] Environment variable SLIPS_AUTH_OAUTH_REDIRECT_URL is not set")
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
