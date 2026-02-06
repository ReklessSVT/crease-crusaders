package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Coordinates for Chattanooga, TN
const (
	Lat  = "35.0456"
	Long = "-85.3097"
)

// --- FETCHERS ---

func getRoster() ([]Player, error) {
	url := "https://se-api.sportsengine.com/v3/microsites/roster_players?roster_id=69763da031a69300010a09c8"
	var response RosterResponse
	err := fetchJSON(url, &response)
	return response.Players, err
}

func getSchedule() ([]RawGame, error) {
	// Note: Fetching page 1, 10 items. You might increase 'per_page' if the season is long.
	url := "https://se-api.sportsengine.com/v3/microsites/events?page=1&per_page=15&program_id=69763d9a3dc6b20df8c68bb9&order_by=starts_at&direction=asc&team_id=11f0fa06-ae85-42fa-bcf3-9e3f2a32c39c&starts_at=2026-02-03T05:00:00.000Z"
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

func getForecast() (HourlyWeather, error) {
	// Free API, no key needed
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%s&longitude=%s&hourly=temperature_2m,weathercode&temperature_unit=fahrenheit&timezone=America%%2FNew_York", Lat, Long)
	var response WeatherResponse
	err := fetchJSON(url, &response)
	return response.Hourly, err
}

// --- HELPERS ---

func fetchJSON(url string, target interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func getWeatherString(gameTime time.Time, weather HourlyWeather) string {
	// Only show weather if game is within the next 7 days
	if time.Until(gameTime) > 7*24*time.Hour || time.Until(gameTime) < 0 {
		return ""
	}

	// Match game time prefix (YYYY-MM-DDTHH)
	targetPrefix := gameTime.Format("2006-01-02T15")

	for i, tStr := range weather.Times {
		if len(tStr) >= 13 && tStr[:13] == targetPrefix {
			return fmt.Sprintf("%.0f°F %s", weather.Temps[i], getWeatherEmoji(weather.WeatherCode[i]))
		}
	}
	return ""
}

func getWeatherEmoji(code int) string {
	switch {
	case code == 0:
		return "☀️"
	case code >= 1 && code <= 3:
		return "⛅"
	case code >= 45 && code <= 48:
		return "🌫️"
	case code >= 51 && code <= 67:
		return "🌧️"
	case code >= 71 && code <= 77:
		return "❄️"
	case code >= 95:
		return "⛈️"
	default:
		return "🌡️"
	}
}
