package reporter

import (
	"encoding/json"
	"io"
)

// JSONReporter formats output as JSON
type JSONReporter struct{}

// JSONOutput represents the JSON output structure
type JSONOutput struct {
	Metadata   JSONMetadata    `json:"metadata"`
	Summary    JSONSummary     `json:"summary"`
	Violations []JSONViolation `json:"violations"`
}

// JSONMetadata contains metadata about the linting run
type JSONMetadata struct {
	LinterVersion string `json:"linter_version"`
	Timestamp     string `json:"timestamp"`
	FilesChecked  int    `json:"files_checked"`
	TotalTimeMs   int64  `json:"total_time_ms"`
}

// JSONSummary contains violation counts
type JSONSummary struct {
	TotalViolations int `json:"total_violations"`
	Critical        int `json:"critical"`
	Warning         int `json:"warning"`
	Info            int `json:"info"`
}

// JSONViolation represents a single violation in JSON format
type JSONViolation struct {
	Rule          string `json:"rule"`
	Severity      string `json:"severity"`
	Message       string `json:"message"`
	FilePath      string `json:"file_path"`
	LineNumber    int    `json:"line_number,omitempty"`
	Line          string `json:"line,omitempty"`
	SQLLineNumber int    `json:"sql_line_number,omitempty"` // Line number within the SQL content
	ChangeSetID   string `json:"changeset_id,omitempty"`
	Author        string `json:"author,omitempty"`
}

// Report implements the Reporter interface for JSON output
func (r *JSONReporter) Report(w io.Writer, result *Result) error {
	summary := NewSummary(result)

	output := JSONOutput{
		Metadata: JSONMetadata{
			LinterVersion: result.LinterVersion,
			Timestamp:     result.Timestamp.Format("2006-01-02T15:04:05Z"),
			FilesChecked:  result.FilesChecked,
			TotalTimeMs:   result.TotalTime.Milliseconds(),
		},
		Summary: JSONSummary{
			TotalViolations: summary.TotalViolations,
			Critical:        summary.CriticalCount,
			Warning:         summary.WarningCount,
			Info:            summary.InfoCount,
		},
		Violations: make([]JSONViolation, 0, len(result.Violations)),
	}

	for _, v := range result.Violations {
		jsonViolation := JSONViolation{
			Rule:        v.Rule,
			Severity:    v.Severity.String(),
			Message:     v.Message,
			FilePath:    v.FilePath,
			Line:        v.Line,
			ChangeSetID: v.ChangeSetID,
			Author:      v.Author,
		}
		// Set line numbers: sql_line_number for lines within SQL content
		if v.LineNumber > 0 {
			jsonViolation.SQLLineNumber = v.LineNumber
			jsonViolation.LineNumber = v.LineNumber // Also populate line_number for backward compatibility
		}
		output.Violations = append(output.Violations, jsonViolation)
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
