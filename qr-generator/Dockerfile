# ---------- Build Stage ----------
FROM golang:1.22 AS builder

WORKDIR /app

# Copy module files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o qr-generator .

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/qr-generator .

# Expose application port
EXPOSE 8080

# Run application
CMD ["./qr-generator"]
