FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

ENV GOPROXY=https://goproxy.io,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o product-service ./cmd/product-service/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/product-service .

EXPOSE 50051

CMD ["./product-service"]