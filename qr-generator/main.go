package main

import (
	"fmt"
	"net/http"

	"github.com/skip2/go-qrcode"
)

// homePage is the HTML form served at "/".
const homePage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>QR Code Generator</title>
  <style>
    body { font-family: system-ui, sans-serif; background:#0f1115; color:#e6e6e6;
           display:flex; min-height:100vh; align-items:center; justify-content:center; margin:0; }
    .card { background:#1a1d24; padding:32px 28px; border-radius:14px; width:min(460px,92vw);
            box-shadow:0 10px 40px rgba(0,0,0,.4); }
    h1 { margin:0 0 4px; font-size:22px; }
    p.sub { margin:0 0 18px; color:#9aa0aa; font-size:13px; }
    textarea { width:100%; padding:12px 14px; border-radius:9px; border:1px solid #2c313c;
            background:#0f1115; color:#fff; font-size:15px; box-sizing:border-box; min-height:80px; resize:vertical; }
    button { margin-top:14px; width:100%; padding:12px; border:0; border-radius:9px;
            background:#4f7cff; color:#fff; font-size:15px; font-weight:600; cursor:pointer; }
    button:hover { background:#3f6af0; }
    .out { margin-top:18px; text-align:center; display:none; }
    .out img { background:#fff; padding:12px; border-radius:10px; max-width:100%; }
    .err { color:#ff6b6b; }
  </style>
</head>
<body>
  <div class="card">
    <h1>🔳 QR Code Generator</h1>
    <p class="sub">Enter text or a URL to encode.</p>
    <textarea id="text" placeholder="https://github.com/Phiraien">https://github.com/Phiraien</textarea>
    <button onclick="gen()">Generate QR</button>
    <div class="out" id="out"></div>
  </div>
  <script>
    async function gen() {
      const out = document.getElementById('out');
      out.style.display = 'block';
      const text = document.getElementById('text').value.trim();
      if (!text) { out.innerHTML = '<span class="err">Enter some text.</span>'; return; }
      out.innerHTML = 'Generating...';
      try {
        const res = await fetch('/qr?text=' + encodeURIComponent(text));
        if (!res.ok) { out.innerHTML = '<span class="err">Failed to generate.</span>'; return; }
        const blob = await res.blob();
        const url = URL.createObjectURL(blob);
        out.innerHTML = '<img src="' + url + '" alt="QR code">';
      } catch (e) {
        out.innerHTML = '<span class="err">Request failed.</span>';
      }
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

func qrHandler(w http.ResponseWriter, r *http.Request) {
	text := r.URL.Query().Get("text")
	if text == "" {
		http.Error(w, "Missing 'text' query param", http.StatusBadRequest)
		return
	}
	// Reject absurdly long input.
	if len(text) > 2000 {
		http.Error(w, "'text' too long (max 2000)", http.StatusBadRequest)
		return
	}

	png, err := qrcode.Encode(text, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Failed to encode QR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "inline; filename=\"qrcode.png\"")
	w.Write(png)
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/qr", qrHandler)
	println("QR code generator running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
