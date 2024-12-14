package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type CLIFramework string

const (
	Cobra  CLIFramework = "cobra"
	Urfave CLIFramework = "urfave"

	baseCommandTemplate = `
{{define "base"}}
{{template "package" .}}

import (
	"context"
	"fmt"
	"log"
	"os"
	{{template "framework_imports" .}}
)

{{template "app_context" .}}
{{template "command_body" .}}
{{end}}

{{define "package"}}
package {{if eq (len .Binaries) 1}}commands{{else}}{{.PackageName}}{{end}}
{{end}}

{{define "app_context"}}
type AppContext struct {
	ConfigPath string
	Debug      bool
}

func NewAppContext() *AppContext {
	return &AppContext{}
}
{{end}}
`

	mainTemplate = `
{{define "main"}}
package main

import (
	"context"
	"fmt"
	"os"
	"{{.ModulePrefix}}/internal/commands{{if gt (len .Binaries) 1}}/{{.PackageName}}{{end}}"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	appCtx := {{if gt (len .Binaries) 1}}{{.PackageName}}{{else}}commands{{end}}.NewAppContext()

	if err := {{if gt (len .Binaries) 1}}{{.PackageName}}{{else}}commands{{end}}.Execute(ctx, appCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
{{end}}
`

	baseRootCommand = `
{{define "command_body"}}
// Common server implementation
func runServer(ctx context.Context, host string, port int) error {
	// Server implementation
	return nil
}

{{template "framework_specific" .}}
{{end}}
`

	baseVersionCommand = `
{{define "command_body"}}
{{template "framework_specific" .}}
{{end}}
`

	baseServerCommand = `
{{define "command_body"}}
{{template "framework_specific" .}}
{{end}}
`
)

var frameworkTemplates = map[CLIFramework]map[string]struct {
	Base          string
	FrameworkBody string
	Template      *template.Template
}{
	Cobra: {
		"root.go": {
			Base: baseRootCommand,
			Template: template.Must(template.New("root").Parse(`
{{define "framework_imports"}}
"github.com/spf13/cobra"
{{end}}
			`)),
			FrameworkBody: `{{define "framework_specific"}}
func CmdRoot(ctx context.Context, appCtx *AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "{{.Binary}}",
		Short: "{{.ProjectName}} CLI",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Setup logging, tracing, etc.
		},
	}

	cmd.PersistentFlags().StringVar(&appCtx.ConfigPath, "config", "", "config file path")
	cmd.PersistentFlags().BoolVar(&appCtx.Debug, "debug", false, "enable debug mode")

	cmd.AddCommand(
		CmdVersion(ctx, appCtx),
		CmdServer(ctx, appCtx),
	)

	return cmd
}

func Execute(ctx context.Context, appCtx *AppContext) error {
	cmd := CmdRoot(ctx, appCtx)
	return cmd.Execute()
}
{{end}}`,
		},
		"version.go": {
			Base: baseVersionCommand,
			Template: template.Must(template.New("version").Parse(`
{{define "framework_imports"}}
"github.com/spf13/cobra"
"{{.ModulePrefix}}/pkg/version"
{{end}}
			`)),
			FrameworkBody: `{{define "framework_specific"}}
func CmdVersion(ctx context.Context, appCtx *AppContext) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("Version: %s\n", version.Version)
			fmt.Printf("Commit: %s\n", version.Commit)
			fmt.Printf("Build Date: %s\n", version.BuildDate)
			return nil
		},
	}
}
{{end}}`,
		},
		"server.go": {
			Base: baseServerCommand,
			Template: template.Must(template.New("server").Parse(`
{{define "framework_imports"}}
"github.com/spf13/cobra"
{{end}}
			`)),
			FrameworkBody: `{{define "framework_specific"}}
func CmdServer(ctx context.Context, appCtx *AppContext) *cobra.Command {
	var (
		port int
		host string
	)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if appCtx.Debug {
				log.Println("Debug mode enabled")
			}
			return runServer(ctx, host, port)
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "server port")
	cmd.Flags().StringVar(&host, "host", "0.0.0.0", "server host")

	return cmd
}
{{end}}`,
		},
	},
	Urfave: {
		"root.go": {
			Base: baseRootCommand,
			Template: template.Must(template.New("root").Parse(`
{{define "framework_imports"}}
"github.com/urfave/cli/v2"
{{end}}
			`)),
			FrameworkBody: `{{define "framework_specific"}}
func CmdRoot(ctx context.Context, appCtx *AppContext) *cli.App {
	return &cli.App{
		Name:  "{{.Binary}}",
		Usage: "{{.ProjectName}} CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "config file path",
				Destination: &appCtx.ConfigPath,
			},
			&cli.BoolFlag{
				Name:        "debug",
				Usage:       "enable debug mode",
				Destination: &appCtx.Debug,
			},
		},
		Before: func(c *cli.Context) error {
			// Setup logging, tracing, etc.
			return nil
		},
		Commands: []*cli.Command{
			CmdVersion(ctx, appCtx),
			CmdServer(ctx, appCtx),
		},
	}
}

func Execute(ctx context.Context, appCtx *AppContext) error {
	app := CmdRoot(ctx, appCtx)
	return app.Run(os.Args)
}
{{end}}`,
		},
		"version.go": {
			Base: baseVersionCommand,
			Template: template.Must(template.New("version").Parse(`
{{define "framework_imports"}}
"github.com/urfave/cli/v2"
"{{.ModulePrefix}}/pkg/version"
{{end}}
			`)),
			FrameworkBody: `{{define "framework_specific"}}
func CmdVersion(ctx context.Context, appCtx *AppContext) *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Print version information",
		Action: func(c *cli.Context) error {
			fmt.Printf("Version: %s\n", version.Version)
			fmt.Printf("Commit: %s\n", version.Commit)
			fmt.Printf("Build Date: %s\n", version.BuildDate)
			return nil
		},
	}
}
{{end}}`,
		},
		"server.go": {
			Base: baseServerCommand,
			Template: template.Must(template.New("server").Parse(`
{{define "framework_imports"}}
"github.com/urfave/cli/v2"
{{end}}
			`)),
			FrameworkBody: `{{define "framework_specific"}}
func CmdServer(ctx context.Context, appCtx *AppContext) *cli.Command {
	var (
		port int
		host string
	)

	return &cli.Command{
		Name:  "server",
		Usage: "Start the server",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Value:       8080,
				Usage:       "server port",
				Destination: &port,
			},
			&cli.StringFlag{
				Name:        "host",
				Value:       "0.0.0.0",
				Usage:       "server host",
				Destination: &host,
			},
		},
		Action: func(c *cli.Context) error {
			if appCtx.Debug {
				log.Println("Debug mode enabled")
			}
			return runServer(ctx, host, port)
		},
	}
}
{{end}}`,
		},
	},
}

func generateCommandFiles(projectPath, binary string, cfg Config, cmdDir string, framework CLIFramework) error {
	if framework == "" {
		framework = "cobra"
	}

	log.Println("Generating command files for", binary, "with framework", framework)

	packageName := strings.Replace(binary, "-", "", -1)

	fullCmdDir := filepath.Join(projectPath, cmdDir)
	if len(cfg.Binaries) > 1 {
		fullCmdDir = filepath.Join(fullCmdDir, binary)
	}

	log.Println("Creating commands directory", fullCmdDir)
	if err := os.MkdirAll(fullCmdDir, 0755); err != nil {
		return fmt.Errorf("failed to create commands directory: %w", err)
	}

	frameworkTmpls, ok := frameworkTemplates[framework]
	if !ok {
		return fmt.Errorf("unsupported CLI framework: %s", framework)
	}

	log.Println("Generating main.go")

	data := struct {
		Binary       string
		PackageName  string
		ProjectName  string
		ModulePrefix string
		ConfigDirs   []string
		ConfigFile   string
		ConfigFormat string
		EnvPrefix    string
		Binaries     []string
	}{
		Binary:       binary,
		PackageName:  packageName,
		ProjectName:  cfg.ProjectName,
		ModulePrefix: cfg.ModulePrefix,
		ConfigDirs:   cfg.ConfigDirs,
		ConfigFile:   cfg.ConfigFile,
		ConfigFormat: cfg.ConfigFormat,
		EnvPrefix:    cfg.EnvPrefix,
		Binaries:     cfg.Binaries,
	}

	// Generate main.go
	if err := generateFileFromTemplate(
		filepath.Join(projectPath, "cmd", binary, "main.go"),
		mainTemplate,
		data,
	); err != nil {
		return fmt.Errorf("failed to generate main.go: %w", err)
	}

	if len(cfg.Binaries) > 1 {
		path := filepath.Join(projectPath, "internal", "commands", binary)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create commands directory: %w", err)
		}
	} else {
		path := filepath.Join(projectPath, "internal", "commands")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create commands directory: %w", err)
		}
	}

	// Generate command files
	for filename, tmpl := range frameworkTmpls {
		log.Println("Generating", filename)

		t := template.Must(template.New("base").Parse(baseCommandTemplate))
		t = template.Must(t.Parse(tmpl.Base))
		t = template.Must(t.Parse(tmpl.FrameworkBody))
		if tmpl.Template != nil {
			for _, templateData := range tmpl.Template.Templates() {
				t = template.Must(t.AddParseTree(templateData.Name(), templateData.Tree))
			}
		}

		// Ensure the directory exists
		if err := os.MkdirAll(fullCmdDir, 0755); err != nil {
			return fmt.Errorf("failed to create commands directory: %w", err)
		}

		var path string
		if len(cfg.Binaries) > 1 {
			path = filepath.Join(projectPath, "internal", "commands", binary)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if err := os.MkdirAll(path, 0755); err != nil {
					return fmt.Errorf("failed to create commands directory: %w", err)
				}
			}
		} else {
			path = filepath.Join(projectPath, "internal", "commands")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				if err := os.MkdirAll(path, 0755); err != nil {
					return fmt.Errorf("failed to create commands directory: %w", err)
				}
			}
		}

		fullPath := filepath.Join(path, filename)

		log.Println("Creating file", fullPath)

		// Generate the command file from the template
		f, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}
		defer f.Close()

		if err := t.ExecuteTemplate(f, "base", data); err != nil {
			return fmt.Errorf("failed to execute template for %s: %w", filename, err)
		}
	}

	return nil
}

func generateFileFromTemplate2(filepath string, tmpl *template.Template, data interface{}) error {
	return tmpl.ExecuteTemplate(os.Stdout, filepath, data)
}
