package crons

import (
	"fmt"
	"log/slog"

	"github.com/kristo-og-logi/premKing/server/initializers"
	"github.com/kristo-og-logi/premKing/server/models"
	"github.com/kristo-og-logi/premKing/server/repositories"
)

func UpdateStatusAndScores(dbFixture models.Fixture, jsonFixture models.SportmonksFixture) (updated bool) {
	updated = false

	if jsonFixture.State == "FT" && !dbFixture.Finished {

		result := ""
		if jsonFixture.Score.Home > jsonFixture.Score.Away {
			result = "1"
		} else if jsonFixture.Score.Home < jsonFixture.Score.Away {
			result = "2"
		} else {
			result = "X"
		}

		dbFixture.Finished = true
		dbFixture.HomeGoals = uint8(jsonFixture.Score.Home)
		dbFixture.AwayGoals = uint8(jsonFixture.Score.Away)
		dbFixture.Result = result

		response := initializers.DB.Save(&dbFixture)

		UpdateBetsByFixture(&dbFixture)

		if response.Error != nil {
			fmt.Printf("error saving new status and score to fixture: %s", response.Error.Error())
			return
		}

		updated = true
	}

	return
}

// TODO
func UpdateOddsApiStatusAndScores(dbFixture models.Fixture, jsonFixture models.SportmonksFixture) (updated bool) {
	updated = false

	if jsonFixture.State == "FT" && !dbFixture.Finished {

		result := ""
		if jsonFixture.Score.Home > jsonFixture.Score.Away {
			result = "1"
		} else if jsonFixture.Score.Home < jsonFixture.Score.Away {
			result = "2"
		} else {
			result = "X"
		}

		dbFixture.Finished = true
		dbFixture.HomeGoals = uint8(jsonFixture.Score.Home)
		dbFixture.AwayGoals = uint8(jsonFixture.Score.Away)
		dbFixture.Result = result

		response := initializers.DB.Save(&dbFixture)

		UpdateBetsByFixture(&dbFixture)

		if response.Error != nil {
			fmt.Printf("error saving new status and score to fixture: %s", response.Error.Error())
			return
		}

		updated = true
	}

	return
}

func UpdateBetsByFixture(dbFix *models.Fixture) {
	updated := 0
	bets, err := repositories.GetBetsByFixtureId(dbFix.ID)
	if err != nil {
		fmt.Printf("ERROR fetching bets by fixture Id: %s", err.Error())
		return
	}

	for idx, bet := range bets {
		if dbFix.Result == bet.Result {
			bets[idx].Won = true
			updated++
		}
	}

	initializers.DB.Save(&bets)

	slog.Info(fmt.Sprintf("Updated %d / %d bets for fixture %s\n", updated, len(bets), dbFix.Name))
}
