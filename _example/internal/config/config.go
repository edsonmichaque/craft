package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration sections
type Config struct {
	Server   ServerConfig   ` + "`mapstructure:\"server\" yaml:\"server\"`" + `
	Database DatabaseConfig ` + "`mapstructure:\"database\" yaml:\"database\"`" + `
	Logger   LoggerConfig  ` + "`mapstructure:\"logger\" yaml:\"logger\"`" + `
}

// Load reads configuration from file and environment variables
func Load(opts ...Option) (*Config, error) {
	// Default options
	options := &options{
		configFormat:   "yaml",
		validateConfig: true,
		configDirs:    []string{"/etc/craft", "$HOME/.config/craft"},
		envPrefix:     "CRAFT",
		logger:        defaultLogger{},
	}

	// Apply provided options
	for _, opt := range opts {
		opt(options)
	}

	v := viper.New()

	// Set config name and type if file is provided
	if options.configFile != "" {
		v.SetConfigFile(options.configFile)
	} else {
		v.SetConfigName("config")
		v.SetConfigType(options.configFormat)
	}

	// Add config paths
	for _, dir := range options.configDirs {
		v.AddConfigPath(dir)
	}

	// Set environment variable prefix
	if options.envPrefix != "" {
		v.SetEnvPrefix(options.envPrefix)
		v.AutomaticEnv()
		v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	}

	// Set defaults if provided
	if options.defaultConfig != nil {
		if err := v.MergeConfigMap(structToMap(options.defaultConfig)); err != nil {
			return nil, fmt.Errorf("failed to set defaults: %w", err)
		}
	}

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		options.logger.Debug("No config file found, using defaults and environment variables")
	}

	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate if enabled
	if options.validateConfig {
		if err := config.Validate(); err != nil {
			return nil, fmt.Errorf("config validation failed: %w", err)
		}
	}

	return config, nil
}

// Helper function to convert struct to map
func structToMap(obj interface{}) map[string]interface{} {
	data, _ := json.Marshal(obj)
	result := make(map[string]interface{})
	json.Unmarshal(data, &result)
	return result
}

// Default logger implementation
type defaultLogger struct{}

func (l defaultLogger) Debug(args ...interface{}) {}
func (l defaultLogger) Info(args ...interface{})  {}
func (l defaultLogger) Error(args ...interface{}) {}

// Add validation method to Config
func (c *Config) Validate() error {
	if c.Server.Port < 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	// Add more validation as needed
	return nil
}`,

		"server.go": `package config

import "time"

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Host           string        ` + "`mapstructure:\"host\" yaml:\"host\"`" + `
	Port           int           ` + "`mapstructure:\"port\" yaml:\"port\"`" + `
	ReadTimeout    time.Duration ` + "`mapstructure:\"read_timeout\" yaml:\"read_timeout\"`" + `
	WriteTimeout   time.Duration ` + "`mapstructure:\"write_timeout\" yaml:\"write_timeout\"`" + `
	MaxHeaderBytes int           ` + "`mapstructure:\"max_header_bytes\" yaml:\"max_header_bytes\"`" + `
	AllowedOrigins []string      ` + "`mapstructure:\"allowed_origins\" yaml:\"allowed_origins\"`" + `
}

// GetAddress returns the full address string for the server
func (c ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}`,

		"database.go": `package config

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host     string ` + "`mapstructure:\"host\" yaml:\"host\"`" + `
	Port     int    ` + "`mapstructure:\"port\" yaml:\"port\"`" + `
	Name     string ` + "`mapstructure:\"name\" yaml:\"name\"`" + `
	User     string ` + "`mapstructure:\"user\" yaml:\"user\"`" + `
	Password string ` + "`mapstructure:\"password\" yaml:\"password\"`" + `
	SSLMode  string ` + "`mapstructure:\"ssl_mode\" yaml:\"ssl_mode\"`" + `
}

// GetDSN returns the database connection string
func (c DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Name, c.User, c.Password, c.SSLMode)
}