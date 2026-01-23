package loader

import (
	"fmt"
	"path/filepath"

	"github.com/senorprogrammer/til/pages"
)

// PageLoader loads page files from disk
type PageLoader struct {
	targetDir string
	extension string
}

// NewPageLoader creates a new PageLoader
func NewPageLoader(targetDir, extension string) *PageLoader {
	return &PageLoader{
		targetDir: targetDir,
		extension: extension,
	}
}

// Load reads page files from disk in reverse chronological order
func (l *PageLoader) Load() ([]*pages.Page, error) {
	pattern := fmt.Sprintf("%s/*.%s", l.targetDir, l.extension)

	filePaths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob page files: %w", err)
	}

	pageSet := make([]*pages.Page, 0, len(filePaths))

	// Read in reverse chronological order (newest first)
	for i := len(filePaths) - 1; i >= 0; i-- {
		page := pages.PageFromFilePath(filePaths[i])
		pageSet = append(pageSet, page)
	}

	return pageSet, nil
}
