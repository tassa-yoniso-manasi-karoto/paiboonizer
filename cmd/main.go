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

	"github.com/tassa-yoniso-manasi-karoto/go-pythainlp"
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

	// Test 2: Corpus test with pure rules (pythainlp tokenization + paiboonizer rules, no dictionary)
	header.Println("\n=== CORPUS TEST (PURE RULES) ===")
	runCorpusPureRules()

	// Test 3: Dictionary accuracy test (paiboonizer rules vs dictionary ground truth)
	// Reuses the pythainlp container via default manager
	header.Println("\n=== DICTIONARY TEST (PAIBOONIZER ACCURACY) ===")
	dictResults := paiboonizer.RunDictionaryTest(paiboonizer.TestModePythainlp)
	printDictResults(dictResults)
}

// printDictResults formats dictionary test results with color
func printDictResults(r paiboonizer.DictTestResults) {
	fmt.Println("Testing pythainlp syllable tokenization + rule-based transliteration")
	fmt.Printf("Dictionary entries: %d, Syllable dict: %d\n\n", 4981, 2772) // TODO: export these

	fmt.Println("=== RESULTS ===")
	fmt.Printf("Total: %d | Passed: %d | Failed: %d\n", r.Total, r.Passed, r.Failed)
	if r.PythainlpFallbacks > 0 {
		fmt.Printf("Pythainlp fallbacks: %d (%.1f%%)\n", r.PythainlpFallbacks, float64(r.PythainlpFallbacks)*100/float64(r.Total))
	}

	boldGreen := color.New(color.Bold, color.FgGreen)
	boldGreen.Printf("\nDICTIONARY ACCURACY: %.2f%%\n", r.Accuracy)

	// Sample failures
	if len(r.Failures) > 0 {
		fmt.Println("\n=== Sample Failures (first 20) ===")
		for i, f := range r.Failures {
			if i >= 20 {
				break
			}
			fmt.Printf("%s: got '%s', expected '%s'\n", f.Thai, f.Got, f.Expected)
		}

		fmt.Println("\n=== Failure Analysis ===")
		fmt.Printf("Tone: ~%d (%.1f%%) | Vowel/length: ~%d (%.1f%%) | Consonant: ~%d (%.1f%%)\n",
			r.ToneErrors, float64(r.ToneErrors)*100/float64(len(r.Failures)),
			r.VowelErrors, float64(r.VowelErrors)*100/float64(len(r.Failures)),
			r.ConsonantErrors, float64(r.ConsonantErrors)*100/float64(len(r.Failures)))
	}
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

// runCorpusPureRules runs corpus test with pythainlp tokenization + pure rule-based transliteration
// (no dictionary lookup). Silent output - just accuracy %.
func runCorpusPureRules() {
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

	wordCorrect := 0
	totalWords := 0

	for i := 0; i < len(inputs); i++ {
		input := strings.TrimSpace(inputs[i])
		// Remove BOM
		input = strings.TrimPrefix(input, "\ufeff")
		exp := normalize(expected[i])

		if input == "" || exp == "" {
			continue
		}

		// Use pythainlp for word tokenization (via package-level function)
		tokenResult, err := pythainlp.Tokenize(input)
		if err != nil || tokenResult == nil || len(tokenResult.Raw) == 0 {
			continue
		}

		// Transliterate each word using pure rules (no dictionary)
		var romanParts []string
		for _, word := range tokenResult.Raw {
			word = strings.TrimSpace(word)
			if word == "" {
				continue
			}
			// Check if it's Thai text
			if containsThai(word) {
				roman := paiboonizer.ComprehensiveTransliterate(word)
				romanParts = append(romanParts, roman)
			} else {
				// Non-Thai passes through (spaces, punctuation, numbers)
				romanParts = append(romanParts, word)
			}
		}

		got := normalize(strings.Join(romanParts, " "))

		// Word-level accuracy
		expWords := splitWords(exp)
		gotWords := splitWords(got)
		totalWords += len(expWords)
		wordCorrect += countMatchingWords(expWords, gotWords)
	}

	wordAccuracy := float64(wordCorrect) / float64(totalWords) * 100
	boldMagenta := color.New(color.Bold, color.FgMagenta)
	boldMagenta.Printf("CORPUS PURE RULES WORD-LEVEL ACCURACY: %.2f%% (%d/%d words)\n", wordAccuracy, wordCorrect, totalWords)
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
