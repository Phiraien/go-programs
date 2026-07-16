# ---------- Build Stage ----------
FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o password-generator .

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/password-generator .

EXPOSE 8080

CMD ["./password-generator"]
