package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/edsonmichaque/craft"
)

//go:embed templates
var templates embed.FS

func main() {
	allGenerators := map[string]craft.Generator{
		"config":   craft.GenerateConfig,
		"docker":   craft.GenerateDockerFiles,
		"script":   craft.GenerateScripts,
		"license":  craft.GenerateLicense,
		"commands": craft.GenerateCommands,
		"version":  craft.GenerateVersion,
		"common":   craft.GenerateCommonFiles,
	}

	data := parseFlags()

	manager := craft.Manager{
		Generators: allGenerators,
		Options: craft.Options{
			Templates: templates,
		},
	}

	gen := make([]string, 0)

	for g := range allGenerators {
		gen = append(gen, g)
	}

	files, err := manager.Generate(context.Background(), data, gen...)
	if err != nil {
		fmt.Println("failed to generate files: %w", err)
		os.Exit(1)
	}

	createdDirs := make(map[string]bool)
	createdFiles := []string{}

	defer func() {
		if err != nil {
			for _, file := range createdFiles {
				os.Remove(file)
			}
			for dir := range createdDirs {
				os.RemoveAll(dir)
			}
		}
	}()

	for k, v := range files {
		fullPath := filepath.Join(data.ProjectName, k)

		if strings.Contains(fullPath, "/") {
			dir := filepath.Dir(fullPath)
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					fmt.Printf("Failed to create directory %s: %v\n", dir, err)
					os.Exit(1)
				}
				createdDirs[dir] = true
			}
		}

		if err := os.WriteFile(fullPath, v, 0644); err != nil {
			fmt.Printf("Failed to write file %s: %v\n", fullPath, err)
			os.Exit(1)
		}

		createdFiles = append(createdFiles, fullPath)
	}
}

func parseFlags() craft.Data {
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

	return craft.Data{
		ProjectName:  *name,
		ModulePrefix: *module,
		Binaries:     binaries,
		Includes:     includes,
		License:      *license,
		GoVersion:    *goVer,
		Author:       *author,
		ConfigDirs:   configDirsList,
		ConfigFile:   *configFile,
		ConfigFormat: *configFormat,
		EnvPrefix:    prefix,
		CLI: craft.CLI{
			Framework: *cliFramework,
		},
		Module:      *module,
		AppName:     *name,
		Description: *name,
		Commands:    []string{},
	}
}
