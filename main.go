package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

var (
	PlayerData map[string]Player
	TeamData   map[string]*Team
	Flag       int
)

type ResponseData struct {
	Data Data `json:"data"`
}

type Data struct {
	Team Team `json:"team"`
}

type Team struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Players []Player `json:"players"`
}

type Player struct {
	ID           string   `json:"id"`
	Country      string   `json:"country"`
	FirstName    string   `json:"firstName"`
	LastName     string   `json:"lastName"`
	Name         string   `json:"name"`
	Position     string   `json:"position"`
	Number       int      `json:"number"`
	Age          string   `json:"age"`
	BirthDate    string   `json:"birthDate"`
	Height       int      `json:"height"`
	Weight       int      `json:"weight"`
	ThumbnailSRC string   `json:"thumbnailSrc"`
	Teams        []string `json:"-"`
}

func main() {
	teamNames := []string{
		"Germany",
		"England",
		"France",
		"Spain",
		"Manchester United",
		"Arsenal",
		"Chelsea",
		"Barcelona",
		"Real Madrid",
		"Bayern Munich",
	}

	TeamData = map[string]*Team{}
	PlayerData = map[string]Player{}

	for _, name := range teamNames {
		TeamData[name] = &Team{Name: name}
	}

	Flag = 0
	// Max 100, all team data found < 100, max id is 9999
	// sequential request avoid blocked from host
	for i := 1; i <= 100; i++ {
		if Flag >= len(teamNames) {
			// All teams ids found, use this id for the next request
			fmt.Println("All data found", i)
			break
		}
		if err := players(i); err != nil {
			fmt.Println(err)
			continue
		}
	}

	show(PlayerData)
}

func players(id int) error {
	url := fmt.Sprintf("https://vintagemonster.onefootball.com/api/teams/en/%d.json", id)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Sprintf("status code is %d, id %d\n", resp.StatusCode, id)
		return errors.New(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	responseData := ResponseData{}

	if err := json.Unmarshal(body, &responseData); err != nil {
		return err
	}

	teamData := TeamData[responseData.Data.Team.Name]
	if teamData == nil {
		return nil
	}

	Flag++
	fmt.Printf("%s found on id %d\n", responseData.Data.Team.Name, id)

	// sequential map player to PlayerData
	for _, player := range responseData.Data.Team.Players {
		playerData := PlayerData[player.Name]
		player.Teams = append(playerData.Teams, responseData.Data.Team.Name)
		PlayerData[player.Name] = player
	}

	return nil
}

func show(players map[string]Player) {
	fmt.Println(len(players))
	keys := make([]string, 0, len(players))
	for key := range players {
		keys = append(keys, key)
	}

	cl := collate.New(language.English)
	cl.SortStrings(keys)

	for i, key := range keys {
		player := players[key]
		fmt.Printf("%d. %s; %s; %s\n", i+1, player.Name, player.Age, strings.Join(player.Teams, ", "))
	}
}
