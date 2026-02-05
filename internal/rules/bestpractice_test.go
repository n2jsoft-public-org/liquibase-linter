package rules

import (
	"testing"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/config"
	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

func TestMissingRollbackRule_Check(t *testing.T) {
	rule := &MissingRollbackRule{}

	tests := []struct {
		changelog      *parser.Changelog
		name           string
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
	tests := []struct {
		changelog      *parser.Changelog
		name           string
		wantMessage    string
		config         config.RuleConfig
		wantViolations int
	}{
		{
			name: "risky-only mode: createTable with preconditions - no violation",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeRiskyOnly,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
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
			name: "risky-only mode: createTable without preconditions - violation",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeRiskyOnly,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:            "1",
						Author:        "test",
						FilePath:      "db/changelog/test.xml",
						Preconditions: nil,
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "Changeset with risky operation 'createTable' requires preconditions (mode: risky-only)",
		},
		{
			name: "risky-only mode: runAlways set - no violation",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeRiskyOnly,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:        "1",
						Author:    "test",
						FilePath:  "db/changelog/test.xml",
						RunAlways: true,
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "risky-only mode: non-risky operation without preconditions - no violation",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeRiskyOnly,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Changes: []parser.Change{
							{Type: "insert", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "all mode: any changeset without preconditions - violation",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeAll,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Changes: []parser.Change{
							{Type: "insert", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "Changeset requires preconditions (mode: all)",
		},
		{
			name: "all mode: changeset with preconditions - no violation",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeAll,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Preconditions: &parser.Precondition{
							Type: "tableExists",
						},
						Changes: []parser.Change{
							{Type: "insert", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "exclude pattern: **/init/** - no violation",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeRiskyOnly,
				ExcludePatterns: []string{"**/init/**"},
			},
			changelog: &parser.Changelog{
				FilePath: "init/test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/init/0-structure/tables.xml",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "exclude pattern: **/seed/** - no violation",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeRiskyOnly,
				ExcludePatterns: []string{"**/seed/**"},
			},
			changelog: &parser.Changelog{
				FilePath: "seed/test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/seed/data.xml",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "exclude pattern: custom pattern db/test/** - no violation",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeRiskyOnly,
				ExcludePatterns: []string{"db/test/**"},
			},
			changelog: &parser.Changelog{
				FilePath: "db/test/test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/test/tables.xml",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "not matching exclude pattern - violation",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeRiskyOnly,
				ExcludePatterns: []string{"**/init/**"},
			},
			changelog: &parser.Changelog{
				FilePath: "sprints/v123/test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/sprints/v123/0-structure/tables.xml",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "Changeset with risky operation 'createTable' requires preconditions (mode: risky-only)",
		},
		{
			name: "default mode when empty - defaults to risky-only",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            "", // Empty mode
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Changes: []parser.Change{
							{Type: "insert", TableName: "users"},
						},
					},
				},
			},
			wantViolations: 0, // insert is not risky, so no violation in risky-only mode
		},
		{
			name: "multiple risky operations - reports once per changeset",
			config: config.RuleConfig{
				Enabled:         true,
				Severity:        "warning",
				Mode:            config.ModeRiskyOnly,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Changes: []parser.Change{
							{Type: "createTable", TableName: "users"},
							{Type: "createIndex", IndexName: "idx_users"},
							{Type: "addForeignKey"},
						},
					},
				},
			},
			wantViolations: 1, // Only one violation per changeset
			wantMessage:    "Changeset with risky operation 'createTable' requires preconditions (mode: risky-only)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewNonIdempotentChangesRule(tt.config)
			violations := rule.Check(tt.changelog)
			if len(violations) != tt.wantViolations {
				t.Errorf("Expected %d violations, got %d", tt.wantViolations, len(violations))
			}
			if tt.wantMessage != "" && len(violations) > 0 {
				if violations[0].Message != tt.wantMessage {
					t.Errorf("Expected message %q, got %q", tt.wantMessage, violations[0].Message)
				}
			}
		})
	}
}

func TestNamingConventionRule_Check(t *testing.T) {
	rule := &NamingConventionRule{}

	tests := []struct {
		changelog      *parser.Changelog
		name           string
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
		changelog      *parser.Changelog
		name           string
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
		changelog      *parser.Changelog
		name           string
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
		changelog      *parser.Changelog
		name           string
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
		changelog      *parser.Changelog
		name           string
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
		changelog      *parser.Changelog
		name           string
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
