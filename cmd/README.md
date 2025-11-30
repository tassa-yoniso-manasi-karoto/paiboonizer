# Paiboonizer Test CLI

## Build & Run

```bash
cd /home/voiduser/go/src/langkit/paiboonizer
go build -o paiboonizer-test ./cmd/main.go
./paiboonizer-test
```

Requires Docker (pythainlp container). First run initializes the container.

## Tests

| Test | Description | Metric |
|------|-------------|--------|
| **Corpus (translitkit)** | Full pipeline: pythainlp tokenization + paiboonizer (with dictionary) | Word-level % |
| **Corpus (pure rules)** | pythainlp tokenization + paiboonizer rules only (no dictionary) | Word-level % |
| **Dictionary** | Paiboonizer rules vs ~5000-word dictionary ground truth | Accuracy % |

## Test Files

```
testing_files/
├── test1.txt                            # Thai input
├── test1_Opus4.5_transliterated.txt     # Ground truth
├── test8.txt                            # More Thai input
├── test8_Opus4.5_transliterated.txt     # More ground truth
├── ...                                  # (auto-discovered testN.txt pairs)
├── draft_dictionary.tsv                 # Generated: words for LLM to transliterate
├── failures_translitkit.txt             # Generated: failure log
└── paiboon_examples.txt                 # Reference examples for LLM
```

## Output

Critical metrics displayed in bold/color. Corpus test writes all failures to `failures_translitkit.txt` for analysis.

The test also generates `draft_dictionary.tsv` containing Thai words that failed transliteration, ready for LLM processing.

---

## LLM Prompts for Corpus Generation

### Prompt: Transliterate test file (sentences)

```
I want you to create a transliteration file for test.txt called test_Opus4.5_transliterated.txt (create it as an artifact)

you will read the entire paiboon_examples.txt file first which contains official Paiboon+ transliterations to get a reminder of how it is done.

The file test_Opus4.5_transliterated.txt will be used as trusted ground truth in future test of a experimental transliteration engine I am working on so it is critical that it is accurate: ultrathink.

IMPORTANT: the transliteration showcased by paiboon_examples.txt showcases transliterations of words ONLY. When transliterating sentences like the ones in test.txt you should make sure to place space between words. For instance "kon-leeo-bɛ̀ɛp-níi-mâi-kuuan-dâi-ráp-gaan-hâi-à~pai" is incorrect but "kon leeo bɛ̀ɛp níi mâi kuuan dâi-ráp gaan-hâi-à~pai" would be correct.

Important constraints for this task:

1. CRITICAL : Proceed iteratively: Due to output token limits, you cannot write the entire transliteration in one response. Instead, work in batches - read the Paiboon+ examples to understand the system, then write the first ~50-100 lines to the file, then continue appending in subsequent responses until complete. Use the Write tool for the first batch, then Edit tool to append subsequent batches.

2. CRITICAL : Stay focused: Your thinking should be strictly focused on the mechanics of Paiboon+ transliteration (consonant mappings, vowel patterns, tone markers, syllable boundaries, special cases like ๆ, clusters, etc.). Do not analyze, summarize, or reflect on the semantic content/story of the Thai text - treat it purely as linguistic input to be transliterated character-by-character and syllable-by-syllable.
```

### Prompt: Fill in draft dictionary (words)

```
I want you to fill in the transliterations in draft_dictionary.tsv.

This is a TSV file where each line has a Thai word followed by a tab. Your task is to add the Paiboon+ romanization after each tab.

First, read paiboon_examples.txt which contains official Paiboon+ transliterations to understand the system (consonant mappings, vowel patterns, tone markers, syllable boundaries, special cases, clusters, etc.).

The completed dictionary will be used as a secondary source in a transliteration engine, so ensuring accuracy is critical.

Format: Each line should be `ThaiWord<TAB>romanization` with no trailing spaces.

Example input:
```
กระตือรือร้น
กระทู้
```

Example output:
```
กระตือรือร้น	grà~dtʉʉ-rʉʉ-rón
กระทู้	grà~túu
```

Important constraints:

1. CRITICAL: Proceed iteratively - work in batches of ~50-100 words at a time due to output token limits. Use Edit tool to update the file in batches.

2. CRITICAL: Stay focused on Paiboon+ mechanics only. Do not analyze word meanings - treat each word purely as linguistic input to be transliterated syllable-by-syllable.

3. Use hyphens between syllables within a word (e.g., "grà~dtʉʉ-rʉʉ-rón").

4. Use tilde (~) for unstressed/reduced syllables where appropriate (e.g., "grà~" not "grà-").
```
