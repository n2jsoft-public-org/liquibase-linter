// Package parser provides functionality for parsing Liquibase changelog files.
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DiscoverChangelogFiles recursively discovers changelog files in a directory.
// It returns absolute file paths for files matching the supported extensions.
// If resourceFilter is provided, only files matching the pattern are returned.
// If followSymlinks is true, symlinks are followed and resolved paths are tracked to prevent loops.
func DiscoverChangelogFiles(dirPath, resourceFilter string, followSymlinks bool) ([]string, error) {
	var files []string
	visited := make(map[string]bool)

	absDir, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve directory path: %w", err)
	}

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Handle symlinks
		if followSymlinks && info.Mode()&os.ModeSymlink != 0 {
			realPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				// Skip broken symlinks
				return nil
			}

			// Check if we've already visited this resolved path
			normalizedReal := NormalizePath(realPath)
			if visited[normalizedReal] {
				return nil
			}
			visited[normalizedReal] = true

			// Get info of the resolved path
			info, err = os.Stat(realPath)
			if err != nil {
				return nil
			}
			path = realPath
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file has a changelog extension
		if !isChangelogFile(path) {
			return nil
		}

		// Apply resource filter if provided
		if resourceFilter != "" {
			matches, err := MatchesResourceFilter(path, resourceFilter)
			if err != nil {
				return fmt.Errorf("invalid resource filter pattern: %w", err)
			}
			if !matches {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to discover changelog files: %w", err)
	}

	return files, nil
}

// isChangelogFile checks if a file has a changelog extension.
func isChangelogFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".xml", ".sql", ".yaml", ".yml", ".json":
		return true
	default:
		return false
	}
}

// MatchesResourceFilter checks if a file path matches a resource filter pattern.
// Supports standard glob patterns plus ** for recursive directory matching.
// Both the file path and pattern are normalized to use forward slashes for consistent matching.
func MatchesResourceFilter(filePath, pattern string) (bool, error) {
	// Normalize both file path and pattern to use forward slashes
	// This ensures consistent matching across platforms
	filePath = filepath.ToSlash(filePath)
	pattern = filepath.ToSlash(pattern)

	// Handle ** for recursive directory matching
	if strings.Contains(pattern, "**") {
		return matchesRecursivePattern(filePath, pattern)
	}

	// Use standard filepath.Match for simple patterns
	matched, err := filepath.Match(pattern, filepath.Base(filePath))
	if err != nil {
		return false, err
	}

	// If base name matches, we're done
	if matched {
		return true, nil
	}

	// Try matching the full path for patterns with directory separators
	// Use forward slash since paths and patterns are normalized
	if strings.Contains(pattern, "/") {
		matched, err = filepath.Match(pattern, filePath)
		if err != nil {
			return false, err
		}
	}

	return matched, nil
}

// matchesRecursivePattern handles patterns with ** for recursive directory matching.
// Examples: **/*.sql matches any .sql file at any depth
//
//	v*/**/*.xml matches .xml files in subdirectories of directories starting with 'v'
//
// Note: Both filePath and pattern should already be normalized to forward slashes by the caller.
func matchesRecursivePattern(filePath, pattern string) (bool, error) {
	// Split pattern by **
	parts := strings.Split(pattern, "**")

	if len(parts) == 1 {
		// No ** in pattern, use standard matching
		return filepath.Match(pattern, filePath)
	}

	// Handle single ** (most common case: **/*.ext)
	if len(parts) == 2 {
		// Use forward slash since paths are normalized
		prefix := strings.TrimSuffix(parts[0], "/")
		suffix := strings.TrimPrefix(parts[1], "/")

		// Check if path starts with prefix (if any)
		if prefix != "" {
			if !strings.HasPrefix(filePath, prefix) {
				// Try matching just the relative part
				if !strings.Contains(filePath, prefix) {
					return false, nil
				}
			}
		}

		// Check if path ends with pattern after **
		if suffix != "" {
			matched, err := filepath.Match(suffix, filepath.Base(filePath))
			if err != nil {
				return false, err
			}
			if !matched {
				// Try matching with directory structure
				// Use forward slash for splitting since paths are normalized
				pathParts := strings.Split(filePath, "/")
				suffixParts := strings.Split(suffix, "/")

				if len(pathParts) >= len(suffixParts) {
					// Check if last N parts match suffix pattern
					for i := 0; i < len(suffixParts); i++ {
						pathIdx := len(pathParts) - len(suffixParts) + i
						matched, err := filepath.Match(suffixParts[i], pathParts[pathIdx])
						if err != nil {
							return false, err
						}
						if !matched {
							return false, nil
						}
					}
					return true, nil
				}
				return false, nil
			}
			return true, nil
		}

		return true, nil
	}

	// Multiple ** - more complex pattern
	// For now, return true if file matches basic criteria
	// This is a simplified implementation
	ext := filepath.Ext(filePath)
	return strings.HasSuffix(pattern, ext), nil
}

// ResolveRelativePath resolves a relative path against a base path.
// Returns an absolute path.
func ResolveRelativePath(basePath, relativePath string) (string, error) {
	// If relativePath is already absolute, return it
	if filepath.IsAbs(relativePath) {
		return filepath.Clean(relativePath), nil
	}

	// Get the directory of the base path
	baseDir := filepath.Dir(basePath)
	if !filepath.IsAbs(baseDir) {
		abs, err := filepath.Abs(baseDir)
		if err != nil {
			return "", fmt.Errorf("failed to resolve base path: %w", err)
		}
		baseDir = abs
	}

	// Join and clean the path
	resolved := filepath.Join(baseDir, relativePath)
	return filepath.Clean(resolved), nil
}

// NormalizePath normalizes a file path for consistent comparison across platforms.
// Converts to absolute path, cleans it, and evaluates symlinks if present.
func NormalizePath(path string) string {
	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		// If we can't get absolute, just clean it
		return filepath.Clean(path)
	}

	// Clean the path
	absPath = filepath.Clean(absPath)

	// Try to evaluate symlinks for more accurate comparison
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If symlink evaluation fails, return the absolute path
		return absPath
	}

	return filepath.Clean(realPath)
}

// GetFileDir returns the directory containing the given file path.
func GetFileDir(filePath string) string {
	return filepath.Dir(filePath)
}

// JoinPath joins path elements and cleans the result.
func JoinPath(elem ...string) string {
	return filepath.Join(elem...)
}

// GetRelativePath returns the relative path from base to target.
// Both paths should be absolute or both relative.
func GetRelativePath(base, target string) (string, error) {
	// Ensure both paths are absolute for consistent behavior
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base path: %w", err)
	}

	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Calculate relative path
	relPath, err := filepath.Rel(absBase, absTarget)
	if err != nil {
		return "", fmt.Errorf("failed to calculate relative path: %w", err)
	}

	return relPath, nil
}
