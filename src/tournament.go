package guestbook

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func showTournaments(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("static", "tournaments.html"))
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
			exist, _, _, err := isExistingTournament(ctx, name)
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

func isExistingTournament(c context.Context, name string) (bool, datastore.Key, Tournament, error) {
	q := datastore.NewQuery("Tournament").Ancestor(guestbookKey(c)).Filter("Name =", name)
	var tournaments []Tournament
	keys, err := q.GetAll(c, &tournaments)
	if err != nil {
		return false, datastore.Key{}, Tournament{}, err
	}
	if len(tournaments) != 0 {
		return true, *keys[0], tournaments[0], nil
	}
	return false, datastore.Key{}, Tournament{}, nil
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

func validateTournamentName(ctx context.Context, name string) error {
	if name == "" {
		return errors.New("tournament name cannot be empty")
	}

	exist, _, _, err := isExistingTournament(ctx, name)
	if err != nil {
		return err
	}
	if !exist {
		return fmt.Errorf("tournament name %s does not exist", name)
	}
	return nil
}
