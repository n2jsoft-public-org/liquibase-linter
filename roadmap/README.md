# Roadmap

This directory contains detailed implementation plans for planned features and rules.

## Planned Rules

### 1. Enhance Non-Idempotent Rule (MERGED: mandatory-preconditions + missing-preconditions)
**File**: [enhance-non-idempotent-rule.md](enhance-non-idempotent-rule.md)  
**Status**: 📋 Planned  
**Priority**: High  
**Effort**: 7-8 hours

Enhance the existing `non-idempotent` rule with configurable modes (`risky-only`, `all`) and exclude patterns, consolidating three overlapping precondition rules into one flexible rule.

> **Note**: This replaces the planned `mandatory-preconditions` rule and addresses the never-implemented `missing-preconditions` rule by adding configuration options to the existing `non-idempotent` rule.

### 2. Label Pattern
**File**: [label-pattern.md](label-pattern.md)  
**Status**: 📋 Planned  
**Priority**: High  
**Effort**: 3-4 hours

Enforces that changeset labels match a configured pattern (e.g., sprint versions like "v123").

### 3. No Manual Transactions
**File**: [no-manual-transactions.md](no-manual-transactions.md)  
**Status**: 📋 Planned  
**Priority**: Medium  
**Effort**: 4-5 hours

Detects manual transaction control statements (BEGIN, COMMIT, ROLLBACK) that interfere with Liquibase's transaction management.

### 4. Redundant onError:HALT
**File**: [redundant-onerror-halt.md](redundant-onerror-halt.md)  
**Status**: 📋 Planned  
**Priority**: Low  
**Effort**: 1-2 hours

Detects redundant `onError:HALT` configuration in preconditions (HALT is the default value).

### 5. No IF EXISTS
**File**: [no-if-exists.md](no-if-exists.md)  
**Status**: 📋 Planned  
**Priority**: Medium  
**Effort**: 3-4 hours

Detects database-specific IF EXISTS patterns in SQL scripts and recommends using Liquibase preconditions for better cross-database compatibility.

### 6. Atomic Changeset
**File**: [atomic-changeset.md](atomic-changeset.md)  
**Status**: 📋 Planned  
**Priority**: Medium  
**Effort**: 7-8 hours

Enforces that each changeset contains only a single change operation for better atomicity, cleaner rollbacks, and easier debugging.

## Implementation Summary

| Rule                   | File                             | Target        | Effort | Priority |
| ---------------------- | -------------------------------- | ------------- | ------ | -------- |
| redundant-onerror-halt | `internal/rules/bestpractice.go` | Existing file | 1-2h   | Low      |
| enhance-non-idempotent | `internal/rules/bestpractice.go` | Existing file | 7-8h   | High     |
| label-pattern          | `internal/rules/bestpractice.go` | Existing file | 3-4h   | High     |
| no-if-exists           | `internal/rules/bestpractice.go` | Existing file | 3-4h   | Medium   |
| atomic-changeset       | `internal/rules/bestpractice.go` | Existing file | 7-8h   | Medium   |
| no-manual-transactions | `internal/rules/reliability.go`  | **NEW FILE**  | 4-5h   | Medium   |

**Total Estimated Effort**: 26-31 hours (including testing, documentation, and integration)

**Effort Savings**: By consolidating three precondition rules into one enhanced rule, we save ~3-5 hours of duplicate implementation work.

## Implementation Order

Recommended order (easiest to most complex):

1. **redundant-onerror-halt** - Simple string comparison on `Precondition.OnError`
2. **label-pattern** - Regex pattern matching on `ChangeSet.Labels`
3. **atomic-changeset** - Count changes per changeset with SQL statement parsing
5. **no-manual-transactions** - Complex SQL analysis with preprocessing
6. **no-manual-transactions** - Complex SQL analysis with preprocessing
5. **enhance-non-idempotent** - Enhance existing rule with modes and exclude patterns (do last to allow other rules to benefit from the enhanced precondition checking)
4. **no-manual-transactions** - Complex SQL analysis with preprocessing

## Quick Start

To implement a rule from this roadmap:

1. Read the detailed task file (e.g., `mandatory-preconditions.md`)
2. Follow the implementation steps in order
3. Use the provided code templates and test cases
4. Complete the validation checklist before marking as done
5. Update this README to mark the rule as ✅ Implemented

## Status Legend

- 📋 **Planned**: Design complete, ready for implementation
- 🚧 **In Progress**: Currently being implemented
- ✅ **Implemented**: Completed and merged
- ⏸️ **On Hold**: Paused pending decisions or dependencies

## Contributing

When adding new planned features to the roadmap:

1. Create a new markdown file following the template structure
2. Include all required sections (see AGENTS.md for details)
3. Add an entry to this README
4. Link to relevant documentation files

## Related Documentation

- [AGENTS.md](../AGENTS.md) - Project overview and development guidelines
- [docs/rules/](../docs/rules/) - Rule documentation
- [docs/development.md](../docs/development.md) - Development guide

---

**Last Updated**: February 5, 2026
