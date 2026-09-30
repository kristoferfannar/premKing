package crons

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kristo-og-logi/premKing/server/external"
	"github.com/kristo-og-logi/premKing/server/initializers"
	"github.com/kristo-og-logi/premKing/server/models"
	"github.com/kristo-og-logi/premKing/server/repositories"
)

// Compares all fixtures between Sportmonks and DB
// Tries to update their dates, odds or status.
//
// Collects and prints statistics about successful updates
func UpdateFixtures() {
	slog.Info("Fetching fixtures for updates")
	fixtures := external.FetchSportmonksFixtures()
	dbFixtures := getFixturesFromDB()
	gwsChanged := map[uint8]bool{}

	var totalOUpdated, totalDUpdated, totalSsUpdated, matches int = 0, 0, 0, 0
	for _, jsonfix := range fixtures {
		matchFound := false
		for _, dbfix := range dbFixtures {
			if dbfix.SportmonksID == jsonfix.Id {
				matchFound = true
				matches++

				if AssignOdds(dbfix, jsonfix) {
					totalOUpdated++
				}
				if UpdateDate(dbfix, jsonfix) {
					totalDUpdated++
					gwsChanged[dbfix.GameWeek] = true
				}
				if UpdateStatusAndScores(dbfix, jsonfix) {
					totalSsUpdated++
				}

				break
			}

		}
		if !matchFound {
			slog.Warn("No match found for API fixture", "ID", jsonfix.Id)
		}
	}
	fmt.Printf("%d matches found!\n", matches)
	fmt.Printf("Updated %d odds\n", totalOUpdated)
	fmt.Printf("Updated %d dates\n", totalDUpdated)
	fmt.Printf("Updated %d statuses\n", totalSsUpdated)

	// If any dates are updated,
	// it might affect when the gameweeks start
	if totalDUpdated > 0 {
		fmt.Printf("gws changed: %+v\n", gwsChanged)
		fmt.Println("\nFixture dates updated... checking for gw updates")
		FindAndSaveNormalFixtures() // we updated a fixture's date, we must check to see whether its normal status has changed
		ChangeGWTimes()
	}
}


func getSportmonksOdds(s models.SportmonksFixture) (string ,float64, string, float64) {
	var err error
	var sHomeOdds float64 = 0.0
	var sAwayOdds float64 = 0.0
	for _, odd := range s.Odds {
		if odd.OriginalLabel == "1" {
			sHomeOdds, err = strconv.ParseFloat(odd.Value, 64)
			if err != nil {
				panic("error converting string to float")
			}
		}
		if odd.OriginalLabel == "2" {
			sAwayOdds, err = strconv.ParseFloat(odd.Value, 64)
			if err != nil {
				panic("error converting string to float")
			}
		}
	}

	names := strings.Split(s.Name, " vs ")
	if len(names) != 2 {
		panic("expected two team names")
	}
	sHome := names[0]
	sAway := names[1]

	return sHome, sHomeOdds, sAway, sAwayOdds
}

func CompareSportmonksAndOdds() {
	slog.Info("Comparing OddsApi and Sportmonks")

	sBytes, err := os.ReadFile("sportmonks.json")
	if err != nil { panic("bad") }
	var sportmonks []models.SportmonksFixture
	err = json.Unmarshal(sBytes, &sportmonks)
	if err != nil { panic("bad") }
	// sportmonks := external.FetchSportmonksFixtures()

	oBytes, err := os.ReadFile("odds.json")
	if err != nil { panic("bad") }
	var odds []external.OddsApiFixture
	err = json.Unmarshal(oBytes, &odds)
	if err != nil { panic("bad") }
	// odds := external.FetchOdds()

	for _, s := range sportmonks {
		sTime, err := time.Parse( "2006-01-02 15:04:05", s.StartingAt)
		if err != nil {
			panic(fmt.Sprintf("bad time parse: err: %s", err.Error()))
		}
		sHome, sHomeOdds, sAway, sAwayOdds := getSportmonksOdds(s)

		var oHomeOdds float32 = 0.0
		var oAwayOdds float32 = 0.0

		for _, o := range odds {
			if sTime != o.CommenceTime {
				continue
			}

			oH, oHO, oA, oAO := o.CalculateOdds(s.Name)

			// odds doesn't sort based off home field
			if sHome == oA {
				tmp := oH
				oH= oA
				oA= tmp

				tmpOdds := oHO
				oHO = oAO
				oAO = tmpOdds
			}

			if sHome == oH && sAway == oA {
				oHomeOdds = oHO
				oAwayOdds = oAO
				break
			}

		}

		// print all games in the next two weeks
		latest := time.Now().Add(21 * 24 * time.Hour)
		earliest := time.Now().Add(-0 * 24 * time.Hour)

		if !(sTime.After(earliest) && sTime.Before(latest)) {
			continue
		}

		s1 := fmt.Sprintf("%s: %s vs %s", s.StartingAt, sHome, sAway)
		s3 := fmt.Sprintf("  Sportmonks: %.2f vs %.2f", sHomeOdds, sAwayOdds)
		s4 := fmt.Sprintf("        Odds: %.2f vs %.2f", oHomeOdds, oAwayOdds)

		fmt.Println(s1)
		fmt.Println(s3)
		fmt.Println(s4)

		slog.Info(s1)
		slog.Info(s3)
		slog.Info(s4)
	}
}

func OddsTeamToSportmonksTeamName(oddsName string) string {
	mp := map[string]string{
		 "Bournemouth": "AFC Bournemouth",
		 "Brighton and Hove Albion": "Brighton & Hove Albion",
	 }

	 lookup := mp[oddsName]

	 if lookup == "" {
		 // identity function
		 return oddsName
	 }

	 return lookup
}

// FindNormalFixtures groups and saves all fixtures
// by their gameweek, and finds out which are normal,
// meaning that they occur within the gameweek's timeframe.
//
// This is not a cron job, but needs to exist here for cron jobs to use
func FindAndSaveNormalFixtures() {
	fixtureList := make([][]models.Fixture, 38)

	// All fixtures have to be fetched before modifying the first one
	// Since calculating a fixture's normal status depends on fixtures in succeeding GWs
	for gw := 1; gw <= 38; gw++ {
		fixtures, err := repositories.FetchFixturesByGameweek(uint8(gw))
		if err != nil {
			slog.Error(fmt.Sprintf("couldn't find fixtures for GW%d", gw), "message", err.Error())
			return
		}

		fixtureList[gw-1] = fixtures
	}

	for gw := 1; gw <= 38; gw++ {
		slog.Info(fmt.Sprintf("GW%d", gw))
		fixtures := fixtureList[gw-1]

		for idx, fix := range fixtures {
			isNormal := false

			// fixture is normal if its in gw 38 or..
			if fix.GameWeek == 38 {
				isNormal = true
			} else {
				succ, err := firstNormalFixture(fixtureList[gw])
				if err != nil {
					slog.Error("error finding first normal fixture", "GW", gw)
					return
				}

				if gw == 1 {
					if fix.MatchDate.Sub(succ.MatchDate).Hours() <= 48 {
						isNormal = true
					}
				} else {
					pred, err := firstNormalFixture(fixtureList[gw-2])
					if err != nil {
						slog.Error("error finding first normal fixture", "GW", gw)
						return
					}

					// ..or if it's at least 48 hours before the first normal fixture of the succeeding gw
					// AND at least 48 hours after the first normal fixture of the preceding gw
					if fix.MatchDate.Sub(succ.MatchDate).Hours() <= 48 && pred.MatchDate.Sub(fix.MatchDate).Hours() <= 48 {
						isNormal = true
					}
				}
			}

			s := fmt.Sprintf("	%v", fix.MatchDate.Format("2006-01-02 15:04"))
			if isNormal {
				s += " - X"
			}
			slog.Info(s)

			fixtures[idx].IsNormal = isNormal
		}
		initializers.DB.Save(&fixtures)
	}
}

func firstNormalFixture(fx []models.Fixture) (*models.Fixture, error) {
	for _, f := range fx {
		if f.IsNormal {
			return &f, nil
		}
	}

	return nil, fmt.Errorf("no normal fixture in group")
}

func getFixturesFromDB() []models.Fixture {
	fixtures := []models.Fixture{}
	result := initializers.DB.Find(&fixtures)
	if result.Error != nil {
		fmt.Printf("error: %s\n", result.Error.Error())
		os.Exit(1)
	}

	fmt.Printf("%d fixtures in db!\n", len(fixtures))

	return fixtures
}
