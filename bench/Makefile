local:
	go build -o local-bench ./bench
	go build -o local-audience ./audience

isucon6f/bench:
	GOOS=linux go build -ldflags="-s -w" -o $@ ./bench

isucon6f/audience:
	GOOS=linux go build -ldflags="-s -w" -o $@ ./audience

isucon6f: isucon6q/bench isucon6q/audience

.PHONY: local
