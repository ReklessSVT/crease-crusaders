package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"time"
)

// --- EMBEDDED FILES ---
// This directive tells Go: "Include the templates folder inside the binary"
//
//go:embed templates/*
var resources embed.FS

// --- CACHE STORAGE ---
var cache struct {
	sync.Mutex
	Data      PageData
	Timestamp time.Time
}

// --- CONTROLLER ---

func handleHome(w http.ResponseWriter, r *http.Request) {
	data := getPageData()

	// Parse specifically from the embedded 'resources' variable
	tmpl, err := template.ParseFS(resources, "templates/index.html")
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

// --- LOGIC ---

func getPageData() PageData {
	cache.Lock()
	defer cache.Unlock()

	// 1. Check Cache (Is it fresh? < 1 Hour)
	if time.Since(cache.Timestamp) < 1*time.Hour && cache.Data.TeamName != "" {
		fmt.Println("⚡ Serving from Cache")
		return cache.Data
	}

	fmt.Println("🐢 Cache expired/empty. Fetching from APIs...")

	// 2. Fetch All Data
	players, _ := getRoster()
	rawGames, _ := getSchedule()
	standings, _ := getStandings()
	weather, _ := getForecast()

	// 3. Process Games + Weather
	var displayGames []GameDisplay
	loc, _ := time.LoadLocation("America/New_York")

	for _, g := range rawGames {
		t, _ := time.Parse(time.RFC3339, g.StartTime)
		localTime := t.In(loc)

		opponent := g.Details.Team2.Name
		homeAway := "vs"
		if g.Details.Team1.Name != "Crease Crusaders" {
			opponent = g.Details.Team1.Name
			homeAway = "@"
		}

		forecast := getWeatherString(localTime, weather)

		displayGames = append(displayGames, GameDisplay{
			Date:     localTime.Format("Mon, Jan 02"),
			Time:     localTime.Format("3:04 PM"),
			Opponent: opponent,
			HomeAway: homeAway,
			Weather:  forecast,
		})
	}

	// 4. Process Standings
	var displayStandings []StandingsDisplay
	bronzeID := "8nLg9ZsBicTerF07t22O"
	var bronzeTeams []TeamRecord

	for _, div := range standings {
		if div.ID == bronzeID {
			bronzeTeams = div.TeamRecords
			break
		}
	}

	for i, team := range bronzeTeams {
		gp := team.Stats.Wins + team.Stats.Losses + team.Stats.Ties
		displayStandings = append(displayStandings, StandingsDisplay{
			Rank: i + 1,
			Team: team.TeamName,
			GP:   gp,
			W:    team.Stats.Wins,
			L:    team.Stats.Losses,
			T:    team.Stats.Ties,
			IsUs: team.TeamName == "Crease Crusaders",
		})
	}

	// 5. Update Cache
	cache.Data = PageData{
		TeamName:  "Crease Crusaders",
		Roster:    players,
		Games:     displayGames,
		Standings: displayStandings,
		Updated:   time.Now().In(loc).Format("3:04 PM"),
	}
	cache.Timestamp = time.Now()

	return cache.Data
}
