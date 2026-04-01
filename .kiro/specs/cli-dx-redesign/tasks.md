# Implementation Plan: CLI DX Redesign

## Overview

This plan implements the remaining components of the Jotform CLI DX Redesign feature. The top-level shortcuts and project context system are already implemented. This task list focuses on implementing the new commands (init, clone, status, open) and shell completion support.

## Tasks

- [ ] 1. Implement jotform init command
  - [x] 1.1 Create init command with interactive and non-interactive modes
    - Implement interactive prompts for existing/new form selection
    - Implement form ID validation (numeric format)
    - Implement form title prompts for new form mode
    - Implement schema file path prompts with default value
    - Handle Ctrl+C gracefully without creating partial files
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 14.1, 14.2, 14.3, 14.4, 14.5, 14.6, 14.7_
  
  - [x] 1.2 Implement form creation and schema export logic
    - Fetch form data from API for existing forms
    - Create new form via API for new form mode
    - Export form schema to local file
    - Create .jotform.yaml with form metadata
    - Display success messages with suggested next commands
    - _Requirements: 3.5, 3.6, 3.7, 3.8, 13.1, 13.2, 13.3, 13.4_
  
  - [ ]* 1.3 Write property test for init atomicity
    - **Property 8: Init Atomicity**
    - **Validates: Requirements 3.6, 3.7**
  
  - [ ]* 1.4 Write unit tests for init command
    - Test interactive mode prompts
    - Test non-interactive mode with flags
    - Test error handling for invalid inputs
    - Test file creation with correct permissions
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 2. Implement jotform clone command
  - [x] 2.1 Create clone command with directory creation logic
    - Implement form ID argument parsing
    - Fetch form data from API
    - Slugify form title for directory name
    - Handle directory name collisions with numeric suffix
    - Support --name flag to override directory name
    - Handle --force flag for overwriting existing directories
    - _Requirements: 4.1, 4.2, 4.3, 4.6, 4.7_
  
  - [x] 2.2 Implement clone workflow integration
    - Create target directory
    - Call SaveProject to create .jotform.yaml
    - Export form schema to new directory
    - Display success messages
    - _Requirements: 4.4, 4.5_
  
  - [ ]* 2.3 Write property test for clone completeness
    - **Property 13: Clone Creates Complete Project**
    - **Validates: Requirements 4.1, 4.4, 4.5**
  
  - [ ]* 2.4 Write property test for slugify safety
    - **Property 11: Slugify Produces Filesystem-Safe Names**
    - **Validates: Requirements 11.1, 11.2, 11.3, 11.4**
  
  - [ ]* 2.5 Write property test for slugify idempotence
    - **Property 12: Slugify Idempotence**
    - **Validates: Requirements 4.2, 11.1, 11.2, 11.3**
  
  - [ ]* 2.6 Write unit tests for clone command
    - Test directory creation with slugified names
    - Test collision handling with numeric suffixes
    - Test --name flag override
    - Test --force flag behavior
    - Test error handling for existing directories
    - _Requirements: 4.1, 4.2, 4.3, 4.6, 4.7_

- [x] 3. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [-] 4. Implement jotform status command
  - [x] 4.1 Create StatusReport data model and change detection
    - Define StatusReport struct with all required fields
    - Define Change struct with type, path, old/new values
    - Implement ChangeType enum (Added, Modified, Deleted)
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_
  
  - [x] 4.2 Implement status computation algorithm
    - Load local schema file
    - Fetch remote form from API
    - Compute structural diff between local and remote
    - Detect added, modified, and deleted fields
    - Use JSON path notation for changed fields
    - _Requirements: 5.4, 5.5, 5.6_
  
  - [-] 4.3 Implement status display and output formatting
    - Display form ID and name
    - Display local and remote modification times
    - Display all changes with type indicators
    - Show old and new values for modified fields
    - Display "no changes" message when schemas match
    - Suggest next actions (push or pull)
    - Support --summary flag for counts only
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.8, 5.9_
  
  - [ ]* 4.4 Write property test for status idempotence
    - **Property 7: Status Idempotence**
    - **Validates: Requirements 5.1, 5.2, 5.3**
  
  - [ ]* 4.5 Write property test for status shows all changes
    - **Property 18: Status Shows All Changes**
    - **Validates: Requirements 5.4, 5.5, 5.6**
  
  - [ ]* 4.6 Write unit tests for status command
    - Test status with no changes
    - Test status with added fields
    - Test status with modified fields
    - Test status with deleted fields
    - Test --summary flag output
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.8, 5.9_

- [ ] 5. Implement jotform open command
  - [ ] 5.1 Create open command with form ID resolution
    - Resolve form ID from args or project context
    - Construct Jotform form URL
    - Validate form ID format before URL construction
    - _Requirements: 6.1, 6.2, 6.3_
  
  - [ ] 5.2 Implement cross-platform browser launching
    - Detect operating system (macOS, Linux, Windows)
    - Use appropriate command for each platform (open, xdg-open, start)
    - Handle browser launch failures gracefully
    - Display form URL when browser launch fails
    - _Requirements: 6.4, 6.5, 6.6, 6.7, 15.1, 15.2, 15.3_
  
  - [ ]* 5.3 Write property test for URL construction
    - **Property 19: URL Construction Correctness**
    - **Validates: Requirement 6.3**
  
  - [ ]* 5.4 Write unit tests for open command
    - Test form ID resolution from context
    - Test form ID resolution from argument
    - Test URL construction
    - Test browser launch on different platforms
    - Test error handling for browser launch failures
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7_

- [ ] 6. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 7. Implement shell completion support
  - [ ] 7.1 Set up Cobra completion command
    - Add completion command to root command
    - Support bash, zsh, fish, and powershell shells
    - Generate completion scripts for each shell
    - _Requirements: 7.1, 7.2, 7.3, 7.4_
  
  - [ ] 7.2 Implement static completion for commands and flags
    - Ensure all command names are included in completion
    - Ensure all flag names are included in completion
    - _Requirements: 7.5, 7.6_
  
  - [ ]* 7.3 Implement optional dynamic form ID completion
    - Add config option for enabling dynamic completion
    - Implement form ID fetching from API
    - Implement caching with configurable TTL
    - Add timeout for completion API calls (500ms)
    - _Requirements: 7.7_
  
  - [ ]* 7.4 Write property test for completion includes all commands
    - **Property 24: Shell Completion Includes All Commands**
    - **Validates: Requirement 7.5**
  
  - [ ]* 7.5 Write property test for completion includes all flags
    - **Property 25: Shell Completion Includes All Flags**
    - **Validates: Requirement 7.6**
  
  - [ ]* 7.6 Write unit tests for completion
    - Test completion script generation for each shell
    - Test static completion for commands
    - Test static completion for flags
    - Test dynamic completion with caching
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7_

- [ ] 8. Integrate context resolution with existing commands
  - [ ] 8.1 Update shortcut commands to use context resolution
    - Update shortGetCmd to use ResolveFormID
    - Update shortRmCmd to use ResolveFormID
    - Update shortPullCmd to use ResolveFormID and ResolveSchemaFile
    - Update shortPushCmd to use ResolveFormID and ResolveSchemaFile
    - Update shortDiffCmd to use ResolveFormID and ResolveSchemaFile
    - Update shortWatchCmd to use ResolveFormID
    - _Requirements: 2.2, 2.3, 2.5, 2.6, 9.1, 9.2, 9.3_
  
  - [ ] 8.2 Add error handling and user guidance
    - Ensure helpful error messages when form ID cannot be resolved
    - Ensure helpful error messages when schema file cannot be resolved
    - Guide users to run 'jotform init' when no context exists
    - _Requirements: 2.7, 10.1, 10.2, 10.3, 10.8_
  
  - [ ]* 8.3 Write property test for explicit arguments override context
    - **Property 5: Explicit Arguments Override Context**
    - **Validates: Requirements 2.5, 9.1, 9.2**
  
  - [ ]* 8.4 Write property test for context fallback behavior
    - **Property 23: Context Fallback Behavior**
    - **Validates: Requirements 2.2, 9.3**
  
  - [ ]* 8.5 Write integration tests for context-aware workflow
    - Test init → diff → push workflow
    - Test clone → status → pull workflow
    - Test explicit args override context
    - _Requirements: 2.2, 2.3, 2.5, 9.1, 9.2, 9.3_

- [ ] 9. Add validation and security measures
  - [ ] 9.1 Implement path validation for schema files
    - Reject absolute paths in .jotform.yaml schema field
    - Reject paths containing parent directory references (..)
    - Validate schema paths exist when loading project
    - _Requirements: 11.5, 11.6, 11.7, 13.7_
  
  - [ ] 9.2 Implement file permission handling
    - Set .jotform.yaml permissions to 0644
    - Set schema file permissions to 0644
    - _Requirements: 11.8_
  
  - [ ] 9.3 Implement form ID validation
    - Validate form ID is non-empty
    - Validate form ID matches numeric format
    - _Requirements: 13.5, 13.6_
  
  - [ ]* 9.4 Write property test for path validation
    - **Property 14: Path Validation Rejects Unsafe Paths**
    - **Validates: Requirements 11.5, 11.6**
  
  - [ ]* 9.5 Write property test for file permissions
    - **Property 20: File Permissions Are Restrictive**
    - **Validates: Requirement 11.8**
  
  - [ ]* 9.6 Write property test for form ID validation
    - **Property 21: Form ID Validation**
    - **Validates: Requirements 13.5, 13.6**
  
  - [ ]* 9.7 Write unit tests for validation
    - Test path validation rejects unsafe paths
    - Test file permissions are set correctly
    - Test form ID validation
    - _Requirements: 11.5, 11.6, 11.7, 11.8, 13.5, 13.6, 13.7_

- [ ] 10. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Top-level shortcuts (Phase 1) are already implemented in cmd/shortcuts.go
- Project context system (Phase 2) is partially implemented in internal/config/context.go
- This task list focuses on implementing the remaining commands and features
