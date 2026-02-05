// Package config provides functionality for loading and managing
// Liquibase Linter configuration settings.
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for the linter.
type Config struct {
	Rules                map[string]RuleConfig       `yaml:"rules"`
	Ignore               []string                    `yaml:"ignore"`
	Output               OutputConfig                `yaml:"output"`
	Parser               ParserConfig                `yaml:"parser"`
	FileStructure        FileStructureConfig         `yaml:"file_structure"`
	LabelPattern         LabelPatternConfig          `yaml:"label_pattern"`
	NoManualTransactions NoManualTransactionsConfig  `yaml:"no_manual_transactions"`
	SeverityThreshold    string                      `yaml:"severity_threshold"`
}

// RuleConfig represents configuration for a single rule.
type RuleConfig struct {
	Enabled         bool     `yaml:"enabled"`
	Severity        string   `yaml:"severity"`
	Mode            string   `yaml:"mode"`
	ExcludePatterns []string `yaml:"exclude_patterns"`
}

// Mode constants for non-idempotent rule
const (
	ModeRiskyOnly = "risky-only"
	ModeAll       = "all"
)

// OutputConfig represents output formatting configuration.
type OutputConfig struct {
	Format   string `yaml:"format"`
	Colorize bool   `yaml:"colorize"`
}

// ParserConfig represents parser behavior configuration.
type ParserConfig struct {
	// MaxIncludeDepth is the maximum depth for nested include/includeAll directives.
	// Must be between 1 and 100. Default is 10.
	MaxIncludeDepth int `yaml:"max_include_depth"`
	// FollowSymlinks determines whether symlinks should be followed during file discovery.
	// Default is true. Symlink loops are automatically detected and prevented.
	FollowSymlinks bool `yaml:"follow_symlinks"`
}

// FileStructureConfig represents file organization rules configuration.
type FileStructureConfig struct {
	// Enabled determines if file structure rules should be enforced.
	Enabled bool `yaml:"enabled"`
	// SprintPattern is a regex pattern for matching sprint folders (e.g., "^v\\d+$").
	// Default: "^v\\d+$" (matches v116, v117, etc.)
	SprintPattern string `yaml:"sprint_pattern"`
	// StructurePattern is a regex pattern for matching structure folders.
	// Default: "(?i)^\\d+\\s*-\\s*structure$" (matches "0 - structure", "0-structure", etc.)
	StructurePattern string `yaml:"structure_pattern"`
	// DataPattern is a regex pattern for matching data folders.
	// Default: "(?i)^\\d+\\s*-\\s*data$" (matches "1 - data", "1-data", etc.)
	DataPattern string `yaml:"data_pattern"`
	// ExcludePatterns are glob patterns to exclude from file structure rules.
	// Default: ["**/init/**"] (excludes initialization scripts)
	ExcludePatterns []string `yaml:"exclude_patterns"`
	// SprintBasePath is the base path where sprint folders are located.
	// If empty, assumes sprints can be anywhere. Example: "changelog/sprints"
	SprintBasePath string `yaml:"sprint_base_path"`
}

// LabelPatternConfig represents configuration for label pattern rule.
type LabelPatternConfig struct {
	// Enabled determines if label pattern validation is active.
	Enabled bool `yaml:"enabled"`
	// Severity is the severity level for violations (info, warning, critical).
	Severity string `yaml:"severity"`
	// Pattern is a single regex pattern for valid labels (e.g., "^v\\d+$").
	Pattern string `yaml:"pattern"`
	// Patterns is a list of regex patterns (alternative to single Pattern).
	// Labels matching any pattern are considered valid.
	Patterns []string `yaml:"patterns"`
	// RequireLabel determines if changesets must have at least one label.
	// Default: true
	RequireLabel bool `yaml:"require_label"`
	// ExcludePatterns are glob patterns to exclude from label validation.
	// Default: ["**/init/**"]
	ExcludePatterns []string `yaml:"exclude_patterns"`
}

// NoManualTransactionsConfig represents configuration for no-manual-transactions rule.
type NoManualTransactionsConfig struct {
	// Enabled determines if the rule is active.
	Enabled bool `yaml:"enabled"`
	// Patterns are regex patterns to match transaction control keywords.
	Patterns []string `yaml:"patterns"`
	// CaseInsensitive determines if pattern matching is case-insensitive.
	CaseInsensitive bool `yaml:"case_insensitive"`
	// ExcludeChangeTypes are change types to skip (e.g., createProcedure).
	ExcludeChangeTypes []string `yaml:"exclude_change_types"`
	// ExcludePatterns are glob patterns to exclude from checking.
	ExcludePatterns []string `yaml:"exclude_patterns"`
}

// Default returns a Config with default values.
func Default() *Config {
	return &Config{
		Rules: map[string]RuleConfig{
			"sql-injection": {
				Enabled:  true,
				Severity: "critical",
			},
			"hardcoded-credentials": {
				Enabled:  true,
				Severity: "critical",
			},
			"dangerous-operations": {
				Enabled:  true,
				Severity: "critical",
			},
			"missing-rollback": {
				Enabled:  true,
				Severity: "warning",
			},
			"non-idempotent": {
				Enabled:         true,
				Severity:        "warning",
				Mode:            ModeRiskyOnly,
				ExcludePatterns: []string{"**/init/**", "**/seed/**"},
			},
			"no-manual-transactions": {
				Enabled:  true,
				Severity: "warning",
			},
		},
		Ignore: []string{},
		Output: OutputConfig{
			Format:   "text",
			Colorize: true,
		},
		Parser: ParserConfig{
			MaxIncludeDepth: 10,
			FollowSymlinks:  true,
		},
		FileStructure: FileStructureConfig{
			Enabled:          true, // Enabled by default to enforce file structure organization
			SprintPattern:    `^v\d+$`,
			StructurePattern: `(?i)^\d+\s*-\s*structure$`,
			DataPattern:      `(?i)^\d+\s*-\s*data$`,
			ExcludePatterns:  []string{"**/init/**"},
			SprintBasePath:   "",
		},
		LabelPattern: LabelPatternConfig{
			Enabled:         false, // Disabled by default (opt-in)
			Severity:        "warning",
			Pattern:         `^v\d+$`,
			RequireLabel:    true,
			ExcludePatterns: []string{"**/init/**"},
		},
		NoManualTransactions: NoManualTransactionsConfig{
			Enabled: true,
			Patterns: []string{
				`\bBEGIN\s+(TRANSACTION|TRAN|WORK)?\b`,
				`\bSTART\s+TRANSACTION\b`,
				`\bCOMMIT\s+(TRANSACTION|TRAN|WORK)?\b`,
				`\bROLLBACK(\s+(TRANSACTION|TRAN|WORK))?\b`,
				`\bSAVE(POINT)?\s+TRANSACTION\b`,
			},
			CaseInsensitive: true,
			ExcludeChangeTypes: []string{
				"createProcedure",
				"createFunction",
				"createTrigger",
			},
		},
		SeverityThreshold: "warning",
	}
}

// Load reads a configuration file and returns a Config.
// If configPath is empty, returns the default configuration.
func Load(configPath string) (*Config, error) {
	// If no config path provided, return default
	if configPath == "" {
		return Default(), nil
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("configuration file not found: %s", configPath)
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse YAML
	cfg := Default() // Start with defaults
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration file: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	// Validate output format
	validFormats := map[string]bool{
		"text":  true,
		"json":  true,
		"sarif": true,
		"junit": true,
	}
	if !validFormats[c.Output.Format] {
		return fmt.Errorf("invalid output format: %s (must be text, json, sarif, or junit)", c.Output.Format)
	}

	// Validate severity threshold
	validSeverities := map[string]bool{
		"info":     true,
		"warning":  true,
		"critical": true,
	}
	if !validSeverities[c.SeverityThreshold] {
		return fmt.Errorf("invalid severity threshold: %s (must be info, warning, or critical)", c.SeverityThreshold)
	}

	// Validate rule severities
	for ruleName, ruleConfig := range c.Rules {
		if ruleConfig.Severity != "" && !validSeverities[ruleConfig.Severity] {
			return fmt.Errorf("invalid severity for rule %s: %s", ruleName, ruleConfig.Severity)
		}
	}

	// Validate parser configuration
	if c.Parser.MaxIncludeDepth < 1 || c.Parser.MaxIncludeDepth > 100 {
		return fmt.Errorf("max_include_depth must be between 1 and 100, got %d", c.Parser.MaxIncludeDepth)
	}

	// Validate file structure configuration
	if c.FileStructure.Enabled {
		if c.FileStructure.SprintPattern == "" {
			return errors.New("file_structure.sprint_pattern cannot be empty when file structure rules are enabled")
		}
		if c.FileStructure.StructurePattern == "" {
			return errors.New("file_structure.structure_pattern cannot be empty when file structure rules are enabled")
		}
		if c.FileStructure.DataPattern == "" {
			return errors.New("file_structure.data_pattern cannot be empty when file structure rules are enabled")
		}
	}

	return nil
}

// InitConfig creates a new configuration file with default settings at the specified path.
func InitConfig(path string) error {
	// Check if file already exists
	if _, err := os.Stat(path); err == nil {
		return errors.New("configuration file already exists")
	}

	cfg := Default()

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to generate configuration: %w", err)
	}

	// Add header comment
	header := `# Liquibase Linter Configuration
# See documentation at: https://github.com/n2jsoft/liquibase-linter

`
	data = append([]byte(header), data...)

	// Write to file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}
