package guestbook

import (
        "encoding/json"
        "path"
        "net/http"
        "regexp"
        "time"
        "appengine"
        "appengine/datastore"
        "appengine/user"
)

func init() {
        http.HandleFunc("/", root)
        http.HandleFunc("/sign", sign)
        http.HandleFunc("/add_user", addUser)
        http.HandleFunc("/submit_user", submitUser)
        http.HandleFunc("/add_match_result", addMatchResult)
        http.HandleFunc("/submit_match_result", submitMatchResult)
        http.HandleFunc("/users", listUsers)
        http.HandleFunc("/latest_match", requestLatestMatch)
        http.HandleFunc("/request_match_data", requestMatchData)
        http.HandleFunc("/request_greetings", requestGreetings)
        http.HandleFunc("/request_matches", requestMatches)
        http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

// guestbookKey returns the key used for all guestbook entries.
func guestbookKey(c appengine.Context) *datastore.Key {
        // The string "default_guestbook" here could be varied to have multiple guestbooks.
        return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

var existLatestMatch = false
var latestMatch Match

// [START func_test_root]
func root(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, path.Join("static", "main.html"))
}

// [START add_user]
func addUser(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, path.Join("static", "add_user.html"))
}

func existUser(c appengine.Context, name string) (bool, datastore.Key, UserProfile, error) {
        q := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(c)).Filter("Name =", name)
        var users []UserProfile
        keys, err := q.GetAll(c, &users)
        if err != nil {
                return false, datastore.Key{}, UserProfile{}, err
        }
        if len(users) != 0 {
                return true, *keys[0], users[0], nil
        }
        return false, datastore.Key{}, UserProfile{}, nil
}

// [START submit_match_result]
func submitUser(w http.ResponseWriter, r *http.Request) {
        // [START new_context]
        c := appengine.NewContext(r)
        // [END new_context]

        // Check valid name
        name := r.FormValue("name")

        re, _ := regexp.Compile("^[A-Za-z0-9_]{3,20}$")

        isValid := re.MatchString(name)
        if !isValid {
                http.Error(w, "Not a valid name", http.StatusBadRequest)
                return
        }

        exist, _, _, err := existUser(c, name)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        if exist {
                http.Error(w, "Already registered", http.StatusBadRequest)
                return
        }

        // Is a valid new user.
        const startingElo = 1200
        g := UserProfile{
                Tournament: "Default",
                Name: name,
                Rating: startingElo,
                JoinDate: time.Now(),
        }

        // [END getall]
        key := datastore.NewIncompleteKey(c, "UserProfile", guestbookKey(c))
        _, err = datastore.Put(c, key, &g)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        http.Redirect(w, r, "/", http.StatusFound)
        // [END if_user]
}

// [START add_match_result]
func addMatchResult(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, path.Join("static", "add_match_result.html"))
}


// [START func_sign]
func sign(w http.ResponseWriter, r *http.Request) {
        // [START new_context]
        c := appengine.NewContext(r)
        // [END new_context]
        g := Greeting{
                Content: r.FormValue("content"),
                Date:    time.Now(),
        }

        // Ignore empty comment.
        if len(g.Content) == 0 {
                http.Redirect(w, r, "/", http.StatusFound)
                return
        }

        // [START if_user]
        if u := user.Current(c); u != nil {
                g.Author = u.String()
        }
        // We set the same parent key on every Greeting entity to ensure each Greeting
        // is in the same entity group. Queries across the single entity group
        // will be consistent. However, the write rate to a single entity group
        // should be limited to ~1/second.
        key := datastore.NewIncompleteKey(c, "Greeting", guestbookKey(c))
        _, err := datastore.Put(c, key, &g)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        http.Redirect(w, r, "/", http.StatusFound)
        // [END if_user]
}
// [END func_sign]

func listUsers(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        queryUser := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(c)).Order("Name")
        var users []UserProfile
        if _, err := queryUser.GetAll(c, &users); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        js, err_js := json.Marshal(users)
        if err_js != nil {
                http.Error(w, err_js.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(js)
}

func requestLatestMatch(w http.ResponseWriter, r *http.Request) {
        if !existLatestMatch {
                nil_js, nil_err_js := json.Marshal(nil)
                if nil_err_js != nil {
                        http.Error(w, nil_err_js.Error(), http.StatusInternalServerError)
                        return
                }
                w.Header().Set("Content-Type", "application/json")
                w.Write(nil_js)
                return
        }

        js, err_js := json.Marshal(latestMatch)
        if err_js != nil {
                http.Error(w, err_js.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(js)
}

func requestMatchData(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        queryUser := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(c)).Order("-Rating")
        var users []UserProfile
        if _, err := queryUser.GetAll(c, &users); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        userDataToShows := make([]UserDataToShow, len(users))
        detailMatchResults := make([]DetailMatchResult, len(users))
        for i, u := range users {
                results := make([]DetailMatchResultEntry, len(users))
                oneUserTotalWin := 0
                oneUserTotalLose := 0

                for j, v := range users {
                        queryOneUserWin := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Filter("Winner =", u.Name).Filter("Loser =", v.Name).KeysOnly()
                        keyWin, err := queryOneUserWin.GetAll(c, nil)
                        if err != nil {
                                http.Error(w, err.Error(), http.StatusInternalServerError)
                                return
                        }
                        queryOneUserLoss := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Filter("Loser =", u.Name).Filter("Winner =", v.Name).KeysOnly()
                        keyLoss, err := queryOneUserLoss.GetAll(c, nil)
                        if err != nil {
                                http.Error(w, err.Error(), http.StatusInternalServerError)
                                return
                        }

                        wins := len(keyWin)
                        losses := len(keyLoss)
                        oneUserTotalWin += wins
                        oneUserTotalLose += losses

                        results[j] = DetailMatchResultEntry {
                                Wins: wins,
                                Losses: losses,
                                Color: getColor(u, v, wins, losses),
                        }
                }

                userDataToShows[i] = UserDataToShow {
                        Name: u.Name,
                        Rating: int(u.Rating),
                        Wins: oneUserTotalWin,
                        Losses: oneUserTotalLose,
                }
                detailMatchResults[i] = DetailMatchResult {
                        Name: u.Name,
                        Results: results,
                }
        }

        matchData := MatchData {
                UserDataToShows: userDataToShows,
                DetailMatchResults: detailMatchResults,
        }

        js, err_js := json.Marshal(matchData)
        if err_js != nil {
                http.Error(w, err_js.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(js)
}

func requestGreetings(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        queryGreeting := datastore.NewQuery("Greeting").Ancestor(guestbookKey(c)).Order("-Date").Limit(20)
        greetings := make([]Greeting, 0, 20)
        if _, err := queryGreeting.GetAll(c, &greetings); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        js, err_js := json.Marshal(greetings)
        if err_js != nil {
                http.Error(w, err_js.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(js)
}

func requestMatches(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        queryMatch := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Order("-Date").Limit(20)
        matches := make([]Match, 0, 20)
        if _, err := queryMatch.GetAll(c, &matches); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        js, err_js := json.Marshal(matches)
        if err_js != nil {
                http.Error(w, err_js.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(js)
}

// Get the color of win/lose/tie
func getColor(u UserProfile, v UserProfile, wins int, losses int) string {
    if u.Name == v.Name {
        return "silver"
    } else if (wins == 0) && (losses == 0) {
        return "white"
    } else if wins > losses {
        return "limegreen"
    } else if wins < losses {
        return "tomato"
    } else {
        return "yellow"
    }
}
