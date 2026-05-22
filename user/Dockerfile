
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o user-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/user-service .

COPY --from=builder /app/migrations ./migrations

COPY --from=builder /app/.env.example .env.example

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 50053

ENTRYPOINT ["./user-service"]
