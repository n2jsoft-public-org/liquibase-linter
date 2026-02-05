# Liquibase Linter

A fast, security-focused command-line tool written in Go that analyzes Liquibase changelog files to identify security vulnerabilities, performance issues, and best practice violations.

## Why Liquibase Linter?

Database migrations are critical infrastructure code that often contains security vulnerabilities and anti-patterns. Liquibase Linter helps teams:

- **Prevent security incidents** by detecting SQL injection, hardcoded credentials, and dangerous operations before they reach production
- **Improve reliability** by ensuring changesets have rollback scripts and proper preconditions
- **Optimize performance** by identifying missing indexes and problematic operations
- **Maintain code quality** through automated enforcement of best practices

## Features

- 🔒 **Security First**: Detect SQL injection risks, dangerous operations, hardcoded credentials, and privilege escalation
- ✅ **Code Quality**: Enforce Liquibase best practices, naming conventions, and proper documentation
- ⚡ **Performance**: Identify missing indexes, table locks, and large data operations
- 🚀 **Fast**: Built in Go for optimal performance (<100ms startup for typical changelogs)
- 📊 **Multiple Output Formats**: Support for text, JSON, SARIF, and JUnit formats
- 🔧 **Configurable**: Flexible rule configuration via YAML
- 🎯 **Inline Suppressions**: Disable specific rules for individual changesets via comments
- 🔌 **CI/CD Ready**: Easy integration with GitHub Actions, GitLab CI, Jenkins, and more

## Installation

```bash
go install github.com/n2jsoft-public-org/liquibase-linter/cmd/liquibase-linter@latest
```

Or download pre-built binaries from the [releases page](https://github.com/n2jsoft-public-org/liquibase-linter/releases).

## Quick Start

```bash
# Check a single changelog file
liquibase-linter check db/changelog/db.changelog-master.xml

# Check all changelogs in a directory
liquibase-linter check db/changelog/

# Use a configuration file
liquibase-linter check --config=.liquibase-linter.yaml db/changelog/

# Output in JSON format for CI/CD
liquibase-linter check --format=json db/changelog/ > results.json

# List all available rules
liquibase-linter rules

# Show details about a specific rule
liquibase-linter rules --info=sql-injection

# Initialize a configuration file
liquibase-linter init
```

### Suppressing Rules for Specific Changesets

Sometimes you need to intentionally bypass a rule for a specific changeset. Add an inline suppression to the changeset comment:

```xml
<changeSet id="1" author="john">
    <comment>liquibase-linter:disable sql-injection</comment>
    <sql>
        -- Known safe parameter from validated config
        INSERT INTO settings VALUES (${config_value});
    </sql>
</changeSet>
```

See [Suppression Documentation](docs/suppression.md) for complete details and format-specific examples.

## Example Output

### Text Format (Default)

```
Checking: db/changelog/changelog-1.0.xml

[CRITICAL] sql-injection (line 15, changeset: create-user-1)
  Potential SQL injection: String concatenation detected in SQL statement
  Use parameterized queries or Liquibase's built-in change types instead.

[WARNING] missing-rollback (line 10, changeset: add-column-1)
  Changeset does not include a rollback script
  Add a <rollback> block to enable safe rollback of this change.

[INFO] naming-conventions (line 25, changeset: create-table-1)
  Table name 'UserData' does not follow naming convention
  Use lowercase snake_case for table names: user_data

Summary:
  Files checked: 1
  Violations: 3 (1 critical, 1 warning, 1 info)
```

### JSON Format

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
    "total_violations": 3,
    "critical": 1,
    "warning": 1,
    "info": 1
  }
}
```

## Configuration

The linter automatically discovers `.liquibase-linter.yaml` configuration files in:
1. The target directory (if checking a directory)
2. Parent directories up to the project root
3. Current working directory

You can also explicitly specify a config file with `--config`.

Create a `.liquibase-linter.yaml` file in your changelog directory or project root:

```yaml
# Enable/disable rules and set severity levels
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
  naming-conventions:
    enabled: true
    severity: info

# Ignore specific files or patterns (relative to target directory)
ignore:
  - "test/fixtures/*.xml"
  - "init/**"                 # Ignore all files in init folder
  - "db/changelog/legacy/**"

# Output configuration
output:
  format: text        # text, json, sarif, junit
  colorize: true      # Enable colored output (text format only)

# Parser configuration
parser:
  max_include_depth: 10    # Maximum nesting for include/includeAll
  follow_symlinks: true    # Follow symlinks during file discovery

# Minimum severity to report
severity_threshold: warning  # info, warning, critical
```

See [Configuration Guide](docs/configuration.md) for all options.

## Documentation

See the [docs](docs/) directory for detailed documentation:

- **[Usage Guide](docs/usage.md)**: Comprehensive guide to using the linter
- **[Rules Reference](docs/rules.md)**: Complete list of available rules with examples
- **[Configuration Guide](docs/configuration.md)**: Detailed configuration options
- **[Suppression Guide](docs/suppression.md)**: How to suppress rules for specific changesets
- **[CI/CD Integration](docs/cicd.md)**: Integration examples for various CI/CD platforms
- **[Development Guide](docs/development.md)**: Guide for contributors and developers

## Rules

The linter includes rules in the following categories:

### Security Rules (Critical)
- **sql-injection**: Detect potential SQL injection vulnerabilities
- **hardcoded-credentials**: Find hardcoded passwords and API keys
- **dangerous-operations**: Detect DROP/TRUNCATE without preconditions
- **privilege-escalation**: Identify excessive privilege grants

### Reliability Rules (Warning)
- **missing-rollback**: Ensure changesets have rollback scripts
- **non-idempotent**: Detect non-idempotent changes
- **missing-preconditions**: Ensure risky operations have safety checks

### Performance Rules (Info/Warning)
- **missing-indexes**: Detect missing indexes on foreign keys
- **table-locks**: Identify operations that may cause table locks
- **large-data-operations**: Detect unbounded data manipulation

### Best Practice Rules (Info)
- **naming-conventions**: Enforce consistent naming standards
- **changelog-organization**: Ensure proper changelog structure
- **documentation**: Verify changesets are documented

See [Rules Reference](docs/rules.md) for detailed descriptions and examples.

## CI/CD Integration

### GitHub Actions

```yaml
name: Liquibase Linting

on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Download Liquibase Linter
      run: |
        curl -L -o liquibase-linter https://github.com/n2jsoft-public-org/liquibase-linter/releases/latest/download/liquibase-linter-linux-amd64
        chmod +x liquibase-linter
    
    - name: Run Liquibase Linter
      run: ./liquibase-linter check --format=sarif db/changelog/ > results.sarif
    
    - name: Upload results
      uses: github/codeql-action/upload-sarif@v3
      with:
        sarif_file: results.sarif
```

### GitLab CI

```yaml
liquibase-lint:
  stage: test
  script:
    - curl -L -o liquibase-linter https://github.com/n2jsoft-public-org/liquibase-linter/releases/latest/download/liquibase-linter-linux-amd64
    - chmod +x liquibase-linter
    - ./liquibase-linter check --format=json db/changelog/
```

See [CI/CD Integration Guide](docs/cicd.md) for more examples.

## Development

### Building from Source

```bash
# Clone the repository
git clone https://github.com/n2jsoft-public-org/liquibase-linter.git
cd liquibase-linter

# Build
go build -o liquibase-linter ./cmd/liquibase-linter

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linters
golangci-lint run
```

See [Development Guide](docs/development.md) for detailed information on contributing.

## Supported Formats

- ✅ XML changelogs
- ✅ SQL formatted changelogs
- ⏳ YAML changelogs (planned)
- ⏳ JSON changelogs (planned)

## Performance

Liquibase Linter is designed for speed:

- Starts in <100ms for typical changelogs
- Processes large changelog directories in seconds
- Uses Go's efficient concurrency for multi-file analysis
- Minimal memory footprint

## Exit Codes

- `0`: No violations found
- `1`: Violations found (based on `fail_on` severity threshold)
- `2`: Error during execution (file not found, parse error, etc.)

## Examples

Check out the [testdata](testdata/) directory for example changelogs:

- [example-good-practices.xml](testdata/example-good-practices.xml): Well-structured changelog with best practices
- [example-good-practices.sql](testdata/example-good-practices.sql): SQL formatted changelog with good practices
- [problematic-changelog.xml](testdata/problematic-changelog.xml): Examples of common issues
- [problematic-changelog.sql](testdata/problematic-changelog.sql): SQL examples with violations

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please see:

- [Development Guide](docs/development.md) for setup and architecture
- [Open Issues](https://github.com/n2jsoft-public-org/liquibase-linter/issues) for things to work on
- [Pull Request Guidelines](docs/development.md#contributing) for the workflow

## Acknowledgments

Built with ❤️ by [n2jsoft](https://github.com/n2jsoft)

Inspired by the need for better security and quality control in database migrations.

## Support

- 📝 [Documentation](docs/)
- 🐛 [Report Issues](https://github.com/n2jsoft-public-org/liquibase-linter/issues)
- 💬 [Discussions](https://github.com/n2jsoft-public-org/liquibase-linter/discussions)
- ⭐ Star the project if you find it useful!
