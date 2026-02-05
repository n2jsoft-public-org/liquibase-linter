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
