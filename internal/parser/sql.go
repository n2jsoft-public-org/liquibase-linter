package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// SQLParser parses SQL-formatted Liquibase changelogs.
type SQLParser struct{}

var (
	// Matches Liquibase SQL changeset headers like: --changeset author:id
	changesetRegex = regexp.MustCompile(`^--\s*changeset\s+([^:]+):(\S+)(.*)$`)
	// Matches rollback marker like: --rollback SQL statement
	rollbackRegex = regexp.MustCompile(`^--\s*rollback\s+(.*)$`)
	// Matches precondition marker like: --preconditions onFail:MARK_RAN
	preconditionsRegex = regexp.MustCompile(`^--\s*preconditions?\s+(.*)$`)
	// Matches comment lines
	commentRegex = regexp.MustCompile(`^--\s*comment\s+(.*)$`)
	// Matches label marker like: --labels: label1, label2
	labelsRegex = regexp.MustCompile(`^--\s*labels?:\s*(.*)$`)
	// Matches context marker like: --context: dev, test
	contextRegex = regexp.MustCompile(`^--\s*context:\s*(.*)$`)
	// Matches empty or whitespace-only lines
	emptyLineRegex = regexp.MustCompile(`^\s*$`)
	// Matches SQL comment lines
	sqlCommentRegex = regexp.MustCompile(`^\s*--`)
)

// Parse parses a SQL-formatted changelog file.
func (p *SQLParser) Parse(filePath string) (*Changelog, error) {
	//nolint:gosec // G304: File path is provided by user for parsing
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to close file: %v\n", closeErr)
		}
	}()

	changelog := &Changelog{
		FilePath:      filePath,
		Format:        FormatSQL,
		ChangeSets:    make([]ChangeSet, 0),
		IncludedFiles: []string{filePath}, // SQL parser doesn't support includes
	}

	scanner := bufio.NewScanner(file)
	var currentChangeSet *ChangeSet
	var currentSQL strings.Builder
	var rollbackSQL string
	var inRollback bool

	for scanner.Scan() {
		line := scanner.Text()

		// Check for changeset header
		if matches := changesetRegex.FindStringSubmatch(line); matches != nil {
			// Save previous changeset if exists
			if currentChangeSet != nil {
				p.finalizeChangeSet(currentChangeSet, currentSQL.String(), rollbackSQL)
				changelog.ChangeSets = append(changelog.ChangeSets, *currentChangeSet)
			}

			// Start new changeset
			currentChangeSet = &ChangeSet{
				Author:        strings.TrimSpace(matches[1]),
				ID:            strings.TrimSpace(matches[2]),
				FilePath:      filePath,
				Changes:       make([]Change, 0),
				Preconditions: nil,
			}

			// Parse attributes from the rest of the line
			p.parseChangeSetAttributes(currentChangeSet, matches[3])

			currentSQL.Reset()
			rollbackSQL = ""
			inRollback = false
			continue
		}

		// No changeset started yet, skip
		if currentChangeSet == nil {
			continue
		}

		// Check for rollback marker
		if matches := rollbackRegex.FindStringSubmatch(line); matches != nil {
			inRollback = true
			rollbackSQL = strings.TrimSpace(matches[1])
			continue
		}

		// Check for preconditions
		if matches := preconditionsRegex.FindStringSubmatch(line); matches != nil {
			p.parsePreconditions(currentChangeSet, matches[1])
			continue
		}

		// Check for comment
		if matches := commentRegex.FindStringSubmatch(line); matches != nil {
			currentChangeSet.Comment = strings.TrimSpace(matches[1])
			continue
		}

		// Check for labels
		if matches := labelsRegex.FindStringSubmatch(line); matches != nil {
			p.parseLabels(currentChangeSet, matches[1])
			continue
		}

		// Check for context
		if matches := contextRegex.FindStringSubmatch(line); matches != nil {
			currentChangeSet.Context = strings.TrimSpace(matches[1])
			continue
		}

		// Skip empty lines and comment-only lines
		if emptyLineRegex.MatchString(line) {
			continue
		}

		// Accumulate SQL
		if inRollback {
			if rollbackSQL != "" {
				rollbackSQL += "\n"
			}
			rollbackSQL += line
		} else if !sqlCommentRegex.MatchString(line) {
			currentSQL.WriteString(line)
			currentSQL.WriteString("\n")
		}
	}

	// Save last changeset
	if currentChangeSet != nil {
		p.finalizeChangeSet(currentChangeSet, currentSQL.String(), rollbackSQL)
		changelog.ChangeSets = append(changelog.ChangeSets, *currentChangeSet)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return changelog, nil
}

// CanParse checks if this parser can handle the given file.
func (p *SQLParser) CanParse(filePath string) bool {
	return DetectFormat(filePath) == FormatSQL
}

// parseChangeSetAttributes parses attributes from the changeset header line.
func (p *SQLParser) parseChangeSetAttributes(cs *ChangeSet, attributesStr string) {
	if attributesStr == "" {
		return
	}

	// Parse attributes like: runAlways:true runOnChange:false
	attributes := strings.Fields(attributesStr)
	for _, attr := range attributes {
		parts := strings.SplitN(attr, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "runalways":
			cs.RunAlways = strings.EqualFold(value, "true")
		case "runonchange":
			cs.RunOnChange = strings.EqualFold(value, "true")
		case "failonerror":
			cs.FailOnError = !strings.EqualFold(value, "false")
		case "context":
			cs.Context = value
		case "labels":
			p.parseLabels(cs, value)
		case "dbms":
			cs.DBMSList = strings.Split(value, ",")
			for i := range cs.DBMSList {
				cs.DBMSList[i] = strings.TrimSpace(cs.DBMSList[i])
			}
		case "logicalfilepath":
			cs.LogicalFilePath = value
		}
	}
}

// parseLabels parses labels from a string.
func (p *SQLParser) parseLabels(cs *ChangeSet, labelsStr string) {
	if labelsStr == "" {
		return
	}

	cs.Labels = strings.Split(labelsStr, ",")
	for i := range cs.Labels {
		cs.Labels[i] = strings.TrimSpace(cs.Labels[i])
	}
}

// parsePreconditions parses preconditions from a string.
func (p *SQLParser) parsePreconditions(cs *ChangeSet, precStr string) {
	// Simple parsing of common preconditions
	// In real implementation, this would be more sophisticated
	precondition := Precondition{
		Type:       "sql",
		Attributes: make(map[string]string),
	}

	// Parse onFail and onError attributes
	parts := strings.Fields(precStr)
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			key := strings.ToLower(strings.TrimSpace(kv[0]))
			value := strings.TrimSpace(kv[1])

			switch key {
			case "onfail":
				precondition.OnFail = value
			case "onerror":
				precondition.OnError = value
			default:
				precondition.Attributes[key] = value
			}
		}
	}

	if cs.Preconditions == nil {
		cs.Preconditions = &precondition
	}
}

// finalizeChangeSet processes the accumulated SQL and creates Change objects.
func (p *SQLParser) finalizeChangeSet(cs *ChangeSet, sql, rollbackSQL string) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return
	}

	// Analyze SQL to determine change type
	changeType := p.detectSQLChangeType(sql)

	change := Change{
		Type:       changeType,
		SQL:        sql,
		Attributes: make(map[string]string),
	}

	// Try to extract table name from SQL
	tableName := p.extractTableName(sql, changeType)
	if tableName != "" {
		change.TableName = tableName
		change.Attributes["tableName"] = tableName
	}

	cs.Changes = append(cs.Changes, change)

	// Add rollback if present
	if rollbackSQL != "" {
		cs.Rollback = &Rollback{
			SQL: strings.TrimSpace(rollbackSQL),
		}
	}
}

// detectSQLChangeType attempts to determine the type of SQL change.
func (p *SQLParser) detectSQLChangeType(sql string) string {
	upperSQL := strings.ToUpper(strings.TrimSpace(sql))

	// Remove leading comments
	lines := strings.Split(upperSQL, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			upperSQL = strings.Join(lines[i:], "\n")
			break
		}
	}

	switch {
	case strings.HasPrefix(upperSQL, "CREATE TABLE"):
		return "createTable"
	case strings.HasPrefix(upperSQL, "DROP TABLE"):
		return "dropTable"
	case strings.HasPrefix(upperSQL, "ALTER TABLE"):
		if strings.Contains(upperSQL, "ADD COLUMN") {
			return "addColumn"
		} else if strings.Contains(upperSQL, "DROP COLUMN") {
			return "dropColumn"
		}
		return "alterTable"
	case strings.HasPrefix(upperSQL, "CREATE INDEX"):
		return "createIndex"
	case strings.HasPrefix(upperSQL, "DROP INDEX"):
		return "dropIndex"
	case strings.HasPrefix(upperSQL, "CREATE UNIQUE INDEX"):
		return "createIndex"
	case strings.HasPrefix(upperSQL, "INSERT INTO"):
		return "insert"
	case strings.HasPrefix(upperSQL, "UPDATE"):
		return "update"
	case strings.HasPrefix(upperSQL, "DELETE FROM"):
		return "delete"
	case strings.HasPrefix(upperSQL, "TRUNCATE"):
		return "truncate"
	case strings.HasPrefix(upperSQL, "GRANT"):
		return "grant"
	case strings.HasPrefix(upperSQL, "REVOKE"):
		return "revoke"
	case strings.HasPrefix(upperSQL, "CREATE USER"):
		return "createUser"
	case strings.HasPrefix(upperSQL, "DROP USER"):
		return "dropUser"
	default:
		return "sql"
	}
}

// extractTableName attempts to extract table name from SQL.
func (p *SQLParser) extractTableName(sql, changeType string) string {
	upperSQL := strings.ToUpper(strings.TrimSpace(sql))

	// Define patterns for different SQL operations
	var pattern *regexp.Regexp

	switch changeType {
	case "createTable":
		pattern = regexp.MustCompile(`CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(?:(\w+)\.)?(\w+)`)
	case "dropTable":
		pattern = regexp.MustCompile(`DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?(?:(\w+)\.)?(\w+)`)
	case "alterTable", "addColumn", "dropColumn":
		pattern = regexp.MustCompile(`ALTER\s+TABLE\s+(?:(\w+)\.)?(\w+)`)
	case "createIndex":
		pattern = regexp.MustCompile(`CREATE\s+(?:UNIQUE\s+)?INDEX\s+\w+\s+ON\s+(?:(\w+)\.)?(\w+)`)
	case "dropIndex":
		// Index drops may or may not specify table
		pattern = regexp.MustCompile(`DROP\s+INDEX\s+\w+(?:\s+ON\s+(?:(\w+)\.)?(\w+))?`)
	case "insert":
		pattern = regexp.MustCompile(`INSERT\s+INTO\s+(?:(\w+)\.)?(\w+)`)
	case "update":
		pattern = regexp.MustCompile(`UPDATE\s+(?:(\w+)\.)?(\w+)`)
	case "delete":
		pattern = regexp.MustCompile(`DELETE\s+FROM\s+(?:(\w+)\.)?(\w+)`)
	case "truncate":
		pattern = regexp.MustCompile(`TRUNCATE\s+(?:TABLE\s+)?(?:(\w+)\.)?(\w+)`)
	}

	if pattern != nil {
		matches := pattern.FindStringSubmatch(upperSQL)
		if len(matches) >= 3 {
			// Return the table name (last capture group)
			return strings.ToLower(matches[len(matches)-1])
		}
	}

	return ""
}
