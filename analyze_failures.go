package paiboonizer

import "fmt"

func AnalyzeFailures() {
	fmt.Println("=== Analyzing Specific Failures ===")
	
	// Test cases that are failing
	tests := []struct {
		thai string
		expected string
	}{
		{"นิ้ว", "níu"},
		{"กด", "gòt"},
		{"เลย", "ləəi"},
		{"หลั่ง", "làng"},
		{"เปื้อน", "bpʉ̂ʉan"},
		{"เรียน", "riian"},
		{"รู้", "rúu"},
		{"ด้วย", "dûuai"},
		{"แล้ว", "lɛ́ɛo"},
		{"อยู่", "yùu"},
	}
	
	for _, test := range tests {
		result := TransliterateWordRulesOnly(test.thai)
		fmt.Printf("\nThai: %s\n", test.thai)
		fmt.Printf("Expected: %s\n", test.expected)
		fmt.Printf("Got:      %s\n", result)
		
		if result != test.expected {
			fmt.Printf("Status:   ❌ FAIL\n")
			
			// Debug: show syllables
			syllables := ExtractSyllables(test.thai)
			fmt.Printf("Syllables: %v\n", syllables)
			
			for _, syl := range syllables {
				comp := parseSyllableComponents(syl)
				fmt.Printf("  %s → Initial:'%s' Vowel:'%s' Final:'%s' Tone:'%s'\n", 
					syl, comp.Initial, comp.Vowel, comp.Final, comp.ToneMark)
				
				// Show rune breakdown
				fmt.Printf("    Runes: ")
				for _, r := range syl {
					fmt.Printf("%c(U+%04X) ", r, r)
				}
				fmt.Println()
			}
		} else {
			fmt.Printf("Status:   ✅ PASS\n")
		}
	}
}