package guestbook

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func showTournaments(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("static", "tournaments.html"))
}

// showTournamentStats handles URL starting with /tournament/
func showTournamentStats(w http.ResponseWriter, r *http.Request) {
	tokens := strings.Split(r.URL.Path, "/")

	// URL is /tournament/<tournament_name>
	if len(tokens) == 3 {
		http.ServeFile(w, r, path.Join("static", "tournament_stats.html"))
		return
	}

	// URL is /tournament/<tournament_name>/<action>
	if len(tokens) == 4 {
		action := tokens[3]

		if action == "add_ffa_match_result" {
			http.ServeFile(w, r, path.Join("static", "add_ffa_match_result.html"))
			return
		}

		if action == "add_tta_match_result" {
			http.ServeFile(w, r, path.Join("static", "add_tta_match_result.html"))
			return
		}

		http.Error(w, "action: "+action+" is not supported", http.StatusBadRequest)

	}

	http.Error(w, "URL must be in the form of /tournament/<name> or /tournament/<name>/<action>", http.StatusBadRequest)
	return
}

// [START submit_tournament]
func submitTournament(w http.ResponseWriter, r *http.Request) {

	// Check valid name
	name := r.FormValue("name")

	re, _ := regexp.Compile("^[A-Za-z0-9_]{3,20}$")

	isValid := re.MatchString(name)
	if !isValid {
		http.Error(w, "Not a valid name for tournament", http.StatusBadRequest)
		return
	}

	ctx := appengine.NewContext(r)
	err := datastore.RunInTransaction(ctx,
		func(ctx context.Context) error {
			exist, _, _, err := findExistingTournament(ctx, name)
			if err != nil {
				return err
			}

			if exist {
				return errors.New("tournament name already exists")
			}

			t := Tournament{
				Name: name,
			}

			// [END getall]
			key := datastore.NewIncompleteKey(ctx, "Tournament", guestbookKey(ctx))
			_, err = datastore.Put(ctx, key, &t)

			return err
		},
		nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/tournament", http.StatusFound)
	return
}

func findExistingTournament(c context.Context, name string) (bool, *datastore.Key, Tournament, error) {
	q := datastore.NewQuery("Tournament").Ancestor(guestbookKey(c)).Filter("Name =", name).Limit(1)
	var tournaments []Tournament
	keys, err := q.GetAll(c, &tournaments)
	if err != nil {
		return false, nil, Tournament{}, err
	}
	if len(tournaments) != 0 {
		return true, keys[0], tournaments[0], nil
	}
	return false, nil, Tournament{}, nil
}

func findExistingTournamentKey(ctx context.Context, tournamentName string) (*datastore.Key, error) {
	exist, tournamentKey, _, err := findExistingTournament(ctx, tournamentName)

	if err != nil {
		return nil, err
	}

	if !exist {
		return nil, fmt.Errorf("tournament %s does not exist", tournamentName)
	}

	return tournamentKey, nil
}

func readTournaments(ctx context.Context) ([]Tournament, error) {
	query := datastore.NewQuery("Tournament").Ancestor(guestbookKey(ctx))
	var tournaments []Tournament
	if _, err := query.GetAll(ctx, &tournaments); err != nil {
		return nil, err
	}

	return tournaments, nil
}

func requestTournaments(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	tournaments, err := readTournaments(ctx)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, err := json.Marshal(tournaments)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func requestDetailMatchResults(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	tournamentName := r.FormValue("tournament")
	if tournamentName == "" {
		tournamentName = "Default"
	}

	tournamentKey, err := findExistingTournamentKey(ctx, tournamentName)
	tournamentID := tournamentKey.IntID()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user tournament stats
	statsList, err := readAllUserStatsForTournament(ctx, tournamentID)
	if err != nil {
		http.Error(w, "Failed to read tournament stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	users := make([]UserProfile, len(statsList))
	for i, stats := range statsList {
		users[i], err = readUserProfile(ctx, stats.UserID)
		if err != nil {
			http.Error(w, "Failed to read user profile: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Get all FFAMatches
	query := datastore.NewQuery("FFAMatch").Ancestor(guestbookKey(ctx)).Filter("TournamentID =", tournamentID)
	var matches []FFAMatch
	if _, err := query.GetAll(ctx, &matches); err != nil {
		http.Error(w, "Failed to read FFAMatches: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// User IDs to index map
	idMap := make(map[int64]int)
	for index, stats := range statsList {
		idMap[stats.UserID] = index
	}

	// Set usernames
	usernames := make([]string, len(users))
	for i, u := range users {
		usernames[i] = u.Name
	}

	// Set resultTable
	resultTable := make([][]DetailMatchResultEntry, len(users))
	for i := range resultTable {
		resultTable[i] = make([]DetailMatchResultEntry, len(users))
		for j := range resultTable[i] {
			resultTable[i][j].Wins = 0
			resultTable[i][j].Losses = 0
		}
	}

	// Generate wins and losses in resultTable
	for _, match := range matches {
		playerIndex := make([]int, len(match.Players))
		for i, playerID := range match.Players {
			index, exist := idMap[playerID]
			if !exist {
				http.Error(w,
					fmt.Sprintf("User ID %d does not exist, cannot process FFAMatch: %+v", playerID, match),
					http.StatusInternalServerError)
				return
			}
			playerIndex[i] = index
		}
		for i := range playerIndex {
			for j := i + 1; j < len(playerIndex); j++ {
				winner := playerIndex[i]
				loser := playerIndex[j]
				resultTable[winner][loser].Wins++
				resultTable[loser][winner].Losses++
			}
		}
	}

	// color-code wins and losses
	for i := range resultTable {
		for j := range resultTable[i] {
			resultTable[i][j].Color = getColor(users[i], users[j], resultTable[i][j].Wins, resultTable[i][j].Losses)
		}
	}

	// Return JSON
	matchData := MatchData{
		Usernames:   usernames,
		ResultTable: resultTable,
	}

	js, err := json.Marshal(matchData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
