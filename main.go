package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
)

type Token struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type RDXBlob struct {
	Tokens []Token
}

var tokenPatterns = map[string]*regexp.Regexp{
	// Matches Markdown headers (ATX-style and setext-style)
	"header":       regexp.MustCompile(`(?m)^(#{1,6})\s*(.+?)\s*(#*)$|^(.+)\n([=-]+)$`),
	"bold":         regexp.MustCompile(`\*\*(.+?)\*\*`),  // todo: fix
	"italic":       regexp.MustCompile(`\*(.+?)\*`), // todo: fix
	"code":         regexp.MustCompile("`([^`]*)`"),
	"link":         regexp.MustCompile(`\[(.*?)\]\((.*?)\)`),
	"unordered":    regexp.MustCompile(`(?m)^\s*[-*+]\s+(.+)$`),
	"ordered":      regexp.MustCompile(`(?m)^\s*\d+\.\s+(.+)$`),
	"blockquote":   regexp.MustCompile(`(?m)^>\s+(.+)$`),
	"horizontal":   regexp.MustCompile(`(?m)^-{3,}$`),
}

// Tokenize splits a CommonMark file content into an RDXBlob
func tokenize(input string) (*RDXBlob, error) {
	if len(input) == 0 {
		return nil, errors.New("empty input")
	}

	var tokens []Token
	for typ, pattern := range tokenPatterns {
		matches := pattern.FindAllStringSubmatch(input, -1)
		for _, match := range matches {
			if typ == "header" {
				if len(match) > 2 {	
					tokens = append(tokens, Token{Type: typ, Value: match[2]})
				}
			} else if len(match) > 1 {
				tokens = append(tokens, Token{Type: typ, Value: match[1]})
			}
		}
	}

	if len(tokens) == 0 {
		return nil, errors.New("no valid tokens found")
	}

	return &RDXBlob{Tokens: tokens}, nil
}




func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./fork-in-regex <filename>")
		return
	}

	filename := os.Args[1]
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	rdxBlob, err := tokenize(string(data))
	if err != nil {
		fmt.Println("Error tokenizing input:", err)
		return
	}

	// Convert the RDXBlob to JSON
	output, err := json.MarshalIndent(rdxBlob, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Println(string(output))
}
