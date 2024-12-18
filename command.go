package craft

import (
	"fmt"
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
		"version": {"internal/commands/base.go.tmpl", fmt.Sprintf("internal/commands/%s_root.go.tmpl", data.Framework), fmt.Sprintf("internal/commands/%s_version.go.tmpl", data.Framework)},
		"server":  {"internal/commands/base.go.tmpl", fmt.Sprintf("internal/commands/%s_root.go.tmpl", data.Framework), fmt.Sprintf("internal/commands/%s_server.go.tmpl", data.Framework)},
	}

	out := make(map[string]RenderOptions)

	if len(data.Binaries) == 1 {
		for key, tmpl := range templates {
			out[fmt.Sprintf("internal/commands/%s.go", key)] = RenderOptions{
				Templates: tmpl,
				Data: CommandOptions{
					Data:    data,
					Binary:  data.Binaries[0],
					Execute: "base",
				},
			}
		}
		return out, nil
	}

	for _, binary := range data.Binaries {
		for key, tmpl := range templates {
			out[fmt.Sprintf("internal/commands/%s/%s.go", binary, key)] = RenderOptions{
				Templates: tmpl,
				Data: CommandOptions{
					Data:    data,
					Binary:  binary,
					Execute: "base",
				},
			}
		}
	}

	for _, binary := range data.Binaries {
		out[fmt.Sprintf("cmd/%s/main.go", binary)] = RenderOptions{
			Templates: []string{"internal/commands/main.go.tmpl"},
			Data: CommandOptions{
				Data:    data,
				Binary:  binary,
				Execute: "main",
			},
		}
	}

	if data.Binaries != nil {
		for _, binary := range data.Binaries {
			out[fmt.Sprintf("internal/commands/%s/README.md", binary)] = RenderOptions{
				Templates: []string{fmt.Sprintf("internal/commands/readme_%s.md.tmpl", data.Framework)},
				Data:      data,
			}
		}
	} else {
		out["internal/commands/README.md"] = RenderOptions{
			Templates: []string{"internal/commands/readme_base.md.tmpl"},
			Data:      data,
		}
	}

	return out, nil
}

type CommandOptions struct {
	Data
	Binary  string
	Execute string
}

func (cmd CommandOptions) PackageName() string {
	return strings.ReplaceAll(strcase.ToKebab(cmd.Binary), "-", "")
}

type RenderOptionsWithExecute struct {
	RenderOptions
	Execute string
}
