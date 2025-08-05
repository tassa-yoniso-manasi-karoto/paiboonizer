package main

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/gookit/color"
	"github.com/k0kubun/pp"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Consonant classes for tone determination
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

// Initial consonant transliterations
var initialConsonants = map[string]string{
	"ก": "g", "ข": "k", "ฃ": "k", "ค": "k", "ฅ": "k", "ฆ": "k", "ง": "ng",
	"จ": "j", "ฉ": "ch", "ช": "ch", "ซ": "s", "ฌ": "ch", "ญ": "y", "ฎ": "d",
	"ฏ": "dt", "ฐ": "t", "ฑ": "t", "ฒ": "t", "ณ": "n", "ด": "d", "ต": "dt",
	"ถ": "t", "ท": "t", "ธ": "t", "น": "n", "บ": "b", "ป": "bp", "ผ": "p",
	"ฝ": "f", "พ": "p", "ฟ": "f", "ภ": "p", "ม": "m", "ย": "y", "ร": "r",
	"ล": "l", "ว": "w", "ศ": "s", "ษ": "s", "ส": "s", "ห": "h", "ฬ": "l",
	"อ": "", "ฮ": "h",
}

// Final consonant transliterations
var finalConsonants = map[string]string{
	"ก": "k", "ข": "k", "ฃ": "k", "ค": "k", "ฅ": "k", "ฆ": "k", "ง": "ng",
	"จ": "t", "ฉ": "t", "ช": "t", "ซ": "t", "ฌ": "t", "ญ": "n", "ฎ": "t",
	"ฏ": "t", "ฐ": "t", "ฑ": "t", "ฒ": "t", "ณ": "n", "ด": "t", "ต": "t",
	"ถ": "t", "ท": "t", "ธ": "t", "น": "n", "บ": "p", "ป": "p", "ผ": "p",
	"ฝ": "p", "พ": "p", "ฟ": "p", "ภ": "p", "ม": "m", "ย": "i", "ร": "n",
	"ล": "n", "ว": "o", "ศ": "t", "ษ": "t", "ส": "t", "ห": "", "ฬ": "n",
	"อ": "", "ฮ": "",
}

// Consonant clusters
var clusters = map[string]string{
	"กร": "gr", "กล": "gl", "กว": "gw",
	"ขร": "kr", "ขล": "kl", "ขว": "kw",
	"คร": "kr", "คล": "kl", "คว": "kw",
	"ปร": "bpr", "ปล": "bpl",
	"พร": "pr", "พล": "pl",
	"ผล": "pl",
	"ฟร": "fr", "ฟล": "fl",
	"ตร": "dtr", "ทร": "s", // ทร is special
	"ดร": "dr",
}

// Special words
var specialWords = map[string]string{
	"ธรรม": "tam",
	"ธรรมะ": "tam-má",
	"จริง": "jing",
	"สัตว์": "sàt",
}

func isConsonant(r string) bool {
	return strings.Contains("กขฃคฅฆงจฉชซฌญฎฏฐฑฒณดตถทธนบปผฝพฟภมยรลวศษสหฬอฮ", r)
}

func isVowel(r string) bool {
	return strings.Contains("ะัาิีึืุูเแโใไ็ำๅ", r)
}

func isToneMark(r string) bool {
	return strings.Contains("่้๊๋", r)
}

func isLeadingVowel(r string) bool {
	return r == "เ" || r == "แ" || r == "โ" || r == "ไ" || r == "ใ"
}

// Syllable represents a Thai syllable
type Syllable struct {
	LeadingVowel  string
	Initial       string
	Vowel         string
	Final         string
	ToneMark      string
	Raw           string
}

func ThaiToRoman(text string) string {
	text = strings.TrimSpace(text)
	
	// Check for special words
	if trans, ok := specialWords[text]; ok {
		return trans
	}
	
	// Process the text
	syllables := parseSyllables(text)
	results := []string{}
	
	for _, syl := range syllables {
		if trans := transliterateSyllable(syl); trans != "" {
			results = append(results, trans)
		}
	}
	
	// Join syllables with hyphen for multi-syllable words
	if len(results) > 1 {
		return strings.Join(results, "-")
	} else if len(results) == 1 {
		return results[0]
	}
	return ""
}

func parseSyllables(text string) []Syllable {
	syllables := []Syllable{}
	runes := []rune(text)
	i := 0
	
	for i < len(runes) {
		syl := Syllable{}
		start := i
		
		// Check for leading vowel
		if i < len(runes) && isLeadingVowel(string(runes[i])) {
			syl.LeadingVowel = string(runes[i])
			i++
		}
		
		// Get initial consonant(s)
		consonants := ""
		for i < len(runes) && isConsonant(string(runes[i])) {
			consonants += string(runes[i])
			i++
			// Check for phinthu (subscript marker)
			if i < len(runes) && string(runes[i]) == "ฺ" {
				i++
			}
			// Stop after 2 consonants unless it's a cluster
			if len([]rune(consonants)) >= 2 {
				break
			}
		}
		
		// Process consonants - check for clusters
		if cluster, ok := clusters[consonants]; ok {
			syl.Initial = cluster
		} else if consonants != "" {
			// Take first consonant as initial
			consRunes := []rune(consonants)
			syl.Initial = string(consRunes[0])
			// If there's a second consonant and no following vowel, it might be final
			if len(consRunes) > 1 {
				// Put it back for processing
				i -= len([]rune(string(consRunes[1])))
			}
		}
		
		// Get vowels and tone marks
		hasVowel := syl.LeadingVowel != ""
		for i < len(runes) {
			r := string(runes[i])
			if isVowel(r) {
				syl.Vowel += r
				hasVowel = true
				i++
			} else if isToneMark(r) {
				syl.ToneMark = r
				i++
			} else if r == "์" { // thanthakhat (silencer)
				i++
				// Skip this consonant
			} else if r == "ํ" { // nikkhahit
				i++
			} else if r == "็" { // mai taikhu
				syl.Vowel += r
				i++
			} else if isConsonant(r) && hasVowel {
				// Could be final consonant
				nextIsVowel := i+1 < len(runes) && (isVowel(string(runes[i+1])) || isLeadingVowel(string(runes[i+1])))
				if !nextIsVowel {
					syl.Final = r
					i++
				}
				break
			} else {
				break
			}
		}
		
		// If we haven't moved, force advance to avoid infinite loop
		if i == start {
			i++
		}
		
		// Store raw for debugging
		syl.Raw = string(runes[start:i])
		syllables = append(syllables, syl)
	}
	
	return syllables
}

func transliterateSyllable(syl Syllable) string {
	result := ""
	
	// Get initial consonant sound
	if syl.Initial != "" {
		if strings.Contains(syl.Initial, "r") || strings.Contains(syl.Initial, "l") || strings.Contains(syl.Initial, "w") {
			// It's already a cluster transliteration
			result = syl.Initial
		} else if trans, ok := initialConsonants[syl.Initial]; ok {
			result = trans
		} else {
			result = syl.Initial
		}
	}
	
	// Process vowel
	vowelSound := getVowelSound(syl)
	result += vowelSound
	
	// Add final consonant
	if syl.Final != "" {
		if trans, ok := finalConsonants[syl.Final]; ok && trans != "" {
			// Don't add if vowel already includes it
			if !strings.HasSuffix(vowelSound, trans) {
				result += trans
			}
		}
	}
	
	// Apply tone
	result = applyTone(result, syl)
	
	return result
}

func getVowelSound(syl Syllable) string {
	// Combine leading vowel and vowel marks
	fullVowel := syl.LeadingVowel + syl.Vowel
	
	// Handle specific vowel patterns
	switch {
	// Short vowels with mai taikhu
	case syl.LeadingVowel == "เ" && strings.Contains(syl.Vowel, "็"):
		return "e"
	case syl.LeadingVowel == "แ" && strings.Contains(syl.Vowel, "็"):
		return "ɛ"
	
	// Complex vowels
	case fullVowel == "เีย":
		return "iia"
	case fullVowel == "เือ":
		return "ʉʉa"
	case fullVowel == "เียะ":
		return "ia"
	case fullVowel == "เือะ":
		return "ʉa"
	case fullVowel == "เา":
		return "ao"
	case fullVowel == "เาะ":
		return "ɔ"
	case fullVowel == "เอะ":
		return "ə"
	case fullVowel == "เอ":
		return "əə"
	case syl.LeadingVowel == "เ" && syl.Vowel == "" && syl.Final != "":
		return "ee"
	case fullVowel == "เะ":
		return "e"
	case syl.LeadingVowel == "เ" && syl.Vowel == "":
		return "ee"
	
	// แ vowels
	case syl.LeadingVowel == "แ" && syl.Vowel == "" && syl.Final != "":
		return "ɛɛ"
	case fullVowel == "แะ":
		return "ɛ"
	case syl.LeadingVowel == "แ" && syl.Vowel == "":
		return "ɛɛ"
	
	// โ vowels
	case fullVowel == "โะ":
		return "o"
	case syl.LeadingVowel == "โ" && syl.Vowel == "":
		return "oo"
	
	// ไ ใ vowels
	case syl.LeadingVowel == "ไ":
		return "ai"
	case syl.LeadingVowel == "ใ":
		return "ai"
	
	// Simple vowels
	case syl.Vowel == "ะ" || syl.Vowel == "ั":
		return "a"
	case syl.Vowel == "า":
		return "aa"
	case syl.Vowel == "ิ":
		return "i"
	case syl.Vowel == "ี":
		return "ii"
	case syl.Vowel == "ึ":
		return "ʉ"
	case syl.Vowel == "ื":
		return "ʉʉ"
	case syl.Vowel == "ุ":
		return "u"
	case syl.Vowel == "ู":
		return "uu"
	case syl.Vowel == "ำ":
		return "am"
	case syl.Vowel == "ัว":
		return "uua"
	case syl.Vowel == "็":
		return "e"
	
	// Vowel + final combinations
	case syl.Vowel == "อ" && syl.Final == "ย":
		return "ɔɔi"
	case syl.Vowel == "า" && syl.Final == "ย":
		return "aai"
	case syl.Vowel == "า" && syl.Final == "ว":
		return "aao"
	case syl.Vowel == "ิ" && syl.Final == "ว":
		return "iu"
	case syl.LeadingVowel == "เ" && syl.Final == "ว":
		return "eeo"
	case syl.LeadingVowel == "แ" && syl.Final == "ว":
		return "ɛɛo"
	case syl.Vowel == "ั" && syl.Final == "ย":
		return "ai"
	case syl.Vowel == "ุ" && syl.Final == "ย":
		return "ui"
	case syl.LeadingVowel == "โ" && syl.Final == "ย":
		return "ooi"
	case syl.LeadingVowel == "เ" && syl.Final == "ย":
		return "əəi"
	
	// No explicit vowel - inherent vowel
	case syl.Vowel == "" && syl.LeadingVowel == "":
		if syl.Final == "" {
			// Open syllable with no vowel mark
			if syl.Initial != "" {
				return "ɔɔ"
			}
		} else {
			// Closed syllable with no vowel mark
			return "o"
		}
	}
	
	return ""
}

func applyTone(text string, syl Syllable) string {
	// Determine tone class
	toneClass := getToneClass(syl)
	
	// Determine if syllable is live or dead
	isLive := isSyllableLive(syl)
	
	// Get tone number based on tone mark and rules
	toneNum := getToneNumber(syl.ToneMark, toneClass, isLive)
	
	// Apply tone mark to the text
	return addToneMark(text, toneNum)
}

func getToneClass(syl Syllable) string {
	// Get the first consonant for tone class
	firstCons := ""
	if syl.Initial != "" {
		// For clusters, check the first character of the Thai original
		// This is simplified - we're using the transliteration
		if len(syl.Initial) > 1 {
			// It's a cluster, need to check the original Thai
			// For now, use the transliteration to guess
			switch syl.Initial[0] {
			case 'k':
				return "high"
			case 'g':
				return "mid"
			case 'p':
				return "high"
			case 'b':
				return "mid"
			case 't':
				return "high"
			case 'd':
				return "mid"
			default:
				return "low"
			}
		}
		firstCons = syl.Initial
	}
	
	// Check consonant class
	if highClass[firstCons] {
		return "high"
	} else if midClass[firstCons] {
		return "mid"
	} else if lowClass[firstCons] {
		return "low"
	}
	
	return "mid"
}

func isSyllableLive(syl Syllable) bool {
	// Live syllables end in sonorants or long vowels
	if syl.Final == "" {
		// Check if it has a long vowel
		vowel := getVowelSound(syl)
		longVowels := []string{"aa", "ii", "ʉʉ", "uu", "ee", "ɛɛ", "oo", "ɔɔ", "əə", "iia", "ʉʉa", "uua", "aao", "iao", "eeo", "ɛɛo"}
		for _, lv := range longVowels {
			if vowel == lv {
				return true
			}
		}
		return false
	}
	
	// Check final consonant
	sonorants := []string{"ง", "น", "ม", "ย", "ว", "ล", "ร"}
	for _, s := range sonorants {
		if syl.Final == s {
			return true
		}
	}
	
	return false
}

func getToneNumber(toneMark string, toneClass string, isLive bool) int {
	// Thai tone rules
	// Return: 0=mid, 1=low, 2=high, 3=falling, 4=rising
	
	if toneMark == "" {
		// No tone mark
		switch toneClass {
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
		case "low":
			if isLive {
				return 0 // mid
			}
			return 2 // high
		}
	}
	
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
		return 2 // high
	case "๋": // mai jattawa
		return 4 // rising
	}
	
	return 0 // mid (default)
}

func addToneMark(text string, toneNum int) string {
	if toneNum == 0 {
		return text // mid tone, no mark
	}
	
	marks := map[int]string{
		1: "\u0300", // grave (low)
		2: "\u0301", // acute (high)
		3: "\u0302", // circumflex (falling)
		4: "\u030C", // caron (rising)
	}
	
	mark := marks[toneNum]
	
	// Find the main vowel to add the tone mark
	runes := []rune(text)
	for i, r := range runes {
		if isVowelChar(r) {
			// Insert tone mark after this vowel
			result := string(runes[:i+1]) + mark + string(runes[i+1:])
			return result
		}
	}
	
	// If no vowel found, return as is
	return text
}

func isVowelChar(r rune) bool {
	vowelChars := "aeiouəɛɔʉ"
	return strings.ContainsRune(vowelChars, r)
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
	pp.Println("Init")
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
			words = append(words, th)
			m[th] = translit
		}
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	fmt.Println("Testing with", len(words), "words from manual vocab")
	
	// First test with Wiktionary test cases
	fmt.Println("\n=== Wiktionary Test Cases ===")
	testWiktionary()
	
	// Then test with loaded vocabulary
	fmt.Println("\n=== Manual Vocab Test Cases ===")
	passCount := 0
	failCount := 0
	for i, th := range words {
		if i > 100 { // Limit for initial testing
			break
		}
		r := ThaiToRoman(th)
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		r, _, _ = transform.String(t, r)
		expected := m[th]
		if r == expected {
			passCount++
		} else {
			failCount++
			if failCount <= 20 { // Show first 20 failures
				fmt.Println(isPassed(r, expected), th, "Got:", r, "Expected:", expected)
			}
		}
	}
	fmt.Printf("\nResults: %d passed, %d failed (%.1f%% accuracy)\n", 
		passCount, failCount, float64(passCount)*100/float64(passCount+failCount))
}

func testWiktionary() {
	// Test cases from original comments
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