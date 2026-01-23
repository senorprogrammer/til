package utils

import (
	"testing"
)

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single word lowercase",
			input:    "hello",
			expected: "Hello",
		},
		{
			name:     "single word uppercase",
			input:    "HELLO",
			expected: "Hello",
		},
		{
			name:     "single word mixed case",
			input:    "hElLo",
			expected: "Hello",
		},
		{
			name:     "multiple words lowercase",
			input:    "hello world",
			expected: "Hello World",
		},
		{
			name:     "multiple words uppercase",
			input:    "HELLO WORLD",
			expected: "Hello World",
		},
		{
			name:     "multiple words mixed case",
			input:    "hElLo WoRlD",
			expected: "Hello World",
		},
		{
			name:     "three words",
			input:    "this is a test",
			expected: "This Is A Test",
		},
		{
			name:     "multiple spaces between words",
			input:    "hello   world",
			expected: "Hello   World",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  hello world  ",
			expected: "  Hello World  ",
		},
		{
			name:     "word with numbers",
			input:    "hello 123 world",
			expected: "Hello 123 World",
		},
		{
			name:     "single character",
			input:    "a",
			expected: "A",
		},
		{
			name:     "word with special characters",
			input:    "hello-world test",
			expected: "Hello-world Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := TitleCase(tt.input)
			if actual != tt.expected {
				t.Errorf("TitleCase(%q) = %q, expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func TestParseTitle(t *testing.T) {
	tests := []struct {
		name       string
		targetFlag string
		args       []string
		expected   string
	}{
		{
			name:       "with target flag and title",
			targetFlag: "a",
			args:       []string{"til", "-t", "a", "my", "test", "title"},
			expected:   "My Test Title",
		},
		{
			name:       "with target flag single word",
			targetFlag: "b",
			args:       []string{"til", "-t", "b", "hello"},
			expected:   "Hello",
		},
		{
			name:       "without target flag and title",
			targetFlag: "",
			args:       []string{"til", "my", "test", "title"},
			expected:   "My Test Title",
		},
		{
			name:       "without target flag single word",
			targetFlag: "",
			args:       []string{"til", "hello"},
			expected:   "Hello",
		},
		{
			name:       "with target flag empty args",
			targetFlag: "a",
			args:       []string{"til", "-t", "a"},
			expected:   "",
		},
		{
			name:       "without target flag empty args",
			targetFlag: "",
			args:       []string{"til"},
			expected:   "",
		},
		{
			name:       "with target flag args shorter than offset",
			targetFlag: "a",
			args:       []string{"til", "-t"},
			expected:   "",
		},
		{
			name:       "without target flag args shorter than offset",
			targetFlag: "",
			args:       []string{},
			expected:   "",
		},
		{
			name:       "with target flag exactly at offset",
			targetFlag: "a",
			args:       []string{"til", "-t", "a"},
			expected:   "",
		},
		{
			name:       "without target flag exactly at offset",
			targetFlag: "",
			args:       []string{"til"},
			expected:   "",
		},
		{
			name:       "with target flag multiple words",
			targetFlag: "b",
			args:       []string{"til", "-target", "b", "this", "is", "a", "long", "title"},
			expected:   "This Is A Long Title",
		},
		{
			name:       "without target flag multiple words",
			targetFlag: "",
			args:       []string{"til", "another", "long", "title", "here"},
			expected:   "Another Long Title Here",
		},
		{
			name:       "with target flag mixed case title",
			targetFlag: "a",
			args:       []string{"til", "-t", "a", "hElLo", "WoRlD"},
			expected:   "Hello World",
		},
		{
			name:       "without target flag mixed case title",
			targetFlag: "",
			args:       []string{"til", "tEsT", "TiTlE"},
			expected:   "Test Title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ParseTitle(tt.targetFlag, tt.args)
			if actual != tt.expected {
				t.Errorf("ParseTitle(%q, %v) = %q, expected %q", tt.targetFlag, tt.args, actual, tt.expected)
			}
		})
	}
}
