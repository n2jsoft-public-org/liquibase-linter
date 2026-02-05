package reporter

import (
	"encoding/json"
	"io"

	"github.com/n2jsoft/liquibase-linter/internal/rules"
)

// SARIFReporter formats output as SARIF (Static Analysis Results Interchange Format) 2.1.0
type SARIFReporter struct{}

// SARIF represents the top-level SARIF structure
type SARIF struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SARIFRun `json:"runs"`
}

// SARIFRun represents a single analysis run
type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

// SARIFTool represents the tool information
type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

// SARIFDriver represents the tool driver information
type SARIFDriver struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Rules   []SARIFReportRule `json:"rules"`
}

// SARIFReportRule represents a rule definition in SARIF
type SARIFReportRule struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	ShortDescription SARIFMessage             `json:"shortDescription"`
	FullDescription  SARIFMessage             `json:"fullDescription"`
	DefaultConfig    SARIFReportingDescriptor `json:"defaultConfiguration"`
}

// SARIFReportingDescriptor represents rule configuration
type SARIFReportingDescriptor struct {
	Level string `json:"level"`
}

// SARIFResult represents a single result (violation)
type SARIFResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   SARIFMessage    `json:"message"`
	Locations []SARIFLocation `json:"locations"`
}

// SARIFMessage represents a message in SARIF
type SARIFMessage struct {
	Text string `json:"text"`
}

// SARIFLocation represents a location in the code
type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

// SARIFPhysicalLocation represents a physical location
type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           *SARIFRegion          `json:"region,omitempty"`
}

// SARIFArtifactLocation represents a file location
type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

// SARIFRegion represents a region in a file
type SARIFRegion struct {
	StartLine int `json:"startLine"`
}

// Report implements the Reporter interface for SARIF output
func (r *SARIFReporter) Report(w io.Writer, result *Result) error {
	// Collect unique rules
	ruleMap := make(map[string]rules.Violation)
	for _, v := range result.Violations {
		if _, exists := ruleMap[v.Rule]; !exists {
			ruleMap[v.Rule] = v
		}
	}

	// Build SARIF rules
	sarifRules := make([]SARIFReportRule, 0, len(ruleMap))
	for ruleID, v := range ruleMap {
		sarifRules = append(sarifRules, SARIFReportRule{
			ID:   ruleID,
			Name: ruleID,
			ShortDescription: SARIFMessage{
				Text: v.Message,
			},
			FullDescription: SARIFMessage{
				Text: v.Message,
			},
			DefaultConfig: SARIFReportingDescriptor{
				Level: r.severityToLevel(v.Severity),
			},
		})
	}

	// Build SARIF results
	sarifResults := make([]SARIFResult, 0, len(result.Violations))
	for _, v := range result.Violations {
		location := SARIFLocation{
			PhysicalLocation: SARIFPhysicalLocation{
				ArtifactLocation: SARIFArtifactLocation{
					URI: v.FilePath,
				},
			},
		}

		if v.LineNumber > 0 {
			location.PhysicalLocation.Region = &SARIFRegion{
				StartLine: v.LineNumber,
			}
		}

		sarifResults = append(sarifResults, SARIFResult{
			RuleID: v.Rule,
			Level:  r.severityToLevel(v.Severity),
			Message: SARIFMessage{
				Text: v.Message,
			},
			Locations: []SARIFLocation{location},
		})
	}

	// Build SARIF output
	output := SARIF{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:    "liquibase-linter",
						Version: result.LinterVersion,
						Rules:   sarifRules,
					},
				},
				Results: sarifResults,
			},
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// severityToLevel converts a rules.Severity to a SARIF level
func (r *SARIFReporter) severityToLevel(severity rules.Severity) string {
	switch severity {
	case rules.SeverityCritical:
		return "error"
	case rules.SeverityWarning:
		return "warning"
	case rules.SeverityInfo:
		return "note"
	default:
		return "note"
	}
}
