package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

type Config struct {
	ProjectName     string
	ModulePrefix    string
	Binaries        []string
	Includes        []string
	License         string
	GoVersion       string
	Author          string
	ConfigDirs      []string
	ConfigFile      string
	ConfigFormat    string
	EnvPrefix       string
	CLIFramework    string
	LicenseTemplate string
}

func main() {
	cfg := parseFlags()

	if err := generateProject(cfg); err != nil {
		fmt.Printf("Error generating project: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully generated project: %s\n", cfg.ProjectName)
}

func parseFlags() Config {
	name := flag.String("name", "", "Name of the project")
	module := flag.String("module", "", "Go module prefix (e.g., github.com/username)")
	bins := flag.String("binaries", "", "Comma-separated list of binaries to generate")
	include := flag.String("include", "", "Comma-separated list of features to include (server,cli,proto)")
	license := flag.String("license", "mit", "License type (mit, apache2, gpl3, bsd3, agpl3, lgpl3, mpl2, unlicense, custom)")
	goVer := flag.String("go", "1.21", "Go version to use")
	author := flag.String("author", "", "Author name for copyright")
	configDirs := flag.String("config-dirs", "", "Comma-separated list of config directories")
	configFile := flag.String("config-file", "config.yml", "Default config filename")
	configFormat := flag.String("config-format", "yml", "Default config format (yml, yaml, json, toml)")
	envPrefix := flag.String("env-prefix", "", "Environment variable prefix (defaults to project name)")
	cliFramework := flag.String("cli", "cobra", "CLI framework to use (cobra or urfave)")
	licenseTemplate := flag.String("license-template", "", "Path to custom license template file")

	flag.Parse()

	if *name == "" || *module == "" {
		fmt.Println("Please provide project name and module prefix")
		flag.Usage()
		os.Exit(1)
	}

	defaultConfigDirs := []string{
		fmt.Sprintf("/etc/%s", *name),
		fmt.Sprintf("$HOME/.config/%s", *name),
	}

	configDirsList := defaultConfigDirs
	if *configDirs != "" {
		configDirsList = strings.Split(*configDirs, ",")
	}

	prefix := *envPrefix
	if prefix == "" {
		prefix = strings.ToUpper(strings.Replace(*name, "-", "_", -1))
	}

	binaries := []string{}
	if *bins != "" {
		binaries = strings.Split(*bins, ",")
	}

	includes := []string{}
	if *include != "" {
		includes = strings.Split(*include, ",")
	}

	return Config{
		ProjectName:     *name,
		ModulePrefix:    *module,
		Binaries:        binaries,
		Includes:        includes,
		License:         *license,
		GoVersion:       *goVer,
		Author:          *author,
		ConfigDirs:      configDirsList,
		ConfigFile:      *configFile,
		ConfigFormat:    *configFormat,
		EnvPrefix:       prefix,
		CLIFramework:    *cliFramework,
		LicenseTemplate: *licenseTemplate,
	}
}

func generateProject(cfg Config) error {
	projectPath := cfg.ProjectName

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	if err := generateDirectoryStructure(projectPath, cfg); err != nil {
		return err
	}

	// Generate main.go files for each binary
	for _, binary := range cfg.Binaries {
		if err := generateMainFile(projectPath, binary, cfg); err != nil {
			return fmt.Errorf("failed to generate main file for %s: %w", binary, err)
		}
	}

	if err := generateCommonFiles(projectPath, cfg); err != nil {
		return err
	}

	if err := generateVersionPackage(projectPath, cfg); err != nil {
		return err
	}

	if err := generateDockerfiles(projectPath, cfg); err != nil {
		return err
	}

	for _, feature := range cfg.Includes {
		switch strings.ToLower(feature) {
		case "server":
			if err := generateServerFiles(projectPath, cfg); err != nil {
				return err
			}
		case "cli":
			if err := generateCLIFiles(projectPath, cfg); err != nil {
				return err
			}
		case "proto":
			if err := generateProtoFiles(projectPath, cfg); err != nil {
				return err
			}
		}
	}

	if err := generateConfigPackage(projectPath, cfg); err != nil {
		return fmt.Errorf("failed to generate config package: %w", err)
	}

	if err := generateSampleConfig(projectPath, cfg); err != nil {
		return fmt.Errorf("failed to generate sample config: %w", err)
	}

	return nil
}

func generateServerFiles(projectPath string, cfg Config) error {
	return nil
}

func generateCLIFiles(projectPath string, cfg Config) error {
	return nil
}

func generateProtoFiles(projectPath string, cfg Config) error {
	return nil
}

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

func generateDirectoryStructure(projectPath string, cfg Config) error {
	dirs := []string{
		"internal",
		"pkg",
		"pkg/version",
		"scripts",
		"hack",
		".github/workflows",
		".gitlab/ci",
		"docker",
		"dist",
	}

	for _, bin := range cfg.Binaries {
		dirs = append(dirs, filepath.Join("cmd", bin))
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(projectPath, dir), 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Create .keep files in dist and hack directories
	keepDirs := []string{"dist", "hack"}
	for _, dir := range keepDirs {
		keepFile := filepath.Join(projectPath, dir, ".keep")
		if err := os.WriteFile(keepFile, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create .keep file in %s: %w", dir, err)
		}
	}

	return nil
}

const versionPackageTemplate = `// Package version provides build and version information for the application.
// This information is populated at build time using -ldflags.
package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Build information, populated at build-time via -ldflags.
var (
	// Version indicates the current version of the application.
	// For releases, this will be a semantic version (e.g., "v1.2.3").
	// For development builds, this will be "dev" or a commit hash.
	Version = "dev"

	// GitCommit is the full git commit hash.
	GitCommit = "unknown"

	// GitBranch is the git branch from which the application was built.
	GitBranch = "unknown"

	// BuildTime is the UTC timestamp of when the binary was built.
	BuildTime = "unknown"

	// BuildUser is the username of who built the binary.
	BuildUser = "unknown"

	// GoVersion is the version of Go used to build the application.
	GoVersion = runtime.Version()

	// Platform is the target platform (OS/architecture combination).
	Platform = runtime.GOOS + "/" + runtime.GOARCH
)

// Info holds all version information.
type Info struct {
	Version      string            ` + "`json:\"version\"`" + `
	GitCommit    string            ` + "`json:\"gitCommit\"`" + `
	GitBranch    string            ` + "`json:\"gitBranch\"`" + `
	BuildTime    string            ` + "`json:\"buildTime\"`" + `
	BuildUser    string            ` + "`json:\"buildUser\"`" + `
	GoVersion    string            ` + "`json:\"goVersion\"`" + `
	Platform     string            ` + "`json:\"platform\"`" + `
	Dependencies map[string]string ` + "`json:\"dependencies,omitempty\"`" + `
}

// Get returns the version information as a structured object.
func Get() Info {
	return Info{
		Version:      Version,
		GitCommit:    GitCommit,
		GitBranch:    GitBranch,
		BuildTime:    BuildTime,
		BuildUser:    BuildUser,
		GoVersion:    GoVersion,
		Platform:     Platform,
		Dependencies: getDependencyVersions(),
	}
}

// String returns a human-readable version string.
func String() string {
	info := Get()
	return fmt.Sprintf(
		"Version:      %s\n"+
			"Git Commit:   %s\n"+
			"Git Branch:   %s\n"+
			"Built:        %s\n"+
			"Built By:     %s\n"+
			"Go Version:   %s\n"+
			"Platform:     %s",
		info.Version,
		info.GitCommit,
		info.GitBranch,
		info.BuildTime,
		info.BuildUser,
		info.GoVersion,
		info.Platform,
	)
}

// JSON returns version information in JSON format.
func JSON(indent bool) string {
	info := Get()
	var data []byte
	var err error

	if indent {
		data, err = json.MarshalIndent(info, "", "  ")
	} else {
		data, err = json.Marshal(info)
	}

	if err != nil {
		return fmt.Sprintf("{\n  \"error\": \"%s\"\n}", err)
	}
	return string(data)
}

// Short returns a condensed version string.
func Short() string {
	commit := GitCommit
	if len(commit) > 8 {
		commit = commit[:8]
	}
	if Version == "dev" {
		return fmt.Sprintf("dev+%s", commit)
	}
	return fmt.Sprintf("%s+%s", Version, commit)
}

// IsRelease returns true if the current version represents a release build.
func IsRelease() bool {
	return Version != "dev" && strings.HasPrefix(Version, "v")
}

// IsDevelopment returns true if this is a development build.
func IsDevelopment() bool {
	return Version == "dev"
}

// getDependencyVersions returns a map of direct dependencies and their versions.
func getDependencyVersions() map[string]string {
	deps := make(map[string]string)
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return deps
	}

	for _, dep := range info.Deps {
		// Only include direct dependencies
		if !strings.Contains(dep.Path, "/internal/") {
			deps[dep.Path] = dep.Version
		}
	}
	return deps
}

// BuildContext returns a map of build-time variables.
func BuildContext() map[string]string {
	return map[string]string{
		"version":    Version,
		"gitCommit": GitCommit,
		"gitBranch": GitBranch,
		"buildTime": BuildTime,
		"buildUser": BuildUser,
		"goVersion": GoVersion,
		"platform":  Platform,
	}
}

// Validate checks if the version information appears to be properly populated.
func Validate() error {
	if Version == "dev" && GitCommit == "unknown" {
		return fmt.Errorf("version information not properly initialized")
	}
	return nil
}
`

func generateVersionPackage(projectPath string, cfg Config) error {
	return os.WriteFile(
		filepath.Join(projectPath, "pkg/version/version.go"),
		[]byte(versionPackageTemplate),
		0644,
	)
}

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

// Also add an .air.toml configuration template
const airConfigTemplate = `root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/{{.Binary}}"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_error = true

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false`

func generateDockerfiles(projectPath string, cfg Config) error {
	// Generate development Dockerfiles
	if err := generateDevDockerfiles(projectPath, cfg); err != nil {
		return fmt.Errorf("failed to generate dev dockerfiles: %w", err)
	}

	// Generate production Dockerfiles
	if err := generateProdDockerfiles(projectPath, cfg); err != nil {
		return fmt.Errorf("failed to generate prod dockerfiles: %w", err)
	}

	// Generate docker-compose.yml in docker directory
	dockerDir := filepath.Join(projectPath, "docker")
	if err := generateFileFromTemplate(
		filepath.Join(dockerDir, "docker-compose.yml"),
		dockerComposeTemplate,
		cfg,
	); err != nil {
		return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
	}

	return nil
}

func generateDevDockerfiles(projectPath string, cfg Config) error {
	dockerDir := filepath.Join(projectPath, "docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		return fmt.Errorf("failed to create docker directory: %w", err)
	}

	tmpl, err := template.New("dockerfile").Parse(devDockerfileTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse dockerfile template: %w", err)
	}

	for _, binary := range cfg.Binaries {
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

		fileName := filepath.Join(dockerDir, binary+".Dockerfile")
		f, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create dockerfile for %s: %w", binary, err)
		}
		defer f.Close()

		if err := tmpl.Execute(f, data); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", binary, err)
		}
	}

	return nil
}

func generateProdDockerfiles(projectPath string, cfg Config) error {
	dockerDir := filepath.Join(projectPath, "build", "docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		return fmt.Errorf("failed to create docker directory: %w", err)
	}

	tmpl, err := template.New("dockerfile").Parse(prodDockerfileTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse dockerfile template: %w", err)
	}

	for _, binary := range cfg.Binaries {
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

		fileName := filepath.Join(dockerDir, binary+".Dockerfile")
		f, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create dockerfile for %s: %w", binary, err)
		}
		defer f.Close()

		if err := tmpl.Execute(f, data); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", binary, err)
		}
	}

	return nil
}

const makefileTemplate = `BINARIES ?= {{range .Binaries}}{{.}} {{end}}

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse HEAD)
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -X '{{.ModulePrefix}}/pkg/version.Version=${VERSION}' \
           -X '{{.ModulePrefix}}/pkg/version.GitCommit=${COMMIT}' \
           -X '{{.ModulePrefix}}/pkg/version.BuildTime=${DATE}'

.PHONY: all
all: deps build

.PHONY: deps
deps:
	go mod download

.PHONY: build
build:
	@for binary in $(BINARIES); do \
		echo "Building $$binary..." ; \
		go build -ldflags "$(LDFLAGS)" -o bin/$$binary ./cmd/$$binary ; \
	done

.PHONY: test
test:
	go test -v -race -cover ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: docker
docker:
	@for binary in $(BINARIES); do \
		echo "Building Docker image for $$binary..." ; \
		docker build \
			--build-arg VERSION=$(VERSION) \
			--build-arg COMMIT=$(COMMIT) \
			--build-arg BUILD_TIME=$(DATE) \
			-t {{.ProjectName}}-$$binary:$(VERSION) \
			-f docker/$$binary.Dockerfile . ; \
	done

.PHONY: proto
proto:
	@if [ -d "proto" ]; then \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			proto/*.proto ; \
	fi

.PHONY: clean
clean:
	rm -rf bin/
	rm -rf dist/
`

const mainTemplate = `package main

import (
{{- if eq (len .Binaries) 1}}
	"{{.ModulePrefix}}/internal/commands"
{{- else}}
	"{{.ModulePrefix}}/internal/commands/{{.Binary}}"
{{- end}}
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
`

func generateMainFile(projectPath, binary string, cfg Config) error {
	// Convert kebab-case to valid package name
	packageName := strings.Replace(binary, "-", "", -1)

	// Create the binary directory if it doesn't exist
	binaryDir := filepath.Join(projectPath, "cmd", binary)
	if err := os.MkdirAll(binaryDir, 0755); err != nil {
		return fmt.Errorf("failed to create binary directory: %w", err)
	}

	// Parse the template
	tmpl, err := template.New("main").Funcs(template.FuncMap{
		"cleanPackageName": cleanPackageName,
	}).Parse(mainTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse main template: %w", err)
	}

	// Create the main.go file
	mainFile := filepath.Join(binaryDir, "main.go")
	f, err := os.Create(mainFile)
	if err != nil {
		return fmt.Errorf("failed to create main.go: %w", err)
	}
	defer f.Close()

	// Execute the template with the config data
	data := struct {
		Binary       string
		PackageName  string
		ModulePrefix string
		ProjectName  string
		Binaries     []string
	}{
		Binary:       binary,
		PackageName:  packageName,
		ModulePrefix: cfg.ModulePrefix,
		ProjectName:  cfg.ProjectName,
		Binaries:     cfg.Binaries,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute main template: %w", err)
	}

	return nil
}

const goModTemplate = `module {{.ModulePrefix}}

go {{.GoVersion}}
`

func generateGoMod(cfg Config) string {
	tmpl, err := template.New("gomod").Parse(goModTemplate)
	if err != nil {
		// Return empty string or handle error as needed
		return ""
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		// Return empty string or handle error as needed
		return ""
	}

	return buf.String()
}

func generateReadme(cfg Config) string {
	const readmeTemplate = `# {{.ProjectName}}

## Overview

This project was generated using a Go project generator.

## Project Structure

### Binaries
{{range .Binaries}}- {{.}}
{{end}}
### Features
{{range .Includes}}- {{.}}
{{end}}
## Requirements

- Go {{.GoVersion}} or higher
- Docker (optional)
- Make

## Getting Started

### Installation

` + "```" + `bash
go get {{.ModulePrefix}}
` + "```" + `

### Building

Build all binaries:
` + "```" + `bash
make build
` + "```" + `

Run tests:
` + "```" + `bash
make test
` + "```" + `

Run linter:
` + "```" + `bash
make lint
` + "```" + `

### Docker

Build the Docker images:
` + "```" + `bash
make docker
` + "```" + `

### Development

The project structure follows the standard Go project layout:

- /cmd - Main applications
- /internal - Private application and library code
- /pkg - Library code that's ok to use by external applications
- /hack - Tools and scripts to help with development
- /scripts - Scripts for CI/CD and other automation

## License

This project is licensed under the {{.License | ToUpper}} License - see the LICENSE file for details.`

	tmpl, err := template.New("readme").Funcs(template.FuncMap{
		"ToUpper": strings.ToUpper,
	}).Parse(readmeTemplate)
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, cfg); err != nil {
		return ""
	}

	return buf.String()
}

const (
	mitLicense = `MIT License

Copyright (c) {{.Year}} {{.Fullname}}

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.`

	apache2License = `                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

   Copyright {{.Year}} {{.Fullname}}

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.`

	gpl3License = `                    GNU GENERAL PUBLIC LICENSE
                       Version 3, 29 June 2007

 Copyright (C) {{.Year}} {{.Fullname}}

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU General Public License as published by
 the Free Software Foundation, either version 3 of the License, or
 (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU General Public License for more details.

 You should have received a copy of the GNU General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.`

	bsd3License = `BSD 3-Clause License

Copyright (c) {{.Year}}, {{.Fullname}}
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its
   contributors may be used to endorse or promote products derived from
   this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.`

	agpl3License = `                    GNU AFFERO GENERAL PUBLIC LICENSE
                       Version 3, 19 November 2007

Copyright (C) {{.Year}} {{.Fullname}}

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.`

	lgpl3License = `                   GNU LESSER GENERAL PUBLIC LICENSE
                       Version 3, 29 June 2007

Copyright (C) {{.Year}} {{.Fullname}}

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.`

	mpl2License = `Mozilla Public License Version 2.0
==================================

1. Definitions
--------------

1.1. "Contributor"
    means each individual or legal entity that creates, contributes to
    the creation of, or owns Covered Software.

Copyright (C) {{.Year}} {{.Fullname}}

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at http://mozilla.org/MPL/2.0/.`

	unlicenseLicense = `This is free and unencumbered software released into the public domain.

Anyone is free to copy, modify, publish, use, compile, sell, or
distribute this software, either in source code form or as a compiled
binary, for any purpose, commercial or non-commercial, and by any
means.

Copyright (C) {{.Year}} {{.Fullname}}

To the extent possible under law, the author has dedicated all copyright
and related and neighboring rights to this software to the public domain
worldwide. This software is distributed without any warranty.

You should have received a copy of the CC0 Public Domain Dedication along
with this software. If not, see <http://creativecommons.org/publicdomain/zero/1.0/>.`
)

func generateLicense(cfg Config, licenseTemplatePath string) string {
	// If a custom template is provided, use it
	if licenseTemplatePath != "" {
		content, err := os.ReadFile(licenseTemplatePath)
		if err != nil {
			fmt.Printf("Warning: Could not read license template file: %v\n", err)
			return getDefaultLicense(cfg)
		}

		tmpl, err := template.New("custom-license").Parse(string(content))
		if err != nil {
			fmt.Printf("Warning: Could not parse license template: %v\n", err)
			return getDefaultLicense(cfg)
		}

		data := struct {
			Year         int
			Fullname     string
			ProjectName  string
			ModulePrefix string
		}{
			Year:         time.Now().Year(),
			Fullname:     cfg.Author,
			ProjectName:  cfg.ProjectName,
			ModulePrefix: cfg.ModulePrefix,
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			fmt.Printf("Warning: Could not execute license template: %v\n", err)
			return getDefaultLicense(cfg)
		}

		return buf.String()
	}

	// Otherwise use built-in licenses
	var licenseText string
	switch strings.ToLower(cfg.License) {
	case "mit":
		licenseText = mitLicense
	case "apache2", "apache-2.0":
		licenseText = apache2License
	case "gpl3", "gpl-3.0":
		licenseText = gpl3License
	case "bsd3", "bsd-3-clause":
		licenseText = bsd3License
	case "agpl3", "agpl-3.0":
		licenseText = agpl3License
	case "lgpl3", "lgpl-3.0":
		licenseText = lgpl3License
	case "mpl2", "mpl-2.0":
		licenseText = mpl2License
	case "unlicense":
		licenseText = unlicenseLicense
	default:
		return getDefaultLicense(cfg)
	}

	tmpl := template.New("license")
	tmpl, err := tmpl.Parse(licenseText)
	if err != nil {
		return ""
	}

	data := struct {
		Year     int
		Fullname string
	}{
		Year:     time.Now().Year(),
		Fullname: cfg.Author,
	}

	// Use module prefix if author not provided
	if data.Fullname == "" {
		data.Fullname = cfg.ModulePrefix
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return ""
	}

	return buf.String()
}

func getDefaultLicense(cfg Config) string {
	const licenseTemplate = `# {{.ProjectName}}

## Overview

This project was generated using a Go project generator.

## Project Structure

### Binaries
- {{.Binaries}}

### Features
- {{.Features}}

## Requirements

- Go {{.GoVersion}} or higher
- Docker (optional)
- Make

## Getting Started

### Installation

` + "```" + `bash
go get {{.ModulePrefix}}
` + "```" + `

### Building

Build all binaries:
` + "```" + `bash
make build
` + "```" + `

Run tests:
` + "```" + `bash
make test
` + "```" + `

Run linter:
` + "```" + `bash
make lint
` + "```" + `

### Docker

Build the Docker images:
` + "```" + `bash
make docker
` + "```" + `

### Development

The project structure follows the standard Go project layout:

- /cmd - Main applications
- /internal - Private application and library code
- /pkg - Library code that's ok to use by external applications
- /hack - Tools and scripts to help with development
- /scripts - Scripts for CI/CD and other automation

## License

This project is licensed under the {{.License}} License - see the LICENSE file for details.`

	tmpl := template.New("license")
	tmpl, err := tmpl.Parse(licenseTemplate)
	if err != nil {
		return ""
	}

	// Convert slices to strings
	binariesStr := strings.Join(cfg.Binaries, "\n- ")
	featuresStr := strings.Join(cfg.Includes, "\n- ")

	data := struct {
		ProjectName  string
		Binaries     string
		Features     string
		GoVersion    string
		ModulePrefix string
		License      string
	}{
		ProjectName:  cfg.ProjectName,
		Binaries:     binariesStr,
		Features:     featuresStr,
		GoVersion:    cfg.GoVersion,
		ModulePrefix: cfg.ModulePrefix,
		License:      strings.ToUpper(cfg.License),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return ""
	}

	return buf.String()
}

func generateCommandsPackage(projectPath string, cfg Config) error {
	if len(cfg.Binaries) == 1 {
		// Single binary: create commands directly in internal/commands/
		return generateCommandFiles(projectPath, cfg.Binaries[0], cfg, "internal/commands")
	}

	// Multiple binaries: create commands in internal/commands/<binary>/
	for _, binary := range cfg.Binaries {
		cmdDir := filepath.Join("internal/commands", binary)
		if err := generateCommandFiles(projectPath, binary, cfg, cmdDir); err != nil {
			return fmt.Errorf("failed to generate commands for %s: %w", binary, err)
		}
	}

	return nil
}

func generateCommandFiles(projectPath, binary string, cfg Config, cmdDir string) error {
	// Convert kebab-case to valid package name
	packageName := strings.Replace(binary, "-", "", -1)

	fullCmdDir := filepath.Join(projectPath, cmdDir)
	if err := os.MkdirAll(fullCmdDir, 0755); err != nil {
		return fmt.Errorf("failed to create commands directory: %w", err)
	}

	data := struct {
		Binary       string
		PackageName  string
		ProjectName  string
		ModulePrefix string
		ConfigDirs   []string
		ConfigFile   string
		ConfigFormat string
		EnvPrefix    string
	}{
		Binary:       binary,
		PackageName:  packageName,
		ProjectName:  cfg.ProjectName,
		ModulePrefix: cfg.ModulePrefix,
		ConfigDirs:   cfg.ConfigDirs,
		ConfigFile:   cfg.ConfigFile,
		ConfigFormat: cfg.ConfigFormat,
		EnvPrefix:    cfg.EnvPrefix,
	}

	var templates map[string]string
	if cfg.CLIFramework == "urfave" {
		templates = map[string]string{
			"root.go":    urfaveMainTemplate,
			"version.go": urfaveVersionCommandTemplate,
			"server.go":  urfaveServerCommandTemplate,
		}
	} else {
		templates = map[string]string{
			"root.go":    rootCommandTemplate,
			"version.go": versionCommandTemplate,
			"server.go":  serverCommandTemplate,
		}
	}

	// Generate command files
	for filename, tmpl := range templates {
		if err := generateFileFromTemplate(
			filepath.Join(fullCmdDir, filename),
			tmpl,
			data,
		); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
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

const mainBinaryTemplate = `package main

import (
{{- if eq (len .Binaries) 1}}
	"{{.ModulePrefix}}/internal/commands"
{{- else}}
	"{{.ModulePrefix}}/internal/commands/{{.Binary}}"
{{- end}}
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
`

func generateFileFromTemplate(filepath, tmplContent string, data interface{}) error {
	tmpl, err := template.New(path.Base(filepath)).Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}

func generateConfigPackage(projectPath string, cfg Config) error {
	configDir := filepath.Join(projectPath, "internal/config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Map of config files to their templates
	configFiles := map[string]string{
		"config.go": `package config

import (
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

const (
	rootCommandTemplate = `package commands

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)
`

	versionCommandTemplate = `package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"{{.ModulePrefix}}/pkg/version"
)
`

	serverCommandTemplate = `package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)
`

	urfaveMainTemplate = `package commands

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)
`

	urfaveVersionCommandTemplate = `package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"{{.ModulePrefix}}/pkg/version"
)
`

	urfaveServerCommandTemplate = `package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"
)
`
)

func generateCommonFiles(projectPath string, cfg Config) error {
	// Map of filenames to their content generation functions
	files := map[string]func(Config) string{
		"README.md": generateReadme,
		"LICENSE": func(c Config) string {
			return generateLicense(c, c.LicenseTemplate)
		},
		"go.mod": generateGoMod,
		"Makefile": func(c Config) string {
			tmpl, err := template.New("makefile").Parse(makefileTemplate)
			if err != nil {
				return ""
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, c); err != nil {
				return ""
			}
			return buf.String()
		},
		".gitignore": func(c Config) string { return defaultGitignore },
		".env": func(c Config) string {
			tmpl, err := template.New("env").Parse(envFileTemplate)
			if err != nil {
				return ""
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, c); err != nil {
				return ""
			}
			return buf.String()
		},
		".air.toml": func(c Config) string {
			binary := c.Binaries[0]
			tmpl, err := template.New("air").Parse(airConfigTemplate)
			if err != nil {
				return ""
			}
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, struct{ Binary string }{Binary: binary}); err != nil {
				return ""
			}
			return buf.String()
		},
	}

	// Generate each file
	for filename, generator := range files {
		content := generator(cfg)
		if content == "" {
			continue // Skip if generator returned empty content
		}

		filepath := path.Join(projectPath, filename)

		// Ensure directory exists for the file
		if dir := path.Dir(filepath); dir != "." {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory for %s: %w", filename, err)
			}
		}

		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	// Generate commands package
	if err := generateCommandsPackage(projectPath, cfg); err != nil {
		return fmt.Errorf("failed to generate commands package: %w", err)
	}

	// Add required dependencies to go.mod based on CLI framework
	if cfg.CLIFramework == "urfave" {
		// Add urfave/cli dependency
		// TODO: Implement go.mod modification for urfave/cli
	} else {
		// Add cobra dependency
		// TODO: Implement go.mod modification for cobra
	}

	return nil
}

const defaultGitignore = `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
vendor/

# IDE specific files
.idea/
.vscode/
*.swp
*.swo

# OS specific files
.DS_Store
.env.local
.env.*.local

# Log files
*.log

# Temporary files
tmp/
temp/
`

func cleanPackageName(name string) string {
	return strings.Replace(name, "-", "", -1)
}
