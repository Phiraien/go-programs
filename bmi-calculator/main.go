package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// bmiResponse is what we return to the client.
type bmiResponse struct {
	BMI       float64 `json:"bmi"`
	Category  string  `json:"category"`
	Unit      string  `json:"unit"` // "metric" or "imperial"
	Error     string  `json:"error,omitempty"`
}

// category maps a BMI value to the WHO classification.
func category(bmi float64) string {
	switch {
	case bmi < 18.5:
		return "Underweight"
	case bmi < 25:
		return "Normal"
	case bmi < 30:
		return "Overweight"
	case bmi < 35:
		return "Obese (Class I)"
	case bmi < 40:
		return "Obese (Class II)"
	default:
		return "Obese (Class III)"
	}
}

func bmiHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	unit := q.Get("unit")
	if unit == "" {
		unit = "metric"
	}

	heightStr := q.Get("height")
	weightStr := q.Get("weight")
	if heightStr == "" || weightStr == "" {
		http.Error(w, "Missing 'height' or 'weight'", http.StatusBadRequest)
		return
	}

	height, err := strconv.ParseFloat(heightStr, 64)
	if err != nil || height <= 0 {
		http.Error(w, "Invalid 'height'", http.StatusBadRequest)
		return
	}
	weight, err := strconv.ParseFloat(weightStr, 64)
	if err != nil || weight <= 0 {
		http.Error(w, "Invalid 'weight'", http.StatusBadRequest)
		return
	}

	var bmi float64
	switch unit {
	case "metric":
		// height in cm -> m
		hM := height / 100
		bmi = weight / (hM * hM)
	case "imperial":
		// height in inches, weight in pounds
		bmi = (weight / (height * height)) * 703
	default:
		http.Error(w, "Invalid 'unit' (metric|imperial)", http.StatusBadRequest)
		return
	}

	resp := bmiResponse{
		BMI:      bmi,
		Category: category(bmi),
		Unit:     unit,
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
  <title>BMI Calculator</title>
  <style>
    body { font-family: system-ui, sans-serif; background:#0f1115; color:#e6e6e6;
           display:flex; min-height:100vh; align-items:center; justify-content:center; margin:0; }
    .card { background:#1a1d24; padding:32px 28px; border-radius:14px; width:min(420px,92vw);
            box-shadow:0 10px 40px rgba(0,0,0,.4); text-align:center; }
    h1 { margin:0 0 4px; font-size:22px; }
    p.sub { margin:0 0 18px; color:#9aa0aa; font-size:13px; }
    .unit { display:flex; gap:8px; justify-content:center; margin-bottom:16px; }
    .unit button { flex:1; padding:9px; border:1px solid #2c313c; background:#0f1115; color:#9aa0aa;
            border-radius:8px; cursor:pointer; font-size:14px; }
    .unit button.active { background:#4f7cff; color:#fff; border-color:#4f7cff; }
    label { display:block; text-align:left; font-size:14px; color:#c7ccd6; margin:12px 0 6px; }
    input { width:100%; padding:12px 14px; border-radius:9px; border:1px solid #2c313c;
            background:#0f1115; color:#fff; font-size:15px; box-sizing:border-box; }
    button.calc { margin-top:18px; width:100%; padding:12px; border:0; border-radius:9px;
            background:#4f7cff; color:#fff; font-size:15px; font-weight:600; cursor:pointer; }
    button.calc:hover { background:#3f6af0; }
    .out { margin-top:20px; display:none; }
    .bmi { font-size:44px; font-weight:700; color:#4f9cff; }
    .cat { font-size:18px; margin-top:4px; }
    .hint { font-size:12px; color:#9aa0aa; margin-top:8px; }
    .err { color:#ff6b6b; }
  </style>
</head>
<body>
  <div class="card">
    <h1>⚖️ BMI Calculator</h1>
    <p class="sub">Body Mass Index (WHO categories).</p>
    <div class="unit">
      <button id="metricBtn" class="active" onclick="setUnit('metric')">Metric (cm/kg)</button>
      <button id="impBtn" onclick="setUnit('imperial')">Imperial (in/lb)</button>
    </div>
    <label id="hLabel">Height (cm)</label>
    <input type="number" id="height" placeholder="170" step="any">
    <label id="wLabel">Weight (kg)</label>
    <input type="number" id="weight" placeholder="65" step="any">
    <button class="calc" onclick="calc()">Calculate</button>
    <div class="out" id="out"></div>
  </div>
  <script>
    let unit = 'metric';
    function setUnit(u) {
      unit = u;
      document.getElementById('metricBtn').classList.toggle('active', u === 'metric');
      document.getElementById('impBtn').classList.toggle('active', u === 'imperial');
      document.getElementById('hLabel').textContent = u === 'metric' ? 'Height (cm)' : 'Height (in)';
      document.getElementById('wLabel').textContent = u === 'metric' ? 'Weight (kg)' : 'Weight (lb)';
    }
    async function calc() {
      const out = document.getElementById('out');
      out.style.display = 'block';
      const height = document.getElementById('height').value.trim();
      const weight = document.getElementById('weight').value.trim();
      if (!height || !weight) { out.innerHTML = '<span class="err">Enter height and weight.</span>'; return; }
      out.innerHTML = 'Calculating...';
      try {
        const res = await fetch('/bmi?unit=' + unit + '&height=' + height + '&weight=' + weight);
        const data = await res.json();
        if (!res.ok || data.error) { out.innerHTML = '<span class="err">' + (data.error || 'Error') + '</span>'; return; }
        out.innerHTML = '<div class="bmi">' + data.bmi + '</div><div class="cat">' + data.category + '</div>' +
          '<div class="hint">' + (unit === 'metric' ? 'kg/m²' : 'lb/in² × 703') + '</div>';
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
	http.HandleFunc("/bmi", bmiHandler)
	println("BMI calculator running on http://localhost:8083")
	http.ListenAndServe(":8083", nil)
}
