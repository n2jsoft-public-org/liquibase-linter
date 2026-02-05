# Configuration Guide

This guide explains how to configure the Liquibase Linter for your project.

## Configuration File

The linter uses a YAML configuration file (`.liquibase-linter.yaml` by default) to define rules, ignore patterns, and output settings.

### Auto-Discovery

The linter automatically searches for a configuration file in the following order:

1. **Specified config path** (if `--config` flag is provided)
2. **Target directory** (if checking a directory)
3. **Parent directory of target file** (if checking a single file)
4. **Current working directory**
5. **Parent directories** (walking up to the filesystem root)

The linter looks for these filenames:
- `.liquibase-linter.yaml`
- `.liquibase-linter.yml`

**Example:**
```bash
# Config file will be auto-discovered from the changelog directory
liquibase-linter check db/changelog/

# Or explicitly specify a config file
liquibase-linter check --config=.liquibase-linter.yaml db/changelog/
```

### Creating a Configuration File

```bash
# Create a configuration file with default settings
liquibase-linter init

# Create with a custom filename
liquibase-linter init --output=custom-config.yaml
```

### Configuration Structure

```yaml
# Rules configuration
rules:
  sql-injection:
    enabled: true
    severity: critical
  hardcoded-credentials:
    enabled: true
    severity: critical
  dangerous-operations:
    enabled: true
    severity: critical
  missing-rollback:
    enabled: true
    severity: warning
  non-idempotent:
    enabled: true
    severity: warning
    mode: "risky-only"  # Options: "risky-only" (default), "all"
    exclude_patterns:
      - "**/init/**"
      - "**/seed/**"

# File patterns to ignore
ignore:
  - "test/**/*.xml"
  - "fixtures/**/*.sql"

# Output configuration
output:
  format: text        # Options: text, json, sarif, junit
  colorize: true      # Enable colored output (text format only)

# Parser behavior configuration
parser:
  # Maximum depth for nested include/includeAll directives (1-100)
  max_include_depth: 10
  
  # Follow symlinks during file discovery
  # Symlink loops are automatically detected and prevented
  follow_symlinks: true

# File structure rules configuration
file_structure:
  # Enable file structure validation
  enabled: false  # Disabled by default
  
  # Sprint folder pattern (regex)
  # Default matches v116, v117, etc.
  sprint_pattern: "^v\\d+$"
  
  # Structure folder pattern (regex, case-insensitive)
  # Default matches "0 - structure", "1-structure", etc.
  structure_pattern: "(?i)^\\d+\\s*-\\s*structure$"
  
  # Data folder pattern (regex, case-insensitive)
  # Default matches "1 - data", "2-data", etc.
  data_pattern: "(?i)^\\d+\\s*-\\s*data$"
  
  # Patterns to exclude from file structure rules
  # Use glob patterns with ** for recursive matching
  exclude_patterns:
    - "**/init/**"  # Exclude initialization scripts
    - "**/legacy/**"  # Optional: exclude legacy migrations
  
  # Optional: specify base path for sprint folders
  # If empty, sprints can be anywhere
  sprint_base_path: "changelog/sprints"

# Minimum severity to report
severity_threshold: warning  # Options: info, warning, critical
```

## Rule Configuration

Each rule can be configured with the following properties:

- **enabled**: Whether the rule is active (default: true)
- **severity**: Rule severity level - `info`, `warning`, or `critical`
- **mode**: (non-idempotent only) Enforcement mode - `risky-only` or `all`
- **exclude_patterns**: (non-idempotent only) Glob patterns to exclude from checks

### Available Rules

- **sql-injection**: Detect SQL injection vulnerabilities
- **hardcoded-credentials**: Find hardcoded passwords and API keys
- **dangerous-operations**: Detect DROP/TRUNCATE without preconditions
- **missing-rollback**: Ensure changesets have rollback scripts
- **non-idempotent**: Detect non-idempotent changes (supports modes and exclude patterns)
- **sprint-folder-structure**: Enforce sprint-based folder organization
- **ddl-location**: Ensure DDL changes are in structure directories
- **dml-location**: Ensure DML changes are in data directories

### Non-Idempotent Rule Configuration

The `non-idempotent` rule supports additional configuration options:

```yaml
rules:
  non-idempotent:
    enabled: true
    severity: warning
    mode: "risky-only"  # or "all"
    exclude_patterns:
      - "**/init/**"
      - "**/seed/**"
```

**Mode Options:**
- **risky-only** (default): Only checks specific risky operations (CREATE TABLE, ADD COLUMN, etc.)
- **all**: Requires preconditions on every changeset (strictest enforcement)

**Exclude Patterns:**
- Use glob patterns with `**` for recursive directory matching
- Default patterns: `**/init/**`, `**/seed/**`
- Useful for exempting initialization scripts, seed data, and test fixtures

For more details, see the [non-idempotent rule documentation](rules/non-idempotent.md).

## File Structure Configuration

The file structure rules help enforce consistent organization of database changes across sprints.

### Enabling File Structure Rules

```yaml
file_structure:
  enabled: true  # Must be explicitly enabled
```

### Sprint Pattern Customization

Customize the regex pattern to match your sprint naming convention:

```yaml
file_structure:
  # Examples:
  sprint_pattern: "^v\\d+$"           # Matches: v116, v117, v200
  sprint_pattern: "^sprint-\\d+$"     # Matches: sprint-116, sprint-200
  sprint_pattern: "^release_\\d+\\.\\d+$"  # Matches: release_1.0, release_2.5
```

### Folder Organization Patterns

Configure how structure and data folders should be named:

```yaml
file_structure:
  # With numbers (default)
  structure_pattern: "(?i)^\\d+\\s*-\\s*structure$"  # Matches: "0 - structure", "1-structure"
  data_pattern: "(?i)^\\d+\\s*-\\s*data$"            # Matches: "1 - data", "2-data"
  
  # Without numbers
  structure_pattern: "^structure$"  # Matches: "structure"
  data_pattern: "^data$"            # Matches: "data"
  
  # Custom names
  structure_pattern: "^ddl$"  # Matches: "ddl"
  data_pattern: "^dml$"       # Matches: "dml"
```

### Exclude Patterns

Exclude certain paths from file structure validation:

```yaml
file_structure:
  exclude_patterns:
    - "**/init/**"      # Initialization scripts
    - "**/legacy/**"    # Legacy migrations
    - "**/hotfix/**"    # Emergency hotfixes
    - "**/rollback/**"  # Rollback scripts
```

### Example Organization

With the default configuration:

```
db/changelog/
  init/                      # ✅ Excluded from validation
    0 - structure/
      001-schema.sql
    1 - data/
      002-seed.sql
  sprints/
    v116/                    # ✅ Matches sprint pattern
      0 - structure/         # ✅ Matches structure pattern
        create_users.sql     # ✅ DDL in structure folder
      1 - data/              # ✅ Matches data pattern
        insert_users.sql     # ✅ DML in data folder
    v117/
      0 - structure/
        alter_users.sql      # ✅ DDL in structure folder
```

### Common Validation Scenarios

**Scenario 1: Sprint folder structure**
```yaml
# File: changelog/hotfix/tables.sql
# ❌ Violation: Not in a sprint folder
```

**Scenario 2: DDL in wrong location**
```yaml
# File: changelog/sprints/v116/1 - data/create_table.sql
# ❌ Violation: DDL (CREATE TABLE) in data directory
```

**Scenario 3: DML in wrong location**
```yaml
# File: changelog/sprints/v116/0 - structure/insert_data.sql
# ❌ Violation: DML (INSERT) in structure directory
```

## Minimum severity to report
severity_threshold: warning  # Options: info, warning, critical
```

## Rule Configuration

Each rule can be configured with the following properties:

- **enabled**: Whether the rule is active (default: true)
- **severity**: Rule severity level - `info`, `warning`, or `critical`

### Available Rules (Phase 3)

Rules will be implemented in Phase 3. Planned rules include:

- **sql-injection**: Detect SQL injection vulnerabilities
- **hardcoded-credentials**: Find hardcoded passwords and API keys
- **dangerous-operations**: Detect DROP/TRUNCATE without preconditions
- **missing-rollback**: Ensure changesets have rollback scripts
- **non-idempotent**: Detect non-idempotent changes

## Command-Line Overrides

Configuration can be overridden using command-line flags:

```bash
# Override output format
liquibase-linter check --format=json db/changelog/

# Disable colors
liquibase-linter check --no-color db/changelog/

# Change severity threshold
liquibase-linter check --severity=critical db/changelog/

# Use custom config file
liquibase-linter check --config=.custom-config.yaml db/changelog/
```

## Ignore Patterns

Use glob patterns to exclude files from linting. Patterns are matched against paths **relative to the target directory** being checked.

```yaml
ignore:
  - "test/**/*.xml"          # All XML files in test directories
  - "fixtures/**"            # Everything in fixtures directories  
  - "init/**"                # Ignore entire init folder
  - "db/changelog/old/*.sql" # Specific directory with .sql files
  - "**/*-rollback.sql"      # All files ending in -rollback.sql at any depth
```

### Pattern Matching Rules

- **Patterns are relative**: Matched against paths relative to the directory being checked
- **`**` matches recursively**: Use `init/**` to match all files in init/ at any depth
- **`*` matches within a directory**: Use `*.xml` to match XML files in a single directory
- **Directory matching**: Patterns like `init/**` match all files within that directory tree

**Examples:**

If your changelog structure is:
```
changelog/
  .liquibase-linter.yaml
  init/
    001-schema.sql
    002-data.sql
  sprints/
    v1/
      001-feature.sql
```

And you run: `liquibase-linter check changelog/`

Then the pattern `init/**` will match:
- ✅ `init/001-schema.sql`
- ✅ `init/002-data.sql`
- ❌ `sprints/v1/001-feature.sql`

## Parser Configuration

### Max Include Depth

Control the maximum nesting level for `include` and `includeAll` directives:

```yaml
parser:
  max_include_depth: 10  # Must be between 1 and 100
```

This prevents infinite loops and excessively deep include hierarchies. If your changelog structure legitimately requires deeper nesting, you can increase this value.

**Common scenarios:**
- Default (10): Sufficient for most projects with typical include hierarchies
- Shallow (5): Recommended for simpler projects to catch accidental deep nesting
- Deep (20+): May be needed for large monorepos with complex changelog organization

### Symlink Following

Control whether symlinks are followed during file discovery:

```yaml
parser:
  follow_symlinks: true  # Default: true
```

When enabled:
- Symlinked files and directories are processed
- Symlink loops are automatically detected and prevented
- Resolved paths are tracked to avoid duplicate processing

When disabled:
- Symlinks are skipped entirely
- Useful if symlinks cause issues in your build environment
- Slightly faster for projects without symlinks

## Output Formats

### Text (Default)

Human-readable output with optional colors:

```bash
liquibase-linter check db/changelog/
```

### JSON

Structured output for tooling integration:

```bash
liquibase-linter check --format=json db/changelog/
```

### SARIF

Static Analysis Results Interchange Format for IDE integration:

```bash
liquibase-linter check --format=sarif db/changelog/
```

### JUnit

XML format for CI/CD systems:

```bash
liquibase-linter check --format=junit db/changelog/
```

## File Structure Configuration

Configure file organization rules for sprint-based development workflows.

### Overview

The file structure configuration enforces a consistent directory organization for your changelog files:

```
db/
  changelog/
    sprints/
      v116/              # Sprint version folder
        0 - structure/   # DDL changes (CREATE, ALTER, DROP)
        1 - data/        # DML changes (INSERT, UPDATE, DELETE)
      v117/
        0 - structure/
        1 - data/
    init/                # Excluded from validation
      tables/
      data/
```

### Configuration Options

```yaml
file_structure:
  # Sprint folder pattern (regex)
  # Default: ^v\d+$ (matches v116, v117, etc.)
  sprint_pattern: "^v\\d+$"
  
  # Structure folder pattern (regex)
  # Default: ^\d+ - structure$ (matches "0 - structure", "1 - structure")
  structure_pattern: "^\\d+ - structure$"
  
  # Data folder pattern (regex)
  # Default: ^\d+ - data$ (matches "0 - data", "1 - data")
  data_pattern: "^\\d+ - data$"
  
  # Glob patterns for paths to exclude from validation
  # Default: ["**/init/**"]
  exclude_patterns:
    - "**/init/**"
    - "**/legacy/**"
```

### Pattern Examples

#### Sprint Patterns

```yaml
file_structure:
  # Pattern: v116, v117, v118
  sprint_pattern: "^v\\d+$"
  
  # Pattern: sprint-116, sprint-117
  sprint_pattern: "^sprint-\\d+$"
  
  # Pattern: v1.16, v1.17
  sprint_pattern: "^v\\d+\\.\\d+$"
  
  # Pattern: 2024-Q1, 2024-Q2
  sprint_pattern: "^\\d{4}-Q[1-4]$"
```

#### Folder Patterns

```yaml
file_structure:
  # Numbered folders with hyphens: "0 - structure", "1 - data"
  structure_pattern: "^\\d+ - structure$"
  data_pattern: "^\\d+ - data$"
  
  # Simple folder names: "structure", "data"
  structure_pattern: "^structure$"
  data_pattern: "^data$"
  
  # Numbered without spaces: "0-structure", "1-data"
  structure_pattern: "^\\d+-structure$"
  data_pattern: "^\\d+-data$"
```

#### Exclude Patterns

```yaml
file_structure:
  exclude_patterns:
    # Exclude all files in init directories
    - "**/init/**"
    
    # Exclude legacy migration paths
    - "**/legacy/**"
    - "**/archive/**"
    
    # Exclude hotfix directories (not sprint-based)
    - "**/hotfix/**"
    - "**/emergency/**"
    
    # Exclude test fixtures
    - "**/test/**"
    - "**/fixtures/**"
```

### Use Cases

#### Agile Sprint-Based Development

Organize changes by sprint with clear separation of DDL and DML:

```yaml
file_structure:
  sprint_pattern: "^v\\d+$"
  structure_pattern: "^0 - structure$"
  data_pattern: "^1 - data$"
  exclude_patterns:
    - "**/init/**"
```

#### Quarterly Releases

Organize by quarters:

```yaml
file_structure:
  sprint_pattern: "^\\d{4}-Q[1-4]$"  # 2024-Q1, 2024-Q2, etc.
  structure_pattern: "^structure$"
  data_pattern: "^data$"
  exclude_patterns:
    - "**/baseline/**"
```

#### Migration from Legacy Structure

Gradually adopt the structure by excluding old directories:

```yaml
file_structure:
  sprint_pattern: "^v\\d+$"
  structure_pattern: "^\\d+ - structure$"
  data_pattern: "^\\d+ - data$"
  exclude_patterns:
    - "**/init/**"
    - "**/legacy/**"
    - "**/v1/**"   # Old version format
    - "**/v2/**"
```

### Related Rules

The file structure configuration controls three rules:

- **file-structure-sprint**: Ensures files are in sprint folders
- **file-structure-ddl**: Ensures DDL changes are in structure directories
- **file-structure-dml**: Ensures DML changes are in data directories

Enable or disable these rules individually:

```yaml
rules:
  file-structure-sprint:
    enabled: true
    severity: critical
  
  file-structure-ddl:
    enabled: true
    severity: critical
  
  file-structure-dml:
    enabled: true
    severity: critical
```

### Benefits

- **Clear Organization**: Easy to find and review changes by sprint
- **Separation of Concerns**: DDL and DML changes are clearly separated
- **Release Management**: Simple to identify changes for specific releases
- **Rollback Strategy**: Easier to rollback specific sprints or change types
- **Team Collaboration**: Reduces merge conflicts and improves clarity
- **Audit Trail**: Clear history of when and why changes were made

## Environment Variables

(To be implemented)

- `LIQUIBASE_LINTER_CONFIG`: Default configuration file path
- `LIQUIBASE_LINTER_NO_COLOR`: Disable colored output

## Examples

### Basic Usage

```bash
# Check with default configuration
liquibase-linter check db/changelog/

# Check with custom config
liquibase-linter check --config=.liquibase-linter.yaml db/changelog/
```

### CI/CD Integration

```bash
# Exit with error on warnings or higher
liquibase-linter check --format=json --severity=warning db/changelog/

# Only fail on critical issues
liquibase-linter check --severity=critical db/changelog/
```

## Next Steps

- Learn about [rules](rules.md)
- Set up [CI/CD integration](cicd.md)
