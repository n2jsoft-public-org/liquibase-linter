package rules

import (
	"testing"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/config"
	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

func TestLabelPatternRule_Check(t *testing.T) {
	tests := []struct {
		config         *config.LabelPatternConfig
		changelog      *parser.Changelog
		name           string
		wantMessage    string
		wantViolations int
	}{
		{
			name: "valid - label matches pattern",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         `^v\d+$`,
				RequireLabel:    true,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{"v123"},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "invalid - label does not match pattern",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         `^v\d+$`,
				RequireLabel:    true,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{"sprint123"},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "Label 'sprint123' does not match required pattern",
		},
		{
			name: "invalid - missing required label",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         `^v\d+$`,
				RequireLabel:    true,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "Changeset lacks required label",
		},
		{
			name: "valid - label not required and missing",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         `^v\d+$`,
				RequireLabel:    false,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "invalid - empty label",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         `^v\d+$`,
				RequireLabel:    false,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{""},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "Changeset has empty label",
		},
		{
			name: "valid - multiple patterns, label matches one",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Patterns:        []string{`^v\d+$`, `^hotfix$`, `^init$`},
				RequireLabel:    true,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{"hotfix"},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "invalid - multiple patterns, label matches none",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Patterns:        []string{`^v\d+$`, `^hotfix$`},
				RequireLabel:    true,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{"bugfix"},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "Label 'bugfix' does not match required pattern",
		},
		{
			name: "valid - multiple labels, all match",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Patterns:        []string{`^v\d+$`, `^hotfix$`},
				RequireLabel:    true,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{"v123", "hotfix"},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "invalid - multiple labels, one does not match",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         `^v\d+$`,
				RequireLabel:    true,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{"v123", "invalid"},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "Label 'invalid' does not match required pattern",
		},
		{
			name: "valid - excluded by pattern **/init/**",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         `^v\d+$`,
				RequireLabel:    true,
				ExcludePatterns: []string{"**/init/**"},
			},
			changelog: &parser.Changelog{
				FilePath: "init/test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/init/tables.xml",
						Labels:   []string{},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "valid - excluded by custom pattern",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         `^v\d+$`,
				RequireLabel:    true,
				ExcludePatterns: []string{"db/test/**"},
			},
			changelog: &parser.Changelog{
				FilePath: "db/test/test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/test/tables.xml",
						Labels:   []string{},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "invalid - not excluded",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         `^v\d+$`,
				RequireLabel:    true,
				ExcludePatterns: []string{"**/init/**"},
			},
			changelog: &parser.Changelog{
				FilePath: "sprints/v123/test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/sprints/v123/tables.xml",
						Labels:   []string{},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "Changeset lacks required label",
		},
		{
			name: "valid - no patterns means any label is valid",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         "",
				Patterns:        []string{},
				RequireLabel:    true,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{"any-label"},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "invalid - no patterns but label still required",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         "",
				Patterns:        []string{},
				RequireLabel:    true,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "Changeset lacks required label",
		},
		{
			name: "valid - label with whitespace is trimmed",
			config: &config.LabelPatternConfig{
				Enabled:         true,
				Severity:        "warning",
				Pattern:         `^v\d+$`,
				RequireLabel:    true,
				ExcludePatterns: []string{},
			},
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "db/changelog/test.xml",
						Labels:   []string{"  v123  "},
					},
				},
			},
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewLabelPatternRule(tt.config)
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
