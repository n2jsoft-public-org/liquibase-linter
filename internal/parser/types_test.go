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
