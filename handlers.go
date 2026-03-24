package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

//go:embed templates/*
var resources embed.FS

var cache struct {
	sync.Mutex
	Data      PageData
	Timestamp time.Time
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	data := getPageData()

	tmpl, err := template.ParseFS(resources, "templates/index.html")
	if err != nil {
		http.Error(w, "Template Error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

func getPageData() PageData {
	cache.Lock()
	defer cache.Unlock()

	if time.Since(cache.Timestamp) < 1*time.Hour && cache.Data.TeamName != "" {
		fmt.Println("⚡ Serving from Cache")
		return cache.Data
	}

	fmt.Println("🐢 Cache expired/empty. Fetching from APIs...")

	rawRoster, _ := getRoster()
	rawGames, _ := getSchedule()
	weather, _ := getForecast()
	detailedGames, _ := getDetailedStats() // Has everything we need!

	myTeamName := "Crease Crusaders"

	// ==========================================
	// 1. Process Roster & Stats
	// ==========================================
	statsMap := make(map[string]*PlayerStat)

	for _, p := range rawRoster {
		fullName := p.FirstName + " " + p.LastName
		statsMap[fullName] = &PlayerStat{
			Name:   fullName,
			Jersey: p.Jersey,
		}
	}

	for _, g := range detailedGames {
		if g.HomeTeam == myTeamName || g.AwayTeam == myTeamName {
			for _, goal := range g.Goals {
				if goal.Team == myTeamName {
					playerKey := matchPlayerName(goal.Player, rawRoster)
					if stat, exists := statsMap[playerKey]; exists {
						stat.Goals++
						stat.Points++
					} else {
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

	// Sort Roster by Points
	sort.Slice(displayRoster, func(i, j int) bool {
		if displayRoster[i].Points == displayRoster[j].Points {
			return displayRoster[i].Goals > displayRoster[j].Goals
		}
		return displayRoster[i].Points > displayRoster[j].Points
	})

	// ==========================================
	// 2. Process Custom Standings Logic
	// ==========================================
	standingsMap := make(map[string]*StandingsDisplay)

	for _, g := range detailedGames {
		// Only calculate finished games in the Bronze division
		if g.Division == "Bronze" && g.Status == "submitted" {

			if _, exists := standingsMap[g.HomeTeam]; !exists {
				standingsMap[g.HomeTeam] = &StandingsDisplay{Team: g.HomeTeam, IsUs: g.HomeTeam == myTeamName}
			}
			if _, exists := standingsMap[g.AwayTeam]; !exists {
				standingsMap[g.AwayTeam] = &StandingsDisplay{Team: g.AwayTeam, IsUs: g.AwayTeam == myTeamName}
			}

			h := standingsMap[g.HomeTeam]
			a := standingsMap[g.AwayTeam]

			h.GP++
			a.GP++

			// Check for Overtime/Shootout
			if g.Overtime != nil {
				if g.Overtime.Winner == g.HomeTeam {
					h.W++      // Home Win
					h.Pts += 2 // 2 Pts
					a.OTL++    // Away OT Loss
					a.Pts += 1 // 1 Pt
				} else if g.Overtime.Winner == g.AwayTeam {
					a.W++
					a.Pts += 2
					h.OTL++
					h.Pts += 1
				}
			} else {
				// Regulation Game
				if g.HomeScore > g.AwayScore {
					h.W++
					h.Pts += 2
					a.L++ // 0 Pts
				} else if g.AwayScore > g.HomeScore {
					a.W++
					a.Pts += 2
					h.L++ // 0 Pts
				}
			}
		}
	}

	var displayStandings []StandingsDisplay
	for _, s := range standingsMap {
		displayStandings = append(displayStandings, *s)
	}

	// Sort Standings by Points, then by Wins
	sort.Slice(displayStandings, func(i, j int) bool {
		if displayStandings[i].Pts == displayStandings[j].Pts {
			return displayStandings[i].W > displayStandings[j].W
		}
		return displayStandings[i].Pts > displayStandings[j].Pts
	})

	for i := range displayStandings {
		displayStandings[i].Rank = i + 1
	}

	// ==========================================
	// 3. Process Schedule + Weather
	// ==========================================
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

	// ==========================================
	// 4. Update Cache
	// ==========================================
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
func matchPlayerName(scorekeeperName string, roster []Player) string {
	parts := strings.Split(scorekeeperName, " ")
	if len(parts) < 2 {
		return scorekeeperName
	}
	firstName := parts[0]
	lastName := parts[len(parts)-1]

	for _, p := range roster {
		if strings.EqualFold(p.FirstName, firstName) {
			if len(p.LastName) > 0 && strings.HasPrefix(strings.ToLower(lastName), strings.ToLower(p.LastName)) {
				return p.FirstName + " " + p.LastName
			}
		}
	}
	return scorekeeperName
}
