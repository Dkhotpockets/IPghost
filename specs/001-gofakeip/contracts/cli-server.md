# CLI Contract: gofakeip-server

**Feature Branch**: `001-gofakeip`
**Created**: 2025-10-19
**Binary**: `gofakeip-server`
**Purpose**: SOCKS5/HTTP proxy server with IP masking capabilities

---

## Command Synopsis

```
gofakeip-server [OPTIONS]
gofakeip-server --config <path>
gofakeip-server --version
gofakeip-server --help
```

---

## Global Flags

### `--config, -c`

**Type**: `string`
**Required**: No
**Default**: `./server-config.yaml` (if exists), otherwise uses flag defaults

**Description**: Path to YAML configuration file.

**Example**:
```bash
gofakeip-server --config /etc/gofakeip/server.yaml
```

**Behavior**:
- If file doesn't exist and no other flags provided: Error
- If file doesn't exist but flags provided: Use flag values
- If file exists: Load from file, flags override file values
- Config file format must be valid YAML

---

### `--listen, -l`

**Type**: `string`
**Required**: Conditional (required if no config file)
**Default**: None

**Description**: Address and port to listen on for proxy connections.

**Format**: `IP:PORT` or `:PORT` (binds to all interfaces)

**Example**:
```bash
gofakeip-server --listen 0.0.0.0:1080
gofakeip-server --listen :1080
```

**Validation**:
- Must be valid IP:port format
- Port must be 1-65535
- IP must be valid IPv4/IPv6 or empty (for all interfaces)

---

### `--bind-ip`

**Type**: `string`
**Required**: Conditional (required if no config file)
**Default**: None

**Description**: Source IP address to bind for outbound connections. This IP appears as the source when connecting to destination servers.

**Format**: Valid IPv4 or IPv6 address

**Example**:
```bash
gofakeip-server --bind-ip 203.0.113.50
```

**Validation**:
- Must be a valid IP address
- **CRITICAL**: Must be assigned to one of the server's network interfaces
- Server will validate at startup and fail if IP not found

**Security Note**: This IP must be legitimately owned by the server. Arbitrary IP addresses will not work (see research.md).

---

### `--protocols`

**Type**: `string` (comma-separated)
**Required**: No
**Default**: `socks5,http`

**Description**: Proxy protocols to enable.

**Valid Values**: `socks5`, `http` (comma-separated for multiple)

**Example**:
```bash
gofakeip-server --protocols socks5
gofakeip-server --protocols socks5,http
```

**Validation**:
- Must contain at least one valid protocol
- Unknown protocols are rejected with error

---

### `--auth-username`

**Type**: `string`
**Required**: No (auth disabled if omitted)
**Default**: None

**Description**: Username for proxy authentication. If provided, `--auth-password` is also required.

**Example**:
```bash
gofakeip-server --auth-username admin --auth-password secretpass
```

**Security**:
- Password is hashed with bcrypt before storage
- Never logged or stored in plain text

---

### `--auth-password`

**Type**: `string`
**Required**: Conditional (required if `--auth-username` provided)
**Default**: None

**Description**: Password for proxy authentication.

**Security**:
- Immediately hashed with bcrypt on receipt
- Plain text never persisted to disk or logs
- Consider using config file with pre-hashed password instead

---

### `--max-connections`

**Type**: `int`
**Required**: No
**Default**: `1000`

**Description**: Maximum number of concurrent proxy connections.

**Example**:
```bash
gofakeip-server --max-connections 5000
```

**Validation**:
- Must be > 0
- Recommended: Set based on system resources (file descriptor limits)

---

### `--connection-timeout`

**Type**: `duration`
**Required**: No
**Default**: `30s`

**Description**: Timeout for idle connections.

**Format**: Go duration format (e.g., `30s`, `5m`, `1h30m`)

**Example**:
```bash
gofakeip-server --connection-timeout 60s
```

**Validation**:
- Must be > 0
- Must be valid Go duration format

---

### `--http-inject-xff`

**Type**: `string`
**Required**: No
**Default**: None (no injection)

**Description**: Value to inject in X-Forwarded-For header for HTTP traffic (distorting proxy feature).

**Format**: Any valid IP address string

**Example**:
```bash
gofakeip-server --http-inject-xff 198.51.100.25
```

**Note**: Application-layer only. Does not affect network-layer source IP.

---

### `--http-inject-xreal`

**Type**: `string`
**Required**: No
**Default**: None (no injection)

**Description**: Value to inject in X-Real-IP header for HTTP traffic.

**Format**: Any valid IP address string

**Example**:
```bash
gofakeip-server --http-inject-xreal 198.51.100.25
```

---

### `--http-remove-existing`

**Type**: `bool`
**Required**: No
**Default**: `false`

**Description**: Remove existing proxy headers (X-Forwarded-For, X-Real-IP, Via) before injecting new ones.

**Example**:
```bash
gofakeip-server --http-remove-existing
```

---

### `--version, -v`

**Type**: Flag (boolean)
**Required**: No

**Description**: Print version information and exit.

**Example**:
```bash
gofakeip-server --version
```

**Output Format**:
```
gofakeip-server version 1.0.0
Built: 2025-10-19
Go version: go1.21.5
```

---

### `--help, -h`

**Type**: Flag (boolean)
**Required**: No

**Description**: Print help message and exit.

**Example**:
```bash
gofakeip-server --help
```

---

## Usage Examples

### Basic SOCKS5 Proxy

```bash
gofakeip-server --listen 0.0.0.0:1080 --bind-ip 203.0.113.50
```

Starts a SOCKS5 proxy on all interfaces, port 1080. Outbound connections use 203.0.113.50 as source IP.

---

### SOCKS5 + HTTP with Authentication

```bash
gofakeip-server \
  --listen :1080 \
  --bind-ip 203.0.113.50 \
  --protocols socks5,http \
  --auth-username admin \
  --auth-password mypassword
```

Enables both SOCKS5 and HTTP proxy with authentication required.

---

### HTTP Proxy with Header Injection

```bash
gofakeip-server \
  --listen :8080 \
  --bind-ip 203.0.113.50 \
  --protocols http \
  --http-inject-xff 198.51.100.25 \
  --http-inject-xreal 198.51.100.25 \
  --http-remove-existing
```

HTTP proxy that injects fake IP in HTTP headers (distorting proxy).

---

### Using Configuration File

```bash
gofakeip-server --config /etc/gofakeip/server.yaml
```

Loads all settings from YAML file (see config-schema.md).

---

### Override Config File with Flags

```bash
gofakeip-server --config server.yaml --bind-ip 192.0.2.10
```

Loads from `server.yaml` but overrides `bind_ip` with flag value.

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (clean shutdown) |
| 1 | Configuration error (invalid flags, bad config file) |
| 2 | Network error (cannot bind to address, bind IP not found) |
| 3 | Runtime error (critical failure during operation) |
| 130 | Interrupted by SIGINT (Ctrl+C) |

---

## Signals

### `SIGINT` / `SIGTERM`

**Behavior**: Graceful shutdown
1. Stop accepting new connections
2. Wait for active connections to close (with timeout)
3. Release resources
4. Exit with code 0

**Max Shutdown Time**: 30 seconds (then force exit)

---

### `SIGHUP`

**Behavior**: Reload configuration (if applicable in future versions)

**Current**: Not implemented (ignored)

---

## Output Format

### Standard Output

**Normal Operation**: Minimal output, only critical messages

**Example**:
```
[INFO] Server starting on 0.0.0.0:1080
[INFO] Bind IP: 203.0.113.50
[INFO] Protocols: SOCKS5, HTTP
[INFO] Server ready
```

---

### Verbose Mode (Future Enhancement)

**Flag**: `--verbose` or `--debug`

**Behavior**: Detailed logging including:
- Connection events
- Protocol negotiations
- Data transfer statistics

---

### Error Output (stderr)

**Format**: Human-readable error messages

**Example**:
```
[ERROR] Failed to bind to 0.0.0.0:1080: address already in use
[ERROR] Bind IP 203.0.113.50 not found on any network interface
```

---

## Configuration Precedence

When both config file and flags are provided:

**Precedence (highest to lowest)**:
1. Command-line flags
2. Configuration file values
3. Built-in defaults

**Example**:
```yaml
# config.yaml
server:
  listen_address: "0.0.0.0:1080"
  bind_ip: "203.0.113.50"
```

```bash
gofakeip-server --config config.yaml --bind-ip 192.0.2.10
```

**Result**: Uses `192.0.2.10` (flag overrides config file)

---

## Validation & Error Messages

### Invalid Bind IP

**Command**:
```bash
gofakeip-server --listen :1080 --bind-ip 999.999.999.999
```

**Output**:
```
[ERROR] Invalid bind IP address: 999.999.999.999
Exit code: 1
```

---

### Bind IP Not Found on Interface

**Command**:
```bash
gofakeip-server --listen :1080 --bind-ip 203.0.113.50
```

**Output** (if IP not assigned):
```
[ERROR] Bind IP 203.0.113.50 not found on any network interface
Available interfaces:
  - eth0: 192.168.1.10
  - lo: 127.0.0.1
Exit code: 2
```

---

### Port Already in Use

**Command**:
```bash
gofakeip-server --listen :1080 --bind-ip 203.0.113.50
```

**Output** (if port 1080 in use):
```
[ERROR] Failed to bind to 0.0.0.0:1080: address already in use
Exit code: 2
```

---

### Missing Required Flags

**Command**:
```bash
gofakeip-server
```

**Output** (no config file, no flags):
```
[ERROR] Missing required configuration:
  - listen address (use --listen or --config)
  - bind IP (use --bind-ip or --config)
Exit code: 1
```

---

### Authentication Misconfiguration

**Command**:
```bash
gofakeip-server --listen :1080 --bind-ip 203.0.113.50 --auth-username admin
```

**Output**:
```
[ERROR] --auth-password is required when --auth-username is provided
Exit code: 1
```

---

## Security Considerations

### Command-Line Password Visibility

**Warning**: Using `--auth-password` on command line exposes password in process list.

**Recommendation**: Use configuration file with pre-hashed password:

```yaml
auth:
  enabled: true
  username: "admin"
  password_hash: "$2a$10$..."  # bcrypt hash
```

Generate hash:
```bash
# External tool or future feature:
gofakeip-server hash-password mypassword
```

---

### Bind IP Validation

Server performs strict validation:
1. IP address format check
2. Interface enumeration
3. Fails startup if bind IP not found

**Rationale**: Prevents misconfiguration and ensures IP is legitimately owned (NON-NEGOTIABLE #4).

---

## Performance Tuning

### Max Connections

Set based on system limits:

```bash
# Check file descriptor limit
ulimit -n

# Set max connections accordingly (leave headroom)
gofakeip-server --max-connections 50000
```

---

### Connection Timeout

Shorter timeouts free resources faster:

```bash
gofakeip-server --connection-timeout 15s  # For high-throughput scenarios
```

Longer timeouts better for slow connections:

```bash
gofakeip-server --connection-timeout 5m  # For long-lived connections
```

---

## Future Enhancements

Potential future flags (not in initial implementation):

- `--log-file`: Write logs to file
- `--log-format`: JSON or plain text
- `--tls-cert`: Enable TLS for proxy connections
- `--rate-limit`: Connection rate limiting
- `--blacklist`: IP blacklist file
- `--metrics-port`: Prometheus metrics endpoint

---

## Summary

The `gofakeip-server` CLI provides:

- ✅ Simple configuration via flags or YAML file
- ✅ Strict input validation (NON-NEGOTIABLE #4)
- ✅ Clear error messages with exit codes
- ✅ Secure authentication (bcrypt hashing)
- ✅ Configurable protocols (SOCKS5/HTTP)
- ✅ Optional HTTP header injection (distorting proxy)

**See Also**:
- [config-schema.md](./config-schema.md) - Configuration file format
- [cli-client.md](./cli-client.md) - Client utility CLI
- [../data-model.md](../data-model.md) - Data structures
