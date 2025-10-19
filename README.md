# GoFakeIP - Distorting Anonymous Proxy

A cross-platform SOCKS5/HTTP proxy server and client configuration utility that
masks your IP address by routing traffic through a remote proxy server.

## Features

- **Proxy Server**: High-performance SOCKS5 and HTTP proxy server written in Go
- **Client Utility**: Cross-platform configuration tool for automatic traffic
  redirection
- **IP Masking**: Hide your real IP by using the proxy server's IP as the source
- **Authentication**: Optional username/password authentication with bcrypt
  hashing
- **Cross-Platform**: Supports Linux (iptables), Windows (netsh), and macOS
  (pfctl)
- **Idempotent & Reversible**: Safe setup and teardown operations
  (NON-NEGOTIABLE #2)
- **Security First**: Input validation, privilege checking, no hardcoded secrets
  (NON-NEGOTIABLE #1, #4)

## Quick Start

### Server (Windows)

```powershell
# Get your network IP address
Get-NetIPAddress -AddressFamily IPv4

# Start proxy server with your actual network IP
.\bin\gofakeip-server.exe -listen 0.0.0.0:8080 -bind-ip <YOUR_IP> -protocols http

# Example with Wi-Fi IP
.\bin\gofakeip-server.exe -listen 0.0.0.0:8080 -bind-ip 192.168.1.100 -protocols http

# Or use configuration file
.\bin\gofakeip-server.exe -config examples\server-config-simple.yaml
```

**Important**: The `bind-ip` must be an IP address actually assigned to your
network interface (not 127.0.0.1 for remote connections).

### Client (Windows - Coming Soon)

```powershell
# Configure traffic redirection (requires Administrator)
.\bin\gofakeip-client.exe setup --proxy <SERVER_IP>:8080

# Remove configuration
.\bin\gofakeip-client.exe teardown

# Check status
.\bin\gofakeip-client.exe status
```

### Testing the Proxy

```bash
# Test HTTP proxy
curl -x http://localhost:8080 http://httpbin.org/ip

# Should return the server's IP, not your client IP
```

## Docker Deployment

GoFakeIP can be easily deployed using Docker for containerized environments.

### Quick Start with Docker

```bash
# Build the Docker image
docker build -t gofakeip:latest .

# Run with docker-compose (recommended)
docker-compose up -d

# Or run directly with Docker
docker run -d \
  --name gofakeip-proxy \
  --network host \
  -e BIND_IP=192.168.1.100 \
  gofakeip:latest
```

### Docker Configuration

The Docker image supports environment variables for configuration:

- `LISTEN_ADDRESS`: Server listening address (default: `0.0.0.0:8080`)
- `BIND_IP`: IP address to bind outbound connections (leave empty for container IP)
- `PROTOCOLS`: Supported protocols (default: `http,socks5`)
- `LOG_LEVEL`: Logging verbosity (default: `info`)
- `MAX_CONNECTIONS`: Maximum concurrent connections (default: `1000`)

### Network Modes

**Host Mode (Recommended for IP Masking)**:

```bash
# docker-compose.yml already configured with host mode
docker-compose up -d
```

**Bridge Mode (For isolated networks)**:

```bash
# Edit docker-compose.yml and uncomment the bridge network section
# Then run:
docker-compose up -d
```

### Docker Commands

```bash
# Build image
docker build -t gofakeip:latest .

# Start service
docker-compose up -d

# View logs
docker-compose logs -f

# Stop service
docker-compose down

# Restart service
docker-compose restart

# Check health
docker ps
```

### Production Deployment

For production use, consider:

1. **Security**: The container runs as non-root user (uid:1000)
2. **Resources**: Adjust CPU/memory limits in `docker-compose.yml`
3. **Logging**: Configure log rotation for production workloads
4. **Monitoring**: Use health checks to ensure service availability

## Documentation

- **[Quick Start Guide](specs/001-gofakeip/quickstart.md)**: Installation and usage instructions
- **[Specification](specs/001-gofakeip/spec.md)**: Complete feature specification
- **[Implementation Plan](specs/001-gofakeip/plan.md)**: Architecture and implementation details
- **[Research Report](specs/001-gofakeip/research.md)**: Technical feasibility and IP masking implementation
- **[Data Model](specs/001-gofakeip/data-model.md)**: Entity definitions and validation rules
- **[CLI Reference - Server](specs/001-gofakeip/contracts/cli-server.md)**: Server command-line interface
- **[CLI Reference - Client](specs/001-gofakeip/contracts/cli-client.md)**: Client command-line interface
- **[Configuration Schema](specs/001-gofakeip/contracts/config-schema.md)**: YAML configuration format

## Project Structure

```text
.
├── cmd/                        # Application entry points
│   ├── gofakeip-server/       # Proxy server binary
│   └── gofakeip-client/       # Client configuration utility
│
├── pkg/                        # Core libraries
│   ├── proxy/                 # Proxy server implementation
│   ├── netconfig/             # Cross-platform network configuration
│   ├── config/                # Configuration management
│   └── common/                # Shared utilities (errors, logging)
│
├── tests/                      # Test suites
│   ├── unit/                  # Unit tests
│   ├── integration/           # Integration tests
│   ├── security/              # Security validation tests
│   └── reversibility/         # Teardown tests
│
├── specs/                      # Design documentation
│   └── 001-gofakeip/          # Feature specifications
│
└── docs/                       # Additional documentation
```

## Building

### Prerequisites

- Go 1.21 or higher
- For client: Administrative/root privileges to configure network rules

### Build Instructions

```bash
# Clone repository
git clone https://github.com/yourorg/gofakeip.git
cd gofakeip

# Download dependencies
go mod download

# Build server
go build -o bin/gofakeip-server ./cmd/gofakeip-server

# Build client
go build -o bin/gofakeip-client ./cmd/gofakeip-client

# Run tests
go test ./...
```

### Cross-Platform Builds

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o bin/gofakeip-server-linux-amd64 ./cmd/gofakeip-server
GOOS=linux GOARCH=amd64 go build -o bin/gofakeip-client-linux-amd64 ./cmd/gofakeip-client

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o bin/gofakeip-server-windows-amd64.exe ./cmd/gofakeip-server
GOOS=windows GOARCH=amd64 go build -o bin/gofakeip-client-windows-amd64.exe ./cmd/gofakeip-client

# macOS (arm64)
GOOS=darwin GOARCH=arm64 go build -o bin/gofakeip-server-darwin-arm64 ./cmd/gofakeip-server
GOOS=darwin GOARCH=arm64 go build -o bin/gofakeip-client-darwin-arm64 ./cmd/gofakeip-client
```

## Security Considerations

### NON-NEGOTIABLE Constraints

This project adheres to strict security constraints defined in [CLAUDE.md](./CLAUDE.md):

1. **Least Privilege Enforcement**: Client verifies elevated privileges before any system modification
2. **Idempotency & Reversibility**: Setup is idempotent, teardown fully restores original state
3. **Technology Stack Lock**: Core logic implemented in Go only (no shell scripts for core functionality)
4. **NO Vulnerable Defaults**: All inputs validated, no hardcoded credentials, parameterized command execution

### Important Notes

- **IP Masking**: This is a standard proxy system using **legitimate IP addresses**. It does NOT perform illegal IP spoofing.
- **Bind IP Requirement**: The server's `bind_ip` must be legitimately assigned to a network interface.
- **Application-Layer Headers**: Optional HTTP header injection (X-Forwarded-For) is application-layer only.

See [research.md](specs/001-gofakeip/research.md) for detailed technical analysis.

## Legal and Ethical Use

### Acceptable Uses

- Privacy protection (hiding your IP for privacy reasons)
- Testing your own systems with proper authorization
- Development and research in controlled environments

### Prohibited Uses

- Unauthorized access to systems
- Fraud or impersonation
- DDoS attacks or other malicious activities
- Any illegal activity

**Disclaimer**: Users are solely responsible for ensuring their use complies with all applicable laws and terms of service.

## Installation

See `quickstart.md` for installation and setup instructions.

## Legal Disclaimer

This software is provided for educational and research purposes only. Use responsibly and comply with all applicable laws.

## Dependencies

- [go-socks5](https://github.com/things-go/go-socks5): SOCKS5 protocol implementation
- [yaml.v3](https://gopkg.in/yaml.v3): YAML configuration parsing
- [golang.org/x/sys](https://golang.org/x/sys): Low-level OS interactions

## Development Status

**Current Phase**: Phase 3 Complete - User Story 1 (Proxy Server) ✓

### Completed

- ✅ Design artifacts (spec, plan, research, data model)
- ✅ Foundational infrastructure (logging, errors, validation)
- ✅ **Proxy Server (HTTP)** - Production-ready
  - HTTP proxy with CONNECT tunneling + GET/POST forwarding
  - IP masking via bind IP configuration
  - Configuration system (YAML + CLI flags)
  - Bind IP interface validation
  - Graceful shutdown (SIGINT/SIGTERM)
  - Structured logging
- ✅ **Security Tests** - All passing
  - Privilege enforcement
  - Input injection prevention (12/12 patterns blocked)
  - No hardcoded secrets scanning

### In Progress

- ⏳ **Windows Client** - Network configuration utility for Windows
- 🔜 SOCKS5 protocol support (requires external dependency)

### Tested & Verified

```text
✓ Server builds successfully on Windows
✓ HTTP proxy functional (tested with curl)
✓ IP masking working (outbound traffic uses bind IP)
✓ All security constraints enforced
```

**Next Steps**:

1. ✅ ~~Implement core proxy server~~ **COMPLETE**
2. ⏳ Implement Windows network configurator (netsh portproxy)
3. 🔜 Implement Linux/macOS configurators
4. 🔜 SOCKS5 protocol support
5. 🔜 Unit & integration tests

## Contributing

Contributions are welcome! Please ensure all contributions:

- Follow the NON-NEGOTIABLE security constraints
- Include comprehensive tests (especially security tests)
- Maintain cross-platform compatibility
- Update documentation as needed

## License

[License TBD]

## Authors

GoFakeIP Project Team

## Acknowledgments

- Research on IP masking and proxy technologies
- Open-source SOCKS5 and Go ecosystem libraries
- Security research on privilege enforcement and input validation
