# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the entire source code
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o gofakeip-server \
    ./cmd/gofakeip-server

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS proxying
RUN apk --no-cache add ca-certificates

# Create non-root user for security
RUN addgroup -g 1000 gofakeip && \
    adduser -D -u 1000 -G gofakeip gofakeip

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/gofakeip-server .


# Change ownership to non-root user
RUN chown -R gofakeip:gofakeip /app

# Switch to non-root user
USER gofakeip

# Expose the proxy port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8888/ || exit 1

# Default environment variables
ENV LISTEN_ADDRESS="0.0.0.0:8080" \
    BIND_IP="" \
    PROTOCOLS="http,socks5" \
    LOG_LEVEL="info" \
    MAX_CONNECTIONS="1000"

# Entry point
ENTRYPOINT ["/app/gofakeip-server"]

# Default arguments (can be overridden)
CMD ["-config", "/app/config.yaml", "-listen", "0.0.0.0:8888", "-log-level", "info"]
