build:
	go build -o bin/main ./cmd/main

run: build
	go run ./cmd/main

