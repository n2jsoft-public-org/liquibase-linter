package rules

import (
	"testing"

	"github.com/n2jsoft/liquibase-linter/internal/parser"
)

func TestMissingRollbackRule_Check(t *testing.T) {
	rule := &MissingRollbackRule{}

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "with rollback",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Rollback: &parser.Rollback{
							SQL: "DROP TABLE users",
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "without rollback",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						Rollback: nil,
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
			if len(violations) > 0 && violations[0].Severity != SeverityWarning {
				t.Errorf("Expected warning severity, got %v", violations[0].Severity)
			}
		})
	}
}

func TestNonIdempotentChangesRule_Check(t *testing.T) {
	rule := &NonIdempotentChangesRule{}

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "createTable with preconditions",
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
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "createTable without preconditions",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:            "1",
						Author:        "test",
						Preconditions: nil,
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "runAlways set",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:        "1",
						Author:    "test",
						RunAlways: true,
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
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

func TestNamingConventionRule_Check(t *testing.T) {
	rule := &NamingConventionRule{}

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "valid snake_case",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "user_accounts"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "invalid camelCase",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "userAccounts"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "invalid uppercase",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "USERS"},
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

func TestChangesetDocumentationRule_Check(t *testing.T) {
	rule := &ChangesetDocumentationRule{}

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "with comment",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:      "1",
						Author:  "test",
						Comment: "Create users table",
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "without comment",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:      "1",
						Author:  "test",
						Comment: "",
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

func TestContextMisuseRule_Check(t *testing.T) {
	rule := &ContextMisuseRule{}

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "dangerous op with context",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:      "1",
						Author:  "test",
						Context: "dev",
						Changes: []parser.Change{
							{SQL: "DROP TABLE temp"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "dangerous op without context",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:      "1",
						Author:  "test",
						Context: "",
						Changes: []parser.Change{
							{SQL: "DROP TABLE users"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "safe op without context",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:      "1",
						Author:  "test",
						Context: "",
						Changes: []parser.Change{
							{SQL: "CREATE TABLE users (id INT)"},
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
