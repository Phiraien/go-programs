# Build stage
FROM golang:1.22 AS builder

WORKDIR /app

# Copy Go module files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o currency-converter .

# Runtime stage
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/currency-converter .

EXPOSE 8080

CMD ["./currency-converter"]
