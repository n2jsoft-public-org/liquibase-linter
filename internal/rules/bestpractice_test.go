package rules

import (
	"testing"

	"github.com/n2jsoft/liquibase-linter/internal/config"
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

func TestSprintFolderStructureRule_Check(t *testing.T) {
	cfg := &config.FileStructureConfig{
		Enabled:          true,
		SprintPattern:    `^v\d+$`,
		StructurePattern: `(?i)^\d+\s*-\s*structure$`,
		DataPattern:      `(?i)^\d+\s*-\s*data$`,
		ExcludePatterns:  []string{"**/init/**"},
		SprintBasePath:   "",
	}
	rule := NewSprintFolderStructureRule(cfg)

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "valid sprint structure",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v116/0 - structure/create_users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v116/0 - structure/create_users.sql",
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "valid sprint data",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v117/1 - data/insert_users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v117/1 - data/insert_users.sql",
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "excluded init folder",
			changelog: &parser.Changelog{
				FilePath: "changelog/init/0 - structure/tables/users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/init/0 - structure/tables/users.sql",
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "invalid - no sprint folder",
			changelog: &parser.Changelog{
				FilePath: "changelog/structure/create_users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/structure/create_users.sql",
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "invalid - sprint without structure/data subfolder",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v116/create_users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v116/create_users.sql",
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

func TestDDLLocationRule_Check(t *testing.T) {
	cfg := &config.FileStructureConfig{
		Enabled:          true,
		StructurePattern: `(?i)^\d+\s*-\s*structure$`,
		ExcludePatterns:  []string{"**/init/**"},
	}
	rule := NewDDLLocationRule(cfg)

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "DDL in structure directory",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v116/0 - structure/create_users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v116/0 - structure/create_users.sql",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "DDL SQL in structure directory",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v116/0-structure/alter_table.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v116/0-structure/alter_table.sql",
						Changes: []parser.Change{
							{SQL: "ALTER TABLE users ADD COLUMN age INT"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "DDL in data directory - violation",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v116/1 - data/create_table.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v116/1 - data/create_table.sql",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "DDL in init folder - excluded",
			changelog: &parser.Changelog{
				FilePath: "changelog/init/tables/users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/init/tables/users.sql",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "DML in data directory - no violation",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v116/1 - data/insert_users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v116/1 - data/insert_users.sql",
						Changes: []parser.Change{
							{Type: "insert", TableName: "users"},
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

func TestDMLLocationRule_Check(t *testing.T) {
	cfg := &config.FileStructureConfig{
		Enabled:         true,
		DataPattern:     `(?i)^\d+\s*-\s*data$`,
		ExcludePatterns: []string{"**/init/**"},
	}
	rule := NewDMLLocationRule(cfg)

	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "DML in data directory",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v116/1 - data/insert_users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v116/1 - data/insert_users.sql",
						Changes: []parser.Change{
							{Type: "insert", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "DML SQL in data directory",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v116/1-data/update_users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v116/1-data/update_users.sql",
						Changes: []parser.Change{
							{SQL: "UPDATE users SET status = 'active'"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "DML in structure directory - violation",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v116/0 - structure/insert_users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v116/0 - structure/insert_users.sql",
						Changes: []parser.Change{
							{Type: "insert", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "DML in init folder - excluded",
			changelog: &parser.Changelog{
				FilePath: "changelog/init/data/users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/init/data/users.sql",
						Changes: []parser.Change{
							{Type: "insert", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "DDL in structure directory - no violation",
			changelog: &parser.Changelog{
				FilePath: "changelog/sprints/v116/0 - structure/create_users.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "changelog/sprints/v116/0 - structure/create_users.sql",
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
