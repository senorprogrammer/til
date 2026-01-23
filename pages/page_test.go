package pages

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/olebedev/config"
	"github.com/senorprogrammer/til/src"
)

func TestNewPage(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		targetDir string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "valid page creation",
			title:     "Test Page",
			targetDir: t.TempDir(),
			wantTitle: "Test Page",
			wantErr:   false,
		},
		{
			name:      "page with special characters in title",
			title:     "Hello-World Test",
			targetDir: t.TempDir(),
			wantTitle: "Hello-World Test",
			wantErr:   false,
		},
		{
			name:      "page with numbers in title",
			title:     "Test 123 Page",
			targetDir: t.TempDir(),
			wantTitle: "Test 123 Page",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := NewPage(tt.title, tt.targetDir)

			if page == nil {
				t.Fatal("NewPage() returned nil")
			}

			if page.Title != tt.wantTitle {
				t.Errorf("NewPage() title = %q, want %q", page.Title, tt.wantTitle)
			}

			if page.FilePath == "" {
				t.Error("NewPage() FilePath is empty")
			}

			// Verify file was created
			if _, err := os.Stat(page.FilePath); os.IsNotExist(err) {
				t.Errorf("NewPage() file was not created at %q", page.FilePath)
			}
		})
	}
}

func TestPageFromFilePath(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		fileName    string
		wantTitle   string
		wantDate    string
		wantTags    string
		wantErr     bool
	}{
		{
			name: "valid page file",
			fileContent: `---
date: 2020-05-07T13:13:08-07:00
title: Test Page
tags: go, test
---

# Test Page

Content here.
`,
			fileName:  "test.md",
			wantTitle: "Test Page",
			wantDate:  "2020-05-07T13:13:08-07:00",
			wantTags:  "go, test",
			wantErr:   false,
		},
		{
			name: "page with empty tags",
			fileContent: `---
date: 2020-01-01T00:00:00Z
title: Another Page
tags: 
---

# Another Page

Content.
`,
			fileName:  "another.md",
			wantTitle: "Another Page",
			wantDate:  "2020-01-01T00:00:00Z",
			wantTags:  "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, tt.fileName)

			if err := os.WriteFile(filePath, []byte(tt.fileContent), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			// Note: PageFromFilePath calls src.Defeat on error which exits
			// In a real test environment, we'd need to handle this differently
			// For now, we'll test the happy path
			page := PageFromFilePath(filePath)

			if page == nil {
				t.Fatal("PageFromFilePath() returned nil")
			}

			if page.Title != tt.wantTitle {
				t.Errorf("PageFromFilePath() title = %q, want %q", page.Title, tt.wantTitle)
			}

			if page.Date != tt.wantDate {
				t.Errorf("PageFromFilePath() date = %q, want %q", page.Date, tt.wantDate)
			}

			if page.TagsStr != tt.wantTags {
				t.Errorf("PageFromFilePath() tags = %q, want %q", page.TagsStr, tt.wantTags)
			}
		})
	}
}

func TestPage_CreatedAt(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		wantYear int
		wantMonth int
		wantDay  int
		wantZero bool
	}{
		{
			name:     "valid RFC3339 date",
			date:     "2020-05-07T13:13:08-07:00",
			wantYear: 2020,
			wantMonth: 5,
			wantDay:  7,
			wantZero: false,
		},
		{
			name:     "valid RFC3339 date UTC",
			date:     "2020-01-01T00:00:00Z",
			wantYear: 2020,
			wantMonth: 1,
			wantDay:  1,
			wantZero: false,
		},
		{
			name:     "invalid date format",
			date:     "invalid-date",
			wantYear: 0,
			wantMonth: 0,
			wantDay:  0,
			wantZero: true,
		},
		{
			name:     "empty date",
			date:     "",
			wantYear: 0,
			wantMonth: 0,
			wantDay:  0,
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{Date: tt.date}
			got := page.CreatedAt()

			if got.IsZero() != tt.wantZero {
				t.Errorf("Page.CreatedAt() IsZero() = %v, want %v", got.IsZero(), tt.wantZero)
			}

			if !tt.wantZero {
				if got.Year() != tt.wantYear {
					t.Errorf("Page.CreatedAt() Year() = %d, want %d", got.Year(), tt.wantYear)
				}
				if int(got.Month()) != tt.wantMonth {
					t.Errorf("Page.CreatedAt() Month() = %d, want %d", int(got.Month()), tt.wantMonth)
				}
				if got.Day() != tt.wantDay {
					t.Errorf("Page.CreatedAt() Day() = %d, want %d", got.Day(), tt.wantDay)
				}
			}
		})
	}
}

func TestPage_CreatedMonth(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		want     time.Month
		wantZero bool
	}{
		{
			name:     "valid date in May",
			date:     "2020-05-07T13:13:08-07:00",
			want:     time.May,
			wantZero: false,
		},
		{
			name:     "valid date in January",
			date:     "2020-01-01T00:00:00Z",
			want:     time.January,
			wantZero: false,
		},
		{
			name:     "invalid date",
			date:     "invalid-date",
			want:     0,
			wantZero: true,
		},
		{
			name:     "empty date",
			date:     "",
			want:     0,
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{Date: tt.date}
			got := page.CreatedMonth()

			if got == 0 && !tt.wantZero {
				t.Errorf("Page.CreatedMonth() = 0, want %v", tt.want)
			}

			if !tt.wantZero && got != tt.want {
				t.Errorf("Page.CreatedMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPage_FrontMatter(t *testing.T) {
	tests := []struct {
		name     string
		page     *Page
		wantCont []string // substrings that should be in the front matter
	}{
		{
			name: "complete front matter",
			page: &Page{
				Date:    "2020-05-07T13:13:08-07:00",
				Title:   "Test Page",
				TagsStr: "go, test",
			},
			wantCont: []string{"layout: default", "date: 2020-05-07T13:13:08-07:00", "title: Test Page", "tags: go, test"},
		},
		{
			name: "front matter with empty tags",
			page: &Page{
				Date:    "2020-01-01T00:00:00Z",
				Title:   "Another Page",
				TagsStr: "",
			},
			wantCont: []string{"layout: default", "date: 2020-01-01T00:00:00Z", "title: Another Page", "tags: "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.page.FrontMatter()

			for _, want := range tt.wantCont {
				if !contains(got, want) {
					t.Errorf("Page.FrontMatter() missing substring %q, got %q", want, got)
				}
			}
		})
	}
}

func TestPage_Link(t *testing.T) {
	tests := []struct {
		name     string
		page     *Page
		wantCont []string // substrings that should be in the link
	}{
		{
			name: "valid link",
			page: &Page{
				Date:     "2020-05-07T13:13:08-07:00",
				FilePath: "docs/zombies.md",
				Title:    "Zombies",
			},
			wantCont: []string{"<code>", "May 07, 2020", "[Zombies]", "(zombies.md)"},
		},
		{
			name: "link with different date",
			page: &Page{
				Date:     "2020-01-01T00:00:00Z",
				FilePath: "docs/test.md",
				Title:    "Test Page",
			},
			wantCont: []string{"<code>", "Jan 01, 2020", "[Test Page]", "(test.md)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.page.Link()

			for _, want := range tt.wantCont {
				if !contains(got, want) {
					t.Errorf("Page.Link() missing substring %q, got %q", want, got)
				}
			}
		})
	}
}

func TestPage_PrettyDate(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		want     string
		wantZero bool
	}{
		{
			name:     "valid date",
			date:     "2020-05-07T13:13:08-07:00",
			want:     "May 07, 2020",
			wantZero: false,
		},
		{
			name:     "valid date January",
			date:     "2020-01-01T00:00:00Z",
			want:     "Jan 01, 2020",
			wantZero: false,
		},
		{
			name:     "invalid date",
			date:     "invalid-date",
			want:     "Jan 01, 0001", // zero time formatted
			wantZero: true,
		},
		{
			name:     "empty date",
			date:     "",
			want:     "Jan 01, 0001", // zero time formatted
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{Date: tt.date}
			got := page.PrettyDate()

			if got != tt.want {
				t.Errorf("Page.PrettyDate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPage_Tags(t *testing.T) {
	tests := []struct {
		name      string
		tagsStr   string
		wantCount int
		wantNames []string
	}{
		{
			name:      "single tag",
			tagsStr:   "go",
			wantCount: 1,
			wantNames: []string{"go"},
		},
		{
			name:      "multiple tags",
			tagsStr:   "go, ada, lua",
			wantCount: 3,
			wantNames: []string{"go", "ada", "lua"},
		},
		{
			name:      "tags with spaces",
			tagsStr:   "go, ada, lua",
			wantCount: 3,
			wantNames: []string{"go", "ada", "lua"},
		},
		{
			name:      "empty tags",
			tagsStr:   "",
			wantCount: 1, // empty string splits to one empty element
			wantNames: []string{""},
		},
		{
			name:      "tags with extra spaces",
			tagsStr:   " go , ada , lua ",
			wantCount: 3,
			wantNames: []string{"go", "ada", "lua"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{TagsStr: tt.tagsStr}
			got := page.Tags()

			if len(got) != tt.wantCount {
				t.Errorf("Page.Tags() count = %d, want %d", len(got), tt.wantCount)
			}

			for i, tag := range got {
				if i < len(tt.wantNames) {
					// Tags are trimmed, so check trimmed name
					wantTrimmed := tt.wantNames[i]
					if wantTrimmed != "" && tag.Name != wantTrimmed {
						t.Errorf("Page.Tags()[%d].Name = %q, want %q", i, tag.Name, wantTrimmed)
					}
				}
			}
		})
	}
}

func TestPage_Open(t *testing.T) {
	tests := []struct {
		name        string
		defaultEditor string
		configEditor  string
		wantErr     bool
	}{
		{
			name:        "valid editor from default",
			defaultEditor: "echo",
			configEditor:  "",
			wantErr:     false,
		},
		{
			name:        "valid editor from config",
			defaultEditor: "echo",
			configEditor:  "echo",
			wantErr:     false,
		},
		{
			name:        "invalid editor",
			defaultEditor: "/nonexistent/editor",
			configEditor:  "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up GlobalConfig if configEditor is provided
			if tt.configEditor != "" {
				cfg, _ := config.ParseYamlBytes([]byte("editor: " + tt.configEditor))
				src.GlobalConfig = cfg
				defer func() { src.GlobalConfig = nil }()
			} else {
				src.GlobalConfig = nil
			}

			tmpDir := t.TempDir()
			page := NewPage("Test Page", tmpDir)

			err := page.Open(tt.defaultEditor)

			if (err != nil) != tt.wantErr {
				t.Errorf("Page.Open() error = %v, wantErr %v", err, tt.wantErr)
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
