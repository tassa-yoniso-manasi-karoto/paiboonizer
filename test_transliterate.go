package paiboonizer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

func testTransliterate() {
	// Open input file
	input, err := os.Open("test.txt")
	if err != nil {
		panic(err)
	}
	defer input.Close()

	// Create output file
	output, err := os.Create("test.romanized.txt")
	if err != nil {
		panic(err)
	}
	defer output.Close()

	scanner := bufio.NewScanner(input)
	writer := bufio.NewWriter(output)
	defer writer.Flush()

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		// Skip BOM if present
		if lineNum == 1 && strings.HasPrefix(line, "\ufeff") {
			line = strings.TrimPrefix(line, "\ufeff")
		}
		
		// Skip empty lines or preserve them
		if line == "" {
			writer.WriteString("\n")
			continue
		}
		
		// Check if line starts with # (header) or is not Thai - preserve as is
		if strings.HasPrefix(line, "#") || !containsThai(line) {
			writer.WriteString(line + "\n")
			continue
		}
		
		// Transliterate Thai text
		result := ""
		if globalManager != nil && globalManager.nlpManager != nil {
			ctx := context.Background()
			tokens, err := globalManager.nlpManager.Tokenize(ctx, line)
			if err == nil && tokens != nil && len(tokens.Raw) > 0 {
				// Tokenize and transliterate each word
				results := []string{}
				for _, token := range tokens.Raw {
					if token == " " || token == "" {
						continue
					}
					// Use the main transliteration function with dictionary
					trans := TransliterateWordRulesOnly(token)
					if trans != "" {
						results = append(results, trans)
					}
				}
				result = strings.Join(results, " ")
			}
		}
		
		// Fallback if no result
		if result == "" {
			result = TransliterateWordRulesOnly(line)
		}
		
		writer.WriteString(result + "\n")
		
		// Progress indicator
		if lineNum % 50 == 0 {
			fmt.Printf("Processed %d lines...\n", lineNum)
		}
	}
	
	fmt.Printf("Transliteration complete! Processed %d lines.\n", lineNum)
	fmt.Println("Output saved to test.romanized.txt")
}

// containsThai checks if a string contains Thai characters
func containsThai(s string) bool {
	for _, r := range s {
		if r >= 0x0E00 && r <= 0x0E7F {
			return true
		}
	}
	return false
}