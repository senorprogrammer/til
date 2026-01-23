package utils

import (
	"strings"
	"unicode"
)

// TitleCase converts a string to title case (capitalizes first letter of each word)
func TitleCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	prevSpace := true

	for _, r := range s {
		if unicode.IsSpace(r) {
			result.WriteRune(r)
			prevSpace = true
		} else if prevSpace {
			result.WriteRune(unicode.ToTitle(r))
			prevSpace = false
		} else {
			result.WriteRune(unicode.ToLower(r))
			prevSpace = false
		}
	}

	return result.String()
}

// ParseTitle extracts and formats the title from command line arguments
func ParseTitle(targetFlag string, args []string) string {
	const (
		titleOffsetWithTarget    = 3 // program name + target flag + target value
		titleOffsetWithoutTarget = 1 // program name
	)

	titleOffset := titleOffsetWithTarget
	if targetFlag == "" {
		titleOffset = titleOffsetWithoutTarget
	}

	// Bounds check: if args doesn't have enough elements, return empty string
	if len(args) <= titleOffset {
		return ""
	}

	return TitleCase(strings.Join(args[titleOffset:], " "))
}
