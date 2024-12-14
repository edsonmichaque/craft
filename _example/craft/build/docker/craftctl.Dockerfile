# Production image - using distroless for minimal attack surface
FROM gcr.io/distroless/static:nonroot

# Copy the pre-built binary from dist directory
COPY dist/craftctl /craftctl

# Copy config files
COPY config/ /etc/craft/

# Set environment variables
ENV CRAFT_CONFIG_FILE=/etc/craft/config.yml

# Use non-root user
USER nonroot:nonroot

# Expose default port
EXPOSE 8080

ENTRYPOINT ["/craftctl"]