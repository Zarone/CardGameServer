build:
	go build -o bin/server ./cmd/server

run: build
	go run ./cmd/server

