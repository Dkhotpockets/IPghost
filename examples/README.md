# GoFakeIP Configuration Examples

This directory contains example configuration files for both the server and client components.

## Server Configuration Examples

### server-config.yaml (Complete)
Full configuration with all available options, comments explaining each setting, and recommended production settings.

**Usage**:
```bash
gofakeip-server --config examples/server-config.yaml
```

**Features**:
- Authentication enabled (username/password with bcrypt)
- HTTP header injection (distorting proxy)
- JSON logging to file
- All options documented

---

### server-config-simple.yaml (Minimal)
Minimal configuration for quick start and development.

**Usage**:
```bash
gofakeip-server --config examples/server-config-simple.yaml
```

**Features**:
- Basic settings only
- Uses defaults for most options
- No authentication (for development only)
- Logs to stdout

---

### server-config-no-auth.yaml (Open Proxy)
Configuration for an open proxy (no authentication required).

**⚠️ WARNING**: Only use in controlled/trusted environments!

**Usage**:
```bash
gofakeip-server --config examples/server-config-no-auth.yaml
```

**Features**:
- No authentication (anyone can use)
- High max connections (10,000)
- JSON logging for monitoring

---

## Client Configuration Examples

### client-config.yaml (Specific Ports)
Redirect only HTTP/HTTPS traffic (ports 80, 443, 8080).

**Usage**:
```bash
# Linux/macOS
sudo gofakeip-client setup --config examples/client-config.yaml

# Windows (Administrator PowerShell)
gofakeip-client setup --config examples/client-config.yaml
```

**Features**:
- Specific port redirection
- TCP only
- Comments explaining each option

---

### client-config-all-traffic.yaml (All Ports)
Redirect ALL TCP traffic through the proxy.

**Usage**:
```bash
sudo gofakeip-client setup --config examples/client-config-all-traffic.yaml
```

**Features**:
- Redirects all ports
- Maximum privacy (all traffic through proxy)

---

## Before You Start

### 1. Update Server IP Addresses

**CRITICAL**: Replace `203.0.113.50` with your actual server IP in:
- All server config files (`bind_ip` field)
- All client config files (`proxy.address` field)

**Find your server's IP**:
```bash
# Linux
ip addr show

# macOS
ifconfig

# Windows
ipconfig
```

### 2. Generate Password Hash (Server)

If using authentication, generate a secure password hash:

```bash
# Using htpasswd (install: apt-get install apache2-utils)
htpasswd -nbB admin YOUR_PASSWORD

# Output format: admin:$2a$10$...
# Copy the hash (everything after "admin:") to password_hash field
```

**Never commit plain text passwords or real password hashes to version control!**

### 3. Set Secure File Permissions

**Linux/macOS**:
```bash
# Server config (contains password hash)
sudo chmod 600 /etc/gofakeip/server.yaml
sudo chown root:root /etc/gofakeip/server.yaml

# Client config
sudo chmod 644 /etc/gofakeip/client.yaml
```

**Windows**:
```powershell
# Restrict access to Administrators only
icacls C:\ProgramData\GoFakeIP\server.yaml /inheritance:r /grant:r "NT AUTHORITY\SYSTEM:(F)" "BUILTIN\Administrators:(F)"
```

---

## Testing Your Configuration

### Test Server Configuration

**Syntax check** (dry-run):
```bash
gofakeip-server --config examples/server-config.yaml --dry-run
```

**Start server**:
```bash
gofakeip-server --config examples/server-config.yaml
```

**Expected output**:
```
[INFO] Server starting on 0.0.0.0:1080
[INFO] Bind IP: 203.0.113.50
[INFO] Protocols: SOCKS5, HTTP
[INFO] Authentication: Enabled
[INFO] Server ready
```

### Test Client Configuration

**Dry-run** (show what would be executed):
```bash
sudo gofakeip-client setup --config examples/client-config.yaml --dry-run
```

**Setup**:
```bash
sudo gofakeip-client setup --config examples/client-config.yaml
```

**Verify** (check your public IP):
```bash
curl https://api.ipify.org
# Should show: 203.0.113.50 (your proxy server's IP)
```

**Teardown**:
```bash
sudo gofakeip-client teardown
```

---

## Common Configuration Patterns

### Development Setup

**Server**: Use `server-config-simple.yaml`
- No authentication (easier testing)
- Logs to stdout (easy to see)
- Local bind IP (127.0.0.1)

**Client**: Use command-line flags (faster iteration)
```bash
sudo gofakeip-client setup --proxy 127.0.0.1:1080 --ports 80,443
```

### Production Setup

**Server**: Use `server-config.yaml`
- Authentication enabled
- JSON logging to file
- Production bind IP
- Run as systemd service (see docs/quickstart.md)

**Client**: Use `client-config.yaml` with version control
- Configuration in version control (without secrets)
- Documented for team members
- Consistent across deployments

### High-Privacy Setup

**Server**: Use `server-config-no-auth.yaml` with firewall
- Restrict access with firewall rules (IP whitelist)
- High max connections
- Multiple bind IPs (if available)

**Client**: Use `client-config-all-traffic.yaml`
- Redirect all traffic (no leaks)
- DNS through proxy (if SOCKS5 with DNS support)

---

## Configuration Validation Checklist

Before deploying:

- [ ] **Server bind_ip** is assigned to a network interface
- [ ] **Client proxy.address** is reachable from client machine
- [ ] **Password hash** is bcrypt format (starts with `$2a$`, `$2b$`, or `$2y$`)
- [ ] **No plain text passwords** in configuration files
- [ ] **File permissions** are restrictive (600 for server config)
- [ ] **Firewall rules** allow proxy port (server: allow 1080, client: allow outbound)
- [ ] **Test connectivity**: `telnet SERVER_IP 1080` succeeds
- [ ] **Verify IP masking**: `curl https://api.ipify.org` shows server IP after client setup

---

## Troubleshooting

### Server Issues

**"bind IP not found on any interface"**
- Run `ip addr show` (Linux) or `ipconfig` (Windows) to list available IPs
- Update `bind_ip` to match an existing IP

**"address already in use"**
- Another process is using port 1080
- Find: `sudo lsof -i :1080` (Linux) or `netstat -ano | findstr :1080` (Windows)
- Change port in config or stop conflicting process

**"authentication failed"**
- Verify password hash is correct
- Regenerate hash: `htpasswd -nbB admin PASSWORD`
- Check client is sending correct credentials

### Client Issues

**"insufficient privileges"**
- Must run as root (Linux/macOS) or Administrator (Windows)
- Linux/macOS: Use `sudo gofakeip-client setup ...`
- Windows: Right-click PowerShell → "Run as Administrator"

**"rules already exist"**
- Run `sudo gofakeip-client teardown` first
- Or use `--force` flag to replace existing configuration

**"cannot reach proxy server"**
- Test connectivity: `telnet 203.0.113.50 1080` or `nc -zv 203.0.113.50 1080`
- Check server firewall allows incoming connections on port 1080
- Verify proxy server is running: `systemctl status gofakeip` (if using systemd)

---

## Security Best Practices

1. **Always use authentication** in production (except in fully trusted networks)
2. **Never commit real password hashes** to version control
3. **Use firewall rules** to restrict proxy access by source IP
4. **Rotate passwords** regularly
5. **Monitor logs** for unauthorized access attempts
6. **Use HTTPS** for management interfaces (future feature)
7. **Keep software updated** for security patches

---

## Next Steps

- Read full documentation: [specs/001-gofakeip/quickstart.md](../specs/001-gofakeip/quickstart.md)
- Review security model: [CLAUDE.md](../CLAUDE.md)
- Deploy as system service: See quickstart.md for systemd/Windows Service setup
- Set up monitoring: Configure JSON logging and integrate with your log aggregation system

---

## Getting Help

- **Documentation**: See `specs/001-gofakeip/` directory
- **Issues**: GitHub Issues (repository link TBD)
- **Security Issues**: Follow responsible disclosure (security contact TBD)

---

**Legal Disclaimer**: Use this software responsibly and in compliance with all applicable laws and terms of service. See main README.md for full legal disclaimer.
