package guestbook

import (
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
        // Run matches
        for i, m := range matches {
                idxW := findUserIndex(m.Winner, users)
                idxL := findUserIndex(m.Loser, users)
                if idxW == -1 || idxL == -1 {
                        http.Error(w, err.Error(), http.StatusInternalServerError)
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
        http.Redirect(w, r, "/", http.StatusFound)
}

func findUserIndex(name string, users []UserProfile) int {
        for i, u := range users {
                if name == u.Name {
                        return i
                }
        }
        return -1
}
