package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

type Token struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

var tokenPatterns = map[string]string{
	"keyword":     `\b(int|float|char|return|if|else|while|for|switch)\b`,
	"identifier":  `\b[a-zA-Z_][a-zA-Z0-9_]*\b`,
	"number":      `\b\d+(\.\d+)?\b`,
	"operator":    `[+\-*/=<>!&|]+`,
	"punctuation": `[{}()\[\],;]`,
	"string":      `"(\\.|[^"])*"`,
	"comment":     `//.*|/\*.*?\*/`,
}

func tokenize(input string) []Token {
	var tokens []Token
	for tokenType, pattern := range tokenPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(input, -1)
		for _, match := range matches {
			tokens = append(tokens, Token{Type: tokenType, Value: match})
		}
	}
	return tokens
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

	tokens := tokenize(string(data))
	output, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Println(string(output))

}
