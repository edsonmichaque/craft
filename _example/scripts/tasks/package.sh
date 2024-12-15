#!/usr/bin/env bash
# Package task script
# Builds native packages (DEB, RPM, APK)

source "$(dirname "${BASH_SOURCE[0]}")/../lib/common.sh"

# Package metadata
PACKAGE_NAME="craft"
PACKAGE_VERSION="${VERSION:-0.0.1}"
PACKAGE_RELEASE="1"
PACKAGE_ARCH="amd64"
PACKAGE_DESCRIPTION="craft - craft"
PACKAGE_MAINTAINER="Edson Michaque"
PACKAGE_LICENSE="mit"

# Directory structure
BUILD_DIR="${PROJECT_ROOT}/build"
PACKAGE_ROOT="${PROJECT_ROOT}/packaging"
STAGING_DIR="${PACKAGE_ROOT}/staging"
OUTPUT_DIR="${PROJECT_ROOT}/dist/packages"

create_package_dirs() {
    local dirs=(
        "${STAGING_DIR}/DEBIAN"
        "${STAGING_DIR}/usr/local/bin"
        "${STAGING_DIR}/etc/craft"
        "${STAGING_DIR}/usr/lib/systemd/system"
        "${STAGING_DIR}/etc/init.d"
        "${STAGING_DIR}/etc/conf.d"
        "${OUTPUT_DIR}"
    )

    for dir in "${dirs[@]}"; do
        mkdir -p "$dir"
    done
}

install_service_files() {
    local binary=$1
    local init_system=$2

    case "$init_system" in
        systemd)
            cp "${BUILD_DIR}/init/systemd/${binary}.service" \
                "${STAGING_DIR}/usr/lib/systemd/system/"
            ;;
        openrc)
            cp "${BUILD_DIR}/init/openrc/${binary}-openrc" \
                "${STAGING_DIR}/etc/init.d/${binary}"
            cp "${BUILD_DIR}/config/${binary}.conf" \
                "${STAGING_DIR}/etc/conf.d/${binary}"
            ;;
        sysvinit)
            cp "${BUILD_DIR}/init/sysvinit/${binary}-sysvinit" \
                "${STAGING_DIR}/etc/init.d/${binary}"
            ;;
    esac
}

install_config_files() {
    local binary=$1
    
    # Copy configuration files
    if [[ -d "${BUILD_DIR}/config" ]]; then
        cp -r "${BUILD_DIR}/config/"* "${STAGING_DIR}/etc/craft/"
    fi
}

build_tarball() {
    local binary=$1
    local format=${2:-}  # No default, empty if not provided
    log_info "Building tarball for ${binary}"

    local archive_name="${PACKAGE_NAME}-${PACKAGE_VERSION}-${PACKAGE_ARCH}"
    local temp_dir="${STAGING_DIR}/archive/${PACKAGE_NAME}-${PACKAGE_VERSION}"

    # Create directory structure
    mkdir -p "${temp_dir}"/{bin,etc,init/{systemd,openrc,sysvinit}}

    # Copy files
    cp "${PROJECT_ROOT}/bin/${binary}" "${temp_dir}/bin/"
    cp -r "${BUILD_DIR}/config/"* "${temp_dir}/etc/" 2>/dev/null || true
    cp -r "${BUILD_DIR}/init/"* "${temp_dir}/init/" 2>/dev/null || true
    cp "${PROJECT_ROOT}/README.md" "${temp_dir}/" 2>/dev/null || true
    cp "${PROJECT_ROOT}/LICENSE" "${temp_dir}/" 2>/dev/null || true

    # If no format specified, create tar.gz
    if [[ -z "$format" ]]; then
        tar -czf "${OUTPUT_DIR}/${archive_name}.tar.gz" \
            -C "${STAGING_DIR}/archive" \
            "${PACKAGE_NAME}-${PACKAGE_VERSION}"
        return
    fi

    # Handle specific format if provided
    case "$format" in
        gz|gzip)
            tar -czf "${OUTPUT_DIR}/${archive_name}.tar.gz" \
                -C "${STAGING_DIR}/archive" \
                "${PACKAGE_NAME}-${PACKAGE_VERSION}"
            ;;
        bz2|bzip2)
            tar -cjf "${OUTPUT_DIR}/${archive_name}.tar.bz2" \
                -C "${STAGING_DIR}/archive" \
                "${PACKAGE_NAME}-${PACKAGE_VERSION}"
            ;;
        xz)
            tar -cJf "${OUTPUT_DIR}/${archive_name}.tar.xz" \
                -C "${STAGING_DIR}/archive" \
                "${PACKAGE_NAME}-${PACKAGE_VERSION}"
            ;;
        zst|zstd)
            tar -cf - -C "${STAGING_DIR}/archive" "${PACKAGE_NAME}-${PACKAGE_VERSION}" | \
                zstd -T0 > "${OUTPUT_DIR}/${archive_name}.tar.zst"
            ;;
        zip)
            (cd "${STAGING_DIR}/archive" && \
                zip -r "${OUTPUT_DIR}/${archive_name}.zip" "${PACKAGE_NAME}-${PACKAGE_VERSION}")
            ;;
        *)
            log_error "Unsupported compression format: $format"
            return 1
            ;;
    esac
}

main() {
    log_info "Starting package generation"

    # Clean and create directories
    rm -rf "${STAGING_DIR}" "${OUTPUT_DIR}"
    create_package_dirs

    # Verify required tools
    if command -v dpkg-deb >/dev/null; then
        build_deb "craft"
    else
        log_warn "dpkg-deb not found, skipping DEB package"
    fi

    if command -v rpmbuild >/dev/null; then
        build_rpm "craft"
    else
        log_warn "rpmbuild not found, skipping RPM package"
    fi

    if command -v abuild >/dev/null; then
        build_apk "craft"
    else
        log_warn "abuild not found, skipping APK package"
    fi

    # Always build tarball as fallback
    build_tarball "craft"

    log_info "Package generation complete! Packages available in ${OUTPUT_DIR}"
}

main "$@"