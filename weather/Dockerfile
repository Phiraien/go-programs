# ---------- Build Stage ----------
FROM golang:1.22 AS builder

WORKDIR /app

# Copy Go module files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o weather-service .

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app

# Copy compiled binary
COPY --from=builder /app/weather-service .

# Expose application port
EXPOSE 8082

# Run application
CMD ["./weather-service"]
