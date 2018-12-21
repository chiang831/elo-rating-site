package guestbook

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"

	"google.golang.org/appengine"
)

func showAddFfaMatchResult(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("static", "add_ffa_match_result.html"))
}

// FfaMatchResult represents an FFA game match result, which will be in json format within the http post request
// this should match the format in add_ffa_match_result.js
type FfaMatchResult struct {
	tournament string
	ranking    []string // player name from first place to last place
}

func submitFfaMatchResult(w http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)

	decoder := json.NewDecoder(req.Body)
	var matchResult FfaMatchResult
	err := decoder.Decode(&matchResult)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = validateTournamentName(ctx, matchResult.tournament)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// ranking is an array containing player's names, from first place to last place
	// TODO: support ties

	// Currently this is dummy code to verify json parsing only
	// TODO: Change to real implementation
	io.WriteString(w, fmt.Sprintf("Tournament name: %s\n", matchResult.tournament))
	for _, playerName := range matchResult.ranking {
		io.WriteString(w, fmt.Sprintf("player: %s\n", playerName))
	}
}
