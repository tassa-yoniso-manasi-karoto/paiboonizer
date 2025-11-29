cd /home/voiduser/go/src/langkit/paiboonizer && go build -o paiboonizer-test ./cmd/main.go && ./paiboonizer-test -dictionary-check -pythainlp 2>&1  | grep -E "^(REAL ACCURACY|fallback)"
