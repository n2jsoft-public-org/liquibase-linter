package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg == nil {
		t.Fatal("Default() returned nil")
	}

	if cfg.Output.Format != "text" {
		t.Errorf("Expected default format to be 'text', got '%s'", cfg.Output.Format)
	}

	if !cfg.Output.Colorize {
		t.Error("Expected default colorize to be true")
	}

	if cfg.SeverityThreshold != "warning" {
		t.Errorf("Expected default severity threshold to be 'warning', got '%s'", cfg.SeverityThreshold)
	}

	// Check that some default rules exist
	if len(cfg.Rules) == 0 {
		t.Error("Expected default rules to be non-empty")
	}

	if rule, ok := cfg.Rules["sql-injection"]; ok {
		if !rule.Enabled {
			t.Error("Expected sql-injection rule to be enabled by default")
		}
		if rule.Severity != "critical" {
			t.Errorf("Expected sql-injection severity to be 'critical', got '%s'", rule.Severity)
		}
	} else {
		t.Error("Expected sql-injection rule to exist in defaults")
	}
}

func TestLoad_EmptyPath(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load with empty path should return default config, got error: %v", err)
	}

	if cfg == nil {
		t.Fatal("Load returned nil config")
	}

	// Should return default config
	defaultCfg := Default()
	if cfg.Output.Format != defaultCfg.Output.Format {
		t.Error("Load with empty path should return default config")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("Expected error for nonexistent file, got nil")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `rules:
  sql-injection:
    enabled: true
    severity: critical
  missing-rollback:
    enabled: false
    severity: warning

ignore:
  - "test/*.xml"
  - "fixtures/*.xml"

output:
  format: json
  colorize: false

severity_threshold: critical
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Output.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", cfg.Output.Format)
	}

	if cfg.Output.Colorize {
		t.Error("Expected colorize to be false")
	}

	if cfg.SeverityThreshold != "critical" {
		t.Errorf("Expected severity threshold 'critical', got '%s'", cfg.SeverityThreshold)
	}

	if len(cfg.Ignore) != 2 {
		t.Errorf("Expected 2 ignore patterns, got %d", len(cfg.Ignore))
	}

	if rule, ok := cfg.Rules["sql-injection"]; ok {
		if !rule.Enabled {
			t.Error("Expected sql-injection to be enabled")
		}
	} else {
		t.Error("Expected sql-injection rule to exist")
	}

	if rule, ok := cfg.Rules["missing-rollback"]; ok {
		if rule.Enabled {
			t.Error("Expected missing-rollback to be disabled")
		}
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "bad-config.yaml")

	invalidContent := `rules:
  this is not: valid: yaml::
`

	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Fatal("Expected error for invalid YAML, got nil")
	}
}

func TestValidate_InvalidFormat(t *testing.T) {
	cfg := Default()
	cfg.Output.Format = "invalid-format"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Expected validation error for invalid format, got nil")
	}
}

func TestValidate_InvalidSeverityThreshold(t *testing.T) {
	cfg := Default()
	cfg.SeverityThreshold = "invalid-severity"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Expected validation error for invalid severity threshold, got nil")
	}
}

func TestValidate_InvalidRuleSeverity(t *testing.T) {
	cfg := Default()
	cfg.Rules["test-rule"] = RuleConfig{
		Enabled:  true,
		Severity: "invalid-severity",
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Expected validation error for invalid rule severity, got nil")
	}
}

func TestValidate_ValidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "default config",
			config: Default(),
		},
		{
			name: "json format",
			config: &Config{
				Rules:  map[string]RuleConfig{},
				Ignore: []string{},
				Output: OutputConfig{
					Format:   "json",
					Colorize: false,
				},
				SeverityThreshold: "info",
			},
		},
		{
			name: "sarif format",
			config: &Config{
				Rules:  map[string]RuleConfig{},
				Ignore: []string{},
				Output: OutputConfig{
					Format:   "sarif",
					Colorize: false,
				},
				SeverityThreshold: "critical",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if err != nil {
				t.Errorf("Validate() failed for %s: %v", tt.name, err)
			}
		})
	}
}

func TestInitConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".liquibase-linter.yaml")

	err := InitConfig(configPath)
	if err != nil {
		t.Fatalf("InitConfig failed: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Configuration file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read created config file: %v", err)
	}

	content := string(data)
	if len(content) == 0 {
		t.Error("Created config file is empty")
	}

	// Verify it can be loaded
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load created config: %v", err)
	}

	if cfg == nil {
		t.Fatal("Loaded config is nil")
	}
}

func TestInitConfig_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".liquibase-linter.yaml")

	// Create file first
	if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to init again
	err := InitConfig(configPath)
	if err == nil {
		t.Fatal("Expected error when file already exists, got nil")
	}
}
