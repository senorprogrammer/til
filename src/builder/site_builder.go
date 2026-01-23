package builder

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/senorprogrammer/til/pages"
	"github.com/senorprogrammer/til/src"
)

// FileWriter defines the interface for writing files
type FileWriter interface {
	WriteFile(path string, content []byte, perm os.FileMode) error
}

// OSFileWriter implements FileWriter using os.WriteFile
type OSFileWriter struct{}

func (w *OSFileWriter) WriteFile(path string, content []byte, perm os.FileMode) error {
	return os.WriteFile(path, content, perm)
}

// DefaultFileWriter is a shared instance of OSFileWriter that can be reused
// since it's stateless. This reduces allocations and makes mocking easier.
var DefaultFileWriter FileWriter = &OSFileWriter{}

// Logger defines the interface for logging
type Logger interface {
	Info(msg string)
	Progress(msg string)
}

// SiteBuilder builds index and tag pages
type SiteBuilder struct {
	targetDir string
	writer    FileWriter
	logger    Logger
}

// BuildIndex creates the main index.md page
func (b *SiteBuilder) BuildIndex(pageSet []*pages.Page, tagMap *pages.TagMap) error {
	b.logger.Info("building index page")

	var content strings.Builder

	// Write the tag list into the top of the index
	tagLinks := []string{}
	for _, tagName := range tagMap.SortedTagNames() {
		tags := tagMap.Get(tagName)
		if len(tags) > 0 {
			tagLinks = append(tagLinks, tags[0].Link())
		}
	}

	content.WriteString(strings.Join(tagLinks, ", "))
	content.WriteString("\n")

	// Write the page list into the middle of the page
	content.WriteString(pagesToHTMLUnorderedList(pageSet))
	content.WriteString("\n")

	// Write the footer content into the bottom of the index
	content.WriteString("\n")
	content.WriteString(src.Footer())

	// Write the file to disk
	filePath := fmt.Sprintf("%s/index.%s", b.targetDir, pages.FileExtension)

	err := b.writer.WriteFile(filePath, []byte(content.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write index page: %w", err)
	}

	b.logger.Progress(filePath)
	return nil
}

// BuildTagPages creates tag pages with links to posts tagged with those names.
// Tag pages are built concurrently using goroutines for better performance.
// TagMap methods are read-only during concurrent access, which is safe for concurrent use.
func (b *SiteBuilder) BuildTagPages(tagMap *pages.TagMap) error {
	b.logger.Info("building tag pages")

	tagNames := tagMap.SortedTagNames()
	if len(tagNames) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for _, tagName := range tagNames {
		wg.Add(1)

		go func(tagName string) {
			defer wg.Done()

			var content strings.Builder
			content.WriteString(fmt.Sprintf("## %s\n\n", tagName))

			// Write the page list into the middle of the page
			// TagMap.PagesFor is read-only and safe for concurrent access
			content.WriteString(pagesToHTMLUnorderedList(tagMap.PagesFor(tagName)))

			// Write the footer content into the bottom of the page
			content.WriteString("\n")
			content.WriteString(src.Footer())

			// Write the file to disk
			filePath := fmt.Sprintf("%s/%s.%s", b.targetDir, tagName, pages.FileExtension)

			err := b.writer.WriteFile(filePath, []byte(content.String()), 0644)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to write tag page %s: %w", tagName, err))
				mu.Unlock()
				return
			}

			b.logger.Progress(filePath)
		}(tagName)
	}

	wg.Wait()

	// Return the first error if any occurred, or combine all errors for better diagnostics
	if len(errors) > 0 {
		if len(errors) == 1 {
			return errors[0]
		}
		// Multiple errors occurred - return a combined error message
		errMsgs := make([]string, len(errors))
		for i, err := range errors {
			errMsgs[i] = err.Error()
		}
		return fmt.Errorf("multiple errors building tag pages: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}

// NewSiteBuilder creates a new SiteBuilder
func NewSiteBuilder(targetDir string, writer FileWriter, logger Logger) *SiteBuilder {
	return &SiteBuilder{
		targetDir: targetDir,
		writer:    writer,
		logger:    logger,
	}
}

// pagesToHTMLUnorderedList creates the unordered list of page links
func pagesToHTMLUnorderedList(pageSet []*pages.Page) string {
	var content strings.Builder
	var prevPage *pages.Page

	for _, page := range pageSet {
		if !page.IsContentPage() {
			continue
		}

		// This breaks the page list up by month
		if prevPage != nil && prevPage.CreatedMonth() != page.CreatedMonth() {
			content.WriteString("\n")
		}

		content.WriteString(fmt.Sprintf("* %s\n", page.Link()))
		prevPage = page
	}

	return content.String()
}
