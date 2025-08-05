package main

import "fmt"

func testSimple() {
	// Test some key improvements
	tests := []struct {
		thai string
		expected string
	}{
		{"นอน", "nɔɔn"},
		{"แดง", "dɛɛng"},
		{"โชค", "chôok"},
		{"ลูก", "lûuk"},
		{"เขียว", "kǐao"},
		{"ไทย", "tai"},
		{"พ่อ", "pɔ̂ɔ"},
		{"เดิน", "dəən"},
		{"ตก", "dtòk"},
		{"ความสุข", "kwaam-sùk"},
	}
	
	passed := 0
	total := len(tests)
	
	fmt.Println("\n=== Thai to Paiboon Transliteration Test ===")
	for _, test := range tests {
		result := ThaiToRoman(test.thai)
		if result == test.expected {
			fmt.Printf("✅ %s → %s\n", test.thai, result)
			passed++
		} else {
			fmt.Printf("❌ %s → %s (expected: %s)\n", test.thai, result, test.expected)
		}
	}
	
	fmt.Printf("\n=== Results ===")
	fmt.Printf("\nPassed: %d/%d (%.1f%% accuracy)\n", passed, total, float64(passed)*100/float64(total))
	
	// Test sentences
	fmt.Println("\n=== Sample Sentences ===")
	sentences := []string{
		"สวัสดีครับ",
		"ขอบคุณมาก",
		"อาหารไทยอร่อย",
		"ผมชอบเที่ยวเมืองไทย",
	}
	
	for _, s := range sentences {
		fmt.Printf("%s\n  → %s\n", s, ThaiToRoman(s))
	}
}