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

## Building
```bash
go build -o app.exe main.go   # produces a standalone binary (excluded from git via .gitignore)
```
