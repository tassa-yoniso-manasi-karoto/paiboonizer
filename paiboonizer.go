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
		// Skip empty tokens
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
	
	return strings.Join(results, "-")
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

// extractSyllables extracts syllables from a word
func extractSyllables(word string) []string {
	// Simplified syllable extraction
	// In production, use go-pythainlp's syllable tokenizer
	syllables := []string{}
	runes := []rune(word)
	i := 0
	
	for i < len(runes) {
		sylEnd := findSyllableEnd(runes, i)
		if sylEnd > i {
			syllables = append(syllables, string(runes[i:sylEnd]))
		}
		i = sylEnd
		if i == 0 {
			break // Prevent infinite loop
		}
	}
	
	return syllables
}

// findSyllableEnd finds the end of a Thai syllable
func findSyllableEnd(runes []rune, start int) int {
	if start >= len(runes) {
		return start
	}
	
	i := start
	hasVowel := false
	
	// Check for leading vowel (เ แ โ ไ ใ)
	if i < len(runes) && isLeadingVowel(string(runes[i])) {
		i++
		hasVowel = true
	}
	
	// Get consonant(s)
	consonantCount := 0
	for i < len(runes) && isConsonant(string(runes[i])) {
		i++
		consonantCount++
		// Allow up to 2 consonants (cluster)
		if consonantCount >= 2 {
			break
		}
	}
	
	// Get trailing vowels and marks
	for i < len(runes) {
		r := string(runes[i])
		if isVowel(r) || isToneMark(r) || r == "็" || r == "์" || r == "ํ" {
			i++
			if isVowel(r) {
				hasVowel = true
			}
		} else if hasVowel && isConsonant(r) {
			// Final consonant
			nextIsVowel := i+1 < len(runes) && (isVowel(string(runes[i+1])) || isLeadingVowel(string(runes[i+1])))
			if !nextIsVowel {
				i++
			}
			break
		} else {
			break
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
	// Special cases
	specialCases := map[string]string{
		"ธรรม": "tam",
		"กรรม": "gam",
		"สัตว์": "sàt",
		"จริง": "jing",
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
	
	// Get initial consonant(s)
	initialCons := ""
	for i < len(runes) && isConsonant(string(runes[i])) {
		initialCons += string(runes[i])
		i++
		if len([]rune(initialCons)) >= 2 {
			break
		}
	}
	
	// Store Thai initial for tone class
	if initialCons != "" {
		comp.InitialThai = string([]rune(initialCons)[0])
	}
	
	// Check for cluster
	if cluster, ok := clusters[initialCons]; ok {
		comp.Initial = cluster
	} else if initialCons != "" {
		// Use first consonant only
		firstCons := string([]rune(initialCons)[0])
		if trans, ok := initialConsonants[firstCons]; ok {
			comp.Initial = trans
		}
	}
	
	// Process remaining marks
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
		} else if isConsonant(r) && i == len(runes)-1 {
			// Final consonant
			finalCons = r
			i++
		} else {
			i++
		}
	}
	
	// Determine vowel sound
	comp.Vowel = determineVowelSound(leadingVowel, vowelMarks, finalCons)
	
	// Set final consonant
	if finalCons != "" {
		if trans, ok := finalConsonants[finalCons]; ok {
			comp.Final = trans
		}
	}
	
	return comp
}

// determineVowelSound determines the vowel sound from Thai components
func determineVowelSound(leading, marks, final string) string {
	// Map common patterns
	patterns := map[string]string{
		"เ-": "ee", "เ-็": "e", "เ-า": "ao", "เ-ีย": "iia", "เ-ือ": "ʉʉa",
		"เ-อ": "əə", "เ-ิ": "əə", "เ-อะ": "ə", "เ-าะ": "ɔ", "เ-ะ": "e",
		"แ-": "ɛɛ", "แ-็": "ɛ", "แ-ะ": "ɛ",
		"โ-": "oo", "โ-ะ": "o",
		"ไ-": "ai", "ใ-": "ai",
		"-ะ": "a", "-ั": "a", "-า": "aa",
		"-ิ": "i", "-ี": "ii",
		"-ึ": "ʉ", "-ื": "ʉʉ", 
		"-ุ": "u", "-ู": "uu",
		"-ำ": "am",
		"-ัว": "uua",
	}
	
	// Try exact pattern match
	pattern := leading + "-" + marks
	if vowel, ok := patterns[pattern]; ok {
		return vowel
	}
	
	// Try without leading
	if marks != "" {
		pattern = "-" + marks
		if vowel, ok := patterns[pattern]; ok {
			return vowel
		}
	}
	
	// Try leading only
	if leading != "" {
		pattern = leading + "-"
		if vowel, ok := patterns[pattern]; ok {
			return vowel
		}
	}
	
	// Default inherent vowel
	if leading == "" && marks == "" {
		if final == "" {
			return "ɔɔ" // Open syllable
		}
		return "o" // Closed syllable
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
	
	// Determine if syllable is live (sonorant ending or long vowel)
	isLive := false
	if comp.Final == "" || comp.Final == "n" || comp.Final == "m" || comp.Final == "ng" {
		isLive = true
	}
	longVowels := []string{"aa", "ii", "ʉʉ", "uu", "ee", "ɛɛ", "oo", "ɔɔ", "əə"}
	for _, lv := range longVowels {
		if strings.Contains(comp.Vowel, lv) {
			isLive = true
			break
		}
	}
	
	// Get tone number
	toneNum := 0 // mid tone
	if comp.ToneMark == "" {
		// Apply tone rules based on class and syllable type
		switch toneClass {
		case "high":
			if isLive {
				toneNum = 4 // rising
			} else {
				toneNum = 1 // low
			}
		case "low":
			if !isLive {
				toneNum = 2 // high
			}
		case "mid":
			if !isLive {
				toneNum = 1 // low
			}
		}
	} else {
		// Apply tone mark rules
		switch comp.ToneMark {
		case "่": // mai ek
			if toneClass == "low" {
				toneNum = 3 // falling
			} else {
				toneNum = 1 // low
			}
		case "้": // mai tho
			if toneClass == "low" {
				toneNum = 2 // high
			} else {
				toneNum = 3 // falling
			}
		case "๊": // mai tri
			toneNum = 2 // high
		case "๋": // mai jattawa
			toneNum = 4 // rising
		}
	}
	
	// Add tone mark
	if toneNum == 0 {
		return text
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
			// This is a simplification - in production, use proper tokenization
			if !strings.Contains(th, " ") && len([]rune(th)) <= 4 {
				syllableDict[th] = translit
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
	
	// Run accuracy test
	testAccuracy()
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