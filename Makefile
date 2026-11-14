.PHONY: build test install

build:
	go build -o dist/shipquiz ./cmd/shipquiz

test:
	go test ./...

install:
	mkdir -p "$(HOME)/.shipquiz/bin"
	go build -o "$(HOME)/.shipquiz/bin/shipquiz" ./cmd/shipquiz
