local:
	go build -o local-bench ./cmd/bench
	go build -o local-audience ./cmd/audience

isucon6f/bench:
	GOOS=linux go build -ldflags="-s -w" -o $@ ./cmd/bench

isucon6f/audience:
	GOOS=linux go build -ldflags="-s -w" -o $@ ./cmd/audience

isucon6f: isucon6q/bench isucon6q/audience

.PHONY: local
