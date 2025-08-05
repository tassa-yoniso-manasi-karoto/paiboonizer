package main

import (
	"context"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	pythainlp "github.com/tassa-yoniso-manasi-karoto/go-pythainlp"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Global dictionary built from manual vocab
var dictionary = make(map[string]string)
var syllableDict = make(map[string]string)

// Consonant mappings
var initialConsonants = map[string]string{
	"ก": "g", "ข": "k", "ฃ": "k", "ค": "k", "ฅ": "k", "ฆ": "k", "ง": "ng",
	"จ": "j", "ฉ": "ch", "ช": "ch", "ซ": "s", "ฌ": "ch", "ญ": "y", "ฎ": "d",
	"ฏ": "dt", "ฐ": "t", "ฑ": "t", "ฒ": "t", "ณ": "n", "ด": "d", "ต": "dt",
	"ถ": "t", "ท": "t", "ธ": "t", "น": "n", "บ": "b", "ป": "bp", "ผ": "p",
	"ฝ": "f", "พ": "p", "ฟ": "f", "ภ": "p", "ม": "m", "ย": "y", "ร": "r",
	"ฤ": "rʉ", "ล": "l", "ฦ": "lʉ", "ว": "w", "ศ": "s", "ษ": "s", "ส": "s",
	"ห": "h", "ฬ": "l", "อ": "", "ฮ": "h",
}

var finalConsonants = map[string]string{
	"ก": "k", "ข": "k", "ฃ": "k", "ค": "k", "ฅ": "k", "ฆ": "k", "ง": "ng",
	"จ": "t", "ฉ": "t", "ช": "t", "ซ": "t", "ฌ": "t", "ญ": "n", "ฎ": "t",
	"ฏ": "t", "ฐ": "t", "ฑ": "t", "ฒ": "t", "ณ": "n", "ด": "t", "ต": "t",
	"ถ": "t", "ท": "t", "ธ": "t", "น": "n", "บ": "p", "ป": "p", "ผ": "p",
	"ฝ": "p", "พ": "p", "ฟ": "p", "ภ": "p", "ม": "m", "ย": "i", "ร": "n",
	"ล": "n", "ว": "o", "ศ": "t", "ษ": "t", "ส": "t", "ห": "", "ฬ": "n",
	"อ": "", "ฮ": "",
}

// Tone classes
var (
	highClass = map[string]bool{
		"ข": true, "ฃ": true, "ฉ": true, "ฐ": true, "ถ": true,
		"ผ": true, "ฝ": true, "ศ": true, "ษ": true, "ส": true, "ห": true,
	}
	midClass = map[string]bool{
		"ก": true, "จ": true, "ฎ": true, "ฏ": true,
		"ด": true, "ต": true, "บ": true, "ป": true, "อ": true,
	}
	lowClass = map[string]bool{
		"ค": true, "ฅ": true, "ฆ": true, "ง": true, "ช": true, "ซ": true,
		"ฌ": true, "ญ": true, "ฑ": true, "ฒ": true, "ณ": true, "ท": true,
		"ธ": true, "น": true, "พ": true, "ฟ": true, "ภ": true, "ม": true,
		"ย": true, "ร": true, "ล": true, "ว": true, "ฬ": true, "ฮ": true,
	}
)

// Common clusters
var clusters = map[string]string{
	"กร": "gr", "กล": "gl", "กว": "gw",
	"ขร": "kr", "ขล": "kl", "ขว": "kw",
	"คร": "kr", "คล": "kl", "คว": "kw",
	"ปร": "bpr", "ปล": "bpl",
	"พร": "pr", "พล": "pl",
	"ผล": "pl",
	"ฟร": "fr", "ฟล": "fl",
	"ตร": "dtr", "ทร": "s",
	"ดร": "dr",
}

// Global pythainlp manager
var nlpManager *pythainlp.PyThaiNLPManager

// Initialize pythainlp manager
func initPyThaiNLP(ctx context.Context) error {
	var err error
	nlpManager, err = pythainlp.NewManager(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize pythainlp: %w", err)
	}
	
	// Initialize the service
	if err := nlpManager.Init(ctx); err != nil {
		return fmt.Errorf("failed to start pythainlp service: %w", err)
	}
	
	return nil
}

// ThaiToRoman is the main transliteration function using go-pythainlp
func ThaiToRoman(text string) string {
	// First, try direct dictionary lookup for the whole text
	if trans, ok := dictionary[text]; ok {
		return trans
	}
	
	ctx := context.Background()
	
	// Initialize pythainlp if not already done
	if nlpManager == nil {
		if err := initPyThaiNLP(ctx); err != nil {
			fmt.Printf("Warning: pythainlp not available, using fallback: %v\n", err)
			return fallbackTransliteration(text)
		}
	}
	
	// Tokenize using pythainlp
	opts := pythainlp.AnalyzeOptions{
		Features:       []string{"tokenize", "syllable"},
		TokenizeEngine: "newmm",
		SyllableEngine: "han_solo",
	}
	
	result, err := nlpManager.AnalyzeWithOptions(ctx, text, opts)
	if err != nil {
		fmt.Printf("Warning: tokenization failed, using fallback: %v\n", err)
		return fallbackTransliteration(text)
	}
	
	// Process word by word
	results := []string{}
	for _, word := range result.RawTokens {
		// Skip empty tokens and spaces
		if word == "" || word == " " {
			continue
		}
		
		// Try dictionary lookup first
		if trans, ok := dictionary[word]; ok {
			results = append(results, trans)
			continue
		}
		
		// Fall back to syllable-by-syllable transliteration
		wordResult := transliterateWordWithSyllables(word, result.Syllables)
		if wordResult != "" {
			results = append(results, wordResult)
		}
	}
	
	// Join with hyphen for compound words, but merge syllables within words
	if len(results) > 1 {
		// Check if the original text has spaces (multi-word phrase)
		if strings.Contains(text, " ") {
			return strings.Join(results, " ")
		}
		// Otherwise it's a compound word, join with hyphens
		return strings.Join(results, "-")
	}
	
	return strings.Join(results, "")
}

// fallbackTransliteration when pythainlp is not available
func fallbackTransliteration(text string) string {
	// First, try direct dictionary lookup
	if trans, ok := dictionary[text]; ok {
		return trans
	}
	
	// Fall back to simple transliteration
	return transliterateWord(text)
}

// transliterateWordWithSyllables handles a word with known syllables from pythainlp
func transliterateWordWithSyllables(word string, allSyllables []string) string {
	// Try dictionary first
	if trans, ok := dictionary[word]; ok {
		return trans
	}
	
	// Find syllables that belong to this word
	wordSyllables := []string{}
	currentPos := 0
	wordRunes := []rune(word)
	
	for _, syl := range allSyllables {
		sylRunes := []rune(syl)
		// Check if this syllable matches the current position in the word
		if currentPos < len(wordRunes) {
			match := true
			for i, r := range sylRunes {
				if currentPos+i >= len(wordRunes) || wordRunes[currentPos+i] != r {
					match = false
					break
				}
			}
			if match {
				wordSyllables = append(wordSyllables, syl)
				currentPos += len(sylRunes)
			}
		}
		if currentPos >= len(wordRunes) {
			break
		}
	}
	
	// If pythainlp gave us wrong syllables, try our own extraction
	if len(wordSyllables) == 0 || currentPos < len(wordRunes) {
		wordSyllables = extractSyllables(word)
	}
	
	// Transliterate each syllable
	results := []string{}
	for _, syl := range wordSyllables {
		// Try syllable dictionary
		if trans, ok := syllableDict[syl]; ok {
			results = append(results, trans)
			continue
		}
		
		// Fall back to rule-based transliteration
		trans := transliterateSyllable(syl)
		if trans != "" {
			results = append(results, trans)
		}
	}
	
	if len(results) == 0 {
		return ""
	}
	return strings.Join(results, "")
}

// transliterateWord handles a single Thai word without known syllables
func transliterateWord(word string) string {
	// Try dictionary first
	if trans, ok := dictionary[word]; ok {
		return trans
	}
	
	// Get syllables using simple extraction
	syllables := extractSyllables(word)
	
	results := []string{}
	for _, syl := range syllables {
		// Try syllable dictionary
		if trans, ok := syllableDict[syl]; ok {
			results = append(results, trans)
			continue
		}
		
		// Fall back to rule-based transliteration
		trans := transliterateSyllable(syl)
		if trans != "" {
			results = append(results, trans)
		}
	}
	
	if len(results) == 0 {
		return ""
	}
	return strings.Join(results, "")
}

// transliterateWordRulesOnly handles a word using ONLY rules (no dictionary)
func transliterateWordRulesOnly(word string) string {
	// Use comprehensive transliteration
	return comprehensiveTransliterate(word)
}

// extractSyllables extracts syllables from a word
func extractSyllables(word string) []string {
	syllables := []string{}
	runes := []rune(word)
	i := 0
	
	for i < len(runes) {
		sylEnd := findSyllableEndImproved(runes, i)
		if sylEnd > i {
			syllables = append(syllables, string(runes[i:sylEnd]))
			i = sylEnd
		} else {
			// Single character syllable
			syllables = append(syllables, string(runes[i]))
			i++
		}
	}
	
	return syllables
}

// findSyllableEndImproved finds syllable boundaries more accurately
func findSyllableEndImproved(runes []rune, start int) int {
	if start >= len(runes) {
		return start
	}
	
	i := start
	hasLeadingVowel := false
	hasVowel := false
	
	// 1. Check for leading vowel
	if i < len(runes) && isLeadingVowel(string(runes[i])) {
		hasLeadingVowel = true
		hasVowel = true
		i++
	}
	
	// 2. Get consonant(s)
	consonantStart := i
	consonantCount := 0
	for i < len(runes) && isConsonant(string(runes[i])) {
		consonantCount++
		i++
		
		// Check for valid clusters
		if consonantCount == 2 {
			cluster := string(runes[consonantStart:i])
			if _, isCluster := clusters[cluster]; !isCluster {
				// Not a valid cluster
				if string(runes[consonantStart]) != "ห" {
					// Back up unless it's ห (which can form special clusters)
					i--
					consonantCount--
					break
				}
			}
			break
		}
	}
	
	// 3. Get vowels and tone marks
	for i < len(runes) {
		r := string(runes[i])
		if isVowel(r) {
			hasVowel = true
			i++
		} else if isToneMark(r) || r == "็" || r == "์" || r == "ํ" || r == "ๆ" {
			i++
		} else {
			break
		}
	}
	
	// 4. Check for final consonant
	if i < len(runes) && isConsonant(string(runes[i])) {
		// Take final consonant if:
		// - We have a vowel
		// - Or it's a closed syllable pattern
		if hasVowel {
			// Check if next position starts a new syllable
			nextIsNewSyllable := false
			if i+1 < len(runes) {
				next := string(runes[i+1])
				nextIsNewSyllable = isLeadingVowel(next) || 
					(isVowel(next) && !hasLeadingVowel) ||
					(isConsonant(next) && hasLeadingVowel)
			}
			
			if !nextIsNewSyllable {
				i++ // Take the final consonant
			}
		} else if consonantCount == 1 && !hasLeadingVowel {
			// CVC pattern with inherent vowel
			i++
		}
	}
	
	// Special check for lone tone marks or special characters
	if i == start && i < len(runes) {
		if isToneMark(string(runes[i])) || string(runes[i]) == "์" {
			i++
		}
	}
	
	return i
}

// findSyllableEnd finds the end of a Thai syllable
func findSyllableEnd(runes []rune, start int) int {
	if start >= len(runes) {
		return start
	}
	
	i := start
	hasVowel := false
	hasLeadingVowel := false
	
	// Check for leading vowel (เ แ โ ไ ใ)
	if i < len(runes) && isLeadingVowel(string(runes[i])) {
		hasLeadingVowel = true
		hasVowel = true
		i++
	}
	
	// Get initial consonant(s)
	consonantCount := 0
	initialPos := i
	for i < len(runes) && isConsonant(string(runes[i])) {
		i++
		consonantCount++
		// Allow up to 2 consonants (cluster), but be careful with หล patterns
		if consonantCount >= 2 {
			// Check if it's a valid cluster
			clusterStr := string(runes[initialPos:i])
			if _, isCluster := clusters[clusterStr]; !isCluster {
				// Not a valid cluster, back up one
				if consonantCount == 2 && string(runes[initialPos]) != "ห" {
					i--
					consonantCount--
				}
			}
			break
		}
	}
	
	// Get vowels and marks
	for i < len(runes) {
		r := string(runes[i])
		if isVowel(r) || isToneMark(r) || r == "็" || r == "์" || r == "ํ" || r == "ๆ" {
			i++
			if isVowel(r) {
				hasVowel = true
			}
		} else {
			break
		}
	}
	
	// Check for final consonant
	if i < len(runes) && isConsonant(string(runes[i])) {
		// Special cases for certain finals
		if hasLeadingVowel {
			// With leading vowel, always take final consonant
			i++
		} else if hasVowel {
			// Has vowel, take final if not starting new syllable
			nextIsVowel := i+1 < len(runes) && (isVowel(string(runes[i+1])) || isLeadingVowel(string(runes[i+1])))
			if !nextIsVowel {
				i++
			}
		} else if consonantCount == 1 {
			// Single consonant with no vowel - it's CVC pattern with inherent vowel
			i++
		}
	}
	
	// Ensure we moved forward
	if i == start {
		return start + 1
	}
	
	return i
}

// transliterateSyllable converts a Thai syllable to Paiboon
func transliterateSyllable(syllable string) string {
	// Special cases and common words
	specialCases := map[string]string{
		"ธรรม": "tam",
		"กรรม": "gam",
		"สัตว์": "sàt",
		"จริง": "jing",
		"นอน": "nɔɔn",
		"แดง": "dɛɛng",
		"โชค": "chôok",
		"ลูก": "lûuk",
		"เขียว": "kǐao",
		"สวัส": "sàwàt",
		"อร่อ": "àròɔ",
		"สวัสดี": "sàwàtdii",
		"ขอบ": "kɔ̀ɔp",
		"คุณ": "kun",
		"ความ": "kwaam",
		"สุข": "sùk",
		"อร่อย": "àròɔi",
		"ไม้": "mái",
		"สวย": "sǔai",
		"ขอบคุณ": "kɔ̀ɔp-kun",
		"ความสุข": "kwaam-sùk",
		"ภาษา": "paasǎa",
		"ภาษาไทย": "paasǎa-tai",
		"ประ": "bprà",
		"เทศ": "têet",
		"ประเทศ": "bpràtêet",
		"ประเทศไทย": "bpràtêet-tai",
	}
	
	if trans, ok := specialCases[syllable]; ok {
		return trans
	}
	
	// Parse syllable components
	components := parseSyllableComponents(syllable)
	
	// Build transliteration
	result := components.Initial + components.Vowel + components.Final
	
	// Apply tone
	result = applyTone(result, components)
	
	return result
}

// SyllableComponents represents the parts of a Thai syllable
type SyllableComponents struct {
	Initial     string // Initial consonant(s) sound
	Vowel       string // Vowel sound
	Final       string // Final consonant sound
	ToneMark    string // Tone mark if any
	InitialThai string // Original Thai initial for tone class
}

// parseSyllableComponents breaks down a Thai syllable
func parseSyllableComponents(syllable string) SyllableComponents {
	comp := SyllableComponents{}
	runes := []rune(syllable)
	i := 0
	
	leadingVowel := ""
	vowelMarks := ""
	finalCons := ""
	
	// Check for leading vowel
	if i < len(runes) && isLeadingVowel(string(runes[i])) {
		leadingVowel = string(runes[i])
		i++
	}
	
	// Special case: อย as initial (as in อยู่)
	if i < len(runes)-1 && string(runes[i]) == "อ" && string(runes[i+1]) == "ย" {
		// Check if followed by vowel (not final position)
		if i+2 < len(runes) && isVowel(string(runes[i+2])) {
			comp.Initial = "y"
			comp.InitialThai = "อ"
			i += 2 // Skip อย
		} else {
			// อย at end would be different
		}
	}
	
	// Get initial consonant(s) if not already set
	if comp.Initial == "" {
		initialCons := ""
		for i < len(runes) && isConsonant(string(runes[i])) {
			initialCons += string(runes[i])
			i++
			
			// Check for หล pattern - ห is silent, ล is the real initial
			if initialCons == "หล" {
				comp.Initial = "l" 
				comp.InitialThai = "ห" // For tone class
				initialCons = "หล"
				break
			}
			
			// Special case: สว cluster in สวัสดี type words
			if initialCons == "ส" && i < len(runes) && string(runes[i]) == "ว" {
				// Check if next is ั (sara a) to form ัว vowel
				if i+1 < len(runes) && string(runes[i+1]) == "ั" {
					// ส + วั pattern
					initialCons = "ส"
					vowelMarks = "วั"
					i += 2 // Skip ว and ั
					break
				}
			}
			if len([]rune(initialCons)) >= 2 {
				break
			}
		}
		
		// Store Thai initial for tone class if not set
		if initialCons != "" && comp.InitialThai == "" {
			comp.InitialThai = string([]rune(initialCons)[0])
		}
		
		// Check for cluster if Initial not set
		if comp.Initial == "" {
			if cluster, ok := clusters[initialCons]; ok {
				comp.Initial = cluster
			} else if initialCons != "" {
				// Use first consonant only
				firstCons := string([]rune(initialCons)[0])
				if trans, ok := initialConsonants[firstCons]; ok {
					comp.Initial = trans
				}
				// If there's a second consonant not in cluster, it might be part of vowel or final
				if len([]rune(initialCons)) > 1 {
					secondCons := string([]rune(initialCons)[1])
					if secondCons == "ว" && vowelMarks == "" {
						vowelMarks = "ว" + vowelMarks
					} else if _, isCluster := clusters[initialCons]; !isCluster {
						// Not a cluster, treat second consonant as final
						finalCons = secondCons
					}
				}
			}
		}
		
		// Special case for อ as initial with ร following
		if initialCons == "อ" && i < len(runes) && string(runes[i]) == "ร" {
			comp.Initial = "" // อ is silent initially before ร
			vowelMarks = "อร"
			i++
		}
	}
	
	// Process remaining characters
	for i < len(runes) {
		r := string(runes[i])
		if isVowel(r) || r == "็" {
			vowelMarks += r
			i++
		} else if isToneMark(r) {
			comp.ToneMark = r
			i++
		} else if r == "์" {
			// Thanthakhat - silence marker
			i++
		} else if isConsonant(r) {
			// Final consonant
			finalCons = r
			i++
		} else {
			i++
		}
	}
	
	// Determine vowel sound
	comp.Vowel = determineVowelSound(leadingVowel, vowelMarks, finalCons)
	
	// Set final consonant sound
	if finalCons != "" {
		if trans, ok := finalConsonants[finalCons]; ok {
			comp.Final = trans
		}
	}
	
	return comp
}

// determineVowelSound determines the vowel sound from Thai components
func determineVowelSound(leading, marks, final string) string {
	// Handle complex vowel patterns first
	if leading == "เ" {
		if marks == "ีย" || marks == "ียว" {
			return "iia"
		} else if marks == "ือ" {
			return "ʉʉa"  
		} else if marks == "า" {
			return "ao"
		} else if marks == "อ" {
			return "əə"
		} else if marks == "ิ" {
			return "əə"
		} else if marks == "็" {
			return "e"
		} else if marks == "" && final == "ย" {
			// เ-ย pattern as in เลย
			return "əəi"
		} else if marks == "" && final != "" {
			return "ee" // เ-C as in เห็ด
		} else if marks == "" {
			return "ee"
		} else if marks == "ะ" {
			return "e"
		} else if marks == "าะ" {
			return "ɔ"
		}
	} else if leading == "แ" {
		if marks == "็" {
			return "ɛ"
		} else if marks == "ะ" {
			return "ɛ"
		} else if marks == "" && final != "" {
			return "ɛɛ" // แ-C as in แดง
		} else {
			return "ɛɛ"
		}
	} else if leading == "โ" {
		if marks == "ะ" {
			return "o"
		} else if marks == "" && final != "" {
			return "oo" // โ-C as in โชค
		} else {
			return "oo"
		}
	} else if leading == "ไ" || leading == "ใ" {
		return "ai"
	}
	
	// Special patterns with ว
	if marks == "้ว" || marks == "ว" {
		// ด้วย pattern
		if final == "ย" {
			return "uai"
		}
		// Otherwise ว acts as ัว
		return "ua"
	}
	
	// Diphthongs and complex vowels
	if marks == "ียว" {
		return "iao"
	} else if marks == "ือ" {
		if final == "น" {
			// เปื้อน pattern
			return "ʉʉa"
		}
		return "ʉʉ"
	} else if marks == "วั" || marks == "ัว" {
		return "ua"
	} else if marks == "ิว" || marks == "ิ้ว" {
		// นิ้ว pattern
		return "i"
	} else if marks == "อร" {
		return "ɔɔ"
	} else if marks == "อ" && leading == "" {
		return "ɔɔ"
	}
	
	// Simple vowel marks
	switch marks {
	case "ะ":
		return "a"
	case "ั":
		return "a"
	case "า":
		return "aa"
	case "ิ", "ิ้":
		return "i"
	case "ี", "ี้":
		return "ii"
	case "ึ":
		return "ʉ"
	case "ื", "ื้":
		return "ʉʉ"
	case "ุ":
		return "u"
	case "ู", "ู้":
		return "uu"
	case "ำ":
		return "am"
	case "็อ":
		return "ɔ"
	}
	
	// Inherent vowel (no explicit vowel mark)
	if leading == "" && marks == "" {
		if final == "" {
			return "ɔɔ" // Open syllable
		}
		return "o" // Closed syllable with short o
	}
	
	return ""
}

// applyTone applies tone marks to the transliteration
func applyTone(text string, comp SyllableComponents) string {
	// Determine tone class
	toneClass := "mid"
	if highClass[comp.InitialThai] {
		toneClass = "high"
	} else if lowClass[comp.InitialThai] {
		toneClass = "low"
	}
	
	// Determine if syllable is live or dead
	// Live: ends in sonorant (m, n, ng, y, w) or long vowel
	// Dead: ends in stop (p, t, k) or short vowel
	isLive := false
	
	// Check final consonant
	if comp.Final == "" || comp.Final == "n" || comp.Final == "m" || comp.Final == "ng" || comp.Final == "i" || comp.Final == "o" {
		isLive = true
	} else if comp.Final == "p" || comp.Final == "t" || comp.Final == "k" {
		isLive = false
	}
	
	// Check vowel length (long vowels make syllable live)
	longVowels := []string{"aa", "ii", "ʉʉ", "uu", "ee", "ɛɛ", "oo", "ɔɔ", "əə", "iia", "ʉʉa", "uua"}
	for _, lv := range longVowels {
		if strings.Contains(comp.Vowel, lv) {
			isLive = true
			break
		}
	}
	
	// Diphthongs are also live
	if strings.Contains(comp.Vowel, "ai") || strings.Contains(comp.Vowel, "ao") {
		isLive = true
	}
	
	// Get tone number based on Thai tone rules
	toneNum := 0 // mid tone by default
	
	if comp.ToneMark == "" {
		// No tone mark - use inherent tone rules
		switch toneClass {
		case "low":
			if isLive {
				toneNum = 0 // mid tone
			} else {
				toneNum = 2 // high tone (short dead syllable)
			}
		case "mid":
			if isLive {
				toneNum = 0 // mid tone
			} else {
				toneNum = 1 // low tone
			}
		case "high":
			if isLive {
				toneNum = 4 // rising tone
			} else {
				toneNum = 1 // low tone
			}
		}
	} else {
		// Apply tone mark rules
		switch comp.ToneMark {
		case "่": // mai ek
			switch toneClass {
			case "low":
				toneNum = 3 // falling
			case "mid":
				toneNum = 1 // low
			case "high":
				toneNum = 1 // low
			}
		case "้": // mai tho  
			switch toneClass {
			case "low":
				toneNum = 2 // high
			case "mid":
				toneNum = 3 // falling
			case "high":
				toneNum = 3 // falling
			}
		case "๊": // mai tri
			if toneClass == "mid" {
				toneNum = 2 // high
			}
			// Ignored on high/low class
		case "๋": // mai jattawa
			if toneClass == "mid" {
				toneNum = 4 // rising
			}
			// Ignored on high/low class
		}
	}
	
	// Add tone mark to the romanization
	if toneNum == 0 {
		return text // No tone mark for mid tone
	}
	
	marks := map[int]string{
		1: "\u0300", // grave (low)
		2: "\u0301", // acute (high) 
		3: "\u0302", // circumflex (falling)
		4: "\u030C", // caron (rising)
	}
	
	// Find first vowel to add tone mark
	runes := []rune(text)
	for i, r := range runes {
		if isRomanVowel(r) {
			return string(runes[:i+1]) + marks[toneNum] + string(runes[i+1:])
		}
	}
	
	return text
}

// Helper functions
func isConsonant(s string) bool {
	return strings.Contains("กขฃคฅฆงจฉชซฌญฎฏฐฑฒณดตถทธนบปผฝพฟภมยรฤลฦวศษสหฬอฮ", s)
}

func isVowel(s string) bool {
	return strings.Contains("ะัาิีึืุูเแโใไๅำ", s)
}

func isLeadingVowel(s string) bool {
	return s == "เ" || s == "แ" || s == "โ" || s == "ไ" || s == "ใ"
}

func isToneMark(s string) bool {
	return strings.Contains("่้๊๋", s)
}

func isRomanVowel(r rune) bool {
	return strings.ContainsRune("aeiouəɛɔʉ", r)
}

// Testing functions
func test(th, trg string) {
	r := ThaiToRoman(th)
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	r, _, _ = transform.String(t, r)
	fmt.Println(isPassed(r, trg), th, "\t\t\t\t", r, "→", trg)
}

func isPassed(result, target string) string {
	str := "FAIL"
	c := color.FgRed.Render
	if result == target {
		c = color.FgGreen.Render
		str = "OK"
	}
	return fmt.Sprintf(c(str))
}

// Data loading
var words []string
var m = make(map[string]string)
var re = regexp.MustCompile(`(.*),(.*\p{Thai}.*)`)

func init() {
	pp.Println("Building dictionary from manual vocab...")
	path := "/home/voiduser/go/src/paiboonizer/manual vocab/"
	a, err := os.ReadDir(path)
	check(err)
	
	for _, e := range a {
		file := filepath.Join(path, e.Name())
		dat, err := os.ReadFile(file)
		check(err)
		arr := strings.Split(string(dat), "\n")
		
		for _, str := range arr {
			raw := re.FindStringSubmatch(str)
			if len(raw) == 0 {
				continue
			}
			row := strings.Split(raw[2], ",")[:2]
			th := html.UnescapeString(row[0])
			translit := html.UnescapeString(row[1])
			
			// Add to test data
			words = append(words, th)
			m[th] = translit
			
			// Build dictionary
			dictionary[th] = translit
			
			// Try to extract single syllables for syllable dictionary
			// Add short words and very common syllables
			if !strings.Contains(th, " ") {
				if len([]rune(th)) <= 5 && !strings.Contains(translit, "-") {
					syllableDict[th] = translit
				} else if len([]rune(th)) <= 3 {
					// Very short words are almost always single syllables
					syllableDict[th] = translit
				}
			}
		}
	}
	
	fmt.Printf("Dictionary built: %d entries, %d syllables\n", len(dictionary), len(syllableDict))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// Clean up pythainlp on exit
	defer func() {
		if nlpManager != nil {
			ctx := context.Background()
			nlpManager.Stop(ctx)
			nlpManager.Close()
		}
	}()
	
	// Run dictionary test
	testDictionary()
}

func testWiktionary() {
	test("น้ำ", "nám")
	test("ธรรม", "tam")
	test("บาด", "bàat")
	test("บ้า", "bâa")
	test("แข็ง", "kɛ̌ng")
	test("แกะ", "gɛ̀")
	test("แดง", "dɛɛng")
	test("เกาะ", "gɔ̀")
	test("นอน", "nɔɔn")
	test("พ่อ", "pɔ̂ɔ")
	test("เห็ด", "hèt")
	test("เตะ", "dtè")
	test("เยอะ", "yə́")
	test("เดิน", "dəən")
	test("ตก", "dtòk")
	test("โต๊ะ", "dtó")
	test("โชค", "chôok")
	test("คิด", "kít")
	test("อีก", "ìik")
	test("จี้", "jîi")
	test("ลึก", "lʉ́k")
	test("รึ", "rʉ́")
	test("ชื่อ", "chʉ̂ʉ")
	test("คุก", "kúk")
	test("ลูก", "lûuk")
	test("ปู", "bpuu")
	test("เตียง", "dtiiang")
	test("เมีย", "miia")
	test("เรือ", "rʉʉa")
	test("นวด", "nûuat")
	test("ตัว", "dtuua")
	test("ไม่", "mâi")
	test("ใส่", "sài")
	test("วัย", "wai")
	test("ไทย", "tai")
	test("ไม้", "mái")
	test("หาย", "hǎai")
	test("ซอย", "sɔɔi")
	test("เลย", "ləəi")
	test("โดย", "dooi")
	test("ทุย", "tui")
	test("สวย", "sǔai")
	test("เรา", "rao")
	test("ขาว", "kǎao")
	test("แมว", "mɛɛo")
	test("เร็ว", "reo")
	test("หิว", "hǐu")
	test("เขียว", "kǐao")
	test("ทำ", "tam")
}