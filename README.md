# Go Programs

A collection of small Go HTTP programs. Each subfolder is its own Go module.

## Programs

### 1. `hello-server`
A minimal HTTP server that responds with `Hello, World!` on every request.

```bash
cd hello-server
go run main.go
# Open http://localhost:8080
```

### 2. `calculator`
A JSON calculator service. Send `op`, `a`, and `b` as query parameters.

```bash
cd calculator
go run main.go
```

Example requests:
```
http://localhost:8080/calculate?op=add&a=10&b=5   → {"operation":"add","a":10,"b":5,"result":15}
http://localhost:8080/calculate?op=sub&a=10&b=5   → {"operation":"sub","a":10,"b":5,"result":5}
http://localhost:8080/calculate?op=mul&a=10&b=5   → {"operation":"mul","a":10,"b":5,"result":50}
http://localhost:8080/calculate?op=div&a=10&b=5   → {"operation":"div","a":10,"b":5,"result":2}
http://localhost:8080/calculate?op=div&a=10&b=0   → {"operation":"div","a":10,"b":0,"error":"Division by zero"}
```

Operations: `add`, `sub`, `mul`, `div`. Invalid numbers return HTTP 400; unknown operations return a JSON `error` field.

### 3. `url-shortener`
An in-memory URL shortener. POST a URL to get a short code, then visit the short URL to be redirected.

```bash
cd url-shortener
go run main.go
```

Shorten a URL:
```bash
curl -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/Phiraien/filesync-app"}'
# → {"short_code":"htViyq","short_url":"http://localhost:8080/htViyq"}
```

Follow the redirect:
```
GET http://localhost:8080/htViyq   → 302 → https://github.com/Phiraien/filesync-app
```

Behavior:
- `POST /shorten` with `{"url":"..."}` → returns a 6-char `short_code` (HTTP 201)
- `GET /<code>` → 302 redirect to the original URL
- missing/empty URL → HTTP 400; unknown code → HTTP 404; wrong method → HTTP 405
- **Note:** storage is in-memory only — codes reset when the server restarts.

### 4. `currency-converter`
A currency converter with live exchange rates (falls back to static rates if offline). Ships with a web form at `/`.

```bash
cd currency-converter
go run main.go
# Open http://localhost:8080  (web form)
```

Convert via the API:
```
http://localhost:8080/convert?from=usd&to=inr&amount=100
# → {"from":"USD","to":"INR","amount":100,"rate":95.49,"result":9549.28,"source":"live"}
```

Behavior:
- `GET /convert?from=<CUR>&to=<CUR>&amount=<N>` → JSON with `rate`, `result`, and `source`
- `source` is `"live"` (from open.er-api.com, no key needed) or `"fallback"` (static rates if offline)
- Currencies: USD, EUR, GBP, INR, JPY, CAD, AUD, CHF, CNY, SGD, AED, BTC (codes are case-insensitive)
- missing params or unknown currency → HTTP 400
- Go 1.26+

### 5. `password-generator`
A cryptographically secure password generator (uses `crypto/rand`, no modulo bias). Ships with a web form at `/`.

```bash
cd password-generator
go run main.go
# Open http://localhost:8080  (web form)
```

Generate via the API:
```
http://localhost:8080/generate?length=16&count=1&upper=true&digits=true&symbols=false
# → {"length":16,"count":1,"passwords":["7tVJPwxCvH7HfIG2"]}
```

Query params:
- `length` — password length, 1–128 (default 16)
- `count` — how many to generate, 1–50 (default 1)
- `upper` — include uppercase A–Z (default true)
- `digits` — include numbers 0–9 (default true)
- `symbols` — include symbols e.g. `!@#$%` (default false)

Behavior:
- at least lowercase `a–z` is always included
- invalid `length`/`count` → HTTP 400

### 6. `qr-generator`
Generates a scannable QR code PNG from any text or URL. Uses `github.com/skip2/go-qrcode`.

```bash
cd qr-generator
go run main.go            # serves on :8080
# Open http://localhost:8080  (web form)
```

Generate via the API:
```
http://localhost:8080/qr?text=https://github.com/Phiraien
# → returns a 256px PNG image (Content-Type: image/png)
```

- `text` — any string up to 2000 chars (required; missing → HTTP 400)
- The web form shows the QR inline; scan it with your phone.

### 7. `joke-quote`
Fetches random jokes and quotes from free public APIs. Serves a web page at `/` with Random / Joke / Quote buttons. Runs on **`:8081`** (so it can coexist with other services on 8080).

```bash
cd joke-quote
go run main.go            # serves on :8081
# Open http://localhost:8081  (web UI)
```

Endpoints (all return JSON `{type, text, author?, source}`):
```
http://localhost:8081/random   → a joke or quote
http://localhost:8081/joke     → joke (jokeapi.dev)
http://localhost:8081/quote    → quote (dummyjson.com)
```
- If the primary source is down, `/random` falls back to the other; individual endpoints report an `error` field.

## Requirements
```bash
go build -o app.exe main.go   # produces a standalone binary (excluded from git via .gitignore)
```
