# Requirements Document: Jotform CLI DX Redesign

## Introduction

The Jotform CLI DX Redesign transforms the command-line interface to provide a streamlined, developer-friendly experience. The core philosophy is "Her Zaman Daha Az Yazmak" (Always Write Less) — daily operations should never exceed 2 words. This redesign introduces top-level shortcuts, project context awareness, and enhanced workflow commands while maintaining full backward compatibility with existing grouped commands.

## Glossary

- **CLI**: Command-Line Interface - the text-based interface for interacting with the Jotform system
- **Project_Context**: Configuration stored in `.jotform.yaml` that associates a directory with a specific Jotform form
- **Context_Resolver**: Component that determines form IDs and schema paths from project context or explicit arguments
- **Shortcut_Command**: Top-level command that provides abbreviated syntax for common operations
- **Grouped_Command**: Original hierarchical command structure (e.g., `jotform forms list`)
- **Schema_File**: Local YAML/JSON file containing form structure and configuration
- **Form_ID**: Unique numeric identifier for a Jotform form
- **Slugify**: Process of converting text to filesystem-safe lowercase-hyphenated format

## Requirements

### Requirement 1: Top-Level Command Shortcuts

**User Story:** As a developer, I want to use short, memorable commands for common operations, so that I can work efficiently without typing verbose command hierarchies.

#### Acceptance Criteria

1. WHEN a user types `jotform login`, THE CLI SHALL execute the authentication login flow
2. WHEN a user types `jotform logout`, THE CLI SHALL execute the authentication logout flow
3. WHEN a user types `jotform whoami`, THE CLI SHALL display current user information
4. WHEN a user types `jotform ls`, THE CLI SHALL list all forms
5. WHEN a user types `jotform get [form_id]`, THE CLI SHALL retrieve and display the specified form
6. WHEN a user types `jotform pull [form_id]`, THE CLI SHALL download the form schema to a local file
7. WHEN a user types `jotform push [form_id]`, THE CLI SHALL upload local schema changes to the remote form
8. WHEN a user types `jotform diff [form_id]`, THE CLI SHALL display differences between local and remote schemas
9. WHEN a user types `jotform watch [form_id]`, THE CLI SHALL stream real-time form submissions
10. WHEN a user types `jotform generate [prompt]`, THE CLI SHALL generate a form schema using AI
11. WHEN a user types `jotform new`, THE CLI SHALL create a new form
12. WHEN a user types `jotform rm [form_id]`, THE CLI SHALL delete the specified form

### Requirement 2: Project Context System

**User Story:** As a developer, I want to work within a project directory without repeatedly specifying form IDs and file paths, so that my workflow is faster and less error-prone.

#### Acceptance Criteria

1. WHEN a user runs `jotform init` in a directory, THE CLI SHALL create a `.jotform.yaml` configuration file
2. WHEN a `.jotform.yaml` file exists, THE CLI SHALL automatically resolve form IDs from the configuration
3. WHEN a `.jotform.yaml` file exists, THE CLI SHALL automatically resolve schema file paths from the configuration
4. WHEN no `.jotform.yaml` file exists in the current directory, THE Context_Resolver SHALL search parent directories up to the filesystem root
5. WHEN a user provides an explicit form ID argument, THE CLI SHALL use that argument instead of the project context
6. WHEN a user provides an explicit `--file` flag, THE CLI SHALL use that file path instead of the project context
7. WHEN no project context exists and no explicit arguments are provided, THE CLI SHALL return an error message guiding the user to run `jotform init` or provide arguments

### Requirement 3: Project Initialization

**User Story:** As a developer, I want to initialize a project directory for an existing or new form, so that I can start working with project context immediately.

#### Acceptance Criteria

1. WHEN a user runs `jotform init` without flags, THE CLI SHALL prompt interactively for initialization options
2. WHEN a user selects "existing form" mode, THE CLI SHALL prompt for a form ID
3. WHEN a user selects "new form" mode, THE CLI SHALL prompt for a form title and create a new form
4. WHEN a user provides `--form-id` flag, THE CLI SHALL initialize in non-interactive mode for an existing form
5. WHEN a user provides `--new` flag, THE CLI SHALL initialize in non-interactive mode for a new form
6. WHEN initialization completes, THE CLI SHALL create a `.jotform.yaml` file with form metadata
7. WHEN initialization completes, THE CLI SHALL export the form schema to a local file
8. WHEN initialization completes, THE CLI SHALL display success messages with suggested next commands

### Requirement 4: Form Cloning

**User Story:** As a developer, I want to clone a form into a new directory with all necessary configuration, so that I can quickly set up a project workspace.

#### Acceptance Criteria

1. WHEN a user runs `jotform clone [form_id]`, THE CLI SHALL create a new directory named after the form title
2. WHEN creating the clone directory, THE CLI SHALL slugify the form title to ensure filesystem safety
3. WHEN the target directory already exists, THE CLI SHALL return an error unless `--force` flag is provided
4. WHEN cloning completes, THE CLI SHALL create a `.jotform.yaml` file in the new directory
5. WHEN cloning completes, THE CLI SHALL export the form schema to the new directory
6. WHEN a user provides `--name` flag, THE CLI SHALL use the specified name instead of the slugified form title
7. WHEN the slugified directory name already exists, THE CLI SHALL append a numeric suffix to create a unique name

### Requirement 5: Status Reporting

**User Story:** As a developer, I want to see a summary of differences between my local schema and the remote form, so that I can understand what changes exist before pushing or pulling.

#### Acceptance Criteria

1. WHEN a user runs `jotform status`, THE CLI SHALL display the form ID and name
2. WHEN a user runs `jotform status`, THE CLI SHALL display local schema file path and modification time
3. WHEN a user runs `jotform status`, THE CLI SHALL display remote form modification time
4. WHEN differences exist, THE CLI SHALL list all changes with type indicators (added, modified, deleted)
5. WHEN differences exist, THE CLI SHALL show the JSON path for each changed field
6. WHEN differences exist, THE CLI SHALL display old and new values for modified fields
7. WHEN no differences exist, THE CLI SHALL display a message indicating schemas are in sync
8. WHEN differences exist, THE CLI SHALL suggest next actions (push or pull)
9. WHEN a user provides `--summary` flag, THE CLI SHALL display only change counts without field details

### Requirement 6: Browser Integration

**User Story:** As a developer, I want to quickly open a form in my browser for visual inspection, so that I can verify changes without manually navigating to the Jotform website.

#### Acceptance Criteria

1. WHEN a user runs `jotform open`, THE CLI SHALL resolve the form ID from project context
2. WHEN a user runs `jotform open [form_id]`, THE CLI SHALL use the provided form ID
3. WHEN opening a form, THE CLI SHALL construct the correct Jotform form URL
4. WHEN opening a form, THE CLI SHALL launch the system default browser with the form URL
5. WHEN the browser launch succeeds, THE CLI SHALL return success
6. WHEN the browser launch fails, THE CLI SHALL display the form URL for manual copying
7. WHEN opening a form, THE CLI SHALL work correctly on macOS, Linux, and Windows platforms

### Requirement 7: Shell Completion

**User Story:** As a developer, I want shell completion for commands and flags, so that I can work faster and discover available options.

#### Acceptance Criteria

1. WHEN a user runs `jotform completion bash`, THE CLI SHALL generate bash completion script
2. WHEN a user runs `jotform completion zsh`, THE CLI SHALL generate zsh completion script
3. WHEN a user runs `jotform completion fish`, THE CLI SHALL generate fish completion script
4. WHEN a user runs `jotform completion powershell`, THE CLI SHALL generate PowerShell completion script
5. WHEN shell completion is active, THE CLI SHALL provide completions for all command names
6. WHEN shell completion is active, THE CLI SHALL provide completions for all flag names
7. WHERE dynamic completion is enabled in configuration, THE CLI SHALL provide form ID completions from the API

### Requirement 8: Backward Compatibility

**User Story:** As an existing user, I want all my current commands and scripts to continue working without modification, so that the upgrade is seamless and non-breaking.

#### Acceptance Criteria

1. WHEN a user runs any grouped command (e.g., `jotform forms list`), THE CLI SHALL execute with identical behavior to previous versions
2. WHEN a user provides existing flags, THE CLI SHALL interpret them with identical behavior to previous versions
3. WHEN a command produces JSON output, THE CLI SHALL maintain the same output structure as previous versions
4. WHEN a command produces table output, THE CLI SHALL maintain compatible formatting
5. WHEN a command succeeds, THE CLI SHALL return exit code 0
6. WHEN a command fails, THE CLI SHALL return a non-zero exit code
7. WHEN configuration files exist from previous versions, THE CLI SHALL continue to read and respect them

### Requirement 9: Context Resolution Priority

**User Story:** As a developer, I want explicit command-line arguments to override project context, so that I can work with multiple forms without changing directories.

#### Acceptance Criteria

1. WHEN a user provides a form ID argument and project context exists, THE Context_Resolver SHALL use the argument
2. WHEN a user provides a `--file` flag and project context exists, THE Context_Resolver SHALL use the flag value
3. WHEN no arguments are provided and project context exists, THE Context_Resolver SHALL use the context values
4. WHEN no arguments are provided and no project context exists, THE Context_Resolver SHALL return an error
5. WHERE verbose logging is enabled, THE CLI SHALL indicate when arguments override project context

### Requirement 10: Error Handling and User Guidance

**User Story:** As a developer, I want clear, actionable error messages when something goes wrong, so that I can quickly resolve issues and continue working.

#### Acceptance Criteria

1. WHEN a form ID cannot be resolved, THE CLI SHALL display an error message explaining how to provide one
2. WHEN a `.jotform.yaml` file is malformed, THE CLI SHALL display the parsing error and file location
3. WHEN a schema file is not found, THE CLI SHALL suggest running `jotform pull` to download it
4. WHEN a form ID does not exist or is inaccessible, THE CLI SHALL display a verification message
5. WHEN a directory is not writable, THE CLI SHALL display a permissions error with the directory path
6. WHEN a target directory already exists during clone, THE CLI SHALL suggest using `--force` or choosing a different name
7. WHEN browser launch fails, THE CLI SHALL display the form URL for manual access
8. WHEN API authentication fails, THE CLI SHALL guide the user to run `jotform login`

### Requirement 11: File System Safety

**User Story:** As a developer, I want the CLI to handle file paths and names safely, so that I don't encounter issues with special characters or path traversal vulnerabilities.

#### Acceptance Criteria

1. WHEN slugifying a form title, THE CLI SHALL convert all characters to lowercase
2. WHEN slugifying a form title, THE CLI SHALL replace spaces and special characters with hyphens
3. WHEN slugifying a form title, THE CLI SHALL remove consecutive hyphens
4. WHEN slugifying a form title, THE CLI SHALL limit the output length to 50 characters
5. WHEN validating schema paths from `.jotform.yaml`, THE CLI SHALL reject absolute paths
6. WHEN validating schema paths from `.jotform.yaml`, THE CLI SHALL reject paths containing parent directory references (`..`)
7. WHEN resolving schema paths, THE CLI SHALL convert relative paths to absolute paths based on the `.jotform.yaml` location
8. WHEN creating files, THE CLI SHALL set permissions to 0644 (readable by owner and group, writable by owner only)

### Requirement 12: Destructive Operation Safety

**User Story:** As a developer, I want confirmation prompts for destructive operations, so that I don't accidentally delete forms or overwrite changes.

#### Acceptance Criteria

1. WHEN a user runs `jotform rm [form_id]` without `--force` flag, THE CLI SHALL prompt for confirmation
2. WHEN a user runs `jotform push` with local changes without `--force` flag, THE CLI SHALL prompt for confirmation
3. WHEN a user confirms a destructive operation, THE CLI SHALL proceed with execution
4. WHEN a user declines a destructive operation, THE CLI SHALL cancel and return success
5. WHEN a user provides `--force` flag, THE CLI SHALL skip confirmation prompts
6. WHEN a user provides `--dry-run` flag, THE CLI SHALL display planned actions without executing them
7. WHEN running in `--dry-run` mode, THE CLI SHALL make no API calls that modify data

### Requirement 13: Configuration File Format

**User Story:** As a developer, I want `.jotform.yaml` files to be human-readable and well-documented, so that I can understand and modify them when needed.

#### Acceptance Criteria

1. WHEN creating a `.jotform.yaml` file, THE CLI SHALL include header comments explaining the file purpose
2. WHEN creating a `.jotform.yaml` file, THE CLI SHALL include a `form_id` field with the numeric form identifier
3. WHEN creating a `.jotform.yaml` file, THE CLI SHALL include a `name` field with the human-readable form title
4. WHEN creating a `.jotform.yaml` file, THE CLI SHALL include a `schema` field with the relative path to the schema file
5. WHEN loading a `.jotform.yaml` file, THE CLI SHALL validate that `form_id` is non-empty
6. WHEN loading a `.jotform.yaml` file, THE CLI SHALL validate that `form_id` matches numeric format
7. WHEN loading a `.jotform.yaml` file, THE CLI SHALL validate that `schema` path exists if specified

### Requirement 14: Interactive Prompt Experience

**User Story:** As a developer, I want interactive prompts to be clear and helpful, so that I can easily initialize projects without consulting documentation.

#### Acceptance Criteria

1. WHEN prompting for initialization mode, THE CLI SHALL offer clear options: "existing" or "new"
2. WHEN prompting for form ID, THE CLI SHALL validate the input matches numeric format
3. WHEN prompting for form title, THE CLI SHALL validate the input is non-empty
4. WHEN prompting for schema file path, THE CLI SHALL provide a sensible default value
5. WHEN a prompt validation fails, THE CLI SHALL display an error message and re-prompt
6. WHEN all prompts complete successfully, THE CLI SHALL proceed with initialization
7. WHEN a user cancels a prompt (Ctrl+C), THE CLI SHALL exit gracefully without creating partial files

### Requirement 15: Cross-Platform Compatibility

**User Story:** As a developer, I want the CLI to work consistently across different operating systems, so that I can use the same commands regardless of my platform.

#### Acceptance Criteria

1. WHEN running on macOS, THE CLI SHALL use `open` command for browser launching
2. WHEN running on Linux, THE CLI SHALL use `xdg-open` command for browser launching
3. WHEN running on Windows, THE CLI SHALL use `start` command for browser launching
4. WHEN handling file paths, THE CLI SHALL use platform-appropriate path separators
5. WHEN creating directories, THE CLI SHALL work correctly on all supported platforms
6. WHEN reading environment variables, THE CLI SHALL handle platform-specific conventions
7. WHEN displaying output, THE CLI SHALL handle platform-specific line endings correctly
