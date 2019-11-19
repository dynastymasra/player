package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

const FootballTeam = "https://api.onefootball.com/score-one/api/teams/en/%d.json"

var (
	PlayerData map[string]Player
	TeamData   map[string]*Team
	//Flag       int
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

	job := make(chan int)
	result := make(chan string, len(teamNames))

	TeamData = map[string]*Team{}
	PlayerData = map[string]Player{}

	for _, name := range teamNames {
		TeamData[name] = &Team{Name: name}
	}

	maxIDs := os.Getenv("MAX_IDS")
	maxID, err := strconv.Atoi(maxIDs)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < 10; i++ {
		go func(id int, j chan int, r chan string) {
			players(id, j, r)
		}(i, job, result)
	}

	for i := 1; i <= maxID; i++ {
		job <- i
	}
	close(job)

	for i := 0; i < len(teamNames); i++ {
		res := <-result
		fmt.Printf("Received %s\n", res)
	}

	show(PlayerData)
}

func players(id int, job chan int, result chan string) {
	for j := range job {
		fmt.Printf("Handle request id %d with worker %d\n", j, id)
		url := fmt.Sprintf(FootballTeam, j)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			err := fmt.Sprintf("status code is %d, id %d\n", resp.StatusCode, j)
			fmt.Println(err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			continue
		}

		responseData := ResponseData{}

		if err := json.Unmarshal(body, &responseData); err != nil {
			fmt.Println(err)
			continue
		}

		teamData := TeamData[responseData.Data.Team.Name]
		if teamData == nil {
			continue
		}

		fmt.Printf("%s found on id %d\n", responseData.Data.Team.Name, id)

		// map player to PlayerData
		for _, player := range responseData.Data.Team.Players {
			playerData := PlayerData[player.Name]
			player.Teams = append(playerData.Teams, responseData.Data.Team.Name)
			PlayerData[player.Name] = player
		}

		result <- teamData.Name
	}
}

func show(players map[string]Player) {
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
