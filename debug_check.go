package main

import (
	"fmt"
	"unicode"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/transform"
	"golang.org/x/text/runes"
)

func debugTest() {
	fmt.Println("=== Debug Unicode Issues ===")
	
	// Test cases that are failing despite looking correct
	tests := []struct {
		thai string
		expected string
	}{
		{"น้ำ", "nám"},
		{"คิด", "kít"},
		{"ตก", "dtòk"},
		{"สวัสดี", "sàwàtdii"},
		{"อร่อย", "àròɔi"},
		{"ความสุข", "kwaam-sùk"},
		{"ขอบคุณ", "kɔ̀ɔp-kun"},
	}
	
	for _, test := range tests {
		// Use fallback transliteration to avoid pythainlp startup noise
		result := fallbackTransliteration(test.thai)
		
		// Normalize both for comparison
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		resultNorm, _, _ := transform.String(t, result)
		expectedNorm, _, _ := transform.String(t, test.expected)
		
		fmt.Printf("\n%s:\n", test.thai)
		fmt.Printf("  Result:   %s\n", result)
		fmt.Printf("  Expected: %s\n", test.expected)
		fmt.Printf("  Result bytes:   %q\n", result)
		fmt.Printf("  Expected bytes: %q\n", test.expected)
		fmt.Printf("  Result norm:    %s\n", resultNorm)
		fmt.Printf("  Expected norm:  %s\n", expectedNorm)
		
		// Show syllable breakdown
		syllables := extractSyllables(test.thai)
		fmt.Printf("  Syllables: %v\n", syllables)
		for _, syl := range syllables {
			comp := parseSyllableComponents(syl)
			fmt.Printf("    %s → %s+%s+%s (tone: %s, class: %s)\n", 
				syl, comp.Initial, comp.Vowel, comp.Final, comp.ToneMark, comp.InitialThai)
		}
		
		if result == test.expected {
			fmt.Printf("  ✅ PASS\n")
		} else if resultNorm == expectedNorm {
			fmt.Printf("  ⚠️  PASS (after normalization)\n")
		} else {
			fmt.Printf("  ❌ FAIL\n")
		}
	}
}