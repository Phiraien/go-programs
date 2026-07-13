package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// geocodeResponse matches open-meteo geocoding API.
type geocodeResponse struct {
	Results []struct {
		Name    string  `json:"name"`
		Country string  `json:"country"`
		Lat     float64 `json:"latitude"`
		Lon     float64 `json:"longitude"`
	} `json:"results"`
}

// weatherResponse matches the relevant parts of open-meteo forecast API.
type weatherResponse struct {
	Current struct {
		Temp      float64 `json:"temperature_2m"`
		WindSpeed float64 `json:"wind_speed_10m"`
		Humidity  float64 `json:"relative_humidity_2m"`
		Code      int     `json:"weather_code"`
	} `json:"current"`
	CurrentUnits struct {
		Temp     string `json:"temperature_2m"`
		Wind     string `json:"wind_speed_10m"`
		Humidity string `json:"relative_humidity_2m"`
	} `json:"current_units"`
}

// weatherOut is what we return to the client.
type weatherOut struct {
	City      string `json:"city"`
	Country   string `json:"country"`
	Temp      float64 `json:"temp"`
	TempUnit  string `json:"temp_unit"`
	Wind      float64 `json:"wind"`
	WindUnit  string `json:"wind_unit"`
	Humidity  float64 `json:"humidity"`
	Condition string `json:"condition"`
	Source    string `json:"source"`
	Error     string `json:"error,omitempty"`
}

// weatherCodeText maps WMO weather codes to a short description.
func weatherCodeText(code int) string {
	switch code {
	case 0:
		return "Clear sky"
	case 1, 2, 3:
		return "Partly cloudy"
	case 45, 48:
		return "Fog"
	case 51, 53, 55:
		return "Drizzle"
	case 56, 57:
		return "Freezing drizzle"
	case 61, 63, 65:
		return "Rain"
	case 66, 67:
		return "Freezing rain"
	case 71, 73, 75:
		return "Snow"
	case 77:
		return "Snow grains"
	case 80, 81, 82:
		return "Showers"
	case 85, 86:
		return "Snow showers"
	case 95:
		return "Thunderstorm"
	case 96, 99:
		return "Thunderstorm with hail"
	default:
		return "Unknown"
	}
}

func getJSON(client *http.Client, urlStr string, out interface{}) error {
	resp, err := client.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

func weatherHandler(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "Missing 'city' query param", http.StatusBadRequest)
		return
	}

	client := &http.Client{Timeout: 8 * time.Second}

	// 1) Geocode city -> lat/lon.
	geoURL := "https://geocoding-api.open-meteo.com/v1/search?name=" +
		url.QueryEscape(city) + "&count=1"
	var geo geocodeResponse
	if err := getJSON(client, geoURL, &geo); err != nil || len(geo.Results) == 0 {
		writeWeather(w, weatherOut{Error: "City not found or geocoding failed"})
		return
	}
	res := geo.Results[0]

	// 2) Fetch current weather.
	wxURL := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,wind_speed_10m,weather_code",
		res.Lat, res.Lon,
	)
	var wx weatherResponse
	if err := getJSON(client, wxURL, &wx); err != nil {
		writeWeather(w, weatherOut{Error: "Weather fetch failed"})
		return
	}

	out := weatherOut{
		City:      res.Name,
		Country:   res.Country,
		Temp:      wx.Current.Temp,
		TempUnit:  wx.CurrentUnits.Temp,
		Wind:      wx.Current.WindSpeed,
		WindUnit:  wx.CurrentUnits.Wind,
		Humidity:  wx.Current.Humidity,
		Condition: weatherCodeText(wx.Current.Code),
		Source:    "open-meteo.com",
	}
	writeWeather(w, out)
}

func writeWeather(w http.ResponseWriter, out weatherOut) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

// homePage is the HTML page served at "/".
const homePage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Weather</title>
  <style>
    body { font-family: system-ui, sans-serif; background:#0f1115; color:#e6e6e6;
           display:flex; min-height:100vh; align-items:center; justify-content:center; margin:0; }
    .card { background:#1a1d24; padding:32px 28px; border-radius:14px; width:min(440px,92vw);
            box-shadow:0 10px 40px rgba(0,0,0,.4); text-align:center; }
    h1 { margin:0 0 4px; font-size:22px; }
    p.sub { margin:0 0 18px; color:#9aa0aa; font-size:13px; }
    .row { display:flex; gap:10px; }
    input { flex:1; padding:12px 14px; border-radius:9px; border:1px solid #2c313c;
            background:#0f1115; color:#fff; font-size:15px; box-sizing:border-box; }
    button { padding:12px 18px; border:0; border-radius:9px; background:#4f7cff; color:#fff;
            font-size:15px; font-weight:600; cursor:pointer; }
    button:hover { background:#3f6af0; }
    .out { margin-top:20px; display:none; }
    .temp { font-size:42px; font-weight:700; color:#4f9cff; }
    .cond { font-size:18px; margin-top:4px; }
    .meta { margin-top:12px; color:#9aa0aa; font-size:14px; line-height:1.6; }
    .err { color:#ff6b6b; }
    .loading { color:#9aa0aa; }
  </style>
</head>
<body>
  <div class="card">
    <h1>🌤️ Weather</h1>
    <p class="sub">Live data from open-meteo (no API key).</p>
    <div class="row">
      <input id="city" placeholder="City name (e.g. London)" value="London">
      <button onclick="getW()">Check</button>
    </div>
    <div class="out" id="out"></div>
  </div>
  <script>
    async function getW() {
      const out = document.getElementById('out');
      out.style.display = 'block';
      const city = document.getElementById('city').value.trim();
      if (!city) { out.innerHTML = '<span class="err">Enter a city.</span>'; return; }
      out.innerHTML = '<span class="loading">Loading...</span>';
      try {
        const res = await fetch('/weather?city=' + encodeURIComponent(city));
        const data = await res.json();
        if (data.error) { out.innerHTML = '<span class="err">' + data.error + '</span>'; return; }
        out.innerHTML =
          '<div class="temp">' + data.temp + ' ' + data.temp_unit + '</div>' +
          '<div class="cond">' + data.condition + '</div>' +
          '<div class="meta">' + data.city + ', ' + data.country + '<br>' +
          '💨 Wind: ' + data.wind + ' ' + data.wind_unit + '<br>' +
          '💧 Humidity: ' + data.humidity + '%<br>' +
          '<span style="font-size:11px">' + data.source + '</span></div>';
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
	http.HandleFunc("/weather", weatherHandler)
	println("Weather service running on http://localhost:8082")
	http.ListenAndServe(":8082", nil)
}
