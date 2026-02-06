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

// --- WEATHER ---
type WeatherResponse struct {
	Hourly HourlyWeather `json:"hourly"`
}
type HourlyWeather struct {
	Times       []string  `json:"time"`
	Temps       []float64 `json:"temperature_2m"`
	WeatherCode []int     `json:"weathercode"`
}

// --- PAGE DISPLAY ---
type GameDisplay struct {
	Date     string
	Time     string
	Opponent string
	HomeAway string
	Weather  string // The new field for forecast!
}

type StandingsDisplay struct {
	Rank int
	Team string
	GP   int
	W    int
	L    int
	T    int
	IsUs bool
}

type PageData struct {
	TeamName  string
	Roster    []Player
	Games     []GameDisplay
	Standings []StandingsDisplay
	Updated   string
}
