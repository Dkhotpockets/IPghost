# CLI Contract: gofakeip-client

**Feature Branch**: `001-gofakeip`
**Created**: 2025-10-19
**Binary**: `gofakeip-client`
**Purpose**: Cross-platform network configuration utility for redirecting traffic through GoFakeIP proxy

---

## Command Synopsis

```
gofakeip-client setup [OPTIONS]
gofakeip-client teardown [OPTIONS]
gofakeip-client status [OPTIONS]
gofakeip-client --version
gofakeip-client --help
```

---

## Commands

### `setup`

**Description**: Configure system network rules to redirect traffic through the proxy server.

**Requires**: Elevated privileges (root/sudo on Linux/macOS, Administrator on Windows)

**Behavior**:
1. Verify elevated privileges (fail if insufficient)
2. Detect operating system platform
3. Validate proxy address
4. Capture current network state (for teardown)
5. Check for existing GoFakeIP rules (idempotency)
6. Create platform-specific redirection rules
7. Verify configuration

**Synopsis**:
```bash
gofakeip-client setup --proxy <address:port> [OPTIONS]
```

---

#### Setup Flags

##### `--proxy, -p`

**Type**: `string`
**Required**: Yes

**Description**: Proxy server address and port to redirect traffic to.

**Format**: `HOST:PORT` where HOST is IP or hostname

**Example**:
```bash
gofakeip-client setup --proxy 203.0.113.50:1080
gofakeip-client setup --proxy proxy.example.com:1080
```

**Validation**:
- Must be valid host:port format
- Port must be 1-65535
- Host must be valid IP or resolvable hostname

---

##### `--ports`

**Type**: `string` (comma-separated)
**Required**: No
**Default**: `all` (redirect all TCP traffic)

**Description**: Specific ports to redirect. Use `all` for all traffic.

**Format**: Comma-separated port numbers or ranges

**Example**:
```bash
gofakeip-client setup --proxy 203.0.113.50:1080 --ports 80,443
gofakeip-client setup --proxy 203.0.113.50:1080 --ports 80,443,8000-9000
gofakeip-client setup --proxy 203.0.113.50:1080 --ports all
```

**Validation**:
- Each port must be 1-65535
- Ranges must be valid (start < end)
- "all" cannot be combined with specific ports

---

##### `--protocol`

**Type**: `string`
**Required**: No
**Default**: `tcp`

**Description**: Network protocol to redirect.

**Valid Values**: `tcp`, `udp`, `both`

**Example**:
```bash
gofakeip-client setup --proxy 203.0.113.50:1080 --protocol tcp
gofakeip-client setup --proxy 203.0.113.50:1080 --protocol both
```

**Note**: SOCKS5 supports both TCP and UDP. HTTP proxy only supports TCP.

---

##### `--interface`

**Type**: `string`
**Required**: No
**Default**: Platform-specific (all interfaces on Linux/Windows, primary on macOS)

**Description**: Network interface to apply rules to.

**Example**:
```bash
gofakeip-client setup --proxy 203.0.113.50:1080 --interface eth0
gofakeip-client setup --proxy 203.0.113.50:1080 --interface en0  # macOS
```

**Platform-Specific**:
- **Linux**: Default applies to OUTPUT chain (all interfaces)
- **Windows**: Not applicable (portproxy is interface-agnostic)
- **macOS**: Required for pf rules (default: primary interface)

---

##### `--dry-run`

**Type**: Flag (boolean)
**Required**: No

**Description**: Show what commands would be executed without actually running them.

**Example**:
```bash
gofakeip-client setup --proxy 203.0.113.50:1080 --dry-run
```

**Output**:
```
[DRY RUN] Would execute:
  iptables -t nat -A OUTPUT -p tcp -j REDIRECT --to-ports 1080
[DRY RUN] No changes made
```

---

##### `--force`

**Type**: Flag (boolean)
**Required**: No

**Description**: Force setup even if existing GoFakeIP rules detected.

**Example**:
```bash
gofakeip-client setup --proxy 203.0.113.50:1080 --force
```

**Behavior**:
- Without `--force`: Error if rules already exist
- With `--force`: Remove existing rules, then create new ones

**Warning**: Use with caution. May disrupt active connections.

---

### `teardown`

**Description**: Remove all GoFakeIP network redirection rules and restore original network configuration.

**Requires**: Elevated privileges (root/sudo on Linux/macOS, Administrator on Windows)

**Behavior** (NON-NEGOTIABLE #2 - Reversibility):
1. Verify elevated privileges
2. Detect operating system platform
3. Enumerate GoFakeIP rules (if any)
4. Remove all rules
5. Restore original network state (from setup snapshot)
6. Verify clean state

**Synopsis**:
```bash
gofakeip-client teardown [OPTIONS]
```

---

#### Teardown Flags

##### `--dry-run`

**Type**: Flag (boolean)
**Required**: No

**Description**: Show what would be removed without making changes.

**Example**:
```bash
gofakeip-client teardown --dry-run
```

**Output**:
```
[DRY RUN] Would execute:
  iptables -t nat -D OUTPUT -p tcp -j REDIRECT --to-ports 1080
[DRY RUN] Would restore original network state
[DRY RUN] No changes made
```

---

##### `--all`

**Type**: Flag (boolean)
**Required**: No

**Description**: Remove all proxy-related rules, even if not created by GoFakeIP.

**Example**:
```bash
gofakeip-client teardown --all
```

**Warning**: May remove rules created by other tools. Use with caution.

---

### `status`

**Description**: Show current GoFakeIP configuration status and active rules.

**Requires**: No elevated privileges (read-only)

**Synopsis**:
```bash
gofakeip-client status [OPTIONS]
```

**Output**:
```
GoFakeIP Client Status

Platform: Linux (amd64)
Configuration: Active
Proxy: 203.0.113.50:1080
Setup at: 2025-10-19 14:30:00

Active Rules:
  [1] iptables -t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports 1080
  [2] iptables -t nat -A OUTPUT -p tcp --dport 443 -j REDIRECT --to-ports 1080

Traffic Redirected: TCP ports 80, 443
```

---

#### Status Flags

##### `--json`

**Type**: Flag (boolean)
**Required**: No

**Description**: Output status in JSON format (for scripting).

**Example**:
```bash
gofakeip-client status --json
```

**Output**:
```json
{
  "platform": "linux",
  "state": "active",
  "proxy_address": "203.0.113.50:1080",
  "setup_at": "2025-10-19T14:30:00Z",
  "rules": [
    {
      "id": "gofakeip-tcp-80",
      "command": "iptables -t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports 1080"
    },
    {
      "id": "gofakeip-tcp-443",
      "command": "iptables -t nat -A OUTPUT -p tcp --dport 443 -j REDIRECT --to-ports 1080"
    }
  ]
}
```

---

## Global Flags

### `--version, -v`

**Type**: Flag (boolean)

**Description**: Print version information and exit.

**Example**:
```bash
gofakeip-client --version
```

**Output**:
```
gofakeip-client version 1.0.0
Platform: linux/amd64
Built: 2025-10-19
```

---

### `--help, -h`

**Type**: Flag (boolean)

**Description**: Print help message and exit.

**Example**:
```bash
gofakeip-client --help
gofakeip-client setup --help
```

---

## Usage Examples

### Setup on Linux

```bash
sudo gofakeip-client setup --proxy 203.0.113.50:1080
```

Redirects all TCP traffic through the proxy.

---

### Setup Specific Ports (Windows)

```powershell
# Run as Administrator
gofakeip-client setup --proxy 203.0.113.50:1080 --ports 80,443
```

Redirects only HTTP and HTTPS traffic.

---

### Setup on macOS with Specific Interface

```bash
sudo gofakeip-client setup --proxy 203.0.113.50:1080 --interface en0
```

Applies pf rules to en0 interface.

---

### Teardown

```bash
sudo gofakeip-client teardown
```

Removes all GoFakeIP rules on any platform.

---

### Check Status

```bash
gofakeip-client status
```

No sudo required (read-only operation).

---

### Dry Run

```bash
sudo gofakeip-client setup --proxy 203.0.113.50:1080 --dry-run
```

See what would be executed without making changes.

---

## Platform-Specific Behavior

### Linux (iptables)

**Setup**:
```bash
sudo gofakeip-client setup --proxy 203.0.113.50:1080 --ports 80,443
```

**Executed Commands**:
```bash
iptables -t nat -C OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports 1080 || \
iptables -t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports 1080

iptables -t nat -C OUTPUT -p tcp --dport 443 -j REDIRECT --to-ports 1080 || \
iptables -t nat -A OUTPUT -p tcp --dport 443 -j REDIRECT --to-ports 1080
```

**Idempotency**: Uses `-C` (check) before `-A` (append)

**Teardown Commands**:
```bash
iptables -t nat -D OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports 1080
iptables -t nat -D OUTPUT -p tcp --dport 443 -j REDIRECT --to-ports 1080
```

---

### Windows (netsh)

**Setup**:
```powershell
gofakeip-client setup --proxy 203.0.113.50:1080 --ports 80,443
```

**Executed Commands**:
```powershell
netsh interface portproxy show v4tov4 | findstr "80 " || ^
netsh interface portproxy add v4tov4 listenport=80 connectport=1080 connectaddress=127.0.0.1

netsh interface portproxy show v4tov4 | findstr "443 " || ^
netsh interface portproxy add v4tov4 listenport=443 connectport=1080 connectaddress=127.0.0.1
```

**Idempotency**: Checks existing rules with `show` before adding

**Teardown Commands**:
```powershell
netsh interface portproxy delete v4tov4 listenport=80
netsh interface portproxy delete v4tov4 listenport=443
```

**Note**: Requires proxy server running locally (127.0.0.1) or adjust to remote proxy with additional netsh routing

---

### macOS (pfctl)

**Setup**:
```bash
sudo gofakeip-client setup --proxy 203.0.113.50:1080 --interface en0 --ports 80,443
```

**Executed Commands**:
```bash
# Create temporary pf rule file
cat > /tmp/gofakeip.pf.conf << 'EOF'
rdr pass on en0 inet proto tcp to any port 80 -> 127.0.0.1 port 1080
rdr pass on en0 inet proto tcp to any port 443 -> 127.0.0.1 port 1080
EOF

# Load rules
pfctl -ef /tmp/gofakeip.pf.conf
```

**Idempotency**: Check existing rules with `pfctl -s rules` before loading

**Teardown Commands**:
```bash
pfctl -F all -f /etc/pf.conf  # Restore original pf config
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Invalid arguments or configuration |
| 2 | Insufficient privileges (not root/admin) |
| 3 | Platform not supported |
| 4 | Network configuration error (command failed) |
| 5 | Rules already exist (setup without --force) |
| 6 | No rules found (teardown when nothing configured) |

---

## Error Handling & Messages

### Insufficient Privileges

**Command**:
```bash
gofakeip-client setup --proxy 203.0.113.50:1080
```

**Output** (non-root):
```
[ERROR] Insufficient privileges
This operation requires elevated privileges.

Linux/macOS: Run with sudo
  sudo gofakeip-client setup --proxy 203.0.113.50:1080

Windows: Run as Administrator
  Right-click -> "Run as Administrator"

Exit code: 2
```

---

### Invalid Proxy Address

**Command**:
```bash
sudo gofakeip-client setup --proxy invalid:999999
```

**Output**:
```
[ERROR] Invalid proxy address: invalid:999999
  Port 999999 out of range (1-65535)

Usage:
  gofakeip-client setup --proxy <HOST:PORT>

Exit code: 1
```

---

### Rules Already Exist

**Command**:
```bash
sudo gofakeip-client setup --proxy 203.0.113.50:1080
```

**Output** (if already configured):
```
[ERROR] GoFakeIP rules already exist
Configuration is already active for proxy: 203.0.113.50:1080

Options:
  1. Run 'gofakeip-client teardown' first
  2. Use --force to replace existing configuration

Exit code: 5
```

---

### Teardown on Clean System (Graceful)

**Command**:
```bash
sudo gofakeip-client teardown
```

**Output** (no rules):
```
[INFO] No GoFakeIP rules found
System is already in clean state
Exit code: 0
```

**Note**: Teardown succeeds gracefully (NON-NEGOTIABLE #2)

---

### Platform Not Supported

**Command**:
```bash
gofakeip-client setup --proxy 203.0.113.50:1080
```

**Output** (on FreeBSD, for example):
```
[ERROR] Platform not supported: freebsd
Supported platforms:
  - Linux (iptables)
  - Windows (netsh)
  - macOS (pfctl)

Exit code: 3
```

---

## Security & Safety

### Privilege Verification (NON-NEGOTIABLE #1)

**Implementation**:
```go
func checkPrivileges() error {
    if runtime.GOOS == "windows" {
        // Check Windows Admin token
        return checkWindowsAdmin()
    }
    // Unix-like: check UID
    if os.Getuid() != 0 {
        return ErrInsufficientPrivs
    }
    return nil
}
```

**Enforcement**:
- Privilege check is **first operation** before any system modification
- Zero system commands executed without privileges
- Clear error messages guide users to elevate

---

### Input Sanitization (NON-NEGOTIABLE #4)

**All user inputs validated before system commands**:

```go
func validateProxyAddress(addr string) error {
    // Validate format, no shell metacharacters
    if strings.ContainsAny(addr, ";&|`$()") {
        return ErrCommandInjection
    }
    // Validate as valid host:port
    return validateAddress(addr)
}
```

**No string concatenation for commands**:
```go
// GOOD: Parameterized execution
cmd := exec.Command("iptables", "-t", "nat", "-A", "OUTPUT", ...)

// BAD: String concatenation (NEVER DO THIS)
cmd := exec.Command("sh", "-c", fmt.Sprintf("iptables -t nat -A OUTPUT %s", userInput))
```

---

### Idempotency Enforcement (NON-NEGOTIABLE #2)

**Setup checks before creating rules**:

```go
func setupIdempotent(config ClientConfiguration) error {
    // 1. Check for existing rules
    existingRules, err := listGoFakeIPRules()
    if err != nil {
        return err
    }

    // 2. If rules exist and no --force, error
    if len(existingRules) > 0 && !config.Force {
        return ErrRulesAlreadyExist
    }

    // 3. Create only missing rules
    for _, rule := range config.DesiredRules {
        if !ruleExists(rule, existingRules) {
            if err := createRule(rule); err != nil {
                return err
            }
        }
    }
    return nil
}
```

---

### Reversibility Guarantee (NON-NEGOTIABLE #2)

**State capture before modifications**:

```go
func captureNetworkState() (*NetworkState, error) {
    state := &NetworkState{
        Platform: detectPlatform(),
        CapturedAt: time.Now(),
    }

    // Platform-specific: enumerate existing rules
    switch state.Platform {
    case PlatformLinux:
        state.ExistingRules = captureIPTablesRules()
    case PlatformWindows:
        state.ExistingRules = captureNetshRules()
    case PlatformMacOS:
        state.ExistingRules = capturePFRules()
    }

    return state, nil
}
```

**Restoration during teardown**:

```go
func teardown() error {
    // 1. Load original state
    originalState, err := loadOriginalState()

    // 2. Remove all GoFakeIP rules
    if err := removeAllGoFakeIPRules(); err != nil {
        return err
    }

    // 3. Restore original rules (if any)
    if originalState != nil {
        return restoreOriginalState(originalState)
    }

    return nil
}
```

---

## State Persistence

### Configuration Storage

**Location**:
- **Linux**: `/var/lib/gofakeip/client-state.json`
- **Windows**: `C:\ProgramData\GoFakeIP\client-state.json`
- **macOS**: `/var/lib/gofakeip/client-state.json`

**Format**:
```json
{
  "version": "1.0",
  "platform": "linux",
  "proxy_address": "203.0.113.50:1080",
  "setup_at": "2025-10-19T14:30:00Z",
  "rules": [
    {
      "id": "gofakeip-tcp-80",
      "platform": "linux",
      "rule_type": "iptables_redirect",
      "protocol": "tcp",
      "target_port": 80,
      "proxy_port": 1080,
      "command": "iptables -t nat -A OUTPUT -p tcp --dport 80 -j REDIRECT --to-ports 1080"
    }
  ],
  "original_state": {
    "captured_at": "2025-10-19T14:29:55Z",
    "existing_rules": []
  }
}
```

**Purpose**:
- Enables `status` command
- Supports idempotent setup
- Ensures reversible teardown

---

## Summary

The `gofakeip-client` CLI provides:

- ✅ **Privilege enforcement** (NON-NEGOTIABLE #1): Checks before all operations
- ✅ **Idempotent setup** (NON-NEGOTIABLE #2): No duplicate rules
- ✅ **Reversible teardown** (NON-NEGOTIABLE #2): Complete restoration
- ✅ **Input validation** (NON-NEGOTIABLE #4): No command injection
- ✅ **Cross-platform support**: Linux, Windows, macOS
- ✅ **Clear error messages**: Guides users to resolution
- ✅ **Dry-run mode**: Safe testing before execution

**See Also**:
- [cli-server.md](./cli-server.md) - Server CLI contract
- [config-schema.md](./config-schema.md) - Configuration format
- [../data-model.md](../data-model.md) - Data structures
