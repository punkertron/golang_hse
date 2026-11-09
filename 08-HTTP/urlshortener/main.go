//go:build !solution

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
)

type store struct {
	URLToKey map[string]string
	keyToURL map[string]string
	mu       sync.RWMutex
}

type request struct {
	URL string `json:"url"`
}

type response struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

func main() {
	port := flag.Int("port", 0, "Server port (required)")

	flag.Parse()

	if *port < 1 || *port > 65535 {
		log.Fatal("port is required and must be between 1 and 65535")
	}

	d := &store{
		URLToKey: make(map[string]string),
		keyToURL: make(map[string]string),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /shorten", func(w http.ResponseWriter, r *http.Request) {
		shortenHandler(w, r, d)
	})

	mux.HandleFunc("GET /go/{key}", func(w http.ResponseWriter, r *http.Request) {
		goHandler(w, r, d)
	})

	addr := fmt.Sprintf(":%d", *port)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func shortenHandler(w http.ResponseWriter, r *http.Request, d *store) {
	var req request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		http.Error(w, "Invalid json", http.StatusBadRequest)

		return
	}

	d.mu.Lock()

	key, ok := d.URLToKey[req.URL]
	if !ok {
		for {
			key = generateKey()

			if _, ok := d.keyToURL[key]; !ok {
				break
			}
		}

		d.keyToURL[key] = req.URL
		d.URLToKey[req.URL] = key
	}
	d.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response{URL: req.URL, Key: key}); err != nil {
		log.Println(err)
	}
}

func goHandler(w http.ResponseWriter, r *http.Request, d *store) {
	key := r.PathValue("key")

	d.mu.RLock()
	url, ok := d.keyToURL[key]
	d.mu.RUnlock()

	if ok {
		http.Redirect(w, r, url, http.StatusFound)
	} else {
		http.NotFound(w, r)
	}
}

func generateKey() string {
	const letterBytes = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, 6)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}
