package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"
)

// --- EMBEDDED FILES ---
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

	// 1. Check Cache
	if time.Since(cache.Timestamp) < 1*time.Hour && cache.Data.TeamName != "" {
		fmt.Println("⚡ Serving from Cache")
		return cache.Data
	}

	fmt.Println("🐢 Cache expired/empty. Fetching from APIs...")

	// 2. Fetch All Data
	rawRoster, _ := getRoster()
	rawGames, _ := getSchedule()
	standings, _ := getStandings()
	weather, _ := getForecast()
	detailedGames, _ := getDetailedStats()

	// 3. Process Roster & Stats
	statsMap := make(map[string]*PlayerStat)
	myTeamName := "Crease Crusaders"

	// Initialize map with exact roster data
	for _, p := range rawRoster {
		fullName := p.FirstName + " " + p.LastName
		statsMap[fullName] = &PlayerStat{
			Name:    fullName,
			Jersey:  p.Jersey,
			Goals:   0,
			Assists: 0,
			Points:  0,
		}
	}

	// Tally Goals and Assists using the fuzzy matcher
	for _, g := range detailedGames {
		if g.HomeTeam == myTeamName || g.AwayTeam == myTeamName {
			for _, goal := range g.Goals {
				if goal.Team == myTeamName {

					// Use our fuzzy matcher to get the Roster Key ("Jeremy V")
					playerKey := matchPlayerName(goal.Player, rawRoster)

					if stat, exists := statsMap[playerKey]; exists {
						stat.Goals++
						stat.Points++
					} else {
						// Sub player not on official roster
						statsMap[playerKey] = &PlayerStat{Name: goal.Player, Goals: 1, Points: 1}
					}

					if goal.Assist != "" {
						assistKey := matchPlayerName(goal.Assist, rawRoster)
						if stat, exists := statsMap[assistKey]; exists {
							stat.Assists++
							stat.Points++
						} else {
							statsMap[assistKey] = &PlayerStat{Name: goal.Assist, Assists: 1, Points: 1}
						}
					}
				}
			}
		}
	}

	var displayRoster []PlayerStat
	for _, stat := range statsMap {
		displayRoster = append(displayRoster, *stat)
	}

	// 4. Process Schedule + Weather
	var displayGames []GameDisplay
	loc, _ := time.LoadLocation("America/New_York")

	for _, g := range rawGames {
		t, _ := time.Parse(time.RFC3339, g.StartTime)
		localTime := t.In(loc)

		opponent := g.Details.Team2.Name
		homeAway := "vs"
		if g.Details.Team1.Name != myTeamName {
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

	// 5. Process Standings
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
			IsUs: team.TeamName == myTeamName,
		})
	}

	// 6. Update Cache
	cache.Data = PageData{
		TeamName:  myTeamName,
		Roster:    displayRoster,
		Games:     displayGames,
		Standings: displayStandings,
		Updated:   time.Now().In(loc).Format("3:04 PM"),
	}
	cache.Timestamp = time.Now()

	return cache.Data
}

// --- FUZZY MATCHER HELPER ---

// matchPlayerName tries to match a full name ("Jeremy Vermillion")
// to a roster name ("Jeremy V")
func matchPlayerName(scorekeeperName string, roster []Player) string {
	parts := strings.Split(scorekeeperName, " ")
	if len(parts) < 2 {
		return scorekeeperName // Can't split, just return original
	}

	firstName := parts[0]
	lastName := parts[len(parts)-1]

	for _, p := range roster {
		// Do the first names match exactly? (Case-insensitive)
		if strings.EqualFold(p.FirstName, firstName) {
			// Does the scorekeeper last name start with the roster last name?
			// (e.g. "Vermillion" starts with "V")
			if len(p.LastName) > 0 && strings.HasPrefix(strings.ToLower(lastName), strings.ToLower(p.LastName)) {
				// Match found! Return the key we use in statsMap
				return p.FirstName + " " + p.LastName
			}
		}
	}

	// No match found in roster (likely a substitute player), return original name
	return scorekeeperName
}
