package guestbook

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
)

func readUserProfile(ctx context.Context, userID int64) (UserProfile, error) {
	var userProfile UserProfile
	key := datastore.NewKey(
		ctx,               // context.Context
		"UserProfile",     // Kind
		"",                // String ID; empty means no string ID
		userID,            // Integer ID; if 0, generate automatically. Ignored if string ID specified.
		guestbookKey(ctx), // Parent Key; nil means no parent

	)
	if err := datastore.Get(ctx, key, &userProfile); err != nil {
		return UserProfile{}, err
	}

	return userProfile, nil
}

func readUserProfiles(ctx context.Context, userIDs []int64) ([]UserProfile, error) {
	profiles := make([]UserProfile, len(userIDs))
	for i, userID := range userIDs {
		var err error
		profiles[i], err = readUserProfile(ctx, userID)
		if err != nil {
			return nil, err
		}
	}
	return profiles, nil
}

func readUserIDAndProfileMapping(ctx context.Context, userIDs []int64) (map[int64]UserProfile, error) {
	profiles, err := readUserProfiles(ctx, userIDs)

	if err != nil {
		return nil, err
	}

	m := make(map[int64]UserProfile)
	for i, profile := range profiles {
		m[userIDs[i]] = profile
	}
	return m, nil
}
