package craft

import "fmt"

func GenerateDockerFiles(data Data) (map[string]RenderOptions, error) {
	out := make(map[string]RenderOptions)

	// Generate Dockerfiles for each binary
	for _, binary := range data.Binaries {
		filename := "docker/Dockerfile"
		if len(data.Binaries) > 1 {
			filename = fmt.Sprintf("docker/%s.Dockerfile", binary)
		}
		out[filename] = renderOptions(DockerfileOptions{Binary: binary, Data: data}, "docker/dockerfile.tmpl")
	}

	// Additional Docker-related files
	additionalFiles := map[string]string{
		"docker/README.md":                     "docker/readme.md.tmpl",
		"docker/docker-compose.yml":            "docker/docker-compose.yml.tmpl",
		"docker/mysql/docker-compose.yml":      "docker/mysql/docker-compose.yml.tmpl",
		"docker/postgres/docker-compose.yml":   "docker/postgres/docker-compose.yml.tmpl",
		"docker/mariadb/docker-compose.yml":    "docker/mariadb/docker-compose.yml.tmpl",
		"docker/redis/docker-compose.yml":      "docker/redis/docker-compose.yml.tmpl",
		"docker/grafana/docker-compose.yml":    "docker/grafana/docker-compose.yml.tmpl",
		"docker/prometheus/docker-compose.yml": "docker/prometheus/docker-compose.yml.tmpl",
		"docker/rabbitmq/docker-compose.yml":   "docker/rabbitmq/docker-compose.yml.tmpl",
		"docker/jaeger/docker-compose.yml":     "docker/jaeger/docker-compose.yml.tmpl",
	}

	for k, tmpl := range additionalFiles {
		out[k] = renderOptions(data, tmpl)
	}

	return out, nil
}

type DockerfileOptions struct {
	Binary string
	Data
}
