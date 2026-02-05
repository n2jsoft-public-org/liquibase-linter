package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/n2jsoft/liquibase-linter/internal/rules"
)

func TestGetReporter(t *testing.T) {
	tests := []struct {
		name      string
		format    Format
		colorize  bool
		wantError bool
	}{
		{"text format", FormatText, false, false},
		{"text with color", FormatText, true, false},
		{"json format", FormatJSON, false, false},
		{"sarif format", FormatSARIF, false, false},
		{"invalid format", Format("invalid"), false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter, err := GetReporter(tt.format, tt.colorize)
			if tt.wantError {
				if err == nil {
					t.Error("Expected error for invalid format")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if reporter == nil {
					t.Error("Expected non-nil reporter")
				}
			}
		})
	}
}

func TestNewSummary(t *testing.T) {
	result := &Result{
		Violations: []rules.Violation{
			{Severity: rules.SeverityCritical},
			{Severity: rules.SeverityCritical},
			{Severity: rules.SeverityWarning},
			{Severity: rules.SeverityInfo},
		},
		FilesChecked: 3,
		TotalTime:    100 * time.Millisecond,
	}

	summary := NewSummary(result)
	if summary.TotalViolations != 4 {
		t.Errorf("TotalViolations = %d, want 4", summary.TotalViolations)
	}
	if summary.CriticalCount != 2 {
		t.Errorf("CriticalCount = %d, want 2", summary.CriticalCount)
	}
	if summary.WarningCount != 1 {
		t.Errorf("WarningCount = %d, want 1", summary.WarningCount)
	}
	if summary.InfoCount != 1 {
		t.Errorf("InfoCount = %d, want 1", summary.InfoCount)
	}
	if summary.FileCount != 3 {
		t.Errorf("FileCount = %d, want 3", summary.FileCount)
	}
}

func createTestResult() *Result {
	return &Result{
		Violations: []rules.Violation{
			{
				Rule:        "sql-injection",
				Severity:    rules.SeverityCritical,
				Message:     "SQL injection detected",
				FilePath:    "test.xml",
				ChangeSetID: "1",
				Author:      "test",
			},
			{
				Rule:        "missing-rollback",
				Severity:    rules.SeverityWarning,
				Message:     "Changeset lacks rollback",
				FilePath:    "test.xml",
				ChangeSetID: "2",
				Author:      "test",
			},
		},
		FilesChecked:  1,
		TotalTime:     50 * time.Millisecond,
		Timestamp:     time.Date(2026, 2, 4, 12, 0, 0, 0, time.UTC),
		LinterVersion: "0.1.0",
	}
}

func TestTextReporter_ReportWithViolations(t *testing.T) {
	reporter := &TextReporter{Colorize: false}
	result := createTestResult()
	var buf bytes.Buffer

	err := reporter.Report(&buf, result)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "test.xml") {
		t.Error("Output should contain file name")
	}

	if !strings.Contains(output, "SQL injection detected") {
		t.Error("Output should contain first violation message")
	}
	if !strings.Contains(output, "Changeset lacks rollback") {
		t.Error("Output should contain second violation message")
	}

	if !strings.Contains(output, "SUMMARY") {
		t.Error("Output should contain summary")
	}
	if !strings.Contains(output, "Total Issues: 2") {
		t.Error("Output should contain correct issue count")
	}
}

func TestTextReporter_ReportNoViolations(t *testing.T) {
	reporter := &TextReporter{Colorize: false}
	result := &Result{
		Violations:   []rules.Violation{},
		FilesChecked: 1,
		TotalTime:    10 * time.Millisecond,
	}
	var buf bytes.Buffer

	err := reporter.Report(&buf, result)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No issues found") {
		t.Error("Output should contain success message")
	}
}

func TestTextReporter_Colorize(t *testing.T) {
	reporter := &TextReporter{Colorize: true}
	result := createTestResult()
	var buf bytes.Buffer

	err := reporter.Report(&buf, result)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\033[") {
		t.Error("Colorized output should contain ANSI codes")
	}
}

func TestJSONReporter_Report(t *testing.T) {
	reporter := &JSONReporter{}
	result := createTestResult()
	var buf bytes.Buffer

	err := reporter.Report(&buf, result)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	var output JSONOutput
	err = json.Unmarshal(buf.Bytes(), &output)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if output.Metadata.FilesChecked != 1 {
		t.Errorf("FilesChecked = %d, want 1", output.Metadata.FilesChecked)
	}
	if output.Metadata.LinterVersion != "0.1.0" {
		t.Errorf("LinterVersion = %s, want 0.1.0", output.Metadata.LinterVersion)
	}

	if output.Summary.TotalViolations != 2 {
		t.Errorf("TotalViolations = %d, want 2", output.Summary.TotalViolations)
	}
	if output.Summary.Critical != 1 {
		t.Errorf("Critical = %d, want 1", output.Summary.Critical)
	}
	if output.Summary.Warning != 1 {
		t.Errorf("Warning = %d, want 1", output.Summary.Warning)
	}

	if len(output.Violations) != 2 {
		t.Fatalf("Expected 2 violations, got %d", len(output.Violations))
	}
	v1 := output.Violations[0]
	if v1.Rule != "sql-injection" {
		t.Errorf("Rule = %s, want sql-injection", v1.Rule)
	}
	if v1.Severity != "critical" {
		t.Errorf("Severity = %s, want critical", v1.Severity)
	}
	if v1.Message != "SQL injection detected" {
		t.Errorf("Message = %s", v1.Message)
	}
}

func TestSARIFReporter_Report(t *testing.T) {
	reporter := &SARIFReporter{}
	result := createTestResult()
	var buf bytes.Buffer

	err := reporter.Report(&buf, result)
	if err != nil {
		t.Fatalf("Report failed: %v", err)
	}

	var output SARIF
	err = json.Unmarshal(buf.Bytes(), &output)
	if err != nil {
		t.Fatalf("Failed to parse SARIF: %v", err)
	}

	if output.Version != "2.1.0" {
		t.Errorf("Version = %s, want 2.1.0", output.Version)
	}

	if len(output.Runs) != 1 {
		t.Fatalf("Expected 1 run, got %d", len(output.Runs))
	}
	run := output.Runs[0]

	if run.Tool.Driver.Name != "liquibase-linter" {
		t.Errorf("Tool name = %s", run.Tool.Driver.Name)
	}
	if run.Tool.Driver.Version != "0.1.0" {
		t.Errorf("Tool version = %s", run.Tool.Driver.Version)
	}

	if len(run.Tool.Driver.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(run.Tool.Driver.Rules))
	}

	if len(run.Results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(run.Results))
	}
	result1 := run.Results[0]
	if result1.RuleID != "sql-injection" {
		t.Errorf("RuleID = %s", result1.RuleID)
	}
	if result1.Level != "error" {
		t.Errorf("Level = %s, want error", result1.Level)
	}
	if !strings.Contains(result1.Message.Text, "SQL injection detected") {
		t.Error("Message should contain violation text")
	}

	if len(result1.Locations) != 1 {
		t.Fatalf("Expected 1 location, got %d", len(result1.Locations))
	}
	location := result1.Locations[0]
	if location.PhysicalLocation.ArtifactLocation.URI != "test.xml" {
		t.Errorf("URI = %s", location.PhysicalLocation.ArtifactLocation.URI)
	}
}

func TestSARIFReporter_SeverityMapping(t *testing.T) {
	reporter := &SARIFReporter{}
	tests := []struct {
		severity rules.Severity
		want     string
	}{
		{rules.SeverityCritical, "error"},
		{rules.SeverityWarning, "warning"},
		{rules.SeverityInfo, "note"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := reporter.severityToLevel(tt.severity)
			if got != tt.want {
				t.Errorf("severityToLevel(%v) = %s, want %s", tt.severity, got, tt.want)
			}
		})
	}
}
