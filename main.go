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
	Module          string
	AppName         string
	Description     string
	Commands        []string
	CIServices      []string
}

func main() {
	cfg := parseFlags()

	if err := generateProject(cfg); err != nil {
		fmt.Printf("Error generating project: %v\n", err)
		os.Exit(1)
	}
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
	ciServices := flag.String("ci", "github", "Comma-separated list of CI services to use (e.g., github, gitlab, circleci)")

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

	ciServicesList := strings.Split(*ciServices, ",")

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
		Module:          *module,
		AppName:         *name,
		Description:     *name,
		Commands:        []string{},
		CIServices:      ciServicesList,
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

	// Add this line to generate CI scripts
	if err := generateCIScripts(projectPath, cfg); err != nil {
		return fmt.Errorf("failed to generate CI scripts: %w", err)
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

// const mainTemplate = `package main

// import (
// {{- if eq (len .Binaries) 1}}
// 	"{{.ModulePrefix}}/internal/commands"
// {{- else}}
// 	"{{.ModulePrefix}}/internal/commands/{{.Binary}}"
// {{- end}}
// )

// func main() {
// 	if err := commands.Execute(); err != nil {
// 		os.Exit(1)
// 	}
// }
// `

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

const mainBinaryTemplate = `package main

import (
{{- if eq (len .Binaries) 1}}
	"{{.ModulePrefix}}/internal/commands"
{{- else}}
	"{{.ModulePrefix}}/internal/commands/{{.Binary}}"
{{- end}}
)

func main() {
	if err := {{- if eq (len .Binaries) 1}}commands{{else}}{{.Binary}}{{end}}.Execute(); err != nil {
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

func generateCommonFiles(projectPath string, cfg Config) error {
	// Map of filenames to their content generation functions
	files := map[string]struct {
		template string
		mode     os.FileMode
	}{
		"README.md":  {generateReadme(cfg), 0644},
		"LICENSE":    {generateLicense(cfg, cfg.LicenseTemplate), 0644},
		"go.mod":     {generateGoMod(cfg), 0644},
		"Makefile":   {executeTemplate(makefileTemplate, cfg), 0644},
		".gitignore": {defaultGitignore, 0644},
		".env":       {executeTemplate(envFileTemplate, cfg), 0644},
		".air.toml":  {executeTemplate(airConfigTemplate, struct{ Binary string }{Binary: cfg.Binaries[0]}), 0644},
	}

	for filename, file := range files {
		filepath := path.Join(projectPath, filename)
		if err := os.MkdirAll(path.Dir(filepath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", filename, err)
		}
		if err := os.WriteFile(filepath, []byte(file.template), file.mode); err != nil {
			return fmt.Errorf("failed to write %s: %w", filename, err)
		}
	}

	return nil
}

func executeTemplate(tmpl string, data interface{}) string {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return ""
	}
	return buf.String()
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
