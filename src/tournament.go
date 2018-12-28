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

	ctx := r.Context()
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
	ctx := r.Context()
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
