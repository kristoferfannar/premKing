package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kristo-og-logi/premKing/server/initializers"
	"github.com/kristo-og-logi/premKing/server/models"
)

// SeedTeamsFromJSON reads a JSON file of teams (matching models.Team's
// json tags: id, name, shortName, logo) and writes each one to the DB,
// skipping any that already exist by ID.
func SeedTeamsFromJSON(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("error reading %s: %s\n", path, err.Error())
		os.Exit(1)
	}

	var teams []models.Team
	if err := json.Unmarshal(data, &teams); err != nil {
		fmt.Printf("error parsing %s: %s\n", path, err.Error())
		os.Exit(1)
	}

	added := 0
	for _, team := range teams {
		result := initializers.DB.Where(models.Team{ID: team.ID}).FirstOrCreate(&team)
		if result.Error != nil {
			fmt.Printf("error adding team %s: %s\n", team.Name, result.Error.Error())
			continue
		}

		if result.RowsAffected > 0 {
			added++
			fmt.Printf("added %s\n", team.Name)
		} else {
			fmt.Printf("already exists: %s\n", team.Name)
		}
	}

	fmt.Printf("done: added %d/%d teams\n", added, len(teams))
}
