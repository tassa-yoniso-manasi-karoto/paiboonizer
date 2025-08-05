package main

import (
	"strings"
)

// Comprehensive Thai vowel patterns for Paiboon
var thaiVowelPatterns = []struct {
	pattern string
	paiboon string
	hasFinal bool // if pattern includes final consonant
}{
	// Long vowels with finals
	{"เCียว", "iao", false},
	{"เCีย", "iia", false},
	{"เCือ", "ʉʉa", false},
	{"เCือC", "ʉʉa", true},
	{"เCียC", "iia", true},
	{"CัวC", "ua", true},
	{"Cัว", "ua", false},
	{"Cวย", "uai", false},
	{"Cาย", "aai", false},
	{"Cาว", "aao", false},
	
	// Complex patterns
	{"เCาะ", "ɔ", false},
	{"เCอะ", "ə", false},
	{"เCอ", "əə", false},
	{"เCิC", "əə", true},
	{"เCา", "ao", false},
	{"เCาC", "ao", true},
	{"เCย", "əəi", false},
	{"เCว", "eeo", false},
	{"เCC", "ee", true},
	{"เC", "ee", false},
	{"เC็C", "e", true},
	
	// แ patterns
	{"แCะ", "ɛ", false},
	{"แC็C", "ɛ", true},
	{"แCC", "ɛɛ", true},
	{"แC", "ɛɛ", false},
	{"แCว", "ɛɛo", false},
	
	// โ patterns
	{"โCะ", "o", false},
	{"โCC", "oo", true},
	{"โC", "oo", false},
	{"โCย", "ooi", false},
	
	// ไ ใ patterns
	{"ไC", "ai", false},
	{"ใC", "ai", false},
	{"ไCย", "ai", false},
	
	// Simple vowels
	{"Cะ", "a", false},
	{"CัC", "a", true},
	{"Cั้C", "a", true},
	{"Cา", "aa", false},
	{"CาC", "aa", true},
	{"Cำ", "am", false},
	{"CิC", "i", true},
	{"Cิ", "i", false},
	{"CีC", "ii", true},
	{"Cี", "ii", false},
	{"CึC", "ʉ", true},
	{"Cึ", "ʉ", false},
	{"CืC", "ʉʉ", true},
	{"Cื", "ʉʉ", false},
	{"CุC", "u", true},
	{"Cุ", "u", false},
	{"CูC", "uu", true},
	{"Cู", "uu", false},
	
	// อ patterns
	{"CอC", "ɔɔ", true},
	{"Cอ", "ɔɔ", false},
	{"C็อC", "ɔ", true},
	
	// Special
	{"CรรC", "a", true},
	{"Cรร", "an", false},
	{"Cร", "ɔɔn", false},
	
	// Inherent vowel patterns (C = consonant only)
	{"CC", "o", true}, // Closed syllable with inherent vowel
	{"C", "ɔɔ", false}, // Open syllable with inherent vowel
}

// improvedTransliterate uses pattern matching for better accuracy
func improvedTransliterate(word string) string {
	if word == "" {
		return ""
	}
	
	// Try each pattern
	for _, vp := range thaiVowelPatterns {
		if match, result := matchPattern(word, vp.pattern, vp.paiboon); match {
			return result
		}
	}
	
	// Fallback - return empty to avoid recursion
	return ""
}

// matchPattern checks if word matches a vowel pattern
func matchPattern(word, pattern, paiboon string) (bool, string) {
	runes := []rune(word)
	patRunes := []rune(pattern)
	
	if len(patRunes) == 0 {
		return false, ""
	}
	
	result := ""
	wordIdx := 0
	patIdx := 0
	initialCons := ""
	finalCons := ""
	
	for patIdx < len(patRunes) && wordIdx < len(runes) {
		if patRunes[patIdx] == 'C' {
			// Match consonant
			if !isConsonant(string(runes[wordIdx])) {
				return false, ""
			}
			
			if patIdx == 0 || (patIdx > 0 && patRunes[patIdx-1] != 'C') {
				// Initial consonant
				initialCons = string(runes[wordIdx])
			} else {
				// Final consonant
				finalCons = string(runes[wordIdx])
			}
			wordIdx++
			patIdx++
		} else {
			// Match exact character
			if runes[wordIdx] != patRunes[patIdx] {
				return false, ""
			}
			wordIdx++
			patIdx++
		}
	}
	
	// Check if we matched the whole pattern
	if patIdx != len(patRunes) || wordIdx != len(runes) {
		return false, ""
	}
	
	// Build result
	if initialCons != "" {
		if trans, ok := initialConsonants[initialCons]; ok {
			result = trans
		}
	}
	
	result += paiboon
	
	if finalCons != "" {
		if trans, ok := finalConsonants[finalCons]; ok {
			// Replace final vowel if needed
			if strings.HasSuffix(paiboon, "ɔɔ") && trans != "" {
				result = result[:len(result)-2] + "o" + trans
			} else {
				result += trans
			}
		}
	}
	
	return true, result
}