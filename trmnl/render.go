package trmnl

import (
	"fmt"
	"html"
	"strings"
	"time"
	"trmnl-sports/espn"
)

// SportSection groups a sport's name with its team game data.
type SportSection struct {
	SportName string
	SportKey  string // ESPN league key, used for icon lookup
	TeamData  []TeamDisplay
}

// TeamDisplay holds the last result and next game for one team.
type TeamDisplay struct {
	TeamAbbr string
	TeamName string
	LastGame *espn.TeamGame
	NextGame *espn.TeamGame
}

// MarkupResponse is the JSON payload TRMNL expects from our endpoint.
type MarkupResponse struct {
	Markup          string            `json:"markup"`
	MarkupHalfHoriz string            `json:"markup_half_horizontal"`
	MarkupHalfVert  string            `json:"markup_half_vertical"`
	MarkupQuadrant  string            `json:"markup_quadrant"`
	MergeVariables  map[string]string `json:"merge_variables"`
}

// BuildTeamDisplay creates a TeamDisplay from a slice of games for one team.
// It picks the most recent completed game and the nearest future game.
func BuildTeamDisplay(teamAbbr string, games []espn.TeamGame) TeamDisplay {
	td := TeamDisplay{TeamAbbr: teamAbbr}

	var lastCompleted *espn.TeamGame
	var nextScheduled *espn.TeamGame
	now := time.Now()

	for i := range games {
		g := &games[i]
		if td.TeamName == "" {
			td.TeamName = g.TeamName
		}

		if g.Status == "post" {
			if lastCompleted == nil || g.Date.After(lastCompleted.Date) {
				lastCompleted = g
			}
		} else if g.Status == "pre" {
			if g.Date.After(now) || g.Date.Equal(now) {
				if nextScheduled == nil || g.Date.Before(nextScheduled.Date) {
					nextScheduled = g
				}
			}
		}
	}

	td.LastGame = lastCompleted
	td.NextGame = nextScheduled
	return td
}

// RenderMarkup generates all four TRMNL layout markups from a list of sport sections.
func RenderMarkup(sections []SportSection) MarkupResponse {
	resp := MarkupResponse{
		Markup:          renderFull(sections),
		MarkupHalfHoriz: renderHalfHorizontal(sections),
		MarkupHalfVert:  renderHalfVertical(sections),
		MarkupQuadrant:  renderQuadrant(sections),
	}
	// Also populate merge_variables so the endpoint works with TRMNL
	// Private Plugin "Polling" strategy, where TRMNL merges variables
	// into a user-defined template.
	resp.MergeVariables = map[string]string{
		"markup":                resp.Markup,
		"markup_half_horizontal": resp.MarkupHalfHoriz,
		"markup_half_vertical":   resp.MarkupHalfVert,
		"markup_quadrant":        resp.MarkupQuadrant,
	}
	return resp
}

func renderFull(sections []SportSection) string {
	if len(sections) == 0 {
		return noDataMarkup("full")
	}

	// Collect scores and upcoming into separate lists
	type row struct {
		icon     string
		teamAbbr string
		game     *espn.TeamGame
	}
	var scores, upcoming []row
	for _, sec := range sections {
		icon := SportIcon(sec.SportKey)
		for _, td := range sec.TeamData {
			if td.LastGame != nil {
				scores = append(scores, row{icon, td.TeamAbbr, td.LastGame})
			}
			if td.NextGame != nil {
				upcoming = append(upcoming, row{icon, td.TeamAbbr, td.NextGame})
			}
		}
	}

	// Build Recent Scores table
	var scoresHTML strings.Builder
	if len(scores) > 0 {
		scoresHTML.WriteString(`    <span class="title title--small" style="margin-bottom:4px">Recent Scores</span>
    <table class="table table--small" data-table-limit="true">
      <thead>
        <tr>
          <th></th>
          <th><span class="title title--small">Team</span></th>
          <th><span class="title title--small">Opponent</span></th>
          <th><span class="title title--small">Score</span></th>
          <th><span class="title title--small">Date</span></th>
        </tr>
      </thead>
      <tbody>
`)
		for _, r := range scores {
			prefix := "vs"
			if !r.game.IsHome {
				prefix = "@"
			}
			result := ""
			if r.game.Won != nil {
				if *r.game.Won {
					result = "W "
				} else {
					result = "L "
				}
			}
			scoresHTML.WriteString(fmt.Sprintf(`        <tr>
          <td>%s</td>
          <td><span class="label">%s</span></td>
          <td><span class="label label--small">%s %s</span></td>
          <td><span class="label label--small">%s%s-%s</span></td>
          <td><span class="label label--small">%s</span></td>
        </tr>
`, r.icon, esc(r.teamAbbr), prefix, esc(r.game.OpponentAbbr),
				result, esc(r.game.TeamScore), esc(r.game.OpponentScore),
				r.game.Date.Format("Mon 1/2")))
		}
		scoresHTML.WriteString(`      </tbody>
    </table>`)
	} else {
		scoresHTML.WriteString(`    <span class="title title--small">Recent Scores</span>
    <span class="description">No recent results</span>`)
	}

	// Build Upcoming Games table
	var upcomingHTML strings.Builder
	if len(upcoming) > 0 {
		upcomingHTML.WriteString(`    <span class="title title--small" style="margin-bottom:4px">Upcoming Games</span>
    <table class="table table--small" data-table-limit="true">
      <thead>
        <tr>
          <th></th>
          <th><span class="title title--small">Team</span></th>
          <th><span class="title title--small">Opponent</span></th>
          <th><span class="title title--small">Date</span></th>
        </tr>
      </thead>
      <tbody>
`)
		for _, r := range upcoming {
			prefix := "vs"
			if !r.game.IsHome {
				prefix = "@"
			}
			upcomingHTML.WriteString(fmt.Sprintf(`        <tr>
          <td>%s</td>
          <td><span class="label">%s</span></td>
          <td><span class="label label--small">%s %s</span></td>
          <td><span class="label label--small">%s</span></td>
        </tr>
`, r.icon, esc(r.teamAbbr), prefix, esc(r.game.OpponentAbbr),
				r.game.Date.Format("Mon 1/2 3:04PM")))
		}
		upcomingHTML.WriteString(`      </tbody>
    </table>`)
	} else {
		upcomingHTML.WriteString(`    <span class="title title--small">Upcoming Games</span>
    <span class="description">No upcoming games</span>`)
	}

	return fmt.Sprintf(`<div class="view view--full">
  <div class="layout layout--col">
    <div class="columns">
      <div class="column">
%s
      </div>
      <div class="column">
%s
      </div>
    </div>
  </div>
  <div class="title_bar">
    <span class="title">Sports Scores</span>
    <span class="instance">TRMNL Sports</span>
  </div>
</div>`, scoresHTML.String(), upcomingHTML.String())
}

func renderHalfHorizontal(sections []SportSection) string {
	if len(sections) == 0 {
		return noDataMarkup("half_horizontal")
	}

	type row struct {
		icon     string
		teamAbbr string
		game     *espn.TeamGame
	}
	var scores, upcoming []row
	for _, sec := range sections {
		icon := SportIcon(sec.SportKey)
		for _, td := range sec.TeamData {
			if td.LastGame != nil {
				scores = append(scores, row{icon, td.TeamAbbr, td.LastGame})
			}
			if td.NextGame != nil {
				upcoming = append(upcoming, row{icon, td.TeamAbbr, td.NextGame})
			}
		}
	}

	var scoresHTML strings.Builder
	if len(scores) > 0 {
		scoresHTML.WriteString(`    <span class="title title--small">Scores</span>
    <table class="table table--xsmall" data-table-limit="true">
      <tbody>
`)
		for _, r := range scores {
			prefix := "vs"
			if !r.game.IsHome {
				prefix = "@"
			}
			result := ""
			if r.game.Won != nil {
				if *r.game.Won {
					result = "W "
				} else {
					result = "L "
				}
			}
			scoresHTML.WriteString(fmt.Sprintf(`        <tr>
          <td>%s</td>
          <td><span class="label">%s</span></td>
          <td><span class="label label--small">%s %s</span></td>
          <td><span class="label label--small">%s%s-%s</span></td>
        </tr>
`, r.icon, esc(r.teamAbbr), prefix, esc(r.game.OpponentAbbr),
				result, esc(r.game.TeamScore), esc(r.game.OpponentScore)))
		}
		scoresHTML.WriteString(`      </tbody>
    </table>`)
	}

	var upcomingHTML strings.Builder
	if len(upcoming) > 0 {
		upcomingHTML.WriteString(`    <span class="title title--small">Upcoming</span>
    <table class="table table--xsmall" data-table-limit="true">
      <tbody>
`)
		for _, r := range upcoming {
			prefix := "vs"
			if !r.game.IsHome {
				prefix = "@"
			}
			upcomingHTML.WriteString(fmt.Sprintf(`        <tr>
          <td>%s</td>
          <td><span class="label">%s</span></td>
          <td><span class="label label--small">%s %s</span></td>
          <td><span class="label label--small">%s</span></td>
        </tr>
`, r.icon, esc(r.teamAbbr), prefix, esc(r.game.OpponentAbbr),
				r.game.Date.Format("Mon 1/2 3:04PM")))
		}
		upcomingHTML.WriteString(`      </tbody>
    </table>`)
	}

	return fmt.Sprintf(`<div class="view view--half_horizontal">
  <div class="layout layout--col">
    <div class="columns">
      <div class="column">
%s
      </div>
      <div class="column">
%s
      </div>
    </div>
  </div>
  <div class="title_bar">
    <span class="title">Sports</span>
    <span class="instance">TRMNL Sports</span>
  </div>
</div>`, scoresHTML.String(), upcomingHTML.String())
}

func renderHalfVertical(sections []SportSection) string {
	if len(sections) == 0 {
		return noDataMarkup("half_vertical")
	}

	type entry struct {
		icon     string
		sport    string
		teamAbbr string
		game     *espn.TeamGame
	}
	var scores, upcoming []entry
	for _, sec := range sections {
		icon := SportIcon(sec.SportKey)
		for _, td := range sec.TeamData {
			if td.LastGame != nil {
				scores = append(scores, entry{icon, sec.SportName, td.TeamAbbr, td.LastGame})
			}
			if td.NextGame != nil {
				upcoming = append(upcoming, entry{icon, sec.SportName, td.TeamAbbr, td.NextGame})
			}
		}
	}

	var items strings.Builder
	if len(scores) > 0 {
		items.WriteString(`        <span class="title title--small" style="margin-bottom:2px">Recent Scores</span>
`)
		for _, e := range scores {
			prefix := "vs"
			if !e.game.IsHome {
				prefix = "@"
			}
			result := ""
			if e.game.Won != nil {
				if *e.game.Won {
					result = "W "
				} else {
					result = "L "
				}
			}
			items.WriteString(fmt.Sprintf(`        <div class="item">
          <div class="meta">%s</div>
          <div class="content">
            <span class="title title--small">%s %s</span>
            <span class="description">%s%s-%s %s %s</span>
          </div>
        </div>
`, e.icon, esc(e.sport), esc(e.teamAbbr),
				result, esc(e.game.TeamScore), esc(e.game.OpponentScore),
				prefix, esc(e.game.OpponentAbbr)))
		}
	}

	if len(upcoming) > 0 {
		items.WriteString(`        <span class="title title--small" style="margin-top:4px;margin-bottom:2px">Upcoming</span>
`)
		for _, e := range upcoming {
			prefix := "vs"
			if !e.game.IsHome {
				prefix = "@"
			}
			items.WriteString(fmt.Sprintf(`        <div class="item">
          <div class="meta">%s</div>
          <div class="content">
            <span class="title title--small">%s %s</span>
            <span class="label label--small">%s %s — %s</span>
          </div>
        </div>
`, e.icon, esc(e.sport), esc(e.teamAbbr),
				prefix, esc(e.game.OpponentAbbr),
				e.game.Date.Format("Mon 1/2 3:04PM")))
		}
	}

	return fmt.Sprintf(`<div class="view view--half_vertical">
  <div class="layout layout--col">
    <div class="columns">
      <div class="column">
%s
      </div>
    </div>
  </div>
  <div class="title_bar">
    <span class="title">Sports</span>
    <span class="instance">TRMNL Sports</span>
  </div>
</div>`, items.String())
}

func renderQuadrant(sections []SportSection) string {
	if len(sections) == 0 {
		return noDataMarkup("quadrant")
	}

	type entry struct {
		icon     string
		sport    string
		teamAbbr string
		detail   string
		isScore  bool
	}
	var items []entry
	for _, sec := range sections {
		icon := SportIcon(sec.SportKey)
		for _, td := range sec.TeamData {
			if td.LastGame != nil && td.LastGame.Won != nil {
				r := "L"
				if *td.LastGame.Won {
					r = "W"
				}
				prefix := "vs"
				if !td.LastGame.IsHome {
					prefix = "@"
				}
				items = append(items, entry{icon, sec.SportName, td.TeamAbbr,
					fmt.Sprintf("%s %s-%s %s %s", r,
						esc(td.LastGame.TeamScore), esc(td.LastGame.OpponentScore),
						prefix, esc(td.LastGame.OpponentAbbr)), true})
			}
			if td.NextGame != nil {
				prefix := "vs"
				if !td.NextGame.IsHome {
					prefix = "@"
				}
				items = append(items, entry{icon, sec.SportName, td.TeamAbbr,
					fmt.Sprintf("%s %s %s", prefix, esc(td.NextGame.OpponentAbbr),
						td.NextGame.Date.Format("1/2 3:04PM")), false})
			}
		}
	}

	// Show up to 4 items, prioritize scores
	var out strings.Builder
	count := 0
	for _, e := range items {
		if count >= 4 {
			break
		}
		if e.isScore {
			out.WriteString(fmt.Sprintf(`    <div class="item">
      <div class="meta">%s</div>
      <div class="content">
        <span class="title title--small">%s %s</span>
        <span class="label label--small">%s</span>
      </div>
    </div>
`, e.icon, esc(e.sport), esc(e.teamAbbr), e.detail))
			count++
		}
	}
	for _, e := range items {
		if count >= 4 {
			break
		}
		if !e.isScore {
			out.WriteString(fmt.Sprintf(`    <div class="item">
      <div class="meta">%s</div>
      <div class="content">
        <span class="title title--small">%s %s</span>
        <span class="label label--small">%s</span>
      </div>
    </div>
`, e.icon, esc(e.sport), esc(e.teamAbbr), e.detail))
			count++
		}
	}

	return fmt.Sprintf(`<div class="view view--quadrant">
  <div class="layout layout--col">
%s
  </div>
  <div class="title_bar">
    <span class="title">Sports</span>
  </div>
</div>`, out.String())
}

func noDataMarkup(viewClass string) string {
	return fmt.Sprintf(`<div class="view view--%s">
  <div class="layout layout--col layout--center">
    <span class="title">No Sports Data</span>
    <span class="description">Configure team environment variables to see scores.</span>
  </div>
</div>`, viewClass)
}

func esc(s string) string {
	return html.EscapeString(s)
}
