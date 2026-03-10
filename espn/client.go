package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const baseURL = "https://site.api.espn.com/apis/site/v2/sports"

// Client fetches and caches ESPN scoreboard data.
type Client struct {
	httpClient    *http.Client
	cache         map[string]cacheEntry
	scheduleCache map[string]scheduleCacheEntry
	mu            sync.RWMutex
	cacheTTL      time.Duration
}

type cacheEntry struct {
	data      *ScoreboardResponse
	fetchedAt time.Time
}

type scheduleCacheEntry struct {
	data      *ScheduleResponse
	fetchedAt time.Time
}

// NewClient creates an ESPN API client with the given cache TTL.
func NewClient(cacheTTL time.Duration) *Client {
	return &Client{
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		cache:         make(map[string]cacheEntry),
		scheduleCache: make(map[string]scheduleCacheEntry),
		cacheTTL:      cacheTTL,
	}
}

// FetchScoreboard retrieves the scoreboard for a sport/league.
// sport and league are the ESPN path segments, e.g. "football" and "nfl".
func (c *Client) FetchScoreboard(sport, league string) (*ScoreboardResponse, error) {
	key := sport + "/" + league

	c.mu.RLock()
	entry, ok := c.cache[key]
	c.mu.RUnlock()

	if ok && time.Since(entry.fetchedAt) < c.cacheTTL {
		return entry.data, nil
	}

	url := fmt.Sprintf("%s/%s/%s/scoreboard", baseURL, sport, league)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		// Return stale cache on error
		if ok {
			return entry.data, nil
		}
		return nil, fmt.Errorf("espn: fetching scoreboard for %s: %w", key, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if ok {
			return entry.data, nil
		}
		return nil, fmt.Errorf("espn: scoreboard %s returned status %d", key, resp.StatusCode)
	}

	var sb ScoreboardResponse
	if err := json.NewDecoder(resp.Body).Decode(&sb); err != nil {
		if ok {
			return entry.data, nil
		}
		return nil, fmt.Errorf("espn: decoding scoreboard for %s: %w", key, err)
	}

	c.mu.Lock()
	c.cache[key] = cacheEntry{data: &sb, fetchedAt: time.Now()}
	c.mu.Unlock()

	return &sb, nil
}

// FetchScoreboardWithDates gets the scoreboard for a date range (format: YYYYMMDD-YYYYMMDD or YYYYMMDD).
func (c *Client) FetchScoreboardWithDates(sport, league, dates string) (*ScoreboardResponse, error) {
	key := sport + "/" + league + "?dates=" + dates

	c.mu.RLock()
	entry, ok := c.cache[key]
	c.mu.RUnlock()

	if ok && time.Since(entry.fetchedAt) < c.cacheTTL {
		return entry.data, nil
	}

	url := fmt.Sprintf("%s/%s/%s/scoreboard?dates=%s", baseURL, sport, league, dates)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		if ok {
			return entry.data, nil
		}
		return nil, fmt.Errorf("espn: fetching scoreboard for %s: %w", key, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if ok {
			return entry.data, nil
		}
		return nil, fmt.Errorf("espn: scoreboard %s returned status %d", key, resp.StatusCode)
	}

	var sb ScoreboardResponse
	if err := json.NewDecoder(resp.Body).Decode(&sb); err != nil {
		if ok {
			return entry.data, nil
		}
		return nil, fmt.Errorf("espn: decoding scoreboard for %s: %w", key, err)
	}

	c.mu.Lock()
	c.cache[key] = cacheEntry{data: &sb, fetchedAt: time.Now()}
	c.mu.Unlock()

	return &sb, nil
}

// GetTeamGames extracts games for specific teams from a scoreboard response.
// teamAbbrs should be uppercase abbreviations (e.g. "PHI", "DAL").
func GetTeamGames(sb *ScoreboardResponse, teamAbbrs []string) []TeamGame {
	abbrSet := make(map[string]bool, len(teamAbbrs))
	for _, a := range teamAbbrs {
		abbrSet[strings.ToUpper(strings.TrimSpace(a))] = true
	}

	var games []TeamGame
	for _, event := range sb.Events {
		if len(event.Competitions) == 0 {
			continue
		}
		comp := event.Competitions[0]
		if len(comp.Competitors) < 2 {
			continue
		}

		eventDate, _ := time.Parse(time.RFC3339, event.Date)

		for _, competitor := range comp.Competitors {
			abbr := strings.ToUpper(competitor.Team.Abbreviation)
			if !abbrSet[abbr] {
				continue
			}

			// Find opponent
			var opponent Competitor
			for _, other := range comp.Competitors {
				if other.Team.Abbreviation != competitor.Team.Abbreviation {
					opponent = other
					break
				}
			}

			g := TeamGame{
				EventID:       event.ID,
				Date:          eventDate,
				EventName:     event.Name,
				TeamAbbr:      abbr,
				TeamName:      competitor.Team.ShortDisplayName,
				OpponentAbbr:  strings.ToUpper(opponent.Team.Abbreviation),
				OpponentName:  opponent.Team.ShortDisplayName,
				IsHome:        competitor.HomeAway == "home",
				Status:        event.Status.Type.State,
				TeamScore:     competitor.Score,
				OpponentScore: opponent.Score,
				StatusDesc:    event.Status.Type.Description,
				Venue:         comp.Venue.FullName,
				IsPlayoff:     event.Season.Type.ID == "3" || event.Season.Type.Type == 3,
			}

			if event.Status.Type.State == "post" {
				won := competitor.Winner
				g.Won = &won
			}

			games = append(games, g)
		}
	}
	return games
}

// championshipKeywords are event name substrings that indicate a major championship game.
var championshipKeywords = []string{
	"super bowl", "world series", "stanley cup", "nba finals",
	"mls cup", "championship", "ncaa final", "final four",
	"college football playoff", "cfp",
}

// GetPlayoffGames extracts all postseason games from a scoreboard response,
// regardless of team. Only returns games whose event names match championship keywords
// or that are marked as postseason (season type 3).
func GetPlayoffGames(sb *ScoreboardResponse) []TeamGame {
	var games []TeamGame
	for _, event := range sb.Events {
		isPostseason := event.Season.Type.ID == "3" || event.Season.Type.Type == 3
		if !isPostseason {
			continue
		}

		// Check if this is a major championship-level event
		nameLower := strings.ToLower(event.Name)
		isMajor := false
		for _, kw := range championshipKeywords {
			if strings.Contains(nameLower, kw) {
				isMajor = true
				break
			}
		}
		if !isMajor {
			continue
		}

		if len(event.Competitions) == 0 {
			continue
		}
		comp := event.Competitions[0]
		if len(comp.Competitors) < 2 {
			continue
		}

		eventDate, _ := time.Parse(time.RFC3339, event.Date)
		home := comp.Competitors[0]
		away := comp.Competitors[1]
		if home.HomeAway != "home" {
			home, away = away, home
		}

		g := TeamGame{
			EventID:       event.ID,
			Date:          eventDate,
			EventName:     event.Name,
			TeamAbbr:      strings.ToUpper(home.Team.Abbreviation),
			TeamName:      home.Team.ShortDisplayName,
			OpponentAbbr:  strings.ToUpper(away.Team.Abbreviation),
			OpponentName:  away.Team.ShortDisplayName,
			IsHome:        true,
			Status:        event.Status.Type.State,
			TeamScore:     home.Score,
			OpponentScore: away.Score,
			StatusDesc:    event.Status.Type.Description,
			Venue:         comp.Venue.FullName,
			IsPlayoff:     true,
		}
		if event.Status.Type.State == "post" {
			won := home.Winner
			g.Won = &won
		}
		games = append(games, g)
	}
	return games
}

// FetchTeamSchedule retrieves the schedule for a specific team.
// teamAbbr is the uppercase ESPN abbreviation (e.g. "PHI").
func (c *Client) FetchTeamSchedule(sport, league, teamAbbr string) (*ScheduleResponse, error) {
	key := sport + "/" + league + "/teams/" + strings.ToLower(teamAbbr) + "/schedule"

	c.mu.RLock()
	entry, ok := c.scheduleCache[key]
	c.mu.RUnlock()

	if ok && time.Since(entry.fetchedAt) < c.cacheTTL {
		return entry.data, nil
	}

	url := fmt.Sprintf("%s/%s/%s/teams/%s/schedule", baseURL, sport, league, strings.ToLower(teamAbbr))
	resp, err := c.httpClient.Get(url)
	if err != nil {
		if ok {
			return entry.data, nil
		}
		return nil, fmt.Errorf("espn: fetching schedule for %s: %w", key, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if ok {
			return entry.data, nil
		}
		return nil, fmt.Errorf("espn: schedule %s returned status %d", key, resp.StatusCode)
	}

	var sr ScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		if ok {
			return entry.data, nil
		}
		return nil, fmt.Errorf("espn: decoding schedule for %s: %w", key, err)
	}

	c.mu.Lock()
	c.scheduleCache[key] = scheduleCacheEntry{data: &sr, fetchedAt: time.Now()}
	c.mu.Unlock()

	return &sr, nil
}

// GetUpcomingFromSchedule extracts the next upcoming game for a team from a schedule response.
// It finds games where boxscoreAvailable is false and the date is in the future.
func GetUpcomingFromSchedule(sr *ScheduleResponse, teamAbbr string) *TeamGame {
	now := time.Now()
	teamAbbr = strings.ToUpper(teamAbbr)

	var best *TeamGame
	for _, event := range sr.Events {
		if len(event.Competitions) == 0 {
			continue
		}
		comp := event.Competitions[0]

		// Skip completed games
		if comp.BoxscoreAvailable {
			continue
		}

		eventDate, err := time.Parse(time.RFC3339, event.Date)
		if err != nil {
			// Try alternate format without timezone offset
			eventDate, err = time.Parse("2006-01-02T15:04Z", event.Date)
			if err != nil {
				continue
			}
		}

		// Must be in the future
		if eventDate.Before(now) {
			continue
		}

		if len(comp.Competitors) < 2 {
			continue
		}

		// Find which competitor is our team and which is the opponent
		var team, opponent Competitor
		found := false
		for _, c := range comp.Competitors {
			if strings.ToUpper(c.Team.Abbreviation) == teamAbbr {
				team = c
				found = true
			} else {
				opponent = c
			}
		}
		if !found {
			continue
		}

		g := TeamGame{
			EventID:      event.ID,
			Date:         eventDate,
			EventName:    event.Name,
			TeamAbbr:     teamAbbr,
			TeamName:     team.Team.ShortDisplayName,
			OpponentAbbr: strings.ToUpper(opponent.Team.Abbreviation),
			OpponentName: opponent.Team.ShortDisplayName,
			IsHome:       team.HomeAway == "home",
			Status:       "pre",
			StatusDesc:   "Scheduled",
			Venue:        comp.Venue.FullName,
		}

		if best == nil || g.Date.Before(best.Date) {
			best = &g
		}
	}

	return best
}
