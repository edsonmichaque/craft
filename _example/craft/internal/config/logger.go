package config

// LoggerConfig holds all logging-related configuration
type LoggerConfig struct {
	Level  string            ` + "`mapstructure:\"level\" yaml:\"level\"`" + `
	Format string            ` + "`mapstructure:\"format\" yaml:\"format\"`" + `
	Output string            ` + "`mapstructure:\"output\" yaml:\"output\"`" + `
	Fields map[string]string ` + "`mapstructure:\"fields\" yaml:\"fields\"`" + `
}