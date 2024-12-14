#!/usr/bin/env bash
# Dependencies task script
# Manages project dependencies and tools

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

update_dependencies() {
    log_info "Updating Go dependencies"
    go get -u ./...
    go mod tidy
}

verify_dependencies() {
    log_info "Verifying dependencies"
    go mod verify
}

install_tools() {
    log_info "Installing development tools..."
    
    # Run dependencies task with tools
    "${PROJECT_ROOT}/scripts/ci/tasks/dependencies.sh" --with-tools
    
    # Install additional development tools
    go install \
        github.com/cosmtrek/air@latest \
        github.com/go-delve/delve/cmd/dlv@latest \
        github.com/swaggo/swag/cmd/swag@latest
}

main() {
    log_info "Starting dependencies management"
    
    # Verify required tools
    ensure_command "go"
    
    update_dependencies
    verify_dependencies
    
    if [[ "${1:-}" == "--with-tools" ]]; then
        install_tools
    fi
    
    log_info "Dependencies management complete!"
}

main "$@"
