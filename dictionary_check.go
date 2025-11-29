package paiboonizer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

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

func TestDictionary(useDictionary bool) {
	if useDictionary {
		TestDictionaryWithMode(TestModeFullDictionary)
	} else {
		TestDictionaryWithMode(TestModePureRules)
	}
}

// TestDictionaryWithMode tests transliteration with specific mode
func TestDictionaryWithMode(mode TestMode) {
	switch mode {
	case TestModePureRules:
		fmt.Println("=== Testing Rule-Based Transliterator (No Dictionary, No Pythainlp) ===")
		fmt.Println("Testing pure rule-based syllable extraction + transliteration")
	case TestModePythainlp:
		fmt.Println("=== Testing Pythainlp Tokenization + Rule-Based Transliteration ===")
		fmt.Println("Testing pythainlp syllable tokenization + rule-based transliteration")
	case TestModeFullDictionary:
		fmt.Println("=== Testing With Full Dictionary Lookup ===")
		fmt.Println("Testing WITH dictionary lookup (baseline)")
	}
	fmt.Println("Dictionary entries available:", len(dictionary))
	fmt.Println("Syllable dictionary entries:", len(syllableDict), "\n")

	passed := 0
	total := 0
	failures := []struct {
		thai     string
		expected string
		got      string
	}{}

	// Test each dictionary entry
	for thai, expected := range dictionary {
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

		// Remove hyphens and tildes for comparison (syllable separation style not important)
		expectedNoSep := strings.ReplaceAll(strings.ReplaceAll(cleanExpected, "-", ""), "~", "")
		resultNoSep := strings.ReplaceAll(strings.ReplaceAll(result, "-", ""), "~", "")

		// Legacy variable names for compatibility
		expectedNoHyphen := expectedNoSep
		resultNoHyphen := resultNoSep
		
		// Also normalize Unicode for fair comparison
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		resultNorm, _, _ := transform.String(t, resultNoHyphen)
		expectedNorm, _, _ := transform.String(t, expectedNoHyphen)
		
		if resultNoHyphen == expectedNoHyphen || resultNorm == expectedNorm {
			passed++
		} else {
			// Collect failures for analysis
			if len(failures) < 50 { // Keep first 50 failures for analysis
				failures = append(failures, struct{
					thai string
					expected string
					got string
				}{thai, expected, result})
			}
		}
		
		// Show progress every 500 words
		if total % 500 == 0 {
			fmt.Printf("Progress: %d/%d tested, %.1f%% accuracy so far\n", 
				total, len(dictionary), float64(passed)*100/float64(total))
		}
	}
	
	accuracy := float64(passed) * 100 / float64(total)
	
	fmt.Println("\n=== RESULTS ===")
	fmt.Printf("Total words tested: %d\n", total)
	fmt.Printf("Passed: %d\n", passed)
	fmt.Printf("Failed: %d\n", total-passed)
	fmt.Printf("\nREAL ACCURACY: %.2f%%\n", accuracy)
	
	if accuracy >= 90.0 {
		fmt.Println("\n🎉 SUCCESS! Achieved 90%+ accuracy on dictionary dataset!")
	} else {
		fmt.Printf("\n❌ Need %.2f%% more to reach 90%% target\n", 90.0-accuracy)
	}
	
	// Show some failures for debugging
	fmt.Println("\n=== Sample Failures (first 20) ===")
	for i, f := range failures {
		if i >= 20 {
			break
		}
		fmt.Printf("%s: got '%s', expected '%s'\n", f.thai, f.got, f.expected)
	}
	
	// Analyze failure patterns
	fmt.Println("\n=== Failure Analysis ===")
	toneErrors := 0
	vowelErrors := 0
	consonantErrors := 0
	
	for _, f := range failures {
		// Simple heuristic analysis
		if len(f.got) != len(f.expected) {
			vowelErrors++
		} else {
			// Check if it's mainly tone differences
			t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
			gotNorm, _, _ := transform.String(t, f.got)
			expNorm, _, _ := transform.String(t, f.expected)
			if gotNorm == expNorm {
				toneErrors++
			} else {
				consonantErrors++
			}
		}
	}
	
	if len(failures) > 0 {
		fmt.Printf("Tone errors: ~%d (%.1f%%)\n", toneErrors, float64(toneErrors)*100/float64(len(failures)))
		fmt.Printf("Vowel/length errors: ~%d (%.1f%%)\n", vowelErrors, float64(vowelErrors)*100/float64(len(failures)))
		fmt.Printf("Consonant/other errors: ~%d (%.1f%%)\n", consonantErrors, float64(consonantErrors)*100/float64(len(failures)))
	}
}

// transliterateWithPythainlp uses pythainlp for syllable tokenization
// then transliterates each syllable using rules (no whole-word dictionary lookup)
func transliterateWithPythainlp(word string) string {
	// Ensure pythainlp manager is initialized
	if globalManager == nil || globalManager.nlpManager == nil {
		// Fall back to pure rules if pythainlp not available
		return ComprehensiveTransliterate(word)
	}

	ctx := context.Background()
	result, err := globalManager.nlpManager.SyllableTokenize(ctx, word)
	if err != nil || result == nil || len(result.Syllables) == 0 {
		// Fall back to pure rules
		return ComprehensiveTransliterate(word)
	}

	// Transliterate each syllable using rules (syllable dict + pattern matching)
	results := []string{}
	var lastTrans string // Track last transliteration for ๆ repetition

	for _, syllable := range result.Syllables {
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
	if globalManager != nil {
		return nil // Already initialized
	}

	ctx := context.Background()
	var err error
	globalManager, err = NewManager(ctx)
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
				// Check syllable dict
				if trans, ok := syllableDict[syl]; ok {
					fmt.Printf("  [%d] '%s' → '%s' (syllable dict)\n", i, syl, trans)
				} else if trans, ok := specialCasesGlobal[syl]; ok {
					fmt.Printf("  [%d] '%s' → '%s' (special case)\n", i, syl, trans)
				} else {
					trans := ComprehensiveTransliterate(syl)
					fmt.Printf("  [%d] '%s' → '%s' (rules)\n", i, syl, trans)
				}
			}
		} else {
			fmt.Printf("Pythainlp error: %v\n", err)
		}
	}
	fmt.Println()
}