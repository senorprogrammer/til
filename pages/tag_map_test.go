package pages

import (
	"testing"
)

func TestTagMap_PagesFor(t *testing.T) {
	tests := []struct {
		name      string
		pageSet   []*Page
		tagName   string
		wantCount int
		wantOrder bool // true if pages should be in reverse chronological order
	}{
		{
			name: "single tag with one page",
			pageSet: []*Page{
				{Date: "2020-01-01T00:00:00Z", Title: "Page 1", TagsStr: "go"},
			},
			tagName:   "go",
			wantCount: 1,
			wantOrder: true,
		},
		{
			name: "single tag with multiple pages",
			pageSet: []*Page{
				{Date: "2020-01-01T00:00:00Z", Title: "Page 1", TagsStr: "go"},
				{Date: "2020-02-01T00:00:00Z", Title: "Page 2", TagsStr: "go"},
				{Date: "2020-03-01T00:00:00Z", Title: "Page 3", TagsStr: "go"},
			},
			tagName:   "go",
			wantCount: 3,
			wantOrder: true,
		},
		{
			name: "multiple tags, filter by one",
			pageSet: []*Page{
				{Date: "2020-01-01T00:00:00Z", Title: "Page 1", TagsStr: "go"},
				{Date: "2020-02-01T00:00:00Z", Title: "Page 2", TagsStr: "ada"},
				{Date: "2020-03-01T00:00:00Z", Title: "Page 3", TagsStr: "go"},
			},
			tagName:   "go",
			wantCount: 2,
			wantOrder: true,
		},
		{
			name: "non-existent tag",
			pageSet: []*Page{
				{Date: "2020-01-01T00:00:00Z", Title: "Page 1", TagsStr: "go"},
			},
			tagName:   "nonexistent",
			wantCount: 0,
			wantOrder: false,
		},
		{
			name:      "empty page set",
			pageSet:   []*Page{},
			tagName:   "go",
			wantCount: 0,
			wantOrder: false,
		},
		{
			name: "pages with multiple tags",
			pageSet: []*Page{
				{Date: "2020-01-01T00:00:00Z", Title: "Page 1", TagsStr: "go, test"},
				{Date: "2020-02-01T00:00:00Z", Title: "Page 2", TagsStr: "go"},
			},
			tagName:   "go",
			wantCount: 2,
			wantOrder: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagMap := NewTagMap(tt.pageSet)
			got := tagMap.PagesFor(tt.tagName)

			if len(got) != tt.wantCount {
				t.Errorf("TagMap.PagesFor(%q) count = %d, want %d", tt.tagName, len(got), tt.wantCount)
			}

			if tt.wantOrder && len(got) > 1 {
				// Verify reverse chronological order (newest first)
				for i := 0; i < len(got)-1; i++ {
					if got[i].CreatedAt().Before(got[i+1].CreatedAt()) {
						t.Errorf("TagMap.PagesFor() pages not in reverse chronological order: page %d (%v) is before page %d (%v)",
							i, got[i].CreatedAt(), i+1, got[i+1].CreatedAt())
					}
				}
			}
		})
	}
}

func TestTagMap_SortedTagNames(t *testing.T) {
	tests := []struct {
		name     string
		pageSet  []*Page
		want     []string
	}{
		{
			name: "single tag",
			pageSet: []*Page{
				{TagsStr: "go"},
			},
			want: []string{"go"},
		},
		{
			name: "multiple tags alphabetical",
			pageSet: []*Page{
				{TagsStr: "go, ada, lua"},
			},
			want: []string{"ada", "go", "lua"},
		},
		{
			name: "multiple tags from different pages",
			pageSet: []*Page{
				{TagsStr: "go"},
				{TagsStr: "ada"},
				{TagsStr: "lua"},
			},
			want: []string{"ada", "go", "lua"},
		},
		{
			name:     "empty page set",
			pageSet:  []*Page{},
			want:     []string{},
		},
		{
			name: "duplicate tags",
			pageSet: []*Page{
				{TagsStr: "go"},
				{TagsStr: "go"},
				{TagsStr: "ada"},
			},
			want: []string{"ada", "go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tagMap := NewTagMap(tt.pageSet)
			got := tagMap.SortedTagNames()

			if len(got) != len(tt.want) {
				t.Errorf("TagMap.SortedTagNames() length = %d, want %d", len(got), len(tt.want))
			}

			for i, want := range tt.want {
				if i < len(got) && got[i] != want {
					t.Errorf("TagMap.SortedTagNames()[%d] = %q, want %q", i, got[i], want)
				}
			}
		})
	}
}
