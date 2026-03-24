package main

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

// --- SCHEDULE (SportsEngine) ---
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

// --- DETAILED STATS API (Scorekeeper) ---
type DetailedStatsResponse struct {
	Games []DetailedGame `json:"games"`
}

type DetailedGame struct {
	Division  string         `json:"division"`
	HomeTeam  string         `json:"homeTeam"`
	AwayTeam  string         `json:"awayTeam"`
	HomeScore int            `json:"homeScore"`
	AwayScore int            `json:"awayScore"`
	Status    string         `json:"status"`
	Goals     []DetailedGoal `json:"goals"`
	Overtime  *OvertimeInfo  `json:"overtimeResult"` // Captures the OT winner
}

type OvertimeInfo struct {
	Winner string `json:"winner"`
}

type DetailedGoal struct {
	Team   string `json:"team"`
	Player string `json:"player"`
	Assist string `json:"assist"`
}

// --- WEATHER ---
type WeatherResponse struct {
	Hourly HourlyWeather `json:"hourly"`
}
type HourlyWeather struct {
	Times       []string  `json:"time"`
	Temps       []float64 `json:"temperature_2m"`
	WeatherCode []int     `json:"weathercode"`
}

// --- PAGE DISPLAY MODELS ---
type GameDisplay struct {
	Date     string
	Time     string
	Opponent string
	HomeAway string
	Weather  string
}

type StandingsDisplay struct {
	Rank int
	Team string
	GP   int
	W    int
	L    int
	OTL  int
	Pts  int
	IsUs bool
}

type PlayerStat struct {
	Name    string
	Jersey  string
	Goals   int
	Assists int
	Points  int
}

type PageData struct {
	TeamName  string
	Roster    []PlayerStat
	Games     []GameDisplay
	Standings []StandingsDisplay
	Updated   string
}
