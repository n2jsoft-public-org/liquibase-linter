// Package parser provides functionality for parsing Liquibase changelog files
// in various formats including XML, YAML, JSON, and SQL.
package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ChangelogFormat represents the format of a changelog file.
type ChangelogFormat string

const (
	// FormatXML represents XML format changelogs.
	FormatXML ChangelogFormat = "xml"
	// FormatYAML represents YAML format changelogs.
	FormatYAML ChangelogFormat = "yaml"
	// FormatJSON represents JSON format changelogs.
	FormatJSON ChangelogFormat = "json"
	// FormatSQL represents SQL format changelogs.
	FormatSQL ChangelogFormat = "sql"
	// FormatUnknown represents an unknown format.
	FormatUnknown ChangelogFormat = "unknown"
)

// String returns the string representation of the format.
func (f ChangelogFormat) String() string {
	return string(f)
}

// Changelog represents a parsed Liquibase changelog file.
type Changelog struct {
	FilePath   string
	Format     ChangelogFormat
	ChangeSets []ChangeSet
}

// ChangeSet represents a single changeset in a changelog.
type ChangeSet struct {
	ID              string
	Author          string
	FilePath        string
	Changes         []Change
	Rollback        *Rollback
	Preconditions   *Precondition
	Context         string
	Labels          []string
	DBMSList        []string
	RunAlways       bool
	RunOnChange     bool
	FailOnError     bool
	Comment         string
	LogicalFilePath string
}

// Change represents a database change within a changeset.
type Change struct {
	Type       string
	Attributes map[string]string
	SQL        string
	TableName  string
	ColumnName string
	IndexName  string
}

// Rollback represents rollback instructions for a changeset.
type Rollback struct {
	Changes []Change
	SQL     string
}

// Precondition represents a precondition for a changeset.
type Precondition struct {
	Type       string
	OnFail     string
	OnError    string
	Attributes map[string]string
}

// Parser is the interface that all format-specific parsers must implement.
type Parser interface {
	Parse(filePath string) (*Changelog, error)
	CanParse(filePath string) bool
}

// DetectFormat determines the format of a changelog file based on its extension.
func DetectFormat(filePath string) ChangelogFormat {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".xml":
		return FormatXML
	case ".yaml", ".yml":
		return FormatYAML
	case ".json":
		return FormatJSON
	case ".sql":
		return FormatSQL
	default:
		return FormatUnknown
	}
}

// Parse parses a changelog file using the appropriate parser based on format detection.
func Parse(filePath string) (*Changelog, error) {
	format := DetectFormat(filePath)
	var parser Parser

	switch format {
	case FormatXML:
		parser = &XMLParser{}
	case FormatSQL:
		parser = &SQLParser{}
	case FormatYAML, FormatJSON:
		return nil, fmt.Errorf("format %s not yet implemented (Phase 6)", format)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", filePath)
	}

	return parser.Parse(filePath)
}

// HasRollback checks if a changeset has rollback instructions.
func (cs *ChangeSet) HasRollback() bool {
	return cs.Rollback != nil && (len(cs.Rollback.Changes) > 0 || cs.Rollback.SQL != "")
}

// HasPreconditions checks if a changeset has preconditions.
func (cs *ChangeSet) HasPreconditions() bool {
	return cs.Preconditions != nil
}

// GetChangeType returns a human-readable description of the change type.
func (c *Change) GetChangeType() string {
	if c.Type != "" {
		return c.Type
	}
	if c.SQL != "" {
		return "sql"
	}
	return "unknown"
}
