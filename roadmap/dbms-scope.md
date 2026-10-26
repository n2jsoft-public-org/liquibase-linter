# DBMS Scope Rules Implementation

## Objective

Verify changelogs that run on several database engines, for example during a SQL Server →
PostgreSQL migration in which one changelog serves both engines: every changeset is scoped with
`dbms:`, and SQL Server changes get a PostgreSQL "twin". Five checks, split into two rules:

| # | Rule | Check |
|---|---|---|
| 1 | `dbms-scope` | Every changeset declares `dbms:` |
| 2 | `dbms-scope` | `dbms:` only names known engines (a typo silently skips the changeset everywhere) |
| 3 | `dbms-scope` | A changeset reaching engine X has no precondition SQL matching X's forbidden patterns (e.g. `sys.*`, `OBJECT_ID()` on PostgreSQL) |
| 4 | `dbms-twin` | A changeset that runs on the source engine only has a twin `<id><suffix>` for the target engine in the same file, or suppresses the rule |
| 5 | `dbms-twin` | A changeset whose ID ends with the twin suffix targets exactly the target engine |

`dbms-scope` is generic (any multi-engine changelog). `dbms-twin` serves migrations between two
engines; it is Phase 2.

## Requirements

### Functional Requirements

1. **Required dbms** (`dbms-scope`): report a changeset with no `dbms` attribute, when `require_dbms` is on.
2. **Known engines** (`dbms-scope`): report each `dbms` value, with its `!` prefix stripped, that is
   not `all`, `none` or one of `allowed_values`. Comparison is case-insensitive.
3. **Forbidden precondition SQL** (`dbms-scope`): for each engine in `forbidden_precondition_patterns`,
   report a changeset that reaches that engine and whose precondition SQL matches one of the engine's
   regexes.
4. **Twin required** (`dbms-twin`): in files in scope, a changeset that names `source_dbms` and does
   not reach `target_dbms` must have, in the same file, a changeset with the same author, the ID
   `<id><id_suffix>`, and a `dbms` that reaches `target_dbms`. The exemption is the existing inline
   suppression (`liquibase-linter:disable dbms-twin`), so no new marker is introduced.
5. **Twin engine** (`dbms-twin`): in files in scope, a changeset whose ID ends with `id_suffix` must
   declare `dbms` equal to `target_dbms` alone.
6. **Scope**: both rules honour `exclude_patterns`. `dbms-twin` can also be limited to sprint
   folders from a given number on (`since_path_pattern` + `since`).
7. **Opt-in**: both rules are disabled by default; single-engine projects never see them.
8. **All formats**: XML, YAML, JSON and formatted SQL.

### Non-Functional Requirements

1. One pass over the changesets; the twin lookup is a map per file.
2. Violations carry `FilePath`, `ChangeSetID`, `Author` (needed by `FilterSuppressedViolations`)
   and, for formatted SQL, the line number and text of the changeset header (`LineNumber`, `Line`).
3. Table-driven tests, >80% coverage of new code.

### Definition: "reaches an engine"

Liquibase runs a changeset on engine E when its `dbms` list allows E:

- no `dbms` → reaches every engine;
- a positive `none` → reaches nothing;
- `!E` or `!all` → does not reach E;
- otherwise it reaches E if the list names `E` or `all`, or if it contains only negations.

## Implementation Details

### 1. Parser changes (prerequisite)

Two gaps in `internal/parser` block check 3 and the line numbers:

**a. Precondition SQL is dropped.**
- Formatted SQL: `--precondition-sql-check expectedResult:0 SELECT ...` does not match
  `preconditionsRegex` (`^--\s*preconditions?\s+`) and is skipped as a comment
  (`internal/parser/sql.go:20`).
- XML: `<sqlCheck>` is unmarshalled into `xmlSQLCheck` but `convertPreconditions` ignores it
  (`internal/parser/xml.go:344`). It also keeps only the first precondition (`&precond[0]`).
- YAML/JSON: nested preconditions are flattened with `fmt.Sprintf("%v")` (`internal/parser/yaml.go:316`).

Add a dedicated field instead of reworking `Preconditions` (which other rules read as a single value):

```go
// ChangeSet represents a single changeset in a changelog.
type ChangeSet struct {
	// ...existing fields...
	PreconditionSQL []string // SQL of every sqlCheck precondition, in declaration order
	LineNumber      int      // Line of the changeset header (formatted SQL only; 0 otherwise)
}
```

- SQL parser: add `preconditionSQLCheckRegex = regexp.MustCompile(`^--\s*precondition-sql-check\s+(.*)$`)`,
  checked before `preconditionsRegex`. Strip leading `key:value` tokens (`expectedResult:0`) and
  append the rest. Set `LineNumber` from the scanner's line counter when a `--changeset` header starts.
- XML parser: append each `xmlPreconditions.SQLCheck[i].SQL` (trimmed) to `PreconditionSQL`.
  `sqlCheck` nested in `<and>`/`<or>`/`<not>` is out of scope for now (see Notes).
- YAML/JSON parser: walk the `preConditions` list and append `sqlCheck.sql`.

**b. DBMS helper**, next to `IsDDLChange()`:

```go
// ReachesDBMS reports whether Liquibase runs the changeset on the given engine,
// according to its dbms attribute.
func (cs *ChangeSet) ReachesDBMS(engine string) bool
```

### 2. Rule `dbms-scope`

**File**: `internal/rules/reliability.go`

```go
// DBMSScopeRule checks that changesets declare which database engines they run on,
// name only known engines, and keep engine-specific SQL out of the preconditions of
// changesets that reach other engines.
type DBMSScopeRule struct {
	config            *config.DBMSScopeConfig
	allowed           map[string]bool
	forbiddenPatterns map[string][]*regexp.Regexp // engine -> compiled patterns
}

func NewDBMSScopeRule(cfg *config.DBMSScopeConfig) *DBMSScopeRule

func (r *DBMSScopeRule) ID() string          { return "dbms-scope" }
func (r *DBMSScopeRule) Name() string        { return "DBMS Scope" }
func (r *DBMSScopeRule) Description() string {
	return "Ensures changesets declare a valid dbms and keep engine-specific SQL out of preconditions of other engines"
}
func (r *DBMSScopeRule) Severity() Severity { return SeverityCritical }

func (r *DBMSScopeRule) Check(changelog *parser.Changelog) []Violation {
	// for each changeset not in exclude_patterns:
	//   1. require_dbms && len(DBMSList) == 0            -> violation
	//   2. each value (minus "!") not all/none/allowed     -> violation per value
	//   3. for engine, patterns in forbiddenPatterns:
	//        cs.ReachesDBMS(engine) && any PreconditionSQL matches -> violation (one per changeset)
}
```

Patterns are compiled once in the constructor; an invalid regex is rejected by `config.Validate`.

Messages:

- `changeset <author>:<id> has no dbms. Add dbms:<engine> to its header, or dbms:all for dialect-neutral SQL.`
- `changeset <author>:<id>: unknown engine '<name>' in dbms:<list>. Allowed: <list>. An unknown name matches no database and skips the changeset everywhere.`
- `changeset <author>:<id> runs on <engine> (dbms:<list>) but a precondition matches '<pattern>', which <engine> does not support.`

### 3. Rule `dbms-twin` (Phase 2)

**File**: `internal/rules/reliability.go`

```go
// DBMSTwinRule checks that every changeset restricted to the source engine has a
// counterpart for the target engine in the same file.
type DBMSTwinRule struct {
	config     *config.DBMSTwinConfig
	sincePath  *regexp.Regexp // nil when no sprint threshold is configured
}

func (r *DBMSTwinRule) ID() string { return "dbms-twin" }

func (r *DBMSTwinRule) Check(changelog *parser.Changelog) []Violation {
	// group changesets by FilePath; skip files excluded or before `since`
	// index per file: author + ":" + id -> changeset
	// 4. names source, !ReachesDBMS(target), !HasSuffix(id, suffix):
	//      twin := index[author:id+suffix]; twin missing or !twin.ReachesDBMS(target) -> violation
	// 5. HasSuffix(id, suffix) && DBMSList != [target] -> violation
}
```

`since` scope: when `since_path_pattern` is set, its first capture group is parsed as an integer
and compared to `since` (`sprints/v(\d+)/` with `since: 10` → `v10` and later). A file that does
not match the pattern is out of scope.

Messages:

- `changeset <author>:<id> runs on <source> only and has no <target> twin. Add changeset <id><suffix> with dbms:<target> to this file, or suppress dbms-twin with a reason.`
- `changeset <author>:<id> is a <target> twin (<suffix>) but declares dbms:<list>. Use dbms:<target>.`

### Registration

**File**: `cmd/liquibase-linter/main.go`, next to the other opt-in rules:

```go
if cfg.DBMSScope.Enabled {
	registry.Register(rules.NewDBMSScopeRule(&cfg.DBMSScope))
}
if cfg.DBMSTwin.Enabled {
	registry.Register(rules.NewDBMSTwinRule(&cfg.DBMSTwin))
}
```

Also add both to the `rules` listing command (`main.go` around line 610) so `liquibase-linter rules`
shows them.

## Configuration

**File**: `internal/config/config.go`

```go
// DBMSScopeConfig configures the dbms-scope rule.
type DBMSScopeConfig struct {
	ForbiddenPreconditionPatterns map[string][]string `yaml:"forbidden_precondition_patterns"`
	AllowedValues                 []string            `yaml:"allowed_values"`
	ExcludePatterns               []string            `yaml:"exclude_patterns"`
	Enabled                       bool                `yaml:"enabled"`
	RequireDBMS                   bool                `yaml:"require_dbms"`
}

// DBMSTwinConfig configures the dbms-twin rule.
type DBMSTwinConfig struct {
	SourceDBMS       string   `yaml:"source_dbms"`
	TargetDBMS       string   `yaml:"target_dbms"`
	IDSuffix         string   `yaml:"id_suffix"`
	SincePathPattern string   `yaml:"since_path_pattern"`
	ExcludePatterns  []string `yaml:"exclude_patterns"`
	Since            int      `yaml:"since"`
	Enabled          bool     `yaml:"enabled"`
}
```

### Defaults

- `dbms_scope`: `enabled: false`, `require_dbms: true`, `allowed_values`: Liquibase's short names
  (`db2`, `derby`, `h2`, `hsqldb`, `informix`, `mariadb`, `mssql`, `mysql`, `oracle`, `postgresql`,
  `snowflake`, `sqlite`, `sybase`, `cockroachdb`), `forbidden_precondition_patterns: {}`.
- `dbms_twin`: `enabled: false`, `id_suffix: "-pg"`, no default engines.

### Validation (`config.Validate`)

- `forbidden_precondition_patterns`: every regex compiles.
- `dbms_twin` enabled: `source_dbms`, `target_dbms` and `id_suffix` are set and differ;
  `since_path_pattern`, if set, compiles and has at least one capture group.

### Example configuration (SQL Server → PostgreSQL migration)

```yaml
rules:
  dbms-scope:
    enabled: true
    severity: critical
  dbms-twin:
    enabled: true
    severity: critical

dbms_scope:
  enabled: true
  require_dbms: true
  allowed_values: [mssql, postgresql]
  exclude_patterns:
    - "changelog/generated/**"
  forbidden_precondition_patterns:
    postgresql:
      - '(?i)(^|[^a-z0-9_])sys\.'
      - '(?i)\bobject_id\s*\('
      - '(?i)\bcol_length\s*\('

dbms_twin:
  enabled: true
  source_dbms: mssql
  target_dbms: postgresql
  id_suffix: "-pg"
  since_path_pattern: 'changelog/sprints/v(\d+)/'
  since: 10
```

The global `ignore` hides files from every rule. A folder that other rules skip but that must still
declare `dbms` (a baseline, for instance) belongs in those rules' `exclude_patterns`, not in `ignore`.

## Testing

### Unit tests

**File**: `internal/parser/types_test.go`: `TestChangeSet_ReachesDBMS` (table-driven):

| dbms | engine | reaches |
|---|---|---|
| (none) | postgresql | true |
| `mssql` | postgresql | false |
| `postgresql` | postgresql | true |
| `mssql,postgresql` | postgresql | true |
| `all` | postgresql | true |
| `none` | postgresql | false |
| `!mssql` | postgresql | true |
| `!postgresql` | postgresql | false |
| `!all` | postgresql | false |
| `MSSQL` | mssql | true |

**File**: `internal/parser/sql_test.go`: `--precondition-sql-check expectedResult:0 SELECT ...` fills
`PreconditionSQL` with `SELECT ...`; `LineNumber` is the header line. Same for `<sqlCheck>` in
`xml_test.go` and `sqlCheck.sql` in `yaml_test.go`.

**File**: `internal/rules/reliability_test.go`:

`TestDBMSScopeRule_Check`:
- missing dbms → 1 violation; same with `require_dbms: false` → 0
- `dbms:postgres` (typo) → 1 violation naming `postgres`
- `dbms:!postgresql` → 0 (negation of a known name)
- `dbms:mssql` with `OBJECT_ID(` precondition → 0 (does not reach PostgreSQL)
- `dbms:postgresql` with `sys.tables` precondition → 1
- `dbms:all` with `COL_LENGTH(` precondition → 1
- `myobject_id(` in precondition → 0 (word boundary)
- file matching `exclude_patterns` → 0

`TestDBMSTwinRule_Check`:
- `a:1 dbms:mssql` + `a:1-pg dbms:postgresql` → 0
- `a:1 dbms:mssql` alone → 1
- twin with another author (`b:1-pg`) → 1
- twin in another file → 1
- `a:1-pg dbms:mssql,postgresql` → 1 (check 5)
- `a:1 dbms:mssql,postgresql` → 0 (already reaches the target)
- file in `sprints/v9/` with `since: 10` → 0; `sprints/v10/` → 1
- `--comment: liquibase-linter:disable dbms-twin data fix for production rows` → filtered by
  `FilterSuppressedViolations` (test in `types_test.go`)

### Fixtures

**File**: `testdata/dbms-scope/changelog.sql`

```sql
--liquibase formatted sql

--changeset alice:no-dbms
CREATE TABLE a (id INT);

--changeset alice:typo dbms:postgres
CREATE TABLE b (id INT);

--changeset alice:catalog-on-pg dbms:postgresql
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM sys.tables WHERE name = 'c'
CREATE TABLE c (id INT);

--changeset alice:ok dbms:mssql
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM sys.tables WHERE name = 'd'
CREATE TABLE d (id INT);
```

**File**: `testdata/dbms-twin/sprints/v10/changelog.sql`

```sql
--liquibase formatted sql

--changeset alice:1 dbms:mssql
ALTER TABLE Users ADD Email NVARCHAR(255);

--changeset alice:1-pg dbms:postgresql
ALTER TABLE users ADD COLUMN email VARCHAR(255);

--changeset alice:2 dbms:mssql
ALTER TABLE Users ADD Phone NVARCHAR(50);

--changeset alice:3 dbms:mssql
--comment: liquibase-linter:disable dbms-twin SQL Server full-text index, no PostgreSQL equivalent
CREATE FULLTEXT INDEX ON Users(Name) KEY INDEX PK_Users;
```

Expected: one `dbms-twin` violation, on `alice:2`.

## Integration Steps

1. Parser: `PreconditionSQL`, `LineNumber`, `ReachesDBMS` + tests
2. Config: `DBMSScopeConfig`, `DBMSTwinConfig`, defaults, validation + tests
3. Rule `dbms-scope` + tests + fixture
4. Register `dbms-scope` in `main.go` (check and `rules` listing)
5. Rule `dbms-twin` + tests + fixture (Phase 2)
6. Register `dbms-twin`
7. Docs: `docs/rules/dbms-scope.md`, `docs/rules/dbms-twin.md`, add both to `docs/rules.md` and
   `docs/configuration.md`
8. `make lint` and `make test`

## Validation Checklist

- [ ] `PreconditionSQL` filled for formatted SQL, XML and YAML/JSON
- [ ] `LineNumber` set for formatted SQL changesets
- [ ] `ReachesDBMS` matches the table above
- [ ] Missing `dbms` reported only when `require_dbms` is on
- [ ] Unknown engine reported, negated known engine accepted
- [ ] Forbidden precondition reported only on changesets that reach the engine
- [ ] Twin found only in the same file, with the same author
- [ ] `-pg` changeset with another engine reported
- [ ] `since` threshold honoured
- [ ] Suppression via `liquibase-linter:disable dbms-twin` works
- [ ] Both rules disabled by default; invalid config rejected by `Validate`
- [ ] `liquibase-linter rules` lists both
- [ ] `make lint` and `make test` pass

## Expected Output

With the example configuration above:

```
$ liquibase-linter check --config=.liquibase-linter.yaml testdata/dbms-scope/changelog.sql

testdata/dbms-scope/changelog.sql
  ✖ [CRITICAL] changeset alice:no-dbms has no dbms. Add dbms:<engine> to its header, or dbms:all for dialect-neutral SQL.
    Changeset: no-dbms (author: alice)
    Line 3: --changeset alice:no-dbms
    Rule: dbms-scope
  ✖ [CRITICAL] changeset alice:typo: unknown engine 'postgres' in dbms:postgres. Allowed: mssql, postgresql. An unknown name matches no database and skips the changeset everywhere.
    Changeset: typo (author: alice)
    Line 6: --changeset alice:typo dbms:postgres
    Rule: dbms-scope
  ✖ [CRITICAL] changeset alice:catalog-on-pg runs on postgresql (dbms:postgresql) but a precondition matches '(?i)(^|[^a-z0-9_])sys\.', which postgresql does not support.
    Changeset: catalog-on-pg (author: alice)
    Line 9: --changeset alice:catalog-on-pg dbms:postgresql
    Rule: dbms-scope
```

The text reporter prints `LineNumber` only when `Line` is set, so violations set `Line` to the
changeset header when `LineNumber > 0`.

## Dependencies

- `internal/parser`: `PreconditionSQL`, `LineNumber`, `ReachesDBMS` (step 1, blocking)
- `internal/config`: two new sections
- Reporters: no change

## Estimated Effort

- Parser changes: 2-3 hours
- `dbms-scope`: 3-4 hours
- `dbms-twin`: 3-4 hours
- Docs: 1-2 hours
- **Total**: 9-13 hours

## Notes

- The linter follows `include`/`includeAll` from the root changelog, so a file not yet included is
  not checked, which is also when Liquibase would not run it.
- `dbms` on a changeset does not change its checksum, so adding it to deployed changesets is safe
  and `require_dbms` can be enabled on an existing changelog.
- XML `sqlCheck` nested inside `<and>`/`<or>`/`<not>` is not collected in the first version. Add it
  if an XML user needs it.
- Liquibase also accepts a `dbms` precondition, but not in formatted SQL; this rule only reads the
  changeset attribute.
- The twin naming (`<id><suffix>`, same author, same file) is fixed in the first version. Make the
  pairing configurable when a user needs a different one, not before.
