// Package parser provides functionality for parsing Liquibase changelog files
// in various formats including XML, YAML, JSON, and SQL.
package parser

import (
	"fmt"
	"path/filepath"
	"strings"
)

// CircularIncludeError represents an error when circular includes are detected.
type CircularIncludeError struct {
	IncludeChain   []string
	IsSymlinkCycle bool
}

func (e *CircularIncludeError) Error() string {
	chain := strings.Join(e.IncludeChain, " → ")
	if e.IsSymlinkCycle {
		return fmt.Sprintf("circular include via symlinks detected: %s", chain)
	}
	return fmt.Sprintf("circular include detected: %s", chain)
}

// MaxDepthExceededError represents an error when max include depth is exceeded.
type MaxDepthExceededError struct {
	IncludeChain []string
	MaxDepth     int
}

func (e *MaxDepthExceededError) Error() string {
	chain := strings.Join(e.IncludeChain, " → ")
	return fmt.Sprintf("maximum include depth of %d exceeded: %s", e.MaxDepth, chain)
}

// parseContext holds context for parsing with includes
type parseContext struct {
	visitedFiles       map[string]bool
	symlinkResolutions map[string]string
	processedFiles     *[]string
	basePath           string
	includeChain       []string
	ignorePatterns     []string
	maxDepth           int
	currentDepth       int
	followSymlinks     bool
}

// newParseContext creates a new parse context with initial values
//
//nolint:unparam // maxDepth is a configurable parameter, even though currently always 10
func newParseContext(maxDepth int, followSymlinks bool) *parseContext {
	processedFiles := []string{}
	return &parseContext{
		visitedFiles:       make(map[string]bool),
		symlinkResolutions: make(map[string]string),
		processedFiles:     &processedFiles,
		currentDepth:       0,
		includeChain:       []string{},
		maxDepth:           maxDepth,
		followSymlinks:     followSymlinks,
		ignorePatterns:     []string{},
		basePath:           "",
	}
}

// SetIgnorePatterns sets ignore patterns and base path for filtering
func (ctx *parseContext) SetIgnorePatterns(patterns []string, basePath string) {
	ctx.ignorePatterns = patterns
	ctx.basePath = basePath
}

// ShouldIgnore checks if a file path should be ignored based on patterns
func (ctx *parseContext) ShouldIgnore(filePath string) bool {
	if len(ctx.ignorePatterns) == 0 || ctx.basePath == "" {
		return false
	}

	// Make file path relative to base path for pattern matching
	relPath, err := GetRelativePath(ctx.basePath, filePath)
	if err != nil {
		// If we can't make it relative, try matching with absolute path
		relPath = filePath
	}

	// Don't normalize - keep it relative for pattern matching
	// Just clean the path to handle . and .. segments
	relPath = filepath.Clean(relPath)

	for _, pattern := range ctx.ignorePatterns {
		matched, err := MatchesResourceFilter(relPath, pattern)
		if err == nil && matched {
			return true
		}
	}
	return false
}

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
	FilePath      string
	Format        ChangelogFormat
	ChangeSets    []ChangeSet
	IncludedFiles []string // List of all files processed (root + includes)
}

// ChangeSet represents a single changeset in a changelog.
type ChangeSet struct {
	Rollback        *Rollback
	Preconditions   *Precondition
	Comment         string
	FilePath        string
	Author          string
	Context         string
	ID              string
	LogicalFilePath string
	Changes         []Change
	Labels          []string
	DBMSList        []string
	RunAlways       bool
	RunOnChange     bool
	FailOnError     bool
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
	SQL     string
	Changes []Change
}

// Precondition represents a precondition for a changeset.
type Precondition struct {
	Attributes map[string]string
	Type       string
	OnFail     string
	OnError    string
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
	return ParseWithConfig(filePath, []string{}, "")
}

// ParseWithConfig parses a changelog file with ignore patterns for filtering includes.
// ignorePatterns: glob patterns to ignore during includeAll processing
// basePath: base path for relative pattern matching
func ParseWithConfig(filePath string, ignorePatterns []string, basePath string) (*Changelog, error) {
	format := DetectFormat(filePath)

	switch format {
	case FormatXML:
		parser := &XMLParser{}
		return parser.Parse(filePath)
	case FormatSQL:
		parser := &SQLParser{}
		return parser.Parse(filePath)
	case FormatYAML:
		parser := &YAMLParser{}
		return parser.ParseWithConfig(filePath, ignorePatterns, basePath)
	case FormatJSON:
		parser := &JSONParser{}
		return parser.ParseWithConfig(filePath, ignorePatterns, basePath)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", filePath)
	}
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

// IsDDLChange checks if a change is a Data Definition Language (DDL) operation.
// DDL operations modify database structure (tables, indexes, constraints, etc.).
func (c *Change) IsDDLChange() bool {
	// Normalize change type for comparison
	changeType := strings.ToLower(strings.TrimSpace(c.Type))
	changeType = strings.ReplaceAll(changeType, " ", "")

	// DDL change types
	ddlTypes := map[string]bool{
		"createtable":          true,
		"droptable":            true,
		"renametable":          true,
		"addcolumn":            true,
		"modifycolumn":         true,
		"dropcolumn":           true,
		"renamecolumn":         true,
		"altersequence":        true,
		"createindex":          true,
		"dropindex":            true,
		"createview":           true,
		"dropview":             true,
		"createprocedure":      true,
		"dropprocedure":        true,
		"addprimarykey":        true,
		"dropprimarykey":       true,
		"addunique":            true,
		"adduniqueconstraint":  true,
		"dropuniqueconstraint": true,
		"addforeignkey":        true,
		"dropforeignkey":       true,
		"adddefaultvalue":      true,
		"dropdefaultvalue":     true,
		"addnotnull":           true,
		"dropnotnull":          true,
		"createsequence":       true,
		"dropsequence":         true,
		"createfunction":       true,
		"dropfunction":         true,
		"createtrigger":        true,
		"droptrigger":          true,
		"setcolumnremarks":     true,
		"settableremarks":      true,
	}

	if ddlTypes[changeType] {
		return true
	}

	// Check SQL content for DDL keywords
	if c.SQL != "" {
		sqlUpper := strings.ToUpper(strings.TrimSpace(c.SQL))
		ddlKeywords := []string{
			"CREATE TABLE", "DROP TABLE", "ALTER TABLE", "RENAME TABLE",
			"CREATE INDEX", "DROP INDEX",
			"CREATE VIEW", "DROP VIEW",
			"CREATE PROCEDURE", "DROP PROCEDURE",
			"CREATE FUNCTION", "DROP FUNCTION",
			"CREATE TRIGGER", "DROP TRIGGER",
			"CREATE SEQUENCE", "DROP SEQUENCE",
			"ADD COLUMN", "DROP COLUMN", "MODIFY COLUMN", "ALTER COLUMN",
			"ADD CONSTRAINT", "DROP CONSTRAINT",
			"ADD PRIMARY KEY", "DROP PRIMARY KEY",
			"ADD FOREIGN KEY", "DROP FOREIGN KEY",
		}

		for _, keyword := range ddlKeywords {
			if strings.Contains(sqlUpper, keyword) {
				return true
			}
		}
	}

	return false
}

// IsDMLChange checks if a change is a Data Manipulation Language (DML) operation.
// DML operations modify data within tables (INSERT, UPDATE, DELETE, etc.).
func (c *Change) IsDMLChange() bool {
	// Normalize change type for comparison
	changeType := strings.ToLower(strings.TrimSpace(c.Type))
	changeType = strings.ReplaceAll(changeType, " ", "")

	// DML change types
	dmlTypes := map[string]bool{
		"insert":         true,
		"update":         true,
		"delete":         true,
		"loaddata":       true,
		"loadupdatedata": true,
	}

	if dmlTypes[changeType] {
		return true
	}

	// Check SQL content for DML keywords
	if c.SQL != "" {
		sqlUpper := strings.ToUpper(strings.TrimSpace(c.SQL))

		// Check if it starts with a DML keyword (to avoid matching CREATE, ALTER, etc.)
		dmlKeywords := []string{"INSERT", "UPDATE", "DELETE", "MERGE"}

		for _, keyword := range dmlKeywords {
			if strings.HasPrefix(sqlUpper, keyword) {
				return true
			}
		}
	}

	return false
}
