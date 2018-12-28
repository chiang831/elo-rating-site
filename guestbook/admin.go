package guestbook

import (
	"encoding/json"
	"io"
	"net/http"
	"path"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/file"
	"google.golang.org/appengine/user"
)

// HandleAdmin handles http request to /admin
func HandleAdmin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("static", "admin.html"))
}

// HandleRerunMatches handles http request to /rerun
func HandleRerunMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Get users
	queryUser := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(ctx))
	var users []UserProfile
	keyUsers, err := queryUser.GetAll(ctx, &users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Get matches
	queryMatch := datastore.NewQuery("Match").Ancestor(guestbookKey(ctx)).Order("Date")
	var matches []Match
	keyMatches, err := queryMatch.GetAll(ctx, &matches)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Reset ratings
	for i := range users {
		users[i].Rating = startingElo
		users[i].Wins = 0
		users[i].Losses = 0
	}
	// Name to index map
	mp := make(map[string]int)
	for i, u := range users {
		mp[u.Name] = i
	}
	// Run matches
	for i, m := range matches {
		idxW, existW := mp[m.Winner]
		idxL, existL := mp[m.Loser]
		if !existW || !existL {
			http.Error(w, "Datastore error", http.StatusInternalServerError)
			return
		}
		// Update match
		matches[i] = createMatch(
			users[idxW].Rating, users[idxL].Rating,
			users[idxW].Name, users[idxL].Name,
			m.Tournament, m.Submitter, m.Note, m.Date)
		// Update user
		users[idxW].Rating = matches[i].WinnerRatingAfter
		users[idxW].Wins++
		users[idxL].Rating = matches[i].LoserRatingAfter
		users[idxL].Losses++
	}
	// Restore users
	for i, u := range users {
		datastore.Put(ctx, keyUsers[i], &u)
	}
	// Restore matches
	for i, m := range matches {
		datastore.Put(ctx, keyMatches[i], &m)
	}
	// Clear latest match
	existLatestMatch = false
}

// HandleDeleteMatchEntry deletes a match entry from database
func HandleDeleteMatchEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	encodedString := ""
	ret := ""

	// Get encoded string
	keys, ok := r.URL.Query()["key"]
	if ok && len(keys) == 1 {
		encodedString = keys[0]
	}

	// Get decoded key
	key, err := datastore.DecodeKey(encodedString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		ret = "Error"
	} else {
		// Remove the key
		datastore.Delete(ctx, key)
		HandleRerunMatches(w, r) // TODO(music960633): Should we run this here?
		ret = "OK"
	}

	js, errJs := json.Marshal(ret)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// HandleSwitchMatchUsers switches winner/loser of a match
func HandleSwitchMatchUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	encodedString := ""
	ret := ""

	// Get encoded string
	keys, ok := r.URL.Query()["key"]
	if ok && len(keys) == 1 {
		encodedString = keys[0]
	}

	// Get decoded key
	key, err := datastore.DecodeKey(encodedString)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		ret = "Error"
	} else {
		// Get the entry
		match := Match{}
		err = datastore.Get(ctx, key, &match)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			ret = "Error"
		} else {
			// Swap winner and loser
			tmp := match.Winner
			match.Winner = match.Loser
			match.Loser = tmp

			// Store back
			datastore.Put(ctx, key, &match)
			HandleRerunMatches(w, r) // TODO(music960633): Should we run this here?
			ret = "OK"
		}
	}

	js, errJs := json.Marshal(ret)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// HandleSubmitBadge handles http request to submit a badge
func HandleSubmitBadge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Check if badge already exist
	badgeName := r.FormValue("name")
	queryBadge := datastore.NewQuery("Badge").Ancestor(guestbookKey(ctx)).Filter("Name =", badgeName).KeysOnly()
	keys, err := queryBadge.GetAll(ctx, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(keys) != 0 {
		http.Error(w, "Badge name already registered.", http.StatusInternalServerError)
		return
	}
	// Get bucket
	bucketName, err := file.DefaultBucketName(ctx)
	if err != nil {
		http.Error(w, "failed to get default GCS bucket.", http.StatusInternalServerError)
	}
	client, err := storage.NewClient(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer client.Close()
	bucket := client.Bucket(bucketName)
	r.ParseMultipartForm(32 << 20)
	// Write
	icon, header, err := r.FormFile("icon")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writer := bucket.Object(badgeName).NewWriter(ctx)
	writer.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	writer.ContentType = header.Header.Get("Content-Type")
	if _, err := io.Copy(writer, icon); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := writer.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Add badge
	badge := Badge{
		Name:        badgeName,
		Description: r.FormValue("description"),
		Author:      user.Current(ctx).String(),
		Path:        "https://storage.googleapis.com/" + bucketName + "/" + badgeName,
	}
	key := datastore.NewIncompleteKey(ctx, "Badge", guestbookKey(ctx))
	_, err = datastore.Put(ctx, key, &badge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

// HandleSubmitUserBadge handles http request to submit a user badge
func HandleSubmitUserBadge(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Get user
	userName := r.FormValue("user_name")
	existU, _, _, errUser := existUser(ctx, userName)
	if errUser != nil {
		http.Error(w, errUser.Error(), http.StatusInternalServerError)
		return
	} else if !existU {
		http.Error(w, "User "+userName+" does not exist.", http.StatusInternalServerError)
		return
	}
	// Get badge
	badgeName := r.FormValue("badge_name")
	existB, _, _, errBadge := existBadge(ctx, badgeName)
	if errBadge != nil {
		http.Error(w, errBadge.Error(), http.StatusInternalServerError)
		return
	} else if !existB {
		http.Error(w, "Badge "+badgeName+" does not exist.", http.StatusInternalServerError)
		return
	}
	// Get UserBadge
	queryBadge := datastore.NewQuery("UserBadge").Ancestor(guestbookKey(ctx)).Filter("User =", userName)
	var userBadges []UserBadge
	keys, err := queryBadge.GetAll(ctx, &userBadges)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	key := datastore.Key{}
	userBadge := UserBadge{}
	if len(userBadges) == 0 {
		key = *datastore.NewIncompleteKey(ctx, "UserBadge", guestbookKey(ctx))
		userBadge = UserBadge{
			User:       userName,
			BadgeNames: []string{},
		}
	} else {
		key = *keys[0]
		userBadge = userBadges[0]
	}
	userBadge.BadgeNames = append(userBadge.BadgeNames, badgeName)
	datastore.Put(ctx, &key, &userBadge)
	http.Redirect(w, r, "/admin", http.StatusFound)
}
