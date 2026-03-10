package espn

import (
	"encoding/json"
	"fmt"
	"time"
)

// ScoreboardResponse is the top-level ESPN scoreboard API response.
type ScoreboardResponse struct {
	Events []Event `json:"events"`
}

// Event represents a single game/event from ESPN.
type Event struct {
	ID           string        `json:"id"`
	Date         string        `json:"date"` // ISO 8601
	Name         string        `json:"name"`
	ShortName    string        `json:"shortName"`
	Competitions []Competition `json:"competitions"`
	Status       EventStatus   `json:"status"`
	Season       Season        `json:"season"`
}

// Season identifies what part of the season this event belongs to.
type Season struct {
	Year int        `json:"year"`
	Type SeasonType `json:"type"`
}

// SeasonType indicates pre/regular/post season.
// ESPN returns this as either a bare int (e.g. 3) or an object ({"id":"3","type":3,"name":"Postseason"}).
type SeasonType struct {
	ID   string
	Type int
	Name string
}

// UnmarshalJSON handles both integer and object forms of season type.
func (st *SeasonType) UnmarshalJSON(data []byte) error {
	// Try bare number first (e.g. 2 or 3)
	var num int
	if err := json.Unmarshal(data, &num); err == nil {
		st.Type = num
		st.ID = fmt.Sprintf("%d", num)
		return nil
	}

	// Try object form
	type alias SeasonType
	var obj alias
	if err := json.Unmarshal(data, &obj); err == nil {
		*st = SeasonType(obj)
		return nil
	}

	// Try string form (unlikely but defensive)
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		st.ID = s
		return nil
	}

	return fmt.Errorf("espn: cannot unmarshal season type: %s", string(data))
}

// Competition holds the two competitors and venue info.
type Competition struct {
	Type        CompetitionType `json:"type"`
	Competitors []Competitor     `json:"competitors"`
	Venue       Venue            `json:"venue"`
	Date        string           `json:"date"`
	StartDate   string           `json:"startDate"`
}

// CompetitionType identifies the kind of competition (e.g. Race, Qual, FP1).
type CompetitionType struct {
	ID           string `json:"id"`
	Abbreviation string `json:"abbreviation"`
}

// Competitor is one side of a game (home or away), or a driver entry in racing.
type Competitor struct {
	ID       string        `json:"id"`
	Team     Team          `json:"team"`
	Athlete  Athlete       `json:"athlete"`
	Score    FlexScore     `json:"score"`
	HomeAway string        `json:"homeAway"` // "home" or "away"
	Winner   bool          `json:"winner"`
	Records  []RecordEntry `json:"records"`
	Order    int           `json:"order"` // finishing position in racing
}

// FlexScore handles score fields that can be either a string ("24") or an
// object ({"value":24.0,"displayValue":"24"}) depending on the ESPN endpoint.
type FlexScore string

func (fs *FlexScore) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*fs = FlexScore(s)
		return nil
	}
	// Try object with displayValue
	var obj struct {
		DisplayValue string `json:"displayValue"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		*fs = FlexScore(obj.DisplayValue)
		return nil
	}
	// Fallback: empty
	*fs = ""
	return nil
}

// Athlete identifies a driver or individual competitor (used in racing).
type Athlete struct {
	DisplayName      string `json:"displayName"`
	ShortDisplayName string `json:"shortDisplayName"`
}

// Team contains team identity info.
type Team struct {
	ID               string `json:"id"`
	Abbreviation     string `json:"abbreviation"`
	DisplayName      string `json:"displayName"`
	ShortDisplayName string `json:"shortDisplayName"`
	Logo             string `json:"logo"`
}

// RecordEntry is a team's record (e.g. "10-5").
type RecordEntry struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

// Venue is where the game is played.
type Venue struct {
	FullName string `json:"fullName"`
	City     string `json:"city"`
	State    string `json:"state"`
}

// EventStatus describes whether a game is scheduled, in-progress, or final.
type EventStatus struct {
	Clock        float64    `json:"clock"`
	DisplayClock string     `json:"displayClock"`
	Period       int        `json:"period"`
	Type         StatusType `json:"type"`
}

// StatusType classifies the game state.
type StatusType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`  // "STATUS_SCHEDULED", "STATUS_FINAL", "STATUS_IN_PROGRESS"
	State       string `json:"state"` // "pre", "post", "in"
	Completed   bool   `json:"completed"`
	Description string `json:"description"` // "Scheduled", "Final", etc.
}

// TeamGame is our normalized representation of a game for a specific team.
type TeamGame struct {
	EventID       string
	Date          time.Time
	EventName     string // Full event name, e.g. "Super Bowl LX"
	TeamAbbr      string
	TeamName      string
	OpponentAbbr  string
	OpponentName  string
	IsHome        bool
	Status        string // "pre", "post", "in"
	TeamScore     string
	OpponentScore string
	Won           *bool  // nil if not yet played
	StatusDesc    string // "Final", "Scheduled", etc.
	Venue         string
	IsPlayoff     bool // true if postseason/playoff game
}

// ScheduleResponse is the top-level ESPN team schedule API response.
type ScheduleResponse struct {
	Team   Team            `json:"team"`
	Events []ScheduleEvent `json:"events"`
}

// ScheduleEvent is an event from the team schedule endpoint.
// Unlike scoreboard events, these lack a status.type.state field.
type ScheduleEvent struct {
	ID           string              `json:"id"`
	Date         string              `json:"date"`
	Name         string              `json:"name"`
	ShortName    string              `json:"shortName"`
	Competitions []ScheduleCompetition `json:"competitions"`
}

// ScheduleCompetition holds competitor and venue info from the schedule endpoint.
type ScheduleCompetition struct {
	Competitors       []Competitor `json:"competitors"`
	Venue             Venue        `json:"venue"`
	BoxscoreAvailable bool         `json:"boxscoreAvailable"`
}

// RaceEvent is the normalized representation of a racing event (race weekend).
type RaceEvent struct {
	EventID    string
	Date       time.Time
	RaceName   string
	Circuit    string
	Status     string // "pre", "post", "in"
	Winner     string // Driver name for completed races
	StatusDesc string
}
