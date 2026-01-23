package pages

import (
	"testing"
)

func TestNewTag(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		page     *Page
		wantName string
		wantLen  int
	}{
		{
			name:     "valid tag",
			tagName:  "ada",
			page:     &Page{Title: "test"},
			wantName: "ada",
			wantLen:  1,
		},
		{
			name:     "tag with spaces",
			tagName:  "  go  ",
			page:     &Page{Title: "test"},
			wantName: "go", // trimmed
			wantLen:  1,
		},
		{
			name:     "empty tag name",
			tagName:  "",
			page:     &Page{Title: "test"},
			wantName: "",
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTag(tt.tagName, tt.page)

			if got == nil {
				t.Fatal("NewTag() returned nil")
			}

			if got.Name != tt.wantName {
				t.Errorf("NewTag() Name = %q, want %q", got.Name, tt.wantName)
			}

			if len(got.Pages) != tt.wantLen {
				t.Errorf("NewTag() Pages length = %d, want %d", len(got.Pages), tt.wantLen)
			}

			if len(got.Pages) > 0 && got.Pages[0] != tt.page {
				t.Error("NewTag() Pages[0] != input page")
			}
		})
	}
}

func TestTag_AddPage(t *testing.T) {
	tests := []struct {
		name      string
		initial   *Tag
		pagesToAdd []*Page
		wantLen   int
	}{
		{
			name:      "add single page",
			initial:   NewTag("go", &Page{Title: "page1"}),
			pagesToAdd: []*Page{{Title: "page2"}},
			wantLen:   2,
		},
		{
			name:      "add multiple pages",
			initial:   NewTag("go", &Page{Title: "page1"}),
			pagesToAdd: []*Page{{Title: "page2"}, {Title: "page3"}},
			wantLen:   3,
		},
		{
			name:      "add to empty tag",
			initial:   &Tag{Name: "test", Pages: []*Page{}},
			pagesToAdd: []*Page{{Title: "page1"}},
			wantLen:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, page := range tt.pagesToAdd {
				tt.initial.AddPage(page)
			}

			if len(tt.initial.Pages) != tt.wantLen {
				t.Errorf("Tag.AddPage() Pages length = %d, want %d", len(tt.initial.Pages), tt.wantLen)
			}
		})
	}
}

func TestTag_Link(t *testing.T) {
	tests := []struct {
		name     string
		tag      *Tag
		want     string
		wantEmpty bool
	}{
		{
			name:     "valid tag link",
			tag:      &Tag{Name: "go"},
			want:     "[go](./go)",
			wantEmpty: false,
		},
		{
			name:     "tag with special characters",
			tag:      &Tag{Name: "c++"},
			want:     "[c++](./c++)",
			wantEmpty: false,
		},
		{
			name:     "empty tag name",
			tag:      &Tag{Name: ""},
			want:     "",
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.tag.Link()

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("Tag.Link() = %q, want empty string", got)
				}
			} else {
				if got != tt.want {
					t.Errorf("Tag.Link() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}
