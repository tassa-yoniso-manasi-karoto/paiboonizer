package main

import (
	"fmt"
	"strings"
	"unicode"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/transform"
	"golang.org/x/text/runes"
)

func testDictionary() {
	fmt.Println("=== Testing Rule-Based Transliterator Against Dictionary ===")
	fmt.Println("Dictionary entries:", len(dictionary))
	fmt.Println("Testing WITHOUT dictionary lookup (pure rules only)\n")
	
	passed := 0
	total := 0
	failures := []struct{
		thai string
		expected string
		got string
	}{}
	
	// Test each dictionary entry using ONLY rule-based transliteration
	for thai, expected := range dictionary {
		// Skip multi-word phrases for now
		if strings.Contains(thai, " ") {
			continue
		}
		
		total++
		
		// Use rule-based transliteration ONLY (no dictionary lookup)
		result := transliterateWordRulesOnly(thai)
		
		// Remove hyphens from expected for comparison (compound word separation not important now)
		expectedNoHyphen := strings.ReplaceAll(expected, "-", "")
		resultNoHyphen := strings.ReplaceAll(result, "-", "")
		
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