package paiboonizer

import (
	"strings"

	"github.com/rivo/uniseg"
	"golang.org/x/text/unicode/norm"
)

// ComprehensiveSyllable represents a parsed Thai syllable
type ComprehensiveSyllable struct {
	LeadingVowel string
	Initial1     string // First consonant
	Initial2     string // Second consonant (cluster)
	Vowel1       string // First vowel mark
	Tone         string // Tone mark
	Vowel2       string // Second vowel mark
	Final1       string // First final consonant
	Final2       string // Second final consonant (rare)
	Silent       string // Silent markers
}

// parseThaiSyllable parses a Thai syllable comprehensively
func parseThaiSyllable(syl string) ComprehensiveSyllable {
	var cs ComprehensiveSyllable

	// Remove silent consonants (consonant + ์) before parsing
	syl = RemoveSilentConsonants(syl)

	runes := []rune(syl)
	i := 0
	
	// 1. Leading vowel (เ แ โ ไ ใ)
	if i < len(runes) && isLeadingVowel(string(runes[i])) {
		cs.LeadingVowel = string(runes[i])
		i++
	}
	
	// 2. Initial consonant(s)
	if i < len(runes) && isConsonant(string(runes[i])) {
		cs.Initial1 = string(runes[i])
		i++
		
		// Check for second consonant (cluster)
		if i < len(runes) && isConsonant(string(runes[i])) {
			// Special case for Cร patterns
			if string(runes[i]) == "ร" {
				// Check if followed by ะ or า (กระ, กรา patterns)
				if i+1 < len(runes) && (string(runes[i+1]) == "ะ" || string(runes[i+1]) == "า") {
					cs.Initial2 = string(runes[i])
					i++
					// The vowel will be picked up in the next section
				} else {
					// Regular cluster with ร
					cs.Initial2 = string(runes[i])
					i++
				}
			} else {
				// Check if it's a valid cluster
				cluster := cs.Initial1 + string(runes[i])
				if _, ok := clusters[cluster]; ok {
					cs.Initial2 = string(runes[i])
					i++
				} else if cs.Initial1 == "ห" && (string(runes[i]) == "น" || string(runes[i]) == "ม" || string(runes[i]) == "ล" || string(runes[i]) == "ว" || string(runes[i]) == "ย") {
					// ห leading consonant clusters
					cs.Initial2 = string(runes[i])
					i++
				} else if i+1 < len(runes) && !isVowel(string(runes[i+1])) && !isToneMark(string(runes[i+1])) {
					// Not a cluster, might be final consonant
					// Don't consume it here
				}
			}
		}
	}
	
	// 3. Vowels and tone marks
	for i < len(runes) {
		r := string(runes[i])
		
		if isVowel(r) {
			if cs.Vowel1 == "" {
				cs.Vowel1 = r
			} else {
				cs.Vowel2 += r
			}
			i++
		} else if isToneMark(r) {
			cs.Tone = r
			i++
		} else if r == "็" || r == "์" || r == "ํ" || r == "ๆ" {
			// Special marks
			cs.Silent += r
			i++
		} else if isConsonant(r) {
			// Final consonant(s)
			if cs.Final1 == "" {
				cs.Final1 = r
			} else {
				cs.Final2 = r
			}
			i++
		} else {
			i++
		}
	}
	
	return cs
}

// buildPaiboonFromSyllable converts parsed syllable to Paiboon
func buildPaiboonFromSyllable(cs ComprehensiveSyllable) string {
	result := ""
	vowelSound := ""
	
	// Get initial consonant sound
	initialSound := ""
	if cs.Initial2 != "" {
		// Check for cluster
		cluster := cs.Initial1 + cs.Initial2
		if trans, ok := clusters[cluster]; ok {
			initialSound = trans
		} else if cs.Initial2 == "ร" && (cs.Vowel1 == "ะ" || cs.Vowel1 == "า") {
			// Special Cระ/Cรา patterns (like กระ, ครา)
			if trans, ok := initialConsonants[cs.Initial1]; ok {
				initialSound = trans + "r"
				// Add tilde for shortened vowel in Cระ pattern
				if cs.Vowel1 == "ะ" {
					initialSound = trans + "r" + "à~"
					cs.Vowel1 = "" // Don't process the vowel again
					vowelSound = "skip" // Skip vowel processing
				}
			}
		} else if cs.Initial1 == "ห" {
			// ห is silent in these clusters
			if trans, ok := initialConsonants[cs.Initial2]; ok {
				initialSound = trans
			}
		}
	} else if cs.Initial1 != "" {
		if trans, ok := initialConsonants[cs.Initial1]; ok {
			initialSound = trans
		}
	}
	
	// Determine vowel sound based on pattern (skip if already handled)
	if vowelSound == "skip" {
		vowelSound = "" // Reset for final assembly
	} else {
	// Handle specific patterns
	if cs.LeadingVowel == "เ" {
		if cs.Vowel1 == "ี" && cs.Vowel2 == "ย" {
			vowelSound = "iia"
		} else if cs.Vowel1 == "ี" && cs.Final1 == "ย" {
			// เรียน pattern where ย is part of vowel
			vowelSound = "iia"
			cs.Final1 = cs.Final2
			cs.Final2 = ""
		} else if cs.Vowel1 == "ื" && cs.Vowel2 == "อ" {
			vowelSound = "ʉʉa"
		} else if cs.Vowel1 == "ื" && cs.Final1 == "อ" {
			// เดือน pattern
			vowelSound = "ʉʉa"
			cs.Final1 = "n"
		} else if cs.Vowel1 == "" && cs.Final1 == "ย" {
			vowelSound = "əəi"
			cs.Final1 = "" // ย is part of vowel
		} else if cs.Vowel1 == "า" {
			vowelSound = "ao"
		} else if cs.Vowel1 == "ิ" {
			vowelSound = "əə"
		} else if cs.Vowel1 == "อ" {
			if cs.Final1 == "ะ" {
				vowelSound = "ə"
				cs.Final1 = ""
			} else {
				vowelSound = "əə"
			}
		} else if cs.Vowel1 == "็" {
			vowelSound = "e"
		} else if cs.Vowel1 == "า" && cs.Vowel2 == "ะ" {
			vowelSound = "ɔ"
		} else if cs.Vowel1 == "ี" && cs.Vowel2 == "ย" && cs.Final1 == "ว" {
			vowelSound = "iao"
			cs.Final1 = ""
		} else if cs.Vowel1 == "ี" && cs.Final1 == "่" {
			// เยี่ยม pattern
			vowelSound = "ii"
		} else if cs.Vowel1 == "" {
			vowelSound = "ee"
		}
	} else if cs.LeadingVowel == "แ" {
		if cs.Vowel1 == "" {
			vowelSound = "ɛɛ"
		} else if cs.Vowel1 == "ะ" {
			vowelSound = "ɛ"
		} else if cs.Vowel1 == "็" {
			vowelSound = "ɛ"
		}
	} else if cs.LeadingVowel == "โ" {
		if cs.Vowel1 == "" {
			vowelSound = "oo"
		} else if cs.Vowel1 == "ะ" {
			vowelSound = "o"
		}
	} else if cs.LeadingVowel == "ไ" || cs.LeadingVowel == "ใ" {
		vowelSound = "ai"
	} else {
		// No leading vowel - check complex patterns first
		if cs.Vowel1 == "ั" && cs.Vowel2 == "ว" {
			vowelSound = "ua"
		} else if cs.Vowel1 == "ิ" && cs.Vowel2 == "ว" {
			vowelSound = "io"
		} else if cs.Vowel1 == "ื" && cs.Vowel2 == "อ" {
			vowelSound = "ʉʉa"
		} else if cs.Vowel1 == "า" && cs.Vowel2 == "ย" {
			vowelSound = "aai"
		} else if cs.Vowel1 == "า" && cs.Vowel2 == "ว" {
			vowelSound = "aao"
		} else if cs.Initial1 == "ร" && cs.Vowel1 == "" && cs.Vowel2 == "" {
			// Special case for ร as syllable
			vowelSound = "ɔɔ"
			// Final becomes n
			if cs.Final1 == "" {
				cs.Final1 = "n" // Built-in final n sound
			}
		} else if cs.Vowel1 == "า" {
			vowelSound = "aa"
		} else if cs.Vowel1 == "ะ" {
			vowelSound = "a"
		} else if cs.Vowel1 == "ั" {
			vowelSound = "a"
		} else if cs.Vowel1 == "ิ" {
			vowelSound = "i"
		} else if cs.Vowel1 == "ี" {
			vowelSound = "ii"
		} else if cs.Vowel1 == "ึ" {
			vowelSound = "ʉ"
		} else if cs.Vowel1 == "ื" {
			vowelSound = "ʉʉ"
		} else if cs.Vowel1 == "ุ" {
			vowelSound = "u"
		} else if cs.Vowel1 == "ู" {
			vowelSound = "uu"
		} else if cs.Vowel1 == "ำ" {
			vowelSound = "am"
		} else if cs.Vowel1 == "ว" && cs.Final1 == "" {
			// ว as vowel
			vowelSound = "ua"
		} else if cs.Vowel1 == "" && cs.Vowel2 == "" {
			// Inherent vowel
			if cs.Final1 == "" {
				vowelSound = "ɔɔ" // Open syllable
			} else if cs.Final1 == "ร" {
				// Special inherent vowel before ร
				vowelSound = "ɔɔ"
			} else {
				vowelSound = "o" // Closed syllable
			}
		}
	}
	}
	
	// Get final consonant sound
	finalSound := ""
	if cs.Final1 != "" && !strings.Contains(cs.Silent, "์") { // Not silenced
		if cs.Final1 == "n" {
			// Already set as final n (for ร case)
			finalSound = "n"
		} else if trans, ok := finalConsonants[cs.Final1]; ok {
			finalSound = trans
		}
	}
	
	// Build result
	result = initialSound + vowelSound + finalSound
	
	// Apply tone
	toneClass := "mid"
	if cs.Initial1 == "ห" && cs.Initial2 != "" {
		// ห affects tone class
		toneClass = "high"
	} else if highClass[cs.Initial1] {
		toneClass = "high"
	} else if lowClass[cs.Initial1] {
		toneClass = "low"
	}
	
	// Determine if live or dead syllable
	isLive := finalSound == "" || finalSound == "n" || finalSound == "m" || finalSound == "ng" || 
			strings.Contains(vowelSound, "aa") || strings.Contains(vowelSound, "ii") || 
			strings.Contains(vowelSound, "ʉʉ") || strings.Contains(vowelSound, "uu") ||
			strings.Contains(vowelSound, "ee") || strings.Contains(vowelSound, "ɛɛ") ||
			strings.Contains(vowelSound, "oo") || strings.Contains(vowelSound, "ɔɔ")
	
	// Apply tone mark
	toneNum := 0
	if cs.Tone == "" {
		// Inherent tone
		if toneClass == "low" && !isLive {
			toneNum = 2 // high
		} else if toneClass == "high" && isLive {
			toneNum = 4 // rising
		} else if toneClass == "high" && !isLive {
			toneNum = 1 // low
		} else if toneClass == "mid" && !isLive {
			toneNum = 1 // low
		}
	} else {
		// Explicit tone mark
		switch cs.Tone {
		case "่":
			if toneClass == "low" {
				toneNum = 3 // falling
			} else {
				toneNum = 1 // low
			}
		case "้":
			if toneClass == "low" {
				toneNum = 2 // high
			} else {
				toneNum = 3 // falling
			}
		case "๊":
			if toneClass == "mid" {
				toneNum = 2 // high
			}
		case "๋":
			if toneClass == "mid" {
				toneNum = 4 // rising
			}
		}
	}
	
	// Add tone diacritic to first vowel using proper grapheme handling
	if toneNum > 0 {
		toneMarks := map[int]string{
			1: "\u0300", // grave
			2: "\u0301", // acute 
			3: "\u0302", // circumflex
			4: "\u030C", // caron
		}
		
		// Find first vowel and add tone mark properly
		graphemes := uniseg.NewGraphemes(result)
		var newResult strings.Builder
		tonePlaced := false
		
		for graphemes.Next() {
			cluster := graphemes.Str()
			// Check if this grapheme contains a vowel
			if !tonePlaced && len(cluster) > 0 {
				for _, r := range cluster {
					if isRomanVowel(r) {
						// Add the vowel with the tone mark
						newResult.WriteString(cluster + toneMarks[toneNum])
						tonePlaced = true
						break
					}
				}
				if !tonePlaced {
					newResult.WriteString(cluster)
				}
			} else {
				newResult.WriteString(cluster)
			}
		}
		result = newResult.String()
	}

	// Normalize to NFC for consistent comparison
	return norm.NFC.String(result)
}

// findSyllableEndComprehensive finds syllable boundaries with better pattern recognition
func findSyllableEndComprehensive(runes []rune, start int) int {
	if start >= len(runes) {
		return start
	}
	
	// Check for special patterns first
	if start+3 <= len(runes) {
		// Check เCียน pattern (like เรียน)
		if string(runes[start]) == "เ" && isConsonant(string(runes[start+1])) {
			if start+4 <= len(runes) && string(runes[start+2]) == "ี" && string(runes[start+3]) == "ย" {
				// Check if there's a final consonant
				if start+5 <= len(runes) && isConsonant(string(runes[start+4])) {
					return start + 5 // เCียC pattern
				}
				return start + 4 // เCีย pattern
			}
			// Check เCือน pattern
			if start+4 <= len(runes) && string(runes[start+2]) == "ื" && string(runes[start+3]) == "อ" {
				if start+5 <= len(runes) && isConsonant(string(runes[start+4])) {
					return start + 5
				}
				return start + 4
			}
		}
		
		// Check กระ/กรา patterns and other Cระ/Cรา patterns
		if start+2 < len(runes) && string(runes[start+1]) == "ร" {
			if string(runes[start+2]) == "ะ" {
				// Check for final consonant after Cระ
				if start+3 < len(runes) && isConsonant(string(runes[start+3])) {
					return start + 4 // CระC pattern
				}
				return start + 3 // Cระ pattern
			} else if string(runes[start+2]) == "า" {
				// Check for final consonant after Cรา
				if start+3 < len(runes) && isConsonant(string(runes[start+3])) {
					return start + 4 // CราC pattern
				}
				return start + 3 // Cรา pattern
			}
		}
	}
	
	// Fall back to improved syllable finder
	return findSyllableEndImproved(runes, start)
}

// ComprehensiveTransliterate performs advanced Thai-to-Paiboon transliteration
// using comprehensive syllable parsing, pattern recognition, and tone rules.
// It handles complex vowel patterns, consonant clusters, and special cases.
func ComprehensiveTransliterate(word string) string {
	ensureDictionaryLoaded()
	// Try special cases first (irregular words, loanwords)
	if trans, ok := specialCasesGlobal[word]; ok {
		return norm.NFC.String(trans)
	}

	// Try syllable dictionary for known syllables
	if trans, ok := syllableDict[word]; ok {
		return norm.NFC.String(trans)
	}

	// Try to find longest matching syllables from dictionary and special cases
	results := []string{}
	runes := []rune(word)
	i := 0

	for i < len(runes) {
		found := false
		// Try longest possible match first (maximal matching)
		// Limit search to reasonable syllable lengths (max 8 runes for a syllable)
		maxLen := len(runes) - i
		if maxLen > 8 {
			maxLen = 8
		}

		for length := maxLen; length > 0; length-- {
			if i+length <= len(runes) {
				substr := string(runes[i : i+length])

				// Check if this match would leave an orphan consonant
				// (a single consonant without a vowel at the end)
				if i+length < len(runes) {
					remaining := runes[i+length:]
					if len(remaining) == 1 && isConsonant(string(remaining[0])) {
						// Would leave orphan consonant - skip this match
						// unless it's at the start of a new syllable pattern
						continue
					}
				}

				// Check special cases first
				if trans, ok := specialCasesGlobal[substr]; ok {
					results = append(results, norm.NFC.String(trans))
					i += length
					found = true
					break
				}
				// Then check syllable dictionary
				if trans, ok := syllableDict[substr]; ok {
					results = append(results, norm.NFC.String(trans))
					i += length
					found = true
					break
				}
			}
		}

		if !found {
			// Extract one syllable using improved rules
			end := findSyllableEndComprehensive(runes, i)
			if end > i {
				syl := string(runes[i:end])
				// Try pattern matching on single syllable
				trans := improvedTransliterate(syl)
				if trans == "" {
					// Fall back to comprehensive parsing
					parsed := parseThaiSyllable(syl)
					trans = buildPaiboonFromSyllable(parsed)
				}
				if trans != "" {
					results = append(results, trans)
				}
				i = end
			} else {
				// Single character
				parsed := parseThaiSyllable(string(runes[i]))
				trans := buildPaiboonFromSyllable(parsed)
				if trans != "" {
					results = append(results, trans)
				}
				i++
			}
		}
	}

	if len(results) == 0 {
		return ""
	}
	// Normalize to NFC to match dictionary expectations (precomposed characters)
	return norm.NFC.String(strings.Join(results, ""))
}