package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// checkResult is the outcome for a single URL.
type checkResult struct {
	URL        string  `json:"url"`
	StatusCode int     `json:"status_code"`
	OK         bool    `json:"ok"`
	Up         bool    `json:"up"`
	LatencyMS  float64 `json:"latency_ms"`
	Error      string  `json:"error,omitempty"`
}

// checkResponse is what we return for a batch.
type checkResponse struct {
	Results []checkResult `json:"results"`
	Error   string        `json:"error,omitempty"`
}

// checkOne probes a single URL and fills the result.
func checkOne(url string, out *checkResult) {
	out.URL = url
	client := &http.Client{
		Timeout: 8 * time.Second,
		// Don't follow redirects automatically so we report the real status.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	start := time.Now()
	resp, err := client.Get(url)
	latency := time.Since(start)
	out.LatencyMS = float64(latency.Microseconds()) / 1000.0

	if err != nil {
		out.Error = err.Error()
		out.Up = false
		out.OK = false
		return
	}
	defer resp.Body.Close()

	out.StatusCode = resp.StatusCode
	out.OK = resp.StatusCode >= 200 && resp.StatusCode < 400
	out.Up = true
}

// checkHandler accepts a JSON body {"urls":["...","..."]} OR a comma/
// newline separated list via ?urls=. Returns results in parallel.
func checkHandler(w http.ResponseWriter, r *http.Request) {
	var urls []string

	if r.Method == http.MethodPost {
		var body struct {
			URLs []string `json:"urls"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || len(body.URLs) == 0 {
			http.Error(w, "Invalid JSON body, expected {\"urls\":[\"...\"]}", http.StatusBadRequest)
			return
		}
		urls = body.URLs
	} else {
		raw := r.URL.Query().Get("urls")
		if raw == "" {
			http.Error(w, "Missing 'urls' (POST JSON or ?urls=a,b,c)", http.StatusBadRequest)
			return
		}
		// split on comma or newline
		for _, part := range splitList(raw) {
			if part != "" {
				urls = append(urls, part)
			}
		}
	}

	if len(urls) > 50 {
		http.Error(w, "Too many URLs (max 50)", http.StatusBadRequest)
		return
	}

	results := make([]checkResult, len(urls))
	var wg sync.WaitGroup
	for i, u := range urls {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()
			checkOne(url, &results[idx])
		}(i, u)
	}
	wg.Wait()

	resp := checkResponse{Results: results}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// splitList splits on commas and newlines, trimming whitespace.
func splitList(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ',' || r == '\n' || r == '\r' || r == ' ' || r == '\t' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
		} else {
			cur += string(r)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// homePage is the HTML page served at "/".
const homePage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Site Up/Down Checker</title>
  <style>
    body { font-family: system-ui, sans-serif; background:#0f1115; color:#e6e6e6;
           display:flex; min-height:100vh; align-items:center; justify-content:center; margin:0; }
    .card { background:#1a1d24; padding:32px 28px; border-radius:14px; width:min(560px,94vw);
            box-shadow:0 10px 40px rgba(0,0,0,.4); }
    h1 { margin:0 0 4px; font-size:22px; }
    p.sub { margin:0 0 18px; color:#9aa0aa; font-size:13px; }
    textarea { width:100%; padding:12px 14px; border-radius:9px; border:1px solid #2c313c;
            background:#0f1115; color:#fff; font-size:14px; box-sizing:border-box; min-height:120px; resize:vertical; }
    button { margin-top:14px; width:100%; padding:12px; border:0; border-radius:9px;
            background:#4f7cff; color:#fff; font-size:15px; font-weight:600; cursor:pointer; }
    button:hover { background:#3f6af0; }
    .out { margin-top:18px; display:none; }
    .row { display:flex; justify-content:space-between; align-items:center; gap:10px;
           padding:10px 12px; border-radius:8px; background:#0f1115; border:1px solid #2c313c; margin-top:8px; }
    .row .u { word-break:break-all; font-size:13px; }
    .badge { font-size:12px; font-weight:700; padding:3px 9px; border-radius:20px; white-space:nowrap; }
    .up { background:#16351f; color:#4ade80; }
    .down { background:#3a1d1d; color:#f87171; }
    .meta { color:#9aa0aa; font-size:12px; }
    .err { color:#ff6b6b; }
  </style>
</head>
<body>
  <div class="card">
    <h1>📡 Site Up/Down Checker</h1>
    <p class="sub">Checks many URLs in parallel. One per line or comma-separated.</p>
    <textarea id="urls" placeholder="https://github.com&#10;https://google.com&#10;https://this-domain-does-not-exist-xyz.com">https://github.com
https://google.com
https://this-domain-does-not-exist-xyz.com</textarea>
    <button onclick="check()">Check Sites</button>
    <div class="out" id="out"></div>
  </div>
  <script>
    async function check() {
      const out = document.getElementById('out');
      out.style.display = 'block';
      const raw = document.getElementById('urls').value.trim();
      if (!raw) { out.innerHTML = '<span class="err">Enter at least one URL.</span>'; return; }
      out.innerHTML = 'Checking...';
      try {
        const res = await fetch('/check', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ urls: raw.split(/\n|,/).map(s => s.trim()).filter(s => s) })
        });
        const data = await res.json();
        if (!res.ok || data.error) { out.innerHTML = '<span class="err">' + (data.error || 'Error') + '</span>'; return; }
        let html = '';
        data.results.forEach(r => {
          const cls = r.up ? 'up' : 'down';
          const label = r.up ? ('UP ' + r.status_code) : 'DOWN';
          const detail = r.up ? (r.latency_ms.toFixed(0) + ' ms') : r.error;
          html += '<div class="row"><span class="u">' + r.url + '</span>' +
                  '<span class="badge ' + cls + '">' + label + '</span></div>' +
                  '<div class="meta" style="margin:2px 0 8px 2px">' + detail + '</div>';
        });
        out.innerHTML = html;
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

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/check", checkHandler)
	println("Site checker running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
