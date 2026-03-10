# TRMNL Sports Plugin — Backend Endpoint

## Overview

A **Go** application that serves as a TRMNL plugin backend endpoint. It pulls sports data from the public ESPN API and returns TRMNL-compatible markup showing the **last completed game result** and **next upcoming game** for each configured team.

---

## TRMNL Private Plugin Setup

Create a new Private Plugin in TRMNL and configure it with the settings below.

### Strategy

Select **Polling**.

### Polling URL

```
http://your.url
```

### Polling Verb

```
GET
```

### Polling Headers

```
content-type=application/json
```

### Polling Body

Leave **empty** — no body is needed for the GET request.

### Form Fields

Leave **empty** — team configuration is handled via server-side environment variables (see [Environment Variables](#environment-variables) below).

### Markup Editor

After saving the plugin, click **"Edit Markup"** and enter the following in each tab. These use TRMNL's Liquid templating to pass through the pre-rendered HTML from the polling response.

**Full** tab:

```liquid
{{ markup }}
```

**Half Horizontal** tab:

```liquid
{{ markup_half_horizontal }}
```

**Half Vertical** tab:

```liquid
{{ markup_half_vertical }}
```

**Quadrant** tab:

```liquid
{{ markup_quadrant }}
```

> **Tip:** After saving the markup templates, click **"Force Refresh"** in the plugin settings to pull fresh data and generate the first screen.

---

## Environment Variables

| Variable        | Required | Description                                                       | Example                |
|-----------------|----------|-------------------------------------------------------------------|------------------------|
| `LISTEN_ADDR`   | No       | Host:port the server listens on (default `0.0.0.0:8080`)         | `0.0.0.0:8080`        |
| `NFL_TEAMS`     | No       | Comma-separated ESPN team abbreviations for NFL                   | `PHI,DAL,KC`          |
| `NBA_TEAMS`     | No       | Comma-separated ESPN team abbreviations for NBA                   | `LAL,BOS`             |
| `MLB_TEAMS`     | No       | Comma-separated ESPN team abbreviations for MLB                   | `NYY,LAD`             |
| `NHL_TEAMS`     | No       | Comma-separated ESPN team abbreviations for NHL                   | `NYR,BOS`             |
| `NCAAF_TEAMS`   | No       | Comma-separated ESPN team abbreviations for College Football      | `MICH,OSU`            |
| `NCAAM_TEAMS`   | No       | Comma-separated ESPN team abbreviations for College Basketball    | `DUKE,UNC`            |
| `MLS_TEAMS`     | No       | Comma-separated ESPN team abbreviations for MLS                   | `ATL,LAFC`            |
| `EPL_TEAMS`     | No       | Comma-separated ESPN team abbreviations for Premier League        | `LIV,ARS`             |
| `UFC_FIGHTERS`  | No       | Comma-separated ESPN fighter IDs or names for UFC                 | *(TBD per ESPN API)*  |

At least one team/sport variable must be set for the plugin to return data.
