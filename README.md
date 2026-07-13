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

### 8. `weather`
Live weather by city using open-meteo (no API key). Geocodes the city name, then fetches current conditions. Runs on **`:8082`**.

```bash
cd weather
go run main.go            # serves on :8082
# Open http://localhost:8082  (web form)
```

Query the API:
```
http://localhost:8082/weather?city=London
# → {"city":"London","country":"United Kingdom","temp":19.2,"temp_unit":"°C",
#     "wind":19.8,"wind_unit":"km/h","humidity":66,"condition":"Partly cloudy","source":"open-meteo.com"}
```

- `city` — any city name (required; missing → HTTP 400)
- Unknown city → JSON `error` field
- WMO weather codes are mapped to readable conditions (e.g. Rain, Thunderstorm)

### 9. `bmi-calculator`
Body Mass Index calculator with WHO categories. Metric (cm/kg) and Imperial (in/lb) support. Runs on **`:8083`**.

```bash
cd bmi-calculator
go run main.go            # serves on :8083
# Open http://localhost:8083  (web form)
```

Query the API:
```
http://localhost:8083/bmi?unit=metric&height=170&weight=65
# → {"bmi":22.49,"category":"Normal","unit":"metric"}
```

- `height`, `weight` — required (positive numbers); missing/invalid → HTTP 400
- `unit` — `metric` (default) or `imperial`; invalid → HTTP 400
- Categories: Underweight (<18.5), Normal (<25), Overweight (<30), Obese Class I/II/III

### 10. `site-checker`
Checks whether a list of websites are up or down, in parallel (goroutines + WaitGroup). Reports status code and latency per URL. Runs on **`:8080`**.

```bash
cd site-checker
go run main.go            # serves on :8080
# Open http://localhost:8080  (web form)
```

Check via API — POST JSON:
```bash
curl -X POST http://localhost:8080/check -H "Content-Type: application/json" \
  -d '{"urls":["https://github.com","https://google.com"]}'
# → {"results":[{"url":"https://github.com","status_code":200,"up":true,"latency_ms":83.1}, ...]}
```

Or via query string (comma/newline separated):
```
http://localhost:8080/check?urls=https://github.com,https://google.com
```

- Each URL is checked concurrently; total time ≈ the slowest one
- Redirects are NOT followed (real status like 301/302 is reported)
- Per result: `up`, `status_code`, `latency_ms`, `error` (on failure)
- Max 50 URLs per request; empty list → HTTP 400

## Requirements
```bash
go build -o app.exe main.go   # produces a standalone binary (excluded from git via .gitignore)
```
