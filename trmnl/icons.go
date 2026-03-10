package trmnl

// SportIcon returns an inline SVG line-art icon for the given league key.
// All icons are 24x24, stroke-only (no fill), suitable for e-ink display.
func SportIcon(leagueKey string) string {
	switch leagueKey {
	case "nfl", "college-football":
		return svgFootball
	case "nba", "mens-college-basketball":
		return svgBasketball
	case "mlb":
		return svgBaseball
	case "nhl":
		return svgHockeyPuck
	case "usa.1", "eng.1":
		return svgSoccerBall
	case "ufc":
		return svgFist
	default:
		return svgGenericBall
	}
}

// American football — pointed oval with laces
const svgFootball = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
  <ellipse cx="12" cy="12" rx="10" ry="6" transform="rotate(-30 12 12)"/>
  <line x1="12" y1="7.5" x2="12" y2="16.5"/>
  <line x1="10" y1="9" x2="14" y2="9"/>
  <line x1="10" y1="11" x2="14" y2="11"/>
  <line x1="10" y1="13" x2="14" y2="13"/>
  <line x1="10" y1="15" x2="14" y2="15"/>
</svg>`

// Basketball — circle with seam lines
const svgBasketball = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
  <circle cx="12" cy="12" r="10"/>
  <line x1="12" y1="2" x2="12" y2="22"/>
  <line x1="2" y1="12" x2="22" y2="12"/>
  <path d="M5.2 5.2C8 8 8 16 5.2 18.8"/>
  <path d="M18.8 5.2C16 8 16 16 18.8 18.8"/>
</svg>`

// Baseball — circle with stitching curves
const svgBaseball = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
  <circle cx="12" cy="12" r="10"/>
  <path d="M6.5 3.5C8.5 7 8.5 17 6.5 20.5"/>
  <path d="M17.5 3.5C15.5 7 15.5 17 17.5 20.5"/>
</svg>`

// Hockey puck — flat cylinder from the side
const svgHockeyPuck = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
  <ellipse cx="12" cy="9" rx="10" ry="4"/>
  <line x1="2" y1="9" x2="2" y2="15"/>
  <line x1="22" y1="9" x2="22" y2="15"/>
  <ellipse cx="12" cy="15" rx="10" ry="4"/>
</svg>`

// Soccer ball — circle with pentagon pattern
const svgSoccerBall = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
  <circle cx="12" cy="12" r="10"/>
  <polygon points="12,7 14.5,9 13.5,12 10.5,12 9.5,9"/>
  <line x1="12" y1="7" x2="12" y2="3"/>
  <line x1="14.5" y1="9" x2="19" y2="7"/>
  <line x1="13.5" y1="12" x2="18" y2="15"/>
  <line x1="10.5" y1="12" x2="6" y2="15"/>
  <line x1="9.5" y1="9" x2="5" y2="7"/>
</svg>`

// MMA fist — simple closed fist outline
const svgFist = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
  <path d="M7 12V8a2 2 0 0 1 4 0v1a2 2 0 0 1 4 0v1a2 2 0 0 1 4 0v4c0 4-3 6-6 7H9c-3-1-5-4-5-7v-2a2 2 0 0 1 3 0"/>
</svg>`

// Generic ball fallback
const svgGenericBall = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
  <circle cx="12" cy="12" r="10"/>
</svg>`
