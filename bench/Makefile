bench:
	go build -o local-bench ./bench

isucon6f/bench:
	GOOS=linux go build -ldflags="-s -w" -o $@ ./bench

isucon6f: isucon6q/bench

.PHONY: bench
