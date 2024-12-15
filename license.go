package craft

import (
	"fmt"
)

func GenerateLicense(data Data) (map[string]RenderOptions, error) {
	licenseTemplates := map[string]string{
		"mit":          "license/mit.tmpl",
		"apache-2.0":   "license/apache2.tmpl",
		"agpl-3.0":     "license/agpl3.tmpl",
		"bsd-3-clause": "license/bsd3.tmpl",
		"gpl-3.0":      "license/gpl3.tmpl",
		"mpl-2.0":      "license/mpl2.tmpl",
		"apache":       "license/apache2.tmpl",
		"agpl":         "license/agpl3.tmpl",
		"bsd":          "license/bsd3.tmpl",
		"gpl":          "license/gpl3.tmpl",
		"mpl":          "license/mpl2.tmpl",
	}

	templateFile, exists := licenseTemplates[data.License]
	if !exists {
		return nil, fmt.Errorf("license %s not found", data.License)
	}

	return map[string]RenderOptions{
		"LICENSE": renderOptions(data, templateFile),
	}, nil
}
