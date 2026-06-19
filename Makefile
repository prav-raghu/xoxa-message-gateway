.PHONY: build test vet fmt lint run-api run-worker proto up down

build:
	go build ./...

test:
	go test ./... -v

vet:
	go vet ./...

fmt:
	gofmt -l .

run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/messenger.proto

up:
	docker compose up -d postgres nats

down:
	docker compose down
