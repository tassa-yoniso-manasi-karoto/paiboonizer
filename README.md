Experimental Thai-to-Paiboon romanization library for translitkit integration. It is designed for replacing the scraper of thai2english that translitkit uses to support paiboon transliteration. It is meant to be used with pythainlp for word/syllable tokenization.

## Accuracy

Against:
- **CORPUS: With pythainlp word tokenization + dictionary + syllable segmentation of pythainlp**: ~83%
- CORPUS: With pythainlp word tokenization + *pure Golang rules only*: ~37%
- **DICTIONARY: pure Golang rules only**: ~84.5%

Corpus is 11714 thai sentences extracted from subtitle files.

Ground truth is a transliteration of this corpus made by Claude Opus 4.5.

## Usage

### 👉 With [translitkit](https://github.com/tassa-yoniso-manasi-karoto/translitkit) (RECOMMENDED FOR BEST ACCURACY) 👈

Use the `paiboon-hybrid` scheme which combines pythainlp tokenization with paiboonizer transliteration:

```go
// In translitkit, the scheme handles the pipeline automatically
// pythainlp tokenizes → paiboonizer transliterates
```

### Direct API (good enough for transliterating individual words)

```go
import "github.com/tassa-yoniso-manasi-karoto/paiboonizer"

// Check word dictionary first (~5000 entries)
if trans, found := paiboonizer.LookupDictionary("หน้าต่าง"); found {
    // Returns "nâa-dtàang"
}

// Check syllable dictionary
if trans, found := paiboonizer.LookupSyllable("สวัส"); found {
    // ...
}

// Rule-based transliteration (fallback)
result := paiboonizer.ComprehensiveTransliterate("ความสุข")

// Helper for silent consonant markers (์)
clean := paiboonizer.RemoveSilentConsonants("สันต์") // Returns "สัน"
```

## Dependencies

- go-pythainlp for syllable tokenization (via Docker)
- Vocabulary embedded from CSV files at build time
