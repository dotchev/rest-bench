package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
)

func main() {
	fmt.Println("Version:", runtime.Version())
	fmt.Println("NumCPU:", runtime.NumCPU())
	fmt.Println("GOMAXPROCS:", runtime.GOMAXPROCS(0))

	fmt.Println("Listening on port 3000")
	err := http.ListenAndServe(":3000", http.HandlerFunc(handler))
	fmt.Println(err)
}

func handler(w http.ResponseWriter, r *http.Request) {
	var body Body
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(body)
	if err != nil {
		panic(err)
	}
}

type Body struct {
	Colors []Color `json:"colors"`
}

type Color struct {
	Color    string `json:"color,omitempty"`
	Category string `json:"category,omitempty"`
	Type     string `json:"type,omitempty"`
	Code     Code   `json:"code,omitempty"`
}

type Code struct {
	RGBA []int  `json:"rgba,omitempty"`
	Hex  string `json:"hex,omitempty"`
}
