package crons

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kristo-og-logi/premKing/server/initializers"
	"github.com/kristo-og-logi/premKing/server/models"
)

func UpdateDate(dbFixture models.Fixture, jsonFixture models.SportmonksFixture) (updated bool) {
	updated = false

	jsonTime, err := time.Parse("2006-01-02 15:04:05", jsonFixture.StartingAt)
	if err != nil {
		// fmt.Printf("err - %s\n", err.Error())
		return
	}

	if jsonTime.Compare(dbFixture.MatchDate) != 0 {
		fmt.Printf("%s | from %s to %s\n", dbFixture.Name, dbFixture.MatchDate, jsonTime)
		initializers.DB.Model(&dbFixture).Updates(models.Fixture{MatchDate: jsonTime})
		updated = true
	}
	return
}

func UpdateOddsApiDate(dbFixture models.Fixture, commenceTime time.Time) (updated bool) {
	updated = false

	if commenceTime.Compare(dbFixture.MatchDate) != 0 {
		slog.Info("%s | from %s to %s\n", dbFixture.Name, dbFixture.MatchDate, commenceTime)
		initializers.DB.Model(&dbFixture).Updates(models.Fixture{MatchDate: commenceTime})
		updated = true
	}
	return
}
