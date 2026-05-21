.PHONY: migrate-up migrate-down proto test test-integration

migrate-up:
	migrate -path migrations \
	  -database "postgres://postgres:postgres@localhost:5435/user_db?sslmode=disable" up

migrate-down:
	migrate -path migrations \
	  -database "postgres://postgres:postgres@localhost:5435/user_db?sslmode=disable" down

proto:
	mkdir -p proto/gen
	protoc --proto_path=proto \
	  --go_out=proto/gen --go_opt=paths=source_relative \
	  --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
	  proto/user.proto

test:
	go test ./internal/usecase/... -v -count=1

test-integration:
	go test ./tests/integration/... -v -count=1 -tags=integration