package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	"trmnl-sports/config"
	"trmnl-sports/espn"
	"trmnl-sports/trmnl"
)

func main() {
	config.LoadEnvFile(".env")

	activeSports := config.LoadActiveSports()
	if len(activeSports) == 0 {
		log.Println("WARNING: No sports configured. Set at least one team env var (e.g. NFL_TEAMS=PHI,DAL).")
	} else {
		for _, as := range activeSports {
			log.Printf("Loaded %s: %v", as.Name, as.Teams)
		}
	}

	espnClient := espn.NewClient(15 * time.Minute)
	tz := config.LoadTimezone()
	log.Printf("Timezone: %s", tz)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// CORS headers so the TRMNL web-based preview can fetch this endpoint
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// TRMNL sends POST to the markup URL
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sections := fetchAllSections(espnClient, activeSports)
		markup := trmnl.RenderMarkup(sections, tz)

		w.Header().Set("Content-Type", "application/json")
		// Use SetEscapeHTML(false) so <, >, & are NOT escaped to \u003c etc.
		// TRMNL expects literal HTML angle brackets in the JSON markup strings.
		enc := json.NewEncoder(w)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(markup); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	addr := config.ListenAddr()
	log.Printf("TRMNL Sports plugin listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// fetchAllSections queries ESPN for each active sport and builds display sections.
// Teams with no games within the window are hidden. Championship-level playoff
// games are always included regardless of team selection.
func fetchAllSections(client *espn.Client, sports []config.ActiveSport) []trmnl.SportSection {
	// 3-week window: 1 week back (recent results) + 3 weeks ahead
	now := time.Now()
	startDate := now.AddDate(0, 0, -7).Format("20060102")
	endDate := now.AddDate(0, 0, 21).Format("20060102")
	dateRange := startDate + "-" + endDate
	threeWeeksOut := now.AddDate(0, 0, 21)

	var sections []trmnl.SportSection

	for _, as := range sports {
		sb, err := client.FetchScoreboardWithDates(as.Sport, as.League, dateRange)
		if err != nil {
			log.Printf("Error fetching %s scoreboard: %v", as.Name, err)
			continue
		}

		log.Printf("%s: ESPN returned %d events", as.Name, len(sb.Events))

		// Racing series — extract race events, no team filtering
		if as.IsRacing {
			lastRace, nextRace := espn.GetRaceEvents(sb)
			if lastRace != nil || nextRace != nil {
				sections = append(sections, trmnl.SportSection{
					SportName: as.Name,
					SportKey:  as.League,
					RaceData: &trmnl.RaceDisplay{
						LastRace: lastRace,
						NextRace: nextRace,
					},
				})
			}
			continue
		}

		// Get games for the user's selected teams
		allGames := espn.GetTeamGames(sb, as.Teams)
		log.Printf("%s: %d games matched configured teams %v", as.Name, len(allGames), as.Teams)

		// Group games by team
		byTeam := make(map[string][]espn.TeamGame)
		for _, g := range allGames {
			byTeam[g.TeamAbbr] = append(byTeam[g.TeamAbbr], g)
		}

		var teamDisplays []trmnl.TeamDisplay
		for _, teamAbbr := range as.Teams {
			games := byTeam[teamAbbr]
			td := trmnl.BuildTeamDisplay(teamAbbr, games)

			// If scoreboard didn't have an upcoming game (ESPN caps at 100 events),
			// fall back to the team schedule endpoint.
			if td.NextGame == nil {
				sched, err := client.FetchTeamSchedule(as.Sport, as.League, teamAbbr)
				if err != nil {
					log.Printf("%s/%s: schedule fetch failed: %v", as.Name, teamAbbr, err)
				} else {
					upcoming := espn.GetUpcomingFromSchedule(sched, teamAbbr)
					if upcoming != nil {
						td.NextGame = upcoming
						if td.TeamName == "" {
							td.TeamName = upcoming.TeamName
						}
						log.Printf("%s/%s: found upcoming game from schedule: %s on %s",
							as.Name, teamAbbr, upcoming.OpponentAbbr, upcoming.Date.Format("2006-01-02"))
					}
				}
			}

			// Hide teams with no games in the 3-week window
			if td.LastGame == nil && td.NextGame == nil {
				continue
			}
			// If next game is beyond 3 weeks and no recent result, skip
			if td.LastGame == nil && td.NextGame != nil && td.NextGame.Date.After(threeWeeksOut) {
				continue
			}
			teamDisplays = append(teamDisplays, td)
		}

		// Get championship/playoff games for this league (regardless of team)
		playoffGames := espn.GetPlayoffGames(sb)
		if len(playoffGames) > 0 {
			// Avoid duplicating games already shown for a user's team
			shown := make(map[string]bool)
			for _, td := range teamDisplays {
				if td.LastGame != nil {
					shown[td.LastGame.EventID] = true
				}
				if td.NextGame != nil {
					shown[td.NextGame.EventID] = true
				}
			}
			for _, pg := range playoffGames {
				if shown[pg.EventID] {
					continue
				}
				shown[pg.EventID] = true
				td := trmnl.TeamDisplay{
					TeamAbbr: pg.EventName,
					TeamName: pg.EventName,
				}
				if pg.Status == "post" {
					td.LastGame = &pg
				} else {
					td.NextGame = &pg
				}
				teamDisplays = append(teamDisplays, td)
			}
		}

		if len(teamDisplays) > 0 {
			sections = append(sections, trmnl.SportSection{
				SportName: as.Name,
				SportKey:  as.League,
				TeamData:  teamDisplays,
			})
		}
	}

	return sections
}
