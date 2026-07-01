.PHONY: build run-coordinator run-worker run-target test-local test tidy

build:
	go build -o bin/dlt ./cmd/dlt
	go build -o bin/target ./cmd/target

run-coordinator:
	go run ./cmd/dlt coordinator --config configs/coordinator.yaml

run-worker:
	go run ./cmd/dlt worker --config configs/worker.yaml

run-target:
	go run ./cmd/target -c configs/target.yaml

test-local:
	go run ./cmd/dlt test --coordinator-config configs/coordinator.yaml --worker-config configs/worker.yaml --workers 1

test:
	go test -race ./...

tidy:
	go mod tidy
