.PHONY: build run-target tidy

build:
	go build -o bin/target ./cmd/target

run-target:
	go run ./cmd/target -c configs/target.yaml

tidy:
	go mod tidy
