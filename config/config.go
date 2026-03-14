package config

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"
)

// LoadEnvFile reads a .env file and sets any variables not already in the environment.
// Lines starting with # and blank lines are skipped.
func LoadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		// .env file is optional; silently skip if missing
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		// Don't overwrite existing OS-level env vars
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Warning: error reading %s: %v", path, err)
	}
}

// SportConfig defines one sport's ESPN API path and env var for team selection.
type SportConfig struct {
	Name     string // Display name, e.g. "NFL"
	Sport    string // ESPN sport path segment, e.g. "football"
	League   string // ESPN league path segment, e.g. "nfl"
	EnvVar   string // Environment variable name, e.g. "NFL_TEAMS"
	IsRacing bool   // true for racing series (boolean toggle, no team selection)
}

// AllSports lists every supported sport configuration.
var AllSports = []SportConfig{
	{Name: "NFL", Sport: "football", League: "nfl", EnvVar: "NFL_TEAMS"},
	{Name: "NBA", Sport: "basketball", League: "nba", EnvVar: "NBA_TEAMS"},
	{Name: "MLB", Sport: "baseball", League: "mlb", EnvVar: "MLB_TEAMS"},
	{Name: "NHL", Sport: "hockey", League: "nhl", EnvVar: "NHL_TEAMS"},
	{Name: "NCAAF", Sport: "football", League: "college-football", EnvVar: "NCAAF_TEAMS"},
	{Name: "NCAAM", Sport: "basketball", League: "mens-college-basketball", EnvVar: "NCAAM_TEAMS"},
	{Name: "MLS", Sport: "soccer", League: "usa.1", EnvVar: "MLS_TEAMS"},
	{Name: "EPL", Sport: "soccer", League: "eng.1", EnvVar: "EPL_TEAMS"},
	{Name: "UFC", Sport: "mma", League: "ufc", EnvVar: "UFC_FIGHTERS"},
	{Name: "IndyCar", Sport: "racing", League: "irl", EnvVar: "INDYCAR", IsRacing: true},
	{Name: "NASCAR Cup", Sport: "racing", League: "nascar-premier", EnvVar: "NASCAR", IsRacing: true},
	{Name: "NASCAR Xfinity", Sport: "racing", League: "nascar-secondary", EnvVar: "NASCAR_XFINITY", IsRacing: true},
	{Name: "NASCAR Truck", Sport: "racing", League: "nascar-truck", EnvVar: "NASCAR_TRUCK", IsRacing: true},
	{Name: "Formula 1", Sport: "racing", League: "f1", EnvVar: "F1", IsRacing: true},
}

// ActiveSport is a sport that has at least one team configured.
type ActiveSport struct {
	SportConfig
	Teams []string // Uppercase abbreviations
}

// LoadActiveSports reads environment variables and returns only sports with teams configured.
func LoadActiveSports() []ActiveSport {
	var active []ActiveSport
	for _, sc := range AllSports {
		raw := strings.TrimSpace(os.Getenv(sc.EnvVar))
		if raw == "" {
			continue
		}
		if sc.IsRacing {
			if strings.EqualFold(raw, "true") || raw == "1" {
				active = append(active, ActiveSport{SportConfig: sc})
			}
			continue
		}
		parts := strings.Split(raw, ",")
		var teams []string
		for _, p := range parts {
			t := strings.ToUpper(strings.TrimSpace(p))
			if t != "" {
				teams = append(teams, t)
			}
		}
		if len(teams) > 0 {
			active = append(active, ActiveSport{SportConfig: sc, Teams: teams})
		}
	}
	return active
}

// ListenAddr returns the configured listen address, defaulting to 0.0.0.0:8080.
func ListenAddr() string {
	addr := os.Getenv("LISTEN_ADDR")
	if addr == "" {
		return "0.0.0.0:8080"
	}
	return addr
}

// LoadTimezone returns the *time.Location configured via the TIMEZONE env var.
// Defaults to UTC if not set or invalid.
func LoadTimezone() *time.Location {
	tz := strings.TrimSpace(os.Getenv("TIMEZONE"))
	if tz == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Printf("WARNING: invalid TIMEZONE %q, falling back to UTC: %v", tz, err)
		return time.UTC
	}
	return loc
}
