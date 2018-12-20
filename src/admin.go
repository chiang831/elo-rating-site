package guestbook

import (
	"encoding/json"
	"io"
	"net/http"
	"path"

	"cloud.google.com/go/storage"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/file"
	"google.golang.org/appengine/user"
)

// Admin page
func admin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("static", "admin.html"))
}

// Re-run all matches
func rerunMatches(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// Get users
	queryUser := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(c))
	var users []UserProfile
	keyUsers, err := queryUser.GetAll(c, &users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Get matches
	queryMatch := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Order("Date")
	var matches []Match
	keyMatches, err := queryMatch.GetAll(c, &matches)
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
		matches[i] = createMatch(users[idxW], users[idxL], m.Tournament, m.Submitter, m.Note, m.Date)
		// Update user
		users[idxW].Rating = matches[i].WinnerRatingAfter
		users[idxW].Wins += 1
		users[idxL].Rating = matches[i].LoserRatingAfter
		users[idxL].Losses += 1
	}
	// Restore users
	for i, u := range users {
		datastore.Put(c, keyUsers[i], &u)
	}
	// Restore matches
	for i, m := range matches {
		datastore.Put(c, keyMatches[i], &m)
	}
	// Clear latest match
	existLatestMatch = false
}

// Delete a match entry from database
func deleteMatchEntry(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	encoded_string := ""
	ret := ""

	// Get encoded string
	keys, ok := r.URL.Query()["key"]
	if ok && len(keys) == 1 {
		encoded_string = keys[0]
	}

	// Get decoded key
	key, err := datastore.DecodeKey(encoded_string)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		ret = "Error"
	} else {
		// Remove the key
		datastore.Delete(c, key)
		rerunMatches(w, r) // TODO(music960633): Should we run this here?
		ret = "OK"
	}

	js, err_js := json.Marshal(ret)
	if err_js != nil {
		http.Error(w, err_js.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Switch winner/loser of a match
func switchMatchUsers(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	encoded_string := ""
	ret := ""

	// Get encoded string
	keys, ok := r.URL.Query()["key"]
	if ok && len(keys) == 1 {
		encoded_string = keys[0]
	}

	// Get decoded key
	key, err := datastore.DecodeKey(encoded_string)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		ret = "Error"
	} else {
		// Get the entry
		match := Match{}
		err = datastore.Get(c, key, &match)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			ret = "Error"
		} else {
			// Swap winner and loser
			tmp := match.Winner
			match.Winner = match.Loser
			match.Loser = tmp

			// Store back
			datastore.Put(c, key, &match)
			rerunMatches(w, r) // TODO(music960633): Should we run this here?
			ret = "OK"
		}
	}

	js, err_js := json.Marshal(ret)
	if err_js != nil {
		http.Error(w, err_js.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func submitBadge(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// Check if badge already exist
	badgeName := r.FormValue("name")
	queryBadge := datastore.NewQuery("Badge").Ancestor(guestbookKey(c)).Filter("Name =", badgeName).KeysOnly()
	keys, err := queryBadge.GetAll(c, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if len(keys) != 0 {
		http.Error(w, "Badge name already registered.", http.StatusInternalServerError)
		return
	}
	// Get bucket
	bucketName, err := file.DefaultBucketName(c)
	if err != nil {
		http.Error(w, "failed to get default GCS bucket.", http.StatusInternalServerError)
	}
	client, err := storage.NewClient(c)
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
	writer := bucket.Object(badgeName).NewWriter(c)
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
		Author:      user.Current(c).String(),
		Path:        "https://storage.googleapis.com/" + bucketName + "/" + badgeName,
	}
	key := datastore.NewIncompleteKey(c, "Badge", guestbookKey(c))
	_, err = datastore.Put(c, key, &badge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusFound)
}

func submitUserBadge(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// Get user
	user_name := r.FormValue("user_name")
	existU, _, _, errUser := existUser(c, user_name)
	if errUser != nil {
		http.Error(w, errUser.Error(), http.StatusInternalServerError)
		return
	} else if !existU {
		http.Error(w, "User "+user_name+" does not exist.", http.StatusInternalServerError)
		return
	}
	// Get badge
	badge_name := r.FormValue("badge_name")
	existB, _, _, errBadge := existBadge(c, badge_name)
	if errBadge != nil {
		http.Error(w, errBadge.Error(), http.StatusInternalServerError)
		return
	} else if !existB {
		http.Error(w, "Badge "+badge_name+" does not exist.", http.StatusInternalServerError)
		return
	}
	// Get UserBadge
	queryBadge := datastore.NewQuery("UserBadge").Ancestor(guestbookKey(c)).Filter("User =", user_name)
	var userBadges []UserBadge
	keys, err := queryBadge.GetAll(c, &userBadges)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	key := datastore.Key{}
	userBadge := UserBadge{}
	if len(userBadges) == 0 {
		key = *datastore.NewIncompleteKey(c, "UserBadge", guestbookKey(c))
		userBadge = UserBadge{
			User:       user_name,
			BadgeNames: []string{},
		}
	} else {
		key = *keys[0]
		userBadge = userBadges[0]
	}
	userBadge.BadgeNames = append(userBadge.BadgeNames, badge_name)
	datastore.Put(c, &key, &userBadge)
	http.Redirect(w, r, "/admin", http.StatusFound)
}
