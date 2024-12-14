# craft

## Overview

This project was generated using a Go project generator.

## Project Structure

### Binaries
- craftd
- craftctl

### Features

## Requirements

- Go 1.22 or higher
- Docker (optional)
- Make

## Getting Started

### Installation

```bash
go get github.com/edsonmichaque/craft
```

### Building

Build all binaries:
```bash
make build
```

Run tests:
```bash
make test
```

Run linter:
```bash
make lint
```

### Docker

Build the Docker images:
```bash
make docker
```

### Development

The project structure follows the standard Go project layout:

- /cmd - Main applications
- /internal - Private application and library code
- /pkg - Library code that's ok to use by external applications
- /hack - Tools and scripts to help with development
- /scripts - Scripts for CI/CD and other automation

## License

This project is licensed under the APACHE2 License - see the LICENSE file for details.