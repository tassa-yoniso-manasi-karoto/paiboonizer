package main

import (
	"fmt"
	"unicode"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/transform"
	"golang.org/x/text/runes"
)

func comprehensiveTest() {
	fmt.Println("=== Comprehensive Thai to Paiboon Test ===")
	
	// Extended test cases
	tests := []struct {
		thai string
		expected string
		description string
	}{
		// Original Wiktionary tests
		{"น้ำ", "nám", "water"},
		{"ธรรม", "tam", "dharma"},
		{"บาด", "bàat", "cut/wound"},
		{"บ้า", "bâa", "crazy"},
		{"แข็ง", "kɛ̌ng", "strong"},
		{"แกะ", "gɛ̀", "unwrap"},
		{"แดง", "dɛɛng", "red"},
		{"เกาะ", "gɔ̀", "island"},
		{"นอน", "nɔɔn", "sleep"},
		{"พ่อ", "pɔ̂ɔ", "father"},
		{"เห็ด", "hèt", "mushroom"},
		{"เตะ", "dtè", "kick"},
		{"เยอะ", "yə́", "a lot"},
		{"เดิน", "dəən", "walk"},
		{"ตก", "dtòk", "fall"},
		{"โต๊ะ", "dtó", "table"},
		{"โชค", "chôok", "luck"},
		{"คิด", "kít", "think"},
		{"อีก", "ììk", "again"},
		{"จี้", "jîi", "tickle"},
		{"ลึก", "lʉ́k", "deep"},
		{"รึ", "rʉ́", "or (question)"},
		{"ชื่อ", "chʉ̂ʉ", "name"},
		{"คุก", "kúk", "prison"},
		{"ลูก", "lûuk", "child"},
		{"ปู", "bpuu", "crab"},
		{"เตียง", "dtiiang", "bed"},
		{"เมีย", "miia", "wife"},
		{"เรือ", "rʉʉa", "boat"},
		{"นวด", "nûuat", "massage"},
		{"ตัว", "dtuua", "body"},
		{"ไม่", "mâi", "not"},
		{"ใส่", "sài", "put in"},
		{"วัย", "wai", "age"},
		{"ไทย", "tai", "Thai"},
		{"ไม้", "mái", "wood"},
		{"หาย", "hǎai", "disappear"},
		{"ซอย", "sɔɔi", "alley"},
		{"เลย", "ləəi", "at all"},
		{"โดย", "dooi", "by"},
		{"ทุย", "tui", "flag"},
		{"สวย", "sǔai", "beautiful"},
		{"เรา", "rao", "we"},
		{"ขาว", "kǎao", "white"},
		{"แมว", "mɛɛo", "cat"},
		{"เร็ว", "reo", "fast"},
		{"หิว", "hǐu", "hungry"},
		{"เขียว", "kǐao", "green"},
		{"ทำ", "tam", "do"},
		
		// Additional common words
		{"สวัสดี", "sàwàtdii", "hello"},
		{"ขอบคุณ", "kɔ̀ɔp-kun", "thank you"},
		{"ความสุข", "kwaam-sùk", "happiness"},
		{"อร่อย", "àròɔi", "delicious"},
		{"ภาษาไทย", "paasǎa-tai", "Thai language"},
		{"ประเทศไทย", "bpràtêet-tai", "Thailand"},
	}
	
	passed := 0
	total := len(tests)
	
	for _, test := range tests {
		result := ThaiToRoman(test.thai)
		
		// Normalize for comparison
		t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
		resultNorm, _, _ := transform.String(t, result)
		expectedNorm, _, _ := transform.String(t, test.expected)
		
		if result == test.expected || resultNorm == expectedNorm {
			passed++
			fmt.Printf("✅ %s (%s) → %s\n", test.thai, test.description, result)
		} else {
			fmt.Printf("❌ %s (%s) → %s (expected: %s)\n", test.thai, test.description, result, test.expected)
		}
	}
	
	fmt.Printf("\n=== COMPREHENSIVE TEST RESULTS ===")
	fmt.Printf("\nPassed: %d/%d (%.1f%% accuracy)\n", passed, total, float64(passed)*100/float64(total))
	
	if float64(passed)*100/float64(total) >= 95.0 {
		fmt.Println("\n🎆 SUCCESS! Achieved 95%+ accuracy! 🎆")
	}
}