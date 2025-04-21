FROM golang:1.24-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git ca-certificates tzdata && \
    update-ca-certificates

WORKDIR /app

# Copy go.mod and go.sum first to leverage layer caching
COPY go.mod go.sum ./

# Copy vendor directory for dependencies
COPY vendor/ vendor/

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=vendor \
    -ldflags="-w -s" \
    -o user-service ./cmd/server

FROM alpine:3.19

# Add necessary runtime packages
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/user-service .

# Expose ports
EXPOSE 50051

# Health check
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget -q --spider http://localhost:50051/health || exit 1 || true

# Run the binary
CMD ["./user-service"]
