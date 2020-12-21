package src

import "sort"

// TagMap is a map of tag name to Tag instance
type TagMap struct {
	Tags map[string][]*Tag
}

// NewTagMap creates and returns an instance of TagMap
func NewTagMap(pages []*Page) *TagMap {
	tm := &TagMap{
		Tags: make(map[string][]*Tag),
	}

	tm.BuildFromPages(pages)

	return tm
}

// Add adds a Tag instance to the map
func (tm *TagMap) Add(tag *Tag) {
	if !tag.IsValid() {
		return
	}

	tm.Tags[tag.Name] = append(tm.Tags[tag.Name], tag)
}

// BuildFromPages populates the tag map from a slice of Page instances
func (tm *TagMap) BuildFromPages(pages []*Page) {
	for _, page := range pages {
		for _, tag := range page.Tags() {
			tm.Add(tag)
		}
	}
}

// Get returns the tags for a given tag name
func (tm *TagMap) Get(name string) []*Tag {
	return tm.Tags[name]
}

// Len returns the number of tags in the map
func (tm *TagMap) Len() int {
	return len(tm.Tags)
}

// PagesFor returns a flattened slice of pages for a given tag name, sorted
// in reverse-chronological order
func (tm *TagMap) PagesFor(tagName string) []*Page {
	pages := []*Page{}
	tags := tm.Get(tagName)

	for _, tag := range tags {
		pages = append(pages, tag.Pages...)

		sort.Slice(pages, func(i, j int) bool {
			return pages[i].CreatedAt().After(pages[j].CreatedAt())
		})
	}

	return pages
}

// SortedTagNames returns the tag names in alphabetical order
func (tm *TagMap) SortedTagNames() []string {
	tagArr := make([]string, tm.Len())
	i := 0

	for tag := range tm.Tags {
		tagArr[i] = tag
		i++
	}

	sort.Strings(tagArr)

	return tagArr
}
