package main

import (
	"flag"
	"fmt"

	"github.com/tassa-yoniso-manasi-karoto/paiboonizer"
)

func main() {
	dictionaryCheck := flag.Bool("dictionary-check", false, "Test transliterator accuracy against dictionary")
	analyzeFlag := flag.Bool("analyze", false, "Analyze specific failure cases")
	noDictionary := flag.Bool("no-dict", false, "Disable dictionary lookup for testing pure rules")
	usePythainlp := flag.Bool("pythainlp", false, "Use pythainlp for syllable tokenization (requires Docker)")
	flag.Parse()

	if *dictionaryCheck {
		if *usePythainlp {
			// Initialize pythainlp with recreate to handle port mismatches
			fmt.Println("Initializing pythainlp...")
			if err := paiboonizer.InitPythainlpWithRecreate(true); err != nil {
				fmt.Printf("Failed to initialize pythainlp: %v\n", err)
				fmt.Println("Falling back to pure rules...")
				paiboonizer.TestDictionaryWithMode(paiboonizer.TestModePureRules)
			} else {
				defer paiboonizer.ClosePythainlp()
				paiboonizer.TestDictionaryWithMode(paiboonizer.TestModePythainlp)
			}
		} else if *noDictionary {
			paiboonizer.TestDictionaryWithMode(paiboonizer.TestModePureRules)
		} else {
			paiboonizer.TestDictionaryWithMode(paiboonizer.TestModeFullDictionary)
		}
		return
	}

	if *analyzeFlag {
		paiboonizer.AnalyzeFailures()
		return
	}

	// Debug mode: test specific words from command line args
	args := flag.Args()
	if len(args) > 0 {
		if *usePythainlp {
			fmt.Println("Initializing pythainlp for debug...")
			if err := paiboonizer.InitPythainlpWithRecreate(true); err != nil {
				fmt.Printf("Failed to initialize pythainlp: %v\n", err)
			} else {
				defer paiboonizer.ClosePythainlp()
			}
		}
		for _, word := range args {
			paiboonizer.DebugTransliteration(word)
		}
		return
	}

	fmt.Println("Use -dictionary-check, -analyze flags to run tests")
	fmt.Println("  -no-dict     : Test pure rule-based transliteration")
	fmt.Println("  -pythainlp   : Test with pythainlp syllable tokenization")
	fmt.Println("\nOr pass Thai words as arguments to debug their transliteration:")
	fmt.Println("  ./paiboonizer-test -pythainlp เหมือนกับ เลี้ยว")
}
