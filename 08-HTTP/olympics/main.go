//go:build !solution

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

type athlete struct {
	Athlete string `json:"athlete"`
	Age     int    `json:"age"`
	Country string `json:"country"`
	Year    int    `json:"year"`
	Date    string `json:"date"`
	Sport   string `json:"sport"`
	medals
}

func main() {
	port := flag.Int("port", 0, "Server port (required)")
	data := flag.String("data", "", "Path to JSON")

	flag.Parse()

	if *port < 1 || *port > 65535 {
		log.Fatal("port is required and must be between 1 and 65535")
	}

	athletes, err := parseData(*data)
	if err != nil {
		log.Fatal("incorrect path or JSON: ", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /athlete-info", func(w http.ResponseWriter, r *http.Request) {
		athleteInfoHandler(w, r, athletes)
	})

	mux.HandleFunc("GET /top-athletes-in-sport", func(w http.ResponseWriter, r *http.Request) {
		topAthletesHandler(w, r, athletes)
	})

	mux.HandleFunc("GET /top-countries-in-year", func(w http.ResponseWriter, r *http.Request) {
		topCountriesHandler(w, r, athletes)
	})

	addr := fmt.Sprintf(":%d", *port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func parseData(path string) ([]athlete, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var athletes []athlete

	err = json.NewDecoder(file).Decode(&athletes)
	if err != nil {
		return nil, err
	}

	return athletes, nil
}
