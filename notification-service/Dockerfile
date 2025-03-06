FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o notification-service ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates netcat-openbsd
WORKDIR /root/
COPY --from=builder /app/notification-service .
COPY --from=builder /app/wait-for-kafka.sh .
RUN chmod +x wait-for-kafka.sh
EXPOSE 50051
ENTRYPOINT ["./wait-for-kafka.sh"]
