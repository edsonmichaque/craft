package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"text/template"
)

func generateConfigPackage(projectPath string, cfg Config) error {
	configDir := filepath.Join(projectPath, "internal/config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Map of config files to their templates
	configFiles := map[string]string{
		"config.go": `package config

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
		configDirs:    []string{"/etc/{{.ProjectName}}", "$HOME/.config/{{.ProjectName}}"},
		envPrefix:     "{{.EnvPrefix}}",
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
}`,

		"logger.go": `package config

// LoggerConfig holds all logging-related configuration
type LoggerConfig struct {
	Level  string            ` + "`mapstructure:\"level\" yaml:\"level\"`" + `
	Format string            ` + "`mapstructure:\"format\" yaml:\"format\"`" + `
	Output string            ` + "`mapstructure:\"output\" yaml:\"output\"`" + `
	Fields map[string]string ` + "`mapstructure:\"fields\" yaml:\"fields\"`" + `
}`,

		"config_test.go": `package config

import (
	"testing"
	"os"
	"path/filepath"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yml")
	
	configContent := []byte(` + "`" + `
server:
  host: "127.0.0.1"
  port: 8080
database:
  host: "localhost"
  port: 5432
logger:
  level: "debug"
` + "`" + `)
	
	if err := os.WriteFile(configFile, configContent, 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}
	
	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Test server config
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Expected server host 127.0.0.1, got %s", cfg.Server.Host)
	}
	
	// Test database config
	if cfg.Database.Host != "localhost" {
		t.Errorf("Expected database host localhost, got %s", cfg.Database.Host)
	}
	
	// Test logger config
	if cfg.Logger.Level != "debug" {
		t.Errorf("Expected logger level debug, got %s", cfg.Logger.Level)
	}
}`,
	}

	// Generate each config file
	for filename, content := range configFiles {
		tmpl, err := template.New(filename).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template for %s: %w", filename, err)
		}

		filepath := path.Join(configDir, filename)
		f, err := os.Create(filepath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}
		defer f.Close()

		if err := tmpl.Execute(f, cfg); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", filename, err)
		}
	}

	return nil
}

const envFileTemplate = `# {{.ProjectName}} environment variables
{{.EnvPrefix}}_CONFIG_FILE={{index .ConfigDirs 0}}/{{.ConfigFile}}
{{.EnvPrefix}}_CONFIG_FORMAT={{.ConfigFormat}}

# Server configuration
{{.EnvPrefix}}_SERVER_HOST=0.0.0.0
{{.EnvPrefix}}_SERVER_PORT=8080
{{.EnvPrefix}}_SERVER_READ_TIMEOUT=30s
{{.EnvPrefix}}_SERVER_WRITE_TIMEOUT=30s

# Database configuration
{{.EnvPrefix}}_DATABASE_HOST=localhost
{{.EnvPrefix}}_DATABASE_PORT=5432
{{.EnvPrefix}}_DATABASE_NAME={{.ProjectName}}
{{.EnvPrefix}}_DATABASE_USER=postgres
{{.EnvPrefix}}_DATABASE_PASSWORD=postgres
{{.EnvPrefix}}_DATABASE_SSL_MODE=disable

# Logger configuration
{{.EnvPrefix}}_LOGGER_LEVEL=info
{{.EnvPrefix}}_LOGGER_FORMAT=json
{{.EnvPrefix}}_LOGGER_OUTPUT=stdout

# Binary-specific ports (for docker-compose)
{{- range .Binaries}}
{{$.EnvPrefix}}_{{.}}_PORT=8080
{{- end}}
`

const sampleConfigTemplate = `# {{.ProjectName}} configuration file

server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  max_header_bytes: 1048576
  allowed_origins:
    - "*"

database:
  host: "localhost"
  port: 5432
  name: "{{.ProjectName}}"
  user: "postgres"
  password: ""
  ssl_mode: "disable"

logger:
  level: "info"
  format: "json"
  output: "stdout"
  fields:
    service: "{{.ProjectName}}"
`

func generateSampleConfig(projectPath string, cfg Config) error {
	configDir := filepath.Join(projectPath, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	tmpl, err := template.New("sample-config").Parse(sampleConfigTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse sample config template: %w", err)
	}

	filename := filepath.Join(configDir, cfg.ConfigFile)
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create sample config file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, cfg); err != nil {
		return fmt.Errorf("failed to execute sample config template: %w", err)
	}

	return nil
}
