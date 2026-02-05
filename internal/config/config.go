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
	Rules             map[string]RuleConfig `yaml:"rules"`
	Ignore            []string              `yaml:"ignore"`
	Output            OutputConfig          `yaml:"output"`
	SeverityThreshold string                `yaml:"severity_threshold"`
}

// RuleConfig represents configuration for a single rule.
type RuleConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Severity string `yaml:"severity"`
}

// OutputConfig represents output formatting configuration.
type OutputConfig struct {
	Format   string `yaml:"format"`
	Colorize bool   `yaml:"colorize"`
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
				Enabled:  true,
				Severity: "warning",
			},
		},
		Ignore: []string{},
		Output: OutputConfig{
			Format:   "text",
			Colorize: true,
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
