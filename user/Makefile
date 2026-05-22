
.PHONY: migrate-up migrate-down proto test test-integration test-all clean run build docker-build docker-run docker-stop

DB_URL = postgres://postgres:postgres@localhost:5435/user_db?sslmode=disable
TEST_DB_URL = postgres://postgres:postgres@localhost:5435/user_test?sslmode=disable

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down

migrate-up-test:
	migrate -path migrations -database "$(TEST_DB_URL)" up

migrate-down-test:
	migrate -path migrations -database "$(TEST_DB_URL)" down

proto:
	mkdir -p proto/gen
	protoc --proto_path=proto \
	  --go_out=proto/gen --go_opt=paths=source_relative \
	  --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
	  proto/user.proto

test:
	go test ./internal/usecase/... -v -count=1 -cover

test-integration:
	$(MAKE) migrate-up-test
	go test ./tests/integration/... -v -count=1 -tags=integration
	$(MAKE) migrate-down-test

test-all: test test-integration

clean:
	rm -rf proto/gen
	go clean -testcache

run:
	go run cmd/main.go

build:
	go build -o bin/user-service cmd/main.go

docker-build:
	docker build -t user-service:latest .
	docker tag user-service:latest user-service:$(shell date +%Y%m%d-%H%M%S)

docker-run:
	docker run -d \
		--name user-service \
		-p 50053:50053 \
		--env-file .env \
		user-service:latest

docker-stop:
	docker stop user-service || true
	docker rm user-service || true

docker-up:
	docker-compose -f docker-compose.monitoring.yml up -d

docker-down:
	docker-compose -f docker-compose.monitoring.yml down

fmt:
	go fmt ./...
