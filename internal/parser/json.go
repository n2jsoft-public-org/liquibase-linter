// Package parser provides functionality for parsing Liquibase changelog files.
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// JSONParser parses JSON-formatted Liquibase changelogs.
type JSONParser struct{}

// Parse parses a JSON changelog file.
func (p *JSONParser) Parse(filePath string) (*Changelog, error) {
	return p.ParseWithConfig(filePath, []string{}, "")
}

// ParseWithConfig parses a JSON changelog file with ignore patterns for filtering includes.
func (p *JSONParser) ParseWithConfig(filePath string, ignorePatterns []string, basePath string) (*Changelog, error) {
	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON - reuse YAML struct since it has dual tags
	var doc yamlDatabaseChangeLog
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Create parse context with default config values
	ctx := newParseContext(10, true)
	absPath, _ := filepath.Abs(filePath)
	ctx.includeChain = []string{absPath}

	// Set ignore patterns if provided
	if len(ignorePatterns) > 0 && basePath != "" {
		ctx.SetIgnorePatterns(ignorePatterns, basePath)
	}

	// Parse the changelog with context - reuse YAML parser logic
	return p.parseWithContext(filePath, doc, ctx)
}

// parseWithContext parses a JSON changelog with include tracking
// This reuses the same logic as YAML parser since JSON structure is identical
func (p *JSONParser) parseWithContext(filePath string, doc yamlDatabaseChangeLog, ctx *parseContext) (*Changelog, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve file path: %w", err)
	}

	// Check if already visited (circular include detection)
	normalizedPath := NormalizePath(absPath)
	if ctx.visitedFiles[normalizedPath] {
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
	// Track this file as processed
	*ctx.processedFiles = append(*ctx.processedFiles, absPath)

	// Track symlink resolution
	if ctx.followSymlinks {
		realPath, err := filepath.EvalSymlinks(absPath)
		if err == nil && realPath != absPath {
			ctx.symlinkResolutions[absPath] = realPath
		}
	}

	changelog := &Changelog{
		FilePath:      absPath,
		Format:        FormatJSON,
		ChangeSets:    []ChangeSet{},
		IncludedFiles: []string{}, // Will be populated at the end
	}

	// Use YAML parser instance for shared parsing logic
	yamlParser := &YAMLParser{}

	// Process each element in the databaseChangeLog array
	for _, element := range doc.DatabaseChangeLog {
		for key, value := range element {
			switch key {
			case "changeSet":
				cs, err := yamlParser.parseChangeSet(value, absPath)
				if err != nil {
					return nil, err
				}
				changelog.ChangeSets = append(changelog.ChangeSets, cs)

			case "include":
				included, err := p.handleInclude(value, absPath, ctx)
				if err != nil {
					return nil, fmt.Errorf("in file %s: %w", absPath, err)
				}
				changelog.ChangeSets = append(changelog.ChangeSets, included.ChangeSets...)

			case "includeAll":
				// TODO Phase 6+: includeAll context and labels attributes not yet supported
				included, err := p.handleIncludeAll(value, absPath, ctx)
				if err != nil {
					return nil, fmt.Errorf("in file %s: %w", absPath, err)
				}
				changelog.ChangeSets = append(changelog.ChangeSets, included...)

			case "property":
				continue

			default:
				continue
			}
		}
	}

	// Populate list of all included files
	changelog.IncludedFiles = *ctx.processedFiles

	return changelog, nil
}

// handleInclude processes an include directive
func (p *JSONParser) handleInclude(data any, baseFilePath string, ctx *parseContext) (*Changelog, error) {
	includeMap, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("include must be an object")
	}

	fileAttr, ok := includeMap["file"].(string)
	if !ok {
		return nil, fmt.Errorf("include missing required 'file' attribute")
	}

	includePath, err := ResolveRelativePath(baseFilePath, fileAttr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve include path '%s': %w", fileAttr, err)
	}

	if _, err := os.Stat(includePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("included file not found: %s", includePath)
	}

	childCtx := &parseContext{
		visitedFiles:       ctx.visitedFiles,
		symlinkResolutions: ctx.symlinkResolutions,
		processedFiles:     ctx.processedFiles, // Share processed files list
		currentDepth:       ctx.currentDepth + 1,
		includeChain:       append(append([]string{}, ctx.includeChain...), includePath),
		maxDepth:           ctx.maxDepth,
		followSymlinks:     ctx.followSymlinks,
		ignorePatterns:     ctx.ignorePatterns,
		basePath:           ctx.basePath,
	}

	return parseFileWithContext(includePath, childCtx)
}

// handleIncludeAll processes an includeAll directive
func (p *JSONParser) handleIncludeAll(data any, baseFilePath string, ctx *parseContext) ([]ChangeSet, error) {
	includeAllMap, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("includeAll must be an object")
	}

	pathAttr, ok := includeAllMap["path"].(string)
	if !ok {
		return nil, fmt.Errorf("includeAll missing required 'path' attribute")
	}

	includePath, err := ResolveRelativePath(baseFilePath, pathAttr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve includeAll path '%s': %w", pathAttr, err)
	}

	info, err := os.Stat(includePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("includeAll path not found: %s", includePath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("includeAll path is not a directory: %s", includePath)
	}

	resourceFilter := ""
	if filter, ok := includeAllMap["resourceFilter"].(string); ok {
		resourceFilter = filter
	}

	files, err := DiscoverChangelogFiles(includePath, resourceFilter, ctx.followSymlinks)
	if err != nil {
		return nil, fmt.Errorf("failed to discover files in '%s': %w", includePath, err)
	}

	var allChangeSets []ChangeSet
	for _, file := range files {
		// Check if file should be ignored
		if ctx.ShouldIgnore(file) {
			continue
		}

		childCtx := &parseContext{
			visitedFiles:       ctx.visitedFiles,
			symlinkResolutions: ctx.symlinkResolutions,
			processedFiles:     ctx.processedFiles, // Share processed files list
			currentDepth:       ctx.currentDepth + 1,
			includeChain:       append(append([]string{}, ctx.includeChain...), file),
			maxDepth:           ctx.maxDepth,
			followSymlinks:     ctx.followSymlinks,
			ignorePatterns:     ctx.ignorePatterns,
			basePath:           ctx.basePath,
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
func (p *JSONParser) CanParse(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".json"
}
