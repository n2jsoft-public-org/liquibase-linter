package parser

import (
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     ChangelogFormat
	}{
		{"XML file", "changelog.xml", FormatXML},
		{"SQL file", "changelog.sql", FormatSQL},
		{"YAML file", "changelog.yaml", FormatYAML},
		{"JSON file", "changelog.json", FormatJSON},
		{"Unknown extension", "changelog.txt", FormatUnknown},
		{"No extension", "changelog", FormatUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectFormat(tt.filePath)
			if got != tt.want {
				t.Errorf("DetectFormat(%s) = %v, want %v", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestChangelogFormat_String(t *testing.T) {
	tests := []struct {
		format ChangelogFormat
		want   string
	}{
		{FormatXML, "xml"},
		{FormatSQL, "sql"},
		{FormatYAML, "yaml"},
		{FormatJSON, "json"},
		{FormatUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.format.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeSet_HasRollback(t *testing.T) {
	tests := []struct {
		name string
		cs   ChangeSet
		want bool
	}{
		{
			name: "with rollback",
			cs: ChangeSet{
				ID:     "1",
				Author: "test",
				Rollback: &Rollback{
					SQL: "DROP TABLE test",
				},
			},
			want: true,
		},
		{
			name: "without rollback",
			cs: ChangeSet{
				ID:       "1",
				Author:   "test",
				Rollback: nil,
			},
			want: false,
		},
		{
			name: "with empty rollback",
			cs: ChangeSet{
				ID:       "1",
				Author:   "test",
				Rollback: &Rollback{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cs.HasRollback()
			if got != tt.want {
				t.Errorf("HasRollback() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeSet_HasPreconditions(t *testing.T) {
	tests := []struct {
		name string
		cs   ChangeSet
		want bool
	}{
		{
			name: "with preconditions",
			cs: ChangeSet{
				ID:     "1",
				Author: "test",
				Preconditions: &Precondition{
					OnFail: "MARK_RAN",
				},
			},
			want: true,
		},
		{
			name: "without preconditions",
			cs: ChangeSet{
				ID:            "1",
				Author:        "test",
				Preconditions: nil,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cs.HasPreconditions()
			if got != tt.want {
				t.Errorf("HasPreconditions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSuppressions(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		want    []string
	}{
		{
			name:    "single rule",
			comment: "liquibase-linter:disable sql-injection",
			want:    []string{"sql-injection"},
		},
		{
			name:    "multiple rules comma-separated",
			comment: "liquibase-linter:disable sql-injection,missing-rollback",
			want:    []string{"sql-injection", "missing-rollback"},
		},
		{
			name:    "multiple rules with spaces",
			comment: "liquibase-linter:disable sql-injection, missing-rollback, hardcoded-credentials",
			want:    []string{"sql-injection", "missing-rollback", "hardcoded-credentials"},
		},
		{
			name:    "case insensitive directive",
			comment: "LIQUIBASE-LINTER:DISABLE sql-injection",
			want:    []string{"sql-injection"},
		},
		{
			name:    "mixed case",
			comment: "Liquibase-Linter:Disable sql-injection",
			want:    []string{"sql-injection"},
		},
		{
			name:    "with extra whitespace around colon",
			comment: "liquibase-linter : disable sql-injection",
			want:    []string{"sql-injection"},
		},
		{
			name:    "with extra whitespace after disable",
			comment: "liquibase-linter:disable   sql-injection",
			want:    []string{"sql-injection"},
		},
		{
			name:    "comment with additional text before",
			comment: "This is a comment. liquibase-linter:disable sql-injection",
			want:    []string{"sql-injection"},
		},
		{
			name:    "comment with additional text after",
			comment: "liquibase-linter:disable sql-injection because we know what we're doing",
			want:    []string{"sql-injection"},
		},
		{
			name:    "no directive",
			comment: "Just a regular comment",
			want:    nil,
		},
		{
			name:    "empty comment",
			comment: "",
			want:    nil,
		},
		{
			name:    "invalid format (no rule)",
			comment: "liquibase-linter:disable",
			want:    nil,
		},
		{
			name:    "rules with underscores",
			comment: "liquibase-linter:disable sql_injection,missing_rollback",
			want:    []string{"sql_injection", "missing_rollback"},
		},
		{
			name:    "rules with numbers",
			comment: "liquibase-linter:disable rule1,rule2",
			want:    []string{"rule1", "rule2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSuppressions(tt.comment)
			if len(got) != len(tt.want) {
				t.Errorf("ParseSuppressions() returned %d rules, want %d\nGot: %v\nWant: %v",
					len(got), len(tt.want), got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseSuppressions() rule[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestChangeSet_IsSuppressed(t *testing.T) {
	tests := []struct {
		name   string
		ruleID string
		cs     ChangeSet
		want   bool
	}{
		{
			name: "rule is suppressed",
			cs: ChangeSet{
				ID:              "1",
				Author:          "test",
				SuppressedRules: []string{"sql-injection", "missing-rollback"},
			},
			ruleID: "sql-injection",
			want:   true,
		},
		{
			name: "rule is not suppressed",
			cs: ChangeSet{
				ID:              "1",
				Author:          "test",
				SuppressedRules: []string{"sql-injection"},
			},
			ruleID: "missing-rollback",
			want:   false,
		},
		{
			name: "no suppressions",
			cs: ChangeSet{
				ID:              "1",
				Author:          "test",
				SuppressedRules: []string{},
			},
			ruleID: "sql-injection",
			want:   false,
		},
		{
			name: "nil suppressions",
			cs: ChangeSet{
				ID:              "1",
				Author:          "test",
				SuppressedRules: nil,
			},
			ruleID: "sql-injection",
			want:   false,
		},
		{
			name: "case insensitive matching",
			cs: ChangeSet{
				ID:              "1",
				Author:          "test",
				SuppressedRules: []string{"SQL-Injection"},
			},
			ruleID: "sql-injection",
			want:   true,
		},
		{
			name: "case insensitive matching reverse",
			cs: ChangeSet{
				ID:              "1",
				Author:          "test",
				SuppressedRules: []string{"sql-injection"},
			},
			ruleID: "SQL-INJECTION",
			want:   true,
		},
		{
			name: "whitespace handling",
			cs: ChangeSet{
				ID:              "1",
				Author:          "test",
				SuppressedRules: []string{" sql-injection "},
			},
			ruleID: "sql-injection",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cs.IsSuppressed(tt.ruleID)
			if got != tt.want {
				t.Errorf("IsSuppressed(%q) = %v, want %v", tt.ruleID, got, tt.want)
			}
		})
	}
}
