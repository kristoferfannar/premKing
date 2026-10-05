package models

import "time"

type Fixture struct {
	// gorm.Model
	ID        uint32    `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	GameWeek     uint8     `json:"gameWeek"`
	HomeTeamId   uint16    `gorm:"index" json:"homeTeamId"`
	HomeTeam     Team      `gorm:"foreignKey:HomeTeamId" json:"homeTeam"`
	AwayTeamId   uint16    `gorm:"index" json:"awayTeamId"`
	AwayTeam     Team      `gorm:"foreignKey:AwayTeamId" json:"awayTeam"`
	Finished     bool      `json:"finished"`
	HomeGoals    uint8     `json:"homeGoals"`
	AwayGoals    uint8     `json:"awayGoals"`
	HomeOdds     float32   `json:"homeOdds"`
	DrawOdds     float32   `json:"drawOdds"`
	AwayOdds     float32   `json:"awayOdds"`
	Result       string    `json:"result"`
	MatchDate    time.Time `json:"matchDate"`
	Name         string    `json:"name"`
	LongName     string    `json:"longName"`
	SportmonksID uint32    `gorm:"unique" json:"sportmonksId"`
	IsNormal     bool      `json:"isNormal"`
}

// Api-football fixture
type FixturesResponse struct {
	Response []FixtureJSON `json:"response"`
}

type FixtureJSON struct {
	Fixture struct {
		ID     uint32    `json:"id"`
		Date   time.Time `json:"date"`
		Status struct {
			Elapsed uint16 `json:"elapsed"`
		} `json:"status"`
	} `json:"fixture"`
	League struct {
		Round string `json:"round"`
	} `json:"league"`
	Teams struct {
		Home struct {
			ID     uint16 `json:"id"`
			Name   string `json:"name"`
			Winner bool   `json:"winner"`
		} `json:"home"`
		Away struct {
			ID     uint16 `json:"id"`
			Name   string `json:"name"`
			Winner bool   `json:"winner"`
		} `json:"away"`
	} `json:"teams"`
	Goals struct {
		Home uint8 `json:"home"`
		Away uint8 `json:"away"`
	} `json:"goals"`
}

type SportmonksFixture struct {
	Id         uint32          `json:"id"`
	Name       string          `json:"name"`
	Round      SportmonksRound `json:"round"`
	Odds       []SportmonksOdd `json:"odds"`
	StartingAt string          `json:"starting_at"`
	Score      SportmonksScore `json:"score"`
	State      string          `json:"state"`
}

type SportmonksRound struct {
	Name string `json:"name"`
}

type SportmonksOdd struct {
	Value         string `json:"value"`
	OriginalLabel string `json:"original_label"`
}

type SportmonksScore struct {
	Home int `json:"home"`
	Away int `json:"away"`
}

// OddsApi
type OddsApiOdd struct {
	Name  string
	Price float32
}

type OddsAPIFixture struct {
	Name string
	CommenceTime time.Time
	Odds []OddsApiOdd
}
