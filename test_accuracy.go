package main

import (
	"fmt"
	"strings"
)

func testAccuracy() {
	// Test on Wiktionary cases (mostly NOT in dictionary)
	wiktionaryTests := []struct {
		thai     string
		expected string
	}{
		{"น้ำ", "nám"},
		{"ธรรม", "tam"},
		{"บาด", "bàat"},
		{"บ้า", "bâa"},
		{"แข็ง", "kɛ̌ng"},
		{"แกะ", "gɛ̀"},
		{"แดง", "dɛɛng"},
		{"เกาะ", "gɔ̀"},
		{"นอน", "nɔɔn"},
		{"พ่อ", "pɔ̂ɔ"},
		{"เห็ด", "hèt"},
		{"เตะ", "dtè"},
		{"เยอะ", "yə́"},
		{"เดิน", "dəən"},
		{"ตก", "dtòk"},
		{"โต๊ะ", "dtó"},
		{"โชค", "chôok"},
		{"คิด", "kít"},
		{"อีก", "ìik"},
		{"จี้", "jîi"},
		{"ลึก", "lʉ́k"},
		{"รึ", "rʉ́"},
		{"ชื่อ", "chʉ̂ʉ"},
		{"คุก", "kúk"},
		{"ลูก", "lûuk"},
		{"ปู", "bpuu"},
		{"เตียง", "dtiiang"},
		{"เมีย", "miia"},
		{"เรือ", "rʉʉa"},
		{"นวด", "nûuat"},
		{"ตัว", "dtuua"},
		{"ไม่", "mâi"},
		{"ใส่", "sài"},
		{"วัย", "wai"},
		{"ไทย", "tai"},
		{"ไม้", "mái"},
		{"หาย", "hǎai"},
		{"ซอย", "sɔɔi"},
		{"เลย", "ləəi"},
		{"โดย", "dooi"},
		{"ทุย", "tui"},
		{"สวย", "sǔai"},
		{"เรา", "rao"},
		{"ขาว", "kǎao"},
		{"แมว", "mɛɛo"},
		{"เร็ว", "reo"},
		{"หิว", "hǐu"},
		{"เขียว", "kǐao"},
		{"ทำ", "tam"},
	}
	
	passCount := 0
	failCount := 0
	
	fmt.Println("\n=== Testing on Wiktionary words (NOT using dictionary) ===")
	for _, test := range wiktionaryTests {
		// Skip if it's in the dictionary (to test actual transliteration)
		if _, inDict := dictionary[test.thai]; inDict {
			continue
		}
		
		result := ThaiToRoman(test.thai)
		// Remove tone marks for comparison
		resultClean := removeToneMarks(result)
		expectedClean := removeToneMarks(test.expected)
		
		if result == test.expected {
			passCount++
			fmt.Printf("✅ %s → %s\n", test.thai, result)
		} else if resultClean == expectedClean {
			failCount++
			fmt.Printf("⚠️  %s → %s (expected: %s) [tone error]\n", test.thai, result, test.expected)
		} else {
			failCount++
			fmt.Printf("❌ %s → %s (expected: %s)\n", test.thai, result, test.expected)
		}
	}
	
	accuracy := float64(passCount) * 100 / float64(passCount+failCount)
	fmt.Printf("\n=== ACTUAL ACCURACY (non-dictionary words) ===\n")
	fmt.Printf("Passed: %d, Failed: %d\n", passCount, failCount)
	fmt.Printf("Accuracy: %.1f%%\n\n", accuracy)
	
	// Test on arbitrary sentences
	fmt.Println("=== Testing on arbitrary Thai sentences ===")
	sentences := []struct {
		thai     string
		description string
	}{
		{"วันนี้อากาศดีมาก", "Today the weather is very good"},
		{"ฉันชอบกินอาหารไทย", "I like to eat Thai food"},
		{"เขาไปโรงเรียนทุกวัน", "He goes to school every day"},
		{"คุณพูดภาษาอังกฤษได้ไหม", "Can you speak English?"},
		{"ร้านนี้ขายของถูกมาก", "This shop sells things very cheap"},
	}
	
	for _, s := range sentences {
		result := ThaiToRoman(s.thai)
		fmt.Printf("%s\n", s.description)
		fmt.Printf("  Thai: %s\n", s.thai)
		fmt.Printf("  Result: %s\n\n", result)
	}
}

func removeToneMarks(s string) string {
	// Remove common tone marks for comparison
	s = strings.ReplaceAll(s, "̀", "")
	s = strings.ReplaceAll(s, "́", "")
	s = strings.ReplaceAll(s, "̂", "")
	s = strings.ReplaceAll(s, "̌", "")
	return s
}

// Run this test by calling testRealAccuracy() from main()
// or compile this file separately with the main package