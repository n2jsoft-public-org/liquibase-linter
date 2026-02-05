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

# Minimum severity to report
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
