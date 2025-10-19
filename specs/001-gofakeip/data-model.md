# Data Model: GoFakeIP

**Feature Branch**: `001-gofakeip`
**Created**: 2025-10-19
**Status**: Phase 1 - Design Artifacts
**Related**: [spec.md](./spec.md), [plan.md](./plan.md), [research.md](./research.md)

---

## Overview

This document defines the core data entities, their attributes, relationships, and state transitions for the GoFakeIP proxy system. The data model supports both the proxy server component and the client configuration utility.

---

## Core Entities

### 1. ProxyServer

**Description**: Represents the running proxy server instance that accepts client connections and forwards traffic with IP masking.

**Attributes**:

| Attribute | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `ListenAddress` | `string` | Yes | Address:port for proxy to listen on | Valid IP:port format, e.g., "0.0.0.0:1080" |
| `BindIP` | `net.IP` | Yes | Source IP for outbound connections | Must be assigned to server's network interface |
| `Protocols` | `[]Protocol` | Yes | Enabled proxy protocols | One or more of: SOCKS5, HTTP |
| `AuthConfig` | `*AuthConfig` | No | Authentication settings | nil = no auth required |
| `HTTPHeaderConfig` | `*HTTPHeaderConfig` | No | HTTP header injection settings | nil = no header manipulation |
| `MaxConnections` | `int` | Yes | Maximum concurrent connections | >0, default: 1000 |
| `ConnectionTimeout` | `time.Duration` | Yes | Timeout for idle connections | >0, default: 30s |
| `State` | `ServerState` | Yes | Current server state | One of: Stopped, Starting, Running, Stopping, Error |

**Go Struct**:
```go
type ProxyServer struct {
    ListenAddress      string
    BindIP             net.IP
    Protocols          []Protocol
    AuthConfig         *AuthConfig
    HTTPHeaderConfig   *HTTPHeaderConfig
    MaxConnections     int
    ConnectionTimeout  time.Duration

    // Runtime state
    State              ServerState
    ActiveConnections  map[string]*Connection
    connectionPool     *ConnectionPool
    listener           net.Listener
    mu                 sync.RWMutex
}

type Protocol string

const (
    ProtocolSOCKS5 Protocol = "socks5"
    ProtocolHTTP   Protocol = "http"
)

type ServerState int

const (
    ServerStateStopped ServerState = iota
    ServerStateStarting
    ServerStateRunning
    ServerStateStopping
    ServerStateError
)
```

**Relationships**:
- Contains multiple `Connection` instances (one-to-many)
- Has optional `AuthConfig` (one-to-one)
- Has optional `HTTPHeaderConfig` (one-to-one)

**Lifecycle**:
1. **Created** → Configuration loaded from file/flags
2. **Starting** → Bind to listen address, initialize connection pool
3. **Running** → Accept and handle connections
4. **Stopping** → Gracefully close active connections
5. **Stopped** → Release resources, close listener
6. **Error** → Critical failure, requires restart

**Validation Rules**:
- `BindIP` must be a valid IP address assigned to a server network interface
- `ListenAddress` must be a valid IP:port combination
- `Protocols` must contain at least one protocol
- `MaxConnections` must be > 0
- `ConnectionTimeout` must be > 0

---

### 2. Connection

**Description**: Represents an active proxied connection between a client and a destination through the proxy server.

**Attributes**:

| Attribute | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `ID` | `string` | Yes | Unique connection identifier | UUID format |
| `ClientAddr` | `net.Addr` | Yes | Client's address | - |
| `DestinationAddr` | `string` | Yes | Target destination address | Host:port format |
| `Protocol` | `Protocol` | Yes | Protocol used for this connection | SOCKS5 or HTTP |
| `MaskedSourceIP` | `net.IP` | Yes | Source IP presented to destination | Same as server's BindIP |
| `CreatedAt` | `time.Time` | Yes | Connection establishment time | - |
| `BytesReceived` | `uint64` | Yes | Bytes received from destination | >=0 |
| `BytesSent` | `uint64` | Yes | Bytes sent to destination | >=0 |
| `State` | `ConnectionState` | Yes | Current connection state | - |
| `Error` | `error` | No | Last error encountered | nil if no error |

**Go Struct**:
```go
type Connection struct {
    ID               string
    ClientAddr       net.Addr
    DestinationAddr  string
    Protocol         Protocol
    MaskedSourceIP   net.IP
    CreatedAt        time.Time
    BytesReceived    uint64
    BytesSent        uint64
    State            ConnectionState
    Error            error

    // Internal
    clientConn       net.Conn
    destConn         net.Conn
    mu               sync.RWMutex
}

type ConnectionState int

const (
    ConnectionStateEstablishing ConnectionState = iota
    ConnectionStateActive
    ConnectionStateClosing
    ConnectionStateClosed
    ConnectionStateError
)
```

**Relationships**:
- Belongs to one `ProxyServer` (many-to-one)

**Lifecycle**:
1. **Establishing** → Client connects, proxy negotiates protocol
2. **Active** → Bidirectional data relay in progress
3. **Closing** → Graceful shutdown initiated
4. **Closed** → Both connections closed, resources released
5. **Error** → Connection failed, cleanup in progress

**Validation Rules**:
- `ClientAddr` and `DestinationAddr` must be valid network addresses
- `MaskedSourceIP` must match the proxy server's `BindIP`
- `BytesReceived` and `BytesSent` must not overflow

---

### 3. ClientConfiguration

**Description**: Represents the client-side network redirection configuration state and settings for routing traffic through the proxy.

**Attributes**:

| Attribute | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `ProxyAddress` | `string` | Yes | Proxy server address:port | Valid host:port format |
| `Platform` | `Platform` | Yes | Operating system platform | Linux, Windows, or macOS |
| `Rules` | `[]NetworkRule` | No | Active network redirection rules | Platform-specific |
| `State` | `ConfigState` | Yes | Configuration state | Inactive or Active |
| `SetupAt` | `*time.Time` | No | When configuration was activated | nil if inactive |
| `OriginalState` | `*NetworkState` | No | Pre-setup network state for restoration | nil if never set up |

**Go Struct**:
```go
type ClientConfiguration struct {
    ProxyAddress   string
    Platform       Platform
    Rules          []NetworkRule
    State          ConfigState
    SetupAt        *time.Time
    OriginalState  *NetworkState
}

type Platform string

const (
    PlatformLinux   Platform = "linux"
    PlatformWindows Platform = "windows"
    PlatformMacOS   Platform = "darwin"
)

type ConfigState int

const (
    ConfigStateInactive ConfigState = iota
    ConfigStateActive
)
```

**Relationships**:
- Contains multiple `NetworkRule` instances (one-to-many)
- Has one `NetworkState` snapshot (one-to-one, optional)

**Lifecycle**:
1. **Inactive** → No rules configured
2. **Setup** → Privilege check → Add rules → Capture original state
3. **Active** → Traffic redirected through proxy
4. **Teardown** → Remove rules → Restore original state
5. **Inactive** → Clean state, ready for new setup

**Validation Rules**:
- `ProxyAddress` must be a valid reachable address
- Platform must be one of the supported platforms
- When `State` is Active, `Rules` must not be empty
- `OriginalState` must be captured before setup completes

---

### 4. NetworkRule

**Description**: Represents a single platform-specific network redirection rule that routes traffic to the proxy.

**Attributes**:

| Attribute | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `ID` | `string` | Yes | Unique rule identifier | Platform-specific format |
| `Platform` | `Platform` | Yes | Target platform | Linux, Windows, or macOS |
| `RuleType` | `RuleType` | Yes | Type of rule | REDIRECT, PORTPROXY, or PF |
| `Protocol` | `string` | Yes | Network protocol | "tcp", "udp", or "all" |
| `TargetPort` | `int` | No | Specific port to redirect (0=all) | 0-65535 |
| `ProxyPort` | `int` | Yes | Destination proxy port | 1-65535 |
| `Command` | `string` | Yes | Actual system command executed | Platform-specific |
| `RawOutput` | `string` | No | Command output for idempotency checks | - |

**Go Struct**:
```go
type NetworkRule struct {
    ID         string
    Platform   Platform
    RuleType   RuleType
    Protocol   string
    TargetPort int
    ProxyPort  int
    Command    string
    RawOutput  string
}

type RuleType string

const (
    RuleTypeIPTablesRedirect RuleType = "iptables_redirect"  // Linux
    RuleTypePortProxy        RuleType = "portproxy"          // Windows
    RuleTypePF               RuleType = "pf"                 // macOS
)
```

**Platform-Specific Examples**:

**Linux (iptables)**:
```go
NetworkRule{
    ID:         "gofakeip-tcp-80",
    Platform:   PlatformLinux,
    RuleType:   RuleTypeIPTablesRedirect,
    Protocol:   "tcp",
    TargetPort: 80,
    ProxyPort:  1080,
    Command:    "iptables -t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports 1080",
    RawOutput:  "",
}
```

**Windows (netsh)**:
```go
NetworkRule{
    ID:         "portproxy-80-1080",
    Platform:   PlatformWindows,
    RuleType:   RuleTypePortProxy,
    Protocol:   "tcp",
    TargetPort: 80,
    ProxyPort:  1080,
    Command:    "netsh interface portproxy add v4tov4 listenport=80 connectport=1080 connectaddress=127.0.0.1",
    RawOutput:  "",
}
```

**macOS (pfctl)**:
```go
NetworkRule{
    ID:         "gofakeip-pf-80",
    Platform:   PlatformMacOS,
    RuleType:   RuleTypePF,
    Protocol:   "tcp",
    TargetPort: 80,
    ProxyPort:  1080,
    Command:    "echo 'rdr pass on en0 inet proto tcp to any port 80 -> 127.0.0.1 port 1080' | pfctl -ef -",
    RawOutput:  "",
}
```

**Relationships**:
- Belongs to one `ClientConfiguration` (many-to-one)

**Validation Rules**:
- `Command` must pass input validation before execution (no injection)
- `TargetPort` and `ProxyPort` must be valid port numbers
- `Platform` must match the current operating system
- `Protocol` must be valid for the platform

---

### 5. AuthConfig

**Description**: Authentication configuration for proxy server access control.

**Attributes**:

| Attribute | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `Enabled` | `bool` | Yes | Whether authentication is required | - |
| `Method` | `AuthMethod` | Yes | Authentication method | UsernamePassword or None |
| `Username` | `string` | Conditional | Username for auth | Required if Enabled=true |
| `PasswordHash` | `string` | Conditional | Bcrypt hash of password | Required if Enabled=true, never plain text |

**Go Struct**:
```go
type AuthConfig struct {
    Enabled      bool
    Method       AuthMethod
    Username     string
    PasswordHash string  // bcrypt hash, NEVER plain text
}

type AuthMethod string

const (
    AuthMethodNone             AuthMethod = "none"
    AuthMethodUsernamePassword AuthMethod = "username_password"
)
```

**Security Requirements**:
- Passwords MUST be stored as bcrypt hashes (NON-NEGOTIABLE #4)
- Plain text passwords MUST NOT appear in config files or source code
- Username must be non-empty if authentication is enabled

---

### 6. HTTPHeaderConfig

**Description**: Configuration for HTTP header manipulation (distorting proxy functionality).

**Attributes**:

| Attribute | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `InjectXForwardedFor` | `string` | No | Value to inject in X-Forwarded-For header | Valid IP address string or empty |
| `InjectXRealIP` | `string` | No | Value to inject in X-Real-IP header | Valid IP address string or empty |
| `InjectVia` | `string` | No | Value to inject in Via header | Any string or empty |
| `RemoveExisting` | `bool` | Yes | Remove existing proxy headers before injection | - |

**Go Struct**:
```go
type HTTPHeaderConfig struct {
    InjectXForwardedFor string
    InjectXRealIP       string
    InjectVia           string
    RemoveExisting      bool
}
```

**Notes**:
- This is application-layer only (HTTP traffic)
- Does not affect network-layer source IP
- Optional feature for "distorting proxy" classification

---

### 7. NetworkState

**Description**: Snapshot of network configuration state for restoration during teardown.

**Attributes**:

| Attribute | Type | Required | Description | Constraints |
|-----------|------|----------|-------------|-------------|
| `Platform` | `Platform` | Yes | Operating system | - |
| `CapturedAt` | `time.Time` | Yes | When state was captured | - |
| `ExistingRules` | `[]string` | No | Existing rules before setup | Platform-specific format |
| `InterfaceStates` | `map[string]InterfaceState` | No | Network interface states | - |

**Go Struct**:
```go
type NetworkState struct {
    Platform        Platform
    CapturedAt      time.Time
    ExistingRules   []string
    InterfaceStates map[string]InterfaceState
}

type InterfaceState struct {
    Name       string
    Addresses  []net.IP
    IsUp       bool
    MTU        int
}
```

**Purpose**:
- Enables complete restoration during teardown (NON-NEGOTIABLE #2)
- Idempotency check (avoid duplicate rules)

---

## Entity Relationships Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                       ProxyServer                             │
│  - ListenAddress, BindIP, Protocols                          │
│  - MaxConnections, State                                     │
└───────────────┬────────────────┬─────────────────────────────┘
                │                │
                │ 1              │ 1
                │                │
        ┌───────▼──────┐   ┌────▼────────────┐
        │  AuthConfig  │   │ HTTPHeaderConfig│
        │  (optional)  │   │   (optional)    │
        └──────────────┘   └─────────────────┘
                │
                │ 1
                │
                │ *
        ┌───────▼──────────────────────────┐
        │         Connection               │
        │  - ID, ClientAddr, DestAddr      │
        │  - MaskedSourceIP, State         │
        └──────────────────────────────────┘


┌──────────────────────────────────────────────────────────────┐
│                  ClientConfiguration                          │
│  - ProxyAddress, Platform, State                             │
│  - SetupAt                                                   │
└───────────────┬─────────────────┬────────────────────────────┘
                │ 1               │ 1
                │                 │
                │ *               │ 0..1
        ┌───────▼──────┐    ┌────▼───────────┐
        │ NetworkRule  │    │  NetworkState  │
        │ (platform-   │    │  (snapshot)    │
        │  specific)   │    │                │
        └──────────────┘    └────────────────┘
```

---

## Data Validation & Sanitization

### Input Validation Strategy (NON-NEGOTIABLE #4)

All user inputs must pass through validation before use in system commands.

**IP Address Validation**:
```go
func ValidateIP(ip string) error {
    parsed := net.ParseIP(ip)
    if parsed == nil {
        return errors.New("invalid IP address format")
    }
    return nil
}
```

**Port Validation**:
```go
func ValidatePort(port int) error {
    if port < 1 || port > 65535 {
        return errors.New("port must be between 1 and 65535")
    }
    return nil
}
```

**Address Validation**:
```go
func ValidateAddress(addr string) error {
    host, port, err := net.SplitHostPort(addr)
    if err != nil {
        return fmt.Errorf("invalid address format: %w", err)
    }

    // Validate host (IP or hostname)
    if net.ParseIP(host) == nil {
        // If not IP, validate as hostname
        if !isValidHostname(host) {
            return errors.New("invalid hostname")
        }
    }

    // Validate port
    portNum, err := strconv.Atoi(port)
    if err != nil {
        return fmt.Errorf("invalid port: %w", err)
    }
    return ValidatePort(portNum)
}
```

**Command Argument Sanitization**:
```go
// ValidateCommandArg ensures no shell injection characters
func ValidateCommandArg(arg string) error {
    // Blacklist approach for shell metacharacters
    forbidden := []string{";", "&", "|", "`", "$", "(", ")", "<", ">", "\n", "\r"}
    for _, char := range forbidden {
        if strings.Contains(arg, char) {
            return fmt.Errorf("forbidden character '%s' in argument", char)
        }
    }

    // Length limit
    if len(arg) > 256 {
        return errors.New("argument too long (max 256 characters)")
    }

    return nil
}
```

---

## State Transition Rules

### ProxyServer State Transitions

```
Stopped ──[Start()]──> Starting ──[Success]──> Running
                           │
                           └──[Error]──> Error

Running ──[Stop()]──> Stopping ──> Stopped
    │
    └──[Critical Error]──> Error

Error ──[Restart()]──> Starting
```

**Invariants**:
- Cannot accept new connections unless in `Running` state
- Must gracefully close all connections before reaching `Stopped`
- Errors must be logged before transitioning to `Error` state

### Connection State Transitions

```
Establishing ──[Handshake Success]──> Active
     │
     └──[Handshake Failed]──> Error ──> Closed

Active ──[Close Request]──> Closing ──> Closed
   │
   └──[Connection Error]──> Error ──> Closed
```

**Invariants**:
- Data relay only occurs in `Active` state
- Resources must be released in `Closed` state
- Errors must be propagated to client before `Closed`

### ClientConfiguration State Transitions

```
Inactive ──[Setup()]──> [Privilege Check] ──[Pass]──> Active
                              │
                              └──[Fail]──> Inactive

Active ──[Teardown()]──> Inactive
```

**Invariants** (NON-NEGOTIABLE #2):
- Setup must be idempotent (running twice = same state as once)
- Teardown must restore original state completely
- Teardown must succeed even when run multiple times

---

## Persistence & Configuration

### Server Configuration File (YAML)

```yaml
# server-config.yaml
server:
  listen_address: "0.0.0.0:1080"
  bind_ip: "203.0.113.50"  # Must be assigned to server interface
  protocols:
    - socks5
    - http
  max_connections: 1000
  connection_timeout: 30s

auth:
  enabled: true
  method: username_password
  username: "admin"
  password_hash: "$2a$10$..." # bcrypt hash

http_headers:
  inject_x_forwarded_for: "198.51.100.25"
  inject_x_real_ip: "198.51.100.25"
  remove_existing: true
```

### Client Configuration File (YAML)

```yaml
# client-config.yaml
proxy:
  address: "203.0.113.50:1080"

# Platform detected automatically, stored for verification
platform: linux

# Rules are managed dynamically, not in config file
```

---

## Concurrency & Thread Safety

### Thread-Safe Operations

All entities with mutable state MUST use synchronization:

**ProxyServer**:
- `ActiveConnections` map: Protected by `sync.RWMutex`
- Connection count updates: Atomic operations

**Connection**:
- Byte counters: `atomic.AddUint64` for thread-safe increments
- State changes: Protected by `sync.RWMutex`

**ClientConfiguration**:
- Rule management: Serialized (only one setup/teardown at a time)

---

## Error Handling

### Error Types

```go
// Domain-specific errors
var (
    ErrInvalidIP           = errors.New("invalid IP address")
    ErrInvalidPort         = errors.New("invalid port number")
    ErrInvalidAddress      = errors.New("invalid network address")
    ErrCommandInjection    = errors.New("potential command injection detected")
    ErrInsufficientPrivs   = errors.New("insufficient privileges (root/admin required)")
    ErrPlatformUnsupported = errors.New("platform not supported")
    ErrRuleAlreadyExists   = errors.New("network rule already exists")
    ErrRuleNotFound        = errors.New("network rule not found")
    ErrConnectionFailed    = errors.New("connection to destination failed")
    ErrAuthFailed          = errors.New("authentication failed")
    ErrServerNotRunning    = errors.New("server not running")
)
```

---

## Summary

This data model provides:

1. **Clear entity definitions** with attributes, types, and constraints
2. **Validation rules** to enforce NON-NEGOTIABLE security requirements
3. **State machines** for predictable lifecycle management
4. **Thread safety** considerations for concurrent operations
5. **Platform-specific abstractions** for cross-platform support

All entities are designed to comply with the security constraints:
- ✅ Least privilege (privilege checks before system modifications)
- ✅ Idempotency & Reversibility (state snapshots, rule detection)
- ✅ Go-only implementation (no kernel modules)
- ✅ No vulnerable defaults (validation, sanitization, bcrypt hashing)

**Next Steps**: Use this data model to implement Go structs in `pkg/` and create validation functions in `pkg/common/validator.go`.
