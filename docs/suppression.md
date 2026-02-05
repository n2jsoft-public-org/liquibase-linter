# Rule Suppression

Sometimes you need to intentionally violate a linting rule for legitimate reasons. The Liquibase Linter provides inline comment-based suppression to disable specific rules for individual changesets.

## Syntax

Add a suppression directive to your changeset's comment using the following syntax:

```
liquibase-linter:disable <rule-id>[,<rule-id>,...]
```

**Key Points:**
- Directive is case-insensitive: `liquibase-linter:disable` or `LIQUIBASE-LINTER:DISABLE`
- Rule IDs are case-sensitive: use exact rule ID (e.g., `sql-injection` not `SQL-INJECTION`)
- Multiple rules: separate with commas (spaces optional)
- Additional text: can appear before or after the directive in the comment

## Format-Specific Examples

### XML Format

```xml
<changeSet id="1" author="john">
    <comment>liquibase-linter:disable sql-injection</comment>
    <sql>
        INSERT INTO users VALUES (1, ${username});
    </sql>
</changeSet>

<!-- Multiple rules -->
<changeSet id="2" author="jane">
    <comment>liquibase-linter:disable sql-injection,missing-rollback</comment>
    <sql>
        UPDATE users SET name = ${newname};
    </sql>
</changeSet>

<!-- With explanatory text -->
<changeSet id="3" author="bob">
    <comment>Admin creation for test. liquibase-linter:disable hardcoded-credentials This is for CI/CD only</comment>
    <sql>
        CREATE USER 'admin'@'localhost' IDENTIFIED BY 'test123';
    </sql>
</changeSet>
```

### SQL Format

```sql
--liquibase formatted sql

--changeset john:1
--comment: liquibase-linter:disable sql-injection
INSERT INTO users VALUES (1, ${username});

--changeset jane:2
--comment: liquibase-linter:disable sql-injection,missing-rollback
UPDATE users SET name = ${newname};

--changeset bob:3
--comment: Test data. liquibase-linter:disable hardcoded-credentials for development
CREATE USER 'testuser'@'localhost' IDENTIFIED BY 'testpass';
```

### YAML Format

```yaml
databaseChangeLog:
  - changeSet:
      id: 1
      author: john
      comment: "liquibase-linter:disable sql-injection"
      changes:
        - sql:
            sql: "INSERT INTO users VALUES (1, ${username});"

  - changeSet:
      id: 2
      author: jane
      comment: "liquibase-linter:disable sql-injection,missing-rollback"
      changes:
        - sql:
            sql: "UPDATE users SET name = ${newname};"
```

### JSON Format

```json
{
  "databaseChangeLog": [
    {
      "changeSet": {
        "id": "1",
        "author": "john",
        "comment": "liquibase-linter:disable sql-injection",
        "changes": [
          {
            "sql": {
              "sql": "INSERT INTO users VALUES (1, ${username});"
            }
          }
        ]
      }
    }
  ]
}
```

## Available Rule IDs

To get a list of all available rule IDs, run:

```bash
liquibase-linter rules
```

Common rule IDs include:
- `sql-injection` - SQL injection vulnerabilities
- `hardcoded-credentials` - Hardcoded passwords or API keys
- `dangerous-operations` - Unsafe operations (DROP, TRUNCATE without preconditions)
- `missing-rollback` - Changesets without rollback instructions
- `missing-preconditions` - Missing safety preconditions
- `non-idempotent` - Changes that fail on re-run
- `privilege-escalation` - Excessive privilege grants
- `table-locks` - Operations causing table locks
- `naming-conventions` - Naming standard violations
- `changeset-documentation` - Missing comments

See [Rules Documentation](rules.md) for the complete list.

## Behavior

### Filtered Violations
Suppressed violations are completely **hidden from output**. They do not appear in:
- Text output
- JSON output
- SARIF output
- Violation counts

### Validation Warnings
Invalid rule IDs in suppression directives generate **warnings on stderr** but do not fail the linting process:

```
Warning: Unknown rule 'invalid-rule' in suppression directive (changeset john:1 in db/changelog.xml)
```

**Common issues:**
- Typos in rule IDs
- Incorrect case (rule IDs are case-sensitive)
- Non-existent rules

### Exit Codes
Suppressions do not affect exit codes. The linter exits with:
- `0`: No violations found (after suppression filtering)
- `1`: Violations found (after suppression filtering)
- `2`: Error during execution

## Best Practices

### ✅ DO

**1. Be Specific**
```xml
<!-- Good: Only suppress the necessary rule -->
<comment>liquibase-linter:disable sql-injection</comment>
```

**2. Add Explanations**
```xml
<!-- Good: Explain WHY you're suppressing -->
<comment>liquibase-linter:disable hardcoded-credentials Test account for CI, removed in production</comment>
```

**3. Limit Scope**
```sql
-- Good: Suppress only for the changeset that needs it
--changeset john:1
--comment: liquibase-linter:disable sql-injection
-- Known safe parameter from config
INSERT INTO settings VALUES (${config_value});

--changeset john:2
-- No suppression - normal validation
INSERT INTO users VALUES ('john', 'john@example.com');
```

**4. Review Suppressions in Code Reviews**
Treat suppression directives like `@SuppressWarnings` in code - they should be justified and reviewed.

### ❌ DON'T

**1. Suppress Multiple Unrelated Rules**
```xml
<!-- Bad: Too broad -->
<comment>liquibase-linter:disable sql-injection,hardcoded-credentials,missing-rollback,dangerous-operations</comment>
```

**2. Use Suppressions to Hide Problems**
```xml
<!-- Bad: Fix the underlying issue instead -->
<comment>liquibase-linter:disable sql-injection</comment>
<sql>
    DELETE FROM users WHERE name = ${user_input}; -- This is actually dangerous!
</sql>
```

**3. Use Incorrect Case**
```xml
<!-- Bad: Rule IDs are case-sensitive -->
<comment>liquibase-linter:disable SQL-INJECTION</comment>
<!-- This will generate a warning and NOT work -->
```

**4. Forget to Remove Temporary Suppressions**
```sql
-- Bad: Left behind after fixing the issue
--comment: liquibase-linter:disable missing-rollback TODO: add rollback later
-- Rollback was added but suppression remains
```

## Use Cases

### Legitimate Suppression Scenarios

**1. Test/Development Data**
```yaml
comment: "liquibase-linter:disable hardcoded-credentials Development credentials only"
```

**2. Legacy Migration**
```sql
--comment: liquibase-linter:disable non-idempotent Legacy data migration, runs once
```

**3. Known Safe Variables**
```xml
<comment>liquibase-linter:disable sql-injection Config value validated at application layer</comment>
```

**4. Framework Limitations**
```yaml
comment: "liquibase-linter:disable missing-rollback Rollback not possible for this data migration"
```

## Integration with CI/CD

Suppression warnings appear on stderr and are visible in CI/CD logs:

```bash
./liquibase-linter check db/changelog.xml 2>&1 | tee linter.log
```

**Example in GitHub Actions:**

```yaml
- name: Lint Liquibase Changelogs
  run: |
    ./liquibase-linter check db/changelog/ 2>&1 | tee linter-output.txt
    if [ $? -ne 0 ]; then
      echo "::error::Liquibase linting failed"
      exit 1
    fi
    
    # Check for suppression warnings
    if grep -q "Warning: Unknown rule" linter-output.txt; then
      echo "::warning::Invalid suppression directives found"
    fi
```

## Troubleshooting

### Suppression Not Working

**Problem:** Violations still appear even with suppression directive

**Checklist:**
1. ✅ Check rule ID case (must match exactly: `sql-injection` not `SQL-INJECTION`)
2. ✅ Verify rule ID spelling
3. ✅ Ensure directive is in the changeset's `comment` field (not in SQL comments)
4. ✅ Check that the changeset ID and author match the suppressed changeset

**Get valid rule IDs:**
```bash
liquibase-linter rules
```

### Warning: Unknown Rule

**Problem:** `Warning: Unknown rule 'xyz' in suppression directive`

**Solutions:**
1. Fix typo in rule ID
2. Use correct case (e.g., `sql-injection` not `Sql-Injection`)
3. Remove suppression if rule no longer exists
4. Run `liquibase-linter rules` to see available rules

### Suppression Ignored

**Problem:** No warning but suppression seems ignored

**Check:**
- Comment field populated correctly in all formats
- Directive syntax: `liquibase-linter:disable rule-id`
- No extra characters interrupting the directive

## Related Documentation

- [Rules Overview](rules.md) - Complete list of rules
- [Configuration](configuration.md) - Global rule configuration
- [Usage Guide](usage.md) - Command-line usage
- [CI/CD Integration](cicd.md) - Continuous integration setup

---

For questions or issues with suppression, please file an issue on GitHub.
