package guestbook

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// insertFFAMatch inserts an FFAMatch object into datastore
func insertFFAMatch(ctx context.Context, match FFAMatch) error {
	key := datastore.NewIncompleteKey(ctx, "FFAMatch", guestbookKey(ctx))
	_, err := datastore.Put(ctx, key, &match)
	return err
}

// createFFAMatch creates an FFAMatch object based on input parameters
func createFFAMatch(
	tournamentID int64,
	players []int64,
	draws []bool,
	preGameUserStatsList []UserTournamentStats,
	postGameUserStatsList []UserTournamentStats,
	outcomeProbability float64,
	note string,
	submitter string,
	submissionTime time.Time) FFAMatch {

	ffaMatch := FFAMatch{
		TournamentID: tournamentID,
		Players:      players,
		Draws:        draws,

		OutcomeProbability: outcomeProbability,
		// Additional information
		Note:           note,
		Submitter:      submitter,
		SubmissionTime: submissionTime,
	}

	// Fill in pre-game stats
	ffaMatch.PreGameTrueSkillMu,
		ffaMatch.PreGameTrueSkillSigma,
		ffaMatch.PreGameTrueSkillRating = getMuSigmaRating(preGameUserStatsList)

	// Fill in post-game stats
	ffaMatch.PostGameTrueSkillMu,
		ffaMatch.PostGameTrueSkillSigma,
		ffaMatch.PostGameTrueSkillRating = getMuSigmaRating(postGameUserStatsList)

	return ffaMatch
}

func getMuSigmaRating(userStatsList []UserTournamentStats) ([]float64, []float64, []float64) {
	var mu, sigma, rating []float64
	for _, stats := range userStatsList {
		mu = append(mu, stats.TrueSkillMu)
		sigma = append(sigma, stats.TrueSkillSigma)
		rating = append(rating, stats.TrueSkillRating)
	}
	return mu, sigma, rating
}
