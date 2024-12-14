#!/usr/bin/env bash
# Setup development environment
# Installs required tools and dependencies

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

setup_git_hooks() {
    log_info "Setting up Git hooks..."
    
    local hooks_dir="${PROJECT_ROOT}/.git/hooks"
    local ci_hooks_dir="${PROJECT_ROOT}/scripts/ci/hooks"
    
    # Create hooks directory if it doesn't exist
    mkdir -p "$hooks_dir"
    
    # Link all hooks
    for hook in "$ci_hooks_dir"/*; do
        if [[ -f "$hook" ]]; then
            local hook_name
            hook_name=$(basename "$hook")
            ln -sf "$hook" "${hooks_dir}/${hook_name}"
            chmod +x "${hooks_dir}/${hook_name}"
        fi
    done
}

setup_tools() {
    log_info "Installing development tools..."
    
    # Run dependencies task with tools
    "${PROJECT_ROOT}/scripts/ci/tasks/dependencies.sh" --with-tools
    
    # Install additional development tools
    go install \
        github.com/cosmtrek/air@latest \
        github.com/go-delve/delve/cmd/dlv@latest \
        github.com/swaggo/swag/cmd/swag@latest
}

setup_env() {
    log_info "Setting up environment..."
    
    local env_file="${PROJECT_ROOT}/.env"
    local example_env="${PROJECT_ROOT}/.env.example"
    local template_env="${PROJECT_ROOT}/.env.template"
    
    # Check for existing .env file
    if [[ -f "$env_file" ]]; then
        log_info "Existing .env file found"
        
        # Optional: Check for missing variables
        if [[ -f "$example_env" ]]; then
            local missing_vars
            missing_vars=$(grep -v '^#' "$example_env" | cut -d '=' -f1 | while read -r var; do
                grep -q "^${var}=" "$env_file" || echo "$var"
            done)
            
            if [[ -n "$missing_vars" ]]; then
                log_warn "Missing environment variables in .env file:"
                echo "$missing_vars" | sed 's/^/  - /'
            fi
        fi
        
        return 0
    fi
    
    # Try to create .env file from available templates
    if [[ -f "$example_env" ]]; then
        log_info "Creating .env file from .env.example"
        cp "$example_env" "$env_file"
    elif [[ -f "$template_env" ]]; then
        log_info "Creating .env file from .env.template"
        cp "$template_env" "$env_file"
    else
        log_warn "No environment template found (.env.example or .env.template)"
        log_info "Creating minimal .env file"
        cat > "$env_file" << EOF
# Environment Configuration
# Generated on $(date -u +"%Y-%m-%d %H:%M:%S UTC")

# Application
APP_ENV=development
APP_DEBUG=true
APP_PORT=8080

# Add your environment variables below
EOF
    fi
    
    # Set appropriate permissions
    chmod 600 "$env_file"
    
    # Validate environment file
    if ! grep -q "APP_ENV=" "$env_file"; then
        log_warn "Environment file may be missing critical variables"
    fi
    
    log_info "Environment file created at: $env_file"
    log_info "Please review and update the environment variables as needed"
}

main() {
    log_info "Setting up development environment..."
    
    setup_git_hooks
    setup_tools
    setup_env
    
    log_info "Development environment setup complete!"
}

main "$@"
