package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

/* -------------------- Page -------------------- */

func Test_Page_CreatedAt(t *testing.T) {
	page := &Page{Date: "2020-05-07T13:13:08-07:00"}

	actual := page.CreatedAt()

	assert.Equal(t, 2020, int(actual.Year()))
	assert.Equal(t, 5, int(actual.Month()))
	assert.Equal(t, 7, int(actual.Day()))
}

func Test_Page_CreatedMonth(t *testing.T) {
	page := &Page{Date: "2020-05-07T13:13:08-07:00"}

	actual := page.CreatedMonth()

	assert.Equal(t, 5, int(actual))
}

func Test_IsContentPage(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected bool
	}{
		{
			name:     "when is not content page",
			title:    "",
			expected: false,
		},
		{
			name:     "when is content page",
			title:    "test",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := &Page{Title: tt.title}

			actual := page.IsContentPage()

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_Page_Link(t *testing.T) {
	page := &Page{
		Date:     "2020-05-07T13:13:08-07:00",
		FilePath: "docs/zombies.md",
		Title:    "Zombies",
	}

	actual := page.Link()

	assert.Equal(t, "<code>May 07, 2020</code> [Zombies](zombies.md)", actual)
}

func Test_Page_PrettDate(t *testing.T) {
	page := &Page{Date: "2020-05-07T13:13:08-07:00"}

	actual := page.PrettyDate()

	assert.Equal(t, "May 07, 2020", actual)
}

/* -------------------- Tag -------------------- */

func Test_Tag_NewTag(t *testing.T) {
	actual := NewTag("ada", &Page{Title: "test"})

	assert.IsType(t, &Tag{}, actual)
	assert.Equal(t, "ada", actual.Name)
	assert.Equal(t, "test", actual.Pages[0].Title)
}

func Test_Tag_AddPage(t *testing.T) {
	tag := NewTag("ada", &Page{Title: "test"})
	tag.AddPage(&Page{Title: "zombies"})

	assert.Equal(t, 2, len(tag.Pages))
	assert.Equal(t, "zombies", tag.Pages[1].Title)
}

func Test_Tag_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected bool
	}{
		{
			name:     "when invalid",
			title:    "",
			expected: false,
		},
		{
			name:     "when valid",
			title:    "test",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := NewTag(tt.title, &Page{})

			actual := tag.IsValid()

			assert.Equal(t, tt.expected, actual)
		})
	}
}

/* -------------------- TagMap -------------------- */

func Test_NewTagMap(t *testing.T) {
	tests := []struct {
		name        string
		pages       []*Page
		expectedLen int
	}{
		{
			name:        "with no pages",
			pages:       []*Page{},
			expectedLen: 0,
		},
		{
			name: "with pages",
			pages: []*Page{
				{TagsStr: "go, ada"},
			},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := NewTagMap(tt.pages).Tags

			assert.Equal(t, tt.expectedLen, len(actual))
		})
	}
}

func Test_TagMap_Add(t *testing.T) {
	tests := []struct {
		name        string
		tag         *Tag
		expectedLen int
	}{
		{
			name:        "with an invalid tag",
			tag:         &Tag{},
			expectedLen: 0,
		},
		{
			name:        "with a new tag",
			tag:         &Tag{Name: "go"},
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tMap := NewTagMap([]*Page{})
			tMap.Add(tt.tag)

			actual := tMap.Tags

			assert.Equal(t, tt.expectedLen, len(actual))
		})
	}
}

func Test_TagMap_BuildFromPages(t *testing.T) {
	tests := []struct {
		name        string
		pages       []*Page
		expectedLen int
	}{
		{
			name:        "with no pages",
			pages:       []*Page{},
			expectedLen: 0,
		},
		{
			name: "with pages",
			pages: []*Page{
				{TagsStr: "go"},
				{TagsStr: "ada"},
			},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tMap := NewTagMap([]*Page{})
			tMap.BuildFromPages(tt.pages)

			actual := tMap.Tags

			assert.Equal(t, tt.expectedLen, len(actual))
		})
	}
}

func Test_TagMap_Get(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedLen int
	}{
		{
			name:        "with missing tag",
			input:       "ada",
			expectedLen: 0,
		},
		{
			name:        "with valid tag",
			input:       "go",
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		pages := []*Page{&Page{TagsStr: "go"}}
		tMap := NewTagMap(pages)

		actual := tMap.Get(tt.input)

		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedLen, len(actual))
		})
	}
}

func Test_TagMap_Len(t *testing.T) {
	tests := []struct {
		name        string
		page        *Page
		expectedLen int
	}{
		{
			name:        "with missing tag",
			page:        &Page{},
			expectedLen: 0,
		},
		{
			name:        "with valid tag",
			page:        &Page{TagsStr: "go"},
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		pages := []*Page{tt.page}
		tMap := NewTagMap(pages)

		actual := tMap.Len()

		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedLen, actual)
		})
	}
}

func Test_TagMap_SortedTagNames(t *testing.T) {
	pages := []*Page{&Page{TagsStr: "go, ada, lua"}}
	tMap := NewTagMap(pages)

	expected := []string{"ada", "go", "lua"}
	actual := tMap.SortedTagNames()

	assert.Equal(t, expected, actual)
}
