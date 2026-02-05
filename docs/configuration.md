# Configuration Guide

This guide explains how to configure the Liquibase Linter for your project.

## Configuration File

The linter uses a YAML configuration file (`.liquibase-linter.yaml` by default) to define rules, ignore patterns, and output settings.

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

Use glob patterns to exclude files from linting:

```yaml
ignore:
  - "test/**/*.xml"          # All XML files in test directories
  - "fixtures/**"            # Everything in fixtures directories
  - "db/changelog/old/*.sql" # Specific directory
```

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
