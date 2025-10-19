# Configuration Schema: GoFakeIP

**Feature Branch**: `001-gofakeip`
**Created**: 2025-10-19
**Format**: YAML
**Related**: [cli-server.md](./cli-server.md), [cli-client.md](./cli-client.md)

---

## Overview

GoFakeIP supports configuration via YAML files for both server and client components. This document defines the schema, validation rules, and examples.

**Benefits of YAML Configuration**:
- No passwords on command line (process list security)
- Pre-hashed passwords for authentication
- Reusable configuration profiles
- Version control friendly (with secret management)

---

## Server Configuration

### File Locations

**Default Search Paths** (in order):
1. Path specified by `--config` flag
2. `./server-config.yaml` (current directory)
3. `/etc/gofakeip/server.yaml` (Linux/macOS)
4. `C:\ProgramData\GoFakeIP\server.yaml` (Windows)

---

### Schema

```yaml
# server-config.yaml

server:
  listen_address: string        # Required
  bind_ip: string              # Required (must be assigned to server interface)
  protocols: []string          # Optional, default: [socks5, http]
  max_connections: int         # Optional, default: 1000
  connection_timeout: duration # Optional, default: 30s

auth:
  enabled: bool                # Optional, default: false
  method: string               # Required if enabled, values: none | username_password
  username: string             # Required if enabled
  password_hash: string        # Required if enabled (bcrypt hash, NOT plain text)

http_headers:
  inject_x_forwarded_for: string  # Optional, default: "" (disabled)
  inject_x_real_ip: string        # Optional, default: "" (disabled)
  inject_via: string              # Optional, default: "" (disabled)
  remove_existing: bool           # Optional, default: false

logging:
  level: string                # Optional, default: info, values: debug | info | warn | error
  format: string               # Optional, default: text, values: text | json
  file: string                 # Optional, default: "" (stdout)
```

---

### Complete Example

```yaml
# server-config.yaml
server:
  listen_address: "0.0.0.0:1080"
  bind_ip: "203.0.113.50"
  protocols:
    - socks5
    - http
  max_connections: 5000
  connection_timeout: 60s

auth:
  enabled: true
  method: username_password
  username: "admin"
  # Generate with: gofakeip-server hash-password mypassword (future feature)
  password_hash: "$2a$10$rZQk3pXx8kZJmG7Z5vY8AO0.8mVFqK4kV1YzXmQzLpYvEqZ7LqK5S"

http_headers:
  inject_x_forwarded_for: "198.51.100.25"
  inject_x_real_ip: "198.51.100.25"
  inject_via: "1.1 gofakeip-proxy"
  remove_existing: true

logging:
  level: info
  format: json
  file: "/var/log/gofakeip/server.log"
```

---

### Minimal Example (Development)

```yaml
# server-config.yaml
server:
  listen_address: "127.0.0.1:1080"
  bind_ip: "127.0.0.1"
```

All other values use defaults.

---

### Field Definitions

#### `server.listen_address`

**Type**: `string`
**Required**: Yes
**Format**: `IP:PORT` or `:PORT`

**Description**: Address and port to listen on for proxy connections.

**Examples**:
```yaml
listen_address: "0.0.0.0:1080"    # All interfaces
listen_address: "127.0.0.1:1080"  # Localhost only
listen_address: ":1080"            # All interfaces, port 1080
listen_address: "[::]:1080"        # All IPv6 interfaces
```

**Validation**:
- Must be valid IP:port or :port format
- Port must be 1-65535
- IP must be valid IPv4/IPv6 or empty

---

#### `server.bind_ip`

**Type**: `string`
**Required**: Yes
**Format**: Valid IPv4 or IPv6 address

**Description**: Source IP address for outbound connections to destination servers. This is the IP that destination servers will see.

**Examples**:
```yaml
bind_ip: "203.0.113.50"        # IPv4
bind_ip: "2001:db8::1"         # IPv6
```

**Validation**:
- Must be valid IP address
- **CRITICAL**: Must be assigned to a server network interface
- Server validates at startup, fails if IP not found

**Security Note**: Must be legitimately owned by the server (see research.md).

---

#### `server.protocols`

**Type**: `[]string`
**Required**: No
**Default**: `[socks5, http]`

**Description**: List of proxy protocols to enable.

**Valid Values**: `socks5`, `http`

**Examples**:
```yaml
protocols:
  - socks5         # SOCKS5 only
```

```yaml
protocols:
  - socks5
  - http           # Both SOCKS5 and HTTP
```

**Validation**:
- Must contain at least one valid protocol
- Unknown protocols cause validation error

---

#### `server.max_connections`

**Type**: `int`
**Required**: No
**Default**: `1000`

**Description**: Maximum number of concurrent proxy connections.

**Examples**:
```yaml
max_connections: 5000    # High-traffic server
max_connections: 100     # Low-resource environment
```

**Validation**:
- Must be > 0
- Consider system file descriptor limits (ulimit -n)

---

#### `server.connection_timeout`

**Type**: `duration` (Go duration string)
**Required**: No
**Default**: `30s`

**Description**: Timeout for idle connections.

**Format**: Go duration format (e.g., `30s`, `5m`, `1h30m`)

**Examples**:
```yaml
connection_timeout: 30s    # 30 seconds
connection_timeout: 5m     # 5 minutes
connection_timeout: 1h30m  # 1 hour 30 minutes
```

**Validation**:
- Must be valid Go duration
- Must be > 0

---

#### `auth.enabled`

**Type**: `bool`
**Required**: No
**Default**: `false`

**Description**: Whether to require authentication for proxy connections.

**Examples**:
```yaml
auth:
  enabled: true    # Require authentication
```

```yaml
auth:
  enabled: false   # Open proxy (NOT recommended for production)
```

**Security**: Set to `true` for production to prevent unauthorized use.

---

#### `auth.method`

**Type**: `string`
**Required**: Conditional (required if `auth.enabled: true`)
**Default**: `none`

**Description**: Authentication method to use.

**Valid Values**:
- `none`: No authentication (requires `enabled: false`)
- `username_password`: Username and password authentication (SOCKS5/HTTP Basic)

**Examples**:
```yaml
auth:
  enabled: true
  method: username_password
```

---

#### `auth.username`

**Type**: `string`
**Required**: Conditional (required if `auth.enabled: true`)

**Description**: Username for authentication.

**Examples**:
```yaml
auth:
  enabled: true
  method: username_password
  username: "admin"
```

**Validation**:
- Must be non-empty if auth enabled
- No special character restrictions (URL-encoded in HTTP)

---

#### `auth.password_hash`

**Type**: `string`
**Required**: Conditional (required if `auth.enabled: true`)
**Format**: bcrypt hash

**Description**: Bcrypt hash of the password. **NEVER use plain text passwords.**

**Examples**:
```yaml
auth:
  enabled: true
  method: username_password
  username: "admin"
  password_hash: "$2a$10$rZQk3pXx8kZJmG7Z5vY8AO0.8mVFqK4kV1YzXmQzLpYvEqZ7LqK5S"
```

**Generating Hash**:
```bash
# Future feature:
gofakeip-server hash-password mypassword

# External tool (bcrypt):
htpasswd -nbB admin mypassword
# Output: admin:$2y$10$...
```

**Validation**:
- Must be valid bcrypt hash format
- Must start with `$2a$`, `$2b$`, or `$2y$`

**Security** (NON-NEGOTIABLE #4):
- Plain text passwords FORBIDDEN in config files
- Bcrypt provides protection even if config file leaked
- Never log or output password hash

---

#### `http_headers.inject_x_forwarded_for`

**Type**: `string`
**Required**: No
**Default**: `""` (disabled)

**Description**: Value to inject in X-Forwarded-For HTTP header (distorting proxy feature).

**Examples**:
```yaml
http_headers:
  inject_x_forwarded_for: "198.51.100.25"
```

**Notes**:
- Application-layer only (HTTP traffic)
- Does not affect network-layer source IP
- Optional feature for "distorting proxy" classification

---

#### `http_headers.inject_x_real_ip`

**Type**: `string`
**Required**: No
**Default**: `""` (disabled)

**Description**: Value to inject in X-Real-IP HTTP header.

**Examples**:
```yaml
http_headers:
  inject_x_real_ip: "198.51.100.25"
```

---

#### `http_headers.inject_via`

**Type**: `string`
**Required**: No
**Default**: `""` (disabled)

**Description**: Value to inject in Via HTTP header.

**Examples**:
```yaml
http_headers:
  inject_via: "1.1 gofakeip-proxy"
```

**Format**: HTTP Via header format (RFC 7230)

---

#### `http_headers.remove_existing`

**Type**: `bool`
**Required**: No
**Default**: `false`

**Description**: Remove existing proxy headers (X-Forwarded-For, X-Real-IP, Via) before injecting new ones.

**Examples**:
```yaml
http_headers:
  inject_x_forwarded_for: "198.51.100.25"
  remove_existing: true   # Remove client-provided X-Forwarded-For first
```

**Use Case**: Prevent clients from spoofing proxy headers.

---

#### `logging.level`

**Type**: `string`
**Required**: No
**Default**: `info`

**Description**: Logging verbosity level.

**Valid Values**: `debug`, `info`, `warn`, `error`

**Examples**:
```yaml
logging:
  level: debug    # Verbose (development)
  level: info     # Normal (production)
  level: error    # Minimal (only errors)
```

---

#### `logging.format`

**Type**: `string`
**Required**: No
**Default**: `text`

**Description**: Log output format.

**Valid Values**:
- `text`: Human-readable text format
- `json`: JSON format (for log aggregation tools)

**Examples**:
```yaml
logging:
  format: json    # {"level":"info","time":"2025-10-19T14:30:00Z","msg":"Server started"}
  format: text    # [INFO] 2025-10-19 14:30:00 Server started
```

---

#### `logging.file`

**Type**: `string`
**Required**: No
**Default**: `""` (stdout)

**Description**: Path to log file. If empty, logs to stdout.

**Examples**:
```yaml
logging:
  file: "/var/log/gofakeip/server.log"
  file: "C:\\ProgramData\\GoFakeIP\\server.log"  # Windows
  file: ""  # stdout (default)
```

**Note**: Ensure directory exists and is writable.

---

## Client Configuration

### File Locations

**Default Search Paths**:
1. Path specified by `--config` flag
2. `./client-config.yaml` (current directory)
3. `/etc/gofakeip/client.yaml` (Linux/macOS)
4. `C:\ProgramData\GoFakeIP\client.yaml` (Windows)

---

### Schema

```yaml
# client-config.yaml

proxy:
  address: string           # Required (proxy server address:port)

redirection:
  ports: []int              # Optional, default: all
  protocol: string          # Optional, default: tcp, values: tcp | udp | both
  interface: string         # Optional, platform-specific

platform:
  override: string          # Optional, for testing, values: linux | windows | darwin
```

---

### Complete Example

```yaml
# client-config.yaml
proxy:
  address: "203.0.113.50:1080"

redirection:
  ports:
    - 80
    - 443
    - 8080
  protocol: tcp
  interface: "eth0"  # Linux/macOS specific
```

---

### Minimal Example

```yaml
# client-config.yaml
proxy:
  address: "203.0.113.50:1080"
```

Redirects all TCP traffic through the proxy.

---

### Field Definitions

#### `proxy.address`

**Type**: `string`
**Required**: Yes
**Format**: `HOST:PORT`

**Description**: Proxy server address and port.

**Examples**:
```yaml
proxy:
  address: "203.0.113.50:1080"
  address: "proxy.example.com:1080"
```

**Validation**:
- Must be valid host:port format
- Port must be 1-65535
- Host must be valid IP or resolvable hostname

---

#### `redirection.ports`

**Type**: `[]int`
**Required**: No
**Default**: All ports

**Description**: List of ports to redirect. If empty or omitted, all ports redirected.

**Examples**:
```yaml
redirection:
  ports:
    - 80
    - 443      # HTTP and HTTPS only
```

```yaml
redirection:
  ports: []    # All ports (default)
```

**Validation**:
- Each port must be 1-65535

**Note**: Port ranges not supported in YAML (use CLI flags for ranges).

---

#### `redirection.protocol`

**Type**: `string`
**Required**: No
**Default**: `tcp`

**Description**: Network protocol to redirect.

**Valid Values**: `tcp`, `udp`, `both`

**Examples**:
```yaml
redirection:
  protocol: tcp    # TCP only
  protocol: both   # TCP and UDP
```

---

#### `redirection.interface`

**Type**: `string`
**Required**: No
**Default**: Platform-specific

**Description**: Network interface to apply rules to.

**Examples**:
```yaml
redirection:
  interface: "eth0"   # Linux
  interface: "en0"    # macOS
```

**Platform-Specific**:
- **Linux**: Optional (default: all interfaces via OUTPUT chain)
- **Windows**: Not applicable (portproxy is interface-agnostic)
- **macOS**: Recommended (default: primary interface)

---

#### `platform.override`

**Type**: `string`
**Required**: No
**Default**: Auto-detected

**Description**: Override platform detection (for testing only).

**Valid Values**: `linux`, `windows`, `darwin`

**Example**:
```yaml
platform:
  override: "linux"  # Force Linux mode (testing only)
```

**Warning**: Only for testing. Incorrect platform may cause failures.

---

## Validation

### Server Config Validation

**Required Fields**:
- `server.listen_address`
- `server.bind_ip`

**Conditional Requirements**:
- If `auth.enabled: true`:
  - `auth.method` required
  - `auth.username` required
  - `auth.password_hash` required

**Validation Rules**:
1. `bind_ip` must be valid IP address
2. `bind_ip` must be assigned to a server interface (checked at runtime)
3. `listen_address` must be valid IP:port
4. `protocols` must contain at least one valid protocol
5. `password_hash` must be valid bcrypt format if auth enabled
6. All durations must be valid Go duration strings

---

### Client Config Validation

**Required Fields**:
- `proxy.address`

**Validation Rules**:
1. `proxy.address` must be valid host:port
2. All ports in `redirection.ports` must be 1-65535
3. `redirection.protocol` must be valid value
4. No shell metacharacters in any field (injection prevention)

---

## Security Considerations (NON-NEGOTIABLE #4)

### NO Hardcoded Secrets

```yaml
# ❌ FORBIDDEN - Plain text password
auth:
  password: "mypassword"  # NEVER DO THIS

# ✅ CORRECT - Bcrypt hash
auth:
  password_hash: "$2a$10$..."
```

---

### Input Validation

All config values validated before use:

```go
func validateConfig(config *ServerConfig) error {
    // Validate bind_ip format
    if net.ParseIP(config.Server.BindIP) == nil {
        return fmt.Errorf("invalid bind_ip: %s", config.Server.BindIP)
    }

    // Validate bind_ip is assigned to interface
    if !isIPAssignedToInterface(config.Server.BindIP) {
        return fmt.Errorf("bind_ip not found on any interface: %s", config.Server.BindIP)
    }

    // Validate bcrypt hash if auth enabled
    if config.Auth.Enabled {
        if !isValidBcryptHash(config.Auth.PasswordHash) {
            return errors.New("invalid bcrypt password_hash")
        }
    }

    return nil
}
```

---

### File Permissions

**Recommended Permissions**:

**Linux/macOS**:
```bash
chmod 600 /etc/gofakeip/server.yaml
chown root:root /etc/gofakeip/server.yaml
```

**Windows**:
```powershell
icacls C:\ProgramData\GoFakeIP\server.yaml /inheritance:r /grant:r "NT AUTHORITY\SYSTEM:(F)" "BUILTIN\Administrators:(F)"
```

**Rationale**: Config files contain password hashes (sensitive data).

---

## Configuration Examples

### Public SOCKS5 Proxy (No Auth)

```yaml
# ⚠️ WARNING: Open proxy - use only in controlled environments
server:
  listen_address: "0.0.0.0:1080"
  bind_ip: "203.0.113.50"
  protocols:
    - socks5
  max_connections: 10000

auth:
  enabled: false
```

**Security Warning**: Open proxies can be abused. Use authentication in production.

---

### Private HTTP Proxy with Auth

```yaml
server:
  listen_address: "0.0.0.0:8080"
  bind_ip: "203.0.113.50"
  protocols:
    - http
  max_connections: 1000

auth:
  enabled: true
  method: username_password
  username: "admin"
  password_hash: "$2a$10$rZQk3pXx8kZJmG7Z5vY8AO0.8mVFqK4kV1YzXmQzLpYvEqZ7LqK5S"

logging:
  level: info
  format: json
  file: "/var/log/gofakeip/server.log"
```

---

### Distorting Proxy (HTTP Headers)

```yaml
server:
  listen_address: ":8080"
  bind_ip: "203.0.113.50"
  protocols:
    - http

http_headers:
  inject_x_forwarded_for: "198.51.100.25"
  inject_x_real_ip: "198.51.100.25"
  inject_via: "1.1 gofakeip"
  remove_existing: true
```

---

## Environment Variable Substitution (Future Enhancement)

**Planned Feature**: Support environment variable interpolation

```yaml
# Future feature
server:
  bind_ip: "${BIND_IP}"

auth:
  username: "${PROXY_USERNAME}"
  password_hash: "${PROXY_PASSWORD_HASH}"
```

**Status**: Not implemented in initial version

---

## JSON Schema (for Validation Tools)

Future enhancement: Provide JSON Schema for YAML validation in IDEs.

---

## Summary

The GoFakeIP configuration schema provides:

- ✅ **Clear structure** for server and client settings
- ✅ **Security by default** (bcrypt hashing, validation)
- ✅ **Flexible deployment** (YAML files + CLI flag overrides)
- ✅ **NO vulnerable defaults** (NON-NEGOTIABLE #4)
- ✅ **Comprehensive validation** (format, ranges, platform checks)

**See Also**:
- [cli-server.md](./cli-server.md) - Server CLI flags
- [cli-client.md](./cli-client.md) - Client CLI flags
- [../data-model.md](../data-model.md) - Data structures
