# Gemini Code Assistant Context

This document provides context for the Gemini Code Assistant to understand the GoFakeIP project.

## Project Overview

GoFakeIP is a cross-platform SOCKS5/HTTP proxy server and client configuration utility written in Go. Its primary purpose is to mask a user's IP address by routing their traffic through a remote proxy server. The project emphasizes security, idempotency, and reversibility.

**Key Features:**

*   **Proxy Server:** A high-performance SOCKS5 and HTTP proxy server.
*   **Client Utility:** A cross-platform tool for automatically redirecting traffic to the proxy server.
*   **IP Masking:** Hides the user's real IP address by using the proxy server's IP as the source for outbound connections.
*   **Authentication:** Supports optional username/password authentication with bcrypt hashing.
*   **Cross-Platform:** The client utility is designed to work on Linux (iptables), Windows (netsh), and macOS (pfctl).

**Current Status:**

*   The proxy server is fully functional.
*   The Windows client is implemented but has a build issue.
*   The Linux and macOS clients are not yet implemented.

## Building and Running

### Prerequisites

*   Go 1.21 or higher
*   Docker (for containerized deployment)
*   Administrative/root privileges for the client utility to configure network rules.

### Build Instructions

```bash
# Clone the repository
git clone https://github.com/yourorg/gofakeip.git
cd gofakeip

# Download dependencies
go mod download

# Build the server
go build -o bin/gofakeip-server ./cmd/gofakeip-server

# Build the client
go build -o bin/gofakeip-client ./cmd/gofakeip-client
```

### Running the Server

The server can be run directly from the command line or using the provided configuration files in the `examples` directory.

```bash
# Run the server with command-line flags
./bin/gofakeip-server -listen 0.0.0.0:8080 -bind-ip <YOUR_IP> -protocols http

# Run the server with a configuration file
./bin/gofakeip-server -config examples/server-config-simple.yaml
```

### Running with Docker

The project is configured to run in a Docker container.

```bash
# Build the Docker image
docker build -t gofakeip:latest .

# Run with docker-compose
docker-compose up -d
```

### Testing

```bash
# Run all tests
go test ./...
```

## Development Conventions

*   **Security First:** The project adheres to strict security constraints, including least privilege enforcement, input validation, and no hardcoded secrets.
*   **Idempotency and Reversibility:** The client utility's setup and teardown operations are designed to be idempotent and fully reversible.
*   **Cross-Platform Compatibility:** The code is written to be compatible with Linux, Windows, and macOS. Platform-specific implementations are separated using build tags.
*   **Interfaces:** The project makes extensive use of interfaces to decouple components and facilitate testing.
*   **Configuration:** Configuration can be provided via command-line flags or a YAML file.
*   **Logging:** The server uses a structured logger for clear and informative output.
*   **Error Handling:** The project uses custom error types defined in `pkg/common/errors.go`.