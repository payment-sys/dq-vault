# Stage 1 (to create a "build" image)
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git=2.46.0-r0 gcc=13.2.1_git20230522-r1 musl-dev=1.2.4-r1

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o dq-vault .

# Stage 2 (to create a vault container with executable)
FROM hashicorp/vault:1.15.6

# Make new directory for plugins
RUN mkdir -p /vault/plugins

# Install required packages
# hadolint ignore=DL3018,DL3008
RUN apk --no-cache add ca-certificates=20241121-r1 wget=1.21.4-r0 perl-utils=5.36.2-r1 && \
    wget -q -O /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub && \
    wget --progress=dot:giga https://github.com/sgerrand/alpine-pkg-glibc/releases/download/2.34-r0/glibc-2.34-r0.apk && \
    # hadolint ignore=DL3018,DL3008
    apk add --no-cache --force-overwrite glibc-2.34-r0.apk && \
    rm -rf /var/cache/apk/* glibc-2.34-r0.apk

# Copy executable from builder to vault
COPY --from=builder /app/dq-vault /vault/plugins/dq-vault

# Set proper permissions
RUN chown vault:vault /vault/plugins/dq-vault && \
    chmod +x /vault/plugins/dq-vault

# Create necessary directories
RUN mkdir -p /var/lib/vault/data && \
    chown -R vault:vault /var/lib/vault

# Switch to vault user
USER vault

# Expose port
EXPOSE 8200

# Default command
CMD ["vault", "server", "-config=/etc/vault.d/vault.hcl"]
