package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type athleteInfoHeap []*athleteInfo

func (h athleteInfoHeap) Len() int { return len(h) }

func (h athleteInfoHeap) Less(i, j int) bool {
	a, b := h[i], h[j]

	if a.Medals.Gold != b.Medals.Gold {
		return a.Medals.Gold > b.Medals.Gold
	}

	if a.Medals.Silver != b.Medals.Silver {
		return a.Medals.Silver > b.Medals.Silver
	}

	if a.Medals.Bronze != b.Medals.Bronze {
		return a.Medals.Bronze > b.Medals.Bronze
	}

	return a.Athlete < b.Athlete
}

func (h *athleteInfoHeap) Push(x any) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(*athleteInfo))
}

func (h *athleteInfoHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]

	return x
}

func (h athleteInfoHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func topAthletesHandler(w http.ResponseWriter, r *http.Request, athletes []athlete) {
	query := r.URL.Query()

	sport := query.Get("sport")
	if sport == "" {
		http.Error(w, "sport is required", http.StatusBadRequest)
		return
	}

	var limit int

	limitStr := query.Get("limit")
	if limitStr == "" {
		limit = 3
	} else {
		var err error

		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			http.Error(w, "limit must be a positive number number", http.StatusBadRequest)
			return
		}
	}

	m := make(map[string]*athleteInfo)

	for _, athlete := range athletes {
		if athlete.Sport == sport {
			if m[athlete.Athlete] == nil {
				m[athlete.Athlete] = &athleteInfo{
					Athlete:      athlete.Athlete,
					Country:      athlete.Country,
					MedalsByYear: make(map[int]*medals),
				}
			}

			ath := m[athlete.Athlete]
			ath.Medals.Gold += athlete.Gold
			ath.Medals.Silver += athlete.Silver
			ath.Medals.Bronze += athlete.Bronze
			ath.Medals.Total += athlete.Total

			if ath.MedalsByYear[athlete.Year] == nil {
				ath.MedalsByYear[athlete.Year] = &medals{}
			}

			athMedals := ath.MedalsByYear[athlete.Year]

			athMedals.Gold += athlete.Gold
			athMedals.Silver += athlete.Silver
			athMedals.Bronze += athlete.Bronze
			athMedals.Total += athlete.Total
		}
	}

	if len(m) == 0 {
		http.Error(w, fmt.Sprintf("sport '%s' is not found", sport), http.StatusNotFound)
		return
	}

	h := make(athleteInfoHeap, 0, len(m))
	for _, athlete := range m {
		h = append(h, athlete)
	}

	heap.Init(&h)

	limit = min(limit, len(m))

	res := make(athleteInfoHeap, 0, limit)
	for range limit {
		res = append(res, heap.Pop(&h).(*athleteInfo))
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println("problem with json encoding: ", err)
		http.Error(w, "problem with json encoding", http.StatusInternalServerError)

		return
	}
}
