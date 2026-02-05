# Usage Guide

This guide provides detailed information on using the Liquibase Linter effectively.

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Command Reference](#command-reference)
- [Output Formats](#output-formats)
- [Working with Results](#working-with-results)
- [Best Practices](#best-practices)

## Installation

### Using Go Install

```bash
go install github.com/n2jsoft-public-org/liquibase-linter/cmd/liquibase-linter@latest
```

### Download Pre-built Binary

Download from the [releases page](https://github.com/n2jsoft-public-org/liquibase-linter/releases):

```bash
# Linux
curl -L -o liquibase-linter https://github.com/n2jsoft-public-org/liquibase-linter/releases/latest/download/liquibase-linter-linux-amd64
chmod +x liquibase-linter
sudo mv liquibase-linter /usr/local/bin/

# macOS
curl -L -o liquibase-linter https://github.com/n2jsoft-public-org/liquibase-linter/releases/latest/download/liquibase-linter-darwin-amd64
chmod +x liquibase-linter
sudo mv liquibase-linter /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/n2jsoft-public-org/liquibase-linter/releases/latest/download/liquibase-linter-windows-amd64.exe -OutFile liquibase-linter.exe
```

### Build from Source

```bash
git clone https://github.com/n2jsoft-public-org/liquibase-linter.git
cd liquibase-linter
go build -o liquibase-linter ./cmd/liquibase-linter
```

## Basic Usage

### Supported Changelog Formats

The linter supports all major Liquibase changelog formats:

- **XML** (`.xml`) - Traditional Liquibase format with full feature support
- **YAML** (`.yaml`, `.yml`) - YAML format with `include` and `includeAll` directives
- **JSON** (`.json`) - JSON format with `include` and `includeAll` directives
- **SQL** (`.sql`) - Formatted SQL with Liquibase comments

#### Include and IncludeAll Support

YAML and JSON changelogs support `include` and `includeAll` directives:

```yaml
databaseChangeLog:
  # Include a single file
  - include:
      file: changelog/v1.0.xml
  
  # Include all files in a directory
  - includeAll:
      path: changelog/migrations/
      resourceFilter: "**/*.sql"
```

**Features:**
- Mixed-format includes (YAML can include XML, SQL, or other YAML files)
- Recursive directory scanning with `includeAll`
- Resource filtering with glob patterns (e.g., `**/*.sql`, `v*/*.xml`)
- Circular include detection with symlink awareness
- Configurable maximum include depth (default: 10 levels)

**Current Limitations:**
- The `context` and `labels` attributes for `includeAll` are not yet supported and will be added in a future release
- All files matching the filter will be included regardless of context/labels

### Check a Single File

```bash
liquibase-linter check db/changelog/changelog-1.0.xml
```

### Check Multiple Files

```bash
# Check all XML files
liquibase-linter check db/changelog/*.xml

# Check all files in a directory
liquibase-linter check db/changelog/

# Check multiple specific files
liquibase-linter check file1.xml file2.sql file3.yaml
```

### Use Configuration File

```bash
liquibase-linter check --config=.liquibase-linter.yaml db/changelog/
```

## Command Reference

### check

Analyze Liquibase changelog files for issues.

```bash
liquibase-linter check [flags] <path>...
```

**Flags**:
- `--config=FILE`: Path to configuration file (default: `.liquibase-linter.yaml`)
- `--format=FORMAT`: Output format: `text`, `json`, `sarif`, `junit` (default: `text`)
- `--no-color`: Disable colored output
- `--severity=LEVEL`: Minimum severity to report: `info`, `warning`, `critical` (default: `info`)
- `--fail-on=LEVEL`: Exit with error on violations of this severity or higher (default: `critical`)

**Examples**:

```bash
# Basic check with default settings
liquibase-linter check db/changelog/

# Check with custom config
liquibase-linter check --config=my-config.yaml db/changelog/

# JSON output for CI/CD
liquibase-linter check --format=json db/changelog/ > results.json

# Only report critical issues
liquibase-linter check --severity=critical db/changelog/

# Fail on warnings or higher
liquibase-linter check --fail-on=warning db/changelog/
```

### rules

List available linting rules.

```bash
liquibase-linter rules [flags]
```

**Flags**:
- `--info=RULE_ID`: Show detailed information about a specific rule

**Examples**:

```bash
# List all rules
liquibase-linter rules

# Show details for a specific rule
liquibase-linter rules --info=sql-injection
```

### init

Initialize a configuration file with default settings.

```bash
liquibase-linter init [flags]
```

**Flags**:
- `--output=FILE`: Output file path (default: `.liquibase-linter.yaml`)
- `--force`: Overwrite existing configuration file

**Examples**:

```bash
# Create default config
liquibase-linter init

# Create custom config
liquibase-linter init --output=custom-config.yaml

# Overwrite existing config
liquibase-linter init --force
```

### version

Display version information.

```bash
liquibase-linter version
```

## Output Formats

### Text (Default)

Human-readable output with colors:

```
Checking: db/changelog/changelog-1.0.xml

[CRITICAL] sql-injection (line 15, changeset: create-user-1)
  Potential SQL injection: String concatenation detected in SQL statement
  
[WARNING] missing-rollback (line 10, changeset: add-column-1)
  Changeset does not include a rollback script

Summary:
  Files checked: 1
  Violations: 2 (1 critical, 1 warning, 0 info)
```

### JSON

Structured output for programmatic processing:

```json
{
  "version": "1.0.0",
  "timestamp": "2024-02-04T10:30:00Z",
  "files": [
    {
      "path": "db/changelog/changelog-1.0.xml",
      "violations": [
        {
          "rule": "sql-injection",
          "severity": "critical",
          "message": "Potential SQL injection: String concatenation detected",
          "line": 15,
          "changeset_id": "create-user-1"
        }
      ]
    }
  ],
  "summary": {
    "files_checked": 1,
    "total_violations": 2,
    "critical": 1,
    "warning": 1,
    "info": 0
  }
}
```

### SARIF

Static Analysis Results Interchange Format for IDE integration:

```json
{
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Liquibase Linter",
          "version": "1.0.0",
          "rules": [...]
        }
      },
      "results": [...]
    }
  ]
}
```

### JUnit

XML format for CI/CD systems:

```xml
<testsuite name="Liquibase Linter" tests="2" failures="2">
  <testcase name="sql-injection" classname="db.changelog.changelog-1.0.xml">
    <failure message="Potential SQL injection detected" type="critical"/>
  </testcase>
</testsuite>
```

## Working with Results

### Exit Codes

The linter uses standard exit codes to indicate results:

- `0`: No violations found (or only below fail threshold)
- `1`: Violations found at or above fail threshold
- `2`: Error during execution

### Filtering Results

Use `--severity` to filter what gets reported:

```bash
# Only show critical issues
liquibase-linter check --severity=critical db/changelog/

# Show warnings and critical (excludes info)
liquibase-linter check --severity=warning db/changelog/
```

Use `--fail-on` to control exit code behavior:

```bash
# Exit with error only on critical issues (default)
liquibase-linter check --fail-on=critical db/changelog/

# Exit with error on any warning or higher
liquibase-linter check --fail-on=warning db/changelog/

# Never exit with error (useful for reporting only)
liquibase-linter check --fail-on=none db/changelog/
```

### Saving Results

Redirect output to a file for later analysis:

```bash
# Text format
liquibase-linter check db/changelog/ > lint-results.txt

# JSON format
liquibase-linter check --format=json db/changelog/ > results.json

# SARIF format
liquibase-linter check --format=sarif db/changelog/ > results.sarif
```

## Best Practices

### 1. Run Early and Often

Integrate the linter into your development workflow:

```bash
# Pre-commit hook
liquibase-linter check --fail-on=critical db/changelog/

# Local development
liquibase-linter check db/changelog/
```

### 2. Use Configuration Files

Create a `.liquibase-linter.yaml` file in your repository root to maintain consistent settings across the team:

```yaml
rules:
  sql-injection:
    enabled: true
    severity: critical
  missing-rollback:
    enabled: true
    severity: warning

ignore:
  - "test/**/*.xml"

output:
  format: text
  colorize: true

severity_threshold: warning
```

### 3. Progressive Adoption

If you have existing changelogs with many violations:

1. Start with critical issues only:
   ```bash
   liquibase-linter check --severity=critical db/changelog/
   ```

2. Add exceptions for files you'll fix later:
   ```yaml
   ignore:
     - "db/changelog/legacy/*.xml"
   ```

3. Gradually enable more rules and reduce exceptions

### 4. CI/CD Integration

Always run the linter in your CI/CD pipeline:

```yaml
# GitHub Actions example
- name: Lint Liquibase Changelogs
  run: liquibase-linter check --format=json db/changelog/
```

See the [CI/CD Integration Guide](cicd.md) for detailed examples.

### 5. Document Exceptions

If you must ignore certain violations, document why:

```yaml
rules:
  missing-rollback:
    enabled: false  # Disabled because we use snapshot-based rollbacks

ignore:
  - "db/changelog/2024-01-*.xml"  # Legacy migrations, will be refactored in Q2
```

### 6. Review Reports Regularly

Schedule regular reviews of linting reports:

```bash
# Generate comprehensive report
liquibase-linter check --format=json db/changelog/ > report-$(date +%Y%m%d).json
```

### 7. Educate Your Team

Share common violations and their fixes:

- Include linter results in code reviews
- Create internal documentation with project-specific examples
- Run workshops on secure Liquibase practices

## Troubleshooting

### No Output

If the linter produces no output:

1. Check if files exist at the specified path
2. Verify file format is supported (XML, SQL, YAML, JSON)
3. Try with `--severity=info` to see all issues

### False Positives

If you encounter false positives:

1. Review the specific rule documentation
2. Add the file to `ignore` patterns if appropriate
3. Report the issue on GitHub with a minimal example

### Performance Issues

For large codebases:

1. Use specific file patterns instead of checking entire directories
2. Exclude test fixtures and generated files
3. Run checks in parallel using CI/CD job matrices

## Getting Help

- Report bugs: https://github.com/n2jsoft-public-org/liquibase-linter/issues
- Documentation: https://github.com/n2jsoft-public-org/liquibase-linter/docs
- Discussions: https://github.com/n2jsoft-public-org/liquibase-linter/discussions
