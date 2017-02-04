package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// put me in a cache or something
var gs GameStates

func main() {
	gs = make(GameStates)
	http.HandleFunc("/filter/", filterHandler)

	log.Fatal(http.ListenAndServe(":8082", nil))
}

func filterHandler(w http.ResponseWriter, r *http.Request) {
	gameCode := r.URL.Path[len("/filter/"):]
	log.Printf("filtering %s\n", gameCode)

	var pbp PlayByPlayGame
	dec := json.NewDecoder(r.Body)
	for dec.More() {

		err := dec.Decode(&pbp)
		if err != nil {
			log.Printf("error decoding game: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, `{"error":"server error occurred"}`)
			return
		}
	}

	gs.FilterNewPlays(&pbp)

	var msg string
	if len(pbp.Plays) > 0 {
		msg = "new plays"
	} else {
		msg = "no new plays"
	}
	log.Printf("%s: %+v\n", msg, pbp)

	enc := json.NewEncoder(w)
	err := enc.Encode(&pbp)
	if err != nil {
		log.Printf("error encoding game: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, `{"error":"server error occurred"}`)
		return
	}

}

type GameStates map[string]PlayByPlayGame

// FilterNewPlays updates the cache holding all "done" plays
// and updates the input game to only hold "new" plays
func (s GameStates) FilterNewPlays(pbp *PlayByPlayGame) {
	// for now, just compare lengths
	cachedGame := s[pbp.GameCode()]
	donePlays := cachedGame.Plays

	newPlays := pbp.Plays[len(donePlays):] // careful about index here

	// update cache
	cachedGame.Plays = pbp.Plays
	s[pbp.GameCode()] = cachedGame

	// swap current game
	pbp.Plays = newPlays
}

type PlayByPlayGame struct {
	Game
	Plays []Play
}

type Game struct {
	Id           string    `json:"gameId"`
	StartTime    time.Time `json:"startTimeUTC"`
	VisitingTeam Team      `json:"vTeam"`
	HomeTeam     Team      `json:"hTeam"`
	Period       Period    `json:"period"`
	Active       bool      `json:"isGameActivated"`
}

func (g Game) GameCode() string {
	return fmt.Sprintf("%s%s", g.VisitingTeam.TriCode, g.HomeTeam.TriCode)
}

type Play struct {
	Clock            string        `json:"clock"`
	Description      string        `json:"description"`
	PersonId         string        `json:"personId"`
	TeamId           string        `json:"teamId"`
	VistingTeamScore string        `json:"vTeamScore"`
	HomeTeamScore    string        `json:"hTeamScore"`
	IsScoreChange    bool          `json:"isScoreChange"`
	Formatted        FormattedPlay `json:"formatted"`
}

type Team struct {
	Id      string `json:"teamId"`
	TriCode string `json:"triCode"`
}

type Period struct {
	Current int
}

type FormattedPlay struct {
	Description string `json:"description"`
}
