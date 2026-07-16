# -------- Build Stage --------
FROM golang:1.22 AS builder

WORKDIR /app

# Copy Go module files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o jokes-quotes .

# -------- Runtime Stage --------
FROM alpine:latest

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/jokes-quotes .

# Expose the application's port
EXPOSE 8081

# Run the application
CMD ["./jokes-quotes"]
