package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type GameData struct {
	Games []struct {
		Game Game `json:"game"`
	} `json:"games"`
}

type Game struct {
	GameID           string `json:"gameID"`
	Away             Team   `json:"away"`
	FinalMessage     string `json:"finalMessage"`
	BracketRound     string `json:"bracketRound"`
	Title            string `json:"title"`
	ContestName      string `json:"contestName"`
	URL              string `json:"url"`
	Network          string `json:"network"`
	Home             Team   `json:"home"`
	LiveVideoEnabled bool   `json:"liveVideoEnabled"`
	StartTime        string `json:"startTime"`
	StartTimeEpoch   string `json:"startTimeEpoch"`
	BracketID        string `json:"bracketId"`
	GameState        string `json:"gameState"`
	StartDate        string `json:"startDate"`
	CurrentPeriod    string `json:"currentPeriod"`
	VideoState       string `json:"videoState"`
	BracketRegion    string `json:"bracketRegion"`
	ContestClock     string `json:"contestClock"`
}

type Team struct {
	Score       string       `json:"score"`
	Names       TeamNames    `json:"names"`
	Winner      bool         `json:"winner"`
	Seed        string       `json:"seed"`
	Description string       `json:"description"`
	Rank        string       `json:"rank"`
	Conferences []Conference `json:"conferences"`
}

type TeamNames struct {
	Char6 string `json:"char6"`
	Short string `json:"short"`
	Seo   string `json:"seo"`
	Full  string `json:"full"`
}

type Conference struct {
	ConferenceName string `json:"conferenceName"`
	ConferenceSeo  string `json:"conferenceSeo"`
}

const scoresEndpoint = "https://ncaa-api.henrygd.me/scoreboard/football/fbs/"

func getGameData(endpoint string) (*GameData, error) {
	resp, err := http.Get(scoresEndpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var gameData GameData
	err = json.Unmarshal(body, &gameData)
	if err != nil {
		return nil, err
	}

	return &gameData, nil
}

func selectGame(data []Game, main *tview.Table) func(index int, mainText string, secondaryText string, shortcut rune) {
	return func(index int, mainText string, secondaryText string, shortcut rune) {
		main.SetCell(0, 0, tview.NewTableCell("Team").
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter))
		main.SetCell(0, 1, tview.NewTableCell("Score").
			SetTextColor(tcell.ColorYellow).
			SetAlign(tview.AlignCenter))

		home := data[index].Home
		var homeRank string
		if home.Rank != "" {
			homeRank = fmt.Sprintf("(%s) ", home.Rank)
		}
		main.SetCell(1, 0, tview.NewTableCell(fmt.Sprintf("%s%s %s", homeRank, home.Names.Short, home.Description)).
			SetAlign(tview.AlignLeft))
		main.SetCell(1, 1, tview.NewTableCell(home.Score).
			SetAlign(tview.AlignLeft))

		away := data[index].Away
		var awayRank string
		if away.Rank != "" {
			awayRank = fmt.Sprintf("(%s) ", away.Rank)
		}
		main.SetCell(2, 0, tview.NewTableCell(fmt.Sprintf("%s%s %s", awayRank, data[index].Away.Names.Short, home.Description)).
			SetAlign(tview.AlignLeft))
		main.SetCell(2, 1, tview.NewTableCell(away.Score).
			SetAlign(tview.AlignLeft))
	}
}

func main() {
	data, err := getGameData(scoresEndpoint)
	if err != nil {
		log.Fatalf("failed to get game data: %v", err)
	}

	menu := tview.NewList()
	menu.SetBorder(true).
		SetTitle("Games")

	main := tview.NewTable().
		SetBorders(true)
	main.SetBorder(true).
		SetTitle("Stats")

	var displayGames []Game

	for _, g := range data.Games {
		if g.Game.GameState == "live" {
			displayGames = append(displayGames, g.Game)
			displayName := fmt.Sprintf("%s @ %s", g.Game.Away.Names.Short, g.Game.Home.Names.Short)
			menu.AddItem(displayName, fmt.Sprintf("%s Quarter, %s remaining", g.Game.CurrentPeriod, g.Game.ContestClock), 0, nil).
				SetChangedFunc(selectGame(displayGames, main)).SetSelectedFunc(selectGame(displayGames, main))
		}
	}

	grid := tview.NewGrid().
		SetColumns(-1, -3)

	// Layout for screens wider than 100 cells.
	grid.AddItem(menu, 0, 0, 1, 1, 0, 100, false).
		AddItem(main, 0, 1, 1, 1, 0, 100, false)

	// Ensure first element is selected
	menu.SetCurrentItem(1).SetCurrentItem(0)

	if err := tview.NewApplication().SetRoot(grid, true).SetFocus(menu).Run(); err != nil {
		log.Fatalf("failed to launch: %v", err)
	}

}
