# Liquibase Linter

A command-line tool written in Go that analyzes Liquibase SQL scripts to identify security vulnerabilities, anti-patterns, and best practice violations.

## Features

- 🔒 **Security First**: Detect SQL injection risks, dangerous operations, and security anti-patterns
- ✅ **Code Quality**: Enforce Liquibase best practices and coding standards
- 🚀 **Fast**: Built in Go for optimal performance
- 📊 **Multiple Output Formats**: Support for text, JSON, SARIF, and JUnit formats
- 🔧 **Configurable**: Flexible rule configuration via YAML

## Installation

```bash
go install github.com/n2jsoft/liquibase-linter/cmd/liquibase-linter@latest
```

Or download pre-built binaries from the [releases page](https://github.com/n2jsoft/liquibase-linter/releases).

## Quick Start

```bash
# Check a single changelog file
liquibase-linter check db/changelog/db.changelog-master.xml

# Check all changelogs in a directory
liquibase-linter check db/changelog/

# Use a configuration file
liquibase-linter check --config=.liquibase-linter.yaml db/changelog/

# Output in JSON format
liquibase-linter check --format=json db/changelog/

# List all available rules
liquibase-linter rules

# Initialize a configuration file
liquibase-linter init
```

## Configuration

Create a `.liquibase-linter.yaml` file in your project root:

```yaml
rules:
  sql-injection:
    enabled: true
    severity: critical
  missing-rollback:
    enabled: true
    severity: warning
  
ignore:
  - "test/fixtures/*.xml"
  
output:
  format: text
  colorize: true
```

## Documentation

See the [docs](docs/) directory for detailed documentation:
- [Rules Reference](docs/rules.md)
- [Configuration Guide](docs/configuration.md)
- [CI/CD Integration](docs/cicd.md)

## Development

```bash
# Build
go build -o liquibase-linter ./cmd/liquibase-linter

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linters
golangci-lint run
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details.
