package external

import (
	"math"
	"testing"
)

// Real Arsenal vs Leeds United fixture pulled with regions=eu (no bookmakers
// filter), giving odds from 21 bookmakers. betfair_ex_eu additionally
// carries an h2h_lay market, which AverageOdds must ignore.
func getMultiBookmakerOddsData() []byte {
	out := `
{
  "commence_time": "2026-10-10T11:30:00Z",
  "home_team": "Arsenal",
  "away_team": "Leeds United",
  "bookmakers": [
    {"key": "codere_it", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.42}, {"name": "Leeds United", "price": 7.2}, {"name": "Draw", "price": 4.8}]}]},
    {"key": "unibet_fr", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.35}, {"name": "Leeds United", "price": 6.4}, {"name": "Draw", "price": 4.3}]}]},
    {"key": "winamax_de", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.36}, {"name": "Leeds United", "price": 6.75}, {"name": "Draw", "price": 4.8}]}]},
    {"key": "sport888", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.36}, {"name": "Leeds United", "price": 7.5}, {"name": "Draw", "price": 4.6}]}]},
    {"key": "betclic_fr", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.37}, {"name": "Leeds United", "price": 6.75}, {"name": "Draw", "price": 4.5}]}]},
    {"key": "onexbet", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.44}, {"name": "Leeds United", "price": 8.2}, {"name": "Draw", "price": 5.03}]}]},
    {"key": "mybookieag", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.4}, {"name": "Leeds United", "price": 6.9}, {"name": "Draw", "price": 4.6}]}]},
    {"key": "williamhill", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.36}, {"name": "Leeds United", "price": 7.5}, {"name": "Draw", "price": 4.6}]}]},
    {"key": "winamax_fr", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.36}, {"name": "Leeds United", "price": 6.75}, {"name": "Draw", "price": 4.7}]}]},
    {"key": "marathonbet", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.4}, {"name": "Leeds United", "price": 8.0}, {"name": "Draw", "price": 4.9}]}]},
    {"key": "betfair_ex_eu", "markets": [
      {"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.39}, {"name": "Leeds United", "price": 9.2}, {"name": "Draw", "price": 5.6}]},
      {"key": "h2h_lay", "outcomes": [{"name": "Arsenal", "price": 1.4}, {"name": "Leeds United", "price": 9.8}, {"name": "Draw", "price": 5.7}]}
    ]},
    {"key": "tipico_de", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.35}, {"name": "Leeds United", "price": 7.8}, {"name": "Draw", "price": 5.2}]}]},
    {"key": "pinnacle", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.38}, {"name": "Leeds United", "price": 7.65}, {"name": "Draw", "price": 4.93}]}]},
    {"key": "leovegas_se", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.36}, {"name": "Leeds United", "price": 8.0}, {"name": "Draw", "price": 5.1}]}]},
    {"key": "unibet_se", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.38}, {"name": "Leeds United", "price": 8.5}, {"name": "Draw", "price": 5.2}]}]},
    {"key": "unibet_nl", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.38}, {"name": "Leeds United", "price": 8.5}, {"name": "Draw", "price": 5.2}]}]},
    {"key": "pmu_fr", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.32}, {"name": "Leeds United", "price": 7.5}, {"name": "Draw", "price": 4.7}]}]},
    {"key": "everygame", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.36}, {"name": "Leeds United", "price": 7.25}, {"name": "Draw", "price": 4.75}]}]},
    {"key": "nordicbet", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.38}, {"name": "Leeds United", "price": 8.7}, {"name": "Draw", "price": 4.7}]}]},
    {"key": "coolbet", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.38}, {"name": "Leeds United", "price": 8.5}, {"name": "Draw", "price": 5.0}]}]},
    {"key": "betonlineag", "markets": [{"key": "h2h", "outcomes": [{"name": "Arsenal", "price": 1.38}, {"name": "Leeds United", "price": 8.25}, {"name": "Draw", "price": 5.25}]}]}
  ]
}
`
	return []byte(out)
}

func Test_CalculateOdds_MultipleBookmakers(t *testing.T) {
	data := getMultiBookmakerOddsData()

	fixtures, err := toOddsApiFixtures([]byte("[" + string(data) + "]"))
	if err != nil {
		t.Fatalf("bad: %s", err.Error())
	}
	if len(fixtures) != 1 {
		t.Fatalf("expected 1 fixture, got %d", len(fixtures))
	}
	fixture := fixtures[0]

	if len(fixture.Bookmakers) != 21 {
		t.Fatalf("expected 21 bookmakers, got %d", len(fixture.Bookmakers))
	}

	homeTeam, homeAvg, awayTeam, awayAvg := fixture.CalculateOdds("Arsenal vs Leeds United")

	if homeTeam != "Arsenal" {
		t.Fatalf("expected home team Arsenal, got %s", homeTeam)
	}
	if awayTeam != "Leeds United" {
		t.Fatalf("expected away team Leeds United, got %s", awayTeam)
	}

	// expected values computed independently from the same 21 bookmakers,
	// excluding betfair_ex_eu's h2h_lay market
	wantHome := float32(1.375238)
	wantAway := float32(7.704762)

	if math.Abs(float64(homeAvg-wantHome)) > 0.0001 {
		t.Fatalf("expected home avg ~%.6f, got %.6f", wantHome, homeAvg)
	}
	if math.Abs(float64(awayAvg-wantAway)) > 0.0001 {
		t.Fatalf("expected away avg ~%.6f, got %.6f", wantAway, awayAvg)
	}
}
