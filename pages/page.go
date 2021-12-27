package pages

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ericaro/frontmatter"
	"github.com/senorprogrammer/til/src"
)

const (
	// A custom datetime format that plays nicely with GitHub Pages filename restrictions
	ghFriendlyDateFormat = "2006-01-02T15-04-05"

	// FileExtension defines the extension to write on the generated file
	FileExtension = "md"
)

// Page represents a TIL page
type Page struct {
	Content  string `fm:"content" yaml:"-"`
	Date     string `yaml:"date"`
	FilePath string `yaml:"filepath"`
	TagsStr  string `yaml:"tags"`
	Title    string `yaml:"title"`
}

// NewPage creates and returns an instance of page
func NewPage(title string, targetDir string) *Page {
	date := time.Now()

	page := &Page{
		Date: date.Format(time.RFC3339),
		FilePath: fmt.Sprintf(
			"%s/%s-%s.%s",
			targetDir,
			date.Format(ghFriendlyDateFormat),
			strings.ReplaceAll(strings.ToLower(title), " ", "-"),
			FileExtension,
		),
		Title: title,
	}

	page.Save()

	return page
}

// PageFromFilePath creates and returns a Page instance from a file path
func PageFromFilePath(filePath string) *Page {
	page := new(Page)

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		src.Defeat(err)
	}

	err = frontmatter.Unmarshal(data, page)
	if err != nil {
		src.Defeat(err)
	}

	page.FilePath = filePath

	return page
}

// CreatedAt returns a time instance representing when the page was created
func (page *Page) CreatedAt() time.Time {
	date, err := time.Parse(time.RFC3339, page.Date)
	if err != nil {
		return time.Time{}
	}

	return date
}

// CreatedMonth returns the month the page was created
func (page *Page) CreatedMonth() time.Month {
	if page.CreatedAt().IsZero() {
		return 0
	}

	return page.CreatedAt().Month()
}

// FrontMatter returns the front-matter of the page
func (page *Page) FrontMatter() string {
	return fmt.Sprintf(
		"---\nlayout: default\ndate: %s\ntitle: %s\ntags: %s\n---\n\n",
		page.Date,
		page.Title,
		page.TagsStr,
	)
}

// IsContentPage returns true if the page is a valid entry page, false if it is not
func (page *Page) IsContentPage() bool {
	return page.Title != ""
}

// Link returns a link string suitable for embedding in a Markdown page
func (page *Page) Link() string {
	return fmt.Sprintf(
		"<code>%s</code> [%s](%s)",
		page.PrettyDate(),
		page.Title,
		filepath.Base(page.FilePath),
	)
}

// Open tll the OS to open the newly-created page in the editor (as specified in the config)
// If there's no editor explicitly defined by the user, tell the OS to try and open it
func (page *Page) Open(defaultEditor string) error {
	editor := src.GlobalConfig.UString("editor", defaultEditor)
	if editor == "" {
		editor = defaultEditor
	}

	cmd := exec.Command(editor, page.FilePath)
	err := cmd.Run()

	return err
}

// PrettyDate returns a human-friendly representation of the CreatedAt date
func (page *Page) PrettyDate() string {
	return page.CreatedAt().Format("Jan 02, 2006")
}

// Save writes the content of the page to file
func (page *Page) Save() {
	pageSrc := page.FrontMatter()
	pageSrc += fmt.Sprintf("# %s\n\n", page.Title)

	err := ioutil.WriteFile(page.FilePath, []byte(pageSrc), 0644)
	if err != nil {
		src.Defeat(err)
	}
}

// Tags returns a slice of tags assigned to this page
func (page *Page) Tags() []*Tag {
	tags := []*Tag{}

	names := strings.Split(page.TagsStr, ",")
	for _, name := range names {
		tags = append(tags, NewTag(name, page))
	}

	return tags
}
