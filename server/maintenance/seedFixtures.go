package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kristo-og-logi/premKing/server/external"
	"github.com/kristo-og-logi/premKing/server/initializers"
	"github.com/kristo-og-logi/premKing/server/models"
	"github.com/kristo-og-logi/premKing/server/utils"
)

// SeedFixturesFromJSON reads a JSON file of raw Sportmonks fixtures
// (external.Match, as crawled by scripts/crawl_fixtures.sh) and writes
// each one to the DB as a models.Fixture, resolving home/away teams by
// name against teams already in the DB. Skips fixtures that already
// exist by ID.
func SeedFixturesFromJSON(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("error reading %s: %s\n", path, err.Error())
		os.Exit(1)
	}

	var matches []external.Match
	if err := json.Unmarshal(data, &matches); err != nil {
		fmt.Printf("error parsing %s: %s\n", path, err.Error())
		os.Exit(1)
	}

	var teams []models.Team
	if result := initializers.DB.Find(&teams); result.Error != nil {
		fmt.Printf("error fetching teams: %s\n", result.Error.Error())
		os.Exit(1)
	}
	teamsByName := map[string]models.Team{}
	for _, team := range teams {
		teamsByName[team.Name] = team
	}

	added := 0
	for _, match := range matches {
		names := strings.Split(match.Name, " vs ")
		if len(names) != 2 {
			fmt.Printf("skipping fixture with unexpected name format: %s\n", match.Name)
			continue
		}

		homeTeam, homeOk := teamsByName[names[0]]
		awayTeam, awayOk := teamsByName[names[1]]
		if !homeOk || !awayOk {
			fmt.Printf("skipping fixture %s: team(s) not found in db\n", match.Name)
			continue
		}

		matchDate, err := time.Parse("2006-01-02 15:04:05", match.StartingAt)
		if err != nil {
			fmt.Printf("skipping fixture %s: could not parse starting_at %q: %s\n", match.Name, match.StartingAt, err.Error())
			continue
		}

		gameWeek, err := strconv.Atoi(match.Round.Name)
		if err != nil {
			fmt.Printf("skipping fixture %s: could not parse round %q\n", match.Name, match.Round.Name)
			continue
		}

		var homeOdds, drawOdds, awayOdds float32
		for _, odd := range match.Odds {
			value, _ := strconv.ParseFloat(odd.Value, 32)
			switch odd.OriginalLabel {
			case "1":
				homeOdds = float32(value)
			case "2":
				awayOdds = float32(value)
			case "Draw":
				drawOdds = float32(value)
			}
		}

		var homeGoals, awayGoals uint8
		for _, score := range match.Scores {
			if score.Description != "2ND_HALF" {
				continue
			}
			switch score.Score.Participant {
			case "home":
				homeGoals = uint8(score.Score.Goals)
			case "away":
				awayGoals = uint8(score.Score.Goals)
			}
		}

		finished := match.State.State == "FT"
		result := ""
		if finished {
			if homeGoals > awayGoals {
				result = "1"
			} else if homeGoals < awayGoals {
				result = "2"
			} else {
				result = "X"
			}
		}

		fixture := models.Fixture{
			ID:           uint32(match.ID),
			GameWeek:     uint8(gameWeek),
			HomeTeamId:   homeTeam.ID,
			AwayTeamId:   awayTeam.ID,
			Finished:     finished,
			HomeGoals:    homeGoals,
			AwayGoals:    awayGoals,
			HomeOdds:     homeOdds,
			DrawOdds:     drawOdds,
			AwayOdds:     awayOdds,
			Result:       result,
			MatchDate:    matchDate,
			Name:         utils.CreateFixtureName(homeTeam, awayTeam),
			LongName:     match.Name,
			SportmonksID: uint32(match.ID),
			IsNormal:     true,
		}

		dbResult := initializers.DB.Where(models.Fixture{ID: fixture.ID}).FirstOrCreate(&fixture)
		if dbResult.Error != nil {
			fmt.Printf("error adding fixture %s: %s\n", fixture.Name, dbResult.Error.Error())
			continue
		}

		if dbResult.RowsAffected > 0 {
			added++
			fmt.Printf("added %s (GW%d)\n", fixture.Name, fixture.GameWeek)
		} else {
			fmt.Printf("already exists: %s\n", fixture.Name)
		}
	}

	fmt.Printf("done: added %d/%d fixtures\n", added, len(matches))
}
