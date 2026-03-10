# TRMNL Sports Plugin — Backend Endpoint

## Overview

A **Go** application that serves as a TRMNL plugin backend endpoint. It pulls sports data from the public ESPN API and returns TRMNL-compatible markup showing the **last completed game result** and **next upcoming game** for each configured team.

---

## Data Source

- **ESPN API**: Structure and endpoints documented at <https://github.com/pseudo-r/Public-ESPN-API>
- Key endpoint: `https://site.api.espn.com/apis/site/v2/sports/{sport}/{league}/scoreboard`
- Team schedule: `https://site.api.espn.com/apis/site/v2/sports/{sport}/{league}/teams/{id}/schedule`
- No authentication required for ESPN API.
- Only completed and scheduled games are displayed (no live/in-progress scores).

## TRMNL Screen Generation

- Markup and screen generation flow follows <https://docs.trmnl.com/go/plugin-marketplace/plugin-screen-generation-flow>
- TRMNL framework CSS/JS: `https://trmnl.com/css/latest/plugins.css` and `https://trmnl.com/js/latest/plugins.js`
- Framework docs: <https://trmnl.com/framework/docs>
- The plugin must support **all TRMNL screen sizes**: full (800×480), half-horizontal (800×240), half-vertical (400×480), quadrant.
- TRMNL POSTs to the markup URL; response must be JSON with keys: `markup`, `markup_half_horizontal`, `markup_half_vertical`, `markup_quadrant`.

---

## Supported Sports

The application must support **all** of the following ESPN sport/league slugs:

| Sport                  | ESPN Slug          | Env Var            |
|------------------------|--------------------|--------------------|
| NFL                    | `football/nfl`     | `NFL_TEAMS`        |
| NBA                    | `basketball/nba`   | `NBA_TEAMS`        |
| MLB                    | `baseball/mlb`     | `MLB_TEAMS`        |
| NHL                    | `hockey/nhl`       | `NHL_TEAMS`        |
| College Football       | `football/college-football` | `NCAAF_TEAMS` |
| College Basketball     | `basketball/mens-college-basketball` | `NCAAM_TEAMS` |
| MLS                    | `soccer/usa.1`     | `MLS_TEAMS`        |
| Premier League         | `soccer/eng.1`     | `EPL_TEAMS`        |
| UFC / MMA             | `mma/ufc`          | `UFC_FIGHTERS`     |

---

## Environment Variables

| Variable        | Required | Description                                                       | Example                |
|-----------------|----------|-------------------------------------------------------------------|------------------------|
| `LISTEN_ADDR`   | No       | Host:port the server listens on (default `0.0.0.0:8080`)         | `0.0.0.0:8080`        |
| `NFL_TEAMS`     | No       | Comma-separated ESPN team abbreviations for NFL                   | `PHI,DAL,KC`          |
| `NBA_TEAMS`     | No       | Comma-separated ESPN team abbreviations for NBA                   | `LAL,BOS`             |
| `MLB_TEAMS`     | No       | Comma-separated ESPN team abbreviations for MLB                   | `NYY,LAD`             |
| `NHL_TEAMS`     | No       | Comma-separated ESPN team abbreviations for NHL                   | `NYR,BOS`             |
| `NCAAF_TEAMS`   | No       | Comma-separated ESPN team IDs/abbreviations for College Football  | `MICH,OSU`            |
| `NCAAM_TEAMS`   | No       | Comma-separated ESPN team IDs/abbreviations for College Basketball| `DUKE,UNC`            |
| `MLS_TEAMS`     | No       | Comma-separated ESPN team abbreviations for MLS                   | `ATL,LAFC`            |
| `EPL_TEAMS`     | No       | Comma-separated ESPN team abbreviations for Premier League        | `LIV,ARS`             |
| `UFC_FIGHTERS`  | No       | Comma-separated ESPN fighter IDs or names for UFC                 | `(TBD per ESPN API)`  |

At least one team/sport env var must be set for the plugin to return data.

**Configuration format**: One environment variable per sport, with comma-separated team identifiers.

---

## Display Behavior

For **each configured team**, the endpoint returns markup showing:

1. **Last completed game** — opponent, final score, W/L indicator, date.
2. **Next upcoming game** — opponent, scheduled date/time, venue (if available).

If no recent completed game exists (e.g. off-season), only the next game is shown, and vice versa.

---

## Endpoint Behavior

- **No authentication** is required on the endpoint (no API key or webhook verification).
- TRMNL will poll this endpoint approximately **every hour**.
- The server should cache ESPN API responses internally to avoid excessive upstream calls.
- Response format must conform to the TRMNL plugin screen generation flow (HTML/markup returned in the expected JSON envelope).

---

## Technical Requirements

- **Language**: Go
- **Deployment**: Portable — should run as a standalone binary or in a Docker container.
- The application should gracefully handle:
  - Sports with no configured teams (skip them).
  - ESPN API errors or timeouts (show stale cached data or a friendly message).
  - Off-season / no scheduled games for a team.

---

## Project Structure

```
TRMNL-Sports/
├── main.go              # Entry point, HTTP server
├── config/
│   └── config.go        # Environment variable parsing, sport definitions
├── espn/
│   ├── client.go        # ESPN API HTTP client with caching
│   └── types.go         # ESPN API response structs
├── trmnl/
│   └── render.go        # TRMNL HTML markup generation for all layouts
├── go.mod
├── go.sum
├── Dockerfile
├── PROMP.md             # This requirements document
└── README.md
```
