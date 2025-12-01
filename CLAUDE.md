# Paiboonizer - LLM Context

## Architecture: go-pythainlp × paiboonizer

**Paiboonizer is a "dumb" transliterator.** It does not perform word segmentation—it receives pre-tokenized words from go-pythainlp and transliterates them syllable-by-syllable.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         TRANSLITKIT PIPELINE                            │
│                                                                         │
│   Thai text ──► go-pythainlp ──► paiboonizer ──► Romanized output      │
│                (tokenization)    (transliteration)                      │
│                                                                         │
│   "แม่งเอ๊ย" ──► ["แม่ง", "เอ๊ย"] ──► "mɛ̂ɛng əi"    ✓ CORRECT         │
│   "แม่งเอ๊ย" ──► ["แม่", "ง", "เอ๊ย"] ──► "mɛ̂ɛ ngɔɔ ée"  ✗ WRONG      │
│                    ↑                                                    │
│                    └── pythainlp segmentation error                     │
└─────────────────────────────────────────────────────────────────────────┘
```

## Where errors come from

The measured accuracy reflects **multiple potential error sources**:

### 1. pythainlp segmentation errors (tokenization)

Many failures are caused by pythainlp's faulty closing consonant detection, not paiboonizer's transliteration rules.

| Thai | Expected | Got | Issue |
|------|----------|-----|-------|
| แม่งเอ๊ย | mɛ̂ng əi | mɛ̂ɛ ngɔɔ ée | "ง" split as separate word |
| เราแค่อยากบอกว่า | ...bɔ̀ɔk wâa | ...bɔɔ gwàa | "ก" split from บอก |

When pythainlp segments "แม่ง" as ["แม่", "ง"], paiboonizer correctly transliterates each token: "แม่" → "mɛ̂ɛ" and "ง" → "ngɔɔ". The error is in the segmentation, not the transliteration.

### 2. opus_dictionary.tsv errors

The dictionary was LLM-generated and may contain wrong vowel lengths or tones. Always **ask the user to verify against the official Paiboon app** before assuming the engine is wrong.

**IMPORTANT:** If the engine output seems wrong, first check if the word exists in `opus_dictionary.tsv` with an incorrect transliteration. Fix the dictionary entry, don't add workarounds.

### 3. Ground truth errors (test files)

The `*_Opus4.5_transliterated.txt` files were LLM-generated and can occasionally contain systematic errors (wrong vowel lengths, tones, etc.). The official Paiboon app is the source of truth.

**IMPORTANT:** When fixing ground truth with sed, **always use word boundaries** to avoid partial matches:
```bash
# CORRECT - with word boundaries
sed -i 's/ wrongWord / correctWord /g' file.txt
sed -i 's/^wrongWord /correctWord /g' file.txt
sed -i 's/ wrongWord$/ correctWord/g' file.txt

# WRONG - will corrupt other words containing the pattern
sed -i 's/wrongWord/correctWord/g' file.txt
```

## Where to patch word segmentation

If word segmentation needs to be fixed/patched, do it in **translitkit**, not paiboonizer:

```
/home/voiduser/go/src/langkit/translitkit/lang/tha/pythainlp.go
→ ProcessFlowController()
```

This is where pythainlp's tokenization output can be post-processed before being passed to paiboonizer. For example, you could merge orphan closing consonants back into the preceding word.

## Debugging transliteration failures

1. **First check**: Is the word segmented correctly by pythainlp?
2. **If segmentation is wrong**: Patch in `translitkit/lang/tha/pythainlp.go` (ProcessFlowController)
3. **If segmentation is correct but transliteration is wrong**: Fix in paiboonizer (dictionary or rules)

The test suite (`cmd/main.go`) outputs failures to `testing_files/failures_translitkit.txt` for analysis.

## Test commands

```bash
# Build and run full test suite
cd /home/voiduser/go/src/langkit/paiboonizer
go build -o paiboonizer-test ./cmd/main.go
./paiboonizer-test

# Quick accuracy check
./paiboonizer-test -dictionary-check -pythainlp 2>&1 | grep -E "^(REAL ACCURACY|fallback)"
```

## Key files

- `paiboonizer.go` - Main dictionary lookup and rule entry points
- `paiboonizer_comprehensive.go` - Rule-based transliteration engine
- `paiboonizer_improved.go` - Tone calculation logic
- `cmd/main.go` - Test suite
- `testing_files/failures_translitkit.txt` - Generated failure log for analysis
