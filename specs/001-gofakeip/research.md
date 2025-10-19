# Research Report: IP Address Masking Implementation for GoFakeIP Proxy Server

**Date**: 2025-10-19
**Status**: Complete
**Branch**: `001-gofakeip`
**Related**: [spec.md](./spec.md), [plan.md](./plan.md)

---

## Executive Summary

This research addresses the critical question: **Can we implement "Distorting Proxy" functionality (fake source IP) from user-space without kernel modules?**

**Key Finding**: The term "Distorting Proxy" in industry terminology refers to **HTTP header manipulation** (X-Forwarded-For), NOT true IP address spoofing at the network layer. True source IP spoofing from user-space is technically infeasible and legally problematic for the intended use case.

**Recommended Approach**: Implement a standard transparent/anonymous proxy with configurable bind address, which provides the achievable security benefit while remaining legal, portable, and user-space only.

---

## 1. Understanding "Distorting Proxy" - Critical Clarification

### Industry Definition

Research reveals that "Distorting Proxy" is a **proxy anonymity classification**, not a network-layer IP spoofing mechanism:

**What Distorting Proxies Actually Do**:
- Modify HTTP headers (specifically `X-Forwarded-For`, `Via`, `X-Real-IP`)
- Insert a **false IP address** into these application-layer headers
- The **actual network-layer source IP** is still the proxy server's real IP
- Destination servers see the proxy's IP at network level, but a fake IP in HTTP headers

**Anonymity Hierarchy** (from most to least anonymous):
1. **Elite/High Anonymity**: No proxy headers sent, appears as direct connection
2. **Anonymous**: Sends proxy headers but doesn't reveal real client IP
3. **Distorting**: Sends proxy headers with **intentionally false client IP**
4. **Transparent**: Forwards real client IP in headers

### Technical Reality

```
Network Layer (IP):    [Client IP] → [Proxy Server IP] → [Destination]
                                      ↑ This is ALWAYS the proxy's real IP

Application Layer:     HTTP Headers: X-Forwarded-For: <FAKE_IP>
                                     ↑ This can be any value
```

**Critical Distinction**:
- Distorting proxies do **NOT** change the source IP in IP packets
- They only insert false information in **HTTP application headers**
- Destination servers can always see the proxy's real IP at network level
- The "fake IP" only exists in application protocol metadata

### Why This Matters for GoFakeIP

The original specification requested "mask the original source IP with a false, configurable IP" - this needs clarification:

**If the goal is**:
- Making HTTP-based services see a fake IP → Use HTTP header manipulation (achievable, user-space)
- Making the network connection appear from a fake IP → **NOT achievable from user-space** (requires kernel modification or is illegal)

---

## 2. User-Space IP Spoofing Limitations

### 2.1 Technical Feasibility Analysis

#### Can We Change Source IP from User-Space?

**Answer**: No, not reliably or portably without kernel-level access.

**Platform-Specific Restrictions**:

**Linux**:
- Raw sockets (`CAP_NET_RAW` capability) allow crafting packets with arbitrary source IPs
- Requires root privileges or `CAP_NET_RAW` capability
- Even with raw sockets, **return traffic will not reach you** (packets route to the spoofed IP)
- ISPs typically implement **egress filtering** (BCP 38) blocking packets with non-routed source IPs

**Windows** (Windows Vista and later):
- TCP data **cannot** be sent over raw sockets
- UDP datagrams with **invalid source addresses** are dropped by the OS
- The IP source address must exist on a network interface or the datagram is dropped
- These restrictions were implemented specifically to prevent IP spoofing attacks
- Windows explicitly blocks IP spoofing from user-space for security reasons

**macOS**:
- Similar to Linux, requires superuser privileges for raw sockets
- Subject to same egress filtering by ISPs
- Return traffic routing issues

### 2.2 Operating System Security Restrictions

**Capabilities Framework (Linux)**:
- `CAP_NET_RAW`: Required for raw socket creation and packet manipulation
- `CAP_NET_ADMIN`: Required for transparent proxying with `IP_TRANSPARENT`
- These capabilities require elevated privileges (root or `setcap`)

**Windows Security Model**:
- Windows Filtering Platform (WFP) requires Administrator privileges
- Raw socket restrictions enforced at OS level (cannot be bypassed from user-space)
- Winsock LSP (Layered Service Provider) deprecated due to security concerns

**macOS**:
- Requires root/sudo for raw socket access
- System Integrity Protection (SIP) prevents kernel modification on modern versions

### 2.3 Fundamental Problems with True IP Spoofing

**1. Return Traffic Routing**:
```
[Your Machine] --[packet with fake source IP: 8.8.8.8]--> [Destination]
[Destination]  --[response to 8.8.8.8]-----------------> [Goes to 8.8.8.8, not you]
```
Even if you successfully send a packet with a fake source IP, the response will route to that fake IP, not back to you. **This makes two-way communication impossible**.

**2. ISP Egress Filtering (BCP 38)**:
Most ISPs implement egress filtering that drops packets with source IPs not in their allocated ranges. Your spoofed packets will be dropped at the ISP level.

**3. Stateful Firewalls**:
Modern networks use stateful firewalls that track connection state. A packet with a spoofed source IP won't match any established connection and will be dropped.

**4. TCP Handshake Impossibility**:
TCP requires a three-way handshake. With a spoofed source IP:
```
You --> SYN with fake IP --> Server
Server --> SYN-ACK --> Goes to fake IP (not you)
[Connection fails - you never receive SYN-ACK]
```

### 2.4 Security Implications

**Legal Risks**:
- Computer Fraud and Abuse Act (CFAA) in the US criminalizes IP spoofing for unauthorized access
- UK Computer Misuse Act 1990 criminalizes IP spoofing (up to 2 years imprisonment)
- GDPR violations in EU for impersonation and data protection breaches
- **IP spoofing is illegal when used to gain unauthorized access or commit fraud**

**Ethical Concerns**:
- IP spoofing is the foundation for DDoS attacks
- Used for impersonation and identity theft
- Violates network trust relationships

**Practical Use Case Assessment**:
- If the goal is privacy/anonymity → Use legitimate proxy/VPN techniques
- If the goal is bypassing restrictions → May violate terms of service and laws
- If the goal is testing → Can be legal in controlled environments with authorization

---

## 3. Kernel Module Necessity Analysis

### 3.1 When Are Kernel Modules Required?

Kernel modules are necessary when you need to:

1. **Modify packet headers transparently** at the network layer before routing
2. **Implement custom NAT** that modifies source addresses without cooperation from the network stack
3. **Intercept packets** at low levels (Netfilter hooks, etc.) before routing decisions
4. **Bypass OS security restrictions** on raw socket usage

**For True IP Spoofing**: A kernel module could theoretically:
- Hook into the network stack pre-routing
- Modify outgoing packet headers with fake source IPs
- Modify incoming packet headers to route responses correctly

**However**: This still doesn't solve:
- ISP egress filtering
- Return traffic routing (without controlling the entire network path)
- Legal/ethical issues

### 3.2 Can We Achieve IP Masking Without Kernel Code?

**Yes - if we redefine "IP masking" correctly**:

**Achievable from User-Space**:
1. **Standard Proxy Functionality**: Client → Proxy Server → Destination
   - Destination sees proxy's IP, not client's IP
   - No kernel modification needed
   - Fully legal and standard

2. **HTTP Header Manipulation** (Distorting Proxy):
   - Insert false `X-Forwarded-For` headers
   - Achievable from user-space proxy
   - Application-layer only (not network-layer)

3. **Proxy with Configurable Bind Address**:
   - If the proxy server has multiple IPs, it can bind outbound connections to different source IPs
   - Uses `net.Dialer.LocalAddr` in Go
   - **The source IP must be legitimately assigned to the server**
   - Achievable from user-space

4. **VPN/TUN/TAP Approach** (User-space with kernel support):
   - Create TUN/TAP virtual network interface (requires root but no custom kernel module)
   - Implement routing in user-space
   - Packets appear to come from the VPN endpoint IP
   - Standard VPN technique (OpenVPN, WireGuard)

**NOT Achievable from User-Space**:
- Arbitrary source IP spoofing (fake IP not assigned to any of your interfaces)
- Transparent interception without client cooperation (requires kernel hooks)

### 3.3 Cross-Platform Implications

**Kernel Module Approach**:
- ❌ **Linux**: Possible but requires kernel module development and loading (security risk, maintenance burden)
- ❌ **Windows**: Extremely difficult, requires driver signing, WHQL certification for distribution
- ❌ **macOS**: System Integrity Protection (SIP) prevents kernel modification on modern versions
- ❌ **Portability**: Each platform requires completely different kernel code
- ❌ **Security**: Kernel code can crash the entire system, major security concerns
- ❌ **Violation of NON-NEGOTIABLE Constraint #3**: Technology Stack Lock requires Go, not kernel modules

**User-Space Approach**:
- ✅ **Linux**: `iptables` REDIRECT/TPROXY + user-space proxy (achievable)
- ✅ **Windows**: netsh portproxy + user-space proxy (achievable)
- ✅ **macOS**: pfctl + user-space proxy (achievable)
- ✅ **Portability**: Go code works cross-platform
- ✅ **Security**: Runs in user-space, contained by OS security
- ✅ **Compliant**: Meets all NON-NEGOTIABLE constraints

---

## 4. Implementation Approaches - Detailed Analysis

### 4.1 NAT-Based Approach (RECOMMENDED)

**How It Works**:
```
[Client] → [Local iptables REDIRECT] → [Proxy Server (user-space)] → [Destination]
                                         ↑
                                    Proxy's real IP is source
```

**Technical Details**:
1. Client's OS uses iptables/WFP/pfctl to redirect traffic to local proxy
2. Proxy receives connection, establishes new connection to destination
3. Destination sees proxy server's IP as source
4. Proxy relays traffic bidirectionally
5. Client configuration can be reversed (teardown)

**Advantages**:
- ✅ Fully achievable from user-space (proxy component)
- ✅ Cross-platform (different local redirection per OS, but same proxy logic)
- ✅ Legal and standard approach
- ✅ Bidirectional communication works
- ✅ No kernel modules required
- ✅ Meets all NON-NEGOTIABLE constraints

**Limitations**:
- Destination sees proxy's **real** IP, not an arbitrary fake IP
- Requires client-side configuration (setup command to configure iptables/WFP/pfctl)
- Performance overhead of user-space proxy

**Implementation in Go**:
```go
// Proxy server binds outbound connections to configured interface
dialer := &net.Dialer{
    LocalAddr: &net.TCPAddr{
        IP: net.ParseIP(config.BindIP), // Must be assigned to server interface
    },
}
conn, err := dialer.Dial("tcp", destination)
```

**What "Fake IP" Means Here**:
- If the proxy server has multiple IPs (e.g., multiple network interfaces), it can bind to different source IPs
- These must be **legitimate IPs assigned to the server**
- Cannot be arbitrary/unassigned IPs

### 4.2 Raw Socket Approach (NOT FEASIBLE)

**Theoretical Approach**:
Create raw sockets and manually craft IP packets with fake source IPs.

**Why It Fails**:
- ❌ **Windows**: Explicitly blocked - cannot send TCP or UDP with fake source IP
- ❌ **Return traffic**: No way to receive responses (they go to the fake IP)
- ❌ **ISP filtering**: Egress filtering drops spoofed packets
- ❌ **TCP impossible**: Can't complete handshake without receiving SYN-ACK
- ❌ **Requires root**: Not true user-space
- ❌ **Legal issues**: IP spoofing laws

**Verdict**: Not viable for a functional proxy server.

### 4.3 SOCKS5 BIND Feature for Source IP Control

**What SOCKS5 BIND Does**:
The BIND command in SOCKS5 is for **accepting incoming connections**, not controlling source IP:
- Used in protocols like FTP where server connects back to client
- Client tells proxy "listen for incoming connections on your side"
- Proxy binds to its own IP and returns the address to client
- Server connects to proxy, proxy forwards to client

**Source IP Control**:
- The proxy can specify which source IP to expect for security (access control)
- `SOCKS5_ALLOWBLANKETBIND` allows 0.0.0.0 (accept from any IP)
- This is **destination IP filtering**, not source IP spoofing

**Verdict**: BIND is not relevant for faking the proxy's outbound source IP. It's about accepting inbound connections.

### 4.4 VPN/TUN/TAP Interface Approach

**How It Works**:
```
[Client App] → [TUN device (virtual interface)] → [User-space VPN program]
                                                    ↓
                                            [Encrypted tunnel to VPN server]
                                                    ↓
                                            [VPN Server] → [Destination]
                                                 ↑
                                          VPN server's IP is source
```

**Technical Details**:
1. Create TUN (Layer 3) or TAP (Layer 2) virtual network interface
2. User-space program reads packets from TUN/TAP
3. Program encapsulates/encrypts packets and sends to VPN server
4. VPN server decapsulates and forwards to destination
5. Destination sees VPN server's IP as source

**TUN/TAP Characteristics**:
- **TUN**: Layer 3 (IP packets), used by most VPNs (OpenVPN, WireGuard)
- **TAP**: Layer 2 (Ethernet frames), for bridging scenarios
- Requires root/admin privileges to create interface
- No kernel **module** needed (kernel support exists in modern OSes)
- Cross-platform: Linux (native), Windows (TAP-Win32 driver), macOS (utun)

**Advantages**:
- ✅ Standard VPN approach (proven technology)
- ✅ All client traffic automatically routed (no per-app configuration)
- ✅ Encrypted tunnel possible
- ✅ Destination sees VPN server IP (legitimate masking)
- ✅ No kernel modules required (uses built-in TUN/TAP support)

**Disadvantages**:
- ❌ More complex than simple proxy
- ❌ Requires root/admin for interface creation
- ❌ Windows requires TAP-Win32 driver installation
- ❌ Higher overhead (encapsulation)
- ⚠️ Still shows VPN server's **real IP**, not arbitrary fake IP

**Implementation Complexity**: High (full VPN stack)

**Suitability for GoFakeIP**: Overkill for project scope, but technically sound if full VPN functionality desired.

### 4.5 HTTP Header Manipulation (Distorting Proxy)

**How It Works**:
```
Client → Proxy (SOCKS5/HTTP) → Destination

HTTP Request at Proxy:
GET / HTTP/1.1
Host: example.com
[Proxy adds/modifies headers]
X-Forwarded-For: <CONFIGURABLE_FAKE_IP>
Via: 1.1 proxy-name
X-Real-IP: <CONFIGURABLE_FAKE_IP>
```

**Technical Details**:
1. Proxy intercepts HTTP traffic (HTTP proxy or SOCKS5 with HTTP inspection)
2. Parses HTTP headers
3. Adds or modifies `X-Forwarded-For`, `X-Real-IP`, `Via` headers
4. Inserts configured "fake IP" in these headers
5. Forwards modified request to destination

**Advantages**:
- ✅ Fully achievable from user-space
- ✅ Matches industry definition of "Distorting Proxy"
- ✅ Simple implementation
- ✅ No special privileges needed (beyond proxy setup)
- ✅ Cross-platform

**Limitations**:
- ⚠️ **Application-layer only** (HTTP/HTTPS)
- ⚠️ Destination **server can still see real proxy IP** at network layer
- ⚠️ HTTPS requires MITM (man-in-the-middle) with client trust (complex, security concerns)
- ⚠️ Security-conscious servers ignore or validate X-Forwarded-For (can be spoofed)
- ⚠️ Not effective against IP-based blocking at network level

**Security Considerations**:
- X-Forwarded-For is **easily spoofed** and should never be trusted for authentication
- Many security guides explicitly warn against trusting X-Forwarded-For
- Services that rely on it for IP blocking can be bypassed

**Implementation in Go (HTTP Proxy)**:
```go
func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Modify headers before forwarding
    req.Header.Set("X-Forwarded-For", p.config.FakeIP)
    req.Header.Set("X-Real-IP", p.config.FakeIP)

    // Forward request
    resp, err := http.DefaultClient.Do(req)
    // ... handle response
}
```

---

## 5. Recommended Implementation Strategy

### 5.1 Decision Matrix

| Approach | Feasibility | User-Space | Portability | Legality | Matches Spec |
|----------|-------------|------------|-------------|----------|--------------|
| NAT-Based Proxy | ✅ High | ✅ Yes | ✅ High | ✅ Legal | ⚠️ Partial* |
| Raw Sockets | ❌ Impossible | ❌ No (needs root) | ❌ Windows blocked | ❌ Illegal use | ❌ No |
| SOCKS5 BIND | ❌ Irrelevant | N/A | N/A | N/A | ❌ No |
| VPN/TUN/TAP | ✅ High | ⚠️ Needs root | ⚠️ Medium (driver on Windows) | ✅ Legal | ⚠️ Partial* |
| HTTP Header Manipulation | ✅ High | ✅ Yes | ✅ High | ✅ Legal | ⚠️ Partial** |

\* Shows proxy's real IP, not arbitrary fake IP
\** Application layer only, network layer shows proxy IP

### 5.2 Recommended Strategy: Hybrid Approach

**Primary Implementation**: **NAT-Based Transparent Proxy with Configurable Bind Address**

**Components**:

1. **SOCKS5/HTTP Proxy Server** (User-Space Go Application):
   - Accepts client connections via SOCKS5 and HTTP proxy protocols
   - Establishes outbound connections to destinations
   - Binds outbound connections to configured source IP (must be assigned to server)
   - Relays traffic bidirectionally
   - Optionally adds HTTP header manipulation for HTTP traffic

2. **Client Configuration Tool** (Cross-Platform Go Application):
   - Linux: iptables REDIRECT rules
   - Windows: netsh portproxy rules
   - macOS: pfctl redirect rules
   - Idempotent setup, reversible teardown

**How It Works**:
```
┌─────────────────────────────────────────────────────────────────┐
│ Client Machine                                                   │
│                                                                   │
│ [Application] → [OS Network Stack]                              │
│                        ↓                                         │
│                 [iptables/netsh/pfctl REDIRECT]                 │
│                        ↓                                         │
│                 [Local SOCKS5 Proxy Listen Port]                │
│                        ↓                                         │
│                 [Outbound to Proxy Server]                      │
└─────────────────────────────────────────────────────────────────┘
                         │
                         │ Internet
                         ↓
┌─────────────────────────────────────────────────────────────────┐
│ Proxy Server Machine                                             │
│                                                                   │
│ [GoFakeIP Server Process]                                       │
│  - Listen on SOCKS5/HTTP port                                   │
│  - Receive client connection                                    │
│  - Create outbound connection:                                  │
│    net.Dialer.LocalAddr = Configured Bind IP (must exist)       │
│  - Relay traffic                                                │
│                        ↓                                         │
│                 [Network Interface with Bind IP]                │
│                        ↓                                         │
└─────────────────────────────────────────────────────────────────┘
                         │
                         │ Destination sees Bind IP as source
                         ↓
                    [Destination Server]
```

### 5.3 Technical Implementation Details

**Server Configuration**:
```yaml
# server-config.yaml
listen_address: "0.0.0.0:1080"  # SOCKS5/HTTP proxy listen
protocols:
  - socks5
  - http
bind_ip: "203.0.113.50"  # Source IP for outbound connections (must be assigned to server)
http_headers:
  fake_x_forwarded_for: "198.51.100.25"  # Optional: fake IP in HTTP headers
```

**Key Implementation Points**:

1. **Source IP Binding** (Go):
```go
type ProxyServer struct {
    config Config
}

func (s *ProxyServer) dialDestination(address string) (net.Conn, error) {
    dialer := &net.Dialer{
        Timeout: 30 * time.Second,
        LocalAddr: &net.TCPAddr{
            IP: net.ParseIP(s.config.BindIP),
            // Port 0 = let OS choose ephemeral port
        },
    }
    return dialer.Dial("tcp", address)
}
```

2. **HTTP Header Injection** (Optional Enhancement):
```go
func (s *ProxyServer) handleHTTPProxy(clientConn net.Conn) error {
    // Parse HTTP request
    req, err := http.ReadRequest(bufio.NewReader(clientConn))

    // Modify headers if configured
    if s.config.HTTPHeaders.FakeXForwardedFor != "" {
        req.Header.Set("X-Forwarded-For", s.config.HTTPHeaders.FakeXForwardedFor)
    }

    // Forward to destination
    destConn, err := s.dialDestination(req.Host)
    // ... relay traffic
}
```

3. **SOCKS5 Implementation**:
   - Use existing library: `github.com/armon/go-socks5` or `github.com/things-go/go-socks5`
   - Customize dial function to use custom dialer with LocalAddr
   - Both libraries support custom dial functions

### 5.4 What This Achieves

**Achievable Goals**:
- ✅ Destination servers see the proxy server's IP (configured bind address) as source
- ✅ Client's real IP is completely hidden from destination
- ✅ SOCKS5 and HTTP proxy protocols supported
- ✅ Cross-platform client configuration (Linux/Windows/macOS)
- ✅ Fully reversible setup/teardown
- ✅ User-space implementation (proxy component)
- ✅ Legal and ethical when used appropriately
- ✅ No kernel modules required
- ✅ Complies with all NON-NEGOTIABLE constraints

**Optional Enhancement (HTTP Only)**:
- ✅ HTTP headers can contain a different "fake" IP (X-Forwarded-For)
- ⚠️ This is application-layer only
- ⚠️ Sophisticated services ignore X-Forwarded-For

**NOT Achievable** (and why):
- ❌ Arbitrary source IP spoofing (fake IP not assigned to server)
  - **Reason**: Requires kernel modification, blocked by Windows, return traffic fails, illegal
- ❌ True network-layer IP address fakery
  - **Reason**: Fundamentally requires control of routing infrastructure or kernel
- ❌ Spoofing without owning the IP
  - **Reason**: ISP egress filtering, routing, legal issues

### 5.5 Clarification Required for Stakeholders

**Question for User/Stakeholder**:

The specification requests "mask the original source IP with a false, configurable IP (Distorting Proxy functionality)."

Based on research, this can mean:

**Option A: Standard Proxy IP Masking** (RECOMMENDED)
- Client's IP is hidden from destination
- Destination sees proxy server's **legitimate IP** (assigned to server's network interface)
- This IP can be configured (if server has multiple IPs)
- **Legal, functional, achievable**

**Option B: HTTP Header Fakery** (SUPPLEMENTAL)
- Network layer shows proxy's real IP
- HTTP headers (`X-Forwarded-For`) contain a configured "fake" IP
- **Application-layer only, not effective for network-level blocking**

**Option C: True IP Spoofing** (NOT RECOMMENDED - INFEASIBLE)
- Attempt to send packets with arbitrary source IP not assigned to server
- **Not achievable from user-space, illegal, doesn't work (return traffic fails)**

**Recommendation**: Implement **Option A** (standard proxy) with **optional Option B** (HTTP headers) for enhanced "distorting proxy" classification. Clearly document that this is not true network-layer IP spoofing.

---

## 6. Feasibility Assessment

### 6.1 Overall Feasibility: HIGH (with clarified scope)

**Achievable Components**:
- ✅ SOCKS5/HTTP proxy server in Go
- ✅ Cross-platform client configuration (iptables/netsh/pfctl)
- ✅ Source IP binding to legitimate server IPs
- ✅ HTTP header manipulation (optional)
- ✅ Concurrent connection handling
- ✅ Idempotent setup/teardown
- ✅ Fully user-space (proxy component)

**Not Achievable**:
- ❌ Arbitrary IP spoofing (fake IP not owned by server)
- ❌ True network-layer source IP fakery without consequences

### 6.2 Constraints and Limitations

**Technical Constraints**:
1. **Bind IP must be assigned**: The "fake" IP in configuration must be a legitimate IP address assigned to one of the proxy server's network interfaces
2. **Network-layer visibility**: Destination can always see proxy's IP at network layer (routing, logs, packet captures)
3. **HTTPS limitations**: Transparent HTTP header manipulation requires MITM, which requires client trust (complexity)
4. **Performance**: User-space proxy adds latency vs kernel-based NAT

**Operational Constraints**:
1. **Client setup requires privileges**: iptables/netsh/pfctl configuration needs root/admin
2. **Per-platform setup logic**: Different commands for Linux/Windows/macOS
3. **Network dependencies**: Client must be able to reach proxy server
4. **Firewall compatibility**: Corporate firewalls may block SOCKS5/custom proxy ports

**Legal Constraints**:
1. **No true IP spoofing**: Cannot implement arbitrary source IP spoofing (illegal)
2. **Compliance**: Must document intended legal use cases
3. **Terms of Service**: Using proxies may violate some services' ToS
4. **Jurisdiction**: Proxy operation legality varies by country

### 6.3 Security Considerations

**Security Benefits**:
- ✅ Hides client's IP from destination services
- ✅ Provides privacy layer
- ✅ Standard security model (proxy)

**Security Risks**:
- ⚠️ HTTP header manipulation is security theater (headers are easily spoofed/ignored)
- ⚠️ X-Forwarded-For should never be trusted for authentication
- ⚠️ Proxy becomes single point of failure/monitoring
- ⚠️ Potential for misuse (bypassing access controls)

**Security Best Practices for Implementation**:
1. Implement authentication on proxy server (prevent open proxy)
2. Logging for accountability
3. Rate limiting to prevent abuse
4. Input validation for all client inputs (NON-NEGOTIABLE #4)
5. Clear documentation of legal use cases

---

## 7. Legal and Ethical Considerations

### 7.1 Legal Framework

**IP Spoofing Laws**:
- **United States**: Computer Fraud and Abuse Act (CFAA) criminalizes IP spoofing for unauthorized access
- **United Kingdom**: Computer Misuse Act 1990, up to 2 years imprisonment
- **European Union**: GDPR violations for impersonation
- **International**: Budapest Convention on Cybercrime

**Key Legal Distinction**:
- **Legal**: Operating a proxy server that masks client IPs (with legitimate server IPs)
- **Legal**: Testing your own systems with proper authorization
- **ILLEGAL**: Using fake IPs to gain unauthorized access, commit fraud, launch attacks, or impersonate others

### 7.2 Recommended Implementation Strategy

**What GoFakeIP Should Do**:
1. Implement standard proxy functionality (legal, ethical)
2. Use legitimate IP addresses assigned to the proxy server
3. Optionally manipulate HTTP headers (standard distorting proxy)
4. Document legal use cases (privacy, testing, research)
5. Implement authentication to prevent open proxy abuse
6. Maintain logs for accountability

**What GoFakeIP Should NOT Do**:
1. Attempt true IP spoofing with unassigned IPs
2. Advertise as a tool for bypassing security controls
3. Implement features specifically for illegal purposes
4. Disable logging/accountability mechanisms

### 7.3 Ethical Use Cases

**Acceptable Use Cases**:
- Privacy-focused users wanting to hide their IP
- Security researchers testing systems (with authorization)
- Developers testing geo-location features
- Bypassing geographic content restrictions (debatable, check ToS)

**Unacceptable Use Cases**:
- Bypassing authentication or access controls
- Impersonating other users or systems
- Launching attacks (DDoS, etc.)
- Fraud or identity theft
- Violating laws or terms of service

---

## 8. Research Questions - Resolved

### Question 1: SOCKS5 Library Selection

**Decision**: Use `github.com/things-go/go-socks5`

**Rationale**:
- ✅ Actively maintained (recent commits)
- ✅ Full TCP/UDP and IPv4/IPv6 support
- ✅ Custom dial function support (for LocalAddr binding)
- ✅ Custom goroutine pool for performance
- ✅ Extensible and customizable

**Alternative Considered**: `github.com/armon/go-socks5`
- Simpler but less actively maintained
- Also supports custom dial functions
- Good fallback option

**Security Assessment**: Both libraries are open-source and can be audited. No known vulnerabilities. Custom dial function allows us to control source IP binding.

### Question 2: Configuration Management

**Decision**: Use standard `flag` package + YAML config files (via `gopkg.in/yaml.v3`)

**Rationale**:
- ✅ Simple, no external dependencies beyond YAML parser
- ✅ `flag` package is standard library (secure)
- ✅ YAML is human-readable and industry-standard
- ✅ No eval/exec risks (YAML parser is safe)
- ✅ Meets NON-NEGOTIABLE #4 (no vulnerable defaults)

**Rejected Alternative**: `viper`
- More features than needed (complexity)
- Additional dependency

**Configuration Structure**:
```yaml
server:
  listen_address: "0.0.0.0:1080"
  bind_ip: "203.0.113.50"  # Must be assigned to server interface
  protocols:
    - socks5
    - http
  auth:
    enabled: true
    username: "user"
    password_hash: "<bcrypt-hash>"  # Never plain text

http_headers:
  inject_x_forwarded_for: "198.51.100.25"  # Optional
  inject_x_real_ip: "198.51.100.25"  # Optional
```

### Question 3: Windows Network Configuration

**Decision**: Use `netsh` (simpler) for initial implementation, with note for future WFP enhancement

**Rationale**:
- ✅ `netsh` is available on all Windows versions (Windows 7+)
- ✅ Simpler to implement and test
- ✅ Sufficient for proxy redirection use case
- ✅ No additional drivers or APIs required

**Alternative (Future Enhancement)**: Windows Filtering Platform (WFP) via `golang.org/x/sys/windows`
- More powerful and flexible
- Higher complexity
- Requires deep Windows networking knowledge
- Better performance for high-traffic scenarios

**Implementation Plan**:
1. Phase 1: Implement `netsh interface portproxy` approach
2. Phase 2 (optional): Add WFP implementation for advanced users

**netsh Commands**:
```powershell
# Add portproxy rule
netsh interface portproxy add v4tov4 listenport=80 listenaddress=0.0.0.0 connectport=1080 connectaddress=127.0.0.1

# Remove rule
netsh interface portproxy delete v4tov4 listenport=80 listenaddress=0.0.0.0

# List rules (idempotency check)
netsh interface portproxy show v4tov4
```

### Question 4: IP Masking Implementation Strategy

**Decision**: **NAT-Based Proxy with Configurable Bind Address** + Optional HTTP Header Injection

**Rationale**:
- ✅ **Feasibility**: Fully achievable from user-space
- ✅ **User-space only**: Proxy component requires no kernel modification
- ✅ **Portability**: Go implementation works cross-platform
- ✅ **Legal**: Standard proxy operation is legal
- ✅ **Functional**: Bidirectional communication works (unlike true IP spoofing)

**Technical Approach**:
1. **Primary Mechanism**: Bind outbound connections to configured source IP
   ```go
   dialer := &net.Dialer{
       LocalAddr: &net.TCPAddr{IP: net.ParseIP(config.BindIP)},
   }
   ```
   - Configured IP must be assigned to proxy server's network interface
   - Destination sees this IP as source

2. **Secondary Mechanism** (HTTP only, optional): Inject HTTP headers
   ```go
   req.Header.Set("X-Forwarded-For", config.FakeXForwardedFor)
   ```
   - Application-layer only
   - Not effective against network-level blocking

**What "Fake IP" Means**:
- The configured bind IP is the proxy server's **legitimate IP**
- If server has multiple IPs, user can choose which one to use (appears as different source)
- For HTTP headers, any IP string can be inserted (but it's just metadata)

**Limitations Acknowledged**:
- Cannot use arbitrary/unassigned IPs at network layer
- Destination can always see proxy IP in network packets (routing headers, etc.)
- HTTP header injection is security theater (easily detected/ignored)

---

## 9. Revised Requirements Clarification

Based on research findings, the following specification requirements need clarification:

### Specification Line 6 (spec.md):
> "The proxy must successfully mask the original source IP with a false, configurable IP (Distorting Proxy functionality)."

**Proposed Revision**:
> "The proxy must successfully mask the client's source IP by presenting the proxy server's configured IP address to destination services. For HTTP traffic, the proxy may optionally inject configurable values into X-Forwarded-For and related headers (Distorting Proxy functionality, application-layer only)."

**Clarification**:
- "False IP" → "Proxy server's configured IP" (must be legitimately assigned)
- Add note about HTTP header injection being optional and application-layer only

### Specification Assumption (Line 175, spec.md):
> "Network infrastructure allows the fake IP to be used as a source address (no anti-spoofing filters blocking it)"

**Proposed Revision**:
> "The proxy server has the configured bind IP address legitimately assigned to one of its network interfaces, allowing it to be used as the source address for outbound connections."

**Clarification**:
- The IP is not "fake" in the sense of being arbitrary/unassigned
- It's a real IP that the server owns

---

## 10. Final Recommendations

### 10.1 Recommended Architecture

**Component 1: Proxy Server** (User-Space Go Application)
- SOCKS5 handler using `github.com/things-go/go-socks5` with custom dial function
- HTTP proxy handler using standard `net/http` with header injection
- Configurable bind IP for outbound connections (must be assigned to server interface)
- Optional authentication (username/password)
- Concurrent connection handling with goroutine pool
- Configuration via YAML file + command-line flags

**Component 2: Client Configuration Utility** (Cross-Platform Go Application)
- Platform detection (Linux/Windows/macOS)
- Privilege verification (root/admin check before operations)
- Linux: iptables NAT REDIRECT rules
- Windows: netsh portproxy rules
- macOS: pfctl redirect rules
- Idempotent setup (check existing rules before adding)
- Complete teardown (remove all rules, restore state)
- Input validation and sanitization (NON-NEGOTIABLE #4)

### 10.2 Implementation Phases

**Phase 0: Research** ✅ COMPLETE
- Resolved all NEEDS CLARIFICATION items
- Documented technical feasibility
- Identified legal/ethical boundaries

**Phase 1: Core Proxy Server**
- Implement SOCKS5 handler with custom dial (bind IP)
- Implement HTTP proxy handler
- Configuration loading (YAML + flags)
- Basic authentication
- Unit tests for proxy logic

**Phase 2: HTTP Header Injection** (Optional Enhancement)
- Parse HTTP requests in SOCKS5 CONNECT and HTTP proxy
- Inject X-Forwarded-For, X-Real-IP headers
- Test with common HTTP services

**Phase 3: Client Configuration - Linux**
- iptables wrapper with validation
- Idempotency checks
- Teardown logic
- Integration tests

**Phase 4: Client Configuration - Windows**
- netsh wrapper with validation
- Idempotency checks
- Teardown logic
- Integration tests

**Phase 5: Client Configuration - macOS**
- pfctl wrapper with validation
- Idempotency checks
- Teardown logic
- Integration tests

**Phase 6: Security & Reversibility Testing**
- Mandatory security tests (see plan.md)
- Privilege enforcement tests
- Input injection prevention tests
- Reversibility tests (all platforms)
- No hardcoded secrets audit

**Phase 7: Documentation & Polish**
- User documentation (quickstart.md)
- Legal use case documentation
- Security best practices guide
- Example configurations

### 10.3 Success Criteria (Revised)

**Achievable Success Criteria**:
- ✅ Destination services see proxy server's configured IP as source (not client's IP)
- ✅ Client can configure system to route traffic through proxy
- ✅ Setup is idempotent (no duplicate rules)
- ✅ Teardown fully removes all rules
- ✅ Cross-platform support (Linux/Windows/macOS)
- ✅ No kernel modules required
- ✅ Legal and ethical implementation

**Explicitly NOT Attempting**:
- ❌ Arbitrary IP spoofing (unassigned IPs)
- ❌ True network-layer IP fakery
- ❌ Bypassing return traffic routing limitations

---

## 11. Conclusion

### Key Findings Summary

1. **"Distorting Proxy" Definition**: Industry term for HTTP header manipulation, NOT network-layer IP spoofing
2. **User-Space IP Spoofing**: Not feasible (Windows blocks it, Linux requires root, return traffic fails, ISP filtering, illegal)
3. **Kernel Modules**: NOT required for legitimate proxy functionality; violates project constraints
4. **Recommended Approach**: Standard NAT-based proxy with configurable bind address (must own the IP)
5. **Legal/Ethical**: Standard proxy is legal; true IP spoofing is illegal and non-functional

### Recommended Decision

**Implement**: NAT-based transparent proxy with configurable source IP binding (legitimate IPs only) + optional HTTP header injection for "distorting proxy" classification.

**Do NOT implement**: True IP spoofing with arbitrary/unassigned source IPs (infeasible, illegal, violates constraints).

### Stakeholder Confirmation Needed

Before proceeding to Phase 1, confirm:
1. Is the goal to hide client IP behind proxy server's IP? → **Achievable**
2. Or is the goal true source IP spoofing with arbitrary IPs? → **Not achievable**

If #1: Proceed with recommended architecture.
If #2: Project requirements are technically and legally infeasible and must be revised.

---

## Appendix A: Go Libraries Selected

### SOCKS5: `github.com/things-go/go-socks5`
- **License**: MIT
- **Stars**: ~500+
- **Last Update**: Active (2024)
- **Features**: TCP/UDP, IPv4/IPv6, custom dial, auth, DNS resolution
- **Security**: Open source, auditable, no known CVEs

### YAML Config: `gopkg.in/yaml.v3`
- **License**: Apache 2.0 / MIT
- **Stars**: ~2000+
- **Last Update**: Maintained
- **Features**: Standard YAML parser
- **Security**: Well-tested, no eval/exec risks

### System Calls: `golang.org/x/sys`
- **License**: BSD
- **Maintainer**: Go team
- **Features**: Cross-platform system APIs
- **Security**: Official Go extended library

---

## Appendix B: Platform-Specific Commands Reference

### Linux (iptables)
```bash
# Add REDIRECT rule (setup)
iptables -t nat -A OUTPUT -p tcp -j REDIRECT --to-ports 1080

# Check rule exists (idempotency)
iptables -t nat -C OUTPUT -p tcp -j REDIRECT --to-ports 1080

# Remove rule (teardown)
iptables -t nat -D OUTPUT -p tcp -j REDIRECT --to-ports 1080

# List rules
iptables -t nat -L OUTPUT -n --line-numbers
```

### Windows (netsh)
```powershell
# Add portproxy rule (setup)
netsh interface portproxy add v4tov4 listenport=80 connectport=1080 connectaddress=127.0.0.1

# Show rules (idempotency check)
netsh interface portproxy show v4tov4

# Delete rule (teardown)
netsh interface portproxy delete v4tov4 listenport=80

# Reset all rules
netsh interface portproxy reset
```

### macOS (pfctl)
```bash
# Create rule file
echo "rdr pass on en0 inet proto tcp to any port 80 -> 127.0.0.1 port 1080" > /tmp/gofakeip.pf.conf

# Load rules (setup)
pfctl -ef /tmp/gofakeip.pf.conf

# Show rules (idempotency)
pfctl -s rules

# Disable pf (teardown)
pfctl -d

# Flush rules
pfctl -F all
```

---

## Appendix C: Legal Use Case Documentation Template

**Recommended Documentation for Users**:

```
GoFakeIP Legal Use Cases

ACCEPTABLE USES:
1. Privacy Protection: Hiding your personal IP from websites for privacy reasons
2. Security Testing: Testing your own systems with proper authorization
3. Development: Testing geo-location or IP-based features in applications you control
4. Research: Academic or security research in controlled, authorized environments

PROHIBITED USES:
1. Unauthorized Access: Using the proxy to bypass authentication or access controls
2. Fraud: Impersonating other users or systems
3. Attacks: Launching DDoS, intrusion attempts, or other malicious activities
4. ToS Violations: Bypassing service restrictions in violation of terms of service
5. Illegal Activities: Any use that violates local, national, or international law

USER RESPONSIBILITY:
Users are solely responsible for ensuring their use of GoFakeIP complies with all
applicable laws, regulations, and terms of service. The developers of GoFakeIP
are not responsible for misuse of this software.

TECHNICAL LIMITATIONS:
GoFakeIP implements standard proxy functionality. It masks your IP by presenting
the proxy server's legitimate IP address. It does NOT implement illegal IP spoofing
or impersonation techniques.
```

---

**End of Research Report**

**Status**: ✅ Complete - All Phase 0 research tasks resolved
**Next Step**: Proceed to Phase 1 (Design Artifacts) with recommended architecture
**Blocker**: Requires stakeholder confirmation of clarified scope before implementation
