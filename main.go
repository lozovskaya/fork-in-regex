package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"github.com/lozovskaya/go-commonmark"
)

type TokenType int

func cleanText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")

	paragraphs := strings.Split(text, "\n\n")

	for i, p := range paragraphs {
		if !strings.HasPrefix(p, ">") && !strings.HasPrefix(p, "- ") &&
			!strings.HasPrefix(p, "* ") && !strings.HasPrefix(p, "```") {
			paragraphs[i] = strings.ReplaceAll(p, "\n", " ")
		}
	}

	return strings.TrimSpace(strings.Join(paragraphs, "\n\n"))
}

func tokenize(text string) (*RDXBlob, error) {
	text = cleanText(text)

	var tokens []Token
	remaining := text

	patterns := []struct {
		Type  string `json:"type"`
		Regex *regexp.Regexp
	}{
		{"BLOCKQUOTE", regexp.MustCompile(`^>\s*`)},
		{"BOLD", regexp.MustCompile(`^\*\*`)},
		{"ITALIC", regexp.MustCompile(`^\*`)},
		{"HEADING", regexp.MustCompile(`^#{1,6}\s+`)},
		{"LIST_ITEM", regexp.MustCompile(`^[-*+]\s+|^\d+\.\s+`)},
		{"LINK_START", regexp.MustCompile(`^\[`)},
		{"LINK_END", regexp.MustCompile(`^\]`)},
		{"URL_START", regexp.MustCompile(`^\(`)},
		{"URL_END", regexp.MustCompile(`^\)`)},
		{"CODE", regexp.MustCompile("^`")},
		{"HORIZONTAL_RULE", regexp.MustCompile(`^---+|^===+|^___+`)},
	}

	textPattern := regexp.MustCompile(`^[^>\*#\[\]\(\)\` + "`" + `\-=_]+`)

	for len(remaining) > 0 {
		matched := false

		for _, pattern := range patterns {
			if match := pattern.Regex.FindString(remaining); match != "" {
				tokens = append(tokens, Token{Type: pattern.Type, Value: match})
				remaining = remaining[len(match):]
				matched = true
				break
			}
		}

		if !matched {
			if match := textPattern.FindString(remaining); match != "" {
				tokens = append(tokens, Token{Type: "TEXT", Value: match})
				remaining = remaining[len(match):]
			} else {
				tokens = append(tokens, Token{Type: "TEXT", Value: string(remaining[0])})
				remaining = remaining[1:]
			}
		}
	}

	return &RDXBlob{Tokens: tokens}, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./fork-in-regex <filename>")
		return
	}

	filename := os.Args[1]

	file_extension := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if file_extension != "md" {
		fmt.Println("Error: The file must be a markdown file with .md extension.")
		return
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	blocks, refMap := commonmark.Parse(data)
	
	// Convert the parsed blocks and refMap into RDXBlob
    rdxBlob := RenderRDX(blocks, refMap)

	// rdxBlob, err := tokenize(string(data))
	// if err != nil {
	// 	fmt.Println("Error tokenizing input:", err)
	// 	return
	// }

	// Convert the RDXBlob to JSON
	output, err := json.MarshalIndent(rdxBlob, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Println(string(output))
}
