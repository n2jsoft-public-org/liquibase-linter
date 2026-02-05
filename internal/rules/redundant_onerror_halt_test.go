package rules

import (
	"testing"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

func TestRedundantOnErrorHaltRule_Check(t *testing.T) {
	rule := NewRedundantOnErrorHaltRule()

	tests := []struct {
		name          string
		changelog     *parser.Changelog
		wantViolation bool
	}{
		{
			name: "redundant onError:HALT",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "1",
						Author:   "test",
						FilePath: "/test/changelog.xml",
						Preconditions: &parser.Precondition{
							Type:    "tableExists",
							OnFail:  "MARK_RAN",
							OnError: "HALT",
						},
					},
				},
			},
			wantViolation: true,
		},
		{
			name: "onError:HALT lowercase",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "2",
						Author:   "test",
						FilePath: "/test/changelog.xml",
						Preconditions: &parser.Precondition{
							Type:    "tableExists",
							OnFail:  "MARK_RAN",
							OnError: "halt",
						},
					},
				},
			},
			wantViolation: true,
		},
		{
			name: "onError:HALT mixed case",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "3",
						Author:   "test",
						FilePath: "/test/changelog.xml",
						Preconditions: &parser.Precondition{
							Type:    "tableExists",
							OnFail:  "MARK_RAN",
							OnError: "Halt",
						},
					},
				},
			},
			wantViolation: true,
		},
		{
			name: "onError:WARN - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "4",
						Author:   "test",
						FilePath: "/test/changelog.xml",
						Preconditions: &parser.Precondition{
							Type:    "tableExists",
							OnFail:  "MARK_RAN",
							OnError: "WARN",
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "onError:CONTINUE - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "5",
						Author:   "test",
						FilePath: "/test/changelog.xml",
						Preconditions: &parser.Precondition{
							Type:    "tableExists",
							OnFail:  "MARK_RAN",
							OnError: "CONTINUE",
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "onError:MARK_RAN - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "6",
						Author:   "test",
						FilePath: "/test/changelog.xml",
						Preconditions: &parser.Precondition{
							Type:    "tableExists",
							OnFail:  "MARK_RAN",
							OnError: "MARK_RAN",
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "no onError attribute - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:       "7",
						Author:   "test",
						FilePath: "/test/changelog.xml",
						Preconditions: &parser.Precondition{
							Type:    "tableExists",
							OnFail:  "MARK_RAN",
							OnError: "",
						},
					},
				},
			},
			wantViolation: false,
		},
		{
			name: "no preconditions - no violation",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:            "8",
						Author:        "test",
						FilePath:      "/test/changelog.xml",
						Preconditions: nil,
					},
				},
			},
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.Check(tt.changelog)

			if tt.wantViolation && len(violations) == 0 {
				t.Error("Expected violation but got none")
			}
			if !tt.wantViolation && len(violations) > 0 {
				t.Errorf("Expected no violations but got %d: %v", len(violations), violations)
			}
		})
	}
}
