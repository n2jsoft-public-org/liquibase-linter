package parser

import (
	"path/filepath"
	"testing"
)

func TestXMLParser_ParseValidChangelog(t *testing.T) {
	parser := &XMLParser{}
	filePath := filepath.Join("..", "..", "testdata", "valid-changelog.xml")

	changelog, err := parser.Parse(filePath)
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if len(changelog.ChangeSets) != 2 {
		t.Fatalf("Expected 2 changesets, got %d", len(changelog.ChangeSets))
	}

	// Test first changeset
	cs1 := changelog.ChangeSets[0]
	if cs1.ID != "1" {
		t.Errorf("ChangeSet[0].ID = %v, want 1", cs1.ID)
	}
	if cs1.Author != "test" {
		t.Errorf("ChangeSet[0].Author = %v, want test", cs1.Author)
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

	// Test second changeset
	cs2 := changelog.ChangeSets[1]
	if cs2.ID != "2" {
		t.Errorf("ChangeSet[1].ID = %v, want 2", cs2.ID)
	}
	if len(cs2.Changes) == 0 {
		t.Fatal("ChangeSet[1] has no changes")
	}
}

func TestXMLParser_ParseProblematicChangelog(t *testing.T) {
	parser := &XMLParser{}
	filePath := filepath.Join("..", "..", "testdata", "problematic-changelog.xml")

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
}

func TestXMLParser_FileNotFound(t *testing.T) {
	parser := &XMLParser{}
	_, err := parser.Parse("nonexistent.xml")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
