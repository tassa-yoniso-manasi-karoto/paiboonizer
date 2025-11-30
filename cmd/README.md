# Paiboonizer Test CLI

## Build & Run

```bash
cd /home/voiduser/go/src/langkit/paiboonizer
go build -o paiboonizer-test ./cmd/main.go
./paiboonizer-test
```

## What It Tests

1. **Corpus Test (translitkit)**: Tests transliteration against 804-line ground truth corpus using the full translitkit pipeline (pythainlp tokenization + paiboonizer). Reports word-level and line-level accuracy. Failures written to `testing_files/failures_translitkit.txt`.

2. **Dictionary Test**: Tests paiboonizer's rule-based transliteration against its ~5000-word dictionary using pythainlp for syllable segmentation. This measures how well the rules match known correct transliterations.

## Test Files

- `testing_files/test.txt` - Thai input corpus (804 lines)
- `testing_files/test_Opus4.5_transliterated.txt` - Ground truth transliterations
- `testing_files/failures_translitkit.txt` - Generated failure log

## Requirements

- Docker (for pythainlp container)
- First run may take longer as the container initializes
