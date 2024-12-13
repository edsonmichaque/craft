#!/usr/bin/env bash
# Dependencies task script
# Manages project dependencies

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

update_deps() {
    log_info "Updating Go dependencies"
    go get -u ./...
    go mod tidy
}

verify_deps() {
    log_info "Verifying dependencies"
    go mod verify
}

install_tools() {
    log_info "Installing development tools"
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/golang/protobuf/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
}

main() {
    log_info "Starting dependency management"
    
    # Verify required tools
    ensure_command "go"
    
    update_deps
    verify_deps
    
    if [[ "${1:-}" == "--with-tools" ]]; then
        install_tools
    fi
    
    log_info "Dependency management complete!"
}

main "$@"
