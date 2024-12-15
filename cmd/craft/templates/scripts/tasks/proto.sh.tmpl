#!/usr/bin/env bash
# Proto task script
# Generates code from protobuf definitions

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

generate_protos() {
    local proto_dir="${PROJECT_ROOT}/proto"
    local out_dir="${PROJECT_ROOT}/proto"
    
    log_info "Generating protobuf code..."
    
    # Clean existing generated files (only .pb.go files)
    find "${out_dir}" -type f -name "*.pb.go" -delete
    
    # Find all proto files
    local proto_files
    proto_files=$(find "${proto_dir}" -name "*.proto")
    
    # Generate Go code
    protoc \
        -I "${proto_dir}" \
        -I "/path/to/grpc-gateway"  // Replace with actual path
        -I "/path/to/protoc-gen-validate"  // Replace with actual path
        --go_out="${out_dir}" \
        --go_opt=paths=source_relative \
        --go-grpc_out="${out_dir}" \
        --go-grpc_opt=paths=source_relative \
        --grpc-gateway_out="${out_dir}" \
        --grpc-gateway_opt=paths=source_relative \
        --validate_out="lang=go,paths=source_relative:${out_dir}" \
        ${proto_files}
    
    # Clean up backup files
    find "${out_dir}" -type f -name "*.go.bak" -delete
}

install_proto_tools() {
    log_info "Installing protobuf tools"

    # Install protoc plugins
    go install \
        google.golang.org/protobuf/cmd/protoc-gen-go@latest \
        google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest \
        github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest \
        github.com/envoyproxy/protoc-gen-validate@latest

    # Install buf if available
    if command -v brew >/dev/null 2>&1; then
        brew install buf
    elif command -v go >/dev/null 2>&1; then
        go install github.com/bufbuild/buf/cmd/buf@latest
    fi
}

verify_proto_tools() {
    local missing_tools=()

    # Check required tools
    local tools=(
        "protoc"
        "protoc-gen-go"
        "protoc-gen-go-grpc"
        "protoc-gen-validate"
        "protoc-gen-grpc-gateway"
        "protoc-gen-openapiv2"
    )

    for tool in "${tools[@]}"; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done

    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_warn "Missing required tools: ${missing_tools[*]}"
        log_info "Installing missing tools..."
        install_proto_tools
    fi
}

main() {
    log_info "Starting protobuf generation"
    
    verify_proto_tools
    generate_protos
    
    log_info "Protobuf generation complete!"
}

main "$@"