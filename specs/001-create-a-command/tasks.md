# Tasks: GitHub Stars Monitor CLI

**Input**: Design documents from `/specs/001-create-a-command/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)
```
1. Load plan.md from feature directory
   → ✅ Implementation plan loaded with Go 1.25+ CLI architecture
   → ✅ Extracted: Cobra CLI, go-github client, clean architecture
2. Load optional design documents:
   → ✅ data-model.md: Repository, UserState, APIResponse entities
   → ✅ contracts/: github_client, state_storage, token_manager interfaces
   → ✅ research.md: Technology stack decisions and rationale
   → ✅ quickstart.md: End-to-end user workflow scenarios  
3. Generate tasks by category:
   → ✅ Setup: Go module, dependencies, project structure
   → ✅ Tests: Contract tests, integration tests (TDD)
   → ✅ Core: Models, interfaces, CLI commands
   → ✅ Integration: API clients, storage, authentication
   → ✅ Polish: Unit tests, performance, documentation
4. Apply task rules:
   → ✅ Different files marked [P] for parallel execution
   → ✅ Same file sequential (no [P])
   → ✅ Tests before implementation (TDD)
5. Number tasks sequentially (T001-T039)
6. Generate dependency graph and parallel execution examples
7. Validate task completeness:
   → ✅ All 3 contracts have corresponding tests
   → ✅ All 3 entities have model creation tasks
   → ✅ All quickstart scenarios have integration tests
   → ✅ Username validation and logging tasks added
8. Return: SUCCESS (39 tasks completed successfully)
```

## Format: `[ID] [P?] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions
**Single CLI project**: Paths assume Go project structure from plan.md
- `cmd/star-watcher/` - CLI entry point
- `internal/` - Internal packages (github, storage, auth, cli, monitor)
- `tests/` - Test organization (contract, integration, unit)

## Phase 3.1: Setup
- [x] T001 Initialize Go module with `go mod init github.com/akme/gh-stars-watcher`
- [x] T002 Create project directory structure per plan.md layout (cmd/, internal/, tests/)
- [x] T003 [P] Install Go dependencies: cobra, go-github, keychain libraries
- [x] T004 [P] Configure golint, gofmt, and go vet in Makefile or scripts
- [x] T005 [P] Create README.md with installation and basic usage instructions

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3
**CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation**
- [x] T006 [P] Contract test GitHubClient interface in tests/contract/github_client_test.go
- [x] T007 [P] Contract test StateStorage interface in tests/contract/state_storage_test.go
- [x] T008 [P] Contract test TokenManager interface in tests/contract/token_manager_test.go
- [x] T009 [P] Integration test first run workflow in tests/integration/first_run_test.go
- [x] T010 [P] Integration test subsequent monitoring in tests/integration/monitor_flow_test.go
- [x] T011 [P] Integration test authentication setup in tests/integration/auth_flow_test.go
- [x] T012 [P] Integration test JSON output format in tests/integration/output_format_test.go
- [x] T013 [P] Integration test cleanup command in tests/integration/cleanup_test.go
- [x] T014 [P] Integration test error scenarios in tests/integration/error_handling_test.go

## Phase 3.3: Core Implementation (ONLY after tests are failing)
- [x] T015 [P] Repository model struct in internal/storage/models.go
- [x] T016 UserState model struct in internal/storage/models.go (same file as T015)
- [x] T017 [P] APIResponse model struct with repository details (full name, description, star count, URL, timestamps) in internal/github/models.go
- [x] T018 [P] GitHubClient interface definition in internal/github/client.go
- [x] T019 [P] StateStorage interface definition in internal/storage/state.go
- [x] T020 [P] TokenManager interface definition in internal/auth/token.go
- [x] T021 Main CLI entry point in cmd/star-watcher/main.go
- [x] T022 Root command setup with Cobra and logging configuration in internal/cli/root.go
- [x] T023 Monitor command implementation with GitHub username validation in internal/cli/monitor.go
- [x] T024 Cleanup command implementation in internal/cli/cleanup.go

## Phase 3.4: Integration
- [x] T025 GitHub username validation utility in internal/cli/validation.go
- [x] T026 GitHub API client implementation in internal/github/api.go
- [x] T027 JSON file storage implementation in internal/storage/json.go
- [x] T028 [P] OS keychain integration in internal/auth/keychain.go
- [x] T029 [P] Interactive token prompt in internal/auth/prompt.go
- [x] T030 Core monitoring service in internal/monitor/service.go
- [x] T031 Repository comparison logic in internal/monitor/differ.go
- [x] T032 [P] Progress indication implementation in internal/monitor/progress.go
- [x] T033 [P] Output formatting (JSON/text) in internal/cli/output.go
- [x] T034 Error handling with proper exit codes and configurable logging (quiet/normal/verbose) across all packages

## Phase 3.5: Polish
- [x] T035 [P] Unit tests for repository comparison in tests/unit/differ_test.go
- [x] T036 [P] Unit tests for output formatting in tests/unit/output_test.go
- [x] T037 [P] Unit tests for progress indication in tests/unit/progress_test.go
- [x] T038 Performance tests for GitHub API rate limiting and memory usage
- [x] T039 Update README.md with complete usage examples and quickstart guide

## Dependencies
- Setup (T001-T005) before all other phases
- Tests (T006-T014) before implementation (T015-T034)
- Models (T015-T017) before interfaces (T018-T020)
- Interfaces (T018-T020) before CLI commands (T021-T024)
- CLI commands (T021-T024) before implementations (T025-T034)
- Core implementation (T015-T034) before polish (T035-T039)

**Specific Dependencies**:
- T015, T016 (same file) must be sequential - no [P] on T016
- T023 depends on T025 (monitor command uses username validation)
- T026 depends on T017, T018 (GitHub models and interface)
- T027 depends on T015, T016, T019 (Storage models and interface)  
- T028, T029 depend on T020 (TokenManager interface)
- T030, T031 depend on T018, T019 (service layer uses interfaces)
- T023, T024 depend on T022 (commands use root setup)

## Parallel Example
```
# Phase 3.2 - Launch contract tests together:
Task: "Contract test GitHubClient interface in tests/contract/github_client_test.go"
Task: "Contract test StateStorage interface in tests/contract/state_storage_test.go"
Task: "Contract test TokenManager interface in tests/contract/token_manager_test.go"

# Phase 3.2 - Launch integration tests together:
Task: "Integration test first run workflow in tests/integration/first_run_test.go"
Task: "Integration test subsequent monitoring in tests/integration/monitor_flow_test.go"
Task: "Integration test authentication setup in tests/integration/auth_flow_test.go"

# Phase 3.3 - Launch interface definitions together:
Task: "GitHubClient interface definition in internal/github/client.go"  
Task: "StateStorage interface definition in internal/storage/state.go"
Task: "TokenManager interface definition in internal/auth/token.go"

# Phase 3.4 - Launch independent implementations together:
Task: "OS keychain integration in internal/auth/keychain.go"
Task: "Interactive token prompt in internal/auth/prompt.go"
Task: "Progress indication implementation in internal/monitor/progress.go"
Task: "Output formatting (JSON/text) in internal/cli/output.go"
```

## Notes
- [P] tasks = different files, no dependencies, can run in parallel
- Tests MUST fail before implementing (verify with `go test ./...`)
- Follow constitutional TDD requirements: Red → Green → Refactor
- Each task should result in a working, tested, documented component
- Commit after each task with descriptive commit message
- Use Go idioms: interfaces, error handling, table-driven tests

## Task Generation Rules Applied

1. **From Contracts**: 3 contract files → 3 contract test tasks [P] (T006-T008)
2. **From Data Model**: 3 entities → 3 model creation tasks (T015-T017) 
3. **From User Stories**: Quickstart scenarios → 6 integration tests [P] (T009-T014)
4. **From Plan Structure**: CLI architecture → interface + implementation tasks
5. **Ordering**: Setup → Tests → Models → Interfaces → CLI → Implementations → Polish
6. **Dependencies**: Sequential tasks for same files, parallel for independent files

## Validation Checklist

- [x] All contracts have corresponding tests (github_client, state_storage, token_manager)
- [x] All entities have model tasks (Repository, UserState, APIResponse)
- [x] All tests come before implementation (Phase 3.2 before 3.3)
- [x] Parallel tasks truly independent (different files, no shared dependencies)
- [x] Each task specifies exact file path for implementation
- [x] No task modifies same file as another [P] task (T015/T016 sequential as noted)
- [x] Username validation task added (T025)
- [x] Configurable logging integrated into root command and error handling
- [x] Repository details mapping clarified in T017 APIResponse model
- [x] Exit codes explicitly mentioned in T034 error handling