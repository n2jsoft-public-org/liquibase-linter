package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseContext_ShouldIgnore tests the ignore pattern matching functionality
func TestParseContext_ShouldIgnore(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name           string
		ignorePatterns []string
		basePath       string
		filePath       string
		shouldIgnore   bool
	}{
		{
			name:           "ignore init directory",
			ignorePatterns: []string{"changelog/init/**"},
			basePath:       tmpDir,
			filePath:       filepath.Join(tmpDir, "changelog/init/structure/tables.sql"),
			shouldIgnore:   true,
		},
		{
			name:           "ignore init directory with wildcard",
			ignorePatterns: []string{"changelog/init/**"},
			basePath:       tmpDir,
			filePath:       filepath.Join(tmpDir, "changelog/init/data/inserts.sql"),
			shouldIgnore:   true,
		},
		{
			name:           "don't ignore sprint directory",
			ignorePatterns: []string{"changelog/init/**"},
			basePath:       tmpDir,
			filePath:       filepath.Join(tmpDir, "changelog/sprints/v116/structure.sql"),
			shouldIgnore:   false,
		},
		{
			name:           "ignore specific file pattern",
			ignorePatterns: []string{"**/*.bak"},
			basePath:       tmpDir,
			filePath:       filepath.Join(tmpDir, "changelog/backup.bak"),
			shouldIgnore:   true,
		},
		{
			name:           "don't ignore when pattern doesn't match",
			ignorePatterns: []string{"**/*.bak"},
			basePath:       tmpDir,
			filePath:       filepath.Join(tmpDir, "changelog/main.sql"),
			shouldIgnore:   false,
		},
		{
			name:           "empty ignore patterns",
			ignorePatterns: []string{},
			basePath:       tmpDir,
			filePath:       filepath.Join(tmpDir, "changelog/init/tables.sql"),
			shouldIgnore:   false,
		},
		{
			name:           "multiple patterns - first matches",
			ignorePatterns: []string{"changelog/init/**", "changelog/test/**"},
			basePath:       tmpDir,
			filePath:       filepath.Join(tmpDir, "changelog/init/tables.sql"),
			shouldIgnore:   true,
		},
		{
			name:           "multiple patterns - second matches",
			ignorePatterns: []string{"changelog/init/**", "changelog/test/**"},
			basePath:       tmpDir,
			filePath:       filepath.Join(tmpDir, "changelog/test/fixtures.sql"),
			shouldIgnore:   true,
		},
		{
			name:           "multiple patterns - none match",
			ignorePatterns: []string{"changelog/init/**", "changelog/test/**"},
			basePath:       tmpDir,
			filePath:       filepath.Join(tmpDir, "changelog/sprints/v116/main.sql"),
			shouldIgnore:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newParseContext(10, true)
			ctx.SetIgnorePatterns(tt.ignorePatterns, tt.basePath)

			result := ctx.ShouldIgnore(tt.filePath)
			if result != tt.shouldIgnore {
				t.Errorf("ShouldIgnore() = %v, want %v for file %s with patterns %v",
					result, tt.shouldIgnore, tt.filePath, tt.ignorePatterns)
			}
		})
	}
}

// TestYAMLParser_IncludeAllWithIgnore tests that includeAll respects ignore patterns
func TestYAMLParser_IncludeAllWithIgnore(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create directory structure
	initDir := filepath.Join(tmpDir, "changelog", "init")
	sprintDir := filepath.Join(tmpDir, "changelog", "sprints", "v116")

	if err := os.MkdirAll(initDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sprintDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test changelog files
	initSQL := filepath.Join(initDir, "tables.sql")
	sprintSQL := filepath.Join(sprintDir, "changes.sql")

	initContent := `--liquibase formatted sql
--changeset test:1
CREATE TABLE test (id INT);
`
	sprintContent := `--liquibase formatted sql
--changeset test:2
CREATE TABLE sprint (id INT);
`

	if err := os.WriteFile(initSQL, []byte(initContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sprintSQL, []byte(sprintContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create master changelog that includes all
	masterPath := filepath.Join(tmpDir, "master.yaml")
	masterContent := `databaseChangeLog:
  - includeAll:
      path: changelog
`
	if err := os.WriteFile(masterPath, []byte(masterContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: Parse without ignore patterns
	parser := &YAMLParser{}
	changelog, err := parser.Parse(masterPath)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(changelog.ChangeSets) != 2 {
		t.Errorf("Expected 2 changesets without ignore patterns, got %d", len(changelog.ChangeSets))
	}

	// Test 2: Parse with ignore patterns
	ctx := newParseContext(10, true)
	ctx.SetIgnorePatterns([]string{"changelog/init/**"}, tmpDir)

	changelogWithIgnore, err := parser.ParseWithConfig(masterPath, []string{"changelog/init/**"}, tmpDir)
	if err != nil {
		t.Fatalf("ParseWithConfig() error = %v", err)
	}

	// Should only have the sprint changeset, not the init one
	if len(changelogWithIgnore.ChangeSets) != 1 {
		t.Errorf("Expected 1 changeset with ignore patterns, got %d", len(changelogWithIgnore.ChangeSets))
	}

	// Verify it's the sprint changeset
	if len(changelogWithIgnore.ChangeSets) > 0 {
		cs := changelogWithIgnore.ChangeSets[0]
		if cs.ID != "2" {
			t.Errorf("Expected changeset ID '2' (sprint), got '%s'", cs.ID)
		}
	}
}
