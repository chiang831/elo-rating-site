package guestbook

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	trueskill "github.com/mafredri/go-trueskill"
)

// Constants for initial stats.
const (
	InitialFFAWins = 0
	InitialWins    = 0
	InitialLosses  = 0

	// Initial Elo Rating
	InitialRating = 1200

	// Initial Trueskill Parameter
	InitialTrueSkillMu     = 25.0
	InitialTrueSkillSigma  = 25.0 / 3.0
	InitialTrueSkillRating = 0.0
	DrawProbability        = 5.0 // setting draw probability to 5%, the library uses 0.0-100.0 instead of 0.0-1.0
)

// createTrueSkillConfig creates a TrueSkill config object with default values
func createTrueSkillConfig() (trueskill.Config, error) {
	drawProbabilityOption, err := trueskill.DrawProbability(DrawProbability)
	if err != nil {
		return trueskill.New(), err
	}

	return trueskill.New(
		trueskill.Mu(InitialTrueSkillMu),
		trueskill.Sigma(InitialTrueSkillSigma),
		drawProbabilityOption), nil
}

// findExistingUser tries to find exiting user in the database that matches the
// given username
func findExistingUser(ctx context.Context, userName string) (bool, datastore.Key, UserProfile, error) {
	q := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(ctx)).Filter("Name =", userName).Limit(1)
	var userProfiles []UserProfile
	keys, err := q.GetAll(ctx, &userProfiles)
	if err != nil {
		return false, datastore.Key{}, UserProfile{}, err
	}
	if len(userProfiles) != 0 {
		return true, *keys[0], userProfiles[0], nil
	}
	return false, datastore.Key{}, UserProfile{}, nil
}

// findUserKey finds the datastore key for given user name, and returns an error
// if the user name does not exist
func findUserKey(ctx context.Context, userName string) (datastore.Key, error) {
	exist, key, _, err := findExistingUser(ctx, userName)
	if err != nil {
		return datastore.Key{}, err
	}
	if !exist {
		return datastore.Key{}, fmt.Errorf("username %s does not exist", userName)
	}
	return key, nil
}

// findUserKeys finds the datastore keys for given user names, and returns an
// error if any of the user name does not exist
func findUserKeys(ctx context.Context, userNames []string) ([]datastore.Key, error) {
	keys := make([]datastore.Key, len(userNames))
	var err error
	for i, userName := range userNames {
		keys[i], err = findUserKey(ctx, userName)
		if err != nil {
			return nil, err
		}
	}
	return keys, nil
}

// readStatsWithKey reads an user's stats for a given tournament, using
// datastore ID instead of string names. The first returned value indicates
// whether the stats exists.
func readStatsWithID(ctx context.Context, tournamentID int64, userID int64) (
	bool, *datastore.Key, UserTournamentStats, error) {
	q := datastore.NewQuery("UserTournamentStats").Ancestor(guestbookKey(ctx)).
		Filter("TournamentID =", tournamentID).
		Filter("UserID =", userID).
		Limit(1)
	var stats []UserTournamentStats
	keys, err := q.GetAll(ctx, &stats)
	if err != nil {
		return false, nil, UserTournamentStats{}, err
	}
	if len(stats) == 0 {
		return false, nil, UserTournamentStats{}, err
	}
	return true, keys[0], stats[0], nil
}

func calculateTrueSkillRating(mu float64, sigma float64) float64 {
	return mu - 3*sigma
}

func createInitialUserStats(tournamentID int64, userID int64) UserTournamentStats {
	return UserTournamentStats{
		TournamentID:    tournamentID,
		UserID:          userID,
		FFAWins:         InitialFFAWins,
		Wins:            InitialWins,
		Losses:          InitialLosses,
		Rating:          InitialRating,
		TrueSkillMu:     InitialTrueSkillMu,
		TrueSkillSigma:  InitialTrueSkillSigma,
		TrueSkillRating: InitialTrueSkillRating,
	}
}

// readOrCreateStatsWithID reads user stats with given IDs, and will create a
// new default entry if the records does not exist yet.
func readOrCreateStatsWithID(ctx context.Context, tournamentID int64, userID int64) (
	*datastore.Key, UserTournamentStats, error) {
	exist, key, stats, err := readStatsWithID(ctx, tournamentID, userID)
	if err != nil {
		return nil, UserTournamentStats{}, err
	}

	if exist {
		return key, stats, nil
	}

	exist, key, stats, err = readStatsWithID(ctx, tournamentID, userID)
	if exist {
		// there are multiple queries that were trying to create this entry,
		// and it's now created by some other thread. We can now return stats.
		return nil, UserTournamentStats{}, nil
	}

	incompleteKey := datastore.NewIncompleteKey(ctx, "UserTournamentStats", guestbookKey(ctx))

	stats = createInitialUserStats(tournamentID, userID)

	key, err = datastore.Put(ctx, incompleteKey, &stats)

	if err != nil {
		return nil, UserTournamentStats{}, err
	}

	return key, stats, nil
}

func readAllUserStatsForTournament(ctx context.Context, tournamentID int64) ([]UserTournamentStats, error) {
	query := datastore.NewQuery("UserTournamentStats").Ancestor(guestbookKey(ctx)).
		Filter("TournamentID = ", tournamentID).
		Order("-TrueSkillRating")
	var statsList []UserTournamentStats
	if _, err := query.GetAll(ctx, &statsList); err != nil {
		return nil, err
	}

	return statsList, nil
}

func requestTournamentStats(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	tournamentName := r.FormValue("tournament")
	if tournamentName == "" {
		http.Error(w, "tournament parameter is missing", http.StatusBadRequest)
		return
	}

	tournamentKey, err := findExistingTournamentKey(ctx, tournamentName)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user tournament stats
	statsList, err := readAllUserStatsForTournament(ctx, tournamentKey.IntID())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create public user profile
	userProfileToShows := make([]UserProfileToShow, len(statsList))
	for i, stats := range statsList {
		profile, err := readUserProfile(ctx, stats.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get badges
		userProfileToShows[i] = UserProfileToShow{
			Name:            profile.Name,
			Rating:          stats.Rating,
			TrueSkillMu:     stats.TrueSkillMu,
			TrueSkillSigma:  stats.TrueSkillSigma,
			TrueSkillRating: stats.TrueSkillRating,
			FFAWins:         stats.FFAWins,
			Wins:            stats.Wins,
			Losses:          stats.Losses,
			Badges:          getUserBadges(ctx, profile.Name),
		}
	}

	js, errJs := json.Marshal(userProfileToShows)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}
