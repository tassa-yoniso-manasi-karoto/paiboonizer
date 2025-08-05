package main

import (
	"fmt"
)

func main() {
	// Test cases showing tokenization benefits
	testCases := []struct {
		input    string
		expected string
		description string
	}{
		{
			input:    "ไทย",
			expected: "tai",
			description: "Single word (dictionary)",
		},
		{
			input:    "ภาษาไทย",
			expected: "paa-sǎa-tai",
			description: "Compound word",
		},
		{
			input:    "ฉันรักเมืองไทย",
			expected: "chǎn-rák-mʉʉang-tai",
			description: "Simple sentence",
		},
		{
			input:    "สวัสดีครับ",
			expected: "sà~wàt-dii-kráp",
			description: "Common greeting",
		},
		{
			input:    "ขอบคุณมาก",
			expected: "kɔ̀ɔp-kun-mâak",
			description: "Thank you very much",
		},
	}
	
	fmt.Println("Thai to Paiboon Transliteration Tests")
	fmt.Println("======================================")
	fmt.Println("Testing with go-pythainlp tokenization:")
	fmt.Println()
	
	for _, tc := range testCases {
		result := ThaiToRoman(tc.input)
		status := "❌"
		if result == tc.expected {
			status = "✅"
		}
		fmt.Printf("%s %s\n", status, tc.description)
		fmt.Printf("   Input:    %s\n", tc.input)
		fmt.Printf("   Result:   %s\n", result)
		fmt.Printf("   Expected: %s\n", tc.expected)
		fmt.Println()
	}
	
	// Test with phrases from the manual vocab
	fmt.Println("Testing phrases from manual vocab (should use dictionary):")
	phrases := []string{
		"มีเสื้อผ้าที่สกปรกเยอะมาก",
		"ความสุข",
		"โรงเรียน",
		"ประเทศไทย",
		"อาหารไทย",
	}
	
	for _, phrase := range phrases {
		result := ThaiToRoman(phrase)
		fmt.Printf("   %s → %s\n", phrase, result)
	}
}