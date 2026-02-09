package parser

import (
	"path/filepath"
	"testing"
)

func TestSQLParser_ParseValidChangelog(t *testing.T) {
	parser := &SQLParser{}
	filePath := filepath.Join("..", "..", "testdata", "valid-changelog.sql")

	changelog, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(changelog.ChangeSets) != 3 {
		t.Fatalf("Expected 3 changesets, got %d", len(changelog.ChangeSets))
	}

	// Test first changeset
	cs1 := changelog.ChangeSets[0]
	if cs1.ID != "1" {
		t.Errorf("ChangeSet[0].ID = %v, want 1", cs1.ID)
	}
	if cs1.Author != "john" {
		t.Errorf("ChangeSet[0].Author = %v, want john", cs1.Author)
	}
	if len(cs1.Changes) == 0 {
		t.Fatal("ChangeSet[0] has no changes")
	}
	if cs1.Changes[0].Type != "createTable" {
		t.Errorf("ChangeSet[0].Changes[0].Type = %v, want createTable", cs1.Changes[0].Type)
	}
	if !cs1.HasRollback() {
		t.Error("ChangeSet[0] should have rollback")
	}

	// Test second changeset - has labels
	cs2 := changelog.ChangeSets[1]
	if cs2.ID != "2" {
		t.Errorf("ChangeSet[1].ID = %v, want 2", cs2.ID)
	}
	if len(cs2.Labels) == 0 {
		t.Error("ChangeSet[1] should have labels")
	}

	// Test third changeset - has context
	cs3 := changelog.ChangeSets[2]
	if cs3.ID != "3" {
		t.Errorf("ChangeSet[2].ID = %v, want 3", cs3.ID)
	}
	if cs3.Context == "" {
		t.Error("ChangeSet[2] should have context")
	}
}

func TestSQLParser_ParseProblematicChangelog(t *testing.T) {
	parser := &SQLParser{}
	filePath := filepath.Join("..", "..", "testdata", "problematic-changelog.sql")

	changelog, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(changelog.ChangeSets) < 1 {
		t.Fatal("Expected at least 1 changeset")
	}

	// Check that problematic changesets can still be parsed
	cs1 := changelog.ChangeSets[0]
	if cs1.ID == "" {
		t.Error("ChangeSet ID should not be empty")
	}

	// Verify missing rollback
	if cs1.HasRollback() {
		t.Error("ChangeSet[0] should not have rollback (problematic)")
	}
}

func TestSQLParser_FileNotFound(t *testing.T) {
	parser := &SQLParser{}
	_, err := parser.Parse("nonexistent.sql")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestSQLParser_ParseMultipleComments(t *testing.T) {
	parser := &SQLParser{}
	filePath := filepath.Join("..", "..", "testdata", "test-multiple-comments.sql")

	changelog, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(changelog.ChangeSets) != 7 {
		t.Fatalf("Expected 7 changesets, got %d", len(changelog.ChangeSets))
	}

	// Test changeset 1: suppression in first comment, description in second
	cs1 := changelog.ChangeSets[0]
	if cs1.ID != "1" || cs1.Author != "john" {
		t.Errorf("ChangeSet[0] ID/Author mismatch: %s:%s", cs1.Author, cs1.ID)
	}
	// Comment should contain both lines
	if cs1.Comment == "" {
		t.Error("ChangeSet[0] Comment should not be empty")
	}
	// Should contain both comment contents
	if !contains(cs1.Comment, "liquibase-linter:disable sql-injection") {
		t.Errorf("ChangeSet[0] Comment missing suppression directive: %s", cs1.Comment)
	}
	if !contains(cs1.Comment, "This creates the users table") {
		t.Errorf("ChangeSet[0] Comment missing description: %s", cs1.Comment)
	}
	// Should have sql-injection suppressed
	if !containsRule(cs1.SuppressedRules, "sql-injection") {
		t.Errorf("ChangeSet[0] should suppress sql-injection, got: %v", cs1.SuppressedRules)
	}

	// Test changeset 2: description in first comment, suppression in second
	cs2 := changelog.ChangeSets[1]
	if cs2.ID != "2" || cs2.Author != "jane" {
		t.Errorf("ChangeSet[1] ID/Author mismatch: %s:%s", cs2.Author, cs2.ID)
	}
	if !contains(cs2.Comment, "Update email field") {
		t.Errorf("ChangeSet[1] Comment missing description: %s", cs2.Comment)
	}
	if !contains(cs2.Comment, "liquibase-linter:disable missing-rollback") {
		t.Errorf("ChangeSet[1] Comment missing suppression: %s", cs2.Comment)
	}
	if !containsRule(cs2.SuppressedRules, "missing-rollback") {
		t.Errorf("ChangeSet[1] should suppress missing-rollback, got: %v", cs2.SuppressedRules)
	}

	// Test changeset 3: multiple suppressions across multiple comments
	cs3 := changelog.ChangeSets[2]
	if cs3.ID != "3" || cs3.Author != "bob" {
		t.Errorf("ChangeSet[2] ID/Author mismatch: %s:%s", cs3.Author, cs3.ID)
	}
	// Should have both sql-injection and hardcoded-credentials suppressed
	if !containsRule(cs3.SuppressedRules, "sql-injection") {
		t.Errorf("ChangeSet[2] should suppress sql-injection, got: %v", cs3.SuppressedRules)
	}
	if !containsRule(cs3.SuppressedRules, "hardcoded-credentials") {
		t.Errorf("ChangeSet[2] should suppress hardcoded-credentials, got: %v", cs3.SuppressedRules)
	}
	// Should have all three comment lines
	if !contains(cs3.Comment, "liquibase-linter:disable sql-injection") {
		t.Errorf("ChangeSet[2] Comment missing first suppression: %s", cs3.Comment)
	}
	if !contains(cs3.Comment, "liquibase-linter:disable hardcoded-credentials") {
		t.Errorf("ChangeSet[2] Comment missing second suppression: %s", cs3.Comment)
	}
	if !contains(cs3.Comment, "test environment only") {
		t.Errorf("ChangeSet[2] Comment missing description: %s", cs3.Comment)
	}

	// Test changeset 4: multiple comments without suppressions
	cs4 := changelog.ChangeSets[3]
	if cs4.ID != "4" || cs4.Author != "alice" {
		t.Errorf("ChangeSet[3] ID/Author mismatch: %s:%s", cs4.Author, cs4.ID)
	}
	if !contains(cs4.Comment, "First comment line") {
		t.Errorf("ChangeSet[3] Comment missing first line: %s", cs4.Comment)
	}
	if !contains(cs4.Comment, "Second comment line") {
		t.Errorf("ChangeSet[3] Comment missing second line: %s", cs4.Comment)
	}
	if !contains(cs4.Comment, "Third comment line") {
		t.Errorf("ChangeSet[3] Comment missing third line: %s", cs4.Comment)
	}
	if len(cs4.SuppressedRules) != 0 {
		t.Errorf("ChangeSet[3] should have no suppressions, got: %v", cs4.SuppressedRules)
	}

	// Test changeset 5: multiple rules in one comment
	cs5 := changelog.ChangeSets[4]
	if cs5.ID != "5" || cs5.Author != "charlie" {
		t.Errorf("ChangeSet[4] ID/Author mismatch: %s:%s", cs5.Author, cs5.ID)
	}
	if !containsRule(cs5.SuppressedRules, "sql-injection") {
		t.Errorf("ChangeSet[4] should suppress sql-injection, got: %v", cs5.SuppressedRules)
	}
	if !containsRule(cs5.SuppressedRules, "missing-rollback") {
		t.Errorf("ChangeSet[4] should suppress missing-rollback, got: %v", cs5.SuppressedRules)
	}

	// Test changeset 6: single comment (backward compatibility)
	cs6 := changelog.ChangeSets[5]
	if cs6.ID != "6" || cs6.Author != "dave" {
		t.Errorf("ChangeSet[5] ID/Author mismatch: %s:%s", cs6.Author, cs6.ID)
	}
	if !contains(cs6.Comment, "Regular single comment") {
		t.Errorf("ChangeSet[5] Comment unexpected: %s", cs6.Comment)
	}
	if len(cs6.SuppressedRules) != 0 {
		t.Errorf("ChangeSet[5] should have no suppressions, got: %v", cs6.SuppressedRules)
	}

	// Test changeset 7: suppressions across multiple comment lines with interleaved documentation
	cs7 := changelog.ChangeSets[6]
	if cs7.ID != "7" || cs7.Author != "eve" {
		t.Errorf("ChangeSet[6] ID/Author mismatch: %s:%s", cs7.Author, cs7.ID)
	}
	if !containsRule(cs7.SuppressedRules, "sql-injection") {
		t.Errorf("ChangeSet[6] should suppress sql-injection, got: %v", cs7.SuppressedRules)
	}
	if !containsRule(cs7.SuppressedRules, "missing-rollback") {
		t.Errorf("ChangeSet[6] should suppress missing-rollback, got: %v", cs7.SuppressedRules)
	}
	if !contains(cs7.Comment, "Additional context") {
		t.Errorf("ChangeSet[6] Comment missing context line: %s", cs7.Comment)
	}
	if !contains(cs7.Comment, "Even more documentation") {
		t.Errorf("ChangeSet[6] Comment missing documentation line: %s", cs7.Comment)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to check if a rule is in the suppressed rules list
func containsRule(rules []string, rule string) bool {
	for _, r := range rules {
		if r == rule {
			return true
		}
	}
	return false
}
