package craft

import (
	"fmt"
	"log"
	"strings"

	"github.com/iancoleman/strcase"
)

func GenerateCommands(data Data) (map[string]RenderOptions, error) {
	cliFrameworks := []string{"cobra", "urfave"}

	if !contains(cliFrameworks, data.Framework) {
		return nil, fmt.Errorf("invalid cli framework: %s", data.Framework)
	}

	templates := map[string][]string{
		"root":    {"internal/commands/base.go.tmpl", fmt.Sprintf("internal/commands/%s_root.go.tmpl", data.Framework)},
		"version": {fmt.Sprintf("internal/commands/%s_root.go.tmpl", data.Framework), fmt.Sprintf("internal/commands/%s_version.go.tmpl", data.Framework)},
		"server":  {fmt.Sprintf("internal/commands/%s_root.go.tmpl", data.Framework), fmt.Sprintf("internal/commands/%s_server.go.tmpl", data.Framework)},
	}

	out := make(map[string]RenderOptions)

	if len(data.Binaries) == 1 {
		for key, tmpl := range templates {
			out[fmt.Sprintf("internal/commands/%s.go", key)] = renderOptions(data, tmpl...)
		}
		return out, nil
	}

	for _, binary := range data.Binaries {
		for key, tmpl := range templates {
			out[fmt.Sprintf("internal/commands/%s/%s.go", binary, key)] = renderOptions(CommandOptions{Data: data, Binary: binary}, tmpl...)
		}
	}

	log.Printf("Generated commands: %#+v", out)

	for k, v := range out {
		log.Printf("Generated command: %s %#+v", k, v.Templates)
	}

	for _, binary := range data.Binaries {
		out[fmt.Sprintf("cmd/%s/main.go", binary)] = RenderOptions{
			Templates: []string{"internal/commands/main.go.tmpl"},
			Data:      CommandOptions{Data: data, Binary: binary},
			Execute:   "main",
		}
	}

	return out, nil
}

type CommandOptions struct {
	Data
	Binary string
}

func (cmd CommandOptions) PackageName() string {
	return strings.ReplaceAll(strcase.ToKebab(cmd.Binary), "-", "")
}
