package reporter

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/rules"
)

// TextReporter formats output as human-readable text
type TextReporter struct {
	Colorize bool
}

// Report implements the Reporter interface for text output
func (r *TextReporter) Report(w io.Writer, result *Result) error {
	if len(result.Violations) == 0 {
		fmt.Fprintf(w, "%sNo issues found!%s\n", r.color(colorGreen), r.colorReset())
		fmt.Fprintf(w, "Checked %d file(s) in %v\n", result.FilesChecked, result.TotalTime)
		return nil
	}

	// Group violations by file
	fileViolations := r.groupByFile(result.Violations)

	// Print violations grouped by file
	for _, file := range r.sortedKeys(fileViolations) {
		fmt.Fprintf(w, "\n%s%s%s\n", r.color(colorCyan), file, r.colorReset())
		for _, v := range fileViolations[file] {
			r.writeViolation(w, v)
		}
	}

	// Print summary
	fmt.Fprintln(w, "")
	r.writeSummary(w, NewSummary(result))

	return nil
}

// groupByFile groups violations by their file path
func (r *TextReporter) groupByFile(violations []rules.Violation) map[string][]rules.Violation {
	grouped := make(map[string][]rules.Violation)
	for _, v := range violations {
		grouped[v.FilePath] = append(grouped[v.FilePath], v)
	}
	return grouped
}

// sortedKeys returns sorted keys from a map
func (r *TextReporter) sortedKeys(m map[string][]rules.Violation) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// writeViolation writes a single violation
func (r *TextReporter) writeViolation(w io.Writer, v rules.Violation) {
	icon := r.formatSeverity(v.Severity)
	color := r.severityColor(v.Severity)

	fmt.Fprintf(w, "  %s %s[%s]%s %s\n",
		icon,
		r.color(color),
		strings.ToUpper(v.Severity.String()),
		r.colorReset(),
		v.Message,
	)

	if v.ChangeSetID != "" {
		fmt.Fprintf(w, "    Changeset: %s (author: %s)\n", v.ChangeSetID, v.Author)
	}
	if v.Line != "" {
		if v.LineNumber > 0 {
			fmt.Fprintf(w, "    Line %d: %s\n", v.LineNumber, v.Line)
		} else {
			fmt.Fprintf(w, "    SQL: %s\n", v.Line)
		}
	}
	fmt.Fprintf(w, "    Rule: %s\n", v.Rule)
}

// writeSummary writes the summary section
func (r *TextReporter) writeSummary(w io.Writer, summary Summary) {
	fmt.Fprintf(w, "%s=== SUMMARY ===%s\n", r.color(colorBold), r.colorReset())
	fmt.Fprintf(w, "Total Issues: %d\n", summary.TotalViolations)

	if summary.CriticalCount > 0 {
		fmt.Fprintf(w, "  %sCritical: %d%s\n",
			r.color(colorRed), summary.CriticalCount, r.colorReset())
	}
	if summary.WarningCount > 0 {
		fmt.Fprintf(w, "  %sWarning:  %d%s\n",
			r.color(colorYellow), summary.WarningCount, r.colorReset())
	}
	if summary.InfoCount > 0 {
		fmt.Fprintf(w, "  %sInfo:     %d%s\n",
			r.color(colorBlue), summary.InfoCount, r.colorReset())
	}

	fmt.Fprintf(w, "Files Checked: %d\n", summary.FileCount)
}

// formatSeverity returns an icon for the severity level
func (r *TextReporter) formatSeverity(severity rules.Severity) string {
	switch severity {
	case rules.SeverityCritical:
		return "✖"
	case rules.SeverityWarning:
		return "⚠"
	case rules.SeverityInfo:
		return "ℹ"
	default:
		return "•"
	}
}

// severityColor returns the color code for a severity level
func (r *TextReporter) severityColor(severity rules.Severity) string {
	switch severity {
	case rules.SeverityCritical:
		return colorRed
	case rules.SeverityWarning:
		return colorYellow
	case rules.SeverityInfo:
		return colorBlue
	default:
		return ""
	}
}

// color returns the ANSI color code if colorization is enabled
func (r *TextReporter) color(code string) string {
	if r.Colorize {
		return code
	}
	return ""
}

// colorReset returns the ANSI reset code if colorization is enabled
func (r *TextReporter) colorReset() string {
	if r.Colorize {
		return "\033[0m"
	}
	return ""
}

// ANSI color codes
const (
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)
