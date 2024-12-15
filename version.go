package craft

func GenerateVersion(data Data) (map[string]RenderOptions, error) {
	return map[string]RenderOptions{
		"pkg/version/version.go": renderOptions(data, "pkg/version/version.go.tmpl"),
	}, nil
}
