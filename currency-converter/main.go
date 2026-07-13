package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// apiResponse matches the open.er-api.com/v6/latest/<BASE> shape.
type apiResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

// convertResponse is what we return to the client.
type convertResponse struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
	Rate   float64 `json:"rate"`
	Result float64 `json:"result"`
	Source string  `json:"source"` // "live" or "fallback"
	Error  string  `json:"error,omitempty"`
}

// fallbackRates: rough static rates relative to USD, used only if the
// live API is unreachable. Values are approximate, not real-time.
var fallbackRates = map[string]float64{
	"USD": 1.0, "EUR": 0.92, "GBP": 0.79, "INR": 83.3, "JPY": 157.0,
	"CAD": 1.37, "AUD": 1.52, "CHF": 0.90, "CNY": 7.25, "SGD": 1.35,
	"AED": 3.67, "BTC": 0.0000148,
}

const ratesURL = "https://open.er-api.com/v6/latest/USD"

// fetchRates returns live rates keyed by currency code, relative to USD.
func fetchRates() (map[string]float64, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(ratesURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rates API returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var ar apiResponse
	if err := json.Unmarshal(body, &ar); err != nil {
		return nil, err
	}
	if ar.Rates == nil || len(ar.Rates) == 0 {
		return nil, fmt.Errorf("empty rates")
	}
	return ar.Rates, nil
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")
	amountStr := q.Get("amount")

	if from == "" || to == "" || amountStr == "" {
		http.Error(w, "Missing 'from', 'to', or 'amount' query param", http.StatusBadRequest)
		return
	}
	from = normalize(from)
	to = normalize(to)

	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		http.Error(w, "Invalid 'amount'", http.StatusBadRequest)
		return
	}

	// Try live rates first, fall back to static if offline.
	rates, fetchErr := fetchRates()
	src := "live"
	if fetchErr != nil {
		rates = fallbackRates
		src = "fallback"
	}

	fromRate, ok1 := rates[from]
	toRate, ok2 := rates[to]
	if !ok1 || !ok2 {
		http.Error(w, "Unsupported currency code", http.StatusBadRequest)
		return
	}

	// Convert via USD: amount -> USD -> target.
	usd := amount / fromRate
	result := usd * toRate
	rate := toRate / fromRate

	out := convertResponse{
		From:   from,
		To:     to,
		Amount: amount,
		Rate:   rate,
		Result: result,
		Source: src,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func normalize(c string) string {
	// Currencies are uppercase codes; normalize user input.
	b := []byte(c)
	for i := range b {
		if b[i] >= 'a' && b[i] <= 'z' {
			b[i] -= 'a' - 'A'
		}
	}
	return string(b)
}

// homePage is the HTML form served at "/".
const homePage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Currency Converter</title>
  <style>
    body { font-family: system-ui, sans-serif; background:#0f1115; color:#e6e6e6;
           display:flex; min-height:100vh; align-items:center; justify-content:center; margin:0; }
    .card { background:#1a1d24; padding:32px 28px; border-radius:14px; width:min(460px,92vw);
            box-shadow:0 10px 40px rgba(0,0,0,.4); }
    h1 { margin:0 0 4px; font-size:22px; }
    p.sub { margin:0 0 20px; color:#9aa0aa; font-size:13px; }
    .row { display:flex; gap:10px; }
    input, select { padding:12px 14px; border-radius:9px; border:1px solid #2c313c;
            background:#0f1115; color:#fff; font-size:15px; box-sizing:border-box; }
    input[type=number] { flex:1; }
    select { flex:1; }
    button { margin-top:16px; width:100%; padding:12px; border:0; border-radius:9px;
            background:#4f7cff; color:#fff; font-size:15px; font-weight:600; cursor:pointer; }
    button:hover { background:#3f6af0; }
    .result { margin-top:18px; padding:16px; border-radius:9px; background:#0f1115;
            border:1px solid #2c313c; font-size:18px; text-align:center; display:none; }
    .result .big { font-size:24px; font-weight:700; color:#4f9cff; }
    .result .meta { font-size:12px; color:#9aa0aa; margin-top:6px; }
    .err { color:#ff6b6b; }
  </style>
</head>
<body>
  <div class="card">
    <h1>💱 Currency Converter</h1>
    <p class="sub">Live exchange rates (falls back to static if offline).</p>
    <div class="row">
      <input type="number" id="amount" placeholder="100" value="100" step="any">
      <select id="from"></select>
    </div>
    <div class="row" style="margin-top:10px;">
      <input type="text" id="to" placeholder="To (e.g. INR)" value="EUR">
    </div>
    <button onclick="convert()">Convert</button>
    <div class="result" id="result"></div>
  </div>
  <script>
    const currencies = ["USD","EUR","GBP","INR","JPY","CAD","AUD","CHF","CNY","SGD","AED","BTC"];
    const fromSel = document.getElementById('from');
    currencies.forEach(c => {
      const o = document.createElement('option');
      o.value = c; o.textContent = c;
      if (c === 'USD') o.selected = true;
      fromSel.appendChild(o);
    });

    async function convert() {
      const out = document.getElementById('result');
      out.style.display = 'block';
      const amount = document.getElementById('amount').value.trim();
      const from = document.getElementById('from').value;
      const to = document.getElementById('to').value.trim().toUpperCase();
      if (!amount || !to) { out.innerHTML = '<span class="err">Enter amount and target currency.</span>'; return; }
      out.innerHTML = 'Converting...';
      try {
        const res = await fetch('/convert?from=' + from + '&to=' + to + '&amount=' + amount);
        const data = await res.json();
        if (!res.ok || data.error) { out.innerHTML = '<span class="err">' + (data.error || 'Error') + '</span>'; return; }
        out.innerHTML = '<div class="big">' + data.amount + ' ' + data.from + ' = ' + data.result + ' ' + data.to + '</div>'
          + '<div class="meta">Rate: 1 ' + data.from + ' = ' + data.rate + ' ' + data.to + ' &middot; ' + data.source + ' rates</div>';
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
	http.HandleFunc("/convert", convertHandler)
	println("Currency converter running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
