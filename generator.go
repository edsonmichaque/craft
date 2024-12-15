package craft

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"
)

type CLI struct {
	Framework string
}
type Data struct {
	CLI
	Binaries []string
	License  string

	ProjectName  string
	ModulePrefix string
	Includes     []string
	GoVersion    string
	Author       string
	ConfigDirs   []string
	ConfigFile   string
	ConfigFormat string
	EnvPrefix    string
	Module       string
	AppName      string
	Description  string
	Commands     []string
}

type RenderOptions struct {
	Templates []string
	Data      interface{}
	Execute   string
}

// Generator interface with Configure and Generate methods
type Generator func(data Data) (map[string]RenderOptions, error)

type Manager struct {
	Options    Options
	Generators map[string]Generator
}

func (g *Manager) Configure(options Options) error {
	g.Options = options
	return nil
}

func (g *Manager) generateFiles(ctx context.Context, m map[string]RenderOptions) (map[string][]byte, error) {
	generatedFiles := make(map[string][]byte)

	for dst, opts := range m {
		tpl := template.New(dst)
		tpl.Funcs(template.FuncMap{
			"ToUpper": strings.ToUpper,
		})

		for _, tmpl := range opts.Templates {
			tmplPath := filepath.Join("templates", tmpl)

			tmplFile, err := g.Options.Templates.Open(tmplPath)
			if err != nil {
				log.Printf("Error opening template %s: %v", tmplPath, err)

				return nil, fmt.Errorf("failed to open template %s: %w", tmplPath, err)
			}

			content, err := io.ReadAll(tmplFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read template %s: %v", tmplPath, err)
			}

			tpl, err = tpl.Parse(string(content))
			if err != nil {
				return nil, fmt.Errorf("failed to parse template %s: %w", tmplPath, err)
			}
		}

		content, err := g.generateFile(ctx, dst, tpl, opts.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to generate file for template %s: %w", opts, err)
		}
		for k, v := range content {
			generatedFiles[k] = v
		}
	}

	return generatedFiles, nil
}

func (g *Manager) generateFile(_ context.Context, dst string, template *template.Template, data interface{}) (map[string][]byte, error) {
	buf := bytes.NewBuffer(nil)

	// Use reflection to check if the data has an "Execute" field
	v := reflect.ValueOf(data)
	executeField := v.FieldByName("Execute")

	if executeField.IsValid() && executeField.Kind() == reflect.String && executeField.String() != "" {
		executeTemplateName := executeField.String()
		if err := template.ExecuteTemplate(buf, executeTemplateName, data); err != nil {
			return nil, fmt.Errorf("failed to execute template %s: %w", dst, err)
		}
	} else {
		if err := template.Execute(buf, data); err != nil {
			return nil, fmt.Errorf("failed to execute template %s: %w", dst, err)
		}
	}

	// Return the generated content in a map with the destination as the key
	return map[string][]byte{
		dst: buf.Bytes(),
	}, nil
}

// Options struct to hold configuration parameters, including the file system
type Options struct {
	Templates fs.FS
}

// ScriptGenerator struct
type ScriptGenerator struct {
	Manager
}

func (m *Manager) Generate(ctx context.Context, data Data, generators ...string) (map[string][]byte, error) {
	generatedFiles := make(map[string][]byte)

	for name, generator := range m.Generators {
		log.Println("Generating", name)

		// Check if the generator is in the provided list
		if !contains(generators, name) {
			continue
		}

		mapping, err := generator(data)
		if err != nil {
			return nil, fmt.Errorf("failed to get mapping: %w", err)
		}

		files, err := m.generateFiles(ctx, mapping)
		if err != nil {
			return nil, fmt.Errorf("failed to generate files: %w", err)
		}

		for k, v := range files {
			generatedFiles[k] = v
		}
	}

	return generatedFiles, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

func (d Data) Year() string {
	return fmt.Sprintf("%d", time.Now().Year())
}

func (d Data) Fullname() string {
	return d.Author
}
