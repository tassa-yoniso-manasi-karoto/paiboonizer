package paiboonizer

import (
	"sort"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// VowelPattern represents a Thai vowel pattern for transliteration
type VowelPattern struct {
	pattern  string
	paiboon  string
	hasFinal bool   // if pattern includes final consonant
	priority int    // higher = match first (for same length)
}

// Comprehensive Thai vowel patterns for Paiboon - sorted by specificity
// K = cluster position (กร, กล, etc.), C = single consonant, T = tone mark
var thaiVowelPatterns = []VowelPattern{
	// ===== LONGEST PATTERNS FIRST (6+ chars) =====
	// Complex diphthongs with finals
	{pattern: "เKียวC", paiboon: "iao", hasFinal: true, priority: 100},
	{pattern: "เCียวC", paiboon: "iao", hasFinal: true, priority: 99},
	{pattern: "เKือยC", paiboon: "ʉʉai", hasFinal: true, priority: 98},
	{pattern: "เCือยC", paiboon: "ʉʉai", hasFinal: true, priority: 97},

	// ===== 5 CHARACTER PATTERNS =====
	{pattern: "เKียว", paiboon: "iao", hasFinal: false, priority: 95},
	{pattern: "เCียว", paiboon: "iao", hasFinal: false, priority: 94},
	{pattern: "เKือC", paiboon: "ʉʉa", hasFinal: true, priority: 93},
	{pattern: "เCือC", paiboon: "ʉʉa", hasFinal: true, priority: 92},
	{pattern: "เKียC", paiboon: "iia", hasFinal: true, priority: 91},
	{pattern: "เCียC", paiboon: "iia", hasFinal: true, priority: 90},
	{pattern: "เKิTC", paiboon: "əə", hasFinal: true, priority: 89},
	{pattern: "เCิTC", paiboon: "əə", hasFinal: true, priority: 88},

	// ===== 4 CHARACTER PATTERNS =====
	// เ-ีย patterns
	{pattern: "เKีย", paiboon: "iia", hasFinal: false, priority: 85},
	{pattern: "เCีย", paiboon: "iia", hasFinal: false, priority: 84},
	// เ-ือ patterns
	{pattern: "เKือ", paiboon: "ʉʉa", hasFinal: false, priority: 83},
	{pattern: "เCือ", paiboon: "ʉʉa", hasFinal: false, priority: 82},
	// เ-าะ patterns
	{pattern: "เKาะ", paiboon: "ɔ", hasFinal: false, priority: 81},
	{pattern: "เCาะ", paiboon: "ɔ", hasFinal: false, priority: 80},
	// เ-อะ patterns
	{pattern: "เKอะ", paiboon: "ə", hasFinal: false, priority: 79},
	{pattern: "เCอะ", paiboon: "ə", hasFinal: false, priority: 78},
	// เ-ิ patterns with finals
	{pattern: "เKิC", paiboon: "əə", hasFinal: true, priority: 77},
	{pattern: "เCิC", paiboon: "əə", hasFinal: true, priority: 76},
	// เ-า patterns with finals
	{pattern: "เKาC", paiboon: "ao", hasFinal: true, priority: 75},
	{pattern: "เCาC", paiboon: "ao", hasFinal: true, priority: 74},
	// ัว patterns
	{pattern: "KัวC", paiboon: "ua", hasFinal: true, priority: 73},
	{pattern: "CัวC", paiboon: "ua", hasFinal: true, priority: 72},
	// า+ย/ว patterns with clusters
	{pattern: "Kาย", paiboon: "aai", hasFinal: false, priority: 71},
	{pattern: "Kาว", paiboon: "aao", hasFinal: false, priority: 70},
	// แ patterns with clusters
	{pattern: "แK็C", paiboon: "ɛ", hasFinal: true, priority: 69},
	{pattern: "แC็C", paiboon: "ɛ", hasFinal: true, priority: 68},
	{pattern: "แKCC", paiboon: "ɛɛ", hasFinal: true, priority: 67},
	// โ patterns with clusters
	{pattern: "โKCC", paiboon: "oo", hasFinal: true, priority: 66},
	// รร patterns
	{pattern: "KรรC", paiboon: "a", hasFinal: true, priority: 65},
	{pattern: "CรรC", paiboon: "a", hasFinal: true, priority: 64},

	// ===== 3 CHARACTER PATTERNS =====
	// Cระ/Cรา patterns (very common)
	{pattern: "KระC", paiboon: "à", hasFinal: true, priority: 68},
	{pattern: "CระC", paiboon: "à", hasFinal: true, priority: 67},
	{pattern: "Kระ", paiboon: "à", hasFinal: false, priority: 66},
	{pattern: "Cระ", paiboon: "à", hasFinal: false, priority: 65},
	{pattern: "KราC", paiboon: "aa", hasFinal: true, priority: 64},
	{pattern: "CราC", paiboon: "aa", hasFinal: true, priority: 63},
	{pattern: "Kรา", paiboon: "aa", hasFinal: false, priority: 62},
	{pattern: "Cรา", paiboon: "aa", hasFinal: false, priority: 61},
	// เ patterns
	{pattern: "เKอ", paiboon: "əə", hasFinal: false, priority: 60},
	{pattern: "เCอ", paiboon: "əə", hasFinal: false, priority: 59},
	{pattern: "เKา", paiboon: "ao", hasFinal: false, priority: 58},
	{pattern: "เCา", paiboon: "ao", hasFinal: false, priority: 57},
	{pattern: "เKย", paiboon: "əəi", hasFinal: false, priority: 56},
	{pattern: "เCย", paiboon: "əəi", hasFinal: false, priority: 55},
	{pattern: "เKว", paiboon: "eeo", hasFinal: false, priority: 54},
	{pattern: "เCว", paiboon: "eeo", hasFinal: false, priority: 53},
	{pattern: "เK็C", paiboon: "e", hasFinal: true, priority: 52},
	{pattern: "เC็C", paiboon: "e", hasFinal: true, priority: 51},
	{pattern: "เKC", paiboon: "ee", hasFinal: true, priority: 50},
	{pattern: "เCC", paiboon: "ee", hasFinal: true, priority: 49},
	// แ patterns
	{pattern: "แKะ", paiboon: "ɛ", hasFinal: false, priority: 48},
	{pattern: "แCะ", paiboon: "ɛ", hasFinal: false, priority: 47},
	{pattern: "แKC", paiboon: "ɛɛ", hasFinal: true, priority: 46},
	{pattern: "แCC", paiboon: "ɛɛ", hasFinal: true, priority: 45},
	{pattern: "แKว", paiboon: "ɛɛo", hasFinal: false, priority: 44},
	{pattern: "แCว", paiboon: "ɛɛo", hasFinal: false, priority: 43},
	// โ patterns
	{pattern: "โKะ", paiboon: "o", hasFinal: false, priority: 42},
	{pattern: "โCะ", paiboon: "o", hasFinal: false, priority: 41},
	{pattern: "โKC", paiboon: "oo", hasFinal: true, priority: 40},
	{pattern: "โCC", paiboon: "oo", hasFinal: true, priority: 39},
	{pattern: "โKย", paiboon: "ooi", hasFinal: false, priority: 38},
	{pattern: "โCย", paiboon: "ooi", hasFinal: false, priority: 37},
	// ไ ใ patterns
	{pattern: "ไKย", paiboon: "ai", hasFinal: false, priority: 36},
	{pattern: "ไCย", paiboon: "ai", hasFinal: false, priority: 35},
	{pattern: "ใKย", paiboon: "ai", hasFinal: false, priority: 34},
	{pattern: "ใCย", paiboon: "ai", hasFinal: false, priority: 33},
	// ัว patterns
	{pattern: "Kัว", paiboon: "ua", hasFinal: false, priority: 32},
	{pattern: "Cัว", paiboon: "ua", hasFinal: false, priority: 31},
	// วย patterns
	{pattern: "Kวย", paiboon: "uai", hasFinal: false, priority: 32},
	{pattern: "Cวย", paiboon: "uai", hasFinal: false, priority: 31},
	// า+ย/ว patterns
	{pattern: "Cาย", paiboon: "aai", hasFinal: false, priority: 28},
	{pattern: "Cาว", paiboon: "aao", hasFinal: false, priority: 27},
	// รร patterns
	{pattern: "Kรร", paiboon: "an", hasFinal: false, priority: 26},
	{pattern: "Cรร", paiboon: "an", hasFinal: false, priority: 25},
	// Vowels with clusters and finals
	{pattern: "KัC", paiboon: "a", hasFinal: true, priority: 24},
	{pattern: "KาC", paiboon: "aa", hasFinal: true, priority: 23},
	{pattern: "KิC", paiboon: "i", hasFinal: true, priority: 22},
	{pattern: "KีC", paiboon: "ii", hasFinal: true, priority: 21},
	{pattern: "KึC", paiboon: "ʉ", hasFinal: true, priority: 20},
	{pattern: "KืC", paiboon: "ʉʉ", hasFinal: true, priority: 19},
	{pattern: "KุC", paiboon: "u", hasFinal: true, priority: 18},
	{pattern: "KูC", paiboon: "uu", hasFinal: true, priority: 17},
	{pattern: "KอC", paiboon: "ɔɔ", hasFinal: true, priority: 16},

	// ===== 2 CHARACTER PATTERNS =====
	{pattern: "เK", paiboon: "ee", hasFinal: false, priority: 15},
	{pattern: "เC", paiboon: "ee", hasFinal: false, priority: 14},
	{pattern: "แK", paiboon: "ɛɛ", hasFinal: false, priority: 13},
	{pattern: "แC", paiboon: "ɛɛ", hasFinal: false, priority: 12},
	{pattern: "โK", paiboon: "oo", hasFinal: false, priority: 11},
	{pattern: "โC", paiboon: "oo", hasFinal: false, priority: 10},
	{pattern: "ไK", paiboon: "ai", hasFinal: false, priority: 9},
	{pattern: "ไC", paiboon: "ai", hasFinal: false, priority: 8},
	{pattern: "ใK", paiboon: "ai", hasFinal: false, priority: 7},
	{pattern: "ใC", paiboon: "ai", hasFinal: false, priority: 6},
	// Simple vowels with finals
	{pattern: "Cะ", paiboon: "a", hasFinal: false, priority: 5},
	{pattern: "CัTC", paiboon: "a", hasFinal: true, priority: 4},   // With tone mark
	{pattern: "CัC", paiboon: "a", hasFinal: true, priority: 3},
	{pattern: "Cา", paiboon: "aa", hasFinal: false, priority: 2},
	{pattern: "CาTC", paiboon: "aa", hasFinal: true, priority: 1},  // With tone mark
	{pattern: "CาC", paiboon: "aa", hasFinal: true, priority: 0},
	{pattern: "Cำ", paiboon: "am", hasFinal: false, priority: -1},
	{pattern: "CิTC", paiboon: "i", hasFinal: true, priority: -2},  // With tone mark
	{pattern: "CิC", paiboon: "i", hasFinal: true, priority: -3},
	{pattern: "Cิ", paiboon: "i", hasFinal: false, priority: -4},
	{pattern: "CีTC", paiboon: "ii", hasFinal: true, priority: -5}, // With tone mark
	{pattern: "CีC", paiboon: "ii", hasFinal: true, priority: -6},
	{pattern: "Cี", paiboon: "ii", hasFinal: false, priority: -7},
	{pattern: "CึTC", paiboon: "ʉ", hasFinal: true, priority: -8},  // With tone mark
	{pattern: "CึC", paiboon: "ʉ", hasFinal: true, priority: -9},
	{pattern: "Cึ", paiboon: "ʉ", hasFinal: false, priority: -10},
	{pattern: "CืC", paiboon: "ʉʉ", hasFinal: true, priority: -11},
	{pattern: "Cื", paiboon: "ʉʉ", hasFinal: false, priority: -12},
	{pattern: "CุTC", paiboon: "u", hasFinal: true, priority: -13}, // With tone mark
	{pattern: "CุC", paiboon: "u", hasFinal: true, priority: -14},
	{pattern: "Cุ", paiboon: "u", hasFinal: false, priority: -15},
	{pattern: "CูTC", paiboon: "uu", hasFinal: true, priority: -16}, // With tone mark
	{pattern: "CูC", paiboon: "uu", hasFinal: true, priority: -17},
	{pattern: "Cู", paiboon: "uu", hasFinal: false, priority: -18},
	{pattern: "CTอC", paiboon: "ɔɔ", hasFinal: true, priority: -19}, // Tone before อ
	{pattern: "CอTC", paiboon: "ɔɔ", hasFinal: true, priority: -20}, // Tone after อ
	{pattern: "CอC", paiboon: "ɔɔ", hasFinal: true, priority: -21},
	{pattern: "Cอ", paiboon: "ɔɔ", hasFinal: false, priority: -22},
	{pattern: "C็อC", paiboon: "ɔ", hasFinal: true, priority: -23},
	{pattern: "Cร", paiboon: "ɔɔn", hasFinal: false, priority: -24},

	// ===== SHORTEST PATTERNS (fallback) =====
	{pattern: "CC", paiboon: "o", hasFinal: true, priority: -100},  // Closed syllable inherent
	{pattern: "C", paiboon: "ɔɔ", hasFinal: false, priority: -101}, // Open syllable inherent
}

// sortedVowelPatterns holds patterns sorted by length then priority
var sortedVowelPatterns []VowelPattern

func init() {
	// Sort patterns: longer patterns first, then by priority within same length
	sortedVowelPatterns = make([]VowelPattern, len(thaiVowelPatterns))
	copy(sortedVowelPatterns, thaiVowelPatterns)

	sort.Slice(sortedVowelPatterns, func(i, j int) bool {
		lenI := len([]rune(sortedVowelPatterns[i].pattern))
		lenJ := len([]rune(sortedVowelPatterns[j].pattern))
		if lenI != lenJ {
			return lenI > lenJ // Longer first
		}
		return sortedVowelPatterns[i].priority > sortedVowelPatterns[j].priority
	})
}

// improvedTransliterate uses pattern matching for better accuracy
func improvedTransliterate(word string) string {
	if word == "" {
		return ""
	}

	// Remove silent consonants first
	word = RemoveSilentConsonants(word)

	// Try each pattern from sorted list (longest first)
	for _, vp := range sortedVowelPatterns {
		if match, result := matchPatternImproved(word, vp.pattern, vp.paiboon); match {
			return result
		}
	}

	// Fallback - return empty to avoid recursion
	return ""
}

// matchPatternImproved checks if word matches a vowel pattern
// K = cluster (2 consonants), C = single consonant, T = tone mark
func matchPatternImproved(word, pattern, paiboon string) (bool, string) {
	runes := []rune(word)
	patRunes := []rune(pattern)

	if len(patRunes) == 0 || len(runes) == 0 {
		return false, ""
	}

	wordIdx := 0
	patIdx := 0
	initialCons := ""
	initialCluster := ""
	finalCons := ""
	toneMark := ""
	isCluster := false

	for patIdx < len(patRunes) && wordIdx < len(runes) {
		patChar := patRunes[patIdx]

		switch patChar {
		case 'K': // Cluster (2 consonants)
			// Must match 2 consonants that form a valid cluster
			if wordIdx+1 >= len(runes) {
				return false, ""
			}
			c1 := string(runes[wordIdx])
			c2 := string(runes[wordIdx+1])
			if !isConsonant(c1) || !isConsonant(c2) {
				return false, ""
			}
			cluster := c1 + c2
			if _, ok := clusters[cluster]; !ok {
				return false, "" // Not a valid cluster
			}
			initialCluster = cluster
			initialCons = c1 // For tone class
			isCluster = true
			wordIdx += 2
			patIdx++

		case 'C': // Single consonant
			if !isConsonant(string(runes[wordIdx])) {
				return false, ""
			}
			// Determine if this is initial or final
			if initialCons == "" && !isCluster {
				initialCons = string(runes[wordIdx])
			} else {
				finalCons = string(runes[wordIdx])
			}
			wordIdx++
			patIdx++

		case 'T': // Tone mark
			if wordIdx < len(runes) && isToneMark(string(runes[wordIdx])) {
				toneMark = string(runes[wordIdx])
				wordIdx++
			}
			// Tone mark is optional in pattern
			patIdx++

		default:
			// Match exact character (vowel markers, etc.)
			if runes[wordIdx] != patChar {
				return false, ""
			}
			wordIdx++
			patIdx++
		}
	}

	// Check if we matched the whole pattern and word
	if patIdx != len(patRunes) || wordIdx != len(runes) {
		return false, ""
	}

	// Build result
	result := ""

	// Initial consonant/cluster
	if isCluster {
		if trans, ok := clusters[initialCluster]; ok {
			result = trans
		}
	} else if initialCons != "" {
		if trans, ok := initialConsonants[initialCons]; ok {
			result = trans
		}
	}

	// Vowel
	result += paiboon

	// Final consonant
	if finalCons != "" {
		if trans, ok := finalConsonants[finalCons]; ok {
			// Adjust for closed syllable inherent vowel
			if strings.HasSuffix(paiboon, "ɔɔ") && trans != "" {
				result = result[:len(result)-2] + "o" + trans
			} else {
				result += trans
			}
		}
	}

	// Apply tone
	result = applyToneToResult(result, initialCons, initialCluster, toneMark, paiboon, finalCons)

	return true, result
}

// applyToneToResult applies tone marking to the romanized result
func applyToneToResult(result, initialCons, cluster, toneMark, vowel, finalCons string) string {
	// Determine tone class
	toneClass := "mid"
	toneKey := initialCons
	if cluster != "" {
		// Check if cluster has special tone class (ห-clusters)
		if tc, ok := clusterToneClass[cluster]; ok {
			toneClass = tc
		} else {
			toneKey = string([]rune(cluster)[0])
		}
	}

	if toneClass == "mid" {
		if highClass[toneKey] {
			toneClass = "high"
		} else if lowClass[toneKey] {
			toneClass = "low"
		}
	}

	// Determine live/dead syllable
	isLive := isLiveSyllable(vowel, finalCons)

	// Determine vowel length (for dead-short vs dead-long tone distinction)
	longVowel := isLongVowel(vowel)

	// Calculate tone number
	toneNum := calculateToneNum(toneClass, isLive, toneMark, longVowel)

	if toneNum == 0 {
		return result
	}

	// Add tone diacritic
	return addToneDiacritic(result, toneNum)
}

// isLiveSyllable determines if a syllable is live or dead
func isLiveSyllable(vowel, finalCons string) bool {
	// Dead endings
	deadFinals := map[string]bool{"p": true, "t": true, "k": true}
	if finalCons != "" {
		if trans, ok := finalConsonants[finalCons]; ok {
			if deadFinals[trans] {
				return false
			}
		}
	}

	// Long vowels and sonorant endings make syllable live
	longVowels := []string{"aa", "ii", "ʉʉ", "uu", "ee", "ɛɛ", "oo", "ɔɔ", "əə", "iia", "ʉʉa", "uua", "ai", "ao", "aai", "aao"}
	for _, lv := range longVowels {
		if strings.Contains(vowel, lv) {
			return true
		}
	}

	// Short vowels with no final or sonorant final are live
	if finalCons == "" {
		return true
	}

	// Check for sonorant finals (m, n, ng, y, w)
	if trans, ok := finalConsonants[finalCons]; ok {
		sonorants := map[string]bool{"m": true, "n": true, "ng": true, "i": true, "o": true}
		return sonorants[trans]
	}

	return false
}

// isLongVowel checks if a romanized vowel is long
// Long vowels have doubled letters (aa, ii, uu, etc.) or specific patterns
// Note: "ai" and "ao" are SHORT diphthongs; "aai" and "aao" are long
func isLongVowel(vowel string) bool {
	// Check long diphthongs first (before short ones would match via Contains)
	longPatterns := []string{"aai", "aao", "aa", "ii", "ʉʉ", "uu", "ee", "ɛɛ", "oo", "ɔɔ", "əə", "iia", "ʉʉa", "uua"}
	for _, lp := range longPatterns {
		if strings.Contains(vowel, lp) {
			return true
		}
	}
	return false
}

// calculateToneNum calculates the tone number based on Thai tone rules
// isLongVowelParam is used to distinguish dead-short vs dead-long for low-class consonants
func calculateToneNum(toneClass string, isLive bool, toneMark string, isLongVowelParam bool) int {
	if toneMark == "" {
		// Inherent tone rules
		switch toneClass {
		case "low":
			if isLive {
				return 0 // mid
			}
			// Dead syllable with low-class initial:
			// - Short vowel → HIGH tone (Wiktionary: dead-short → high)
			// - Long vowel → FALLING tone (Wiktionary: dead-long → falling)
			if isLongVowelParam {
				return 3 // falling
			}
			return 2 // high
		case "mid":
			if isLive {
				return 0 // mid
			}
			return 1 // low
		case "high":
			if isLive {
				return 4 // rising
			}
			return 1 // low
		}
	} else {
		// Tone mark rules
		switch toneMark {
		case "่": // mai ek
			if toneClass == "low" {
				return 3 // falling
			}
			return 1 // low
		case "้": // mai tho
			if toneClass == "low" {
				return 2 // high
			}
			return 3 // falling
		case "๊": // mai tri
			if toneClass == "mid" {
				return 2 // high
			}
		case "๋": // mai jattawa
			if toneClass == "mid" {
				return 4 // rising
			}
		}
	}
	return 0
}

// addToneDiacritic adds tone diacritic to first vowel in result
func addToneDiacritic(text string, toneNum int) string {
	if toneNum == 0 {
		return text
	}

	marks := map[int]string{
		1: "\u0300", // grave (low)
		2: "\u0301", // acute (high)
		3: "\u0302", // circumflex (falling)
		4: "\u030C", // caron (rising)
	}

	// Find first vowel and add tone mark
	result := []rune(text)
	for i, r := range result {
		if isRomanVowel(r) {
			// Insert tone mark after this vowel
			before := string(result[:i+1])
			after := string(result[i+1:])
			// Normalize to NFC for consistent comparison
			return norm.NFC.String(before + marks[toneNum] + after)
		}
	}

	return text
}