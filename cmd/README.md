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
├── test.txt                         # Thai input (804 lines)
├── test_Opus4.5_transliterated.txt  # Ground truth
└── failures_translitkit.txt         # Generated failure log
```

## Output

Critical metrics displayed in bold/color. Corpus test writes all failures to `failures_translitkit.txt` for analysis.
