package guestbook

import (
	"math"
	"net/http"
	"time"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

// Functions about creating match and calculating ELO ratings

// [START submit_match_result]
func submitMatchResult(w http.ResponseWriter, r *http.Request) {
	// [START new_context]
	ctx := r.Context()
	// [END new_context]

	keyWinner := datastore.Key{}
	keyLoser := datastore.Key{}
	winner := UserProfile{}
	loser := UserProfile{}
	exist := false
	var err error

	winnerName := r.FormValue("winner")
	loserName := r.FormValue("loser")

	if winnerName == loserName {
		http.Error(w, "Winner should not be the same as loser.",
			http.StatusBadRequest)
	}

	// Check winner is registered.
	exist, keyWinner, winner, err = existUser(ctx, winnerName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !exist {
		http.Error(w, "Winner has not registered", http.StatusBadRequest)
		return
	}

	// Check loser is registered.
	exist, keyLoser, loser, err = existUser(ctx, loserName)
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
	submitter := user.Current(ctx).String()
	note := r.FormValue("note")
	date := time.Now()
	match := createMatch(
		winner.Rating, loser.Rating,
		winner.Name, loser.Name,
		tournament, submitter, note, date)

	// Insert match entry
	key := datastore.NewIncompleteKey(ctx, "Match", guestbookKey(ctx))
	keyMatch := &datastore.Key{}
	keyMatch, err = datastore.Put(ctx, key, &match)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Try to update winner
	winner.Rating = match.WinnerRatingAfter
	winner.Wins++
	_, err = datastore.Put(ctx, &keyWinner, &winner)
	if err != nil {
		// Remove match entity as best-effort fallback.
		datastore.Delete(ctx, keyMatch)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Try to update loser rating.
	loser.Rating = match.LoserRatingAfter
	loser.Losses++
	_, err = datastore.Put(ctx, &keyLoser, &loser)
	if err != nil {
		// Remove match entity as best-effort fallback.
		datastore.Delete(ctx, keyMatch)
		// Change winner rating back.
		winner.Rating = match.WinnerRatingBefore
		winner.Wins--
		datastore.Put(ctx, &keyWinner, &winner)

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update latest match
	existLatestMatch = true
	latestMatch = match

	http.Redirect(w, r, "/add_match_result", http.StatusFound)
}

// Create a match for two players in default tournament
func createMatch(
	winnerOldRating float64, loserOldRating float64,
	winnerName string, loserName string,
	tournament string, submitter string, note string, date time.Time) Match {
	//Get new ELO value
	winnerNewRating, loserNewRating := newRatings(winnerOldRating, loserOldRating)

	match := Match{
		Tournament:         tournament,
		Submitter:          submitter,
		Winner:             winnerName,
		Loser:              loserName,
		WinnerRatingBefore: winnerOldRating,
		WinnerRatingAfter:  winnerNewRating,
		LoserRatingBefore:  loserOldRating,
		LoserRatingAfter:   loserNewRating,
		Expected:           winnerOldRating >= loserOldRating,
		Note:               note,
		Date:               date,
	}
	return match
}

// Get the new ratings of two players after a match
func newRatings(oldRatingW, oldRatingL float64) (float64, float64) {
	//Get new ELO value
	expectedScoreW := expectedScore(oldRatingW, oldRatingL)
	winnerNewRating := newElo(oldRatingW, expectedScoreW, 1.0)
	expectedScoreL := expectedScore(oldRatingL, oldRatingW)
	loserNewRating := newElo(oldRatingL, expectedScoreL, 0.0)
	return winnerNewRating, loserNewRating
}

// Expected score of elo_a in a match against elo_b
func expectedScore(eloA, eloB float64) float64 {
	return 1 / (1 + math.Pow(10, (eloB-eloA)/400))
}

// Get the new Elo rating.
func newElo(oldElo, expected, score float64) float64 {
	return oldElo + 32.0*(score-expected)
}
