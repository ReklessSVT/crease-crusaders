package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

// ==========================================
// PART 1: API DATA MODELS
// ==========================================

// --- ROSTER ---
type RosterResponse struct {
	Players []Player `json:"result"`
}
type Player struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Jersey    string `json:"jersey_number"`
	Position  string `json:"positions"`
	TeamName  string `json:"team_name"`
}

// --- SCHEDULE ---
type ScheduleResponse struct {
	Games []RawGame `json:"result"`
}
type RawGame struct {
	ID        string `json:"id"`
	StartTime string `json:"start_date_time"`
	Details   struct {
		Team1 struct {
			Name   string `json:"name"`
			IsHome bool   `json:"is_home_team"`
		} `json:"team_1"`
		Team2 struct {
			Name   string `json:"name"`
			IsHome bool   `json:"is_home_team"`
		} `json:"team_2"`
	} `json:"game_details"`
}

// --- STANDINGS ---
type StandingsResponse struct {
	Divisions []Division `json:"result"`
}

type Division struct {
	ID          string       `json:"id"`
	TeamRecords []TeamRecord `json:"teamRecords"`
}

type TeamRecord struct {
	TeamName string      `json:"team_name"`
	Stats    RecordStats `json:"values"`
}

type RecordStats struct {
	Wins   int `json:"w"`
	Losses int `json:"l"`
	Ties   int `json:"t"`
}

// ==========================================
// PART 2: DISPLAY MODELS
// ==========================================

type GameDisplay struct {
	Date     string
	Time     string
	Opponent string
	HomeAway string
}

type StandingsDisplay struct {
	Rank int
	Team string
	GP   int // Games Played
	W    int
	L    int
	T    int
	// Pts removed as requested
	IsUs bool
}

type PageData struct {
	TeamName  string
	Roster    []Player
	Games     []GameDisplay
	Standings []StandingsDisplay
	Updated   string
}

// ==========================================
// PART 3: SERVER & LOGIC
// ==========================================

func main() {
	http.HandleFunc("/", handleHome)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🏒 Crease Crusaders Hub is live on port %s\n", port)

	// Listen on the variable port
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	// 1. Fetch Data
	players, _ := getRoster()
	rawGames, _ := getSchedule()
	standings, _ := getStandings()

	// 2. Process Games
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

		displayGames = append(displayGames, GameDisplay{
			Date:     localTime.Format("Mon, Jan 02"),
			Time:     localTime.Format("3:04 PM"),
			Opponent: opponent,
			HomeAway: homeAway,
		})
	}

	// 3. Process Standings
	var displayStandings []StandingsDisplay

	// Filter for Bronze Division ID
	bronzeID := "8nLg9ZsBicTerF07t22O"
	var bronzeTeams []TeamRecord

	for _, div := range standings {
		if div.ID == bronzeID {
			bronzeTeams = div.TeamRecords
			break
		}
	}

	// Convert to Display format
	for i, team := range bronzeTeams {
		gamesPlayed := team.Stats.Wins + team.Stats.Losses + team.Stats.Ties

		displayStandings = append(displayStandings, StandingsDisplay{
			Rank: i + 1,
			Team: team.TeamName,
			GP:   gamesPlayed,
			W:    team.Stats.Wins,
			L:    team.Stats.Losses,
			T:    team.Stats.Ties,
			IsUs: team.TeamName == "Crease Crusaders",
		})
	}

	// 4. Render
	data := PageData{
		TeamName:  "Crease Crusaders",
		Roster:    players,
		Games:     displayGames,
		Standings: displayStandings,
		Updated:   time.Now().Format("3:04 PM"),
	}

	tmpl, err := template.New("index").Parse(htmlTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, data)
}

// ==========================================
// PART 4: API CLIENTS
// ==========================================

func getRoster() ([]Player, error) {
	url := "https://se-api.sportsengine.com/v3/microsites/roster_players?roster_id=69763da031a69300010a09c8"
	var response RosterResponse
	err := fetchJSON(url, &response)
	return response.Players, err
}

func getSchedule() ([]RawGame, error) {
	url := "https://se-api.sportsengine.com/v3/microsites/events?page=1&per_page=10&program_id=69763d9a3dc6b20df8c68bb9&order_by=starts_at&direction=asc&team_id=11f0fa06-ae85-42fa-bcf3-9e3f2a32c39c&starts_at=2026-02-03T05:00:00.000Z"
	var response ScheduleResponse
	err := fetchJSON(url, &response)
	return response.Games, err
}

func getStandings() ([]Division, error) {
	url := "https://se-api.sportsengine.com/v3/microsites/standings?program_id=69763d9a3dc6b20df8c68bb9"
	var response StandingsResponse
	err := fetchJSON(url, &response)
	return response.Divisions, err
}

func fetchJSON(url string, target interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}

// ==========================================
// PART 5: FRONTEND
// ==========================================

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>{{.TeamName}}</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        /* BASE STYLES */
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; background: #f0f2f5; color: #333; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 15px; }
        h1 { text-align: center; color: #003366; margin: 10px 0 20px 0; }
        
        /* TABS NAVIGATION */
        .tabs { display: flex; background: white; border-bottom: 1px solid #ddd; position: sticky; top: 0; z-index: 100; }
        .tab-btn { 
            flex: 1; text-align: center; padding: 15px; cursor: pointer; 
            font-weight: 600; color: #666; border-bottom: 3px solid transparent; 
            background: none; border-top: none; border-left: none; border-right: none;
            font-size: 1rem;
        }
        .tab-btn.active { color: #003366; border-bottom-color: #003366; }
        
        /* SECTIONS */
        .tab-content { display: none; animation: fadeIn 0.3s; }
        .tab-content.active { display: block; }
        @keyframes fadeIn { from { opacity: 0; } to { opacity: 1; } }

        /* CARDS & TABLES (Same as before) */
        .card { background: white; border-radius: 10px; box-shadow: 0 2px 5px rgba(0,0,0,0.05); margin-bottom: 10px; overflow: hidden; }
        table { width: 100%; border-collapse: collapse; font-size: 0.9em; }
        th { background: #f8f9fa; color: #666; font-weight: 600; text-align: center; padding: 12px 5px; border-bottom: 1px solid #eee; }
        td { padding: 12px 5px; text-align: center; border-bottom: 1px solid #eee; }
        th.text-left, td.text-left { text-align: left; }
        .my-team { background-color: #e3f2fd; font-weight: bold; }
        
        .game-row { display: flex; align-items: center; padding: 15px; border-bottom: 1px solid #eee; }
        .date-box { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 6px; padding: 8px 12px; text-align: center; margin-right: 15px; min-width: 60px; }
        .date-day { font-weight: bold; display: block; font-size: 0.9em; }
        .date-time { font-size: 0.8em; color: #666; }
        .matchup { flex-grow: 1; font-weight: 500; }
        .vs-badge { font-size: 0.8em; background: #e9ecef; color: #555; padding: 2px 6px; border-radius: 4px; margin-right: 6px; }

        .player-row { display: flex; justify-content: space-between; padding: 12px 15px; border-bottom: 1px solid #eee; }
        .jersey { font-weight: bold; color: #003366; width: 30px; display:inline-block; }
    </style>
</head>
<body>

    <div class="tabs">
        <button class="tab-btn active" onclick="openTab('games')">Games</button>
        <button class="tab-btn" onclick="openTab('standings')">Standings</button>
        <button class="tab-btn" onclick="openTab('roster')">Roster</button>
    </div>

    <div class="container">
        <h1>{{.TeamName}}</h1>

        <div id="games" class="tab-content active">
            <div class="card">
                {{range .Games}}
                <div class="game-row">
                    <div class="date-box">
                        <span class="date-day">{{.Date}}</span>
                        <span class="date-time">{{.Time}}</span>
                    </div>
                    <div class="matchup">
                        <span class="vs-badge">{{.HomeAway}}</span> {{.Opponent}}
                    </div>
                </div>
                {{else}}
                <div style="padding: 20px; text-align: center; color: #888;">No upcoming games found.</div>
                {{end}}
            </div>
        </div>

        <div id="standings" class="tab-content">
            <div class="card">
                <table>
                    <thead>
                        <tr>
                            <th class="text-left" style="padding-left:15px;">Team</th>
                            <th>GP</th>
                            <th>W</th>
                            <th>L</th>
                            <th>T</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Standings}}
                        <tr class="{{if .IsUs}}my-team{{end}}">
                            <td class="text-left" style="padding-left:15px;">{{.Team}}</td>
                            <td>{{.GP}}</td>
                            <td>{{.W}}</td>
                            <td>{{.L}}</td>
                            <td>{{.T}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </div>
        </div>

        <div id="roster" class="tab-content">
            <div class="card">
                {{range .Roster}}
                <div class="player-row">
                    <div>
                        <span class="jersey">{{if .Jersey}}#{{.Jersey}}{{else}}--{{end}}</span>
                        {{.FirstName}} {{.LastName}}
                    </div>
                    <div style="color:#888; font-size:0.9em;">{{if .Position}}{{.Position}}{{end}}</div>
                </div>
                {{end}}
            </div>
        </div>

        <p style="text-align:center; color:#999; font-size:0.8em; margin-top:30px;">
            Updated at {{.Updated}}
        </p>
    </div>

    <script>
        function openTab(tabName) {
            // 1. Hide all tab contents
            var contents = document.getElementsByClassName("tab-content");
            for (var i = 0; i < contents.length; i++) {
                contents[i].classList.remove("active");
            }

            // 2. Deactivate all buttons
            var buttons = document.getElementsByClassName("tab-btn");
            for (var i = 0; i < buttons.length; i++) {
                buttons[i].classList.remove("active");
            }

            // 3. Show the specific tab and activate the button
            document.getElementById(tabName).classList.add("active");
            
            // Find the button that was clicked (using event.target would be simpler but let's loop to match name)
            // Actually, we can just highlight the button based on index, but simplest way 
            // is to match text or pass 'this' into the function.
            // Let's use a simpler CSS selector approach for the button:
            event.currentTarget.classList.add("active");
        }
    </script>
</body>
</html>
`
