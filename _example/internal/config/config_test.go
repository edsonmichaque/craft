package config

import (
	"testing"
	"os"
	"path/filepath"
)

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yml")
	
	configContent := []byte(`
server:
  host: "127.0.0.1"
  port: 8080
database:
  host: "localhost"
  port: 5432
logger:
  level: "debug"
`)
	
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
}