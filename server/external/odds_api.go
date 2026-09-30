package external

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)


func createOddsUrl(apiKey string) string {
	return fmt.Sprintf("https://api.the-odds-api.com/v4/sports/soccer_epl/odds?regions=eu&apiKey=%s", apiKey)
}

func createOddsRequest() *http.Request {
	oddsApi := "ODDS_API_KEY"
	apiKey := os.Getenv(oddsApi)
	if apiKey == "" {
		fmt.Printf("%s not found in env\n", oddsApi)
		return nil
	}

	url := createOddsUrl(apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("error creating a GET request: %s", err.Error())
		return nil
	}


	return req
}

func toOddsApiFixtures(body []byte) ([]OddsApiFixture, error) {
	var fixtures []OddsApiFixture

	err := json.Unmarshal(body, &fixtures)
	if err != nil {
		fmt.Printf("error unmarshalling body into fixtures struct: %s", err.Error())
		return nil, err
	}

	return fixtures, nil

}

func getOddsResponse(request *http.Request) []OddsApiFixture {
	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		fmt.Printf("error fetching request: %s", err.Error())
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error reading response body: %s", err.Error())
		return nil
	}

	fixtures, err := toOddsApiFixtures(body)
	if err != nil {
		return nil
	}

	return fixtures
}


func FetchOdds() []OddsApiFixture{
	req := createOddsRequest()
	resp := getOddsResponse(req)

	bytes, err := json.Marshal(resp)
	if err != nil {
		fmt.Errorf("error marshalling: %w", err)
	}

	filename := fmt.Sprintf("odds-%s.json", time.Now().Format("2006-01-02_15-04-05"))
	slog.Info("wrote new OddsApi data", "filename", filename)
	os.WriteFile(filename, bytes, 0644)
	os.Remove("odds.json")
	os.Symlink(filename, "odds.json")

	return resp
}

func (odds *OddsApiFixture) CalculateOdds(name string) (string, float32, string, float32) {
	homeTeam := ""
	awayTeam := ""
	var home float32 = 0.0
	var away float32 = 0.0
	found := 0

	if len(odds.Bookmakers) == 0 {
		slog.Warn("OddsApi: No bookmakers", "fixture", name)
		return homeTeam, home, awayTeam, away
	}

	for _, bookmaker := range odds.Bookmakers {
		for _, market := range bookmaker.Markets {
			if market.Key != "h2h" {
				continue
			}
			found += 1
			h2h := market.Outcomes
			homeTeam = h2h[0].Name
			awayTeam = h2h[1].Name
			home += h2h[0].Price
			away += h2h[1].Price
		}
	}

	if found > 0 {
		home = home / float32(found)
		away = away / float32(found)
	}

	return homeTeam, home, awayTeam, away
}


// Odds
type OddsApiOutcome struct {
	Name string `json:"name"`
	Price float32 `json:"price"`
}

type OddsApiMarket struct {
	Key string `json:"key"`
	Outcomes []OddsApiOutcome `json:"outcomes"`
}

type OddsApiBookmaker struct {
	Key string `json:"key"`
	Markets []OddsApiMarket `json:"markets"`
}

type OddsApiFixture struct {
	CommenceTime time.Time `json:"commence_time"`
	Bookmakers []OddsApiBookmaker `json:"bookmakers"`
}
