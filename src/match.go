package guestbook

import (
	"math"
	"net/http"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

// Functions about creating match and calculating ELO ratings

// [START submit_match_result]
func submitMatchResult(w http.ResponseWriter, r *http.Request) {
	// [START new_context]
	c := appengine.NewContext(r)
	// [END new_context]

	keyWinner := datastore.Key{}
	keyLoser := datastore.Key{}
	winner := UserProfile{}
	loser := UserProfile{}
	exist := false
	var err error

	winner_name := r.FormValue("winner")
	loser_name := r.FormValue("loser")

	if winner_name == loser_name {
		http.Error(w, "Winner should not be the same as loser.",
			http.StatusBadRequest)
	}

	// Check winner is registered.
	exist, keyWinner, winner, err = existUser(c, winner_name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exist {
		http.Error(w, "Winner has not registered", http.StatusBadRequest)
		return
	}

	// Check loser is registered.
	exist, keyLoser, loser, err = existUser(c, loser_name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exist {
		http.Error(w, "Loser has not registered", http.StatusBadRequest)
		return
	}

	// Create match entry
	tournament := "Default"
	submitter := user.Current(c).String()
	note := r.FormValue("note")
	date := time.Now()
	match := createMatch(winner, loser, tournament, submitter, note, date)

	// Insert match entry
	key := datastore.NewIncompleteKey(c, "Match", guestbookKey(c))
	keyMatch := &datastore.Key{}
	keyMatch, err = datastore.Put(c, key, &match)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Try to update winner
	winner.Rating = match.WinnerRatingAfter
	winner.Wins += 1
	_, err = datastore.Put(c, &keyWinner, &winner)
	if err != nil {
		// Remove match entity as best-effort fallback.
		datastore.Delete(c, keyMatch)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Try to update loser rating.
	loser.Rating = match.LoserRatingAfter
	loser.Losses += 1
	_, err = datastore.Put(c, &keyLoser, &loser)
	if err != nil {
		// Remove match entity as best-effort fallback.
		datastore.Delete(c, keyMatch)
		// Change winner rating back.
		winner.Rating = match.WinnerRatingBefore
		winner.Wins -= 1
		datastore.Put(c, &keyWinner, &winner)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update latest match
	existLatestMatch = true
	latestMatch = match

	http.Redirect(w, r, "/add_match_result", http.StatusFound)
}

// Create a match for two players
func createMatch(winner UserProfile, loser UserProfile, tournament string, submitter string, note string, date time.Time) Match {
	oldRatingW := winner.Rating
	oldRatingL := loser.Rating

	//Get new ELO value
	newRatingW, newRatingL := newRatings(oldRatingW, oldRatingL)

	match := Match{
		Tournament:         tournament,
		Submitter:          submitter,
		Winner:             winner.Name,
		Loser:              loser.Name,
		WinnerRatingBefore: oldRatingW,
		WinnerRatingAfter:  newRatingW,
		LoserRatingBefore:  oldRatingL,
		LoserRatingAfter:   newRatingL,
		Expected:           oldRatingW >= oldRatingL,
		Note:               note,
		Date:               date,
	}
	return match
}

// Get the new ratings of two players after a match
func newRatings(oldRatingW, oldRatingL float64) (float64, float64) {
	//Get new ELO value
	expectedScoreW := expectedScore(oldRatingW, oldRatingL)
	newRatingW := newElo(oldRatingW, expectedScoreW, 1.0)
	expectedScoreL := expectedScore(oldRatingL, oldRatingW)
	newRatingL := newElo(oldRatingL, expectedScoreL, 0.0)
	return newRatingW, newRatingL
}

// Expected score of elo_a in a match against elo_b
func expectedScore(elo_a, elo_b float64) float64 {
	return 1 / (1 + math.Pow(10, (elo_b-elo_a)/400))
}

// Get the new Elo rating.
func newElo(old_elo, expected, score float64) float64 {
	return old_elo + 32.0*(score-expected)
}
