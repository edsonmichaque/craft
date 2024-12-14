package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

const devDockerfileTemplate = `# Development image with live reload
FROM golang:{{.GoVersion}}-alpine

# Install development tools and build dependencies
RUN apk add --no-cache git make curl \
    && go install github.com/cosmtrek/air@latest

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Set environment variables
ENV {{.EnvPrefix}}_CONFIG_FILE=/app/config/config.yml \
    CGO_ENABLED=0 \
    GO111MODULE=on

# Expose default port
EXPOSE 8080

# Use air for live reload
ENTRYPOINT ["air", "-c", ".air.toml"]`

const prodDockerfileTemplate = `# Production image - using distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot

# Copy the pre-built binary from dist directory
COPY dist/{{.Binary}} /{{.Binary}}

# Copy config files
COPY config/ /etc/{{.ProjectName}}/

# Set environment variables
ENV {{.EnvPrefix}}_CONFIG_FILE=/etc/{{.ProjectName}}/config.yml

# Use non-root user
USER nonroot:nonroot

# Expose default port
EXPOSE 8080

ENTRYPOINT ["/{{.Binary}}"]`

const dockerComposeTemplate = `version: '3.8'

services:
{{- range .Binaries }}
  {{.}}:
    build:
      context: .
      dockerfile: docker/{{.}}.Dockerfile
    volumes:
      - .:/app
      - go-mod-cache:/go/pkg/mod
    env_file:
      - .env
    environment:
      - {{$.EnvPrefix}}_CONFIG_FILE=/app/config/{{$.ConfigFile}}
    ports:
      - "${PORT:-8080}:8080"
    depends_on:
      - postgres
    networks:
      - {{$.ProjectName}}-network

{{- end}}
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: {{.ProjectName}}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - {{.ProjectName}}-network

volumes:
  postgres_data:
  go-mod-cache:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const mysqlComposeTemplate = `version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: {{.ProjectName}}
      MYSQL_USER: {{.ProjectName}}
      MYSQL_PASSWORD: {{.ProjectName}}
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./init:/docker-entrypoint-initdb.d
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mysql_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const postgresComposeTemplate = `version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: {{.ProjectName}}
      POSTGRES_PASSWORD: {{.ProjectName}}
      POSTGRES_DB: {{.ProjectName}}
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init:/docker-entrypoint-initdb.d
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U {{.ProjectName}}"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  postgres_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const mariadbComposeTemplate = `version: '3.8'

services:
  mariadb:
    image: mariadb:10.11
    environment:
      MARIADB_ROOT_PASSWORD: root
      MARIADB_DATABASE: {{.ProjectName}}
      MARIADB_USER: {{.ProjectName}}
      MARIADB_PASSWORD: {{.ProjectName}}
    ports:
      - "3306:3306"
    volumes:
      - mariadb_data:/var/lib/mysql
      - ./init:/docker-entrypoint-initdb.d
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  mariadb_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const pgbouncerComposeTemplate = `version: '3.8'

services:
  pgbouncer:
    image: edoburu/pgbouncer:1.18
    environment:
      DB_USER: {{.ProjectName}}
      DB_PASSWORD: {{.ProjectName}}
      DB_HOST: postgres
      DB_NAME: {{.ProjectName}}
      POOL_MODE: transaction
      MAX_CLIENT_CONN: "1000"
      DEFAULT_POOL_SIZE: "100"
      ADMIN_USERS: "postgres,{{.ProjectName}}"
    ports:
      - "6432:6432"
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -h localhost -p 6432 -U {{.ProjectName}}"]
      interval: 10s
      timeout: 5s
      retries: 5

networks:
  {{.ProjectName}}-network:
    external: true`

const proxysqlComposeTemplate = `version: '3.8'

services:
  proxysql:
    image: proxysql/proxysql:2.5
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_USER: {{.ProjectName}}
      MYSQL_PASSWORD: {{.ProjectName}}
    volumes:
      - ./proxysql.cnf:/etc/proxysql.cnf
      - proxysql_data:/var/lib/proxysql
    ports:
      - "6033:6033" # MySQL client port
      - "6032:6032" # Admin port
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD", "mysql", "-h", "127.0.0.1", "-P", "6032", "-u", "admin", "-padmin", "--execute", "SELECT 1"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  proxysql_data:

networks:
  {{.ProjectName}}-network:
    external: true`

const proxysqlConfigTemplate = `datadir="/var/lib/proxysql"

admin_variables =
{
    admin_credentials="admin:admin"
    mysql_ifaces="0.0.0.0:6032"
}

mysql_variables =
{
    threads=4
    max_connections=2048
    default_query_delay=0
    default_query_timeout=36000000
    have_compress=true
    poll_timeout=2000
    interfaces="0.0.0.0:6033"
    default_schema="{{.ProjectName}}"
    stacksize=1048576
    server_version="8.0.27"
    connect_timeout_server=3000
    monitor_username="{{.ProjectName}}"
    monitor_password="{{.ProjectName}}"
    monitor_history=600000
    monitor_connect_interval=60000
    monitor_ping_interval=10000
    monitor_read_only_interval=1500
    monitor_read_only_timeout=500
    ping_interval_server_msec=120000
    ping_timeout_server=500
    commands_stats=true
    sessions_sort=true
    connect_retries_on_failure=10
}

mysql_servers =
(
    {
        address="mysql"
        port=3306
        hostgroup=0
        max_connections=200
        weight=1
    },
    {
        address="mariadb"
        port=3306
        hostgroup=1
        max_connections=200
        weight=1
    }
)

mysql_users =
(
    {
        username = "{{.ProjectName}}"
        password = "{{.ProjectName}}"
        default_hostgroup = 0
        max_connections = 1000
        default_schema = "{{.ProjectName}}"
        active = 1
    }
)

mysql_query_rules =
(
    {
        rule_id=1
        active=1
        match_pattern="^SELECT .* FOR UPDATE$"
        destination_hostgroup=0
        apply=1
    },
    {
        rule_id=2
        active=1
        match_pattern="^SELECT"
        destination_hostgroup=1
        apply=1
    }
)`

const odysseyComposeTemplate = `version: '3.8'

services:
  odyssey:
    image: yandex/odyssey:latest
    volumes:
      - ./odyssey.conf:/etc/odyssey/odyssey.conf
    ports:
      - "6432:6432"
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -h localhost -p 6432"]
      interval: 10s
      timeout: 5s
      retries: 5

networks:
  {{.ProjectName}}-network:
    external: true`

const odysseyConfigTemplate = `
storage "postgres_server" {
    type "remote"
    host "postgres"
    port 5432
}

database "{{.ProjectName}}" {
    user "{{.ProjectName}}" {
        authentication "clear_text"
        password "{{.ProjectName}}"
        storage "postgres_server"
        pool "session"
        pool_size 100
        pool_timeout 4000
        pool_ttl 3600
        pool_discard no
        pool_cancel yes
        pool_rollback yes
    }
}`

const poolerComposeTemplate = `version: '3.8'

services:
  pgpool:
    image: bitnami/pgpool:latest
    environment:
      PGPOOL_BACKEND_NODES: 0:postgres:5432
      PGPOOL_SR_CHECK_USER: {{.ProjectName}}
      PGPOOL_SR_CHECK_PASSWORD: {{.ProjectName}}
      PGPOOL_ENABLE_LOAD_BALANCING: yes
      PGPOOL_MAX_POOL: 4
      PGPOOL_POSTGRES_USERNAME: {{.ProjectName}}
      PGPOOL_POSTGRES_PASSWORD: {{.ProjectName}}
    ports:
      - "5433:5432"
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD", "/opt/bitnami/scripts/pgpool/healthcheck.sh"]
      interval: 10s
      timeout: 5s
      retries: 5

networks:
  {{.ProjectName}}-network:
    external: true`

const mysqlRouterComposeTemplate = `version: '3.8'

services:
  mysqlrouter:
    image: mysql/mysql-router:8.0
    environment:
      MYSQL_HOST: mysql
      MYSQL_PORT: 3306
      MYSQL_USER: {{.ProjectName}}
      MYSQL_PASSWORD: {{.ProjectName}}
      MYSQL_ROOT_PASSWORD: root
    ports:
      - "6446:6446" # Read-Write port
      - "6447:6447" # Read-Only port
    depends_on:
      mysql:
        condition: service_healthy
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD", "mysqlrouter", "--status"]
      interval: 10s
      timeout: 5s
      retries: 5

networks:
  {{.ProjectName}}-network:
    external: true`

const haproxyComposeTemplate = `version: '3.8'

services:
  haproxy:
    image: haproxy:2.8
    volumes:
      - ./haproxy.cfg:/usr/local/etc/haproxy/haproxy.cfg:ro
    ports:
      - "3307:3307" # MySQL proxy port
      - "8404:8404" # Stats page
    depends_on:
      - mysql
      - mariadb
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD", "haproxy", "-c", "-f", "/usr/local/etc/haproxy/haproxy.cfg"]
      interval: 10s
      timeout: 5s
      retries: 5

networks:
  {{.ProjectName}}-network:
    external: true`

const haproxyConfigTemplate = `global
    maxconn 4096
    stats socket /var/run/haproxy.sock mode 600 level admin

defaults
    mode tcp
    timeout connect 5s
    timeout client 50s
    timeout server 50s

frontend mysql_front
    bind *:3307
    mode tcp
    default_backend mysql_back

backend mysql_back
    mode tcp
    balance roundrobin
    option mysql-check user {{.ProjectName}}
    server mysql mysql:3306 check
    server mariadb mariadb:3306 check backup

listen stats
    bind *:8404
    mode http
    stats enable
    stats uri /
    stats refresh 10s`

const redisComposeTemplate = `version: '3.8'

services:
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  redis_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const rabbitmqComposeTemplate = `version: '3.8'

services:
  rabbitmq:
    image: rabbitmq:3-management
    environment:
      RABBITMQ_DEFAULT_USER: user
      RABBITMQ_DEFAULT_PASS: password
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    networks:
      - {{.ProjectName}}-network
    healthcheck:
      test: ["CMD", "rabbitmqctl", "status"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  rabbitmq_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const kafkaComposeTemplate = `version: '3.8'

services:
  zookeeper:
    image: wurstmeister/zookeeper:3.4.6
    ports:
      - "2181:2181"
    networks:
      - {{.ProjectName}}-network

  kafka:
    image: wurstmeister/kafka:2.13-2.7.0
    ports:
      - "9092:9092"
    environment:
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    networks:
      - {{.ProjectName}}-network

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const minioComposeTemplate = `version: '3.8'

services:
  minio:
    image: minio/minio
    command: server /data
    ports:
      - "9000:9000"
    environment:
      MINIO_ACCESS_KEY: minioadmin
      MINIO_SECRET_KEY: minioadmin
    volumes:
      - minio_data:/data
    networks:
      - {{.ProjectName}}-network

volumes:
  minio_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const localstackComposeTemplate = `version: '3.8'

services:
  localstack:
    image: localstack/localstack
    ports:
      - "4566:4566"  # LocalStack Gateway
      - "4571:4571"  # S3
    environment:
      - SERVICES=s3,lambda,dynamodb
      - DEBUG=1
      - DATA_DIR=/tmp/localstack/data
    volumes:
      - localstack_data:/tmp/localstack
    networks:
      - {{.ProjectName}}-network

volumes:
  localstack_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const valkeyComposeTemplate = `version: '3.8'

services:
  valkey:
    image: valkeyio/valkey:latest
    ports:
      - "6379:6379"
    volumes:
      - valkey_data:/data
    networks:
      - {{.ProjectName}}-network
    environment:
      VALKEY_CONFIG: /etc/valkey/valkey.conf
    command: ["valkey-server", "/etc/valkey/valkey.conf"]

  valkey-sentinel:
    image: valkeyio/valkey-sentinel:latest
    ports:
      - "26379:26379"
    networks:
      - {{.ProjectName}}-network
    environment:
      VALKEY_SENTINEL_CONFIG: /etc/valkey/sentinel.conf
    command: ["valkey-sentinel", "/etc/valkey/sentinel.conf"]

volumes:
  valkey_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const redisSentinelComposeTemplate = `version: '3.8'

services:
  redis:
    image: redis:alpine
    command: redis-server --appendonly yes
    ports:
      - "6379:6379"
    networks:
      - {{.ProjectName}}-network

  sentinel:
    image: redis:alpine
    command: redis-sentinel /etc/sentinel.conf
    ports:
      - "26379:26379"
    networks:
      - {{.ProjectName}}-network
    volumes:
      - ./sentinel.conf:/etc/sentinel.conf

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const prometheusGrafanaComposeTemplate = `version: '3.8'

services:
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    networks:
      - {{.ProjectName}}-network

  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
    networks:
      - {{.ProjectName}}-network
    volumes:
      - grafana_data:/var/lib/grafana

volumes:
  grafana_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const statsdComposeTemplate = `version: '3.8'

services:
  statsd:
    image: statsd/statsd
    ports:
      - "8125:8125/udp"
    networks:
      - {{.ProjectName}}-network

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const influxdbComposeTemplate = `version: '3.8'

services:
  influxdb:
    image: influxdb:latest
    ports:
      - "8086:8086"
    volumes:
      - influxdb_data:/var/lib/influxdb
    networks:
      - {{.ProjectName}}-network

volumes:
  influxdb_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

const pgadminComposeTemplate = `version: '3.8'

services:
  pgadmin:
    image: dpage/pgadmin4
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@example.com
      PGADMIN_DEFAULT_PASSWORD: admin
    ports:
      - "5050:80"
    networks:
      - {{.ProjectName}}-network
    volumes:
      - pgadmin_data:/var/lib/pgadmin

volumes:
  pgadmin_data:

networks:
  {{.ProjectName}}-network:
    driver: bridge`

func generateDockerfiles(projectPath string, cfg Config) error {
	// Create docker directory for dev files
	dockerDir := filepath.Join(projectPath, "docker")
	if err := os.MkdirAll(dockerDir, 0755); err != nil {
		return fmt.Errorf("failed to create docker directory: %w", err)
	}

	// Create build/docker directory for prod files
	buildDockerDir := filepath.Join(projectPath, "build", "docker")
	if err := os.MkdirAll(buildDockerDir, 0755); err != nil {
		return fmt.Errorf("failed to create build/docker directory: %w", err)
	}

	// Generate development Dockerfiles in docker/
	if err := generateDevDockerfiles(dockerDir, cfg); err != nil {
		return fmt.Errorf("failed to generate dev dockerfiles: %w", err)
	}

	// Generate production Dockerfiles in build/docker/
	if err := generateProdDockerfiles(buildDockerDir, cfg); err != nil {
		return fmt.Errorf("failed to generate prod dockerfiles: %w", err)
	}

	// Define a map of compose templates including poolers
	composeTemplates := map[string]string{
		"docker-compose.yml":                    dockerComposeTemplate,
		"mysql/docker-compose.yml":              mysqlComposeTemplate,
		"postgres/docker-compose.yml":           postgresComposeTemplate,
		"pgbouncer/docker-compose.yml":          pgbouncerComposeTemplate,
		"proxysql/docker-compose.yml":           proxysqlComposeTemplate,
		"proxysql/proxysql.cnf":                 proxysqlConfigTemplate,
		"odyssey/docker-compose.yml":            odysseyComposeTemplate,
		"odyssey/odyssey.conf":                  odysseyConfigTemplate,
		"pooler/docker-compose.yml":             poolerComposeTemplate,
		"mysqlrouter/docker-compose.yml":        mysqlRouterComposeTemplate,
		"haproxy/docker-compose.yml":            haproxyComposeTemplate,
		"haproxy/haproxy.cfg":                   haproxyConfigTemplate,
		"mariadb/docker-compose.yml":            mariadbComposeTemplate,
		"redis/docker-compose.yml":              redisComposeTemplate,
		"rabbitmq/docker-compose.yml":           rabbitmqComposeTemplate,
		"kafka/docker-compose.yml":              kafkaComposeTemplate,
		"minio/docker-compose.yml":              minioComposeTemplate,
		"localstack/docker-compose.yml":         localstackComposeTemplate,
		"valkey/docker-compose.yml":             valkeyComposeTemplate,
		"redis-sentinel/docker-compose.yml":     redisSentinelComposeTemplate,
		"prometheus-grafana/docker-compose.yml": prometheusGrafanaComposeTemplate,
		"statsd/docker-compose.yml":             statsdComposeTemplate,
		"influxdb/docker-compose.yml":           influxdbComposeTemplate,
		"pgadmin/docker-compose.yml":            pgadminComposeTemplate,
	}

	// Iterate over the map to generate each compose file
	for fileName, templateContent := range composeTemplates {
		composeFile := filepath.Join(dockerDir, fileName)
		data := struct {
			Config
			ConfigFile string
		}{
			Config:     cfg,
			ConfigFile: "config.yml",
		}

		if err := generateFileFromTemplate(composeFile, templateContent, data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", fileName, err)
		}
	}

	return nil
}

func generateDevDockerfiles(dockerDir string, cfg Config) error {
	tmpl, err := template.New("dockerfile").Parse(devDockerfileTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse dev dockerfile template: %w", err)
	}

	for _, binary := range cfg.Binaries {
		fileName := filepath.Join(dockerDir, binary+".Dockerfile")
		if err := generateSingleDockerfile(fileName, tmpl, cfg, binary); err != nil {
			return fmt.Errorf("failed to generate dev dockerfile for %s: %w", binary, err)
		}
	}

	return nil
}

func generateProdDockerfiles(buildDockerDir string, cfg Config) error {
	tmpl, err := template.New("dockerfile").Parse(prodDockerfileTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse prod dockerfile template: %w", err)
	}

	for _, binary := range cfg.Binaries {
		fileName := filepath.Join(buildDockerDir, binary+".Dockerfile")
		if err := generateSingleDockerfile(fileName, tmpl, cfg, binary); err != nil {
			return fmt.Errorf("failed to generate prod dockerfile for %s: %w", binary, err)
		}
	}

	return nil
}

// Helper function to generate a single Dockerfile
func generateSingleDockerfile(fileName string, tmpl *template.Template, cfg Config, binary string) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", fileName, err)
	}
	defer f.Close()

	data := struct {
		Binary       string
		GoVersion    string
		ModulePrefix string
		ProjectName  string
		EnvPrefix    string
	}{
		Binary:       binary,
		GoVersion:    cfg.GoVersion,
		ModulePrefix: cfg.ModulePrefix,
		ProjectName:  cfg.ProjectName,
		EnvPrefix:    cfg.EnvPrefix,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}
