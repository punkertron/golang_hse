package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type countryInfo struct {
	Country string `json:"country"`
	medals
}

type countryInfoHeap []*countryInfo

func (h countryInfoHeap) Len() int { return len(h) }

func (h countryInfoHeap) Less(i, j int) bool {
	a, b := h[i], h[j]

	if a.Gold != b.Gold {
		return a.Gold > b.Gold
	}

	if a.Silver != b.Silver {
		return a.Silver > b.Silver
	}

	if a.Bronze != b.Bronze {
		return a.Bronze > b.Bronze
	}

	return a.Country < b.Country
}

func (h *countryInfoHeap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*countryInfo))
}

func (h *countryInfoHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]

	return x
}

func (h countryInfoHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func topCountriesHandler(w http.ResponseWriter, r *http.Request, athletes []athlete) {
	query := r.URL.Query()

	yearStr := query.Get("year")
	if yearStr == "" {
		http.Error(w, "year is required", http.StatusBadRequest)
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year <= 0 {
		http.Error(w, "year must be a positive number", http.StatusBadRequest)
		return
	}

	var limit int

	limitStr := query.Get("limit")
	if limitStr == "" {
		limit = 3
	} else {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			http.Error(w, "limit must be a positive number number", http.StatusBadRequest)
			return
		}
	}

	m := make(map[string]*countryInfo)

	for _, athlete := range athletes {
		if athlete.Year == year {
			if m[athlete.Country] == nil {
				m[athlete.Country] = &countryInfo{
					Country: athlete.Country,
				}
			}

			country := m[athlete.Country]
			country.Gold += athlete.Gold
			country.Silver += athlete.Silver
			country.Bronze += athlete.Bronze
			country.Total += athlete.Total
		}
	}

	if len(m) == 0 {
		http.Error(w, fmt.Sprintf("year %d is not found", year), http.StatusNotFound)
		return
	}

	h := make(countryInfoHeap, 0, len(m))
	for _, athlete := range m {
		h = append(h, athlete)
	}

	heap.Init(&h)

	limit = min(limit, len(m))

	res := make(countryInfoHeap, 0, limit)
	for range limit {
		res = append(res, heap.Pop(&h).(*countryInfo))
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println("problem with json encoding: ", err)
		http.Error(w, "problem with json encoding", http.StatusInternalServerError)

		return
	}
}
