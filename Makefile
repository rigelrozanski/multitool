
build:
	go build ./cmd/...

install:
	go install ./cmd/...

test:
	go test ./cmd/...

.PHONY: build install test
