package paiboonizer

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/tassa-yoniso-manasi-karoto/go-pythainlp"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// stripSpecialMarkers removes annotation markers like <sth>, <sone>, <n> etc.
// These are not Thai words and should be ignored in transliteration
var specialMarkerRegex = regexp.MustCompile(`<[^>]+>`)

func stripSpecialMarkers(s string) string {
	return specialMarkerRegex.ReplaceAllString(s, "")
}

// TestMode controls how transliteration is performed
type TestMode int

const (
	TestModePureRules      TestMode = iota // No dictionary, no pythainlp
	TestModePythainlp                      // Pythainlp tokenization + rules (no whole-word dict)
	TestModeFullDictionary                 // Full dictionary lookup (baseline)
)

// Track pythainlp failures that fell back to pure rules
var pythainlpFallbackCount int

// DictTestFailure represents a single test failure
type DictTestFailure struct {
	Thai     string
	Expected string
	Got      string
}

// DictTestResults contains the results of dictionary testing
type DictTestResults struct {
	Mode               TestMode
	Total              int
	Passed             int
	Failed             int
	Accuracy           float64
	PythainlpFallbacks int
	Failures           []DictTestFailure
	ToneErrors         int
	VowelErrors        int
	ConsonantErrors    int
}

// RunDictionaryTest runs dictionary test and returns results
func RunDictionaryTest(mode TestMode) DictTestResults {
	if mode == TestModePythainlp {
		pythainlpFallbackCount = 0
	}

	passed := 0
	total := 0
	var failures []DictTestFailure

	// Sort dictionary keys for deterministic iteration order
	sortedKeys := make([]string, 0, len(dictionary))
	for k := range dictionary {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	// Test each dictionary entry in deterministic order
	for _, thai := range sortedKeys {
		expected := dictionary[thai]
		// Skip multi-word phrases for now
		if strings.Contains(thai, " ") {
			continue
		}

		total++

		// Strip special markers from Thai text before transliteration
		cleanThai := stripSpecialMarkers(thai)

		// Transliterate based on mode
		var result string
		switch mode {
		case TestModePureRules:
			result = ComprehensiveTransliterate(cleanThai)
		case TestModePythainlp:
			result = transliterateWithPythainlp(cleanThai)
		case TestModeFullDictionary:
			result = TransliterateWordRulesOnly(cleanThai)
		}

		// Strip special markers from expected result too
		cleanExpected := stripSpecialMarkers(expected)

		// Remove hyphens and tildes for comparison
		expectedNoSep := strings.ReplaceAll(strings.ReplaceAll(cleanExpected, "-", ""), "~", "")
		resultNoSep := strings.ReplaceAll(strings.ReplaceAll(result, "-", ""), "~", "")

		// Also normalize Unicode for fair comparison
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		resultNorm, _, _ := transform.String(t, resultNoSep)
		expectedNorm, _, _ := transform.String(t, expectedNoSep)

		if resultNoSep == expectedNoSep || resultNorm == expectedNorm {
			passed++
		} else {
			if len(failures) < 50 {
				failures = append(failures, DictTestFailure{
					Thai:     thai,
					Expected: expected,
					Got:      result,
				})
			}
		}
	}

	// Analyze failure patterns
	toneErrors := 0
	vowelErrors := 0
	consonantErrors := 0

	for _, f := range failures {
		if len(f.Got) != len(f.Expected) {
			vowelErrors++
		} else {
			t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
			gotNorm, _, _ := transform.String(t, f.Got)
			expNorm, _, _ := transform.String(t, f.Expected)
			if gotNorm == expNorm {
				toneErrors++
			} else {
				consonantErrors++
			}
		}
	}

	return DictTestResults{
		Mode:               mode,
		Total:              total,
		Passed:             passed,
		Failed:             total - passed,
		Accuracy:           float64(passed) * 100 / float64(total),
		PythainlpFallbacks: pythainlpFallbackCount,
		Failures:           failures,
		ToneErrors:         toneErrors,
		VowelErrors:        vowelErrors,
		ConsonantErrors:    consonantErrors,
	}
}

// transliterateWithPythainlp uses pythainlp for syllable tokenization
// then transliterates each syllable using rules (no whole-word dictionary lookup)
func transliterateWithPythainlp(word string) string {
	var syllables []string

	if globalManager != nil && globalManager.nlpManager != nil {
		// Use paiboonizer's own manager (standalone mode)
		ctx := context.Background()
		result, err := globalManager.nlpManager.SyllableTokenize(ctx, word)
		if err != nil || result == nil || len(result.Syllables) == 0 {
			pythainlpFallbackCount++
			return ComprehensiveTransliterate(word)
		}
		syllables = result.Syllables
	} else {
		// Try package-level function (uses default manager set by translitkit)
		result, err := pythainlp.SyllableTokenize(word)
		if err != nil || result == nil || len(result.Syllables) == 0 {
			pythainlpFallbackCount++
			return ComprehensiveTransliterate(word)
		}
		syllables = result.Syllables
	}

	// Transliterate each syllable using rules (syllable dict + pattern matching)
	results := []string{}
	var lastTrans string // Track last transliteration for ๆ repetition

	for _, syllable := range syllables {
		// Handle ๆ (mai yamok) - repeat previous syllable
		if syllable == "ๆ" {
			if lastTrans != "" {
				results = append(results, lastTrans)
			}
			continue
		}

		// Strip trailing silent consonants (consonant + ์) before lookup
		// This handles syllables like สันต์ → สัน
		cleanSyllable := RemoveSilentConsonants(syllable)
		if cleanSyllable == "" {
			continue // Skip syllables that are entirely silent
		}

		var trans string

		// Try syllable dictionary first (but NOT whole-word dictionary)
		if t, ok := syllableDict[cleanSyllable]; ok {
			trans = t
		} else if t, ok := specialCasesGlobal[cleanSyllable]; ok {
			// Try special cases for this syllable
			trans = t
		} else {
			// Fall back to rule-based transliteration for this syllable
			trans = ComprehensiveTransliterate(cleanSyllable)
		}

		if trans != "" {
			results = append(results, trans)
			lastTrans = trans
		}
	}

	if len(results) == 0 {
		return ""
	}
	return strings.Join(results, "-")
}

// InitPythainlp initializes the pythainlp manager for testing
func InitPythainlp() error {
	return InitPythainlpWithRecreate(false)
}

// InitPythainlpWithRecreate initializes pythainlp, optionally recreating the container
func InitPythainlpWithRecreate(recreate bool) error {
	if globalManager != nil && !recreate {
		return nil // Already initialized
	}

	ctx := context.Background()
	var err error
	globalManager, err = NewManagerWithRecreate(ctx, recreate)
	if err != nil {
		return fmt.Errorf("failed to initialize pythainlp: %w", err)
	}
	return nil
}

// ClosePythainlp closes the pythainlp manager
func ClosePythainlp() {
	if globalManager != nil {
		globalManager.Close()
		globalManager = nil
	}
}

// DebugTransliteration shows detailed breakdown of how a word is transliterated
func DebugTransliteration(word string) {
	fmt.Printf("\n=== Debug: %s ===\n", word)

	// Show expected from dictionary
	if expected, ok := dictionary[word]; ok {
		fmt.Printf("Expected (dictionary): %s\n", expected)
	}

	// Pure rule-based
	pureResult := ComprehensiveTransliterate(word)
	fmt.Printf("Pure rules result: %s\n", pureResult)

	// Show rule-based syllable extraction
	syllables := ExtractSyllables(word)
	fmt.Printf("Rule-based syllables: %v\n", syllables)
	for i, syl := range syllables {
		trans := ComprehensiveTransliterate(syl)
		fmt.Printf("  [%d] '%s' → '%s'\n", i, syl, trans)
	}

	// With pythainlp (if available)
	if globalManager != nil && globalManager.nlpManager != nil {
		ctx := context.Background()
		result, err := globalManager.nlpManager.SyllableTokenize(ctx, word)
		if err == nil && result != nil {
			fmt.Printf("Pythainlp syllables: %v\n", result.Syllables)
			for i, syl := range result.Syllables {
				// Clean syllable first (same as actual test flow)
				cleanSyl := RemoveSilentConsonants(syl)
				// Check syllable dict
				if trans, ok := syllableDict[cleanSyl]; ok {
					fmt.Printf("  [%d] '%s' → '%s' (syllable dict)\n", i, syl, trans)
				} else if trans, ok := specialCasesGlobal[cleanSyl]; ok {
					fmt.Printf("  [%d] '%s' → '%s' (special case)\n", i, syl, trans)
				} else {
					trans := ComprehensiveTransliterate(cleanSyl)
					fmt.Printf("  [%d] '%s' → '%s' (rules, cleaned: '%s')\n", i, syl, trans, cleanSyl)
				}
			}
		} else {
			fmt.Printf("Pythainlp error: %v\n", err)
		}
	}
	fmt.Println()
}