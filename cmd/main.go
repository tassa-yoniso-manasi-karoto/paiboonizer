package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"golang.org/x/text/unicode/norm"

	"github.com/tassa-yoniso-manasi-karoto/go-pythainlp"
	"github.com/tassa-yoniso-manasi-karoto/paiboonizer"

	_ "github.com/tassa-yoniso-manasi-karoto/translitkit"
	"github.com/tassa-yoniso-manasi-karoto/translitkit/common"
)

const failuresFile = "testing_files/failures_translitkit.txt"

// testPair represents a matched pair of Thai input and expected transliteration
type testPair struct {
	name          string
	inputLines    []string
	expectedLines []string
}

// corpusFailure represents a single failed transliteration
type corpusFailure struct {
	file     string
	lineNum  int
	input    string
	expected string
	got      string
}

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

// discoverCorpus finds all testN.txt + testN_Opus4.5_transliterated.txt pairs
func discoverCorpus(dir string) ([]testPair, error) {
	pattern := filepath.Join(dir, "testing_files", "test*.txt")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	warn := color.New(color.FgYellow)
	errColor := color.New(color.FgRed)

	var pairs []testPair
	for _, inputPath := range matches {
		// Skip transliterated files
		if strings.Contains(inputPath, "_Opus4.5_transliterated") {
			continue
		}

		// Derive expected path: testN.txt -> testN_Opus4.5_transliterated.txt
		base := strings.TrimSuffix(filepath.Base(inputPath), ".txt")
		expectedPath := filepath.Join(filepath.Dir(inputPath), base+"_Opus4.5_transliterated.txt")

		// Check expected file exists
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			warn.Printf("WARNING: No transliteration for %s, skipping\n", base)
			continue
		}

		// Load files
		inputs, err := loadLines(inputPath)
		if err != nil {
			errColor.Printf("ERROR: Failed to load %s: %v\n", inputPath, err)
			continue
		}
		expected, err := loadLines(expectedPath)
		if err != nil {
			errColor.Printf("ERROR: Failed to load %s: %v\n", expectedPath, err)
			continue
		}

		// VALIDATION: Line count must match
		if len(inputs) != len(expected) {
			errColor.Printf("ERROR: Line mismatch in %s: %d vs %d, skipping\n",
				base, len(inputs), len(expected))
			continue
		}

		pairs = append(pairs, testPair{
			name:          base,
			inputLines:    inputs,
			expectedLines: expected,
		})
	}

	// Sort for consistent order (test1, test2, test8...)
	sort.Slice(pairs, func(i, j int) bool {
		return naturalLess(pairs[i].name, pairs[j].name)
	})

	return pairs, nil
}

// naturalLess compares strings with embedded numbers naturally
// e.g., "test2" < "test10"
func naturalLess(a, b string) bool {
	numA := extractNumber(a)
	numB := extractNumber(b)
	if numA != numB {
		return numA < numB
	}
	return a < b
}

// extractNumber extracts the first number from a string
func extractNumber(s string) int {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(s)
	if match == "" {
		return 0
	}
	n, _ := strconv.Atoi(match)
	return n
}

// loadLines reads a file and returns all lines
// Aegisub \N markers are replaced with single spaces
func loadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Replace Aegisub subtitle line breaks with single space
		line = strings.ReplaceAll(line, "\\N", " ")
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

// punctuationRegex matches Unicode punctuation characters
var punctuationRegex = regexp.MustCompile(`[\p{P}\p{S}]`)

// normalize prepares strings for comparison
func normalize(s string) string {
	// Remove BOM if present
	s = strings.TrimPrefix(s, "\ufeff")
	s = norm.NFC.String(s)
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	// Remove all Unicode punctuation and symbols
	s = punctuationRegex.ReplaceAllString(s, " ")
	// Normalize ALL whitespace (tabs, multiple spaces, etc.) to single space
	fields := strings.Fields(s)
	s = strings.Join(fields, " ")
	// Normalize ambiguous tones (both are valid for ไหม question particle)
	s = strings.ReplaceAll(s, " mǎi ", " mai ")
	s = strings.ReplaceAll(s, " mái ", " mai ")
	if strings.HasSuffix(s, " mǎi") {
		s = s[:len(s)-len(" mǎi")] + " mai"
	}
	if strings.HasSuffix(s, " mái") {
		s = s[:len(s)-len(" mái")] + " mai"
	}
	// Normalize numbers to Thai romanization for fair comparison
	s = normalizeNumbers(s)
	return s
}

// normalizeNumbers converts Arabic numerals to Thai number romanization
func normalizeNumbers(s string) string {
	// Find and replace number sequences
	var result strings.Builder
	i := 0
	runes := []rune(s)

	for i < len(runes) {
		if runes[i] >= '0' && runes[i] <= '9' {
			// Collect the full number
			numStart := i
			for i < len(runes) && runes[i] >= '0' && runes[i] <= '9' {
				i++
			}
			numStr := string(runes[numStart:i])
			thai := numberToThai(numStr)
			if result.Len() > 0 && result.String()[result.Len()-1] != ' ' {
				result.WriteString(" ")
			}
			result.WriteString(thai)
		} else {
			result.WriteRune(runes[i])
			i++
		}
	}
	return result.String()
}

// numberToThai converts an Arabic numeral string to Thai romanization
func numberToThai(num string) string {
	units := []string{"", "nʉ̀ng", "sɔ̌ɔng", "sǎam", "sìi", "hâa", "hòk", "jèt", "bpɛ̀ɛt", "gâao"}
	tens := []string{"", "sìp", "yîi sìp", "sǎam sìp", "sìi sìp", "hâa sìp", "hòk sìp", "jèt sìp", "bpɛ̀ɛt sìp", "gâao sìp"}

	// Handle single digit
	if len(num) == 1 {
		d := int(num[0] - '0')
		if d == 0 {
			return "sǔun"
		}
		return units[d]
	}

	// Handle two digits (10-99)
	if len(num) == 2 {
		t := int(num[0] - '0')
		u := int(num[1] - '0')
		result := tens[t]
		if u > 0 {
			if u == 1 && t > 0 {
				result += " èt" // Special: 11, 21, 31... use "èt" not "nʉ̀ng"
			} else {
				result += " " + units[u]
			}
		}
		return result
	}

	// For larger numbers, just convert digit by digit for simplicity
	var parts []string
	for _, r := range num {
		d := int(r - '0')
		if d == 0 {
			parts = append(parts, "sǔun")
		} else {
			parts = append(parts, units[d])
		}
	}
	return strings.Join(parts, " ")
}

// runCorpusTranslitkit runs corpus test via translitkit with full failure analysis
func runCorpusTranslitkit(module *common.Module) {
	dir := getTestDir()
	corpus, err := discoverCorpus(dir)
	if err != nil {
		fmt.Printf("Error discovering corpus: %v\n", err)
		return
	}
	if len(corpus) == 0 {
		fmt.Println("No valid test pairs found")
		return
	}

	// Report discovered files
	fmt.Printf("Discovered %d test files:\n", len(corpus))
	totalCorpusLines := 0
	for _, p := range corpus {
		fmt.Printf("  %s: %d lines\n", p.name, len(p.inputLines))
		totalCorpusLines += len(p.inputLines)
	}
	fmt.Printf("Total corpus: %d lines\n\n", totalCorpusLines)

	// Flatten corpus for processing, tracking source file for each line
	type lineInfo struct {
		input    string
		expected string
		file     string
		lineNum  int // Line number within the source file
	}
	var allLines []lineInfo
	for _, p := range corpus {
		for i := range p.inputLines {
			allLines = append(allLines, lineInfo{
				input:    p.inputLines[i],
				expected: p.expectedLines[i],
				file:     p.name,
				lineNum:  i + 1,
			})
		}
	}

	lineCorrect := 0
	totalLines := 0
	wordCorrect := 0
	totalWords := 0
	fallbacks := 0

	var failures []corpusFailure

	for _, line := range allLines {
		input := strings.TrimSpace(line.input)
		exp := normalize(line.expected)

		if input == "" || exp == "" {
			continue
		}
		// Skip Aegisub header lines
		if strings.HasPrefix(input, "#") && strings.Contains(input, "Aegisub") {
			continue
		}
		// Skip lines containing Arabic numerals (unfair to measure)
		if containsDigit(input) {
			continue
		}
		// Skip lines where ground truth uses precomposed accented characters
		// (can't reliably compare with engine output which uses combining marks)
		if hasPrecomposedAccents(line.expected) {
			continue
		}
		totalLines++

		// Use translitkit for transliteration
		result, err := module.Roman(input)
		if err != nil {
			fmt.Printf("Error on [%s:%d]: %v\n", line.file, line.lineNum, err)
			fallbacks++
			continue
		}

		got := normalize(result)

		// Line-level accuracy
		if got == exp {
			lineCorrect++
		} else {
			failures = append(failures, corpusFailure{
				file:     line.file,
				lineNum:  line.lineNum,
				input:    input,
				expected: line.expected,
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
			fmt.Printf("[%s:%d] %s\n", f.file, f.lineNum, f.input)
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
				fmt.Fprintf(file, "[%s:%d] %s\n", f.file, f.lineNum, f.input)
				fmt.Fprintf(file, "  Expected: %s\n", f.expected)
				fmt.Fprintf(file, "  Got:      %s\n\n", f.got)
			}
			fmt.Printf("\nAll %d failures written to: %s\n", len(failures), failuresFile)
		}
	}

	// Generate draft dictionary from failing words
	failedWords := extractFailingWords(failures)
	if len(failedWords) > 0 {
		draftPath := filepath.Join(dir, "testing_files/draft_dictionary.tsv")
		file, err := os.Create(draftPath)
		if err != nil {
			fmt.Printf("Error creating draft dictionary: %v\n", err)
		} else {
			defer file.Close()
			// Sort for consistent output
			sortedWords := make([]string, 0, len(failedWords))
			for word := range failedWords {
				sortedWords = append(sortedWords, word)
			}
			sort.Strings(sortedWords)
			for _, word := range sortedWords {
				fmt.Fprintf(file, "%s\t\n", word)
			}
			fmt.Printf("Draft dictionary: %d words written to %s\n", len(failedWords), "testing_files/draft_dictionary.tsv")
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
	corpus, err := discoverCorpus(dir)
	if err != nil || len(corpus) == 0 {
		fmt.Println("No valid test pairs found")
		return
	}

	// Flatten corpus
	var allInputs, allExpected []string
	for _, p := range corpus {
		allInputs = append(allInputs, p.inputLines...)
		allExpected = append(allExpected, p.expectedLines...)
	}

	wordCorrect := 0
	totalWords := 0

	for i := 0; i < len(allInputs); i++ {
		input := strings.TrimSpace(allInputs[i])
		// Remove BOM
		input = strings.TrimPrefix(input, "\ufeff")
		exp := normalize(allExpected[i])

		if input == "" || exp == "" {
			continue
		}
		// Skip Aegisub header lines
		if strings.HasPrefix(input, "#") && strings.Contains(input, "Aegisub") {
			continue
		}
		// Skip lines containing Arabic numerals (unfair to measure)
		if containsDigit(input) {
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

// containsDigit checks if a string contains Arabic numerals (0-9)
func containsDigit(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// hasPrecomposedAccents checks if ground truth uses precomposed accented vowels
// that official Paiboon doesn't use. Paiboon uses precomposed à, á, â, ǎ, ě, ǐ, ǒ, ǔ
// but uses combining marks for e, i, o, u with grave/acute/circumflex.
// Skip only if ground truth has precomposed forms Paiboon doesn't use.
func hasPrecomposedAccents(s string) bool {
	for _, r := range s {
		switch r {
		// e with grave/acute/circumflex (Paiboon uses combining, not precomposed)
		case 'è', 'é', 'ê': // U+00E8-EA
			return true
		// i with grave/acute/circumflex
		case 'ì', 'í', 'î': // U+00EC-EE
			return true
		// o with grave/acute/circumflex
		case 'ò', 'ó', 'ô': // U+00F2-F4
			return true
		// u with grave/acute/circumflex
		case 'ù', 'ú', 'û': // U+00F9-FB
			return true
		}
	}
	return false
}

// extractFailingWords tokenizes failing Thai inputs and collects unique words
// that aren't in the official dictionary
func extractFailingWords(failures []corpusFailure) map[string]struct{} {
	failedWords := make(map[string]struct{})

	for _, f := range failures {
		// Tokenize the Thai input
		input := strings.TrimPrefix(f.input, "\ufeff")
		tokenResult, err := pythainlp.Tokenize(input)
		if err != nil || tokenResult == nil || len(tokenResult.Raw) == 0 {
			continue
		}

		// Collect Thai words not in official dictionary
		for _, word := range tokenResult.Raw {
			word = strings.TrimSpace(word)
			if word == "" || !containsThai(word) {
				continue
			}
			// Skip if already in official dictionary
			if _, ok := paiboonizer.LookupDictionary(word); ok {
				continue
			}
			// Skip very short words (likely particles or fragments)
			if len([]rune(word)) < 2 {
				continue
			}
			// Skip silent consonant artifacts (e.g., ฟ์, ร์, ว์)
			if paiboonizer.RemoveSilentConsonants(word) == "" {
				continue
			}
			// Skip ๆ (mai yamok) - handled at translitkit level
			if strings.Contains(word, "ๆ") {
				continue
			}
			failedWords[word] = struct{}{}
		}
	}

	return failedWords
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
