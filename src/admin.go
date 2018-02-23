package guestbook

import (
        "encoding/json"
        "net/http"
        "appengine"
        "appengine/datastore"
)

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
                datastore.Put(c, keyMatches[i], &m);
        }
        // Clear latest match
        existLatestMatch = false
}

func findUserIndex(name string, users []UserProfile) int {
        for i, u := range users {
                if name == u.Name {
                        return i
                }
        }
        return -1
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

