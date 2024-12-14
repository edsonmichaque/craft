package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const devDockerfileTemplate = `# Development image with live reload
FROM golang:{{.GoVersion}}-alpine

# Install development tools and build dependencies
RUN apk add --no-cache git make curl \
    && go install github.com/cosmtrek/air@latest

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Set environment variables
ENV {{.EnvPrefix}}_CONFIG_FILE=/app/config/config.yml \
    CGO_ENABLED=0 \
    GO111MODULE=on

# Expose default port
EXPOSE 8080

# Use air for live reload
ENTRYPOINT ["air", "-c", ".air.toml"]`

const prodDockerfileTemplate = `# Production image - using distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot

# Copy the pre-built binary from dist directory
COPY dist/{{.Binary}} /{{.Binary}}

# Copy config files
COPY config/ /etc/{{.ProjectName}}/

# Set environment variables
ENV {{.EnvPrefix}}_CONFIG_FILE=/etc/{{.ProjectName}}/config.yml

# Use non-root user
USER nonroot:nonroot

# Expose default port
EXPOSE 8080

ENTRYPOINT ["/{{.Binary}}"]`

const dockerComposeTemplate = `version: '3.8'

services:
{{- range .Binaries }}
  {{.}}:
    build:
      context: .
      dockerfile: docker/{{.}}.Dockerfile
    volumes:
      - .:/app
      - go-mod-cache:/go/pkg/mod
    env_file:
      - .env
    environment:
      - {{$.EnvPrefix}}_CONFIG_FILE=/app/config/{{$.ConfigFile}}
    ports:
      - "${PORT:-8080}:8080"
    depends_on:
      - postgres
    networks:
      - {{$.ProjectName}}-network

{{- end}}
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: {{.ProjectName}}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - {{.ProjectName}}-network

volumes:
  postgres_data:
  go-mod-cache:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

func generateDockerfiles(projectPath string, cfg Config) error {
	// Create docker directory for dev files
	dockerDir := filepath.Join(projectPath, "docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		return fmt.Errorf("failed to create docker directory: %w", err)
	}

	// Create build/docker directory for prod files
	buildDockerDir := filepath.Join(projectPath, "build", "docker")
	if err := os.MkdirAll(buildDockerDir, 0755); err != nil {
		return fmt.Errorf("failed to create build/docker directory: %w", err)
	}

	// Generate development Dockerfiles in docker/
	if err := generateDevDockerfiles(dockerDir, cfg); err != nil {
		return fmt.Errorf("failed to generate dev dockerfiles: %w", err)
	}

	// Generate production Dockerfiles in build/docker/
	if err := generateProdDockerfiles(buildDockerDir, cfg); err != nil {
		return fmt.Errorf("failed to generate prod dockerfiles: %w", err)
	}

	// Generate docker-compose in docker/
	composeFile := filepath.Join(dockerDir, "docker-compose.yml")
	data := struct {
		Config
		ConfigFile string
	}{
		Config:     cfg,
		ConfigFile: "config.yml",
	}

	if err := generateFileFromTemplate(composeFile, dockerComposeTemplate, data); err != nil {
		return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
	}

	return nil
}

func generateDevDockerfiles(dockerDir string, cfg Config) error {
	tmpl, err := template.New("dockerfile").Parse(devDockerfileTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse dev dockerfile template: %w", err)
	}

	for _, binary := range cfg.Binaries {
		fileName := filepath.Join(dockerDir, binary+".Dockerfile")
		if err := generateSingleDockerfile(fileName, tmpl, cfg, binary); err != nil {
			return fmt.Errorf("failed to generate dev dockerfile for %s: %w", binary, err)
		}
	}

	return nil
}

func generateProdDockerfiles(buildDockerDir string, cfg Config) error {
	tmpl, err := template.New("dockerfile").Parse(prodDockerfileTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse prod dockerfile template: %w", err)
	}

	for _, binary := range cfg.Binaries {
		fileName := filepath.Join(buildDockerDir, binary+".Dockerfile")
		if err := generateSingleDockerfile(fileName, tmpl, cfg, binary); err != nil {
			return fmt.Errorf("failed to generate prod dockerfile for %s: %w", binary, err)
		}
	}

	return nil
}

// Helper function to generate a single Dockerfile
func generateSingleDockerfile(fileName string, tmpl *template.Template, cfg Config, binary string) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fileName, err)
	}
	defer f.Close()

	data := struct {
		Binary       string
		GoVersion    string
		ModulePrefix string
		ProjectName  string
		EnvPrefix    string
	}{
		Binary:       binary,
		GoVersion:    cfg.GoVersion,
		ModulePrefix: cfg.ModulePrefix,
		ProjectName:  cfg.ProjectName,
		EnvPrefix:    cfg.EnvPrefix,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
