package config

import "time"

// ServerConfig holds all server-related configuration
type ServerConfig struct {
	Host           string        `mapstructure:"host" yaml:"host"`
	Port           int           `mapstructure:"port" yaml:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout" yaml:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout" yaml:"write_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes" yaml:"max_header_bytes"`
	AllowedOrigins []string      `mapstructure:"allowed_origins" yaml:"allowed_origins"`
}

// GetAddress returns the full address string for the server
func (c ServerConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}