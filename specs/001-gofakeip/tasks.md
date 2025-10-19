# Tasks: GoFakeIP - Distorting Anonymous Proxy

**Input**: Design documents from `/specs/001-gofakeip/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: Security and reversibility tests are MANDATORY per NON-NEGOTIABLE constraints in CLAUDE.md. Other tests are included to ensure quality.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

---

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions
GoFakeIP uses Go project structure:
- `cmd/` - Application entry points (server, client binaries)
- `pkg/` - Core libraries
- `tests/` - Test suites

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure based on research.md decisions

- [ ] T001 Create Go module with dependencies: github.com/things-go/go-socks5, golang.org/x/sys, gopkg.in/yaml.v3
- [ ] T002 [P] Create README.md with installation instructions and legal disclaimer
- [ ] T003 [P] Create .gitignore for Go projects (exclude binaries, secrets, state files)
- [ ] T004 [P] Create example configuration files in examples/ directory

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

**Foundational Components** (from data-model.md and plan.md):

- [X] T005 [P] Implement error definitions in pkg/common/errors.go per data-model.md
- [X] T006 [P] Implement logging utilities in pkg/common/logging.go with levels (debug, info, warn, error)
- [X] T007 [P] Implement input validator in pkg/netconfig/validator.go (ValidateIP, ValidatePort, ValidateAddress, ValidateCommandArg)
- [X] T008 [P] Implement privilege checker in pkg/netconfig/privileges.go (CheckPrivileges for Linux/Windows/macOS)
- [X] T009 Implement NetworkConfigurator interface in pkg/netconfig/netconfig.go per data-model.md
- [X] T010 [P] Implement safe command execution wrapper in pkg/netconfig/internal/command.go (NON-NEGOTIABLE #4: parameterized execution)
- [X] T011 [P] Implement configuration state tracker in pkg/netconfig/internal/state.go (for teardown restoration)
- [X] T012 Implement server configuration loader in pkg/config/server.go (YAML + flags with validation)
- [X] T013 Implement client configuration loader in pkg/config/client.go (YAML + flags with validation)

**Security Tests (MANDATORY - NON-NEGOTIABLE constraints)**:

- [ ] T014 [P] Create privilege enforcement test in tests/security/privilege_enforcement_test.go (verify privilege check before all system calls)
- [ ] T015 [P] Create input injection prevention test in tests/security/input_injection_test.go (test command injection patterns)
- [ ] T016 [P] Create no hardcoded secrets test in tests/security/no_hardcoded_secrets_test.go (scan source for secret patterns)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Deploy Proxy Server (Priority: P1) 🎯 MVP

**Goal**: A system administrator deploys the GoFakeIP proxy server on a remote machine to provide IP masking services. The server accepts SOCKS5 and HTTP proxy connections and masks the client's real IP with a configured fake IP.

**Independent Test**: Start server, configure SOCKS5 client, verify outbound connections show fake IP (bind IP) rather than client's real IP.

**Acceptance Criteria**:
1. Administrator starts server with specified fake IP → server listens on configured port
2. Client connects via SOCKS5 → connection established, traffic proxied with fake source IP
3. Client connects via HTTP proxy → connection established, traffic proxied with fake source IP
4. External services receive connection → they observe configured fake IP as source

### Implementation for User Story 1

**Core Proxy Components**:

- [ ] T017 [P] [US1] Implement Connection entity in pkg/proxy/connection.go per data-model.md (ID, ClientAddr, DestinationAddr, MaskedSourceIP, state tracking)
- [ ] T018 [P] [US1] Implement ProxyServer entity in pkg/proxy/server.go per data-model.md (ListenAddress, BindIP, Protocols, state management)
- [ ] T019 [US1] Implement connection pool manager in pkg/proxy/pool.go (concurrent connection handling, max connections enforcement)
- [ ] T020 [US1] Implement IP masking layer in pkg/proxy/masking.go (net.Dialer with LocalAddr binding per research.md)
- [ ] T021 [US1] Implement SOCKS5 handler in pkg/proxy/socks5.go using github.com/things-go/go-socks5 with custom dial function
- [ ] T022 [US1] Implement HTTP proxy handler in pkg/proxy/http.go (CONNECT method, GET/POST forwarding)
- [ ] T023 [US1] Implement server initialization in cmd/gofakeip-server/main.go (config loading, validation, server startup)
- [ ] T024 [US1] Add bind IP validation at startup in pkg/proxy/server.go (verify IP assigned to interface, fail early)
- [ ] T025 [US1] Implement graceful shutdown in pkg/proxy/server.go (SIGINT/SIGTERM handling, connection draining)

**Security & Validation**:

- [ ] T026 [US1] Add authentication support (optional) in pkg/proxy/auth.go (username/password with bcrypt, NON-NEGOTIABLE #4)
- [ ] T027 [US1] Add configuration validation in pkg/config/server.go (validate bind IP exists on interface)
- [ ] T028 [US1] Add error handling and logging throughout proxy components

**Tests for User Story 1**:

- [ ] T029 [P] [US1] Unit test for IP masking in tests/unit/proxy/ip_masking_test.go (verify LocalAddr binding)
- [ ] T030 [P] [US1] Unit test for SOCKS5 handler in tests/unit/proxy/socks5_handler_test.go (connection flow)
- [ ] T031 [P] [US1] Unit test for HTTP proxy handler in tests/unit/proxy/http_handler_test.go (CONNECT, GET/POST)
- [ ] T032 [US1] Integration test for proxy E2E in tests/integration/proxy_e2e_test.go (start server, connect client, verify IP masking)

**Checkpoint**: At this point, User Story 1 (proxy server) should be fully functional and testable independently

---

## Phase 4: User Story 2 - Configure Client on Linux (Priority: P2)

**Goal**: A user on a Linux machine uses the client utility to automatically configure their system to redirect all local traffic through the GoFakeIP proxy server using iptables/Netfilter rules. Configuration must be fully reversible.

**Independent Test**: Run client setup command on Linux, verify iptables rules are created, test that traffic is redirected through proxy, confirm teardown removes all rules.

**Acceptance Criteria**:
1. User runs setup without elevated privileges → utility refuses to run and prompts for sudo
2. User runs setup with sudo → iptables rules are created to redirect traffic
3. User accesses internet services → all traffic routes through proxy
4. User runs setup again → no duplicate rules created (idempotent)

### Implementation for User Story 2

**Linux Platform Implementation**:

- [ ] T033 [P] [US2] Implement LinuxConfigurator struct in pkg/netconfig/linux_iptables.go (implements NetworkConfigurator interface)
- [ ] T034 [US2] Implement Setup() for Linux in pkg/netconfig/linux_iptables.go (iptables REDIRECT rules with idempotency checks using -C flag)
- [ ] T035 [US2] Implement Teardown() for Linux in pkg/netconfig/linux_iptables.go (remove all GoFakeIP rules, restore original state)
- [ ] T036 [US2] Implement Status() for Linux in pkg/netconfig/linux_iptables.go (enumerate current rules)
- [ ] T037 [US2] Implement rule idempotency checking in pkg/netconfig/linux_iptables.go (check with -C before adding with -A)
- [ ] T038 [US2] Implement network state capture for Linux in pkg/netconfig/linux_iptables.go (save existing rules before setup)
- [ ] T039 [US2] Add iptables command validation in pkg/netconfig/linux_iptables.go (validate all args before exec.Command)

**Client CLI Implementation**:

- [ ] T040 [US2] Implement setup command in cmd/gofakeip-client/main.go (handleSetup function with flag parsing)
- [ ] T041 [US2] Implement teardown command in cmd/gofakeip-client/main.go (handleTeardown function)
- [ ] T042 [US2] Implement status command in cmd/gofakeip-client/main.go (handleStatus function with JSON output option)
- [ ] T043 [US2] Add platform detection in cmd/gofakeip-client/main.go (runtime.GOOS)
- [ ] T044 [US2] Add privilege check at client startup in cmd/gofakeip-client/main.go (fail fast if not root/admin)
- [ ] T045 [US2] Implement state persistence in pkg/netconfig/internal/state.go (save/load client state to JSON file)

**Tests for User Story 2 (Linux)**:

- [ ] T046 [P] [US2] Unit test for iptables command generation in tests/unit/netconfig/linux_iptables_test.go
- [ ] T047 [P] [US2] Idempotency test for Linux setup in tests/security/idempotency_test.go (run setup twice, verify same state)
- [ ] T048 [US2] Reversibility test for Linux teardown in tests/reversibility/linux_teardown_test.go (setup, teardown, verify original state restored)
- [ ] T049 [US2] Integration test for Linux client in tests/integration/client_setup_teardown_test.go (full setup/teardown cycle)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently (server + Linux client)

---

## Phase 5: User Story 3 - Configure Client on Windows (Priority: P2)

**Goal**: A user on a Windows machine uses the client utility to configure their system using Windows Filtering Platform (WFP) or netsh to redirect traffic through the proxy server.

**Independent Test**: Run client on Windows, verify WFP/netsh rules are created, confirm traffic redirection and clean teardown.

**Acceptance Criteria**:
1. User runs setup without Administrator privileges → utility refuses with error
2. User runs setup with Admin access → WFP/netsh rules created successfully
3. User accesses internet → traffic routes through proxy
4. User runs setup again → existing rules preserved without duplication

### Implementation for User Story 3

**Windows Platform Implementation**:

- [ ] T050 [P] [US3] Implement WindowsConfigurator struct in pkg/netconfig/windows_wfp.go (implements NetworkConfigurator interface)
- [ ] T051 [US3] Implement Setup() for Windows in pkg/netconfig/windows_wfp.go (netsh portproxy rules with idempotency checks)
- [ ] T052 [US3] Implement Teardown() for Windows in pkg/netconfig/windows_wfp.go (remove all GoFakeIP portproxy rules)
- [ ] T053 [US3] Implement Status() for Windows in pkg/netconfig/windows_wfp.go (query portproxy rules)
- [ ] T054 [US3] Implement rule idempotency checking for Windows in pkg/netconfig/windows_wfp.go (check existing rules with netsh show)
- [ ] T055 [US3] Implement network state capture for Windows in pkg/netconfig/windows_wfp.go (save existing portproxy rules)
- [ ] T056 [US3] Add Windows Admin check in pkg/netconfig/privileges.go (check token elevation via golang.org/x/sys/windows)
- [ ] T057 [US3] Add netsh command validation in pkg/netconfig/windows_wfp.go (validate all args before exec.Command)

**Tests for User Story 3 (Windows)**:

- [ ] T058 [P] [US3] Unit test for netsh command generation in tests/unit/netconfig/windows_wfp_test.go
- [ ] T059 [P] [US3] Idempotency test for Windows setup in tests/security/idempotency_test.go (Windows variant)
- [ ] T060 [US3] Reversibility test for Windows teardown in tests/reversibility/windows_teardown_test.go (setup, teardown, verify restoration)
- [ ] T061 [US3] Integration test for Windows client in tests/integration/client_setup_teardown_test.go (Windows variant)

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work independently (server + Linux client + Windows client)

---

## Phase 6: User Story 4 - Configure Client on macOS (Priority: P2)

**Goal**: A user on macOS uses the client utility to configure packet filter (pfctl) rules that redirect traffic through the proxy server.

**Independent Test**: Run setup on macOS, verify pfctl rules, test traffic redirection, confirm complete teardown.

**Acceptance Criteria**:
1. User runs setup without sudo → utility exits with error requesting elevated privileges
2. User runs setup with sudo → pfctl rules created to redirect traffic
3. User browses internet → connections route through proxy server

### Implementation for User Story 4

**macOS Platform Implementation**:

- [ ] T062 [P] [US4] Implement MacOSConfigurator struct in pkg/netconfig/macos_pf.go (implements NetworkConfigurator interface)
- [ ] T063 [US4] Implement Setup() for macOS in pkg/netconfig/macos_pf.go (pfctl rules via temp file or stdin with idempotency checks)
- [ ] T064 [US4] Implement Teardown() for macOS in pkg/netconfig/macos_pf.go (remove GoFakeIP anchor, restore original pf.conf)
- [ ] T065 [US4] Implement Status() for macOS in pkg/netconfig/macos_pf.go (pfctl -s rules query)
- [ ] T066 [US4] Implement rule idempotency checking for macOS in pkg/netconfig/macos_pf.go (check existing rules before loading)
- [ ] T067 [US4] Implement network state capture for macOS in pkg/netconfig/macos_pf.go (save current pf rules)
- [ ] T068 [US4] Add pfctl command validation in pkg/netconfig/macos_pf.go (validate all args before exec.Command)

**Tests for User Story 4 (macOS)**:

- [ ] T069 [P] [US4] Unit test for pfctl command generation in tests/unit/netconfig/macos_pf_test.go
- [ ] T070 [P] [US4] Idempotency test for macOS setup in tests/security/idempotency_test.go (macOS variant)
- [ ] T071 [US4] Reversibility test for macOS teardown in tests/reversibility/macos_teardown_test.go (setup, teardown, verify restoration)
- [ ] T072 [US4] Integration test for macOS client in tests/integration/client_setup_teardown_test.go (macOS variant)

**Checkpoint**: All cross-platform user stories complete - server + all client platforms working

---

## Phase 7: User Story 5 - Remove Client Configuration (Priority: P3)

**Goal**: A user wants to stop using the proxy and restore their machine to its original network configuration. The teardown process must remove all network rules completely and restore normal internet connectivity.

**Independent Test**: Run setup, then run teardown, verify all rules are removed and internet access works normally without proxy.

**Acceptance Criteria**:
1. Client configuration active on any platform → teardown removes all redirection rules completely
2. After teardown → connections work normally without routing through proxy
3. No client configuration exists → teardown succeeds without errors (graceful handling)
4. Teardown run multiple times → each execution succeeds without errors

### Implementation for User Story 5

**Note**: Teardown implementation is covered in User Stories 2, 3, 4 above. This phase adds cross-platform validation and edge case handling.

**Cross-Platform Teardown Enhancements**:

- [ ] T073 [P] [US5] Add graceful handling for teardown on clean system in pkg/netconfig/netconfig.go (return success if no rules found)
- [ ] T074 [P] [US5] Add multiple teardown execution safety in all platform implementations (ensure idempotent removal)
- [ ] T075 [US5] Add state file cleanup in pkg/netconfig/internal/state.go (remove state file after successful teardown)
- [ ] T076 [US5] Add teardown verification in cmd/gofakeip-client/main.go (check internet connectivity after teardown)

**Tests for User Story 5 (Teardown Validation)**:

- [ ] T077 [P] [US5] Test teardown on clean system in tests/reversibility/teardown_clean_system_test.go (verify graceful handling)
- [ ] T078 [P] [US5] Test multiple teardown executions in tests/reversibility/teardown_multiple_test.go (run teardown 3 times, verify all succeed)
- [ ] T079 [US5] Cross-platform teardown test in tests/integration/cross_platform_teardown_test.go (verify restoration on all platforms)

**Checkpoint**: Complete teardown functionality validated across all platforms

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories, security hardening, documentation

**Documentation**:

- [ ] T080 [P] Update README.md with build instructions, usage examples, legal disclaimer
- [ ] T081 [P] Create architecture.md in docs/ directory with system architecture details from plan.md
- [ ] T082 [P] Create security.md in docs/ documenting NON-NEGOTIABLE constraints and security model
- [ ] T083 [P] Create development.md in docs/ with development workflow and contribution guidelines

**Optional Features (from spec.md clarifications)**:

- [ ] T084 [P] Implement HTTP header injection in pkg/proxy/http.go (X-Forwarded-For, X-Real-IP distorting proxy feature per research.md)
- [ ] T085 [P] Add header injection configuration in pkg/config/server.go (http_headers section)
- [ ] T086 [P] Create server configuration examples in examples/server-config.yaml with comments

**Build & Distribution**:

- [ ] T087 [P] Create build script for cross-platform binaries in scripts/build.sh (Linux, Windows, macOS)
- [ ] T088 [P] Create systemd service file example in examples/systemd/gofakeip.service
- [ ] T089 [P] Create Windows service installation guide in docs/windows-service.md

**Final Security Audit**:

- [ ] T090 Run security test suite (privilege enforcement, input injection, no secrets)
- [ ] T091 Run reversibility test suite (all platforms, multiple executions, clean system)
- [ ] T092 Verify quickstart.md instructions work end-to-end
- [ ] T093 Run go vet and golint on all packages
- [ ] T094 Generate test coverage report (target >80% for critical paths per plan.md)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup (Phase 1) completion - BLOCKS all user stories
- **User Story 1 - Proxy Server (Phase 3)**: Depends on Foundational (Phase 2) completion
- **User Story 2 - Linux Client (Phase 4)**: Depends on Foundational (Phase 2) completion - can run in parallel with US1
- **User Story 3 - Windows Client (Phase 5)**: Depends on Foundational (Phase 2) completion - can run in parallel with US1/US2
- **User Story 4 - macOS Client (Phase 6)**: Depends on Foundational (Phase 2) completion - can run in parallel with US1/US2/US3
- **User Story 5 - Teardown (Phase 7)**: Depends on US2/US3/US4 (client implementations)
- **Polish (Phase 8)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - Independent (proxy server)
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent (Linux client, works with US1 server)
- **User Story 3 (P2)**: Can start after Foundational (Phase 2) - Independent (Windows client, works with US1 server)
- **User Story 4 (P2)**: Can start after Foundational (Phase 2) - Independent (macOS client, works with US1 server)
- **User Story 5 (P3)**: Depends on client implementations (US2/US3/US4) for teardown validation

### Within Each User Story

- **User Story 1 (Proxy Server)**:
  1. T017-T018 (entities) → T019-T020 (infrastructure) → T021-T022 (handlers) → T023 (main) → T024-T028 (validation/shutdown)
  2. Tests T029-T032 can run in parallel after implementation

- **User Story 2 (Linux Client)**:
  1. T033 (struct) → T034-T039 (implementation) → T040-T045 (CLI)
  2. Tests T046-T049 can run in parallel after implementation

- **User Story 3 (Windows Client)**:
  1. T050 (struct) → T051-T057 (implementation)
  2. Tests T058-T061 can run in parallel after implementation

- **User Story 4 (macOS Client)**:
  1. T062 (struct) → T063-T068 (implementation)
  2. Tests T069-T072 can run in parallel after implementation

### Parallel Opportunities

**Setup Phase (Phase 1)**: T002, T003, T004 can run in parallel

**Foundational Phase (Phase 2)**: T005, T006, T007, T008, T010, T011 can run in parallel; T014, T015, T016 (security tests) can run in parallel

**User Story 1**: T017, T018 can run in parallel; T029, T030, T031 can run in parallel

**User Story 2**: T046, T047 can run in parallel

**User Story 3**: T058, T059 can run in parallel

**User Story 4**: T069, T070 can run in parallel

**User Story 5**: T073, T074, T077, T078 can run in parallel

**Polish Phase**: T080, T081, T082, T083, T084, T085, T086, T087, T088, T089 can all run in parallel

**Cross-Story Parallelism**: After Foundational (Phase 2) completes, User Stories 1, 2, 3, 4 can all be developed in parallel by different team members.

---

## Parallel Example: User Story 1 (Proxy Server)

```bash
# Launch parallel entity creation:
Task: "Implement Connection entity in pkg/proxy/connection.go"
Task: "Implement ProxyServer entity in pkg/proxy/server.go"

# After entities complete, launch parallel tests:
Task: "Unit test for IP masking in tests/unit/proxy/ip_masking_test.go"
Task: "Unit test for SOCKS5 handler in tests/unit/proxy/socks5_handler_test.go"
Task: "Unit test for HTTP proxy handler in tests/unit/proxy/http_handler_test.go"
```

---

## Parallel Example: Cross-Platform Client Development

```bash
# After Foundational phase completes, launch all client platforms in parallel:
Developer A: User Story 2 (Linux Client) - T033 through T049
Developer B: User Story 3 (Windows Client) - T050 through T061
Developer C: User Story 4 (macOS Client) - T062 through T072

# All develop independently, all integrate with same User Story 1 server
```

---

## Implementation Strategy

### MVP First (Minimum Viable Product)

**Goal**: Demonstrate IP masking with server + one client platform

1. ✅ Complete Phase 1: Setup (T001-T004)
2. ✅ Complete Phase 2: Foundational (T005-T016) - **CRITICAL GATE**
3. ✅ Complete Phase 3: User Story 1 - Proxy Server (T017-T032)
4. ✅ Complete Phase 4: User Story 2 - Linux Client (T033-T049)
5. **STOP and VALIDATE**:
   - Deploy server on remote machine
   - Run client on Linux machine
   - Verify IP masking works (curl https://api.ipify.org shows server IP)
   - Test teardown restores original state
6. **Demo/Deploy MVP**: Functional proxy with Linux support

**MVP Scope**: User Stories 1 + 2 = Working proxy server + Linux client

---

### Incremental Delivery

1. **Foundation** (Phase 1-2): Setup + Foundational → Infrastructure ready
2. **MVP** (Phase 3-4): User Stories 1 + 2 → Test independently → Deploy/Demo (Linux support)
3. **Windows Support** (Phase 5): User Story 3 → Test independently → Deploy/Demo
4. **macOS Support** (Phase 6): User Story 4 → Test independently → Deploy/Demo
5. **Complete** (Phase 7-8): Teardown validation + Polish → Full cross-platform support

Each increment adds value without breaking previous functionality.

---

### Parallel Team Strategy

**With multiple developers**:

1. **Week 1**: Team completes Setup + Foundational together (T001-T016)
   - Critical: Security foundation, privilege checking, input validation

2. **Week 2-3**: After Foundational complete, split:
   - **Developer A**: User Story 1 - Proxy Server (T017-T032)
   - **Developer B**: User Story 2 - Linux Client (T033-T049)
   - **Developer C**: User Story 3 - Windows Client (T050-T061)
   - **Developer D**: User Story 4 - macOS Client (T062-T072)

3. **Week 4**: Integration and validation
   - Test all client platforms with server
   - Run security test suite (MANDATORY)
   - Run reversibility test suite (MANDATORY)
   - Complete User Story 5 teardown validation (T073-T079)

4. **Week 5**: Polish (T080-T094)
   - Documentation updates
   - Build scripts
   - Security audit
   - Performance testing

---

## Notes

- **[P] tasks**: Different files, no dependencies - can run in parallel
- **[Story] label**: Maps task to specific user story for traceability
- **Each user story**: Independently completable and testable
- **Security tests**: MANDATORY per NON-NEGOTIABLE constraints (T014-T016, T047-T048, T059-T060, T070-T071, T090-T091)
- **Reversibility tests**: MANDATORY per NON-NEGOTIABLE #2 (T048, T060, T071, T077-T079)
- **Privilege checks**: Must be first operation in all client commands (NON-NEGOTIABLE #1)
- **Input validation**: All user inputs validated before system commands (NON-NEGOTIABLE #4)
- **Commit strategy**: Commit after each task or logical group
- **Checkpoints**: Stop at any checkpoint to validate story independently
- **Avoid**: Vague tasks, same file conflicts, cross-story dependencies that break independence

---

## Task Count Summary

- **Phase 1 (Setup)**: 4 tasks
- **Phase 2 (Foundational)**: 12 tasks (includes 3 mandatory security tests)
- **Phase 3 (US1 - Proxy Server)**: 16 tasks
- **Phase 4 (US2 - Linux Client)**: 17 tasks
- **Phase 5 (US3 - Windows Client)**: 12 tasks
- **Phase 6 (US4 - macOS Client)**: 11 tasks
- **Phase 7 (US5 - Teardown)**: 7 tasks
- **Phase 8 (Polish)**: 15 tasks

**Total**: 94 tasks

**MVP Scope** (US1 + US2): 32 tasks (Setup + Foundational + Proxy Server + Linux Client)

**Parallel Opportunities**: 35+ tasks marked [P] can run in parallel within their phases

**Independent Test Criteria**: Each user story has clear acceptance criteria and independent test verification
