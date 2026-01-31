# Build stage
FROM golang:1.24.11-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (including generated code in gen/)
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o slips-core cmd/server/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary and config from builder
COPY --from=builder /app/slips-core .
COPY --from=builder /app/config.yaml .

EXPOSE 9090

CMD ["./slips-core"]
