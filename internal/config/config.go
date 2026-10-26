// Package config provides functionality for loading and managing
// Liquibase Linter configuration settings.
package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration for the linter.
type Config struct {
	Rules                map[string]RuleConfig      `yaml:"rules"`
	SeverityThreshold    string                     `yaml:"severity_threshold"`
	Ignore               []string                   `yaml:"ignore"`
	Output               OutputConfig               `yaml:"output"`
	FileStructure        FileStructureConfig        `yaml:"file_structure"`
	LabelPattern         LabelPatternConfig         `yaml:"label_pattern"`
	NoManualTransactions NoManualTransactionsConfig `yaml:"no_manual_transactions"`
	DBMSScope            DBMSScopeConfig            `yaml:"dbms_scope"`
	DBMSTwin             DBMSTwinConfig             `yaml:"dbms_twin"`
	AtomicChangeset      AtomicChangesetConfig      `yaml:"atomic_changeset"`
	Parser               ParserConfig               `yaml:"parser"`
}

// RuleConfig represents configuration for a single rule.
type RuleConfig struct {
	Severity        string   `yaml:"severity"`
	Mode            string   `yaml:"mode"`
	ExcludePatterns []string `yaml:"exclude_patterns"`
	Enabled         bool     `yaml:"enabled"`
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
	SprintPattern    string   `yaml:"sprint_pattern"`
	StructurePattern string   `yaml:"structure_pattern"`
	DataPattern      string   `yaml:"data_pattern"`
	SprintBasePath   string   `yaml:"sprint_base_path"`
	ExcludePatterns  []string `yaml:"exclude_patterns"`
	Enabled          bool     `yaml:"enabled"`
}

// LabelPatternConfig represents configuration for label pattern rule.
type LabelPatternConfig struct {
	Severity        string   `yaml:"severity"`
	Pattern         string   `yaml:"pattern"`
	Patterns        []string `yaml:"patterns"`
	ExcludePatterns []string `yaml:"exclude_patterns"`
	Enabled         bool     `yaml:"enabled"`
	RequireLabel    bool     `yaml:"require_label"`
}

// NoManualTransactionsConfig represents configuration for no-manual-transactions rule.
type NoManualTransactionsConfig struct {
	// Patterns are regex patterns to match transaction control keywords.
	Patterns []string `yaml:"patterns"`
	// ExcludeChangeTypes are change types to skip (e.g., createProcedure).
	ExcludeChangeTypes []string `yaml:"exclude_change_types"`
	// ExcludePatterns are glob patterns to exclude from checking.
	ExcludePatterns []string `yaml:"exclude_patterns"`
	// Enabled determines if the rule is active.
	Enabled bool `yaml:"enabled"`
	// CaseInsensitive determines if pattern matching is case-insensitive.
	CaseInsensitive bool `yaml:"case_insensitive"`
}

// AtomicChangesetConfig represents configuration for atomic-changeset rule.
type AtomicChangesetConfig struct {
	// ExcludePatterns are glob patterns to exclude from checking.
	ExcludePatterns []string `yaml:"exclude_patterns"`
	// MaxSQLStatements is the maximum number of SQL statements allowed in a single change.
	MaxSQLStatements int `yaml:"max_sql_statements"`
	// Enabled determines if the rule is active.
	Enabled bool `yaml:"enabled"`
	// AllowTableWithIndexes allows table creation with indexes as one operation.
	AllowTableWithIndexes bool `yaml:"allow_table_with_indexes"`
	// AllowTableWithConstraints allows table with inline constraints.
	AllowTableWithConstraints bool `yaml:"allow_table_with_constraints"`
}

// DBMSScopeConfig represents configuration for dbms-scope rule.
type DBMSScopeConfig struct {
	// ForbiddenPreconditionPatterns maps an engine name to regex patterns that the precondition
	// SQL of a changeset running on that engine must not match.
	ForbiddenPreconditionPatterns map[string][]string `yaml:"forbidden_precondition_patterns"`
	// AllowedValues are the engine names accepted in dbms, in addition to "all" and "none".
	AllowedValues []string `yaml:"allowed_values"`
	// ExcludePatterns are glob patterns to exclude from checking.
	ExcludePatterns []string `yaml:"exclude_patterns"`
	// Enabled determines if the rule is active.
	Enabled bool `yaml:"enabled"`
	// RequireDBMS reports changesets without a dbms attribute.
	RequireDBMS bool `yaml:"require_dbms"`
}

// DBMSTwinConfig represents configuration for dbms-twin rule.
type DBMSTwinConfig struct {
	// SourceDBMS is the engine the changelog is migrated from.
	SourceDBMS string `yaml:"source_dbms"`
	// TargetDBMS is the engine the changelog is migrated to.
	TargetDBMS string `yaml:"target_dbms"`
	// IDSuffix is appended to a source changeset ID to name its twin.
	IDSuffix string `yaml:"id_suffix"`
	// SincePathPattern is a regex whose first capture group is a number; files that do not
	// match it are not checked.
	SincePathPattern string `yaml:"since_path_pattern"`
	// ExcludePatterns are glob patterns to exclude from checking.
	ExcludePatterns []string `yaml:"exclude_patterns"`
	// Since is the lowest number accepted from SincePathPattern.
	Since int `yaml:"since"`
	// Enabled determines if the rule is active.
	Enabled bool `yaml:"enabled"`
}

// LiquibaseDBMSNames are the database short names Liquibase accepts in a dbms attribute.
var LiquibaseDBMSNames = []string{
	"aerospike", "asany", "bigquery", "cassandra", "cloudspanner", "cockroachdb", "cosmosdb",
	"databricks", "db2", "db2i", "db2z", "derby", "dynamodb", "edb", "firebird", "h2", "hana",
	"hsqldb", "informix", "ingres", "mariadb", "maxdb", "mongodb", "mssql", "mysql", "neo4j",
	"oracle", "postgresql", "redshift", "snowflake", "sqlite", "sybase", "teradata", "vertica",
	"yugabytedb",
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
			"redundant-onerror-halt": {
				Enabled:  true,
				Severity: "info",
			},
			"no-if-exists": {
				Enabled:  true,
				Severity: "warning",
			},
			"atomic-changeset": {
				Enabled:  true,
				Severity: "info",
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
		AtomicChangeset: AtomicChangesetConfig{
			Enabled:                   true,
			AllowTableWithIndexes:     false,
			AllowTableWithConstraints: true,
			MaxSQLStatements:          1,
			ExcludePatterns:           []string{"**/init/**"},
		},
		DBMSScope: DBMSScopeConfig{
			Enabled:       false, // Disabled by default (opt-in, for changelogs shared by several engines)
			RequireDBMS:   true,
			AllowedValues: append([]string{}, LiquibaseDBMSNames...),
		},
		DBMSTwin: DBMSTwinConfig{
			Enabled:  false, // Disabled by default (opt-in, for migrations between two engines)
			IDSuffix: "-pg",
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
	//nolint:gosec // G304: File path is provided by user for configuration
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

	// Validate dbms scope patterns, so the rule never has to deal with an invalid regex
	for engine, patterns := range c.DBMSScope.ForbiddenPreconditionPatterns {
		for _, pattern := range patterns {
			if _, err := regexp.Compile(pattern); err != nil {
				return fmt.Errorf("invalid dbms_scope.forbidden_precondition_patterns.%s pattern %q: %w", engine, pattern, err)
			}
		}
	}

	return c.DBMSTwin.validate()
}

// validate checks the dbms-twin configuration when the rule is enabled.
func (t *DBMSTwinConfig) validate() error {
	if !t.Enabled {
		return nil
	}
	if t.SourceDBMS == "" || t.TargetDBMS == "" || t.IDSuffix == "" {
		return errors.New("dbms_twin.source_dbms, dbms_twin.target_dbms and dbms_twin.id_suffix are required when dbms_twin is enabled")
	}
	if strings.EqualFold(t.SourceDBMS, t.TargetDBMS) {
		return fmt.Errorf("dbms_twin.source_dbms and dbms_twin.target_dbms must differ, both are %q", t.SourceDBMS)
	}
	if t.SincePathPattern != "" {
		re, err := regexp.Compile(t.SincePathPattern)
		if err != nil {
			return fmt.Errorf("invalid dbms_twin.since_path_pattern %q: %w", t.SincePathPattern, err)
		}
		if re.NumSubexp() == 0 {
			return fmt.Errorf("dbms_twin.since_path_pattern %q needs a capture group for the number compared to since", t.SincePathPattern)
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
# See documentation at: https://github.com/n2jsoft-public-org/liquibase-linter

`
	data = append([]byte(header), data...)

	// Write to file
	//nolint:gosec // G306: Config file permissions are intentionally 0644 for readability
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	return nil
}
