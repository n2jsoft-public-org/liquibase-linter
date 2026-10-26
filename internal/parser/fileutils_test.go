package parser

import (
	"path/filepath"
	"testing"
)

func TestGetRelativePath(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		target   string
		expected string
		wantErr  bool
	}{
		{
			name:     "same directory",
			base:     "/path/to/base",
			target:   "/path/to/base/file.txt",
			expected: "file.txt",
			wantErr:  false,
		},
		{
			name:     "subdirectory",
			base:     "/path/to/base",
			target:   "/path/to/base/sub/dir/file.txt",
			expected: "sub/dir/file.txt",
			wantErr:  false,
		},
		{
			name:     "parent directory",
			base:     "/path/to/base/sub",
			target:   "/path/to/base/file.txt",
			expected: "../file.txt",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetRelativePath(tt.base, tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRelativePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != filepath.FromSlash(tt.expected) {
				t.Errorf("GetRelativePath() = %v, want %v", result, filepath.FromSlash(tt.expected))
			}
		})
	}
}

func TestGetFileDir(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "unix path",
			filePath: "/path/to/file.txt",
			expected: "/path/to",
		},
		{
			name:     "unix path with trailing slash",
			filePath: "/path/to/dir/",
			expected: "/path/to/dir",
		},
		{
			name:     "relative path",
			filePath: "path/to/file.txt",
			expected: "path/to",
		},
		{
			name:     "file in root",
			filePath: "/file.txt",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetFileDir(tt.filePath)
			if result != filepath.FromSlash(tt.expected) {
				t.Errorf("GetFileDir() = %v, want %v", result, filepath.FromSlash(tt.expected))
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		elements []string
	}{
		{
			name:     "two elements",
			elements: []string{"/path/to", "file.txt"},
			expected: "/path/to/file.txt",
		},
		{
			name:     "multiple elements",
			elements: []string{"/path", "to", "sub", "file.txt"},
			expected: "/path/to/sub/file.txt",
		},
		{
			name:     "with trailing slashes",
			elements: []string{"/path/", "to/", "file.txt"},
			expected: "/path/to/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinPath(tt.elements...)
			if result != filepath.FromSlash(tt.expected) {
				t.Errorf("JoinPath() = %v, want %v", result, filepath.FromSlash(tt.expected))
			}
		})
	}
}

func TestMatchesResourceFilter_RecursivePatterns(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		pattern  string
		want     bool
	}{
		// Multiple ** segments
		{"init dir absolute path", "/repo/db/changelog/init/a.sql", "**/init/**", true},
		{"init dir nested", "/repo/db/init/sub/deep/a.xml", "**/init/**", true},
		{"init dir relative path", "init/a.sql", "**/init/**", true},
		{"init dir windows path", `C:\repo\db\init\a.sql`, "**/init/**", true},
		{"init as file name prefix", "/repo/db/initial/a.sql", "**/init/**", false},
		{"init as file name", "/repo/db/init.sql", "**/init/**", false},
		{"generated dir", "/repo/db/generated/v1/a.yaml", "**/generated/**", true},
		{"generated missing", "/repo/db/changelog/a.yaml", "**/generated/**", false},

		// ** in the middle with a segment pattern after it
		{"data under changelog", "changelog/sprint1/data/a.sql", "changelog/**/data/*.sql", true},
		{"data directly under changelog", "changelog/data/a.sql", "changelog/**/data/*.sql", true},
		{"data under absolute changelog", "/repo/changelog/s1/s2/data/a.sql", "changelog/**/data/*.sql", true},
		{"data wrong extension", "changelog/sprint1/data/a.xml", "changelog/**/data/*.sql", false},
		{"data file too deep", "changelog/data/sub/a.sql", "changelog/**/data/*.sql", false},
		{"data outside changelog", "other/sprint1/data/a.sql", "changelog/**/data/*.sql", false},

		// Single ** (existing behaviour)
		{"any sql file", "/repo/db/a.sql", "**/*.sql", true},
		{"any sql file relative", "a.sql", "**/*.sql", true},
		{"not a sql file", "/repo/db/a.xml", "**/*.sql", false},
		{"ignore directory", "changelog/init/001.sql", "changelog/init/**", true},
		{"ignore directory absolute", "/repo/changelog/init/sub/001.sql", "changelog/init/**", true},
		{"ignore directory other", "changelog/v1/001.sql", "changelog/init/**", false},
		{"versioned dirs", "v1/sub/a.xml", "v*/**/*.xml", true},
		{"versioned dirs wrong ext", "v1/sub/a.sql", "v*/**/*.xml", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchesResourceFilter(tt.filePath, tt.pattern)
			if err != nil {
				t.Fatalf("MatchesResourceFilter() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("MatchesResourceFilter(%q, %q) = %v, want %v", tt.filePath, tt.pattern, got, tt.want)
			}
		})
	}
}
