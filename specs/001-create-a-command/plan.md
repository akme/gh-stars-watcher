
# Implementation Plan: GitHub Stars Monitor CLI

**Branch**: `001-create-a-command` | **Date**: 2025-09-29 | **Spec**: [spec.md](./spec.md)  
**Status**: ✅ **COMPLETED** - All phases implemented and validated | **Completion**: 2025-09-29  
**Input**: Feature specification from `/specs/001-create-a-command/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from file system structure or context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
Create a Go CLI application that monitors GitHub user starred repositories and displays only newly starred repositories between runs. The tool fetches current starred repos via GitHub API, persists state locally for comparison, and outputs changes with configurable formats and authentication methods. Supports progress indicators, configurable logging, and cleanup commands for state management.

## Technical Context
**Language/Version**: Go 1.25+ (latest stable features, generics support)  
**Primary Dependencies**: GitHub API client library, CLI framework (cobra), HTTP client with retry logic  
**Storage**: Local JSON files for state persistence, OS keychain integration for secure token storage  
**Testing**: Go standard testing, table-driven tests, integration tests with GitHub API mocking  
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: Single CLI application with clean architecture patterns  
**Performance Goals**: <500ms startup for cached operations, <2s for fresh data fetches, progress indicators >2s operations  
**Constraints**: <50MB memory for typical scenarios (1000+ repos), rate limit compliance, TLS 1.2+ required  
**Scale/Scope**: Handle unlimited repositories via pagination, support automated execution via cron/scheduled tasks

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**I. Code Quality Standards**: ✅ PASS
- Go 1.25+ supports idiomatic practices from `.github/instructions/go.instructions.md`
- Clean architecture with interface abstractions enables clear documentation
- Single-purpose packages and functions align with complexity limits

**II. Testing Excellence (NON-NEGOTIABLE)**: ✅ PASS
- TDD workflow: contract tests → integration tests → implementation
- GitHub API interactions, rate limiting, error scenarios covered in test plan
- 90%+ coverage target with meaningful assertions, not just execution coverage

**III. User Experience Consistency**: ✅ PASS
- CLI interface with consistent flag patterns via cobra framework
- JSON + human-readable output formats as specified
- Actionable error messages and progress indicators for >2s operations
- Graceful network failure handling per constitutional requirements

**IV. Performance Requirements**: ✅ PASS
- Rate limit compliance with exponential backoff design
- Local caching to minimize redundant API calls
- <50MB memory target, <500ms cached/<2s fresh startup targets
- Performance metrics validation in automated tests

**Initial Constitution Check**: PASS - No violations detected

**Post-Design Constitution Check**: ✅ PASS
- **Code Quality Standards**: Clean architecture with interface abstractions supports idiomatic Go and clear documentation
- **Testing Excellence**: Contract tests, integration tests, and unit tests provide comprehensive TDD coverage  
- **User Experience Consistency**: CLI design with cobra framework ensures consistent flag patterns and help text
- **Performance Requirements**: Architecture supports rate limiting, caching, and progress indication requirements

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
cmd/
└── star-watcher/
    └── main.go              # CLI entry point

internal/
├── github/
│   ├── client.go           # GitHub API client interface
│   ├── api.go              # API implementation
│   └── models.go           # GitHub API response models
├── storage/
│   ├── state.go            # State persistence interface
│   ├── json.go             # JSON file storage implementation
│   └── models.go           # Local state models
├── auth/
│   ├── token.go            # Token management interface
│   ├── keychain.go         # OS keychain integration
│   └── prompt.go           # Interactive token prompt
├── cli/
│   ├── root.go             # Root command setup
│   ├── monitor.go          # Main monitoring command
│   ├── cleanup.go          # State cleanup command
│   └── output.go           # Output formatting (JSON/text)
└── monitor/
    ├── service.go          # Core monitoring business logic
    ├── differ.go           # Repository comparison logic
    └── progress.go         # Progress indication

tests/
├── integration/
│   ├── github_api_test.go  # Real API integration tests
│   ├── monitor_flow_test.go # End-to-end workflow tests
│   └── auth_flow_test.go   # Authentication flow tests
├── contract/
│   ├── github_client_test.go # GitHub client contract tests
│   ├── state_storage_test.go # Storage contract tests
│   └── auth_token_test.go   # Auth contract tests
└── unit/
    ├── differ_test.go      # Repository comparison unit tests
    ├── output_test.go      # Output formatting unit tests
    └── progress_test.go    # Progress indication unit tests

go.mod
go.sum
README.md
```

**Structure Decision**: Single Go CLI application using clean architecture with clear separation between external interfaces (github, storage, auth) and business logic (monitor). The `internal/` package prevents external imports while `cmd/` provides the CLI entry point. Test organization follows constitutional TDD requirements with contract, integration, and unit test separation.

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/bash/update-agent-context.sh copilot`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

**Task Generation Strategy**:
- Load `.specify/templates/tasks-template.md` as base
- Generate contract tests from `/contracts/` (github_client, state_storage, token_manager)
- Generate model creation tasks from data-model.md entities (Repository, UserState, APIResponse)
- Generate CLI command tasks (root, monitor, cleanup) with cobra framework integration
- Generate integration tests from quickstart.md scenarios
- Generate service layer tasks (monitoring logic, repository comparison, progress indication)

**Ordering Strategy**:
- Setup: Go module init, dependency installation, project structure
- TDD Phase: Contract tests → Integration tests (all must fail initially)
- Models Phase: Repository, UserState, APIResponse structs with validation
- Interfaces Phase: GitHubClient, StateStorage, TokenManager interfaces
- Implementation Phase: API clients, storage implementations, CLI commands
- Services Phase: Monitor service, differ logic, progress indication
- Polish Phase: Error handling refinement, logging, documentation

**Parallel Execution**:
- Contract tests can run in parallel [P] (different interfaces)
- Model creation can run in parallel [P] (independent structs)
- Interface definitions can run in parallel [P] (separate packages)
- Implementation packages can be developed in parallel after interfaces complete

**Estimated Output**: 35-40 numbered, ordered tasks with clear dependencies and [P] markers for parallel execution

**IMPORTANT**: This phase is executed by the /tasks command, NOT by /plan

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command) - Generated research.md with technology decisions
- [x] Phase 1: Design complete (/plan command) - Generated data-model.md, contracts/, quickstart.md, updated .github/copilot-instructions.md
- [x] Phase 2: Task planning complete (/plan command - describe approach only) - Detailed task generation strategy documented
- [x] Phase 3: Tasks generated (/tasks command) - All 39 tasks created and completed
- [x] Phase 4: Implementation complete - Full CLI application built, tested and validated
- [x] Phase 5: Validation passed - All workflows tested, authentication working, production-ready

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS  
- [x] All NEEDS CLARIFICATION resolved - No technical unknowns remaining
- [x] Complexity deviations documented - No violations requiring justification

---
*Based on Constitution v2.1.1 - See `/memory/constitution.md`*
