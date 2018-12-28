package guestbook

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"path"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

// HandleAddFfaMatchResult handles http request to tournament/name/add_ffa_match_result
func HandleAddFfaMatchResult(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("static", "add_ffa_match_result.html"))
}

// MathcResult1v1 represents a match result between two players
// Both winner and loser fields are array index, not player name
type MathcResult1v1 struct {
	winner int
	loser  int

	// Values in below are used only for sorting

	// Indicating how far away this pair's center is from the "middle" of ranking.
	centerToMiddleDistance float64
	// Indicating how wide this pair is. We'd like to process narrow pairs first
	// if two pairs have the same centerToMiddleDistance.
	radius int
}

// New1v1MatchResult creates a MathcResult1v1 object
func New1v1MatchResult(winner int, loser int, middle float64) MathcResult1v1 {
	return MathcResult1v1{
		winner:                 winner,
		loser:                  loser,
		centerToMiddleDistance: math.Abs(float64(winner+loser)/2 - middle),
		radius:                 loser - winner,
	}
}

// Generate1v1MatchResults generates 1v1 match reuslts for a multi-player FFA
// game.
func Generate1v1MatchResults(numPlayers int) []MathcResult1v1 {
	// Our current logic for FFA: Emulate the elo results by generating emulated 1v1
	// game results from each player against other players within their +-2
	// ranking range, generating total of O(2N) 1v1 games.

	// Example: in a 4-player FFA game, we'll emulate following 5 1v1 games:
	//  *  1st > 2nd
	//  *  1st > 3rd
	//  *  2nd > 3rd
	//  *  2nd > 4th
	//  *  3rd > 4th

	// Additionally, elo rating update is not commutative --- the order of
	// update will affect the final results. To make the final ratings "spreads"
	// more evenly for both positive and negative changes, we'll start by
	// updating ratings from the middle of the ranking, then process the updates
	// near the top and bottom. In the 4-player FFA game in above, the actual
	// update sequence will be:
	//  * 2nd > 3rd (index 1, 2)
	//  * 1st > 3rd (index 0, 2)
	//  * 2nd > 4th (index 1, 3)
	//  * 1st > 2nd (index 0, 1)
	//  * 3rd > 4th (index 2, 3)

	// middle represent the "middle" of possible indexes under current number of
	// players. For example, in 4-player game, possible indices are [0,1,2,3],
	// and the middle will be 1.5.
	var middle = (float64(numPlayers) - 1) / 2

	// generate a slice containing all pairs that we need
	var results []MathcResult1v1
	for winner := 0; winner < numPlayers; winner++ {
		for loser := winner + 1; loser < numPlayers && loser-winner <= 2; loser++ {
			results = append(results, New1v1MatchResult(winner, loser, middle))
		}
	}

	sort.Slice(results, func(i, j int) bool {
		m1, m2 := &results[i], &results[j]

		// if both distance and radius is the same, do smaller index first
		if m1.centerToMiddleDistance == m2.centerToMiddleDistance && m1.radius == m2.radius {
			return m1.winner < m2.winner
		}

		// if distance is the same but radius is different do smaller radius first
		if m1.centerToMiddleDistance == m2.centerToMiddleDistance {
			return m1.radius < m2.radius
		}

		// do the pairs closer to middle first
		return m1.centerToMiddleDistance < m2.centerToMiddleDistance
	})

	return results
}

// FfaMatchResult represents an FFA game match result, which will be in json
// format within the http post request this should match the format in
// add_ffa_match_result.js
type FfaMatchResult struct {
	Tournament string
	Ranking    []string // player name from first place to last place
}

// HandleSubmitFfaMatchResult handles http post request to submit a FFA Match
func HandleSubmitFfaMatchResult(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	decoder := json.NewDecoder(req.Body)
	var matchResult FfaMatchResult
	err := decoder.Decode(&matchResult)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("matchResult: %+v\n", matchResult)

	userStatsKeys, userStats, err := readOrCreateUserTournamentStats(ctx, matchResult.Tournament, matchResult.Ranking)

	// Create match entry
	submitter := user.Current(ctx).String()
	note := fmt.Sprintf(
		"FFA game ranking: [%s]",
		strings.Join(matchResult.Ranking, ", "))

	// Do all updates within a transaction
	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		// Process all emulated 1v1 results
		for _, generated1v1 := range Generate1v1MatchResults(len(matchResult.Ranking)) {
			winner := &userStats[generated1v1.winner]
			loser := &userStats[generated1v1.loser]
			winnerName := matchResult.Ranking[generated1v1.winner]
			loserName := matchResult.Ranking[generated1v1.loser]

			match := createMatch(
				winner.Rating, loser.Rating,
				winnerName, loserName,
				matchResult.Tournament, submitter, note, time.Now())

			// Update wins, losses, and rating
			winner.Wins++
			winner.Rating = match.WinnerRatingAfter
			loser.Losses++
			loser.Rating = match.LoserRatingAfter

			// Insert match entry
			key := datastore.NewIncompleteKey(ctx, "Match", guestbookKey(ctx))
			_, err = datastore.Put(ctx, key, &match)
			if err != nil {
				return err
			}
		}

		// First player should get ++ for FFA Win
		userStats[0].FFAWins++

		// Now update user stats for each user
		for i, statsKey := range userStatsKeys {
			if _, err = datastore.Put(ctx, statsKey, &userStats[i]); err != nil {
				return err
			}
		}

		return nil
	}, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// return nothing if successful
}

// readOrCreateUserTournamentStats will try to read users' stats for a given
// tournament, if the specified user have no record in the specified tournament,
// a new record with default values will be created.
func readOrCreateUserTournamentStats(
	ctx context.Context,
	tournamentName string,
	userNames []string) ([]*datastore.Key, []UserTournamentStats, error) {

	tournamentKey, err := findExistingTournamentKey(ctx, tournamentName)

	if err != nil {
		return nil, nil, err
	}

	userKeys, err := findUserKeys(ctx, userNames)

	if err != nil {
		return nil, nil, err
	}

	userStats := make([]UserTournamentStats, len(userNames))
	userStatsKeys := make([]*datastore.Key, len(userNames))

	// Read user stats or create initial values
	for i, userKey := range userKeys {
		userStatsKeys[i], userStats[i], err = readOrCreateStatsWithID(ctx, tournamentKey.IntID(), userKey.IntID())
		if err != nil {
			return nil, nil, err
		}
	}

	return userStatsKeys, userStats, nil
}
