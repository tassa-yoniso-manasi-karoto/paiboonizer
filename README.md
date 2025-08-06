# Paiboonizer

Thai to Paiboon romanization transliterator

## Features

- Dictionary-based lookup for 4,981+ Thai words
- Rule-based transliteration for unknown words
- Comprehensive syllable parsing and tone rules
- Integration with pythainlp for improved tokenization
- 100% accuracy for dictionary words, 60% for arbitrary text

## Usage

### Main API

```go
// Transliterate a Thai word to Paiboon romanization
result := TransliterateWordRulesOnly("สวัสดี")
// Returns: "sà~wàt-dii"
```

### Advanced Functions

```go
// Extract syllables from a Thai word
syllables := ExtractSyllables("สวัสดี")

// Use comprehensive transliteration (no dictionary)
result := ComprehensiveTransliterate("ความสุข")
```

## Command Line Usage

```bash
# Test accuracy against dictionary
go run *.go -dictionary-check

# Transliterate test.txt file
go run *.go -test

# Analyze specific failure cases
go run *.go -analyze
```

## Dependencies

- go-pythainlp for Thai text tokenization
- Manual vocabulary files in `manual vocab/` directory

## Accuracy

- Dictionary words: 100% (using exact lookup)
- Unknown words: ~60% (using rule-based system)