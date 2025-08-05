package main

import "fmt"

func testSyllables() {
	// Test syllable extraction on common failing words
	words := []string{
		"ความทรงจำ", // kwaam-song-jam
		"ร่วง", // rûuang
		"โชว์", // choo
		"ทะลุ", // tá~lú
		"ตีห้า", // dtii-hâa
		"เหลือ", // lǔǔa
		"คอ", // kɔɔ
		"หนวด", // nùuat
	}
	
	fmt.Println("=== Syllable Extraction Test ===")
	for _, word := range words {
		syllables := extractSyllables(word)
		fmt.Printf("\n%s:\n", word)
		fmt.Printf("  Syllables: %v\n", syllables)
		
		// Show comprehensive parse for each syllable
		for _, syl := range syllables {
			parsed := parseThaiSyllable(syl)
			fmt.Printf("    %s → ", syl)
			fmt.Printf("Lead:'%s' Init:'%s%s' V:'%s%s' Tone:'%s' Final:'%s%s' Silent:'%s'\n",
				parsed.LeadingVowel, parsed.Initial1, parsed.Initial2, 
				parsed.Vowel1, parsed.Vowel2, parsed.Tone, 
				parsed.Final1, parsed.Final2, parsed.Silent)
			
			// Show transliteration
			trans := buildPaiboonFromSyllable(parsed)
			fmt.Printf("      Transliteration: %s\n", trans)
		}
		
		// Show final result
		result := comprehensiveTransliterate(word)
		fmt.Printf("  Final: %s\n", result)
	}
}