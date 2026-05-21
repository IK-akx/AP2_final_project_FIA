.PHONY: up down migrate-up migrate-down proto test test-unit test-integration

up:
	docker-compose -f ../docker-compose.yml up -d postgres-order redis-order nats

down:
	docker-compose -f ../docker-compose.yml down

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5434/order_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5434/order_db?sslmode=disable" down

proto:
	cd ../pharmacy-proto && protoc --proto_path=. --go_out=../pharmacy-proto-gen --go_opt=paths=source_relative --go-grpc_out=../pharmacy-proto-gen --go-grpc_opt=paths=source_relative order/order.proto product/product.proto

test:
	go test ./tests/unit/... -v -count=1

test-integration:
	go test ./tests/integration/... -v -count=1 -tags=integration

run:
	go run cmd/main.go

build:
	go build -o bin/order-service cmd/main.go