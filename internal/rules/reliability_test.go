package rules

import (
	"strings"
	"testing"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/config"
	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

func TestNoManualTransactionsRule_Check(t *testing.T) {
	rule := NewNoManualTransactionsRule(nil)

	tests := []struct {
		changelog     *parser.Changelog
		name          string
		expectedMsg   string
		wantViolation bool
	}{
		{
			name: "BEGIN TRANSACTION detected",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "BEGIN TRANSACTION;\nINSERT INTO users VALUES (1);\nCOMMIT;",
							},
						},
					},
				},
			},
			wantViolation: true,
			expectedMsg:   "BEGIN TRANSACTION",
		},
		{
			name: "START TRANSACTION detected",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "2",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "START TRANSACTION;\nUPDATE users SET active = 1;",
							},
						},
					},
				},
			},
			wantViolation: true,
			expectedMsg:   "START TRANSACTION",
		},
		{
			name: "transaction keyword in comment - should not detect",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "3",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "-- BEGIN TRANSACTION for testing\nINSERT INTO users VALUES (1);",
							},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "transaction keyword in string - should not detect",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "4",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "INSERT INTO logs (message) VALUES ('BEGIN TRANSACTION complete');",
							},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "stored procedure with transaction - excluded",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "5",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "createProcedure",
								SQL:  "CREATE PROCEDURE test AS BEGIN BEGIN TRANSACTION; UPDATE users; COMMIT; END;",
							},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "COMMIT detected",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "6",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "UPDATE users SET active = 1;\nCOMMIT TRANSACTION;",
							},
						},
					},
				},
			},
			wantViolation: true,
			expectedMsg:   "COMMIT",
		},
		{
			name: "ROLLBACK detected",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "7",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "UPDATE users SET active = 1;\nROLLBACK;",
							},
						},
					},
				},
			},
			wantViolation: true,
			expectedMsg:   "ROLLBACK",
		},
		{
			name: "no transaction keywords",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "8",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "INSERT INTO users (name, email) VALUES ('John', 'john@example.com');",
							},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "BEGIN in double-quoted string - should not detect",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "9",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  `INSERT INTO audit_log (message) VALUES ("BEGIN TRANSACTION detected");`,
							},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "COMMIT in multi-line comment - should not detect",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "10",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "/* This is a comment\nCOMMIT TRANSACTION\nEnd of comment */\nUPDATE users SET active = 1;",
							},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "createFunction with transaction - excluded",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "11",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "createFunction",
								SQL:  "CREATE FUNCTION test() RETURNS INT AS BEGIN BEGIN TRANSACTION; UPDATE users; COMMIT; RETURN 1; END;",
							},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "createTrigger with transaction - excluded",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "12",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "createTrigger",
								SQL:  "CREATE TRIGGER test_trigger AFTER INSERT ON users BEGIN START TRANSACTION; UPDATE logs; COMMIT; END;",
							},
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "BEGIN WORK detected (SQL Server syntax)",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "13",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "BEGIN WORK;\nINSERT INTO users VALUES (1);",
							},
						},
					},
				},
			},
			wantViolation: true,
			expectedMsg:   "BEGIN WORK",
		},
		{
			name: "SAVEPOINT TRANSACTION detected",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "14",
						Author:   "test",
						FilePath: "/test/changelog.sql",
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "UPDATE users SET active = 1;\nSAVEPOINT TRANSACTION sp1;",
							},
						},
					},
				},
			},
			wantViolation: true,
			expectedMsg:   "SAVEPOINT TRANSACTION",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Check(tt.changelog)

			if tt.wantViolation && len(violations) == 0 {
				t.Error("Expected violation but got none")
			}
			if !tt.wantViolation && len(violations) > 0 {
				t.Errorf("Expected no violations but got %d: %v", len(violations), violations)
			}
			if tt.wantViolation && tt.expectedMsg != "" && len(violations) > 0 {
				if !strings.Contains(violations[0].Message, tt.expectedMsg) {
					t.Errorf("Expected message to contain '%s', got '%s'", tt.expectedMsg, violations[0].Message)
				}
			}
		})
	}
}

func TestNoManualTransactionsRule_PreprocessSQL(t *testing.T) {
	rule := NewNoManualTransactionsRule(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "remove single-line comment",
			input:    "SELECT * FROM users; -- BEGIN TRANSACTION",
			expected: "SELECT * FROM users;",
		},
		{
			name:     "remove multi-line comment",
			input:    "SELECT * /* BEGIN TRANSACTION */ FROM users;",
			expected: "SELECT *  FROM users;",
		},
		{
			name:     "remove single-quoted string",
			input:    "INSERT INTO logs VALUES ('BEGIN TRANSACTION');",
			expected: "INSERT INTO logs VALUES ();",
		},
		{
			name:     "remove double-quoted string",
			input:    "INSERT INTO logs VALUES (\"START TRANSACTION\");",
			expected: "INSERT INTO logs VALUES ();",
		},
		{
			name:     "handle escaped single quotes",
			input:    "INSERT INTO logs VALUES ('It''s a BEGIN TRANSACTION');",
			expected: "INSERT INTO logs VALUES ();",
		},
		{
			name:     "multiple comments and strings",
			input:    "-- Comment with COMMIT\nSELECT 'BEGIN TRANSACTION' /* ROLLBACK */ FROM users;",
			expected: "SELECT  FROM users;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rule.preprocessSQL(tt.input)
			// Normalize whitespace for comparison (collapse multiple spaces to one)
			result = strings.Join(strings.Fields(strings.TrimSpace(result)), " ")
			expected := strings.Join(strings.Fields(strings.TrimSpace(tt.expected)), " ")
			if result != expected {
				t.Errorf("Expected '%s', got '%s'", expected, result)
			}
		})
	}
}

func TestNoManualTransactionsRule_CustomConfig(t *testing.T) {
	// Test with custom patterns
	cfg := &config.NoManualTransactionsConfig{
		Enabled: true,
		Patterns: []string{
			`\bBEGIN\b`, // Simplified pattern
		},
		CaseInsensitive:    true,
		ExcludeChangeTypes: []string{"sql"}, // Exclude sql changes
	}

	rule := NewNoManualTransactionsRule(cfg)

	changelog := &parser.Changelog{
		ChangeSets: []parser.ChangeSet{
			{
				ID:       "1",
				Author:   "test",
				FilePath: "/test/changelog.sql",
				Changes: []parser.Change{
					{
						Type: "sql",
						SQL:  "BEGIN TRANSACTION;",
					},
				},
			},
		},
	}

	violations := rule.Check(changelog)

	// Should not find violations because sql type is excluded
	if len(violations) > 0 {
		t.Errorf("Expected no violations with excluded change type, but got %d", len(violations))
	}
}

func TestNoManualTransactionsRule_ExcludePatterns(t *testing.T) {
	cfg := &config.NoManualTransactionsConfig{
		Enabled: true,
		Patterns: []string{
			`\bBEGIN\s+TRANSACTION\b`,
		},
		CaseInsensitive: true,
		ExcludePatterns: []string{"**/init/*", "test_*.sql"},
	}

	rule := NewNoManualTransactionsRule(cfg)

	tests := []struct {
		name          string
		filePath      string
		wantViolation bool
	}{
		{
			name:          "excluded by init pattern",
			filePath:      "/changelog/init/structure.sql",
			wantViolation: false,
		},
		{
			name:          "excluded by test pattern",
			filePath:      "/changelog/test_migration.sql",
			wantViolation: false,
		},
		{
			name:          "not excluded",
			filePath:      "/changelog/migration.sql",
			wantViolation: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changelog := &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: tt.filePath,
						Changes: []parser.Change{
							{
								Type: "sql",
								SQL:  "BEGIN TRANSACTION;",
							},
						},
					},
				},
			}

			violations := rule.Check(changelog)

			if tt.wantViolation && len(violations) == 0 {
				t.Error("Expected violation but got none")
			}
			if !tt.wantViolation && len(violations) > 0 {
				t.Errorf("Expected no violations but got %d", len(violations))
			}
		})
	}
}

func TestDBMSScopeRule_Check(t *testing.T) {
	cfg := &config.DBMSScopeConfig{
		RequireDBMS:   true,
		AllowedValues: []string{"mssql", "postgresql"},
		ForbiddenPreconditionPatterns: map[string][]string{
			"postgresql": {`(?i)(^|[^a-z0-9_])sys\.`, `(?i)\bobject_id\s*\(`, `(?i)\bcol_length\s*\(`},
		},
		ExcludePatterns: []string{"changelog/generated/**"},
	}

	tests := []struct {
		name         string
		wantMessages []string
		changeSet    parser.ChangeSet
		requireDBMS  bool
	}{
		{name: "missing dbms", changeSet: parser.ChangeSet{}, requireDBMS: true, wantMessages: []string{"has no dbms"}},
		{name: "missing dbms not required", changeSet: parser.ChangeSet{}, requireDBMS: false},
		{name: "known engine", changeSet: parser.ChangeSet{DBMSList: []string{"mssql"}}, requireDBMS: true},
		{name: "all and none", changeSet: parser.ChangeSet{DBMSList: []string{"all", "!none"}}, requireDBMS: true},
		{name: "typo", changeSet: parser.ChangeSet{DBMSList: []string{"postgres"}}, requireDBMS: true, wantMessages: []string{"unknown engine 'postgres'"}},
		{name: "negated known engine", changeSet: parser.ChangeSet{DBMSList: []string{"!postgresql"}}, requireDBMS: true},
		{name: "negated typo", changeSet: parser.ChangeSet{DBMSList: []string{"!oracle"}}, requireDBMS: true, wantMessages: []string{"unknown engine 'oracle'"}},
		{
			name:        "catalog precondition on other engine",
			changeSet:   parser.ChangeSet{DBMSList: []string{"mssql"}, PreconditionSQL: []string{"SELECT OBJECT_ID('a')"}},
			requireDBMS: true,
		},
		{
			name:         "catalog precondition on target engine",
			changeSet:    parser.ChangeSet{DBMSList: []string{"postgresql"}, PreconditionSQL: []string{"SELECT COUNT(*) FROM sys.tables"}},
			requireDBMS:  true,
			wantMessages: []string{"runs on postgresql (dbms:postgresql)"},
		},
		{
			name:         "catalog precondition with all",
			changeSet:    parser.ChangeSet{DBMSList: []string{"all"}, PreconditionSQL: []string{"SELECT COL_LENGTH('a', 'b')"}},
			requireDBMS:  true,
			wantMessages: []string{"col_length"},
		},
		{
			name:         "catalog precondition without dbms",
			changeSet:    parser.ChangeSet{PreconditionSQL: []string{"SELECT OBJECT_ID('a')"}},
			requireDBMS:  false,
			wantMessages: []string{"(dbms:all)"},
		},
		{
			name:        "identifier containing a forbidden word",
			changeSet:   parser.ChangeSet{DBMSList: []string{"postgresql"}, PreconditionSQL: []string{"SELECT my_object_id(1) FROM mysys.tables"}},
			requireDBMS: true,
		},
		{
			name:        "excluded file",
			changeSet:   parser.ChangeSet{FilePath: "/repo/changelog/generated/data.sql"},
			requireDBMS: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleCfg := *cfg
			ruleCfg.RequireDBMS = tt.requireDBMS
			rule := NewDBMSScopeRule(&ruleCfg)

			cs := tt.changeSet
			cs.ID, cs.Author = "1", "alice"
			if cs.FilePath == "" {
				cs.FilePath = "/repo/changelog/sprints/v1/changelog.sql"
			}

			violations := rule.Check(&parser.Changelog{ChangeSets: []parser.ChangeSet{cs}})
			if len(violations) != len(tt.wantMessages) {
				t.Fatalf("Check() returned %d violations, want %d: %+v", len(violations), len(tt.wantMessages), violations)
			}
			for i, want := range tt.wantMessages {
				if !strings.Contains(violations[i].Message, want) {
					t.Errorf("violation %d message = %q, want it to contain %q", i, violations[i].Message, want)
				}
				if violations[i].Rule != "dbms-scope" || violations[i].Author != "alice" || violations[i].ChangeSetID != "1" {
					t.Errorf("violation %d = %+v, want rule dbms-scope on alice:1", i, violations[i])
				}
			}
		})
	}
}

func TestDBMSScopeRule_Fixture(t *testing.T) {
	changelog, err := parser.Parse("../../testdata/dbms-scope/changelog.sql")
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	rule := NewDBMSScopeRule(&config.DBMSScopeConfig{
		RequireDBMS:   true,
		AllowedValues: []string{"mssql", "postgresql"},
		ForbiddenPreconditionPatterns: map[string][]string{
			"postgresql": {`(?i)(^|[^a-z0-9_])sys\.`},
		},
	})

	want := map[string]int{"no-dbms": 3, "typo": 6, "catalog-on-pg": 9}
	violations := rule.Check(changelog)
	if len(violations) != len(want) {
		t.Fatalf("Check() returned %d violations, want %d: %+v", len(violations), len(want), violations)
	}
	for _, v := range violations {
		line, ok := want[v.ChangeSetID]
		if !ok {
			t.Errorf("unexpected violation on %s: %s", v.ChangeSetID, v.Message)
			continue
		}
		if v.LineNumber != line || !strings.HasPrefix(v.Line, "--changeset alice:"+v.ChangeSetID) {
			t.Errorf("violation on %s points at line %d %q, want line %d", v.ChangeSetID, v.LineNumber, v.Line, line)
		}
	}
}

func TestDBMSScopeRule_DefaultConfig(t *testing.T) {
	rule := NewDBMSScopeRule(nil)
	changelog := &parser.Changelog{ChangeSets: []parser.ChangeSet{
		{ID: "1", Author: "alice", DBMSList: []string{"oracle", "!yugabytedb"}},
		{ID: "2", Author: "alice"},
	}}

	violations := rule.Check(changelog)
	if len(violations) != 1 || violations[0].ChangeSetID != "2" {
		t.Errorf("Check() = %+v, want only the missing dbms on changeset 2", violations)
	}
}

func TestDBMSTwinRule_Check(t *testing.T) {
	const file = "/repo/changelog/sprints/v10/changelog.sql"

	tests := []struct {
		name       string
		changeSets []parser.ChangeSet
		wantIDs    []string // changeset IDs with a violation, in order
	}{
		{
			name: "twin present",
			changeSets: []parser.ChangeSet{
				{ID: "1", Author: "alice", FilePath: file, DBMSList: []string{"mssql"}},
				{ID: "1-pg", Author: "alice", FilePath: file, DBMSList: []string{"postgresql"}},
			},
		},
		{
			name:       "twin missing",
			changeSets: []parser.ChangeSet{{ID: "1", Author: "alice", FilePath: file, DBMSList: []string{"mssql"}}},
			wantIDs:    []string{"1"},
		},
		{
			name: "twin by another author",
			changeSets: []parser.ChangeSet{
				{ID: "1", Author: "alice", FilePath: file, DBMSList: []string{"mssql"}},
				{ID: "1-pg", Author: "bob", FilePath: file, DBMSList: []string{"postgresql"}},
			},
			wantIDs: []string{"1"},
		},
		{
			name: "twin in another file",
			changeSets: []parser.ChangeSet{
				{ID: "1", Author: "alice", FilePath: file, DBMSList: []string{"mssql"}},
				{ID: "1-pg", Author: "alice", FilePath: "/repo/changelog/sprints/v10/other.sql", DBMSList: []string{"postgresql"}},
			},
			wantIDs: []string{"1"},
		},
		{
			name: "twin on the wrong engine",
			changeSets: []parser.ChangeSet{
				{ID: "1", Author: "alice", FilePath: file, DBMSList: []string{"mssql"}},
				{ID: "1-pg", Author: "alice", FilePath: file, DBMSList: []string{"mssql", "postgresql"}},
			},
			wantIDs: []string{"1-pg"},
		},
		{
			name:       "already reaches the target",
			changeSets: []parser.ChangeSet{{ID: "1", Author: "alice", FilePath: file, DBMSList: []string{"mssql", "postgresql"}}},
		},
		{
			name:       "does not name the source",
			changeSets: []parser.ChangeSet{{ID: "1", Author: "alice", FilePath: file, DBMSList: []string{"oracle"}}},
		},
		{
			name:       "before the since threshold",
			changeSets: []parser.ChangeSet{{ID: "1", Author: "alice", FilePath: "/repo/changelog/sprints/v9/changelog.sql", DBMSList: []string{"mssql"}}},
		},
		{
			name:       "outside the since pattern",
			changeSets: []parser.ChangeSet{{ID: "1", Author: "alice", FilePath: "/repo/changelog/init/changelog.sql", DBMSList: []string{"mssql"}}},
		},
		{
			name:       "excluded file",
			changeSets: []parser.ChangeSet{{ID: "1", Author: "alice", FilePath: "/repo/changelog/sprints/v10/generated/data.sql", DBMSList: []string{"mssql"}}},
		},
	}

	rule := NewDBMSTwinRule(&config.DBMSTwinConfig{
		SourceDBMS:       "mssql",
		TargetDBMS:       "postgresql",
		IDSuffix:         "-pg",
		SincePathPattern: `sprints/v(\d+)/`,
		Since:            10,
		ExcludePatterns:  []string{"changelog/sprints/v10/generated/**"},
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Check(&parser.Changelog{ChangeSets: tt.changeSets})
			gotIDs := make([]string, 0, len(violations))
			for _, v := range violations {
				gotIDs = append(gotIDs, v.ChangeSetID)
				if v.Rule != "dbms-twin" || v.Author == "" {
					t.Errorf("violation = %+v, want rule dbms-twin with an author", v)
				}
			}
			if strings.Join(gotIDs, ",") != strings.Join(tt.wantIDs, ",") {
				t.Errorf("violations on %v, want %v: %+v", gotIDs, tt.wantIDs, violations)
			}
		})
	}
}

func TestDBMSTwinRule_NoSinceThreshold(t *testing.T) {
	rule := NewDBMSTwinRule(&config.DBMSTwinConfig{SourceDBMS: "mssql", TargetDBMS: "postgresql", IDSuffix: "-pg"})
	changelog := &parser.Changelog{ChangeSets: []parser.ChangeSet{
		{ID: "1", Author: "alice", FilePath: "/repo/changelog/init/changelog.sql", DBMSList: []string{"MSSQL"}},
	}}

	if violations := rule.Check(changelog); len(violations) != 1 {
		t.Errorf("Check() returned %d violations, want 1 (every file is in scope without since_path_pattern)", len(violations))
	}
}

func TestDBMSTwinRule_Fixture(t *testing.T) {
	rule := NewDBMSTwinRule(&config.DBMSTwinConfig{
		SourceDBMS:       "mssql",
		TargetDBMS:       "postgresql",
		IDSuffix:         "-pg",
		SincePathPattern: `sprints/v(\d+)/`,
		Since:            10,
	})

	got := make(map[string]int)
	for _, path := range []string{"../../testdata/dbms-twin/sprints/v9/changelog.sql", "../../testdata/dbms-twin/sprints/v10/changelog.sql"} {
		changelog, err := parser.Parse(path)
		if err != nil {
			t.Fatalf("Parse(%s) failed: %v", path, err)
		}
		for _, v := range FilterSuppressedViolations(rule.Check(changelog), changelog) {
			got[v.ChangeSetID] = v.LineNumber
		}
	}

	// alice:2 has no twin, alice:4 has none reaching PostgreSQL, alice:4-pg is on the wrong
	// engine; alice:3 is suppressed and v9 is before the threshold.
	want := map[string]int{"2": 9, "4": 16, "4-pg": 19}
	if len(got) != len(want) {
		t.Fatalf("violations on %v, want %v", got, want)
	}
	for id, line := range want {
		if got[id] != line {
			t.Errorf("violation on %s at line %d, want line %d", id, got[id], line)
		}
	}
}
