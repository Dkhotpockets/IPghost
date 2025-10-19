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
*   The Windows client is implemented.
*   The Linux and macOS clients are not yet implemented.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         User's Machine                           │
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │              GoFakeIP Client (gofakeip-client)             │ │
│  │                                                             │ │
│  │  ┌──────────────────────────────────────────────────────┐  │ │
│  │  │         cmd/gofakeip-client/main.go                  │  │ │
│  │  │  - CLI argument parsing                              │  │ │
│  │  │  - Privilege verification                             │  │ │
│  │  │  - Platform detection                                │  │ │
│  │  └────────────────┬─────────────────────────────────────┘  │ │
│  │                   │                                         │ │
│  │                   ▼                                         │ │
│  │  ┌──────────────────────────────────────────────────────┐  │ │
│  │  │      pkg/netconfig (Cross-Platform Abstraction)      │  │ │
│  │  │                                                       │  │ │
│    │  interface NetworkConfigurator {                     │  │ │
│  │    Setup(proxyAddr) error                            │  │ │
│  │    Teardown() error                                  │  │ │
│  │    Status() (ConfigState, error)                     │  │ │
│  │    CheckPrivileges() error                           │  │ │
│  │  }                                                    │  │ │
│  │  └────────────────┬─────────────────────────────────────┘  │ │
│  │                   │                                         │ │
│  │       ┌───────────┼───────────┬───────────────────┐        │ │
│  │       ▼           ▼           ▼                   ▼        │ │
│  │  ┌────────┐  ┌────────┐  ┌─────────┐       ┌──────────┐  │ │
│  │  │ Linux  │  │Windows │  │  macOS  │       │  Input   │  │ │
│  │  │Impl    │  │ Impl   │  │  Impl   │       │Validator │  │ │
│  │  │        │  │        │  │         │       │          │  │ │
│  │  │iptables│  │WFP/    │  │  pfctl  │       │Sanitizer │  │ │
│  │  │wrapper │  │netsh   │  │ wrapper │       │          │  │ │
│  │  └────┬───┘  └───┬────┘  └────┬────┘       └──────────┘  │ │
│  └───────│──────────│──────────────│──────────────────────────┘ │
│          │          │              │                            │
│          ▼          ▼              ▼                            │
│     ┌────────────────────────────────────┐                     │
│     │      OS Network Stack (Kernel)     │                     │
│     │  Netfilter  │   WFP   │     pf     │                     │
│     └────────────────┬───────────────────┘                     │
│                      │                                         │
│            Traffic Redirection via NAT/REDIRECT                │
│                      │                                         │
└──────────────────────┼─────────────────────────────────────────┘
                       │
                       │ Proxied Traffic
                       │
                       ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Remote Proxy Server                           │
│                                                                   │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │            GoFakeIP Server (gofakeip-server)               │ │
│  │                                                             │ │
│  │  ┌──────────────────────────────────────────────────────┐  │ │
│  │  │         cmd/gofakeip-server/main.go                  │  │ │
│  │  │  - Server initialization                             │  │ │
│  │  │  - Configuration loading (fake IP, listen addr)      │  │ │
│  │  └────────────────┬─────────────────────────────────────┘  │ │
│  │                   │                                         │ │
│  │                   ▼                                         │ │
│  │  ┌──────────────────────────────────────────────────────┐  │ │
│  │  │            pkg/proxy (Proxy Core)                    │  │ │
│  │  │                                                       │  │ │
│  │  │  ┌─────────────────┐    ┌──────────────────────┐    │  │ │
│  │  │  │  SOCKS5 Handler │    │  HTTP Proxy Handler  │    │  │ │
│  │  │  │  - Auth (opt)   │    │  - CONNECT method    │    │  │ │
│  │  │  │  - CONNECT cmd  │    │  - GET/POST proxy    │    │  │ │
│  │  │  └────────┬────────┘    └──────────┬───────────┘    │  │ │
│  │  │           │                        │                 │  │ │
│  │  │           └────────────┬───────────┘                 │  │ │
│  │  │                        │                             │  │ │
│  │  │                        ▼                             │  │ │
│  │  │           ┌─────────────────────────┐               │  │ │
│  │  │           │  Connection Pool Mgr    │               │  │ │
│  │  │           │  - Concurrent handling  │               │  │ │
│  │  │           │  - Source IP masking    │               │  │ │
│  │  │           │  - Error handling       │               │  │ │
│  │  │           └──────────┬──────────────┘               │  │ │
│  │  └──────────────────────│──────────────────────────────┘  │ │
│  │                         │                                  │ │
│  │                         ▼                                  │ │
│  │            ┌─────────────────────────┐                     │ │
│  │            │   IP Masking Layer      │                     │ │
│  │            │   (Fake IP injection)   │                     │ │
│  │            └──────────┬──────────────┘                     │ │
│  └───────────────────────┼──────────────────────────────────┘ │
│                          │                                     │
└──────────────────────────┼─────────────────────────────────────┘
                           │
                           │ Outbound connection with FAKE_IP as source
                           │
                           ▼
                  Internet Destination
             (Sees FAKE_IP, not real client IP)
```

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

## Design Principles

*   **Modularity:** The project is divided into logical packages (`cmd`, `pkg`, `tests`, `specs`) to promote code organization and reusability.
*   **Abstraction:** Interfaces are used extensively to define contracts between components, allowing for flexible and testable implementations.
*   **Security by Design:** Security considerations are integrated into every phase of development, from design to implementation and testing.
*   **Test-Driven Development (TDD):** The project emphasizes a comprehensive testing strategy, including unit, integration, and security tests.
*   **Clear State Management:** Entities have well-defined lifecycles and state transitions to ensure predictable behavior.
*   **Concurrency Safety:** Mutable states are protected by synchronization mechanisms to prevent race conditions.

## Building and Running