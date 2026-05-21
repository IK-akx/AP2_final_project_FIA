.PHONY: up down migrate-up migrate-down proto test test-unit test-integration

up:
	docker-compose up -d

down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5434/order_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5434/order_db?sslmode=disable" down

proto:
	protoc --proto_path=proto/order --proto_path=proto/product --go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		proto/order/order.proto
	protoc --proto_path=proto/product --go_out=proto/gen --go_opt=paths=source_relative \
		--go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
		proto/product/product.proto

test:
	go test ./tests/unit/... -v -count=1

test-integration:
	go test ./tests/integration/... -v -count=1 -tags=integration