package craft

func GenerateCommonFiles(data Data) (map[string]RenderOptions, error) {
	return map[string]RenderOptions{
		".gitignore":   renderOptions(data, "common/gitignore.tmpl"),
		".env.example": renderOptions(data, "common/env.tmpl"),
		"go.mod":       renderOptions(data, "common/go.mod.tmpl"),
		"README.md":    renderOptions(data, "common/readme.md.tmpl"),
		".air.toml":    renderOptions(data, "common/air.toml.tmpl"),
	}, nil
}

func renderOptions(data interface{}, templates ...string) RenderOptions {
	return RenderOptions{Templates: templates, Data: data}
}
