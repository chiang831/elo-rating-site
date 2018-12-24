package guestbook

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

// Constants for initial stats.
const (
	InitialFFAWins = 0
	InitialWins    = 0
	InitialLosses  = 0
	InitialRating  = 1200
)

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
	bool, UserTournamentStats, error) {
	q := datastore.NewQuery("UserTournamentStats").Ancestor(guestbookKey(ctx)).
		Filter("TournamentId =", tournamentID).
		Filter("UserId =", userID).
		Limit(1)
	var stats []UserTournamentStats
	_, err := q.GetAll(ctx, &stats)
	if err != nil {
		return false, UserTournamentStats{}, err
	}
	if len(stats) == 0 {
		return false, UserTournamentStats{}, err
	}
	return true, stats[0], nil
}

func readOrCreateStatsWithID(ctx context.Context, tournamentID int64, userID int64) (
	UserTournamentStats, error) {
	exist, stats, err := readStatsWithID(ctx, tournamentID, userID)
	if err != nil {
		return UserTournamentStats{}, err
	}

	if exist {
		return stats, nil
	}

	// create a new entry within a transaction
	err = datastore.RunInTransaction(ctx, func(ctx context.Context) error {
		exist, stats, err = readStatsWithID(ctx, tournamentID, userID)
		if exist {
			// there are multiple queries that were trying to create this entry,
			// and it's now created by some other thread. We can now return stats.
			return nil
		}

		key := datastore.NewIncompleteKey(ctx, "UserTournamentStats", guestbookKey(ctx))
		stats = UserTournamentStats{
			TournamentID: tournamentID,
			UserID:       userID,
			FFAWins:      InitialFFAWins,
			Wins:         InitialWins,
			Losses:       InitialLosses,
			Rating:       InitialRating,
		}

		_, putErr := datastore.Put(ctx, key, stats)

		return putErr
	}, nil)

	if err != nil {
		return UserTournamentStats{}, err
	}

	return stats, nil
}
