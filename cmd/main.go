package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/text/unicode/norm"

	"github.com/tassa-yoniso-manasi-karoto/paiboonizer"

	_ "github.com/tassa-yoniso-manasi-karoto/translitkit"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

const (
	testFile     = "testing_files/test.txt"
	expectedFile = "testing_files/test_Opus4.5_transliterated.txt"
	failuresFile = "testing_files/failures_translitkit.txt"
)

func main() {
	header := color.New(color.Bold, color.FgYellow)

	// Initialize translitkit module (starts pythainlp, sets default manager)
	// Keep it alive for both tests
	module, err := common.GetSchemeModule("tha", "paiboon-hybrid")
	if err != nil {
		fmt.Printf("Error getting translitkit module: %v\n", err)
		return
	}

	fmt.Println("Initializing translitkit (pythainlp + paiboonizer)...")
	if err := module.Init(); err != nil {
		fmt.Printf("Error initializing translitkit: %v\n", err)
		return
	}
	defer module.Close()

	// Test 1: Corpus test with translitkit (full pipeline)
	header.Println("\n=== CORPUS TEST (TRANSLITKIT) ===")
	runCorpusTranslitkit(module)

	// Test 2: Dictionary accuracy test (paiboonizer rules vs dictionary ground truth)
	// Reuses the pythainlp container via default manager
	header.Println("\n=== DICTIONARY TEST (PAIBOONIZER ACCURACY) ===")
	paiboonizer.TestDictionaryWithMode(paiboonizer.TestModePythainlp)
}

// getTestDir returns the directory containing the test files
func getTestDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "."
	}
	return filepath.Dir(filename)
}

// loadLines reads a file and returns all lines
func loadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// normalize prepares strings for comparison
func normalize(s string) string {
	// Remove BOM if present
	s = strings.TrimPrefix(s, "\ufeff")
	s = norm.NFC.String(s)
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// Normalize multiple spaces to single space
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

// runCorpusTranslitkit runs corpus test via translitkit with full failure analysis
func runCorpusTranslitkit(module *common.Module) {
	dir := getTestDir()
	inputPath := filepath.Join(dir, testFile)
	expectedPath := filepath.Join(dir, expectedFile)

	inputs, err := loadLines(inputPath)
	if err != nil {
		fmt.Printf("Error loading input file: %v\n", err)
		return
	}

	expected, err := loadLines(expectedPath)
	if err != nil {
		fmt.Printf("Error loading expected file: %v\n", err)
		return
	}

	if len(inputs) != len(expected) {
		fmt.Printf("Line count mismatch: %d inputs vs %d expected\n", len(inputs), len(expected))
		return
	}

	lineCorrect := 0
	totalLines := 0
	wordCorrect := 0
	totalWords := 0
	fallbacks := 0

	type failure struct {
		line     int
		input    string
		expected string
		got      string
	}
	var failures []failure

	for i := 0; i < len(inputs); i++ {
		input := strings.TrimSpace(inputs[i])
		exp := normalize(expected[i])

		if input == "" || exp == "" {
			continue
		}
		totalLines++

		// Use translitkit for transliteration
		result, err := module.Roman(input)
		if err != nil {
			fmt.Printf("Error on line %d: %v\n", i+1, err)
			fallbacks++
			continue
		}

		got := normalize(result)

		// Line-level accuracy
		if got == exp {
			lineCorrect++
		} else {
			failures = append(failures, failure{
				line:     i + 1,
				input:    input,
				expected: expected[i],
				got:      result,
			})
		}

		// Word-level accuracy
		expWords := splitWords(exp)
		gotWords := splitWords(got)
		totalWords += len(expWords)
		wordCorrect += countMatchingWords(expWords, gotWords)
	}

	// Report fallbacks
	if fallbacks > 0 {
		fmt.Printf("WARNING: Fallbacks occurred: %d\n", fallbacks)
	} else {
		fmt.Printf("Fallbacks: 0 (good!)\n")
	}

	// Show first 30 failures
	showCount := 30
	if len(failures) < showCount {
		showCount = len(failures)
	}

	if showCount > 0 {
		fmt.Printf("\nFirst %d failures:\n", showCount)
		fmt.Println(strings.Repeat("-", 80))
		for i := 0; i < showCount; i++ {
			f := failures[i]
			fmt.Printf("Line %d: %s\n", f.line, f.input)
			fmt.Printf("  Expected: %s\n", f.expected)
			fmt.Printf("  Got:      %s\n", f.got)
		}
		fmt.Println(strings.Repeat("-", 80))
	}

	// Write all failures to file
	failuresPath := filepath.Join(dir, failuresFile)
	if len(failures) > 0 {
		file, err := os.Create(failuresPath)
		if err != nil {
			fmt.Printf("Error creating failures file: %v\n", err)
		} else {
			defer file.Close()
			for _, f := range failures {
				fmt.Fprintf(file, "Line %d: %s\n", f.line, f.input)
				fmt.Fprintf(file, "  Expected: %s\n", f.expected)
				fmt.Fprintf(file, "  Got:      %s\n\n", f.got)
			}
			fmt.Printf("\nAll %d failures written to: %s\n", len(failures), failuresFile)
		}
	}

	lineAccuracy := float64(lineCorrect) / float64(totalLines) * 100
	wordAccuracy := float64(wordCorrect) / float64(totalWords) * 100

	bold := color.New(color.Bold)
	boldCyan := color.New(color.Bold, color.FgCyan)

	fmt.Println()
	bold.Printf("Line-level accuracy: %.2f%% (%d/%d lines)\n", lineAccuracy, lineCorrect, totalLines)
	boldCyan.Printf("CORPUS WORD-LEVEL ACCURACY: %.2f%% (%d/%d words)\n", wordAccuracy, wordCorrect, totalWords)
}

// splitWords splits a romanized string into words by spaces
func splitWords(s string) []string {
	var words []string
	for _, w := range strings.Fields(s) {
		w = strings.TrimSpace(w)
		if w != "" && w != "-" {
			words = append(words, w)
		}
	}
	return words
}

// countMatchingWords counts how many words from expected appear in got (order-sensitive)
func countMatchingWords(expected, got []string) int {
	matches := 0
	gotIdx := 0

	for _, expWord := range expected {
		// Look for this expected word in the remaining got words
		for gotIdx < len(got) {
			if got[gotIdx] == expWord {
				matches++
				gotIdx++
				break
			}
			gotIdx++
		}
	}
	return matches
}
