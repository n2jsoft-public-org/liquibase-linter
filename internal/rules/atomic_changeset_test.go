package rules

import (
	"strings"
	"testing"

	"github.com/n2jsoft/liquibase-linter/internal/config"
	"github.com/n2jsoft/liquibase-linter/internal/parser"
)

func TestAtomicChangesetRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		changelog      *parser.Changelog
		config         *config.AtomicChangesetConfig
		wantViolations int
	}{
		{
			name: "Single SQL statement - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "test.sql",
						Changes: []parser.Change{
							{SQL: "CREATE TABLE users (id INT PRIMARY KEY);"},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:          true,
				MaxSQLStatements: 1,
			},
			wantViolations: 0,
		},
		{
			name: "Multiple SQL statements - violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "2",
						Author:   "test",
						FilePath: "test.sql",
						Changes: []parser.Change{
							{SQL: `CREATE TABLE users (id INT);
								CREATE INDEX idx_email ON users(email);
								CREATE INDEX idx_name ON users(name);`},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:          true,
				MaxSQLStatements: 1,
			},
			wantViolations: 1,
		},
		{
			name: "Multiple changes in XML - violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "3",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{Type: "createTable"},
							{Type: "createIndex"},
							{Type: "addColumn"},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:          true,
				MaxSQLStatements: 1,
			},
			wantViolations: 1,
		},
		{
			name: "Table with index - allowed when configured",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "4",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{Type: "createTable"},
							{Type: "createIndex"},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:               true,
				AllowTableWithIndexes: true,
				MaxSQLStatements:      1,
			},
			wantViolations: 0,
		},
		{
			name: "Table with index - violation when not allowed",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "5",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{Type: "createTable"},
							{Type: "createIndex"},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:               true,
				AllowTableWithIndexes: false,
				MaxSQLStatements:      1,
			},
			wantViolations: 1,
		},
		{
			name: "SQL with semicolons in comments - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "6",
						Author:   "test",
						FilePath: "test.sql",
						Changes: []parser.Change{
							{SQL: `-- This is a comment with; semicolons;
								CREATE TABLE users (id INT);`},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:          true,
				MaxSQLStatements: 1,
			},
			wantViolations: 0,
		},
		{
			name: "SQL with semicolons in strings - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "7",
						Author:   "test",
						FilePath: "test.sql",
						Changes: []parser.Change{
							{SQL: `INSERT INTO logs (message) VALUES ('Statement; with; semicolons;');`},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:          true,
				MaxSQLStatements: 1,
			},
			wantViolations: 0,
		},
		{
			name: "Multiple statements with higher max allowed - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "9",
						Author:   "test",
						FilePath: "test.sql",
						Changes: []parser.Change{
							{SQL: `CREATE TABLE users (id INT);
								CREATE INDEX idx_email ON users(email);`},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:          true,
				MaxSQLStatements: 2,
			},
			wantViolations: 0,
		},
		{
			name: "Empty changeset - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "10",
						Author:   "test",
						FilePath: "test.xml",
						Changes:  []parser.Change{},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:          true,
				MaxSQLStatements: 1,
			},
			wantViolations: 0,
		},
		{
			name: "Rule disabled - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "11",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{Type: "createTable"},
							{Type: "createIndex"},
							{Type: "addColumn"},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:          false,
				MaxSQLStatements: 1,
			},
			wantViolations: 0,
		},
		{
			name: "Complex SQL with multiple statements",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "12",
						Author:   "test",
						FilePath: "test.sql",
						Changes: []parser.Change{
							{SQL: `
								-- Create users table
								CREATE TABLE users (
									id INT PRIMARY KEY,
									email VARCHAR(255)
								);
								
								-- Create index
								CREATE INDEX idx_email ON users(email);
								
								/* Insert default data */
								INSERT INTO users (id, email) VALUES (1, 'admin@example.com');
							`},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:          true,
				MaxSQLStatements: 1,
			},
			wantViolations: 1,
		},
		{
			name: "Single statement without semicolon - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "13",
						Author:   "test",
						FilePath: "test.sql",
						Changes: []parser.Change{
							{SQL: `CREATE TABLE users (id INT PRIMARY KEY, email VARCHAR(255))`},
						},
					},
				},
			},
			config: &config.AtomicChangesetConfig{
				Enabled:          true,
				MaxSQLStatements: 1,
			},
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewAtomicChangesetRule(tt.config)
			violations := rule.Check(tt.changelog)

			if len(violations) != tt.wantViolations {
				t.Errorf("Check() returned %d violations, want %d", len(violations), tt.wantViolations)
				for _, v := range violations {
					t.Logf("  Violation: %s", v.Message)
				}
			}

			// Verify violation properties when expected
			if tt.wantViolations > 0 && len(violations) > 0 {
				v := violations[0]
				if v.Rule != rule.ID() {
					t.Errorf("Violation.Rule = %s, want %s", v.Rule, rule.ID())
				}
				if v.Severity != rule.Severity() {
					t.Errorf("Violation.Severity = %v, want %v", v.Severity, rule.Severity())
				}
				if v.Message == "" {
					t.Error("Violation.Message is empty")
				}
			}
		})
	}
}

func TestAtomicChangesetRule_Metadata(t *testing.T) {
	rule := NewAtomicChangesetRule(&config.AtomicChangesetConfig{Enabled: true})

	if rule.ID() != "atomic-changeset" {
		t.Errorf("ID() = %s, want 'atomic-changeset'", rule.ID())
	}

	if rule.Name() != "Atomic Changeset" {
		t.Errorf("Name() = %s, want 'Atomic Changeset'", rule.Name())
	}

	if rule.Severity() != SeverityInfo {
		t.Errorf("Severity() = %v, want SeverityInfo", rule.Severity())
	}

	if rule.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestAtomicChangesetRule_CountSQLStatements(t *testing.T) {
	rule := NewAtomicChangesetRule(&config.AtomicChangesetConfig{Enabled: true})

	tests := []struct {
		name string
		sql  string
		want int
	}{
		{
			name: "Single statement with semicolon",
			sql:  "CREATE TABLE users (id INT);",
			want: 1,
		},
		{
			name: "Single statement without semicolon",
			sql:  "CREATE TABLE users (id INT)",
			want: 1,
		},
		{
			name: "Two statements",
			sql:  "CREATE TABLE users (id INT); CREATE TABLE orders (id INT);",
			want: 2,
		},
		{
			name: "Three statements",
			sql:  "CREATE TABLE users (id INT); CREATE INDEX idx_id ON users(id); INSERT INTO users VALUES (1);",
			want: 3,
		},
		{
			name: "Statement with semicolon in comment",
			sql:  "-- Comment with; semicolon\nCREATE TABLE users (id INT);",
			want: 1,
		},
		{
			name: "Statement with semicolon in string",
			sql:  "INSERT INTO logs VALUES ('Message; with; semicolons');",
			want: 1,
		},
		{
			name: "Empty SQL",
			sql:  "",
			want: 0,
		},
		{
			name: "Only comments",
			sql:  "-- Just a comment\n/* Another comment */",
			want: 0,
		},
		{
			name: "Complex multiline",
			sql: `
				CREATE TABLE users (
					id INT PRIMARY KEY
				);
				
				CREATE INDEX idx_email ON users(email);
			`,
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.countSQLStatements(tt.sql)
			if got != tt.want {
				t.Errorf("countSQLStatements() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAtomicChangesetRule_PreprocessSQL(t *testing.T) {
	rule := NewAtomicChangesetRule(&config.AtomicChangesetConfig{Enabled: true})

	tests := []struct {
		name  string
		input string
		check func(string) bool
	}{
		{
			name:  "Remove single-line comment",
			input: "SELECT * FROM users -- comment; with; semicolons",
			check: func(output string) bool {
				return !strings.Contains(output, "comment")
			},
		},
		{
			name:  "Remove multi-line comment",
			input: "SELECT * FROM users /* comment; with; semicolons */",
			check: func(output string) bool {
				return !strings.Contains(output, "comment")
			},
		},
		{
			name:  "Remove single-quoted string",
			input: "INSERT INTO logs VALUES ('string; with; semicolons')",
			check: func(output string) bool {
				return !strings.Contains(output, "string")
			},
		},
		{
			name:  "Remove double-quoted identifier",
			input: `SELECT * FROM "table; with; semicolons"`,
			check: func(output string) bool {
				return !strings.Contains(output, "table")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.preprocessSQL(tt.input)
			if !tt.check(got) {
				t.Errorf("preprocessSQL() = %q, check failed", got)
			}
		})
	}
}
