# Development Guide

This guide provides information for contributors and developers working on the Liquibase Linter project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Project Architecture](#project-architecture)
- [Building](#building)
- [Testing](#testing)
- [Code Style](#code-style)
- [Contributing](#contributing)
- [Release Process](#release-process)

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Git
- Make (optional, but recommended)

### Clone the Repository

```bash
git clone https://github.com/n2jsoft/liquibase-linter.git
cd liquibase-linter
```

### Install Dependencies

```bash
go mod download
```

### Build the Project

```bash
go build -o liquibase-linter ./cmd/liquibase-linter
```

### Run Tests

```bash
go test ./...
```

## Development Environment

### Recommended Tools

- **Editor**: VS Code with Go extension
- **Linter**: golangci-lint
- **Formatter**: gofmt (built into Go)
- **Debugger**: Delve (built into VS Code Go extension)

### VS Code Setup

Recommended extensions:
- Go (golang.go)
- Error Lens
- GitLens

Workspace settings (`.vscode/settings.json`):
```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.testOnSave": true,
  "go.coverOnSave": true,
  "editor.formatOnSave": true
}
```

### Install golangci-lint

```bash
# macOS/Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# Or using Homebrew on macOS
brew install golangci-lint
```

## Project Architecture

### Directory Structure

```
liquibase-linter/
├── cmd/
│   └── liquibase-linter/      # Main application entry point
│       └── main.go             # CLI setup and command handling
├── internal/                   # Private application code
│   ├── config/                 # Configuration management
│   │   ├── config.go           # Config loading and validation
│   │   └── config_test.go      # Config tests
│   ├── parser/                 # Changelog parsers
│   │   ├── types.go            # Common types (Changelog, ChangeSet)
│   │   ├── xml.go              # XML parser
│   │   ├── sql.go              # SQL parser
│   │   └── *_test.go           # Parser tests
│   ├── rules/                  # Linting rules
│   │   ├── types.go            # Rule interface and registry
│   │   ├── security.go         # Security rules
│   │   ├── bestpractice.go     # Best practice rules
│   │   ├── performance.go      # Performance rules
│   │   └── *_test.go           # Rule tests
│   └── reporter/               # Output formatting
│       ├── types.go            # Reporter interface
│       ├── text.go             # Text output
│       ├── json.go             # JSON output
│       ├── sarif.go            # SARIF output
│       └── reporter_test.go    # Reporter tests
├── testdata/                   # Test fixtures
├── docs/                       # Documentation
├── scripts/                    # Build scripts
│   ├── build.sh                # Build script
│   └── test.sh                 # Test script
├── go.mod                      # Go module definition
├── go.sum                      # Go dependencies
└── README.md                   # Project readme
```

### Core Components

#### 1. Parser (`internal/parser`)

Responsible for reading and parsing Liquibase changelog files.

**Key files**:
- `types.go`: Defines `Changelog`, `ChangeSet`, `Change` types
- `xml.go`: XML parser implementation
- `sql.go`: SQL formatted changelog parser

**Adding a new parser**:
1. Create a new file (e.g., `yaml.go`)
2. Implement parsing function following the pattern:
   ```go
   func ParseYAML(filePath string) (*Changelog, error)
   ```
3. Add comprehensive tests in `yaml_test.go`
4. Update `types.go` if new types are needed

#### 2. Rules (`internal/rules`)

Defines and implements linting rules.

**Key files**:
- `types.go`: `Rule` interface, `Violation` type, rule registry
- `security.go`: Security-related rules
- `bestpractice.go`: Best practice rules
- `performance.go`: Performance rules

**Adding a new rule**:
1. Create a struct implementing the `Rule` interface:
   ```go
   type MyNewRule struct{}
   
   func (r *MyNewRule) ID() string { return "my-new-rule" }
   func (r *MyNewRule) Name() string { return "My New Rule" }
   func (r *MyNewRule) Description() string { return "Description" }
   func (r *MyNewRule) Severity() Severity { return SeverityWarning }
   func (r *MyNewRule) Check(changelog *Changelog) []Violation {
       // Implementation
   }
   ```
2. Register the rule in `init()`:
   ```go
   func init() {
       RegisterRule(&MyNewRule{})
   }
   ```
3. Add tests in the appropriate `*_test.go` file
4. Document the rule in `docs/rules.md`

#### 3. Reporter (`internal/reporter`)

Formats and outputs linting results.

**Key files**:
- `types.go`: `Reporter` interface
- `text.go`: Human-readable text output
- `json.go`: JSON output
- `sarif.go`: SARIF format output

**Adding a new output format**:
1. Create a new file (e.g., `junit.go`)
2. Implement the `Reporter` interface:
   ```go
   type JUnitReporter struct{}
   
   func (r *JUnitReporter) Report(results []RuleResult) error {
       // Implementation
   }
   ```
3. Register in `GetReporter()` function in `types.go`
4. Add tests

#### 4. Config (`internal/config`)

Manages configuration loading and validation.

**Key files**:
- `config.go`: Config struct, loading, and validation

## Building

### Local Build

```bash
# Build for current platform
go build -o liquibase-linter ./cmd/liquibase-linter

# Build with version information
go build -ldflags "-X main.Version=1.0.0" -o liquibase-linter ./cmd/liquibase-linter
```

### Cross-Platform Build

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o liquibase-linter-linux-amd64 ./cmd/liquibase-linter

# macOS
GOOS=darwin GOARCH=amd64 go build -o liquibase-linter-darwin-amd64 ./cmd/liquibase-linter

# Windows
GOOS=windows GOARCH=amd64 go build -o liquibase-linter-windows-amd64.exe ./cmd/liquibase-linter
```

### Using Build Script

```bash
./scripts/build.sh
```

This creates binaries in the `build/` directory for all platforms.

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run Tests for a Specific Package

```bash
go test ./internal/parser/
go test ./internal/rules/
```

### Run a Specific Test

```bash
go test -run TestParseXML ./internal/parser/
```

### Table-Driven Tests

Follow the standard Go table-driven test pattern:

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "TEST",
            wantErr: false,
        },
        {
            name:    "empty input",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("MyFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("MyFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Code Style

### Follow Effective Go

Read and follow [Effective Go](https://go.dev/doc/effective_go).

### Key Conventions

1. **Formatting**: Use `gofmt` (automatic in most editors)
2. **Naming**:
   - Exported: `PascalCase`
   - Unexported: `camelCase`
   - Acronyms: `HTTP`, `URL`, `ID` (not `Http`, `Url`, `Id`)
3. **Error Handling**: Always check errors explicitly
4. **Comments**: 
   - Package comments start with `// Package name...`
   - Exported symbols must have comments
   - Comments are complete sentences

### Run Linters

```bash
# Run all linters
golangci-lint run

# Run specific linters
golangci-lint run --enable=golint,gofmt,govet
```

### Example: Well-Formatted Code

```go
// Package parser provides functionality for parsing Liquibase changelog files.
package parser

import (
    "encoding/xml"
    "fmt"
    "os"
)

// ParseXML reads and parses an XML-formatted Liquibase changelog file.
// It returns a Changelog structure or an error if parsing fails.
func ParseXML(filePath string) (*Changelog, error) {
    data, err := os.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    var changelog Changelog
    if err := xml.Unmarshal(data, &changelog); err != nil {
        return nil, fmt.Errorf("failed to parse XML: %w", err)
    }
    
    return &changelog, nil
}
```

## Contributing

### Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Add tests for new functionality
5. Run tests: `go test ./...`
6. Run linters: `golangci-lint run`
7. Commit with descriptive message: `git commit -m "Add feature X"`
8. Push to your fork: `git push origin feature/my-feature`
9. Create a Pull Request

### Commit Messages

Follow conventional commits format:

```
type(scope): subject

body

footer
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test additions or changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

**Examples**:
```
feat(parser): add YAML parser support

Implement YAML parser for Liquibase changelogs using gopkg.in/yaml.v3.
Includes comprehensive tests and error handling.

Closes #123
```

```
fix(rules): correct SQL injection detection for prepared statements

The SQL injection rule was incorrectly flagging prepared statements.
Updated the pattern matching to exclude parameterized queries.

Fixes #456
```

### Pull Request Checklist

- [ ] Tests pass locally
- [ ] New functionality includes tests
- [ ] Code is formatted with `gofmt`
- [ ] Linters pass (`golangci-lint run`)
- [ ] Documentation is updated
- [ ] Commit messages follow convention
- [ ] PR description explains changes clearly

## Release Process

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):
- `MAJOR.MINOR.PATCH`
- Example: `1.2.3`

### Creating a Release

1. Update version in code
2. Update CHANGELOG.md
3. Create a git tag:
   ```bash
   git tag -a v1.2.3 -m "Release v1.2.3"
   git push origin v1.2.3
   ```
4. Build release binaries:
   ```bash
   ./scripts/build.sh
   ```
5. Create GitHub release with binaries

### Pre-release Checklist

- [ ] All tests pass
- [ ] Documentation is up to date
- [ ] CHANGELOG.md is updated
- [ ] Version number is bumped
- [ ] Build succeeds for all platforms

## Debugging

### Using Delve

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug ./cmd/liquibase-linter -- check testdata/valid-changelog.xml

# Set breakpoint
(dlv) break parser.ParseXML
(dlv) continue
```

### Logging

Add debug logging during development:

```go
import "log"

log.Printf("Debug: processing changeset %s", changeSet.ID)
```

## Performance Profiling

### CPU Profiling

```go
import _ "net/http/pprof"

// In main():
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Access profiles at http://localhost:6060/debug/pprof/

### Memory Profiling

```bash
go test -memprofile=mem.prof ./internal/parser/
go tool pprof mem.prof
```

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Liquibase Documentation](https://docs.liquibase.com/)

## Getting Help

- Open an issue on GitHub
- Check existing documentation in `docs/`
- Review existing code for patterns

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
