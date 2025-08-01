# User-Managed Catalogs Implementation Checklist

## Project Status: ANALYSIS COMPLETE - READY FOR IMPLEMENTATION

**Last Updated**: August 1, 2025  
**Feature Spec**: [feature-spec.md](./feature-spec.md)  
**Investigation Notes**: `/Users/masegraye/dev/docker/id-writing/scratch/mcp-gateway-investigation.md`

## Development Workflow & TDD Instructions

### Essential Commands for Development

All commands should be run from `/Users/masegraye/dev/docker/workspaces/catalog-management/mcp-gateway/`:

```bash
# Scoped Test-Driven Development Cycle
go test ./cmd/docker-mcp/commands           # Test specific package (fastest)
go test ./cmd/docker-mcp/...               # Test cmd tree (medium scope)
make test                                   # Test entire system (broadest scope)
make docker-mcp                            # Build and install binary (only after tests pass)
make lint-darwin                           # Run linting (only when implementation complete)

# Manual testing commands
./dist/docker-mcp --help                    # Verify binary works
./dist/docker-mcp catalog --help           # Check catalog commands
./dist/docker-mcp gateway run --help       # Check gateway flags
```

### TDD Development Process

**CRITICAL**: Follow this exact workflow with progressively broader scope:

#### 1. Red Phase - Write Failing Tests First
```bash
# Test just the specific package you're working on
go test ./cmd/docker-mcp/commands    # Example: testing feature commands
# Should show failing tests for new functionality
```

#### 2. Green Phase - Make Tests Pass
```bash
# Write minimal implementation to make tests pass
go test ./cmd/docker-mcp/commands    # Test same package again
# Should show all tests passing in that package
```

#### 3. Package Integration Test
```bash
# Test all related packages together
go test ./cmd/docker-mcp/...         # Test entire cmd tree
# Should show all tests passing across packages
```

#### 4. Full System Test
```bash
# Test the entire system
make test
# Should show all tests passing system-wide
```

#### 5. Build Phase - Verify Compilation
```bash
# Only build after all tests pass
make docker-mcp
# Should build successfully and install to ~/.docker/cli-plugins/docker-mcp
```

#### 6. Manual Verification
```bash
# Test the actual CLI functionality
./dist/docker-mcp [your-new-command] --help
# Verify your changes work as expected
```

#### 7. Refactor Phase (Optional)
```bash
# Improve code while keeping tests green
go test ./path/to/package    # Test specific package after changes
```

#### 8. Lint Phase (End of Implementation Only)
```bash
# Only run linting when feature is complete
make lint-darwin
# Should show 0 issues
```

### Test Categories & Strategy

#### Unit Tests (Primary Focus)
- **Location**: Tests should be in `*_test.go` files alongside implementation
- **Speed**: Fast (< 1 second per test file)
- **Focus**: Individual function/method behavior
- **Run Command**: `make test`

#### Integration Tests (Secondary)
- **Location**: `cmd/docker-mcp/integration_test.go` and similar
- **Speed**: Slower (requires Docker daemon)
- **Focus**: End-to-end command workflows
- **Note**: Some are skipped in normal `make test` runs

#### Manual Tests (Verification)
- **Purpose**: Verify CLI behavior matches expectations
- **When**: After successful build, before marking tasks complete
- **Focus**: User experience and error messages

### Implementation Guidelines

1. **Test First**: Always write tests before implementation
2. **Small Steps**: Implement one small piece at a time
3. **Green Tests**: Never commit with failing tests
4. **Build Validation**: Only build after tests pass
5. **Manual Check**: Always verify CLI behavior manually
6. **Lint Last**: Only run linting when implementation is complete

### Error Handling Strategy

#### Test Failures
```bash
make test
# If tests fail, fix the code, don't skip tests
```

#### Build Failures
```bash
make docker-mcp
# If build fails, check compilation errors and fix
# Don't proceed to manual testing until build succeeds
```

#### Lint Failures (End of Development)
```bash
make lint-darwin
# Fix all linting issues before marking feature complete
```

### Baseline Verification ✅

The following baseline tests have been verified to work:
- `make docker-mcp` - Builds successfully 
- `make test` - All existing tests pass
- `make lint-darwin` - 0 linting issues
- CLI functionality verified working

**Ready to begin TDD implementation following the above workflow.**

## 🧪 Test Strategy Overview

### Required Test Files for TDD Implementation

Each implementation section below specifies **TEST FIRST** requirements. These test files must be created with failing tests before writing implementation code:

| Component | Test File | Purpose |
|-----------|-----------|---------|
| Feature Management | `cmd/docker-mcp/commands/feature_test.go` | Test feature enable/disable/list commands |
| Gateway Enhancement | `cmd/docker-mcp/commands/gateway_test.go` | Test --use-configured-catalogs flag validation |
| Catalog Loading | `cmd/docker-mcp/internal/catalog/catalog_test.go` | Test catalog precedence and loading logic |
| Export Command | `cmd/docker-mcp/catalog/export_test.go` | Test export functionality and protection |
| Command Visibility | `cmd/docker-mcp/catalog/*_test.go` | Test unhidden commands appear in help |

### Test-First Workflow Reminder

```bash
# 1. Write failing tests first
go test ./cmd/docker-mcp/commands -v  # Should show failing tests

# 2. Write minimal implementation  
go test ./cmd/docker-mcp/commands -v  # Should show passing tests

# 3. Test broader scope
go test ./cmd/docker-mcp/... -v      # Should show all integration tests pass

# 4. Test full system
make test                            # Should show all system tests pass

# 5. Build and verify
make docker-mcp                      # Should build successfully
```

## Quick Context for Claude Code Sessions

### What This Feature Does
Enable users to create and manage custom MCP server catalogs that automatically work with the gateway runtime, while maintaining backward compatibility with Docker Desktop through feature flag gating.

### Current Architecture Status
- ✅ **Catalog CRUD System**: Fully implemented by David Gageot (June-July 2025)
- ✅ **Multi-Catalog Gateway**: Implemented by Jim Clark (July 21, 2025) 
- ✅ **Infrastructure**: All underlying systems exist and work
- ❌ **Integration Gap**: Gateway runtime ignores catalog management system
- ❌ **User Discovery**: Catalog management commands are hidden

### Implementation Strategy
**Feature Flag Approach**: Use Docker CLI's existing `features` config system to gate new functionality, ensuring Docker Desktop compatibility.

## Implementation Checklist

### Phase 1: Core Implementation

#### 1.1 Feature Management System

**🧪 TEST FIRST**: `cmd/docker-mcp/commands/feature_test.go`
- [ ] **Write tests for feature command structure**
  ```go
  // Test cases needed:
  func TestFeatureEnableCommand(t *testing.T)     // Test enabling configured-catalogs
  func TestFeatureDisableCommand(t *testing.T)    // Test disabling configured-catalogs  
  func TestFeatureListCommand(t *testing.T)       // Test listing all features and status
  func TestFeatureInvalidFeature(t *testing.T)    // Test error for unknown feature names
  ```

- [ ] **Create feature command structure**
  - [ ] File: `cmd/docker-mcp/commands/feature.go`
  - [ ] Commands: `enable <feature>`, `disable <feature>`, `list`
  - [ ] Target: `~/.docker/config.json` → `features.configured-catalogs`

**🧪 TEST FIRST**: `cmd/docker-mcp/commands/feature_test.go` (validation utilities)
- [ ] **Write tests for feature validation utilities**
  ```go
  // Test cases needed:
  func TestIsFeatureEnabledTrue(t *testing.T)       // Test when feature is enabled
  func TestIsFeatureEnabledFalse(t *testing.T)      // Test when feature is disabled
  func TestIsFeatureEnabledMissing(t *testing.T)    // Test when config missing
  func TestIsFeatureEnabledCorrupt(t *testing.T)    // Test when config corrupted
  ```

- [ ] **Feature validation utilities**
  - [ ] Function: `isFeatureEnabled(dockerCli command.Cli, feature string) bool`
  - [ ] Handle missing config file gracefully
  - [ ] Support container mode detection

- [ ] **Integration with root command**
  - [ ] Add feature command to `cmd/docker-mcp/commands/root.go`
  - [ ] Ensure proper dockerCli context passing

#### 1.2 Gateway Command Enhancement  

**🧪 TEST FIRST**: `cmd/docker-mcp/commands/gateway_test.go`
- [ ] **Write tests for gateway flag validation**
  ```go
  // Test cases needed:
  func TestGatewayUseConfiguredCatalogsEnabled(t *testing.T)   // Test flag works when feature enabled
  func TestGatewayUseConfiguredCatalogsDisabled(t *testing.T)  // Test flag fails when feature disabled
  func TestGatewayFeatureFlagErrorMessage(t *testing.T)        // Test error message clarity
  func TestGatewayContainerModeDetection(t *testing.T)         // Test container mode handling
  ```

- [ ] **Add --use-configured-catalogs flag**
  - [ ] File: `cmd/docker-mcp/commands/gateway.go`
  - [ ] Flag: `--use-configured-catalogs` (boolean)
  - [ ] Validation: Check feature flag before allowing flag usage

- [ ] **Feature validation integration**
  - [ ] PreRunE validation for feature flag requirement
  - [ ] Clear error messages with exact enable command
  - [ ] Container mode detection and helpful errors

- [ ] **Pass flag to catalog system**
  - [ ] Update `catalog.Get()` call site
  - [ ] Pass useConfigured boolean parameter

#### 1.3 Catalog Loading Enhancement

**🧪 TEST FIRST**: `cmd/docker-mcp/internal/catalog/catalog_test.go`
- [ ] **Write tests for catalog loading logic**
  ```go
  // Test cases needed:
  func TestCatalogGetWithConfigured(t *testing.T)         // Test loading configured catalogs
  func TestCatalogGetWithoutConfigured(t *testing.T)      // Test default behavior unchanged
  func TestGetConfiguredCatalogsSuccess(t *testing.T)     // Test reading catalog.json
  func TestGetConfiguredCatalogsMissing(t *testing.T)     // Test missing catalog.json
  func TestGetConfiguredCatalogsCorrupt(t *testing.T)     // Test corrupted catalog.json
  func TestCatalogPrecedenceOrder(t *testing.T)           // Test Docker → Configured → CLI order
  ```

- [ ] **Update catalog.Get() signature**  
  - [ ] File: `cmd/docker-mcp/internal/catalog/catalog.go`
  - [ ] New signature: `Get(ctx context.Context, useConfigured bool, additionalCatalogs []string) (Catalog, error)`
  - [ ] Backward compatibility: Keep current `Get()` for existing callers

- [ ] **Implement getConfiguredCatalogs()**
  - [ ] Read from `~/.docker/mcp/catalog.json`
  - [ ] Return list of catalog file names
  - [ ] Handle missing/corrupted catalog registry gracefully

- [ ] **Implement catalog precedence logic**
  - [ ] Order: Built-in → Docker → Configured → CLI-specified
  - [ ] Use existing `ReadFrom()` function with ordered list
  - [ ] Comprehensive logging of loading process

#### 1.4 Enhanced Conflict Resolution & Logging

**🧪 TEST FIRST**: `cmd/docker-mcp/internal/catalog/catalog_test.go` (conflict resolution)
- [ ] **Write tests for conflict resolution**
  ```go
  // Test cases needed:
  func TestReadFromConflictResolution(t *testing.T)     // Test server name conflicts
  func TestReadFromLogging(t *testing.T)               // Test logging output
  func TestReadFromSourceTracking(t *testing.T)        // Test server source tracking
  func TestReadFromLastWinsPrecedence(t *testing.T)    // Test "last wins" behavior
  ```

- [ ] **Update ReadFrom() logging**  
  - [ ] File: `cmd/docker-mcp/internal/catalog/catalog.go`
  - [ ] Log catalog loading progress
  - [ ] Log server additions and conflicts
  - [ ] Log final catalog statistics

- [ ] **Server conflict handling**
  - [ ] Track server source catalogs
  - [ ] Log when servers are overridden
  - [ ] Maintain "last wins" precedence behavior

#### 1.5 Export Command Implementation

**🧪 TEST FIRST**: `cmd/docker-mcp/catalog/export_test.go`
- [ ] **Write tests for export command**
  ```go
  // Test cases needed:
  func TestExportCommandSuccess(t *testing.T)           // Test successful export
  func TestExportCommandDefaultFilename(t *testing.T)   // Test default filename generation
  func TestExportCommandCustomFilename(t *testing.T)    // Test custom output file
  func TestExportCommandDockerCatalogBlocked(t *testing.T) // Test docker-mcp protection
  func TestExportCommandCatalogNotFound(t *testing.T)   // Test missing catalog error
  func TestExportCommandPermissionError(t *testing.T)   // Test file permission errors
  ```

- [ ] **Create export command**
  - [ ] File: `cmd/docker-mcp/catalog/export.go` (new)
  - [ ] Command: `export <catalog-name> [output-file]`
  - [ ] Protection: Prevent export of `docker-mcp` official catalog

- [ ] **Export functionality**
  - [ ] Read catalog from `~/.docker/mcp/catalogs/{name}.yaml`
  - [ ] Default output filename: `./{catalog-name}.yaml`
  - [ ] Custom output file support
  - [ ] Validate catalog exists and is user-managed

- [ ] **Error handling**
  - [ ] Clear error for attempting to export official Docker catalog
  - [ ] Helpful error when catalog doesn't exist
  - [ ] File system permission error handling

#### 1.6 Command Visibility Updates

**🧪 TEST FIRST**: `cmd/docker-mcp/catalog/*_test.go` (existing command tests)
- [ ] **Write tests for command visibility**
  ```go
  // Test cases needed (add to existing test files):
  func TestImportCommandVisible(t *testing.T)         // Test import command shows in help
  func TestExportCommandVisible(t *testing.T)         // Test export command shows in help  
  func TestCreateCommandVisible(t *testing.T)         // Test create command shows in help
  func TestAddCommandVisible(t *testing.T)            // Test add command shows in help
  func TestForkCommandVisible(t *testing.T)           // Test fork command shows in help
  func TestRmCommandVisible(t *testing.T)             // Test rm command shows in help
  ```

- [ ] **Unhide catalog CRUD commands**
  - [ ] Files: `cmd/docker-mcp/catalog/{import,export,create,add,fork,rm}.go`  
  - [ ] Remove `Hidden: true` from command definitions
  - [ ] Update help text to mention feature flag requirement

- [ ] **Update command descriptions**
  - [ ] Reference feature flag in command descriptions
  - [ ] Provide clear usage examples
  - [ ] Link to feature enablement instructions

### Phase 2: Testing & Validation

#### 2.1 Unit Tests

- [ ] **Feature flag validation tests**
  - [ ] Test enabled/disabled/missing scenarios
  - [ ] Test Docker CLI integration
  - [ ] Test container mode behavior

- [ ] **Catalog loading tests**  
  - [ ] Test precedence order with multiple catalogs  
  - [ ] Test conflict resolution logging
  - [ ] Test graceful failure handling

- [ ] **Export command tests**
  - [ ] Test successful export of user catalogs
  - [ ] Test prevention of Docker catalog export
  - [ ] Test default vs custom output filenames
  - [ ] Test catalog not found scenarios

- [ ] **Gateway integration tests**
  - [ ] Test flag validation
  - [ ] Test feature flag gating
  - [ ] Test error message clarity

#### 2.2 Integration Tests

- [ ] **End-to-end workflow tests**
  - [ ] Create catalog → Add server → Enable feature → Run gateway
  - [ ] Export catalog → Import catalog → Verify server availability
  - [ ] Verify server availability in running gateway
  - [ ] Test catalog precedence with conflicts

- [ ] **Docker Desktop compatibility tests**
  - [ ] Verify `docker mcp gateway run` unchanged
  - [ ] Verify `docker mcp catalog show docker-mcp` unchanged  
  - [ ] Test Docker Desktop workflow regression

- [ ] **Container mode tests**
  - [ ] Test without volume mounts (graceful degradation)
  - [ ] Test with config volume mount (full functionality)
  - [ ] Test error messages for missing config

#### 2.3 Manual Testing Scenarios

- [ ] **Development workflow testing**
  - [ ] Test feature enable/disable cycle
  - [ ] Test catalog creation and management
  - [ ] Test export/import roundtrip workflow
  - [ ] Test gateway startup with configured catalogs

- [ ] **Multi-catalog testing**
  - [ ] Test overlapping server names across catalogs
  - [ ] Verify precedence order correctness
  - [ ] Test logging output comprehensiveness

### Phase 3: Documentation & Polish

#### 3.1 Documentation Updates

- [ ] **CLI reference updates**
  - [ ] Update generated docs for newly visible commands
  - [ ] Document feature flag requirement
  - [ ] Provide usage examples

- [ ] **User workflow documentation**
  - [ ] Development workflow example
  - [ ] Production setup guide
  - [ ] Container deployment instructions

#### 3.2 Error Handling Polish

- [ ] **Comprehensive error messages**
  - [ ] Feature not enabled → exact enable command
  - [ ] Container mode → volume mount instructions  
  - [ ] Catalog loading failures → helpful diagnostics

- [ ] **Logging improvements**
  - [ ] Structured logging for catalog operations
  - [ ] Clear startup messages about active catalogs
  - [ ] Performance timing for catalog loading

### Phase 4: Advanced Features (Future)

#### 4.1 Performance Optimizations
- [ ] **Catalog caching**
  - [ ] Cache loaded catalogs between gateway runs
  - [ ] Invalidate cache on catalog updates
  - [ ] Performance benchmarking

#### 4.2 Enhanced Validation
- [ ] **Catalog URL validation**
  - [ ] HTTPS requirement for remote catalogs
  - [ ] URL safety validation
  - [ ] Optional signature verification

#### 4.3 Docker Desktop Integration
- [ ] **UI for catalog management**
  - [ ] List configured catalogs in Docker Desktop
  - [ ] Enable/disable specific catalogs
  - [ ] Import catalogs via UI

## Progress Tracking

### Completed Analysis Work ✅
- [x] **Architecture Investigation**: Complete understanding of existing systems
- [x] **Git Blame Analysis**: Identified key contributors and implementation timeline  
- [x] **Docker Desktop Integration**: Understood compatibility requirements
- [x] **Binary Distribution Analysis**: Confirmed feature flag system works across all modes
- [x] **Feature Flag Research**: Identified Docker CLI `features` system as perfect solution
- [x] **Implementation Planning**: Complete technical design and user workflows
- [x] **Feature Specification**: Comprehensive spec document created
- [x] **Risk Assessment**: Identified and mitigated major risks

### Current Status
**ANALYSIS COMPLETE - READY FOR IMPLEMENTATION**

All research, investigation, and planning work is complete. The next steps are pure implementation following the detailed plan in the feature specification.

## Key Implementation Notes

### Critical Success Factors
1. **Maintain Docker Desktop Compatibility**: Never change default behavior of `docker mcp gateway run`
2. **Feature Flag Gating**: All new functionality MUST be gated behind `configured-catalogs` feature
3. **Graceful Degradation**: System must work properly when config is inaccessible
4. **Clear User Communication**: Error messages must provide exact commands to fix issues

### Technical Constraints  
1. **Backward Compatibility**: All existing command signatures and behaviors preserved
2. **Container Support**: Feature must work in container mode with proper volume mounts
3. **Performance**: Catalog loading must not significantly impact gateway startup time
4. **Security**: No privilege escalation or unsafe file operations

### Architecture Decisions Made
1. **Feature Flag System**: Use Docker CLI's existing `features` config approach
2. **Catalog Precedence**: Last-loaded-wins for server conflicts (Docker → Configured → CLI)
3. **Command Structure**: Add `--use-configured-catalogs` flag rather than change defaults
4. **Error Handling**: Fail-safe approach where config problems disable feature rather than break gateway

## File Locations Reference

### Implementation Target Files
- **Feature Commands**: `cmd/docker-mcp/commands/feature.go` (new)
- **Export Command**: `cmd/docker-mcp/catalog/export.go` (new)
- **Gateway Enhancement**: `cmd/docker-mcp/commands/gateway.go` 
- **Catalog Loading**: `cmd/docker-mcp/catalog/catalog.go`
- **Command Visibility**: `cmd/docker-mcp/catalog/{import,export,create,add,fork,rm}.go`

### Key Reference Files  
- **Command Structure**: `cmd/docker-mcp/commands/root.go`
- **Docker CLI Integration**: `cmd/docker-mcp/main.go:33-35`
- **Config Pattern**: `cli/cli/config/configfile/file.go:120` (Features map)
- **Storage Pattern**: `cmd/docker-mcp/internal/config/readwrite.go`

### Investigation Documentation
- **Detailed Investigation**: `/Users/masegraye/dev/docker/id-writing/scratch/mcp-gateway-investigation.md`
- **Feature Specification**: `docs/feature-specs/catalog-management/feature-spec.md`

## Contact Information

**Primary Technical Contact**: Jim Clark (@slimslenderslacks)
- Implemented multi-catalog gateway support (July 21, 2025)
- Best person to consult for architecture questions

**Secondary Contact**: David Gageot  
- Implemented catalog management system (June-July 2025)
- Expert on catalog CRUD operations and storage patterns

---

## Quick Start for New Claude Code Sessions

1. **Read the feature spec**: Start with `feature-spec.md` for complete context
2. **Check current status**: Review completed items in this checklist
3. **Start implementation**: Begin with Phase 1.1 (Feature Management System)
4. **Test thoroughly**: Each component must work in plugin, standalone, and container modes  
5. **Maintain compatibility**: Docker Desktop behavior must never change

**Key Command to Validate Success**:
```bash
# This workflow should work after implementation
docker mcp feature enable configured-catalogs
docker mcp catalog create my-servers  
docker mcp catalog add my-servers test-server ./server.yaml
docker mcp catalog export my-servers ./backup.yaml
docker mcp gateway run --use-configured-catalogs
# Gateway should start with both Docker servers AND test-server available

# Validate export protection
docker mcp catalog export docker-mcp ./should-fail.yaml
# Should fail with: "Cannot export official Docker catalog 'docker-mcp'"
```