package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
)

// generateResponse is returned to the client.
type generateResponse struct {
	Length     int      `json:"length"`
	Count      int      `json:"count"`
	Passwords  []string `json:"passwords"`
	Error      string   `json:"error,omitempty"`
}

// charset pools.
const (
	lower   = "abcdefghijklmnopqrstuvwxyz"
	upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits  = "0123456789"
	symbols = "!@#$%^&*()-_=+[]{};:,.<>?"
)

// secureRandInt returns a crypto-secure uniform random int in [0, n).
func secureRandInt(n int) int {
	if n <= 0 {
		return 0
	}
	// crypto/rand.Int returns a uniform value in [0, n) with no modulo bias.
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(v.Int64())
}

// generateOne builds a single password of the given length from the pool.
func generateOne(length int, useUpper, useDigits, useSymbols bool) string {
	pool := lower
	if useUpper {
		pool += upper
	}
	if useDigits {
		pool += digits
	}
	if useSymbols {
		pool += symbols
	}

	b := make([]byte, length)
	poolLen := len(pool)
	for i := range b {
		b[i] = pool[secureRandInt(poolLen)]
	}
	return string(b)
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	length := 16
	if l := q.Get("length"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil || v < 1 || v > 128 {
			http.Error(w, "Invalid 'length' (1-128)", http.StatusBadRequest)
			return
		}
		length = v
	}

	count := 1
	if c := q.Get("count"); c != "" {
		v, err := strconv.Atoi(c)
		if err != nil || v < 1 || v > 50 {
			http.Error(w, "Invalid 'count' (1-50)", http.StatusBadRequest)
			return
		}
		count = v
	}

	useUpper := q.Get("upper") != "false"   // default true
	useDigits := q.Get("digits") != "false" // default true
	useSymbols := q.Get("symbols") == "true" // default false

	passwords := make([]string, count)
	for i := 0; i < count; i++ {
		passwords[i] = generateOne(length, useUpper, useDigits, useSymbols)
	}

	resp := generateResponse{
		Length:    length,
		Count:     count,
		Passwords: passwords,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// homePage is the HTML form served at "/".
const homePage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Password Generator</title>
  <style>
    body { font-family: system-ui, sans-serif; background:#0f1115; color:#e6e6e6;
           display:flex; min-height:100vh; align-items:center; justify-content:center; margin:0; }
    .card { background:#1a1d24; padding:32px 28px; border-radius:14px; width:min(480px,92vw);
            box-shadow:0 10px 40px rgba(0,0,0,.4); }
    h1 { margin:0 0 4px; font-size:22px; }
    p.sub { margin:0 0 20px; color:#9aa0aa; font-size:13px; }
    label { display:block; margin:14px 0 6px; font-size:14px; color:#c7ccd6; }
    input[type=range] { width:100%; }
    .val { color:#4f9cff; font-weight:700; }
    .checks { display:flex; flex-wrap:wrap; gap:14px; margin-top:8px; }
    .checks label { display:flex; align-items:center; gap:6px; margin:0; font-size:14px; }
    button { margin-top:20px; width:100%; padding:12px; border:0; border-radius:9px;
            background:#4f7cff; color:#fff; font-size:15px; font-weight:600; cursor:pointer; }
    button:hover { background:#3f6af0; }
    .out { margin-top:18px; display:none; }
    .pw { background:#0f1115; border:1px solid #2c313c; border-radius:9px; padding:12px 14px;
          margin-top:8px; font-family:monospace; font-size:16px; display:flex; justify-content:space-between; align-items:center; gap:10px; }
    .pw span { word-break:break-all; }
    .copy { background:#2c313c; border:0; color:#9aa0aa; border-radius:6px; padding:6px 10px; cursor:pointer; font-size:12px; }
    .copy:hover { color:#fff; }
  </style>
</head>
<body>
  <div class="card">
    <h1>🔐 Password Generator</h1>
    <p class="sub">Cryptographically secure (crypto/rand).</p>
    <label>Length: <span class="val" id="lenVal">16</span></label>
    <input type="range" id="length" min="4" max="64" value="16" oninput="document.getElementById('lenVal').textContent=this.value">
    <label>How many: <span class="val" id="cntVal">1</span></label>
    <input type="range" id="count" min="1" max="10" value="1" oninput="document.getElementById('cntVal').textContent=this.value">
    <div class="checks">
      <label><input type="checkbox" id="upper" checked> Uppercase</label>
      <label><input type="checkbox" id="digits" checked> Numbers</label>
      <label><input type="checkbox" id="symbols"> Symbols</label>
    </div>
    <button onclick="gen()">Generate</button>
    <div class="out" id="out"></div>
  </div>
  <script>
    async function gen() {
      const out = document.getElementById('out');
      out.style.display = 'block';
      const length = document.getElementById('length').value;
      const count = document.getElementById('count').value;
      const upper = document.getElementById('upper').checked;
      const digits = document.getElementById('digits').checked;
      const symbols = document.getElementById('symbols').checked;
      const qs = 'length=' + length + '&count=' + count + '&upper=' + upper + '&digits=' + digits + '&symbols=' + symbols;
      out.innerHTML = 'Generating...';
      try {
        const res = await fetch('/generate?' + qs);
        const data = await res.json();
        if (!res.ok || data.error) { out.innerHTML = '<span style="color:#ff6b6b">' + (data.error || 'Error') + '</span>'; return; }
        let html = '';
        data.passwords.forEach(p => {
          html += '<div class="pw"><span>' + p + '</span><button class="copy" onclick="copyPw(this)">Copy</button></div>';
        });
        out.innerHTML = html;
      } catch (e) {
        out.innerHTML = '<span style="color:#ff6b6b">Request failed.</span>';
      }
    }
    function copyPw(btn) {
      const pw = btn.parentElement.querySelector('span').textContent;
      navigator.clipboard.writeText(pw);
      btn.textContent = 'Copied';
      setTimeout(() => btn.textContent = 'Copy', 1200);
    }
  </script>
</body>
</html>`

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, homePage)
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/generate", generateHandler)
	println("Password generator running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
