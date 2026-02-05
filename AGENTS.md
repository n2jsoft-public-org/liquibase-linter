# AGENTS.md - Liquibase Linter

## Project Overview

**Liquibase Linter** is a command-line tool written in Go that analyzes Liquibase SQL scripts to identify security vulnerabilities, anti-patterns, and best practice violations. The linter helps teams maintain secure and high-quality database migration code.

## Project Goals

- **Security First**: Detect SQL injection risks, dangerous operations, and security anti-patterns
- **Code Quality**: Enforce Liquibase best practices and coding standards
- **Human Readable**: Maintain clear, understandable code with comprehensive documentation
- **Go Best Practices**: Follow standard Go conventions and idioms
- **CLI Excellence**: Provide an intuitive, fast, and reliable command-line experience

## Technology Stack

- **Language**: Go 1.25+
- **Architecture**: CLI application
- **Testing**: Go standard testing library with table-driven tests
- **Build**: Go modules for dependency management

## Project Structure (Standard Go Layout)

```
liquibase-linter/
├── cmd/
│   └── liquibase-linter/      # Main application entry point
│       └── main.go
├── internal/                   # Private application code
│   ├── parser/                 # Liquibase changelog parser
│   ├── rules/                  # Linting rules and validators
│   ├── reporter/               # Output formatting and reporting
│   └── config/                 # Configuration management
├── pkg/                        # Public libraries (if needed)
├── testdata/                   # Test fixtures and sample changelogs
├── docs/                       # Documentation
├── scripts/                    # Build and utility scripts
├── .github/                    # GitHub workflows (if applicable)
├── go.mod
├── go.sum
├── README.md
├── AGENTS.md                   # This file
├── LICENSE
└── .gitignore
```

## Core Components

### 1. Parser (`internal/parser`)

**Purpose**: Parse Liquibase changelog files (XML, YAML, JSON, SQL formats)

**Responsibilities**:
- Read and validate Liquibase changelog files
- Build an AST (Abstract Syntax Tree) representation
- Support all Liquibase formats: XML, YAML, JSON, and formatted SQL
- Extract changesets, preconditions, and rollback scripts

**Key Types**:
```go
type Changelog struct {
    FilePath    string
    Format      ChangelogFormat
    ChangeSets  []ChangeSet
}

type ChangeSet struct {
    ID          string
    Author      string
    Changes     []Change
    Rollback    *Rollback
    Context     string
    Labels      []string
}
```

### 2. Rules Engine (`internal/rules`)

**Purpose**: Define and execute linting rules

**Responsibilities**:
- Register and manage linting rules
- Execute rules against parsed changelogs
- Collect and categorize violations
- Support custom rule plugins

**Rule Categories**:
- **Security**: SQL injection, privilege escalation, password exposure
- **Performance**: Missing indexes, table locks, large data operations
- **Reliability**: Missing rollback, invalid preconditions, non-idempotent changes
- **Best Practices**: Naming conventions, changelog organization, documentation

**Key Types**:
```go
type Rule interface {
    ID() string
    Name() string
    Description() string
    Severity() Severity
    Check(changelog *Changelog) []Violation
}

type Violation struct {
    Rule        string
    Severity    Severity
    Message     string
    FilePath    string
    LineNumber  int
    ChangeSetID string
}

type Severity int
const (
    SeverityInfo Severity = iota
    SeverityWarning
    SeverityCritical
)
```

### 3. Reporter (`internal/reporter`)

**Purpose**: Format and output linting results

**Responsibilities**:
- Generate human-readable reports
- Support multiple output formats (text, JSON, SARIF, JUnit)
- Colorized terminal output
- Summary statistics

**Output Formats**:
- **Text**: Human-readable console output (default)
- **JSON**: Structured output for tooling integration
- **SARIF**: Static Analysis Results Interchange Format for IDE integration
- **JUnit**: XML format for CI/CD systems

### 4. Config (`internal/config`)

**Purpose**: Manage linter configuration

**Responsibilities**:
- Load configuration from files (.liquibase-linter.yaml)
- Support command-line flag overrides
- Rule enabling/disabling
- Severity threshold configuration
- Custom rule paths

**Configuration Example**:
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

## CLI Design

### Commands

```bash
# Basic usage
liquibase-linter check <path-to-changelog>

# Check multiple files/directories
liquibase-linter check db/changelog/*.xml

# With specific output format
liquibase-linter check --format=json db/changelog/

# With configuration file
liquibase-linter check --config=.liquibase-linter.yaml db/changelog/

# List available rules
liquibase-linter rules

# Show rule details
liquibase-linter rules --info sql-injection

# Initialize configuration file
liquibase-linter init
```

### Exit Codes

- `0`: No violations found
- `1`: Violations found (based on severity threshold)
- `2`: Error during execution (file not found, parse error, etc.)

## Development Guidelines for AI Agents

### Code Style

1. **Follow Effective Go**: https://go.dev/doc/effective_go
2. **Use `gofmt`**: All code must be formatted with `gofmt`
3. **Run `golint`**: Code should pass golint checks
4. **Error Handling**: Always check and handle errors explicitly
5. **Naming Conventions**:
   - Use camelCase for unexported names
   - Use PascalCase for exported names
   - Prefer short, descriptive names in limited scopes
   - Use full descriptive names for package-level declarations

### Testing Requirements

1. **Test Coverage**: Aim for >80% code coverage
2. **Table-Driven Tests**: Use table-driven test pattern for multiple test cases
3. **Test Files**: Place test files alongside source files (`*_test.go`)
4. **Test Data**: Store fixtures in `testdata/` directory
5. **Naming**: Test functions should be named `TestFunctionName_Scenario`

**Example Test Pattern**:
```go
func TestParser_ParseXML_ValidChangelog(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Changelog
        wantErr bool
    }{
        {
            name:  "simple changeset",
            input: "testdata/simple.xml",
            want:  &Changelog{/* ... */},
            wantErr: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseXML(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseXML() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("ParseXML() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Error Handling

1. **Use error types**: Define custom error types for domain-specific errors
2. **Wrap errors**: Use `fmt.Errorf` with `%w` to wrap errors with context
3. **Sentinel errors**: Use `errors.New` for predefined errors
4. **Error checking**: Never ignore errors

**Example**:
```go
var ErrInvalidChangelog = errors.New("invalid changelog format")

func ParseChangelog(path string) (*Changelog, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read changelog: %w", err)
    }
    // ...
}
```

### Package Organization

1. **internal/**: Code that should not be imported by other projects
2. **pkg/**: Reusable libraries (use sparingly, prefer internal/)
3. **cmd/**: One subdirectory per executable
4. **Avoid circular dependencies**: Design packages to depend only on lower-level packages
5. **Keep packages focused**: Each package should have a single, clear purpose

### Documentation

1. **Package Documentation**: Every package must have a package comment
2. **Exported Symbols**: All exported functions, types, constants must have comments
3. **Comment Format**: Comments should be complete sentences starting with the symbol name
4. **Examples**: Provide example tests for complex functions

**Example**:
```go
// Package parser provides functionality for parsing Liquibase changelog files
// in various formats including XML, YAML, JSON, and SQL.
package parser

// ParseXML reads and parses an XML-formatted Liquibase changelog file.
// It returns a Changelog structure or an error if parsing fails.
func ParseXML(path string) (*Changelog, error) {
    // Implementation
}
```

### Performance Considerations

1. **Avoid premature optimization**: Write clear code first
2. **Use benchmarks**: Create benchmark tests for critical paths
3. **Profile when needed**: Use `pprof` to identify bottlenecks
4. **Efficient parsing**: Use streaming parsers for large files
5. **Concurrency**: Consider concurrent processing for multiple files

## Implementation Phases

### Phase 1: Foundation ✅ COMPLETED
- [x] Set up standard Go project layout
- [x] Implement basic CLI with flag package
- [x] Create configuration loading system
- [x] Set up testing framework and CI/CD

### Phase 2: Core Parser ✅ COMPLETED
- [x] Implement XML parser for Liquibase changelogs
- [x] Implement SQL format parser
- [x] Add comprehensive parser tests

### Phase 3: Rules Engine ✅ COMPLETED
- [x] Design rule interface and registry
- [x] Implement basic security rules (SQL injection detection)
- [x] Implement best practice rules (naming conventions)
- [x] Implement performance rules (missing indexes)
- [x] Add rule unit tests

### Phase 4: Reporter ✅ COMPLETED
- [x] Implement text output formatter
- [x] Implement JSON output formatter
- [x] Add colorized terminal output
- [x] Implement SARIF format for IDE integration
- [x] Add summary statistics

### Phase 5: Documentation ✅ COMPLETED
- [x] Add comprehensive documentation
- [x] Create example changelogs and test cases

### Phase 6: Polish
- [ ] Performance optimization (parallel parsing for large includeAll directories)
- [ ] Add context/labels filtering support for includeAll directives
- [ ] Release preparation (versioning, changelog)

### Phase 7: Extend ✅ COMPLETED
- [x] Implement YAML parser with include/includeAll support
- [x] Implement JSON parser with include/includeAll support
- [x] Add resourceFilter pattern matching for includeAll
- [x] Implement circular include detection with symlink awareness
- [x] Add configurable max include depth
- [x] Support mixed-format includes (YAML→XML→SQL)

### Phase 8: File Structure Organization ✅ COMPLETED
- [x] Add FileStructureConfig to configuration system with regex patterns
- [x] Implement IsDDLChange() and IsDMLChange() helpers in parser
- [x] Create SprintFolderStructureRule to enforce sprint-based organization
- [x] Create DDLLocationRule to ensure DDL changes in structure directories
- [x] Create DMLLocationRule to ensure DML changes in data directories
- [x] Add configurable exclude patterns (default: **/init/**)
- [x] Support custom sprint, structure, and data folder naming patterns
- [x] Add comprehensive tests for all file structure rules
- [x] Update documentation with file structure configuration and rules

## Security Rules to Implement

### High Priority
1. **SQL Injection Detection**: Detect string concatenation in SQL
2. **Hardcoded Credentials**: Find passwords, API keys in changelogs
3. **Dangerous Operations**: DROP TABLE, TRUNCATE without preconditions
4. **Privilege Escalation**: GRANT ALL, CREATE USER with excessive privileges
5. **Data Exposure**: SELECT * into logs, unencrypted sensitive data

### Medium Priority
6. **Missing Rollback**: Changesets without rollback scripts
7. **Non-Idempotent Changes**: Changes that fail on re-run
8. **Context Misuse**: Production changes without proper context
9. **Missing Preconditions**: Risky operations without safety checks
10. **Broad Wildcards**: REVOKE/GRANT with %@%

## Best Practices for Human Readability

1. **Clear Function Names**: Names should describe what the function does
2. **Small Functions**: Keep functions focused on a single task
3. **Meaningful Variable Names**: Avoid single-letter names except in limited scopes
4. **Comments for Why**: Comment the reasoning, not the obvious
5. **Consistent Formatting**: Use gofmt and follow Go conventions
6. **Avoid Magic Numbers**: Use named constants
7. **Group Related Code**: Use blank lines to separate logical sections

## Dependencies Guidelines

1. **Minimize Dependencies**: Use standard library when possible
2. **Vet Dependencies**: Check popularity, maintenance, and license
3. **Pin Versions**: Use exact versions in go.mod
4. **Security Scanning**: Regularly scan for vulnerabilities

### Recommended Libraries

- **CLI Framework**: `github.com/spf13/cobra` (optional, flag package is sufficient)
- **YAML Parsing**: `gopkg.in/yaml.v3`
- **JSON Parsing**: Standard library `encoding/json`
- **XML Parsing**: Standard library `encoding/xml`
- **Terminal Colors**: `github.com/fatih/color`
- **Testing**: Standard library `testing`

## Build and Release

```bash
# Build for current platform
go build -o liquibase-linter ./cmd/liquibase-linter

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o liquibase-linter-linux-amd64 ./cmd/liquibase-linter
GOOS=darwin GOARCH=amd64 go build -o liquibase-linter-darwin-amd64 ./cmd/liquibase-linter
GOOS=windows GOARCH=amd64 go build -o liquibase-linter-windows-amd64.exe ./cmd/liquibase-linter

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linters
golangci-lint run
```

## CI/CD Integration

The linter should integrate seamlessly with:
- GitHub Actions
- GitLab CI
- Jenkins
- CircleCI
- Travis CI

**Example GitHub Action**:
```yaml
- name: Run Liquibase Linter
  run: |
    ./liquibase-linter check --format=json db/changelog/ > results.json
    exit_code=$?
    if [ $exit_code -ne 0 ]; then
      cat results.json
      exit $exit_code
    fi
```

## Success Criteria

The project is considered successful when:
1. ✅ Code follows standard Go project layout
2. ✅ All core security rules are implemented and tested
3. ✅ CLI is intuitive and fast (<100ms startup for small changelogs)
4. ✅ Test coverage exceeds 80%
5. ✅ Documentation is comprehensive and clear
6. ✅ Code passes golangci-lint checks
7. ✅ Successfully integrates with CI/CD pipelines
8. ✅ Handles edge cases gracefully with helpful error messages

## Resources

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Liquibase Documentation](https://docs.liquibase.com/)
- [SARIF Format](https://sarifweb.azurewebsites.net/)

---

**Last Updated**: February 4, 2026  
**Version**: 1.0.0  
**Maintainer**: n2jsoft