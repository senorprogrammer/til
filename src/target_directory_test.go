package src

import (
	"os"
	"testing"

	"github.com/olebedev/config"
)

func TestGetTargetDir(t *testing.T) {
	tests := []struct {
		name          string
		configYaml    string
		targetDirFlag string
		withDocsDir   bool
		wantErr       bool
		wantContains  string
	}{
		{
			name: "single directory without flag",
			configYaml: `targetDirectories:
  a: ~/Documents/blog`,
			targetDirFlag: "",
			withDocsDir:  true,
			wantErr:      false,
			wantContains: "blog/docs",
		},
		{
			name: "single directory with docs",
			configYaml: `targetDirectories:
  a: ~/Documents/blog`,
			targetDirFlag: "",
			withDocsDir:  false,
			wantErr:      false,
			wantContains: "blog",
		},
		{
			name: "multiple directories with flag",
			configYaml: `targetDirectories:
  a: ~/Documents/blog
  b: ~/Documents/notes`,
			targetDirFlag: "a",
			withDocsDir:  true,
			wantErr:      false,
			wantContains: "blog/docs",
		},
		{
			name: "multiple directories without flag",
			configYaml: `targetDirectories:
  a: ~/Documents/blog
  b: ~/Documents/notes`,
			targetDirFlag: "",
			withDocsDir:  true,
			wantErr:      true,
		},
		{
			name: "absolute path without tilde",
			configYaml: `targetDirectories:
  a: /tmp/test`,
			targetDirFlag: "",
			withDocsDir:  true,
			wantErr:      false,
			wantContains: "/tmp/test/docs",
		},
		{
			name: "nil config",
			configYaml: "",
			targetDirFlag: "",
			withDocsDir:  true,
			wantErr:      true,
		},
		{
			name: "invalid key with multiple directories",
			configYaml: `targetDirectories:
  a: ~/Documents/blog
  b: ~/Documents/notes`,
			targetDirFlag: "nonexistent",
			withDocsDir:  true,
			wantErr:      true,
		},
		{
			name: "non-string value in config",
			configYaml: `targetDirectories:
  a: 123`,
			targetDirFlag: "",
			withDocsDir:  true,
			wantErr:      true,
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

			got, err := GetTargetDir(cfg, tt.targetDirFlag, tt.withDocsDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTargetDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.wantContains != "" && !contains(got, tt.wantContains) {
					t.Errorf("GetTargetDir() = %q, want to contain %q", got, tt.wantContains)
				}
			}
		})
	}
}

func TestGetConfigFilePath(t *testing.T) {
	tests := []struct {
		name         string
		xdgConfigHome string
		wantContains string
		wantErr      bool
	}{
		{
			name:         "default XDG config",
			xdgConfigHome: "",
			wantContains: "config.yml",
			wantErr:      false,
		},
		{
			name:         "custom XDG_CONFIG_HOME",
			xdgConfigHome: "/tmp/test",
			wantContains: "config.yml",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.xdgConfigHome != "" {
				os.Setenv("XDG_CONFIG_HOME", tt.xdgConfigHome)
				defer os.Unsetenv("XDG_CONFIG_HOME")
			}

			got, err := GetConfigFilePath()

			if (err != nil) != tt.wantErr {
				t.Errorf("GetConfigFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if !contains(got, tt.wantContains) {
					t.Errorf("GetConfigFilePath() = %q, want to contain %q", got, tt.wantContains)
				}
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
