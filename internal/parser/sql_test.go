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
	if len(cs3.Context) == 0 {
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
