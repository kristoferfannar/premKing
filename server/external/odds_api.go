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
	return fmt.Sprintf("https://api.the-odds-api.com/v4/sports/soccer_epl/odds?bookmakers=betsson&apiKey=%s", apiKey)
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
