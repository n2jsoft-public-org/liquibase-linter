package parser

import (
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
			if !tt.wantErr && result != tt.expected {
				t.Errorf("GetRelativePath() = %v, want %v", result, tt.expected)
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
			if result != tt.expected {
				t.Errorf("GetFileDir() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		name     string
		elements []string
		expected string
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
			if result != tt.expected {
				t.Errorf("JoinPath() = %v, want %v", result, tt.expected)
			}
		})
	}
}
