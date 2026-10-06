package crons

import (
	"fmt"
	"strconv"

	"github.com/kristo-og-logi/premKing/server/initializers"
	"github.com/kristo-og-logi/premKing/server/models"
)

func AssignOdds(dbFixture models.Fixture, jsonFixture models.SportmonksFixture) (updated bool) {
	updated = false
	// no need to reassign odds
	if dbFixture.AwayOdds >= 0.1 || dbFixture.HomeOdds >= 0.1 || dbFixture.DrawOdds >= 0.1 {
		return
	}
	// if odds haven't appeard in API, don't update
	if len(jsonFixture.Odds) == 0 {
		return
	}

	var homeOdd, drawOdd, awayOdd float32
	for _, odd := range jsonFixture.Odds {
		if odd.OriginalLabel == "2" {
			oddValue, _ := strconv.ParseFloat(odd.Value, 32)
			awayOdd = float32(oddValue)
		} else if odd.OriginalLabel == "1" {
			oddValue, _ := strconv.ParseFloat(odd.Value, 32)
			homeOdd = float32(oddValue)
		} else if odd.OriginalLabel == "Draw" {
			oddValue, _ := strconv.ParseFloat(odd.Value, 32)
			drawOdd = float32(oddValue)
		} else {
			panic(fmt.Sprintf("found odd with original_label: %v\n", odd.OriginalLabel))
		}
	}
	initializers.DB.Model(&dbFixture).Updates(models.Fixture{HomeOdds: homeOdd, DrawOdds: drawOdd, AwayOdds: awayOdd})
	updated = true
	return
}

func AssignOddsApiOdds(dbFixture models.Fixture, homeOdds float32, drawOdds float32, awayOdds float32) (updated bool) {
	updated = false
	// no need to reassign odds
	if dbFixture.AwayOdds >= 0.1 || dbFixture.HomeOdds >= 0.1 || dbFixture.DrawOdds >= 0.1 {
		return
	}

	initializers.DB.Model(&dbFixture).Updates(models.Fixture{HomeOdds: homeOdds, DrawOdds: drawOdds, AwayOdds: awayOdds})
	updated = true
	return
}
