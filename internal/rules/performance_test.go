package rules

import (
	"testing"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

func TestMissingIndexRule_Check(t *testing.T) {
	rule := &MissingIndexRule{}

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "foreign key with index",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{Type: "addForeignKey", TableName: "orders", ColumnName: "user_id"},
							{Type: "createIndex", TableName: "orders", ColumnName: "user_id"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "foreign key without index",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{Type: "addForeignKey", TableName: "orders", ColumnName: "user_id"},
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

func TestTableLockRule_Check(t *testing.T) {
	rule := &TableLockRule{}

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "ALTER TABLE with context",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:      "1",
						Author:  "test",
						Context: "dev",
						Changes: []parser.Change{
							{SQL: "ALTER TABLE users ADD COLUMN email VARCHAR(100)"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "ALTER TABLE without context",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:      "1",
						Author:  "test",
						Context: "",
						Changes: []parser.Change{
							{SQL: "ALTER TABLE users ADD COLUMN email VARCHAR(100)"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "addColumn change type without context",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:      "1",
						Author:  "test",
						Context: "",
						Changes: []parser.Change{
							{Type: "addColumn", TableName: "users"},
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

func TestLargeDataOperationRule_Check(t *testing.T) {
	rule := &LargeDataOperationRule{}

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "UPDATE with WHERE",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "UPDATE users SET active = 1 WHERE id = 123"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "UPDATE without WHERE",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "UPDATE users SET active = 1;"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "DELETE without WHERE",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "DELETE FROM users;"},
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

func TestSelectStarRule_Check(t *testing.T) {
	rule := &SelectStarRule{}

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "SELECT with explicit columns",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "SELECT id, name FROM users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "SELECT *",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "SELECT * FROM users"},
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
