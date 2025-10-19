# QuickStart Guide: GoFakeIP

**Feature Branch**: `001-gofakeip`
**Created**: 2025-10-19
**Audience**: End users, system administrators
**Related**: [spec.md](./spec.md), [cli-server.md](./contracts/cli-server.md), [cli-client.md](./contracts/cli-client.md)

---

## What is GoFakeIP?

GoFakeIP is a cross-platform proxy system that masks your IP address by routing traffic through a remote proxy server. It consists of two components:

1. **Server**: A SOCKS5/HTTP proxy server that forwards your traffic with a different source IP
2. **Client**: A configuration utility that automatically redirects your system's traffic through the proxy

**Privacy Model**: Destination servers see the proxy server's IP address, not your real IP.

**Important**: This is a standard proxy system using legitimate IP addresses. It does NOT perform illegal IP spoofing. See [research.md](./research.md) for technical details.

---

## Table of Contents

1. [Installation](#installation)
2. [Server Setup](#server-setup)
3. [Client Setup](#client-setup)
4. [Verification](#verification)
5. [Teardown](#teardown)
6. [Troubleshooting](#troubleshooting)
7. [Advanced Configuration](#advanced-configuration)

---

## Installation

### Prerequisites

**Server**:
- Go 1.21 or higher (for building from source)
- Linux, Windows, or macOS
- At least one network interface with an assigned IP address

**Client**:
- Linux with iptables, Windows with netsh, or macOS with pfctl
- Administrative/root privileges
- Network connectivity to proxy server

---

### Download Pre-Built Binaries

```bash
# Linux (amd64)
curl -LO https://github.com/yourorg/gofakeip/releases/latest/download/gofakeip-linux-amd64.tar.gz
tar -xzf gofakeip-linux-amd64.tar.gz
sudo mv gofakeip-server gofakeip-client /usr/local/bin/

# macOS (arm64)
curl -LO https://github.com/yourorg/gofakeip/releases/latest/download/gofakeip-darwin-arm64.tar.gz
tar -xzf gofakeip-darwin-arm64.tar.gz
sudo mv gofakeip-server gofakeip-client /usr/local/bin/

# Windows (amd64)
# Download from: https://github.com/yourorg/gofakeip/releases/latest/download/gofakeip-windows-amd64.zip
# Extract to C:\Program Files\GoFakeIP\
```

---

### Build from Source

```bash
# Clone repository
git clone https://github.com/yourorg/gofakeip.git
cd gofakeip

# Build server
go build -o gofakeip-server ./cmd/gofakeip-server

# Build client
go build -o gofakeip-client ./cmd/gofakeip-client

# Install (Linux/macOS)
sudo mv gofakeip-server gofakeip-client /usr/local/bin/

# Or add to PATH (Windows)
# Move binaries to C:\Program Files\GoFakeIP\ and add to PATH
```

---

## Server Setup

The proxy server should be deployed on a remote machine (VPS, cloud instance, etc.) with a public IP address.

### Quick Start (No Authentication)

```bash
# Find your server's IP address
ip addr show  # Linux
ipconfig      # Windows

# Start the proxy server
gofakeip-server --listen 0.0.0.0:1080 --bind-ip <SERVER_IP>

# Example:
gofakeip-server --listen 0.0.0.0:1080 --bind-ip 203.0.113.50
```

**Output**:
```
[INFO] Server starting on 0.0.0.0:1080
[INFO] Bind IP: 203.0.113.50
[INFO] Protocols: SOCKS5, HTTP
[INFO] Authentication: Disabled
[INFO] Server ready
```

Server is now accepting connections on port 1080.

---

### Recommended: With Authentication

For security, enable authentication to prevent unauthorized use:

#### Step 1: Create Configuration File

```bash
sudo mkdir -p /etc/gofakeip
sudo nano /etc/gofakeip/server.yaml
```

#### Step 2: Add Configuration

```yaml
# /etc/gofakeip/server.yaml
server:
  listen_address: "0.0.0.0:1080"
  bind_ip: "203.0.113.50"  # Replace with your server's IP
  protocols:
    - socks5
    - http
  max_connections: 1000

auth:
  enabled: true
  method: username_password
  username: "admin"
  # Generate hash with: echo "mypassword" | gofakeip-server hash-password
  # Or use: htpasswd -nbB admin mypassword
  password_hash: "$2a$10$rZQk3pXx8kZJmG7Z5vY8AO0.8mVFqK4kV1YzXmQzLpYvEqZ7LqK5S"
```

#### Step 3: Set Secure Permissions

```bash
sudo chmod 600 /etc/gofakeip/server.yaml
sudo chown root:root /etc/gofakeip/server.yaml
```

#### Step 4: Start Server

```bash
gofakeip-server --config /etc/gofakeip/server.yaml
```

---

### Running as a System Service

#### Linux (systemd)

```bash
# Create service file
sudo nano /etc/systemd/system/gofakeip.service
```

```ini
[Unit]
Description=GoFakeIP Proxy Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/gofakeip-server --config /etc/gofakeip/server.yaml
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable gofakeip
sudo systemctl start gofakeip

# Check status
sudo systemctl status gofakeip
```

---

#### Windows (Service)

```powershell
# Using NSSM (Non-Sucking Service Manager)
# Download from: https://nssm.cc/download

nssm install GoFakeIP "C:\Program Files\GoFakeIP\gofakeip-server.exe"
nssm set GoFakeIP AppParameters "--config C:\ProgramData\GoFakeIP\server.yaml"
nssm set GoFakeIP AppDirectory "C:\Program Files\GoFakeIP"
nssm set GoFakeIP Start SERVICE_AUTO_START

# Start service
sc start GoFakeIP

# Check status
sc query GoFakeIP
```

---

## Client Setup

The client utility configures your local machine to redirect traffic through the proxy server.

**Important**: Requires root/Administrator privileges.

---

### Linux

#### Step 1: Configure Redirection

```bash
# Redirect all TCP traffic
sudo gofakeip-client setup --proxy 203.0.113.50:1080

# Or redirect specific ports
sudo gofakeip-client setup --proxy 203.0.113.50:1080 --ports 80,443
```

**Output**:
```
[INFO] Platform detected: Linux (iptables)
[INFO] Privilege check: OK (root)
[INFO] Proxy: 203.0.113.50:1080
[INFO] Creating iptables rules...
[INFO] Rule created: tcp/80 -> 1080
[INFO] Rule created: tcp/443 -> 1080
[INFO] Setup complete
[INFO] Run 'sudo gofakeip-client teardown' to remove configuration
```

#### Step 2: Verify Configuration

```bash
# Check status
gofakeip-client status
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
```

#### Step 3: Verify iptables Rules

```bash
sudo iptables -t nat -L OUTPUT -n --line-numbers
```

---

### Windows

#### Step 1: Open Administrator PowerShell

Right-click PowerShell → "Run as Administrator"

#### Step 2: Configure Redirection

```powershell
# Redirect all TCP traffic
gofakeip-client setup --proxy 203.0.113.50:1080

# Or redirect specific ports
gofakeip-client setup --proxy 203.0.113.50:1080 --ports 80,443
```

**Output**:
```
[INFO] Platform detected: Windows (netsh)
[INFO] Privilege check: OK (Administrator)
[INFO] Proxy: 203.0.113.50:1080
[INFO] Creating portproxy rules...
[INFO] Rule created: 80 -> 1080
[INFO] Rule created: 443 -> 1080
[INFO] Setup complete
```

#### Step 3: Verify Configuration

```powershell
gofakeip-client status
netsh interface portproxy show v4tov4
```

---

### macOS

#### Step 1: Configure Redirection

```bash
# Redirect all TCP traffic on primary interface
sudo gofakeip-client setup --proxy 203.0.113.50:1080

# Or specify interface and ports
sudo gofakeip-client setup --proxy 203.0.113.50:1080 --interface en0 --ports 80,443
```

**Output**:
```
[INFO] Platform detected: macOS (pfctl)
[INFO] Privilege check: OK (root)
[INFO] Proxy: 203.0.113.50:1080
[INFO] Creating pfctl rules...
[INFO] Rules loaded successfully
[INFO] Setup complete
```

#### Step 2: Verify Configuration

```bash
gofakeip-client status
sudo pfctl -s rules
```

---

## Verification

### Test 1: Check Your Public IP

Before setup:
```bash
curl https://api.ipify.org
# Output: <your real IP>
```

After setup:
```bash
curl https://api.ipify.org
# Output: 203.0.113.50 (proxy server's IP)
```

---

### Test 2: HTTP Header Check (if using HTTP proxy)

```bash
curl -v https://httpbin.org/headers
```

Look for X-Forwarded-For or similar headers if HTTP header injection is enabled.

---

### Test 3: DNS Leak Test

Visit: https://dnsleaktest.com/

Verify that your original IP is not leaked via DNS queries.

---

## Teardown

When you want to stop using the proxy and restore normal network settings:

### All Platforms

```bash
# Linux/macOS
sudo gofakeip-client teardown

# Windows (Administrator PowerShell)
gofakeip-client teardown
```

**Output**:
```
[INFO] Platform detected: Linux
[INFO] Removing GoFakeIP rules...
[INFO] Rule removed: tcp/80 -> 1080
[INFO] Rule removed: tcp/443 -> 1080
[INFO] Restoring original network state...
[INFO] Teardown complete
[INFO] Network configuration restored
```

**Verification**:
```bash
# Check public IP (should be your real IP again)
curl https://api.ipify.org
```

---

## Troubleshooting

### Server Issues

#### Problem: "bind: address already in use"

**Cause**: Another process is using port 1080.

**Solution**:
```bash
# Find process using the port
sudo lsof -i :1080  # Linux/macOS
netstat -ano | findstr :1080  # Windows

# Kill process or change port
gofakeip-server --listen 0.0.0.0:8080 --bind-ip <IP>
```

---

#### Problem: "bind IP not found on any interface"

**Cause**: Configured `bind_ip` is not assigned to any network interface.

**Solution**:
```bash
# List available IPs
ip addr show  # Linux
ipconfig      # Windows
ifconfig      # macOS

# Use a valid IP from the list
gofakeip-server --listen 0.0.0.0:1080 --bind-ip <VALID_IP>
```

---

### Client Issues

#### Problem: "Insufficient privileges"

**Cause**: Not running as root/Administrator.

**Solution**:
```bash
# Linux/macOS
sudo gofakeip-client setup --proxy <PROXY>

# Windows: Right-click PowerShell -> Run as Administrator
```

---

#### Problem: "Rules already exist"

**Cause**: GoFakeIP already configured on this machine.

**Solution**:
```bash
# Option 1: Teardown first
sudo gofakeip-client teardown
sudo gofakeip-client setup --proxy <NEW_PROXY>

# Option 2: Force reconfiguration
sudo gofakeip-client setup --proxy <PROXY> --force
```

---

#### Problem: "Platform not supported"

**Cause**: Running on unsupported OS (e.g., FreeBSD).

**Supported Platforms**: Linux, Windows, macOS only.

---

### Connectivity Issues

#### Problem: Cannot reach proxy server

**Test**:
```bash
# Test connectivity
telnet 203.0.113.50 1080
nc -zv 203.0.113.50 1080

# Check firewall
# Linux
sudo iptables -L INPUT -n | grep 1080

# Windows
netsh advfirewall firewall show rule name=all | findstr 1080
```

**Solution**: Open firewall port on server.

---

#### Problem: Traffic still shows original IP

**Possible Causes**:
1. Proxy server not running
2. Client redirection rules not created
3. Application bypassing proxy (using raw sockets)
4. DNS leak

**Debug**:
```bash
# Check client status
gofakeip-client status

# Check server logs
journalctl -u gofakeip -f  # Linux systemd

# Test specific application
curl --proxy socks5://203.0.113.50:1080 https://api.ipify.org
```

---

## Advanced Configuration

### HTTP Header Injection (Distorting Proxy)

Server config:
```yaml
http_headers:
  inject_x_forwarded_for: "198.51.100.25"
  inject_x_real_ip: "198.51.100.25"
  remove_existing: true
```

**Note**: This only affects HTTP headers, not network-layer IP.

---

### Multiple Bind IPs (Server with Multiple Interfaces)

If your server has multiple IP addresses, you can choose which one to use:

```bash
# List all IPs
ip addr show

# Use specific IP for outbound connections
gofakeip-server --listen 0.0.0.0:1080 --bind-ip 203.0.113.50
```

---

### Port-Specific Redirection

Redirect only specific ports (e.g., HTTP/HTTPS):

```bash
# Linux/macOS
sudo gofakeip-client setup --proxy 203.0.113.50:1080 --ports 80,443,8080

# Windows
gofakeip-client setup --proxy 203.0.113.50:1080 --ports 80,443
```

---

### Using Configuration Files

#### Server

```yaml
# /etc/gofakeip/server.yaml
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
  password_hash: "$2a$10$..."

logging:
  level: info
  format: json
  file: "/var/log/gofakeip/server.log"
```

Start:
```bash
gofakeip-server --config /etc/gofakeip/server.yaml
```

---

#### Client

```yaml
# /etc/gofakeip/client.yaml
proxy:
  address: "203.0.113.50:1080"

redirection:
  ports:
    - 80
    - 443
  protocol: tcp
```

Setup:
```bash
sudo gofakeip-client setup --config /etc/gofakeip/client.yaml
```

---

## Security Best Practices

### 1. Always Use Authentication

```yaml
auth:
  enabled: true
  method: username_password
  username: "admin"
  password_hash: "$2a$10$..."  # NEVER use plain text
```

---

### 2. Firewall Configuration

**Server**:
```bash
# Linux (UFW)
sudo ufw allow 1080/tcp
sudo ufw enable

# Linux (iptables)
sudo iptables -A INPUT -p tcp --dport 1080 -j ACCEPT

# Windows Firewall
netsh advfirewall firewall add rule name="GoFakeIP" dir=in action=allow protocol=TCP localport=1080
```

---

### 3. Secure Configuration Files

```bash
# Linux/macOS
sudo chmod 600 /etc/gofakeip/server.yaml
sudo chown root:root /etc/gofakeip/server.yaml

# Windows
icacls C:\ProgramData\GoFakeIP\server.yaml /inheritance:r /grant:r "NT AUTHORITY\SYSTEM:(F)"
```

---

### 4. Use TLS/VPN for Proxy Connection (Future Enhancement)

For additional security, tunnel proxy traffic through VPN or use TLS-wrapped SOCKS5 (future feature).

---

## Legal and Ethical Use

**Acceptable Uses**:
- Privacy protection (hiding your IP for privacy reasons)
- Testing your own systems
- Bypassing geographic restrictions (check ToS)

**Prohibited Uses**:
- Unauthorized access to systems
- Fraud or impersonation
- DDoS attacks
- Any illegal activity

**Disclaimer**: Users are responsible for ensuring their use complies with all applicable laws and terms of service.

---

## Performance Tuning

### Server

```yaml
server:
  max_connections: 10000  # Increase for high traffic
  connection_timeout: 120s  # Adjust based on use case

# Also increase system limits
# /etc/security/limits.conf
* soft nofile 65536
* hard nofile 65536
```

---

### Client

For high-throughput scenarios, consider:
1. Redirecting only necessary ports (not all traffic)
2. Using multiple proxy servers (load balancing - manual)
3. Optimizing MTU settings

---

## Next Steps

- **Production Deployment**: Set up as system service
- **Monitoring**: Enable JSON logging for log aggregation
- **Backup**: Document configuration for disaster recovery
- **Security**: Regular updates and security audits

---

## Getting Help

**Documentation**:
- [Full Specification](./spec.md)
- [CLI Reference - Server](./contracts/cli-server.md)
- [CLI Reference - Client](./contracts/cli-client.md)
- [Configuration Schema](./contracts/config-schema.md)

**Issues & Support**:
- GitHub Issues: https://github.com/yourorg/gofakeip/issues
- Security Issues: security@yourorg.com

---

## Summary

You've learned how to:
- ✅ Install GoFakeIP server and client
- ✅ Deploy a proxy server with authentication
- ✅ Configure client to redirect traffic through the proxy
- ✅ Verify IP masking is working
- ✅ Remove configuration and restore normal network settings
- ✅ Troubleshoot common issues

Your traffic is now routed through the proxy server, providing IP address masking for privacy and security.

**Remember**: Use responsibly and in compliance with all applicable laws and terms of service.
