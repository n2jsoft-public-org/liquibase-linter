// Package parser provides functionality for parsing Liquibase changelog files.
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLParser parses YAML-formatted Liquibase changelogs.
type YAMLParser struct{}

// yamlDatabaseChangeLog represents the root structure of a YAML changelog
type yamlDatabaseChangeLog struct {
	DatabaseChangeLog []map[string]any `yaml:"databaseChangeLog" json:"databaseChangeLog"`
}

// Parse parses a YAML changelog file.
func (p *YAMLParser) Parse(filePath string) (*Changelog, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML
	var doc yamlDatabaseChangeLog
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Create parse context with default config values
	// These will be overridden if config is available
	ctx := newParseContext(10, true)
	absPath, _ := filepath.Abs(filePath)
	ctx.includeChain = []string{absPath}

	// Parse the changelog with context
	return p.parseWithContext(filePath, doc, ctx)
}

// parseWithContext parses a YAML changelog with include tracking
func (p *YAMLParser) parseWithContext(filePath string, doc yamlDatabaseChangeLog, ctx *parseContext) (*Changelog, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve file path: %w", err)
	}

	// Check if already visited (circular include detection)
	normalizedPath := NormalizePath(absPath)
	if ctx.visitedFiles[normalizedPath] {
		// Check if this is due to symlinks
		isSymlinkCycle := false
		if ctx.followSymlinks {
			if resolved, ok := ctx.symlinkResolutions[absPath]; ok && resolved != absPath {
				isSymlinkCycle = true
			}
		}
		return nil, &CircularIncludeError{
			IncludeChain:   append(ctx.includeChain, absPath),
			IsSymlinkCycle: isSymlinkCycle,
		}
	}

	// Check depth limit
	if ctx.currentDepth >= ctx.maxDepth {
		return nil, &MaxDepthExceededError{
			IncludeChain: ctx.includeChain,
			MaxDepth:     ctx.maxDepth,
		}
	}

	// Mark as visited
	ctx.visitedFiles[normalizedPath] = true

	// Track symlink resolution
	if ctx.followSymlinks {
		realPath, err := filepath.EvalSymlinks(absPath)
		if err == nil && realPath != absPath {
			ctx.symlinkResolutions[absPath] = realPath
		}
	}

	changelog := &Changelog{
		FilePath:   absPath,
		Format:     FormatYAML,
		ChangeSets: []ChangeSet{},
	}

	// Process each element in the databaseChangeLog array
	for _, element := range doc.DatabaseChangeLog {
		// Each element can be a changeSet, include, includeAll, or property
		for key, value := range element {
			switch key {
			case "changeSet":
				// Parse changeset
				cs, err := p.parseChangeSet(value, absPath)
				if err != nil {
					return nil, err
				}
				changelog.ChangeSets = append(changelog.ChangeSets, cs)

			case "include":
				// Handle include directive
				included, err := p.handleInclude(value, absPath, ctx)
				if err != nil {
					return nil, fmt.Errorf("in file %s: %w", absPath, err)
				}
				changelog.ChangeSets = append(changelog.ChangeSets, included.ChangeSets...)

			case "includeAll":
				// Handle includeAll directive
				// TODO Phase 6+: includeAll context and labels attributes not yet supported
				included, err := p.handleIncludeAll(value, absPath, ctx)
				if err != nil {
					return nil, fmt.Errorf("in file %s: %w", absPath, err)
				}
				changelog.ChangeSets = append(changelog.ChangeSets, included...)

			case "property":
				// Properties are metadata, skip for now
				continue

			default:
				// Unknown element type, skip
				continue
			}
		}
	}

	return changelog, nil
}

// parseChangeSet parses a changeSet element
func (p *YAMLParser) parseChangeSet(data any, filePath string) (ChangeSet, error) {
	csMap, ok := data.(map[string]any)
	if !ok {
		return ChangeSet{}, fmt.Errorf("changeSet must be an object, got %T", data)
	}

	cs := ChangeSet{
		FilePath: filePath,
		Changes:  []Change{},
	}

	// Extract ID (required)
	if id, ok := csMap["id"]; ok {
		cs.ID = fmt.Sprintf("%v", id)
	} else {
		return ChangeSet{}, fmt.Errorf("changeSet missing required 'id' field")
	}

	// Extract Author (required)
	if author, ok := csMap["author"]; ok {
		cs.Author = fmt.Sprintf("%v", author)
	} else {
		return ChangeSet{}, fmt.Errorf("changeSet[%s] missing required 'author' field", cs.ID)
	}

	// Extract optional fields
	if context, ok := csMap["context"].(string); ok {
		cs.Context = context
	}
	if labels, ok := csMap["labels"].(string); ok {
		cs.Labels = strings.Split(labels, ",")
	}
	if dbms, ok := csMap["dbms"].(string); ok {
		cs.DBMSList = strings.Split(dbms, ",")
	}
	if runAlways, ok := csMap["runAlways"].(bool); ok {
		cs.RunAlways = runAlways
	}
	if runOnChange, ok := csMap["runOnChange"].(bool); ok {
		cs.RunOnChange = runOnChange
	}
	if failOnError, ok := csMap["failOnError"].(bool); ok {
		cs.FailOnError = failOnError
	}
	if comment, ok := csMap["comment"].(string); ok {
		cs.Comment = comment
	}
	if logicalFilePath, ok := csMap["logicalFilePath"].(string); ok {
		cs.LogicalFilePath = logicalFilePath
	}

	// Parse changes
	if changes, ok := csMap["changes"].([]any); ok {
		for _, changeData := range changes {
			changeMap, ok := changeData.(map[string]any)
			if !ok {
				continue
			}

			// Each change is a map with one key (the change type)
			for changeType, changeValue := range changeMap {
				change, err := ConvertChange(changeType, changeValue, filePath, cs.ID)
				if err != nil {
					return ChangeSet{}, err
				}
				cs.Changes = append(cs.Changes, change)
			}
		}
	}

	// Parse rollback
	if rollback, ok := csMap["rollback"]; ok {
		rb, err := p.parseRollback(rollback, filePath, cs.ID)
		if err != nil {
			return ChangeSet{}, err
		}
		cs.Rollback = rb
	}

	// Parse preconditions
	if preconditions, ok := csMap["preConditions"]; ok {
		pc, err := p.parsePreconditions(preconditions, filePath, cs.ID)
		if err != nil {
			return ChangeSet{}, err
		}
		cs.Preconditions = pc
	}

	return cs, nil
}

// parseRollback parses rollback instructions
func (p *YAMLParser) parseRollback(data any, filePath, changesetID string) (*Rollback, error) {
	rb := &Rollback{
		Changes: []Change{},
	}

	switch v := data.(type) {
	case string:
		// Simple SQL rollback
		rb.SQL = v
	case []any:
		// Array of changes
		for _, changeData := range v {
			changeMap, ok := changeData.(map[string]any)
			if !ok {
				continue
			}
			for changeType, changeValue := range changeMap {
				change, err := ConvertChange(changeType, changeValue, filePath, changesetID)
				if err != nil {
					return nil, err
				}
				rb.Changes = append(rb.Changes, change)
			}
		}
	case map[string]any:
		// Single change or SQL
		if sql, ok := v["sql"].(string); ok {
			rb.SQL = sql
		} else {
			// It's a single change
			for changeType, changeValue := range v {
				change, err := ConvertChange(changeType, changeValue, filePath, changesetID)
				if err != nil {
					return nil, err
				}
				rb.Changes = append(rb.Changes, change)
			}
		}
	}

	return rb, nil
}

// parsePreconditions parses precondition element
func (p *YAMLParser) parsePreconditions(data any, filePath, changesetID string) (*Precondition, error) {
	pcMap, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("preconditions must be an object")
	}

	pc := &Precondition{
		Attributes: make(map[string]string),
	}

	if onFail, ok := pcMap["onFail"].(string); ok {
		pc.OnFail = onFail
	}
	if onError, ok := pcMap["onError"].(string); ok {
		pc.OnError = onError
	}

	// Extract other attributes
	for k, v := range pcMap {
		if k != "onFail" && k != "onError" {
			if strVal, ok := v.(string); ok {
				pc.Attributes[k] = strVal
			} else {
				pc.Attributes[k] = fmt.Sprintf("%v", v)
			}
		}
	}

	// Determine precondition type (simplified)
	if len(pc.Attributes) > 0 {
		for k := range pc.Attributes {
			pc.Type = k
			break
		}
	}

	return pc, nil
}

// handleInclude processes an include directive
func (p *YAMLParser) handleInclude(data any, baseFilePath string, ctx *parseContext) (*Changelog, error) {
	includeMap, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("include must be an object")
	}

	fileAttr, ok := includeMap["file"].(string)
	if !ok {
		return nil, fmt.Errorf("include missing required 'file' attribute")
	}

	// Resolve relative path
	includePath, err := ResolveRelativePath(baseFilePath, fileAttr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve include path '%s': %w", fileAttr, err)
	}

	// Check if file exists
	if _, err := os.Stat(includePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("included file not found: %s", includePath)
	}

	// Create new context for included file
	childCtx := &parseContext{
		visitedFiles:       ctx.visitedFiles,
		symlinkResolutions: ctx.symlinkResolutions,
		currentDepth:       ctx.currentDepth + 1,
		includeChain:       append(append([]string{}, ctx.includeChain...), includePath),
		maxDepth:           ctx.maxDepth,
		followSymlinks:     ctx.followSymlinks,
	}

	// Parse the included file using format auto-detection
	return parseFileWithContext(includePath, childCtx)
}

// handleIncludeAll processes an includeAll directive
func (p *YAMLParser) handleIncludeAll(data any, baseFilePath string, ctx *parseContext) ([]ChangeSet, error) {
	includeAllMap, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("includeAll must be an object")
	}

	pathAttr, ok := includeAllMap["path"].(string)
	if !ok {
		return nil, fmt.Errorf("includeAll missing required 'path' attribute")
	}

	// Resolve relative path
	includePath, err := ResolveRelativePath(baseFilePath, pathAttr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve includeAll path '%s': %w", pathAttr, err)
	}

	// Check if directory exists
	info, err := os.Stat(includePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("includeAll path not found: %s", includePath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("includeAll path is not a directory: %s", includePath)
	}

	// Get resource filter if present
	resourceFilter := ""
	if filter, ok := includeAllMap["resourceFilter"].(string); ok {
		resourceFilter = filter
	}

	// Discover files
	files, err := DiscoverChangelogFiles(includePath, resourceFilter, ctx.followSymlinks)
	if err != nil {
		return nil, fmt.Errorf("failed to discover files in '%s': %w", includePath, err)
	}

	// Parse each discovered file
	var allChangeSets []ChangeSet
	for _, file := range files {
		// Create new context for included file
		childCtx := &parseContext{
			visitedFiles:       ctx.visitedFiles,
			symlinkResolutions: ctx.symlinkResolutions,
			currentDepth:       ctx.currentDepth + 1,
			includeChain:       append(append([]string{}, ctx.includeChain...), file),
			maxDepth:           ctx.maxDepth,
			followSymlinks:     ctx.followSymlinks,
		}

		changelog, err := parseFileWithContext(file, childCtx)
		if err != nil {
			return nil, fmt.Errorf("failed to parse included file '%s': %w", file, err)
		}

		allChangeSets = append(allChangeSets, changelog.ChangeSets...)
	}

	return allChangeSets, nil
}

// CanParse checks if this parser can handle the given file.
func (p *YAMLParser) CanParse(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".yaml" || ext == ".yml"
}

// parseFileWithContext parses a file with the given context using format auto-detection
func parseFileWithContext(filePath string, ctx *parseContext) (*Changelog, error) {
	format := DetectFormat(filePath)

	switch format {
	case FormatYAML:
		// Read and parse YAML
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		var doc yamlDatabaseChangeLog
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		parser := &YAMLParser{}
		return parser.parseWithContext(filePath, doc, ctx)

	case FormatXML:
		parser := &XMLParser{}
		return parser.Parse(filePath)

	case FormatSQL:
		parser := &SQLParser{}
		return parser.Parse(filePath)

	case FormatJSON:
		// Parse JSON
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		var doc yamlDatabaseChangeLog
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}

		parser := &JSONParser{}
		return parser.parseWithContext(filePath, doc, ctx)

	default:
		return nil, fmt.Errorf("unsupported file format: %s", filePath)
	}
}
