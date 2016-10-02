bindata.go: data/*.json
	go-bindata -pkg seed $^

test: bindata.go
	go test

.PHONY: bindata.go
