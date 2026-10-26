package parser

import (
	"path/filepath"
	"testing"
)

func TestYAMLParser_Parse_PreconditionList(t *testing.T) {
	parser := &YAMLParser{}
	changelog, err := parser.Parse(filepath.Join("..", "..", "testdata", "dbms-scope", "changelog.yaml"))
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	cs := findChangeSet(t, changelog, "catalog-on-pg")
	want := "SELECT COUNT(*) FROM sys.tables WHERE name = 'c'"
	if len(cs.PreconditionSQL) != 1 || cs.PreconditionSQL[0] != want {
		t.Errorf("PreconditionSQL = %q, want [%q]", cs.PreconditionSQL, want)
	}
	if cs.Preconditions == nil || cs.Preconditions.OnFail != "MARK_RAN" {
		t.Errorf("Preconditions = %+v, want onFail MARK_RAN", cs.Preconditions)
	}
}

func TestPreconditionSQLChecks(t *testing.T) {
	tests := []struct {
		data any
		name string
		want int
	}{
		{
			name: "list with two checks",
			data: []any{
				map[string]any{"onFail": "MARK_RAN"},
				map[string]any{"sqlCheck": map[string]any{"sql": "SELECT 1"}},
				map[string]any{"sqlCheck": map[string]any{"sql": "SELECT 2"}},
			},
			want: 2,
		},
		{
			name: "single object",
			data: map[string]any{"sqlCheck": map[string]any{"expectedResult": 0, "sql": "SELECT 1"}},
			want: 1,
		},
		{
			name: "no sqlCheck",
			data: []any{map[string]any{"tableExists": map[string]any{"tableName": "a"}}},
			want: 0,
		},
		{
			name: "empty sql",
			data: map[string]any{"sqlCheck": map[string]any{"sql": "  "}},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := preconditionSQLChecks(tt.data); len(got) != tt.want {
				t.Errorf("preconditionSQLChecks() = %q, want %d entries", got, tt.want)
			}
		})
	}
}
