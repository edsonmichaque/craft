# Development image with live reload
FROM golang:1.22-alpine

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
ENV CRAFT_CONFIG_FILE=/app/config/config.yml \
    CGO_ENABLED=0 \
    GO111MODULE=on

# Expose default port
EXPOSE 8080

# Use air for live reload
ENTRYPOINT ["air", "-c", ".air.toml"]