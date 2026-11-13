package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type athleteInfo struct {
	Athlete      string          `json:"athlete"`
	Country      string          `json:"country"`
	Medals       medals          `json:"medals"`
	MedalsByYear map[int]*medals `json:"medals_by_year"`
}

type medals struct {
	Gold   int `json:"gold"`
	Silver int `json:"silver"`
	Bronze int `json:"bronze"`
	Total  int `json:"total"`
}

func athleteInfoHandler(w http.ResponseWriter, r *http.Request, athletes []athlete) {
	query := r.URL.Query()

	name := query.Get("name")
	if name == "" {
		http.Error(w, "name is empty", http.StatusBadRequest)
		return
	}

	info := athleteInfo{
		MedalsByYear: make(map[int]*medals),
	}

	for _, athlete := range athletes {
		if athlete.Athlete == name {
			if info.Country == "" {
				info.Country = athlete.Country
			}

			info.Medals.Gold += athlete.Gold
			info.Medals.Silver += athlete.Silver
			info.Medals.Bronze += athlete.Bronze
			info.Medals.Total += athlete.Total

			if info.MedalsByYear[athlete.Year] == nil {
				info.MedalsByYear[athlete.Year] = &medals{}
			}

			info.MedalsByYear[athlete.Year].Gold += athlete.Gold
			info.MedalsByYear[athlete.Year].Silver += athlete.Silver
			info.MedalsByYear[athlete.Year].Bronze += athlete.Bronze
			info.MedalsByYear[athlete.Year].Total += athlete.Total
		}
	}

	if info.Country == "" {
		http.Error(w, fmt.Sprintf("athlete %s not found", name), http.StatusNotFound)
		return
	}

	info.Athlete = name

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(info)
	if err != nil {
		log.Println("problem with json encoding: ", err)
		http.Error(w, "problem with json encoding", http.StatusInternalServerError)

		return
	}
}
