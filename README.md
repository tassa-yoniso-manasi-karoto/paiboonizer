# Paiboonizer

Experimental Thai-to-Paiboon romanization library for translitkit integration.

## Overview

Paiboonizer provides Thai text transliteration to the Paiboon romanization system. It's designed to be used with pythainlp for word/syllable tokenization, achieving ~83% accuracy on arbitrary Thai text.

## Usage

### With translitkit (recommended)

Use the `paiboon-hybrid` scheme which combines pythainlp tokenization with paiboonizer transliteration:

```go
// In translitkit, the scheme handles the pipeline automatically
// pythainlp tokenizes → paiboonizer transliterates
```

### Direct API

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

## Accuracy

- **With pythainlp + dictionary**: ~85%+ (dictionary lookup first, then syllable rules)
- **With pythainlp + rules only**: ~83%
- **Pure rules (no pythainlp)**: ~70%

## Dependencies

- go-pythainlp for syllable tokenization (via Docker)
- Vocabulary embedded from CSV files at build time
