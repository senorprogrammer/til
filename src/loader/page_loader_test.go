package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPageLoader_Load(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  []string
		extension   string
		wantCount   int
		wantErr     bool
		wantOrder   bool // true if pages should be in reverse chronological order
	}{
		{
			name: "single page file",
			setupFiles: []string{
				"2020-01-01T00-00-00-test.md",
			},
			extension: "md",
			wantCount: 1,
			wantErr:   false,
			wantOrder: true,
		},
		{
			name: "multiple page files",
			setupFiles: []string{
				"2020-01-01T00-00-00-first.md",
				"2020-02-01T00-00-00-second.md",
				"2020-03-01T00-00-00-third.md",
			},
			extension: "md",
			wantCount: 3,
			wantErr:   false,
			wantOrder: true,
		},
		{
			name: "no page files",
			setupFiles: []string{},
			extension: "md",
			wantCount: 0,
			wantErr:   false,
			wantOrder: false,
		},
		{
			name: "different extension",
			setupFiles: []string{
				"2020-01-01T00-00-00-test.txt",
			},
			extension: "md",
			wantCount: 0,
			wantErr:   false,
			wantOrder: false,
		},
		{
			name: "mixed extensions",
			setupFiles: []string{
				"2020-01-01T00-00-00-test.md",
				"2020-02-01T00-00-00-other.txt",
			},
			extension: "md",
			wantCount: 1,
			wantErr:   false,
			wantOrder: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create test files
			for _, fileName := range tt.setupFiles {
				filePath := filepath.Join(tmpDir, fileName)
				content := `---
date: 2020-01-01T00:00:00Z
title: Test Page
tags: test
---

# Test Page

Content.
`
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					t.Fatalf("failed to create test file %q: %v", fileName, err)
				}
			}

			loader := NewPageLoader(tmpDir, tt.extension)
			got, err := loader.Load()

			if (err != nil) != tt.wantErr {
				t.Errorf("PageLoader.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("PageLoader.Load() count = %d, want %d", len(got), tt.wantCount)
			}

			// Note: PageFromFilePath calls src.Defeat on error which exits
			// So we can only test the happy path here
		})
	}
}

func TestNewPageLoader(t *testing.T) {
	tests := []struct {
		name      string
		targetDir string
		extension string
	}{
		{
			name:      "valid loader",
			targetDir: "/tmp/test",
			extension: "md",
		},
		{
			name:      "different extension",
			targetDir: "/tmp/test",
			extension: "txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewPageLoader(tt.targetDir, tt.extension)

			if loader == nil {
				t.Fatal("NewPageLoader() returned nil")
			}

			if loader.targetDir != tt.targetDir {
				t.Errorf("NewPageLoader() targetDir = %q, want %q", loader.targetDir, tt.targetDir)
			}

			if loader.extension != tt.extension {
				t.Errorf("NewPageLoader() extension = %q, want %q", loader.extension, tt.extension)
			}
		})
	}
}
