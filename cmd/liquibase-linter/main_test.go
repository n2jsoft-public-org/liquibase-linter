package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/config"
	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

func TestDiscoverConfigFile(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create nested directories
	targetDir := filepath.Join(tmpDir, "db", "changelog")
	//nolint:gosec // G301: Test directory, permissions are acceptable
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Create a config file in the changelog directory
	configPath := filepath.Join(targetDir, ".liquibase-linter.yaml")
	//nolint:gosec // G306: Test configuration file, permissions are acceptable
	err = os.WriteFile(configPath, []byte("ignore:\n  - \"test/**\"\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test discovery
	discovered := discoverConfigFile(targetDir)
	if discovered == "" {
		t.Fatal("Config file should have been discovered")
	}

	if discovered != configPath {
		t.Errorf("Expected config path %s, got %s", configPath, discovered)
	}
}

func TestShouldIgnore(t *testing.T) {
	tests := []struct {
		name            string
		file            string
		basePath        string
		ignorePatterns  []string
		expectedIgnored bool
	}{
		{
			name:            "ignore init folder",
			file:            "/path/to/changelog/init/001-schema.sql",
			basePath:        "/path/to/changelog",
			ignorePatterns:  []string{"init/**"},
			expectedIgnored: true,
		},
		{
			name:            "don't ignore sprints folder",
			file:            "/path/to/changelog/sprints/v1/001-feature.sql",
			basePath:        "/path/to/changelog",
			ignorePatterns:  []string{"init/**"},
			expectedIgnored: false,
		},
		{
			name:            "ignore test files",
			file:            "/path/to/changelog/test/fixture.sql",
			basePath:        "/path/to/changelog",
			ignorePatterns:  []string{"test/**"},
			expectedIgnored: true,
		},
		{
			name:            "multiple patterns",
			file:            "/path/to/changelog/init/001-schema.sql",
			basePath:        "/path/to/changelog",
			ignorePatterns:  []string{"test/**", "init/**", "fixtures/**"},
			expectedIgnored: true,
		},
		{
			name:            "no patterns",
			file:            "/path/to/changelog/init/001-schema.sql",
			basePath:        "/path/to/changelog",
			ignorePatterns:  []string{},
			expectedIgnored: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ignored := shouldIgnore(tt.file, tt.basePath, tt.ignorePatterns)
			if ignored != tt.expectedIgnored {
				t.Errorf("shouldIgnore() = %v, want %v", ignored, tt.expectedIgnored)
			}
		})
	}
}

func TestDiscoverFiles_WithIgnorePattern(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create directories
	initDir := filepath.Join(tmpDir, "init")
	sprintsDir := filepath.Join(tmpDir, "sprints")
	//nolint:gosec // G301: Test directory, permissions are acceptable
	err := os.MkdirAll(initDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create init dir: %v", err)
	}
	//nolint:gosec // G301: Test directory, permissions are acceptable
	err = os.MkdirAll(sprintsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sprints dir: %v", err)
	}

	// Create test files
	initFile := filepath.Join(initDir, "001-schema.sql")
	sprintsFile := filepath.Join(sprintsDir, "001-feature.sql")

	initContent := `--liquibase formatted sql
--changeset test:1
CREATE TABLE test1 (id INT);
`
	sprintsContent := `--liquibase formatted sql
--changeset test:2
CREATE TABLE test2 (id INT);
`

	//nolint:gosec // G306: Test file, permissions are acceptable
	err = os.WriteFile(initFile, []byte(initContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create init file: %v", err)
	}
	//nolint:gosec // G306: Test file, permissions are acceptable
	err = os.WriteFile(sprintsFile, []byte(sprintsContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create sprints file: %v", err)
	}

	// Create config with ignore pattern
	cfg := config.Default()
	cfg.Ignore = []string{"init/**"}

	// Discover files
	files, err := discoverFiles(tmpDir, cfg)
	if err != nil {
		t.Fatalf("Failed to discover files: %v", err)
	}

	// Verify that init file is not included
	absSprintsFile, _ := parser.ResolveRelativePath(".", sprintsFile)
	found := false
	for _, f := range files {
		if f == absSprintsFile {
			found = true
		}
		if filepath.Base(filepath.Dir(f)) == "init" {
			t.Errorf("Init file should have been ignored: %s", f)
		}
	}

	if !found {
		t.Error("Sprints file should have been discovered")
	}
}
