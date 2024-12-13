#!/usr/bin/env bash
# Development environment setup utility

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

setup_tools() {
    log_info "Installing development tools"
    
    # Install Go tools
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install github.com/golang/protobuf/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    go install github.com/cosmtrek/air@latest
    
    # Install additional tools based on OS
    case "$(uname)" in
        "Darwin")
            brew install protobuf
            ;;
        "Linux")
            sudo apt-get update
            sudo apt-get install -y protobuf-compiler
            ;;
    esac
}

setup_hooks() {
    log_info "Setting up Git hooks"
    
    local hooks_dir="${PROJECT_ROOT}/.git/hooks"
    local ci_hooks_dir="${PROJECT_ROOT}/scripts/ci/hooks"
    
    # Link all hooks
    for hook in "$ci_hooks_dir"/*; do
        if [[ -f "$hook" ]]; then
            ln -sf "$hook" "${hooks_dir}/$(basename "$hook")"
        fi
    done
}

main() {
    log_info "Setting up development environment"
    
    setup_tools
    setup_hooks
    
    # Initialize environment
    cp -n "${PROJECT_ROOT}/.env.example" "${PROJECT_ROOT}/.env" 2>/dev/null || true
    
    log_info "Development environment setup complete!"
}

main "$@"
