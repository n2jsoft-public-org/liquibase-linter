// Package main provides the command-line interface for the Liquibase Linter tool.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/config"
	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
	"github.com/n2jsoft-public-org/liquibase-linter/internal/reporter"
	"github.com/n2jsoft-public-org/liquibase-linter/internal/rules"
)

// Build information, can be overridden at build time with -ldflags
var (
	version    = "dev"
	buildDate  = ""
	commitHash = ""
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "check":
		checkCmd()
	case "rules":
		rulesCmd()
	case "init":
		initCmd()
	case "version":
		versionCmd()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Println(`liquibase-linter - A linter for Liquibase changelog files

Usage:
  liquibase-linter <command> [flags]

Commands:
  check      Check Liquibase changelog files for issues
  rules      List available linting rules
  init       Initialize a configuration file
  version    Print version information
  help       Show this help message

Examples:
  # Check a single changelog file
  liquibase-linter check db/changelog/db.changelog-master.xml

  # Check all changelogs in a directory
  liquibase-linter check db/changelog/

  # Use a configuration file
  liquibase-linter check --config=.liquibase-linter.yaml db/changelog/

  # Output in JSON format
  liquibase-linter check --format=json db/changelog/

  # List all available rules
  liquibase-linter rules

Use "liquibase-linter <command> -h" for more information about a command.`)
}

func checkCmd() {
	fs := flag.NewFlagSet("check", flag.ExitOnError)
	configPath := fs.String("config", "", "Path to configuration file")
	format := fs.String("format", "text", "Output format (text, json, sarif, junit)")
	noColor := fs.Bool("no-color", false, "Disable colored output")
	severityThreshold := fs.String("severity", "warning", "Minimum severity to report (info, warning, critical)")

	fs.Usage = func() {
		fmt.Println(`Usage: liquibase-linter check [flags] <path>

Check Liquibase changelog files for security vulnerabilities, anti-patterns,
and best practice violations.

Flags:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
  liquibase-linter check db/changelog/db.changelog-master.xml
  liquibase-linter check --format=json db/changelog/
  liquibase-linter check --config=.liquibase-linter.yaml --severity=critical db/`)
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(2)
	}

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: path to changelog file or directory is required")
		fs.Usage()
		os.Exit(2)
	}

	path := fs.Arg(0)

	// Auto-discover config file if not specified
	autoDiscovered := false
	if *configPath == "" {
		*configPath = discoverConfigFile(path)
		if *configPath != "" {
			autoDiscovered = true
		}
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(2)
	}

	// Override config with command-line flags
	if *format != "text" {
		cfg.Output.Format = *format
	}
	if *noColor {
		cfg.Output.Colorize = false
	}
	if *severityThreshold != "warning" {
		cfg.SeverityThreshold = *severityThreshold
	}

	// Inform user about auto-discovered config (only for non-JSON output)
	if autoDiscovered && cfg.Output.Format != "json" {
		fmt.Printf("Using configuration file: %s\n", *configPath)
	}

	// Print configuration summary (skip for JSON output)
	if cfg.Output.Format != "json" {
		printConfigSummary(cfg, path)
	}

	// Discover files to check
	files, err := discoverFiles(path, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering files: %v\n", err)
		os.Exit(2)
	}

	if len(files) == 0 {
		fmt.Println("No changelog files found")
		os.Exit(0)
	}

	// Create rule registry and register all rules
	registry := rules.NewRuleRegistry()

	// Register security rules
	registry.Register(&rules.SQLInjectionRule{})
	registry.Register(&rules.HardcodedCredentialsRule{})
	registry.Register(&rules.DangerousOperationsRule{})

	// Register best practice rules
	registry.Register(&rules.MissingRollbackRule{})
	registry.Register(rules.NewNonIdempotentChangesRule(cfg.Rules["non-idempotent"]))
	registry.Register(&rules.NamingConventionRule{})
	registry.Register(&rules.ContextMisuseRule{})
	registry.Register(rules.NewRedundantOnErrorHaltRule())
	registry.Register(rules.NewNoIfExistsRule())
	registry.Register(&rules.UniqueChangesetRule{})

	// Register label pattern rule (only if enabled in config)
	if cfg.LabelPattern.Enabled {
		registry.Register(rules.NewLabelPatternRule(&cfg.LabelPattern))
	}

	// Register no-manual-transactions rule (only if enabled in config)
	if cfg.Rules["no-manual-transactions"].Enabled {
		registry.Register(rules.NewNoManualTransactionsRule(&cfg.NoManualTransactions))
	}

	// Register atomic-changeset rule (only if enabled in config)
	if cfg.Rules["atomic-changeset"].Enabled {
		registry.Register(rules.NewAtomicChangesetRule(&cfg.AtomicChangeset))
	}

	// Register file structure rules (only if enabled in config)
	if cfg.FileStructure.Enabled {
		registry.Register(rules.NewSprintFolderStructureRule(&cfg.FileStructure))
		registry.Register(rules.NewDDLLocationRule(&cfg.FileStructure))
		registry.Register(rules.NewDMLLocationRule(&cfg.FileStructure))
	}

	// Register performance rules
	registry.Register(&rules.MissingIndexRule{})
	registry.Register(&rules.TableLockRule{})
	registry.Register(&rules.LargeDataOperationRule{})
	registry.Register(&rules.SelectStarRule{})

	// Apply config to enable/disable rules
	applyRuleConfig(registry, cfg)

	// Track parsed files to avoid duplicates
	parsedFiles := make(map[string]bool)
	var allViolations []rules.Violation
	totalFilesProcessed := make(map[string]bool) // Track unique files across all changelogs

	// Parse and check each file
	for _, file := range files {
		normalizedFile := parser.NormalizePath(file)
		if parsedFiles[normalizedFile] {
			continue
		}
		parsedFiles[normalizedFile] = true

		// Determine base path for ignore pattern matching
		fileInfo, statErr := os.Stat(path)
		var basePath string
		if statErr == nil {
			if fileInfo.IsDir() {
				basePath = path
			} else {
				basePath = parser.GetFileDir(path)
			}
			absBasePath, resolveErr := parser.ResolveRelativePath(".", basePath)
			if resolveErr == nil {
				basePath = absBasePath
			}
		}

		// Parse changelog with ignore patterns
		changelog, parseErr := parser.ParseWithConfig(file, cfg.Ignore, basePath)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", file, parseErr)
			continue
		}

		// Track all files processed (root + includes)
		for _, includedFile := range changelog.IncludedFiles {
			normalizedIncluded := parser.NormalizePath(includedFile)
			totalFilesProcessed[normalizedIncluded] = true
		}

		// Validate suppression directives and print warnings
		suppressionWarnings := rules.ValidateSuppressions(changelog, registry)
		for _, warning := range suppressionWarnings {
			fmt.Fprintf(os.Stderr, "Warning: %s (changeset %s:%s in %s)\n",
				warning.Message, warning.Author, warning.ChangeSetID, warning.FilePath)
		}

		// Check with rules
		violations := registry.CheckChangelog(changelog)

		// Filter out suppressed violations
		violations = rules.FilterSuppressedViolations(violations, changelog)

		allViolations = append(allViolations, violations...)
	}

	// Create result
	result := &reporter.Result{
		Violations:    allViolations,
		FilesChecked:  len(totalFilesProcessed),
		LinterVersion: version,
		Timestamp:     time.Now(),
	}

	// Filter by severity threshold
	result.Violations = filterBySeverity(result.Violations, cfg.SeverityThreshold)

	// Get reporter and output results
	rep, err := reporter.GetReporter(reporter.Format(cfg.Output.Format), cfg.Output.Colorize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating reporter: %v\n", err)
		os.Exit(2)
	}

	if err := rep.Report(os.Stdout, result); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating report: %v\n", err)
		os.Exit(2)
	}

	// Exit with appropriate code
	if len(result.Violations) > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}

// discoverFiles discovers all changelog files to check
func discoverFiles(path string, cfg *config.Config) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("path not found: %s", path)
	}

	var files []string
	var basePath string

	if info.IsDir() {
		// Discover all changelog files in directory
		discovered, discoverErr := parser.DiscoverChangelogFiles(path, "", cfg.Parser.FollowSymlinks)
		if discoverErr != nil {
			return nil, discoverErr
		}
		files = discovered
		basePath = path
	} else {
		// Single file
		absPath, resolveErr := parser.ResolveRelativePath(".", path)
		if resolveErr != nil {
			return nil, err
		}
		files = []string{absPath}
		basePath = parser.GetFileDir(absPath)
	}

	// Get absolute base path for relative comparison
	absBasePath, err := parser.ResolveRelativePath(".", basePath)
	if err != nil {
		return nil, err
	}

	// Apply ignore patterns
	filtered := []string{}
	for _, file := range files {
		if !shouldIgnore(file, absBasePath, cfg.Ignore) {
			filtered = append(filtered, file)
		}
	}

	return filtered, nil
}

// discoverConfigFile looks for a .liquibase-linter.yaml file in:
// 1. The target directory (if path is a directory)
// 2. The parent directory of the target file (if path is a file)
// 3. Current working directory
// 4. Parent directories up to root
func discoverConfigFile(targetPath string) string {
	configNames := []string{".liquibase-linter.yaml", ".liquibase-linter.yml"}

	// Determine starting directory
	info, err := os.Stat(targetPath)
	var startDir string
	if err == nil {
		if info.IsDir() {
			startDir = targetPath
		} else {
			startDir = parser.GetFileDir(targetPath)
		}
	} else {
		// If target doesn't exist, use current directory
		var gwdErr error
		startDir, gwdErr = os.Getwd()
		if gwdErr != nil {
			startDir = "."
		}
	}

	// Convert to absolute path
	startDir, err = parser.ResolveRelativePath(".", startDir)
	if err != nil {
		return ""
	}

	// Search in target directory and up
	dir := startDir
	for {
		for _, name := range configNames {
			configPath := parser.JoinPath(dir, name)
			if _, err := os.Stat(configPath); err == nil {
				return configPath
			}
		}

		// Move to parent directory
		parent := parser.GetFileDir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return ""
}

// printConfigSummary prints a summary of the active configuration
func printConfigSummary(cfg *config.Config, targetPath string) {
	fmt.Println("\n=== Configuration Summary ===")

	// Output format
	fmt.Printf("Output Format: %s\n", cfg.Output.Format)
	fmt.Printf("Colorize: %v\n", cfg.Output.Colorize)
	fmt.Printf("Severity Threshold: %s\n", cfg.SeverityThreshold)

	// Parser settings
	fmt.Printf("Max Include Depth: %d\n", cfg.Parser.MaxIncludeDepth)
	fmt.Printf("Follow Symlinks: %v\n", cfg.Parser.FollowSymlinks)

	// Enabled rules
	fmt.Println("\nEnabled Rules:")
	enabledRules := []string{}
	for ruleID, ruleConfig := range cfg.Rules {
		if ruleConfig.Enabled {
			enabledRules = append(enabledRules, fmt.Sprintf("  - %s [%s]", ruleID, ruleConfig.Severity))
		}
	}
	if len(enabledRules) == 0 {
		fmt.Println("  (using default rules)")
	} else {
		for _, rule := range enabledRules {
			fmt.Println(rule)
		}
	}

	// File structure rules
	fmt.Println("\nFile Structure Rules:")
	if cfg.FileStructure.Enabled {
		fmt.Println("  Status: ENABLED")
		fmt.Printf("  Sprint Pattern: %s\n", cfg.FileStructure.SprintPattern)
		fmt.Printf("  Structure Pattern: %s\n", cfg.FileStructure.StructurePattern)
		fmt.Printf("  Data Pattern: %s\n", cfg.FileStructure.DataPattern)
		if cfg.FileStructure.SprintBasePath != "" {
			fmt.Printf("  Sprint Base Path: %s\n", cfg.FileStructure.SprintBasePath)
		}
		if len(cfg.FileStructure.ExcludePatterns) > 0 {
			fmt.Println("  Exclude Patterns:")
			for _, pattern := range cfg.FileStructure.ExcludePatterns {
				fmt.Printf("    - %s\n", pattern)
			}
		}
		fmt.Println("  Rules: sprint-folder-structure, ddl-location, dml-location")
	} else {
		fmt.Println("  Status: DISABLED")
	}

	// Ignored patterns
	fmt.Println("\nIgnore Patterns:")
	if len(cfg.Ignore) == 0 {
		fmt.Println("  (none)")
	} else {
		// Resolve base path for display
		info, err := os.Stat(targetPath)
		var basePath string
		if err == nil {
			if info.IsDir() {
				basePath = targetPath
			} else {
				basePath = parser.GetFileDir(targetPath)
			}
			absBasePath, err := parser.ResolveRelativePath(".", basePath)
			if err == nil {
				basePath = absBasePath
			}
		}

		for _, pattern := range cfg.Ignore {
			if basePath != "" {
				fmt.Printf("  - %s (relative to: %s)\n", pattern, basePath)
			} else {
				fmt.Printf("  - %s\n", pattern)
			}
		}
	}

	fmt.Println("============================")
}

// shouldIgnore checks if a file matches any ignore pattern
func shouldIgnore(file, basePath string, ignorePatterns []string) bool {
	// Make file path relative to base path for pattern matching
	relPath, err := parser.GetRelativePath(basePath, file)
	if err != nil {
		// If we can't make it relative, try with absolute path
		relPath = file
	}

	// Normalize path separators for consistent matching
	relPath = parser.NormalizePath(relPath)

	for _, pattern := range ignorePatterns {
		matched, err := parser.MatchesResourceFilter(relPath, pattern)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// applyRuleConfig enables/disables rules based on configuration
func applyRuleConfig(registry *rules.RuleRegistry, cfg *config.Config) {
	// Apply rule configurations
	for ruleID, ruleCfg := range cfg.Rules {
		// Disable rule if not enabled
		if !ruleCfg.Enabled {
			registry.Disable(ruleID)
		}

		// Apply severity override if configured
		if ruleCfg.Severity != "" {
			if severity, err := rules.ParseSeverity(ruleCfg.Severity); err == nil {
				registry.SetSeverity(ruleID, severity)
			}
		}
	}
}

// filterBySeverity filters violations by severity threshold
func filterBySeverity(violations []rules.Violation, threshold string) []rules.Violation {
	if threshold == "info" {
		return violations // All violations
	}

	var filtered []rules.Violation
	for _, v := range violations {
		if threshold == "warning" && (v.Severity == rules.SeverityWarning || v.Severity == rules.SeverityCritical) {
			filtered = append(filtered, v)
		} else if threshold == "critical" && v.Severity == rules.SeverityCritical {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func rulesCmd() {
	fs := flag.NewFlagSet("rules", flag.ExitOnError)
	info := fs.String("info", "", "Show detailed information about a specific rule")

	fs.Usage = func() {
		fmt.Println(`Usage: liquibase-linter rules [flags]

List all available linting rules or show detailed information about a specific rule.

Flags:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
  liquibase-linter rules
  liquibase-linter rules --info=sql-injection`)
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(2)
	}

	if *info != "" {
		// Show detailed rule information
		registry := rules.NewRuleRegistry()
		registerAllRules(registry)

		rule, exists := registry.GetRule(*info)
		if !exists {
			fmt.Fprintf(os.Stderr, "Rule not found: %s\n", *info)
			os.Exit(1)
		}

		fmt.Printf("Rule: %s\n", rule.ID())
		fmt.Printf("Name: %s\n", rule.Name())
		fmt.Printf("Severity: %s\n", rule.Severity())
		fmt.Printf("Description: %s\n", rule.Description())
		return
	}

	// List all rules
	registry := rules.NewRuleRegistry()
	registerAllRules(registry)

	fmt.Println("Available Linting Rules:")
	fmt.Println()

	// Group by category
	fmt.Println("Security Rules:")
	for _, rule := range registry.GetAllRules() {
		if isSecurityRule(rule) {
			fmt.Printf("  %-30s [%s] %s\n", rule.ID(), rule.Severity(), rule.Name())
		}
	}

	fmt.Println("\nBest Practice Rules:")
	for _, rule := range registry.GetAllRules() {
		if isBestPracticeRule(rule) {
			fmt.Printf("  %-30s [%s] %s\n", rule.ID(), rule.Severity(), rule.Name())
		}
	}

	fmt.Println("\nPerformance Rules:")
	for _, rule := range registry.GetAllRules() {
		if isPerformanceRule(rule) {
			fmt.Printf("  %-30s [%s] %s\n", rule.ID(), rule.Severity(), rule.Name())
		}
	}

	fmt.Println("\nUse 'liquibase-linter rules --info=<rule-id>' for detailed information.")
}

// registerAllRules registers all available rules
func registerAllRules(registry *rules.RuleRegistry) {
	// Security rules
	registry.Register(&rules.SQLInjectionRule{})
	registry.Register(&rules.HardcodedCredentialsRule{})
	registry.Register(&rules.DangerousOperationsRule{})

	// Best practice rules
	registry.Register(&rules.MissingRollbackRule{})
	registry.Register(&rules.NonIdempotentChangesRule{})
	registry.Register(&rules.NamingConventionRule{})
	registry.Register(&rules.ContextMisuseRule{})

	// Performance rules
	registry.Register(&rules.MissingIndexRule{})
	registry.Register(&rules.TableLockRule{})
	registry.Register(&rules.LargeDataOperationRule{})
	registry.Register(&rules.SelectStarRule{})
}

// isSecurityRule checks if a rule is a security rule
func isSecurityRule(rule rules.Rule) bool {
	id := rule.ID()
	return id == "sql-injection" || id == "hardcoded-credentials" || id == "dangerous-operations"
}

// isBestPracticeRule checks if a rule is a best practice rule
func isBestPracticeRule(rule rules.Rule) bool {
	id := rule.ID()
	return id == "missing-rollback" || id == "non-idempotent" || id == "naming-convention" || id == "context-misuse"
}

// isPerformanceRule checks if a rule is a performance rule
func isPerformanceRule(rule rules.Rule) bool {
	id := rule.ID()
	return id == "missing-index" || id == "table-lock" || id == "large-data-operation" || id == "select-star"
}

func initCmd() {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	outputPath := fs.String("output", ".liquibase-linter.yaml", "Output path for configuration file")

	fs.Usage = func() {
		fmt.Println(`Usage: liquibase-linter init [flags]

Initialize a configuration file with default settings.

Flags:`)
		fs.PrintDefaults()
		fmt.Println(`
Examples:
  liquibase-linter init
  liquibase-linter init --output=custom-config.yaml`)
	}

	if err := fs.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(2)
	}

	if err := config.InitConfig(*outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating configuration file: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("Configuration file created: %s\n", *outputPath)
}

func versionCmd() {
	fmt.Printf("liquibase-linter version %s", version)
	if commitHash != "" || buildDate != "" {
		fmt.Print(" (")
		if commitHash != "" {
			fmt.Printf("commit: %s", commitHash)
			if buildDate != "" {
				fmt.Print(", ")
			}
		}
		if buildDate != "" {
			fmt.Printf("built: %s", buildDate)
		}
		fmt.Print(")")
	}
	fmt.Println()
}
