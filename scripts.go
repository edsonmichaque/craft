package craft

import (
	"fmt"
)

func GenerateScripts(data Data) (map[string]RenderOptions, error) {
	out := make(map[string]RenderOptions)

	// Generate Dockerfiles for each binary
	for _, binary := range data.Binaries {
		filename := "build/docker/Dockerfile"
		if len(data.Binaries) > 1 {
			filename = fmt.Sprintf("build/docker/%s.Dockerfile", binary)
		}

		out[filename] = renderOptions(DockerfileOptions{Binary: binary, Data: data}, "build/docker/dockerfile.tmpl")
	}

	additional := map[string]RenderOptions{
		"scripts/README.md": renderOptions(data, "scripts/readme.md.tmpl"),

		"scripts/lib/common.sh":  renderOptions(data, "scripts/lib/common.sh.tmpl"),
		"scripts/lib/logger.sh":  renderOptions(data, "scripts/lib/logger.sh.tmpl"),
		"scripts/lib/docker.sh":  renderOptions(data, "scripts/lib/docker.sh.tmpl"),
		"scripts/lib/git.sh":     renderOptions(data, "scripts/lib/git.sh.tmpl"),
		"scripts/lib/version.sh": renderOptions(data, "scripts/lib/version.sh.tmpl"),

		"scripts/tasks/build.sh":        renderOptions(data, "scripts/tasks/build.sh.tmpl"),
		"scripts/tasks/test.sh":         renderOptions(data, "scripts/tasks/test.sh.tmpl"),
		"scripts/tasks/lint.sh":         renderOptions(data, "scripts/tasks/lint.sh.tmpl"),
		"scripts/tasks/docker.sh":       renderOptions(data, "scripts/tasks/docker.sh.tmpl"),
		"scripts/tasks/release.sh":      renderOptions(data, "scripts/tasks/release.sh.tmpl"),
		"scripts/tasks/proto.sh":        renderOptions(data, "scripts/tasks/proto.sh.tmpl"),
		"scripts/tasks/dependencies.sh": renderOptions(data, "scripts/tasks/dependencies.sh.tmpl"),
		"scripts/tasks/package.sh":      renderOptions(data, "scripts/tasks/package.sh.tmpl"),
		"scripts/tasks/setup-dev.sh":    renderOptions(data, "scripts/tasks/setup-dev.sh.tmpl"),

		"scripts/tasks/health-check.sh": renderOptions(data, "scripts/tasks/health-check.sh.tmpl"),

		"scripts/build": renderOptions(data, "scripts/build.tmpl"),
		"scripts/test":  renderOptions(data, "scripts/test.tmpl"),
		"scripts/ci":    renderOptions(data, "scripts/ci.tmpl"),

		"scripts/dev": renderOptions(data, "scripts/dev.tmpl"),
		"Makefile":    renderOptions(data, "makefile.tmpl"),
		"Taskfile":    renderOptions(data, "taskfile.tmpl"),
		"dagger.cue":  renderOptions(data, "dagger.cue.tmpl"),

		"build/ansible/main.yml":                   renderOptions(data, "build/ansible/main.yml.tmpl"),
		"build/ansible/application/tasks/main.yml": renderOptions(data, "build/ansible/role.yml.tmpl"),

		"build/k8s/kustomization.yml":                  renderOptions(data, "build/k8s/kustomization.yml.tmpl"),
		"build/k8s/base/deployment.yml":                renderOptions(data, "build/k8s/deployment.yml.tmpl"),
		"build/k8s/base/namespace.yml":                 renderOptions(data, "build/k8s/namespace.yml.tmpl"),
		"build/k8s/base/service.yml":                   renderOptions(data, "build/k8s/service.yml.tmpl"),
		"build/k8s/base/ingress.yml":                   renderOptions(data, "build/k8s/ingress.yml.tmpl"),
		"build/k8s/base/configmap.yml":                 renderOptions(data, "build/k8s/configmap.yml.tmpl"),
		"build/k8s/base/secret.yml":                    renderOptions(data, "build/k8s/secret.yml.tmpl"),
		"build/k8s/overlays/dev/kustomization.yml":     renderOptions(data, "build/k8s/kustomization.yml.tmpl"),
		"build/k8s/overlays/staging/kustomization.yml": renderOptions(data, "build/k8s/kustomization.yml.tmpl"),
		"build/k8s/overlays/prod/kustomization.yml":    renderOptions(data, "build/k8s/kustomization.yml.tmpl"),

		"build/terraform/main.tf":      renderOptions(data, "build/terraform/main.tf.tmpl"),
		"build/terraform/variables.tf": renderOptions(data, "build/terraform/variables.tf.tmpl"),
		"build/terraform/outputs.tf":   renderOptions(data, "build/terraform/outputs.tf.tmpl"),

		"build/helm/chart.yml":                renderOptions(data, "build/helm/chart.yml.tmpl"),
		"build/helm/values.yml":               renderOptions(data, "build/helm/values.yml.tmpl"),
		"build/helm/templates/deployment.yml": renderOptions(data, "build/helm/deployment.yml.tmpl"),
		"build/helm/templates/service.yml":    renderOptions(data, "build/helm/service.yml.tmpl"),

		"build/swarm/docker-compose.yml": renderOptions(data, "build/swarm/docker-compose.yml.tmpl"),
		"build/swarm/README.md":          renderOptions(data, "build/swarm/readme.adoc.tmpl"),

		".github/workflows/ci.yml": renderOptions(data, "github/ci.yml.tmpl"),
		".gitlab-ci.yml":           renderOptions(data, "gitlab/ci.yml.tmpl"),
		".gitlab/ci/test.yml":      renderOptions(data, "gitlab/test.yml.tmpl"),
		".gitlab/ci/build.yml":     renderOptions(data, "gitlab/build.yml.tmpl"),
		".gitlab/ci/release.yml":   renderOptions(data, "gitlab/release.yml.tmpl"),
	}

	for k, v := range additional {
		out[k] = v
	}

	return out, nil
}
