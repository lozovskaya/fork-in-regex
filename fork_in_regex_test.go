package main

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		shouldFail bool
	}{
		{"# Header", 2, false},
		{"**Bold Text**", 3, false},
		{"*Italic Text*", 3, false},
		{"`Code Block`", 3, false},
		{"[Link](http://example.com)", 6, false},
		{"- List Item\n- Another Item", 4, false},
		{"1. Ordered List", 2, false},
		{"> Blockquote", 2, false},
		{"---", 1, false},
		{"This is plain text", 1, false},
	}

	for _, tt := range tests {
		output, err := tokenize(tt.input)
		if tt.shouldFail {
			if err == nil {
				t.Errorf("expected failure but got success for input: %q", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for input %q: %v", tt.input, err)
			}
			if len(output.Tokens) != tt.expected {
				t.Errorf("expected %d tokens, got %d for input: %q", tt.expected, len(output.Tokens), tt.input)
			}
		}
	}
}

func FuzzTokenize(f *testing.F) {
	sampleInputs := []string{
		"# Sample Header",
		"**bold text**",
		"*italic text*",
		"`inline code`",
		"[example](http://test.com)",
		"- List item",
		"1. Ordered item",
		"> Blockquote",
		"---",
		"",
	}

	for _, input := range sampleInputs {
		f.Add(input)
	}

	f.Fuzz(func(t *testing.T, input string) {
		_, _ = tokenize(input) // Ensure no panics occur
	})
}
