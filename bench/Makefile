local:
	go build -o bench

isucon6f/isucon6f-bench:
	GOOS=linux go build -ldflags="-s -w" -o $@

isucon6f: isucon6q/isucon6q-bench

.PHONY: local
