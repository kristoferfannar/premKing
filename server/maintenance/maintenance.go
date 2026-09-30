package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kristo-og-logi/premKing/server/crons"
	"github.com/kristo-og-logi/premKing/server/initializers"
	"github.com/kristo-og-logi/premKing/server/models"
	"github.com/kristo-og-logi/premKing/server/repositories"
	"github.com/kristo-og-logi/premKing/server/utils"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

func main() {
	env := initializers.LoadEnv()
	initializers.Logging(env)
	initializers.ConnectDB()

	crons.CompareSportmonksAndOdds()

	// SendNotification("test@test.com")
	// AddBetsForUser()
	// ShortenFixtureNames()
	// crons.UpdateFixtures()
	// crons.FindAndSaveNormalFixtures()
	// CreateBets()
	// ChangeGWTimes()
	// AddOddsAndWonToBets()
	// RecalculateBetsForGameweek()
	// SeedTeamsFromJSON("teams.json")
	// SeedFixturesFromJSON("fixtures.json")
	// initializers.MigrateGameweeksToDB(initializers.DB)

	// NEWSEASON()
}

func NEWSEASON() {
	// Run this whenever you have a new season

	crons.FindAndSaveNormalFixtures()
	ChangeGWTimes()
	crons.UpdateFixtures()
}

func SendNotification(email string) {
	user := getUser(email)
	client := expo.NewPushClient(nil)
	crons.PublishNotification(user.ExpoPushToken, client, "test", "this")
}

func AddBetsForUser() {
	fmt.Print("gameweek: ")
	gameweekBytes, _, _ := bufio.NewReader(os.Stdin).ReadLine()
	gameweekString := string(gameweekBytes)
	gameweek, err := strconv.Atoi(gameweekString)
	if err != nil {
		panic("invalid gameweek")
	}

	fmt.Print("email: ")
	userEmail, _, _ := bufio.NewReader(os.Stdin).ReadLine()
	user := getUser(string(userEmail))

	curr, _ := repositories.GetCurrentGameWeek()
	if gameweek > int(curr.Gameweek) {
		panic("gameweek has not started")
	}

	dbBets, _ := repositories.GetBetsByUserIdAndGameweek(user.ID, gameweek)
	if len(dbBets) > 0 {
		panic("user already has bets for gameweek")
	}

	fixtures := getNormalFixturesByGW(gameweek)
	bets := []*models.Bet{}
	for _, fix := range fixtures {
		fmt.Println(fix.Name)
		fmt.Print("bet (1 | X | 2): ")
		userGuess, _, _ := bufio.NewReader(os.Stdin).ReadLine()
		guess := string(userGuess)

		var odd float32
		var won bool = false
		switch guess {
		case "1":
			if fix.Result == "1" {
				won = true
			}
			odd = fix.HomeOdds
		case "X":
			if fix.Result == "X" {
				won = true
			}
			odd = fix.DrawOdds
		case "2":
			if fix.Result == "2" {
				won = true
			}
			odd = fix.AwayOdds
		default:
			panic("invalid guess")
		}
		bet := models.Bet{ID: uuid.NewString(), UserId: user.ID, FixtureId: fix.ID, Result: guess, GameWeek: fix.GameWeek, Won: won, Odd: odd}
		bets = append(bets, &bet)
	}
	saveBets(bets)
	calculateBets(bets, fixtures)

	fmt.Println("all bets saved :)")
}

// Make fixture names (team1 vs team2) use teams' short names
func ShortenFixtureNames() {
	gameweeks, _ := repositories.GetAllGameWeeks()

	for _, gw := range gameweeks {
		fixtures, _ := repositories.FetchFixturesByGameweek(gw.Gameweek)

		for idx, fix := range fixtures {
			name := utils.CreateFixtureName(fix.HomeTeam, fix.AwayTeam)
			fixtures[idx].Name = name
		}

		initializers.DB.Save(&fixtures)
		fmt.Printf("saved %d fixtures for GW%d\n", len(fixtures), gw.Gameweek)
	}

}

func findTeamByName(teams []models.Team, name string) *models.Team {
	for _, team := range teams {
		if team.Name == name {
			return &team
		}
	}

	return nil
}

// CalculateLevenshteinDistance calculates the Levenshtein distance between two strings.
func CalculateLevenshteinDistance(s1, s2 string) int {
	lenS1, lenS2 := len(s1), len(s2)

	// Create a 2D slice for dynamic programming
	dp := make([][]int, lenS1+1)
	for i := range dp {
		dp[i] = make([]int, lenS2+1)
	}

	// Initialize base cases
	for i := 0; i <= lenS1; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= lenS2; j++ {
		dp[0][j] = j
	}

	// Fill the dp table
	for i := 1; i <= lenS1; i++ {
		for j := 1; j <= lenS2; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = min(dp[i-1][j]+1, min(dp[i][j-1]+1, dp[i-1][j-1]+1))
			}
		}
	}

	return dp[lenS1][lenS2]
}

// min returns the minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PairSimilarStrings takes two equal-sized arrays of strings and returns an array of pairs of the most similar strings.
func PairSimilarStrings(arr1, arr2 []string) [][2]string {
	if len(arr1) != len(arr2) {
		return nil // Return nil if the arrays are not of equal size
	}

	pairedStrings := make([][2]string, len(arr1))
	usedIndices := make(map[int]bool)

	for i, str1 := range arr1 {
		minDistance := int(^uint(0) >> 1) // Max int value
		bestMatchIndex := -1

		for j, str2 := range arr2 {
			if usedIndices[j] {
				continue // Skip already paired strings
			}

			distance := CalculateLevenshteinDistance(str1, str2)
			if distance < minDistance {
				minDistance = distance
				bestMatchIndex = j
			}
		}

		pairedStrings[i] = [2]string{str1, arr2[bestMatchIndex]}
		usedIndices[bestMatchIndex] = true
	}

	return pairedStrings
}

func RecalculateBetsForGameweek() {
	fmt.Print("gameweek: ")
	gameweekBytes, _, _ := bufio.NewReader(os.Stdin).ReadLine()
	gameweekString := string(gameweekBytes)
	gameweek, err := strconv.Atoi(gameweekString)
	if err != nil {
		panic("invalid gameweek")
	}

	fixtures := getNormalFixturesByGW(gameweek)
	bets := getBetsByGameweek(gameweek)

	calculateBets(bets, fixtures)

	initializers.DB.Save(&bets)

	fmt.Printf("all bets for GW%d recalculated\n", gameweek)
}

func CreateBets() {
	fmt.Print("email: ")
	userEmail, _, _ := bufio.NewReader(os.Stdin).ReadLine()

	fmt.Print("bet (1 | X | 2): ")
	userGuess, _, _ := bufio.NewReader(os.Stdin).ReadLine()

	user := getUser(string(userEmail))
	guess := string(userGuess)

	curr, _ := repositories.GetCurrentGameWeek()

	for gw := 1; gw <= int(curr.Gameweek); gw++ {
		fixtures := getNormalFixturesByGW(gw)
		bets := createBets(user, fixtures, guess)
		saveBets(bets)
		calculateBets(bets, fixtures)
	}

	fmt.Println("all bets saved :)")
}

func AddOddsAndWonToBets() {
	allBets := getAllNonUpdatedBets()
	allFixtures := getAllPastFixtures()

	fmt.Printf("found %d bets\n", len(allBets))
	fmt.Printf("found %d fixtures\n", len(allFixtures))

	updateOddsAndWon(allFixtures, allBets)
}

// ChangeGWTimes updates, if necessary, all gameweeks's
// Opens, Closes and Finishes attributes depending on
// whether the normal fixture matchDates changed
func ChangeGWTimes() {
	gws := getGWs()

	opensUpdated := 0
	closesUpdated := 0
	finishesUpdated := 0
	for idx := range gws {
		fixtures, err := repositories.FetchNormalFixturesByGameweek(uint8(gws[idx].Gameweek))
		if err != nil {
			fmt.Printf("couldn't find fixtures for GW%d: %s", gws[idx].Gameweek, err.Error())
			continue
		}

		if gws[idx].Closes.Compare(fixtures[0].MatchDate.Add(-2*time.Hour)) != 0 {
			gws[idx].Closes = fixtures[0].MatchDate.Add(-2 * time.Hour)
			closesUpdated++
		}
		if gws[idx].Finishes.Compare(fixtures[len(fixtures)-1].MatchDate.Add(2*time.Hour)) != 0 {
			gws[idx].Finishes = fixtures[len(fixtures)-1].MatchDate.Add(2 * time.Hour)
			finishesUpdated++

			if idx < 37 {
				gws[idx+1].Opens = fixtures[len(fixtures)-1].MatchDate.Add(2 * time.Hour)
				opensUpdated++
			}
		}
	}

	initializers.DB.Save(&gws)
	fmt.Printf("gw.Opens updated: %d\n", opensUpdated)
	fmt.Printf("gw.Closes updated: %d\n", closesUpdated)
	fmt.Printf("gw.Finishes updated: %d\n", finishesUpdated)
}

func getGWs() []models.Gameweek {
	gws := []models.Gameweek{}

	result := initializers.DB.Find(&gws)

	if result.Error != nil {
		fmt.Printf("error fetching gameweeks: %s", result.Error.Error())
	}

	sort.Slice(gws, func(i, j int) bool {
		return gws[i].Gameweek < gws[j].Gameweek
	})

	return gws
}

func updateOddsAndWon(fx []models.Fixture, bts []*models.Bet) {
	updated := 0

	for _, f := range fx {
		for idx, b := range bts {
			if b.FixtureId == f.ID {
				updateBet(bts[idx], f)
				updated++
			}
		}
	}

	// Use .Select() to explicitly update Won column
	// , otherwise it doesn't update rows with won = false
	initializers.DB.Save(&bts)

	fmt.Printf("updated %d bets\n", updated)
}

func updateBet(b *models.Bet, f models.Fixture) {
	var odd float32
	switch b.Result {
	case "1":
		odd = f.HomeOdds
	case "X":
		odd = f.DrawOdds
	case "2":
		odd = f.AwayOdds
	}

	b.Odd = odd
	b.Won = f.Finished && b.Result == f.Result
}

// Gets all bets which haven't had their odds updated.
// Naturally, these bets haven't been declared as either won or lost either.
func getAllNonUpdatedBets() []*models.Bet {
	bets := []*models.Bet{}

	result := initializers.DB.Find(&bets, "odd = 0")
	if result.Error != nil {
		fmt.Printf("error fetching all bets: %s", result.Error.Error())
		return nil
	}

	return bets
}

func getAllPastFixtures() []models.Fixture {
	fixtures := []models.Fixture{}

	result := initializers.DB.Find(&fixtures, "home_odds is not null")
	if result.Error != nil {
		fmt.Printf("error fetching all fixtures: %s", result.Error.Error())
		return nil
	}

	return fixtures
}

func getUser(email string) *models.User {
	user := &models.User{}

	result := initializers.DB.Find(user, "email = ?", email)
	if result.Error != nil {
		fmt.Printf("error fetching user with email %s: %s", email, result.Error.Error())
		panic("oh no")
	}

	if result.RowsAffected == 0 {
		panic("No user with provided email")
	}

	return user
}

func createBets(user *models.User, fixtures []models.Fixture, result string) []*models.Bet {
	bets := []*models.Bet{}

	for _, fix := range fixtures {
		var odd float32
		var won bool = false
		switch result {
		case "1":
			if fix.Result == "1" {
				won = true
			}
			odd = fix.HomeOdds
		case "X":
			if fix.Result == "X" {
				won = true
			}
			odd = fix.DrawOdds
		case "2":
			if fix.Result == "2" {
				won = true
			}
			odd = fix.AwayOdds
		}
		bet := models.Bet{ID: uuid.NewString(), UserId: user.ID, FixtureId: fix.ID, Result: result, GameWeek: fix.GameWeek, Won: won, Odd: odd}
		bets = append(bets, &bet)
	}

	return bets
}

func saveBets(bets []*models.Bet) {
	result := initializers.DB.Save(&bets)
	if result.Error != nil {
		fmt.Printf("error saving bets: %s", result.Error.Error())
	}
}

func getNormalFixturesByGW(gw int) []models.Fixture {
	fixtures, err := repositories.FetchNormalFixturesByGameweek(uint8(gw))
	if err != nil {
		fmt.Printf("error fetching normal fixtures for GW%d\n", gw)
		os.Exit(1)
	}

	return fixtures
}

func getBetsByGameweek(gw int) []*models.Bet {
	bets, err := repositories.GetBetsByGameweek(gw)
	if err != nil {
		panic(err.Error())
	}

	return bets
}

func calculateBets(bets []*models.Bet, fixtures []models.Fixture) {
	for _, fix := range fixtures {
		for idx, bet := range bets {
			if fix.ID == bet.FixtureId {
				switch bet.Result {
				case "1":
					bets[idx].Odd = fix.HomeOdds

				case "X":
					bets[idx].Odd = fix.DrawOdds

				case "2":
					bets[idx].Odd = fix.AwayOdds
				}

				bets[idx].Won = fix.Finished && fix.Result == bet.Result
			}
		}
	}
}
