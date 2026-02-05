// Package reporter provides functionality for formatting and outputting linting results.
package reporter

import (
	"fmt"
	"io"
	"time"

	"github.com/n2jsoft/liquibase-linter/internal/rules"
)

// Format represents the output format type
type Format string

const (
	FormatText  Format = "text"
	FormatJSON  Format = "json"
	FormatSARIF Format = "sarif"
)

// Reporter defines the interface for formatting and outputting linting results
type Reporter interface {
	// Report formats and writes the linting results to the provided writer
	Report(w io.Writer, result *Result) error
}

// Result contains the complete linting results
type Result struct {
	Violations    []rules.Violation
	FilesChecked  int
	TotalTime     time.Duration
	Timestamp     time.Time
	LinterVersion string
}

// Summary provides aggregated statistics about violations
type Summary struct {
	TotalViolations int
	CriticalCount   int
	WarningCount    int
	InfoCount       int
	FileCount       int
}

// NewSummary creates a Summary from a Result
func NewSummary(result *Result) Summary {
	summary := Summary{
		TotalViolations: len(result.Violations),
		FileCount:       result.FilesChecked,
	}

	for _, v := range result.Violations {
		switch v.Severity {
		case rules.SeverityCritical:
			summary.CriticalCount++
		case rules.SeverityWarning:
			summary.WarningCount++
		case rules.SeverityInfo:
			summary.InfoCount++
		}
	}

	return summary
}

// GetReporter returns a Reporter for the specified format
func GetReporter(format Format, colorize bool) (Reporter, error) {
	switch format {
	case FormatText:
		return &TextReporter{Colorize: colorize}, nil
	case FormatJSON:
		return &JSONReporter{}, nil
	case FormatSARIF:
		return &SARIFReporter{}, nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
