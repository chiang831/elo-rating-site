package guestbook

import (
        "encoding/json"
        "path"
        "net/http"
        "html/template"
        "regexp"
        "time"
        "appengine"
        "appengine/datastore"
        "appengine/user"
        "strconv"
        "fmt"
)

func init() {
        // Main page
        http.HandleFunc("/", root)
        // Child pages
        http.HandleFunc("/add_user", addUser)
        http.HandleFunc("/add_match_result", addMatchResult)
        http.HandleFunc("/profile", profile)
        // Submit data
        http.HandleFunc("/submit_greeting", submitGreeting)
        http.HandleFunc("/submit_user", submitUser)
        http.HandleFunc("/submit_match_result", submitMatchResult)
        // Requests
        http.HandleFunc("/request_users", requestUsers)
        http.HandleFunc("/request_latest_match", requestLatestMatch)
        http.HandleFunc("/request_user_profiles", requestUserProfiles)
        http.HandleFunc("/request_detail_results", requestDetailMatchResults)
        http.HandleFunc("/request_greetings", requestGreetings)
        http.HandleFunc("/request_recent_matches", requestRecentMatches)
        // Admin area
        http.HandleFunc("/delete_match_entry", deleteMatchEntry)
        http.HandleFunc("/switch_match_users", switchMatchUsers)
        http.HandleFunc("/rerun", rerunMatches)
        // Static files
        http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

// guestbookKey returns the key used for all guestbook entries.
func guestbookKey(c appengine.Context) *datastore.Key {
        // The string "default_guestbook" here could be varied to have multiple guestbooks.
        return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

var existLatestMatch = false
var latestMatch Match
const startingElo float64 = 1200.0

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
        g := UserProfile{
                Tournament: "Default",
                Name: name,
                Rating: startingElo,
                Wins: 0,
                Losses: 0,
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


// [START func_addGreeting]
func submitGreeting(w http.ResponseWriter, r *http.Request) {
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
// [END func_addGreeting]

func requestUsers(w http.ResponseWriter, r *http.Request) {
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


func requestUserProfiles(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        queryUser := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(c)).Order("-Rating")
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

func requestDetailMatchResults(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        // Get users
        queryUser := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(c)).Order("-Rating")
        var users []UserProfile
        if _, err := queryUser.GetAll(c, &users); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        // Get matches
        queryMatch := datastore.NewQuery("Match").Ancestor(guestbookKey(c))
        var matches []Match
        if _, err := queryMatch.GetAll(c, &matches); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        // Name to index map
        mp := make(map[string]int)
        for i, u := range users {
                mp[u.Name] = i
        }
        // Set usernames
        usernames := make([]string, len(users))
        for i, u := range users {
                usernames[i] = u.Name
        }
        // Set resultTable
        resultTable := make([][]DetailMatchResultEntry, len(users))
        for i := range resultTable {
                resultTable[i] = make([]DetailMatchResultEntry, len(users))
                for j := range resultTable[i] {
                        resultTable[i][j].Wins = 0
                        resultTable[i][j].Losses = 0
                }
        }
        for _, m := range matches {
                idxW, existW := mp[m.Winner]
                idxL, existL := mp[m.Loser]
                if !existW || !existL {
                        http.Error(w, "Datastore Error", http.StatusInternalServerError)
                        return
                }
                resultTable[idxW][idxL].Wins += 1
                resultTable[idxL][idxW].Losses += 1
        }
        for i := range resultTable {
                for j := range resultTable[i] {
                        resultTable[i][j].Color = getColor(users[i], users[j], resultTable[i][j].Wins, resultTable[i][j].Losses)
                }
        }

        // Return JSON
        matchData := MatchData {
                Usernames: usernames,
                ResultTable: resultTable,
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

        // Get number of greetings to retrieve
        // If the number is not a positive integer, return nil
        limit := -1
        keys, ok := r.URL.Query()["num"]
        if ok && len(keys) == 1 {
                new_limit, err := strconv.Atoi(keys[0])
                if err == nil && new_limit > 0 {
                        limit = new_limit
                }
        }

        greetings := []Greeting{}
        if limit != -1 {
                queryGreeting := datastore.NewQuery("Greeting").Ancestor(guestbookKey(c)).Order("-Date").Limit(limit)
                greetings = make([]Greeting, 0, limit)
                if _, err := queryGreeting.GetAll(c, &greetings); err != nil {
                        http.Error(w, err.Error(), http.StatusInternalServerError)
                        return
                }
        }

        js, err_js := json.Marshal(greetings)
        if err_js != nil {
                http.Error(w, err_js.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(js)
}

func requestRecentMatches(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)

        // Get number of matches to retrieve
        // If the number is not a positive integer, return nil
        limit := -1
        keys, ok := r.URL.Query()["num"]
        if ok && len(keys) == 1 {
                new_limit, err := strconv.Atoi(keys[0])
                if err == nil && new_limit > 0 {
                        limit = new_limit
                }
        }

        matchWithKeys := []MatchWithKey{}
        if limit != -1 {
                queryMatch := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Order("-Date").Limit(limit)
                var matches []Match
                keyMatches, err := queryMatch.GetAll(c, &matches)
                if err != nil {
                        http.Error(w, err.Error(), http.StatusInternalServerError)
                        return
                }
                matchWithKeys = make([]MatchWithKey, len(matches))
                for i, m := range matches {
                        matchWithKeys[i] = MatchWithKey {
                                Match: m,
                                Key: keyMatches[i].Encode(),
                        }
                }
        }

        js, err_js := json.Marshal(matchWithKeys)
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
        // silver
        return "rgb(192,192,192)"
    } else if (wins == 0) && (losses == 0) {
        // white
        return "rgb(255,255,255)"
    } else if wins >= losses {
        ratio := float64(wins - losses) / float64(wins + losses)
        // limegreen: rgb(50, 205, 50)
        // gold: rgb(255, 215, 0)
        r := 50.0 * ratio + 255.0 * (1.0 - ratio)
        g := 205.0 * ratio + 215.0 * (1.0 - ratio)
        b := 50.0 * ratio + 0.0 * (1.0 - ratio)
        return fmt.Sprintf("rgb(%.0f,%.0f,%.0f)", r, g, b)
    } else {
        ratio := float64(2 * wins) / float64(wins + losses)
        // gold: rgb(255, 215, 0)
        // tomato: rgb(255, 99, 71)
        r := 255.0 * ratio + 255.0 * (1.0 - ratio)
        g := 215.0 * ratio + 99.0 * (1.0 - ratio)
        b := 0.0 * ratio + 71.0 * (1.0 - ratio)
        return fmt.Sprintf("rgb(%.0f,%.0f,%.0f)", r, g, b)
    }
}

func profile(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        // Get username
        username := ""
        keys, ok := r.URL.Query()["user"]
        if ok && len(keys) == 1 {
                username = keys[0]
        }
        exist, _, user, err := existUser(c, username)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        } else if !exist {
                http.Redirect(w, r, "/", http.StatusFound)
                return
        }
        // User template
        profile_path := path.Join("static", "profile.html")
        tmpl, tmpl_err := template.ParseFiles(profile_path)
        if tmpl_err != nil {
                http.Error(w, tmpl_err.Error(), http.StatusInternalServerError)
                return
        }
        if err = tmpl.Execute(w, user); err != nil {
                http.Error(w, tmpl_err.Error(), http.StatusInternalServerError)
                return
        }
        // http.ServeFile(w, r, path.Join("static", "profile.html"))
}

