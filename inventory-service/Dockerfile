FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
COPY vendor/ vendor/

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o inventory-service ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/inventory-service .
EXPOSE 50051
ENTRYPOINT ["./inventory-service"]
