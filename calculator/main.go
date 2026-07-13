package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Response struct {
	Operation string  `json:"operation"`
	A         float64 `json:"a"`
	B         float64 `json:"b"`
	Result    float64 `json:"result,omitempty"`
	Error     string  `json:"error,omitempty"`
}

func calculate(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	op := q.Get("op")
	aStr := q.Get("a")
	bStr := q.Get("b")

	a, err := strconv.ParseFloat(aStr, 64)
	if err != nil {
		http.Error(w, "Invalid value for 'a'", http.StatusBadRequest)
		return
	}

	b, err := strconv.ParseFloat(bStr, 64)
	if err != nil {
		http.Error(w, "Invalid value for 'b'", http.StatusBadRequest)
		return
	}

	resp := Response{
		Operation: op,
		A:         a,
		B:         b,
	}

	switch op {
	case "add":
		resp.Result = a + b
	case "sub":
		resp.Result = a - b
	case "mul":
		resp.Result = a * b
	case "div":
		if b == 0 {
			resp.Error = "Division by zero"
		} else {
			resp.Result = a / b
		}
	default:
		resp.Error = "Unsupported operation. Use add, sub, mul, or div."
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	http.HandleFunc("/calculate", calculate)

	println("Calculator service running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
