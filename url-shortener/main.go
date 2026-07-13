package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// generateCode returns a random 6-character base62 short code.
func generateCode() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand should never fail; fall back to a fixed pattern.
		return "abcdef"
	}
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}

// Store maps a short code to its original URL.
var (
	mu    sync.RWMutex
	store = make(map[string]string)
)

// shortenRequest is the JSON body for POST /shorten.
type shortenRequest struct {
	URL string `json:"url"`
}

// shortenResponse is returned after a successful shorten.
type shortenResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

// homePage is the HTML form served at "/".
const homePage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>URL Shortener</title>
  <style>
    body { font-family: system-ui, sans-serif; background:#0f1115; color:#e6e6e6;
           display:flex; min-height:100vh; align-items:center; justify-content:center; margin:0; }
    .card { background:#1a1d24; padding:32px 28px; border-radius:14px; width:min(440px,92vw);
            box-shadow:0 10px 40px rgba(0,0,0,.4); }
    h1 { margin:0 0 4px; font-size:22px; }
    p.sub { margin:0 0 20px; color:#9aa0aa; font-size:13px; }
    input[type=url] { width:100%; padding:12px 14px; border-radius:9px; border:1px solid #2c313c;
            background:#0f1115; color:#fff; font-size:15px; box-sizing:border-box; }
    button { margin-top:14px; width:100%; padding:12px; border:0; border-radius:9px;
            background:#4f7cff; color:#fff; font-size:15px; font-weight:600; cursor:pointer; }
    button:hover { background:#3f6af0; }
    .result { margin-top:18px; padding:14px; border-radius:9px; background:#0f1115;
            border:1px solid #2c313c; font-size:14px; display:none; word-break:break-all; }
    .result a { color:#4f9cff; }
    .err { color:#ff6b6b; }
  </style>
</head>
<body>
  <div class="card">
    <h1>🔗 URL Shortener</h1>
    <p class="sub">Paste a long URL, get a short link.</p>
    <input type="url" id="url" placeholder="https://example.com/very/long/path" autofocus>
    <button onclick="shorten()">Shorten</button>
    <div class="result" id="result"></div>
  </div>
  <script>
    async function shorten() {
      const box = document.getElementById('url');
      const out = document.getElementById('result');
      const url = box.value.trim();
      out.style.display = 'block';
      out.className = 'result';
      if (!url) { out.innerHTML = '<span class="err">Please enter a URL.</span>'; return; }
      try {
        const res = await fetch('/shorten', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ url })
        });
        const data = await res.json();
        if (!res.ok) { out.innerHTML = '<span class="err">' + (data.error || 'Error') + '</span>'; return; }
        out.innerHTML = 'Short link: <a href="' + data.short_url + '" target="_blank">' + data.short_url + '</a>';
      } catch (e) {
        out.innerHTML = '<span class="err">Request failed.</span>';
      }
    }
    box?.addEventListener('keydown', e => { if (e.key === 'Enter') shorten(); });
  </script>
</body>
</html>`

// rootHandler serves the HTML form at "/" and redirects for any other path.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, homePage)
		return
	}
	// Any other path is treated as a short code → redirect.
	code := r.URL.Path[1:]
	mu.RLock()
	original, ok := store[code]
	mu.RUnlock()

	if !ok {
		http.Error(w, "Short code not found", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, original, http.StatusFound)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "Invalid JSON body, expected {\"url\": \"...\"}", http.StatusBadRequest)
		return
	}

	code := generateCode()
	mu.Lock()
	store[code] = req.URL
	mu.Unlock()

	resp := shortenResponse{
		ShortCode: code,
		ShortURL:  "http://localhost:8080/" + code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", rootHandler)

	println("URL shortener running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
