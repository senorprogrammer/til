package pages

import (
	"fmt"
	"strings"
)

// Tag represents a page tag (e.g.: linux, zombies)
type Tag struct {
	Name  string
	Pages []*Page
}

// NewTag creates and returns an instance of Tag
func NewTag(name string, page *Page) *Tag {
	tag := &Tag{
		Name:  strings.TrimSpace(name),
		Pages: []*Page{page},
	}

	return tag
}

// AddPage adds a page to the list of pages
func (tag *Tag) AddPage(page *Page) {
	tag.Pages = append(tag.Pages, page)
}

// IsValid returns true if this is a valid tag, false if it is not
func (tag *Tag) IsValid() bool {
	return tag.Name != ""
}

// Link returns a link string suitable for embedding in a Markdown page
func (tag *Tag) Link() string {
	if tag.Name == "" {
		return ""
	}

	return fmt.Sprintf(
		"[%s](%s)",
		tag.Name,
		fmt.Sprintf("./%s", tag.Name),
	)
}
