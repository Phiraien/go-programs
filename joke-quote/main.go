package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// JokeResponse from jokeapi.dev.
type jokeAPI struct {
	Error    bool   `json:"error"`
	Joke     string `json:"joke"`     // single-part joke
	Setup    string `json:"setup"`    // two-part
	Delivery string `json:"delivery"` // two-part
}

// QuoteResponse from dummyjson.com (reliable, no key).
type quoteAPI struct {
	Quote  string `json:"quote"`
	Author string `json:"author"`
}

// item is what we return to the client.
type item struct {
	Type    string `json:"type"` // "joke" or "quote"
	Text    string `json:"text"`
	Author  string `json:"author,omitempty"`
	Source  string `json:"source"`
	Error   string `json:"error,omitempty"`
}

func fetchJoke() (item, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://v2.jokeapi.dev/joke/Any?type=single")
	if err != nil {
		return item{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var j jokeAPI
	if err := json.Unmarshal(body, &j); err != nil || j.Error {
		return item{}, fmt.Errorf("joke api error")
	}
	text := j.Joke
	if text == "" {
		text = j.Setup + " " + j.Delivery
	}
	return item{Type: "joke", Text: text, Source: "jokeapi.dev"}, nil
}

func fetchQuote() (item, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://dummyjson.com/quotes/random")
	if err != nil {
		return item{}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var q quoteAPI
	if err := json.Unmarshal(body, &q); err != nil {
		return item{}, fmt.Errorf("quote api error")
	}
	return item{Type: "quote", Text: q.Quote, Author: q.Author, Source: "dummyjson.com"}, nil
}

func randomHandler(w http.ResponseWriter, r *http.Request) {
	// Try quote first, fall back to joke.
	if it, err := fetchQuote(); err == nil {
		writeJSON(w, it)
		return
	}
	if it, err := fetchJoke(); err == nil {
		writeJSON(w, it)
		return
	}
	writeJSON(w, item{Error: "Both joke and quote APIs are unreachable"})
}

func jokeHandler(w http.ResponseWriter, r *http.Request) {
	it, err := fetchJoke()
	if err != nil {
		writeJSON(w, item{Error: "Joke API unreachable"})
		return
	}
	writeJSON(w, it)
}

func quoteHandler(w http.ResponseWriter, r *http.Request) {
	it, err := fetchQuote()
	if err != nil {
		writeJSON(w, item{Error: "Quote API unreachable"})
		return
	}
	writeJSON(w, it)
}

func writeJSON(w http.ResponseWriter, it item) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(it)
}

// homePage is the HTML page served at "/".
const homePage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Jokes & Quotes</title>
  <style>
    body { font-family: system-ui, sans-serif; background:#0f1115; color:#e6e6e6;
           display:flex; min-height:100vh; align-items:center; justify-content:center; margin:0; }
    .card { background:#1a1d24; padding:32px 28px; border-radius:14px; width:min(480px,92vw);
            box-shadow:0 10px 40px rgba(0,0,0,.4); text-align:center; }
    h1 { margin:0 0 4px; font-size:22px; }
    p.sub { margin:0 0 18px; color:#9aa0aa; font-size:13px; }
    .btns { display:flex; gap:10px; }
    button { flex:1; padding:12px; border:0; border-radius:9px; background:#4f7cff; color:#fff;
            font-size:15px; font-weight:600; cursor:pointer; }
    button:hover { background:#3f6af0; }
    .out { margin-top:20px; padding:20px; border-radius:10px; background:#0f1115; border:1px solid #2c313c;
           font-size:18px; line-height:1.5; display:none; min-height:60px; }
    .out .author { display:block; margin-top:12px; color:#9aa0aa; font-size:14px; font-style:italic; }
    .tag { display:inline-block; margin-bottom:10px; font-size:11px; letter-spacing:1px; text-transform:uppercase;
           color:#4f9cff; border:1px solid #2c313c; padding:3px 8px; border-radius:20px; }
    .err { color:#ff6b6b; }
  </style>
</head>
<body>
  <div class="card">
    <h1>😂 Jokes & 💬 Quotes</h1>
    <p class="sub">Random from free public APIs.</p>
    <div class="btns">
      <button onclick="get('/random')">Random</button>
      <button onclick="get('/joke')">Joke</button>
      <button onclick="get('/quote')">Quote</button>
    </div>
    <div class="out" id="out"></div>
  </div>
  <script>
    async function get(path) {
      const out = document.getElementById('out');
      out.style.display = 'block';
      out.innerHTML = 'Loading...';
      try {
        const res = await fetch(path);
        const data = await res.json();
        if (data.error) { out.innerHTML = '<span class="err">' + data.error + '</span>'; return; }
        out.innerHTML = '<span class="tag">' + data.type + '</span>' + data.text +
          (data.author ? '<span class="author">— ' + data.author + '</span>' : '');
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
	http.HandleFunc("/random", randomHandler)
	http.HandleFunc("/joke", jokeHandler)
	http.HandleFunc("/quote", quoteHandler)
	println("Jokes & Quotes running on http://localhost:8081")
	http.ListenAndServe(":8081", nil)
}
