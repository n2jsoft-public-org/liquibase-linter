# Rules Reference

This document provides an overview of all available linting rules in the Liquibase Linter.

## Rule Categories

- **Security**: Detect security vulnerabilities and risks
- **Performance**: Identify performance issues
- **Reliability**: Ensure reliable database migrations
- **Best Practices**: Enforce coding standards and conventions

## Security Rules

| Rule                                                    | Severity | Description                                     |
| ------------------------------------------------------- | -------- | ----------------------------------------------- |
| [sql-injection](rules/sql-injection.md)                 | Critical | Detects potential SQL injection vulnerabilities |
| [hardcoded-credentials](rules/hardcoded-credentials.md) | Critical | Finds hardcoded passwords and API keys          |
| [dangerous-operations](rules/dangerous-operations.md)   | Critical | Detects operations that could cause data loss   |
| [privilege-escalation](rules/privilege-escalation.md)   | Critical | Detects excessive privilege grants              |

## Reliability Rules

| Rule                                                    | Severity | Description                                     |
| ------------------------------------------------------- | -------- | ----------------------------------------------- |
| [missing-rollback](rules/missing-rollback.md)           | Warning  | Ensures changesets have proper rollback scripts |
| [non-idempotent](rules/non-idempotent.md)               | Warning  | Detects operations that may fail on re-run      |
| [missing-preconditions](rules/missing-preconditions.md) | Warning  | Ensures risky operations have preconditions     |

## Performance Rules

| Rule                                                    | Severity | Description                                                |
| ------------------------------------------------------- | -------- | ---------------------------------------------------------- |
| [missing-indexes](rules/missing-indexes.md)             | Info     | Detects tables without proper indexes                      |
| [table-locks](rules/table-locks.md)                     | Warning  | Identifies operations that may cause prolonged table locks |
| [large-data-operations](rules/large-data-operations.md) | Warning  | Detects operations that manipulate large amounts of data   |

## Best Practices Rules

| Rule                                              | Severity | Description                                |
| ------------------------------------------------- | -------- | ------------------------------------------ |
| [naming-conventions](rules/naming-conventions.md) | Info     | Enforces consistent naming conventions     |
| [unique-changeset](rules/unique-changeset.md)     | Critical | Detects duplicate changesets               |
| [documentation](rules/documentation.md)           | Info     | Ensures changesets are properly documented |

## File Structure Rules

| Rule                                                    | Severity | Description                                  |
| ------------------------------------------------------- | -------- | -------------------------------------------- |
| [file-structure-sprint](rules/file-structure-sprint.md) | Critical | Enforces sprint-based directory organization |
| [file-structure-ddl](rules/file-structure-ddl.md)       | Critical | Ensures DDL changes in structure directories |
| [file-structure-dml](rules/file-structure-dml.md)       | Critical | Ensures DML changes in data directories      |

## Configuring Rules

All rules can be individually configured in your `.liquibase-linter.yaml` configuration file:

```yaml
rules:
  # Security rules
  sql-injection:
    enabled: true
    severity: critical
  
  hardcoded-credentials:
    enabled: true
    severity: critical
  
  dangerous-operations:
    enabled: true
    severity: critical
  
  privilege-escalation:
    enabled: true
    severity: critical
  
  # Reliability rules
  missing-rollback:
    enabled: true
    severity: warning
  
  non-idempotent:
    enabled: true
    severity: warning
  
  missing-preconditions:
    enabled: true
    severity: warning
  
  # Performance rules
  missing-indexes:
    enabled: true
    severity: info
  
  table-locks:
    enabled: true
    severity: warning
  
  large-data-operations:
    enabled: true
    severity: warning
  
  # Best practices rules
  naming-conventions:
    enabled: false  # Disable if you have custom conventions
  
  changelog-organization:
    enabled: true
    severity: info
  
  documentation:
    enabled: true
    severity: info
  
  # File structure rules
  file-structure-sprint:
    enabled: true
    severity: critical
  
  file-structure-ddl:
    enabled: true
    severity: critical
  
  file-structure-dml:
    enabled: true
    severity: critical

# File structure configuration
file_structure:
  sprint_pattern: "^v\\d+$"
  structure_pattern: "^\\d+ - structure$"
  data_pattern: "^\\d+ - data$"
  exclude_patterns:
    - "**/init/**"
```

### Severity Levels

- **Critical**: Security vulnerabilities and data loss risks - should fail builds
- **Warning**: Reliability issues and potential problems - may fail builds
- **Info**: Best practice violations and code quality issues - typically advisory

### Disabling Rules

To disable a rule, set `enabled: false`:

```yaml
rules:
  naming-conventions:
    enabled: false
```

### Adjusting Severity

You can adjust the severity level of any rule:

```yaml
rules:
  missing-rollback:
    enabled: true
    severity: critical  # Upgrade from warning to critical
```

## File Structure Configuration

The file structure rules support flexible pattern matching:

```yaml
file_structure:
  # Sprint folder pattern (regex)
  sprint_pattern: "^v\\d+$"          # Matches: v116, v117, etc.
  
  # Structure directory pattern (regex)
  structure_pattern: "^\\d+ - structure$"  # Matches: 0 - structure, 1 - structure
  
  # Data directory pattern (regex)
  data_pattern: "^\\d+ - data$"      # Matches: 0 - data, 1 - data
  
  # Patterns to exclude from validation (glob patterns)
  exclude_patterns:
    - "**/init/**"      # Exclude initialization folders
    - "**/legacy/**"    # Exclude legacy migrations
    - "**/hotfix/**"    # Exclude hotfix folders
```

### Example Custom Patterns

```yaml
# For "sprint-116" naming
file_structure:
  sprint_pattern: "^sprint-\\d+$"
  structure_pattern: "^ddl$"
  data_pattern: "^dml$"

# For release-based versioning
file_structure:
  sprint_pattern: "^release-\\d+\\.\\d+\\.\\d+$"
  structure_pattern: "^schema$"
  data_pattern: "^data$"
```

## Next Steps

- Review individual rule documentation in the [rules/](rules/) directory
- Set up your [configuration file](configuration.md)
- Integrate with your [CI/CD pipeline](cicd.md)
- See [usage examples](usage.md) for common scenarios

## Rule Status

All rules are ✅ **Implemented** and ready to use. Each rule has comprehensive tests and detailed documentation.
