package commands

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/olebedev/config"
	"github.com/senorprogrammer/til/pages"
	"github.com/senorprogrammer/til/src"
	"github.com/senorprogrammer/til/src/builder"
	"github.com/senorprogrammer/til/src/git"
	"github.com/senorprogrammer/til/src/loader"
)

// mockLogger is a test logger that records log calls
type mockLogger struct {
	infoCalls     []string
	progressCalls []string
}

func (m *mockLogger) Info(msg string) {
	m.infoCalls = append(m.infoCalls, msg)
}

func (m *mockLogger) Progress(msg string) {
	m.progressCalls = append(m.progressCalls, msg)
}

func (m *mockLogger) Defeat(err error) {
	m.infoCalls = append(m.infoCalls, "defeat: "+err.Error())
}

func (m *mockLogger) Victory(msg string) {
	m.infoCalls = append(m.infoCalls, "victory: "+msg)
}

// mockRepository is a test git repository
type mockRepository struct {
	commitErr error
	pushErr   error
}

func (m *mockRepository) Commit(msg string) error {
	return m.commitErr
}

func (m *mockRepository) Push() error {
	return m.pushErr
}

func TestCommitMessageResolver_Resolve(t *testing.T) {
	tests := []struct {
		name        string
		configMsg   string
		defaultMsg  string
		args        []string
		nonFlagArgs int
		expected    string
	}{
		{
			name:        "no non-flag args uses config message",
			configMsg:   "from config",
			defaultMsg:  "default message",
			args:        []string{"til", "-t", "a", "-s"},
			nonFlagArgs: 0,
			expected:    "from config",
		},
		{
			name:        "no non-flag args uses default when config empty",
			configMsg:   "",
			defaultMsg:  "default message",
			args:        []string{"til", "-t", "a", "-s"},
			nonFlagArgs: 0,
			expected:    "default message",
		},
		{
			name:        "non-flag args uses args as message",
			configMsg:   "from config",
			defaultMsg:  "default message",
			args:        []string{"til", "-t", "a", "-s", "custom", "message"},
			nonFlagArgs: 2,
			expected:    "custom message",
		},
		{
			name:        "non-flag args single word",
			configMsg:   "from config",
			defaultMsg:  "default message",
			args:        []string{"til", "-s", "single"},
			nonFlagArgs: 1,
			expected:    "single",
		},
		{
			name:        "non-flag args multiple words",
			configMsg:   "from config",
			defaultMsg:  "default message",
			args:        []string{"til", "-s", "this", "is", "a", "long", "message"},
			nonFlagArgs: 5,
			expected:    "this is a long message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgYaml := "commitMessage: " + tt.configMsg
			cfg, err := config.ParseYamlBytes([]byte(cfgYaml))
			if err != nil {
				t.Fatalf("failed to parse config: %v", err)
			}

			resolver := NewCommitMessageResolver(cfg, tt.defaultMsg)
			actual := resolver.Resolve(tt.args, tt.nonFlagArgs)

			if actual != tt.expected {
				t.Errorf("Resolve() = %q, expected %q", actual, tt.expected)
			}
		})
	}
}

func TestListCommand_Execute(t *testing.T) {
	tests := []struct {
		name          string
		configYaml    string
		expectedError bool
		expectedLogs  int
	}{
		{
			name: "success with single directory",
			configYaml: `targetDirectories:
  a: ~/Documents/blog`,
			expectedError: false,
			expectedLogs:  1,
		},
		{
			name: "success with multiple directories",
			configYaml: `targetDirectories:
  a: ~/Documents/blog
  b: ~/Documents/notes`,
			expectedError: false,
			expectedLogs:  2,
		},
		{
			name: "error when config key missing",
			configYaml: `otherKey: value`,
			expectedError: true,
			expectedLogs:  0,
		},
		{
			name:          "error with empty config",
			configYaml:    ``,
			expectedError: true,
			expectedLogs:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.ParseYamlBytes([]byte(tt.configYaml))
			if err != nil {
				t.Fatalf("failed to parse config: %v", err)
			}

			logger := &mockLogger{}
			cmd := NewListCommand(cfg, logger)

			err = cmd.Execute()

			if tt.expectedError && err == nil {
				t.Error("Execute() expected error but got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Execute() unexpected error: %v", err)
			}
			if len(logger.infoCalls) != tt.expectedLogs {
				t.Errorf("Execute() logged %d times, expected %d", len(logger.infoCalls), tt.expectedLogs)
			}
		})
	}
}

func TestBuildCommand_Execute(t *testing.T) {
	tests := []struct {
		name          string
		setupPages    bool
		expectedError bool
		errorContains string
	}{
		{
			name:          "success with pages",
			setupPages:    true,
			expectedError: false,
		},
		{
			name:          "success with no pages",
			setupPages:    false,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			docsDir := filepath.Join(tmpDir, "docs")
			if err := os.MkdirAll(docsDir, 0755); err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}

			if tt.setupPages {
				// Create a test page file
				pageFile := filepath.Join(docsDir, "2020-01-01T00-00-00-test.md")
				pageContent := `---
date: 2020-01-01T00:00:00Z
title: Test Page
tags: go
---

# Test Page

Content here.
`
				if err := os.WriteFile(pageFile, []byte(pageContent), 0644); err != nil {
					t.Fatalf("failed to write test page: %v", err)
				}
			}

			loader := loader.NewPageLoader(docsDir, pages.FileExtension)
			builder := builder.NewSiteBuilder(docsDir, builder.DefaultFileWriter, &mockLogger{})
			logger := &mockLogger{}
			cmd := NewBuildCommand(builder, loader, logger)

			err := cmd.Execute()

			if tt.expectedError {
				if err == nil {
					t.Error("Execute() expected error but got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Execute() error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("Execute() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSaveCommand_Execute(t *testing.T) {
	tests := []struct {
		name          string
		commitErr     error
		pushErr       error
		args          []string
		nonFlagArgs   int
		expectedError bool
		errorContains string
	}{
		{
			name:          "success",
			args:          []string{"til", "-s"},
			nonFlagArgs:   0,
			expectedError: false,
		},
		{
			name:          "commit error",
			args:          []string{"til", "-s"},
			commitErr:     errors.New("commit failed"),
			expectedError: true,
			errorContains: "commit failed",
		},
		{
			name:          "push error",
			args:          []string{"til", "-s"},
			pushErr:       errors.New("push failed"),
			expectedError: true,
			errorContains: "push failed",
		},
		{
			name:          "success with custom commit message",
			args:          []string{"til", "-s", "custom", "message"},
			nonFlagArgs:   2,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up flag parsing for NArg()
			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
			// Define the flags that might be in args
			flag.Bool("s", false, "")
			flag.Bool("save", false, "")
			flag.String("t", "", "")
			flag.String("target", "", "")
			flag.Parse()

			tmpDir := t.TempDir()
			docsDir := filepath.Join(tmpDir, "docs")
			if err := os.MkdirAll(docsDir, 0755); err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}

			// Create a test page file
			pageFile := filepath.Join(docsDir, "2020-01-01T00-00-00-test.md")
			pageContent := `---
date: 2020-01-01T00:00:00Z
title: Test Page
tags: go
---

# Test Page

Content here.
`
			if err := os.WriteFile(pageFile, []byte(pageContent), 0644); err != nil {
				t.Fatalf("failed to write test page: %v", err)
			}

			loader := loader.NewPageLoader(docsDir, pages.FileExtension)
			builder := builder.NewSiteBuilder(docsDir, builder.DefaultFileWriter, &mockLogger{})
			var repo git.Repository = &mockRepository{
				commitErr: tt.commitErr,
				pushErr:   tt.pushErr,
			}

			cfg, _ := config.ParseYamlBytes([]byte("commitMessage: default"))
			resolver := NewCommitMessageResolver(cfg, "default message")

			logger := &mockLogger{}
			cmd := NewSaveCommand(builder, loader, repo, resolver, tt.args, logger)

			err := cmd.Execute()

			if tt.expectedError {
				if err == nil {
					t.Error("Execute() expected error but got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Execute() error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("Execute() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCreateCommand_Execute(t *testing.T) {
	tests := []struct {
		name          string
		targetDir     string
		title         string
		editor        string
		expectedError bool
		errorContains string
	}{
		{
			name:          "success",
			targetDir:     t.TempDir(),
			title:         "Test Page",
			editor:        "echo",
			expectedError: false,
		},
		{
			name:          "error with invalid editor",
			targetDir:     t.TempDir(),
			title:         "Test Page",
			editor:        "/nonexistent/editor",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize GlobalConfig to avoid nil pointer in page.Open()
			cfg, _ := config.ParseYamlBytes([]byte("editor: " + tt.editor))
			src.GlobalConfig = cfg
			defer func() { src.GlobalConfig = nil }() // Clean up after test

			logger := &mockLogger{}
			cmd := NewCreateCommand(tt.targetDir, tt.title, tt.editor, logger)

			err := cmd.Execute()

			if tt.expectedError {
				if err == nil {
					t.Error("Execute() expected error but got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Execute() error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("Execute() unexpected error: %v", err)
				}
				if len(logger.infoCalls) == 0 {
					t.Error("Execute() expected log call but got none")
				}
			}
		})
	}
}

func TestCommitMessageResolver_Resolve_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		configYaml  string
		defaultMsg  string
		args        []string
		nonFlagArgs int
		want        string
	}{
		{
			name:        "nil config uses default",
			configYaml:  "",
			defaultMsg:  "default message",
			args:        []string{"til", "-s"},
			nonFlagArgs: 0,
			want:        "default message",
		},
		{
			name:        "nonFlagArgs exceeds len(args)",
			configYaml:  "commitMessage: from config",
			defaultMsg:  "default",
			args:        []string{"til", "-s"},
			nonFlagArgs: 10, // exceeds len(args), clamped to len(args), msgOffset becomes 0
			want:        "til -s", // Returns all args joined when offset is 0
		},
		{
			name:        "negative offset handled",
			configYaml:  "commitMessage: from config",
			defaultMsg:  "default",
			args:        []string{"til"},
			nonFlagArgs: 5, // exceeds len(args), clamped, msgOffset becomes 0
			want:        "til", // Returns all args joined when offset is 0
		},
		{
			name:        "empty args",
			configYaml:  "",
			defaultMsg:  "default",
			args:        []string{},
			nonFlagArgs: 0,
			want:        "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *config.Config
			if tt.configYaml != "" {
				var err error
				cfg, err = config.ParseYamlBytes([]byte(tt.configYaml))
				if err != nil {
					t.Fatalf("failed to parse config: %v", err)
				}
			}

			resolver := &CommitMessageResolver{
				config:     cfg,
				defaultMsg: tt.defaultMsg,
			}

			got := resolver.Resolve(tt.args, tt.nonFlagArgs)

			if got != tt.want {
				t.Errorf("CommitMessageResolver.Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}
