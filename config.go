package craft

func GenerateConfig(data Data) (map[string]RenderOptions, error) {
	return map[string]RenderOptions{
		"internal/config/README.md":      renderOptions(data, "internal/config/readme.md.tmpl"),
		"internal/config/config.yml":     renderOptions(data, "internal/config/config.yml.tmpl"),
		"internal/config/.env.example":   renderOptions(data, "internal/config/env.tmpl"),
		"internal/config/logger.go":      renderOptions(data, "internal/config/logger.go.tmpl"),
		"internal/config/config_test.go": renderOptions(data, "internal/config/config_test.go.tmpl"),
		"internal/config/config.go":      renderOptions(data, "internal/config/config.go.tmpl"),
	}, nil
}
