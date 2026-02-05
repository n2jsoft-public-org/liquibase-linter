# file-structure-sprint

**Severity**: Critical  
**Category**: Best Practices  
**Status**: ✅ Implemented

## Description

Enforces sprint-based directory organization for changelog files.

## What it detects

- Changelog files outside of sprint folders (e.g., `sprints/v116/`)
- Sprint folders that don't match the configured pattern
- Proper hierarchical organization of database changes

## Example violation

```
db/
  changelog/
    hotfix/           # ❌ Not in a sprint folder
      tables.xml
    random/           # ❌ Not in a sprint folder
      migration.xml
```

## Correct usage

```
db/
  changelog/
    sprints/
      v116/           # ✅ Matches sprint pattern
        0 - structure/
          tables.xml
        1 - data/
          seeds.xml
      v117/           # ✅ Matches sprint pattern
        0 - structure/
          indexes.xml
    init/             # ✅ Excluded from validation
      baseline.xml
```

## Configuration

```yaml
file_structure:
  sprint_pattern: "^v\\d+$"  # Matches v116, v117, etc.
  exclude_patterns:
    - "**/init/**"  # Exclude initialization scripts
```

### Custom Sprint Patterns

You can customize the pattern to match your team's naming convention:

```yaml
file_structure:
  # For "sprint-116" format
  sprint_pattern: "^sprint-\\d+$"
  
  # For "2024.Q1.1" format
  sprint_pattern: "^\\d{4}\\.Q[1-4]\\.\\d+$"
  
  # For "release-1.2.3" format
  sprint_pattern: "^release-\\d+\\.\\d+\\.\\d+$"
```

## Why this matters

Sprint-based organization provides:
- **Clear versioning**: Easy to identify which changes belong to which sprint/release
- **Simplified rollback**: Can easily identify and revert sprint-specific changes
- **Better planning**: Teams can see the scope of database changes per sprint
- **Audit trail**: Clear history of when changes were introduced
- **Deployment strategies**: Enables sprint-based or incremental deployments

## Benefits

- Easy identification of changes by sprint
- Consistent organization across projects
- Clear release tracking
- Simplified deployment pipeline configuration
- Better collaboration between database and application teams

## See Also

- [file-structure-ddl](file-structure-ddl.md) - Ensures DDL changes in structure directories
- [file-structure-dml](file-structure-dml.md) - Ensures DML changes in data directories
