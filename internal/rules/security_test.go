package rules

import (
	"testing"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

func TestSQLInjectionRule_Check(t *testing.T) {
	rule := &SQLInjectionRule{}

	tests := []struct {
		changelog      *parser.Changelog
		name           string
		wantViolations int
	}{
		{
			name: "clean SQL",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "CREATE TABLE users (id INT PRIMARY KEY)"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "SQL with variable interpolation",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "INSERT INTO users (name) VALUES ('${username}')"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "SQL with string concatenation",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "SELECT * FROM users WHERE name = 'test' + ${var}"},
						},
					},
				},
			},
			wantViolations: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Check(tt.changelog)
			if len(violations) != tt.wantViolations {
				t.Errorf("Expected %d violations, got %d", tt.wantViolations, len(violations))
			}
			if len(violations) > 0 && violations[0].Severity != SeverityCritical {
				t.Errorf("Expected critical severity, got %v", violations[0].Severity)
			}
		})
	}
}

func TestHardcodedCredentialsRule_Check(t *testing.T) {
	rule := &HardcodedCredentialsRule{}

	tests := []struct {
		changelog      *parser.Changelog
		name           string
		wantViolations int
	}{
		{
			name: "no credentials",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "CREATE TABLE users (id INT)"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "hardcoded password",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "CREATE USER 'admin' IDENTIFIED BY 'secret123'"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "hardcoded API key",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "UPDATE config SET api_key = 'abc123xyz' WHERE id = 1"},
						},
					},
				},
			},
			wantViolations: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Check(tt.changelog)
			if len(violations) != tt.wantViolations {
				t.Errorf("Expected %d violations, got %d", tt.wantViolations, len(violations))
			}
		})
	}
}

func TestDangerousOperationsRule_Check(t *testing.T) {
	rule := &DangerousOperationsRule{}

	tests := []struct {
		changelog      *parser.Changelog
		name           string
		wantViolations int
	}{
		{
			name: "safe operation",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "CREATE TABLE users (id INT)"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "DROP TABLE without preconditions",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:            "1",
						Author:        "test",
						Preconditions: nil,
						Context:       "",
						Changes: []parser.Change{
							{SQL: "DROP TABLE old_users"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "DROP TABLE with preconditions",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Preconditions: &parser.Precondition{
							Type: "tableExists",
						},
						Changes: []parser.Change{
							{SQL: "DROP TABLE old_users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "TRUNCATE with context",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:      "1",
						Author:  "test",
						Context: "dev",
						Changes: []parser.Change{
							{SQL: "TRUNCATE TABLE temp_data"},
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
				t.Errorf("Expected %d violations, got %d", tt.wantViolations, len(violations))
			}
		})
	}
}

func TestPrivilegeEscalationRule_Check(t *testing.T) {
	rule := &PrivilegeEscalationRule{}

	tests := []struct {
		changelog      *parser.Changelog
		name           string
		wantViolations int
	}{
		{
			name: "normal grant",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "GRANT SELECT ON mydb.users TO 'reader'@'localhost'"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "GRANT ALL privileges",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "GRANT ALL PRIVILEGES ON *.* TO 'admin'@'%'"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "wildcard host",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "CREATE USER 'user'@'%' IDENTIFIED BY 'password'"},
						},
					},
				},
			},
			wantViolations: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Check(tt.changelog)
			if len(violations) != tt.wantViolations {
				t.Errorf("Expected %d violations, got %d", tt.wantViolations, len(violations))
			}
		})
	}
}
