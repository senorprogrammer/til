package main

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/olebedev/config"
	"github.com/senorprogrammer/til/src"
	"github.com/stretchr/testify/assert"
)

func Test_determineCommitMessage(t *testing.T) {
	tests := []struct {
		name       string
		cfgMessage string
		args       []string
		expected   string
	}{
		{
			name:       "passed in via -s",
			cfgMessage: "from the config",
			args:       []string{"test", "-t", "b", "-s", "this", "is", "test"},
			expected:   "this is test",
		},
		{
			name:       "from config file",
			cfgMessage: "from the config",
			args:       []string{"test", "-t", "b", "-s"},
			expected:   "from the config",
		},
		{
			name:       "from default const",
			cfgMessage: "",
			args:       []string{"test", "-t", "b", "-s"},
			expected:   "build, save, push",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			flag.Parse()

			msg := fmt.Sprintf("commitMessage: %s", tt.cfgMessage)
			cfg, _ := config.ParseYamlBytes([]byte(msg))

			actual := determineCommitMessage(cfg, os.Args)

			assert.Equal(t, tt.expected, actual)
		})
	}
}

/* -------------------- Configuration -------------------- */

func Test_getConfigPath(t *testing.T) {
	actual, err := src.GetConfigFilePath()

	assert.Contains(t, actual, "config.yml")
	assert.NoError(t, err)
}

/* -------------------- More Helper Functions -------------------- */

func Test_Colour(t *testing.T) {
	x := src.Colour("yo%soy")
	actual := x("cat")

	assert.Equal(t, "yocatoy", actual)
}

/* -------------------- Page -------------------- */

func Test_Page_CreatedAt(t *testing.T) {
	page := &src.Page{Date: "2020-05-07T13:13:08-07:00"}

	actual := page.CreatedAt()

	assert.Equal(t, 2020, actual.Year())
	assert.Equal(t, 5, int(actual.Month()))
	assert.Equal(t, 7, actual.Day())
}

func Test_Page_CreatedMonth(t *testing.T) {
	page := &src.Page{Date: "2020-05-07T13:13:08-07:00"}

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
			page := &src.Page{Title: tt.title}

			actual := page.IsContentPage()

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_Page_Link(t *testing.T) {
	page := &src.Page{
		Date:     "2020-05-07T13:13:08-07:00",
		FilePath: "docs/zombies.md",
		Title:    "Zombies",
	}

	actual := page.Link()

	assert.Equal(t, "<code>May 07, 2020</code> [Zombies](zombies.md)", actual)
}

func Test_Page_PrettDate(t *testing.T) {
	page := &src.Page{Date: "2020-05-07T13:13:08-07:00"}

	actual := page.PrettyDate()

	assert.Equal(t, "May 07, 2020", actual)
}

/* -------------------- Tag -------------------- */

func Test_Tag_NewTag(t *testing.T) {
	actual := src.NewTag("ada", &src.Page{Title: "test"})

	assert.IsType(t, &src.Tag{}, actual)
	assert.Equal(t, "ada", actual.Name)
	assert.Equal(t, "test", actual.Pages[0].Title)
}

func Test_Tag_AddPage(t *testing.T) {
	tag := src.NewTag("ada", &src.Page{Title: "test"})
	tag.AddPage(&src.Page{Title: "zombies"})

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
			tag := src.NewTag(tt.title, &src.Page{})

			actual := tag.IsValid()

			assert.Equal(t, tt.expected, actual)
		})
	}
}

/* -------------------- TagMap -------------------- */

func Test_NewTagMap(t *testing.T) {
	tests := []struct {
		name        string
		pages       []*src.Page
		expectedLen int
	}{
		{
			name:        "with no pages",
			pages:       []*src.Page{},
			expectedLen: 0,
		},
		{
			name: "with pages",
			pages: []*src.Page{
				{TagsStr: "go, ada"},
			},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := src.NewTagMap(tt.pages).Tags

			assert.Equal(t, tt.expectedLen, len(actual))
		})
	}
}

func Test_TagMap_Add(t *testing.T) {
	tests := []struct {
		name        string
		tag         *src.Tag
		expectedLen int
	}{
		{
			name:        "with an invalid tag",
			tag:         &src.Tag{},
			expectedLen: 0,
		},
		{
			name:        "with a new tag",
			tag:         &src.Tag{Name: "go"},
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tMap := src.NewTagMap([]*src.Page{})
			tMap.Add(tt.tag)

			actual := tMap.Tags

			assert.Equal(t, tt.expectedLen, len(actual))
		})
	}
}

func Test_TagMap_BuildFromPages(t *testing.T) {
	tests := []struct {
		name        string
		pages       []*src.Page
		expectedLen int
	}{
		{
			name:        "with no pages",
			pages:       []*src.Page{},
			expectedLen: 0,
		},
		{
			name: "with pages",
			pages: []*src.Page{
				{TagsStr: "go"},
				{TagsStr: "ada"},
			},
			expectedLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tMap := src.NewTagMap([]*src.Page{})
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
		pages := []*src.Page{{TagsStr: "go"}}
		tMap := src.NewTagMap(pages)

		actual := tMap.Get(tt.input)

		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedLen, len(actual))
		})
	}
}

func Test_TagMap_Len(t *testing.T) {
	tests := []struct {
		name        string
		page        *src.Page
		expectedLen int
	}{
		{
			name:        "with missing tag",
			page:        &src.Page{},
			expectedLen: 0,
		},
		{
			name:        "with valid tag",
			page:        &src.Page{TagsStr: "go"},
			expectedLen: 1,
		},
	}

	for _, tt := range tests {
		pages := []*src.Page{tt.page}
		tMap := src.NewTagMap(pages)

		actual := tMap.Len()

		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedLen, actual)
		})
	}
}

func Test_TagMap_SortedTagNames(t *testing.T) {
	pages := []*src.Page{{TagsStr: "go, ada, lua"}}
	tMap := src.NewTagMap(pages)

	expected := []string{"ada", "go", "lua"}
	actual := tMap.SortedTagNames()

	assert.Equal(t, expected, actual)
}
