# GoFakeIP Project - AI Agent Context

## NON-NEGOTIABLE Security & Engineering Constraints

This file establishes mandatory constraints for all code generation and implementation decisions throughout the GoFakeIP project. These constraints must be verified at every phase of development.

### 1. Least Privilege Enforcement

**Requirement**: Generated code must explicitly check for required elevated privileges (e.g., `sudo` on Unix, Administrator on Windows), and refuse to run privileged configuration commands without them.

**Enforcement Rules**:
- No code may assume persistent root access
- All privilege checks must occur before any privileged operation
- Clear error messages must be displayed when insufficient privileges detected
- Operations must fail safely when privilege checks fail

**Validation**: Every function that executes privileged system commands must include a privilege verification step that executes before the command.

---

### 2. Idempotency and Reversibility

**Requirement**: Implement robust `setup()` and `teardown()` functions for OS configuration. The `teardown()` function must be fully reversible, guaranteed to remove all rules and restore the original network state across all target OS systems without exception. Repeated execution of `setup()` must not create duplicate, conflicting rules.

**Enforcement Rules**:
- Setup operations must detect existing rules and avoid creating duplicates
- Teardown operations must successfully remove all rules even if run multiple times
- Original network state must be preserved or documented before modifications
- All platform-specific implementations (Linux/Windows/macOS) must meet these requirements

**Validation**: Test suite must verify:
- Running setup twice produces identical state as running it once
- Running teardown fully restores pre-setup network configuration
- Teardown succeeds even when no rules exist (graceful handling)

---

### 3. Technology Stack Lock

**Requirement**: Core network processing and abstraction logic MUST be implemented in Golang.

**Enforcement Rules**:
- Proxy server implementation: Go only
- Client utility implementation: Go only
- OS system call wrappers: Go with exec.Command or CGo for native calls
- Shell scripts allowed only for build/deployment automation, not core logic

**Validation**: All source files in `src/`, `pkg/`, `cmd/` directories must be `.go` files (excluding build scripts).

---

### 4. NO VULNERABLE DEFAULTS

**Requirement**: Strictly prohibit hardcoding sensitive credentials, utilizing insecure API parameters, or passing unsanitized user inputs to system commands.

**Enforcement Rules**:
- No credentials, API keys, passwords in source code
- All user inputs must be validated and sanitized before use in system calls
- All command construction must use parameterized execution (not string concatenation)
- Default configurations must follow security best practices (principle of least exposure)

**Validation**: Code review checklist must verify:
- No string concatenation when building system commands
- All user inputs pass through validation functions
- No hardcoded secrets in version control
- Input validation includes bounds checking, type checking, and injection prevention

---

## Purpose

This document serves as persistent context for AI-assisted development, ensuring all generated code complies with project security requirements. Reference this file when:
- Generating new code
- Reviewing implementation plans
- Writing tests
- Evaluating design decisions

**Version**: 1.0
**Created**: 2025-10-18
**Status**: Active - Enforced across all development phases
