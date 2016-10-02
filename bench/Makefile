local: seed/bindata.go
	go build -o local-bench ./cmd/bench
	go build -o local-audience ./cmd/audience
	go build -o local-worker ./cmd/worker

isucon6f/bench: seed/bindata.go
	GOOS=linux go build -ldflags="-s -w" -o $@ ./cmd/bench

isucon6f/audience:
	GOOS=linux go build -ldflags="-s -w" -o $@ ./cmd/audience

isucon6f/worker:
	GOOS=linux go build -ldflags="-s -w" -o $@ ./cmd/worker

isucon6f: isucon6f/bench isucon6f/audience isucon6f/worker

seed/bindata.go:
	$(MAKE) -C seed

.PHONY: local
