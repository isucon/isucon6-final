bench:
	go build -o local-bench ./cmd/bench

isucon6f/bench:
	GOOS=linux go build -ldflags="-s -w" -o $@ ./cmd/bench

isucon6f: isucon6q/bench

.PHONY: cmd
