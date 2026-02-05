package rules

import (
	"testing"

	"github.com/n2jsoft/liquibase-linter/internal/parser"
)

func TestNoIfExistsRule_Check(t *testing.T) {
	rule := NewNoIfExistsRule()

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "SQL Server IF EXISTS",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `IF EXISTS (SELECT * FROM sys.tables WHERE name = 'users')
								DROP TABLE users;`},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "SQL Server IF NOT EXISTS",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "2",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `IF NOT EXISTS (SELECT * FROM sys.columns WHERE name = 'email')
								ALTER TABLE users ADD email VARCHAR(255);`},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "SQL Server IF OBJECT_ID",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "3",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `IF OBJECT_ID('dbo.sp_GetUser', 'P') IS NOT NULL
								DROP PROCEDURE sp_GetUser;`},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "PostgreSQL DO block with IF EXISTS",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "4",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `DO $$ 
								BEGIN 
									IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'users') THEN
										CREATE TABLE users (id INT);
									END IF;
								END $$;`},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "MySQL DROP PROCEDURE IF EXISTS",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "5",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `DROP PROCEDURE IF EXISTS sp_GetUser;`},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "MySQL CREATE TABLE IF NOT EXISTS",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "6",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `CREATE TABLE IF NOT EXISTS users (
								id INT PRIMARY KEY
							);`},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "MySQL CREATE INDEX IF NOT EXISTS",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "7",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `CREATE INDEX IF NOT EXISTS idx_email ON users(email);`},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "Oracle BEGIN IF EXISTS block",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "8",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `BEGIN
								IF EXISTS (SELECT 1 FROM user_tables WHERE table_name = 'USERS') THEN
									EXECUTE IMMEDIATE 'DROP TABLE users';
								END IF;
							END;`},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "Case insensitive - if exists lowercase",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "9",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `if exists (select * from sys.tables where name = 'users')
								drop table users;`},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "Valid SQL without IF EXISTS",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "10",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `CREATE TABLE users (
								id INT PRIMARY KEY,
								email VARCHAR(255)
							);`},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Valid with Liquibase preconditions",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "11",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `ALTER TABLE users ADD COLUMN email VARCHAR(255);`},
						},
						Preconditions: &parser.Precondition{
							Type: "tableExists",
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "IF EXISTS in comment - should not trigger",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "12",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `-- Note: Use IF EXISTS for manual cleanup
							CREATE TABLE users (id INT PRIMARY KEY);`},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "IF EXISTS in string literal - should not trigger",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "13",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `INSERT INTO logs (message) VALUES ('Check IF EXISTS first');`},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Multiple changes, one with IF EXISTS",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "14",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: `CREATE TABLE users (id INT);`},
							{SQL: `IF EXISTS (SELECT * FROM sys.tables WHERE name = 'temp')
								DROP TABLE temp;`},
							{SQL: `CREATE TABLE orders (id INT);`},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "Empty SQL - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "15",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{SQL: ""},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Non-SQL change - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "16",
						Author:   "test",
						FilePath: "test.xml",
						Changes: []parser.Change{
							{Type: "createTable"},
						},
					},
				},
			},
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Check(tt.changelog)
			if len(violations) != tt.wantViolations {
				t.Errorf("NoIfExistsRule.Check() returned %d violations, want %d", len(violations), tt.wantViolations)
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

func TestNoIfExistsRule_Metadata(t *testing.T) {
	rule := NewNoIfExistsRule()

	if rule.ID() != "no-if-exists" {
		t.Errorf("ID() = %s, want 'no-if-exists'", rule.ID())
	}

	if rule.Name() != "No IF EXISTS" {
		t.Errorf("Name() = %s, want 'No IF EXISTS'", rule.Name())
	}

	if rule.Severity() != SeverityWarning {
		t.Errorf("Severity() = %v, want SeverityWarning", rule.Severity())
	}

	if rule.Description() == "" {
		t.Error("Description() should not be empty")
	}
}

func TestNoIfExistsRule_PreprocessSQL(t *testing.T) {
	rule := NewNoIfExistsRule()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Remove single-line comment",
			input: "SELECT * FROM users -- IF EXISTS (comment)",
			want:  "SELECT * FROM users ",
		},
		{
			name:  "Remove multi-line comment",
			input: "SELECT * FROM users /* IF EXISTS comment */",
			want:  "SELECT * FROM users ",
		},
		{
			name:  "Remove single-quoted string",
			input: "INSERT INTO logs VALUES ('IF EXISTS check')",
			want:  "INSERT INTO logs VALUES ()",
		},
		{
			name:  "Remove double-quoted identifier",
			input: `SELECT * FROM "IF EXISTS table"`,
			want:  "SELECT * FROM ",
		},
		{
			name: "Mixed removal",
			input: `-- Comment with IF EXISTS
			SELECT * FROM users WHERE name = 'IF NOT EXISTS' /* block comment */`,
			want: "\n\t\t\tSELECT * FROM users WHERE name =  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.preprocessSQL(tt.input)
			if got != tt.want {
				t.Errorf("preprocessSQL() = %q, want %q", got, tt.want)
			}
		})
	}
}
