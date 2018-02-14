package guestbook

import (
        "encoding/json"
        "path"
        "log"
        "math"
        "net/http"
        "regexp"
        "time"
        "appengine"
        "appengine/datastore"
        "appengine/user"
)

// [START greeting_struct]
type Greeting struct {
        Author  string
        Content string
        Date    time.Time
}
// [END greeting_struct]

// [START match_struct]
type Match struct {
        Tournament string
        Submitter  string
        Winner     string
        Loser      string
        WinnerRatingBefore int // Just for showing the history. Int is enough.
        WinnerRatingAfter int // Just for showing the history. Int is enough.
        LoserRatingBefore int // Just for showing the history. Int is enough.
        LoserRatingAfter int // Just for showing the history. Int is enough.
        Note       string
        Date       time.Time
}
// [END match_struct]

// [START user_profile]
type UserProfile struct {
        Tournament string
        Name       string
        Rating     float64
        JoinDate   time.Time
}

type UserDataToShow struct {
        Name        string
        Rating      int
        Wins        int
        Losses      int
}

type MatchToShow struct {
        Match       Match
        Expected    bool //Use this to show different icon for underdog.
}

type DetailMatchResultEntry struct {
        Wins        int
        Losses      int
        Color       string
}

type DetailMatchResult struct {
        Name        string
        Results     []DetailMatchResultEntry
}

type MatchData struct {
        UserDataToShows []UserDataToShow
        DetailMatchResults []DetailMatchResult
}

type RootPageVars struct {
        Greetings []Greeting
        MatchToShows []MatchToShow
        UserDataToShows []UserDataToShow
        DetailMatchResults []DetailMatchResult
}

func init() {
        http.HandleFunc("/", root)
        http.HandleFunc("/sign", sign)
        http.HandleFunc("/add_user", addUser)
        http.HandleFunc("/submit_user", submitUser)
        http.HandleFunc("/add_match_result", addMatchResult)
        http.HandleFunc("/submit_match_result", submitMatchResult)
        http.HandleFunc("/users", listUsers)
        http.HandleFunc("/latest_match", latestMatch)
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
var latestMatchToShow MatchToShow

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

// [START submit_match_result]
func submitMatchResult(w http.ResponseWriter, r *http.Request) {
        // [START new_context]
        c := appengine.NewContext(r)
        // [END new_context]

        keyWinner := datastore.Key{}
        keyLoser:= datastore.Key{}
        winner := UserProfile{}
        loser := UserProfile{}
        exist := false
        var err error

        winner_name := r.FormValue("winner")
        loser_name := r.FormValue("loser")

        log.Printf("winner_name: %s", winner_name)
        log.Printf("loser_name: %s", loser_name)

        if winner_name == loser_name {
                http.Error(w, "Winner should not be the same as loser.",
                           http.StatusBadRequest)
        }

        // Check winner is registered.
        exist, keyWinner, winner, err = existUser(c, winner_name)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        if !exist {
                http.Error(w, "Winner has not registered", http.StatusBadRequest)
                return
        }

        // Check loser is registered.
        exist, keyLoser, loser, err = existUser(c, loser_name)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        if !exist {
                http.Error(w, "Loser has not registered", http.StatusBadRequest)
                return
        }

        oldRatingW := winner.Rating
        oldRatingL := loser.Rating

        //Get new ELO value
        expectedScoreW := expectedScore(winner.Rating, loser.Rating)
        newRatingW := newElo(winner.Rating, expectedScoreW, 1.0)

        expectedScoreL := expectedScore(loser.Rating, winner.Rating)
        newRatingL := newElo(loser.Rating, expectedScoreL, 0.0)

        g := Match{
                Tournament: "Default",
                Winner: winner_name,
                Loser: loser_name,
                WinnerRatingBefore: int(oldRatingW),
                WinnerRatingAfter: int(newRatingW),
                LoserRatingBefore: int(oldRatingL),
                LoserRatingAfter: int(newRatingL),
                Note: r.FormValue("note"),
                Date:    time.Now(),
        }

        // [START if_user]
        if u := user.Current(c); u != nil {
                g.Submitter= u.String()
        }

        key := datastore.NewIncompleteKey(c, "Match", guestbookKey(c))

        keyMatch := &datastore.Key{}
        keyMatch, err = datastore.Put(c, key, &g)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        winner.Rating = newRatingW
        loser.Rating = newRatingL

        // Try to update winner rating.
        _, err = datastore.Put(c, &keyWinner, &winner)
        if err != nil {
                // Remove match entity as best-effort fallback.
                datastore.Delete(c, keyMatch)

                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        // Try to update loser rating.
        _, err = datastore.Put(c, &keyLoser, &loser)
        if err != nil {
                // Remove match entity as best-effort fallback.
                datastore.Delete(c, keyMatch)
                // Change winner rating back.
                winner.Rating = oldRatingW
                datastore.Put(c, &keyWinner, &winner)

                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }


        existLatestMatch = true;
        latestMatchToShow = MatchToShow {
                Match: g,
                Expected: oldRatingW >= oldRatingL,
        }

        http.Redirect(w, r, "/add_match_result", http.StatusFound)
        // [END if_user]

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

func latestMatch(w http.ResponseWriter, r *http.Request) {
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

        js, err_js := json.Marshal(latestMatchToShow)
        if err_js != nil {
                http.Error(w, err_js.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(js)
}

func requestMatchData(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        queryUser := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(c)).Order("Name")
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
                        queryOneUserWin := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Filter("Winner =", u.Name).Filter("Loser =", v.Name)
                        var oneUserWin []Match
                        if _, err := queryOneUserWin.GetAll(c, &oneUserWin); err != nil {
                                http.Error(w, err.Error(), http.StatusInternalServerError)
                                return
                        }
                        queryOneUserLoss := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Filter("Loser =", u.Name).Filter("Winner =", v.Name)
                        var oneUserLoss []Match
                        if _, err := queryOneUserLoss.GetAll(c, &oneUserLoss); err != nil {
                                http.Error(w, err.Error(), http.StatusInternalServerError)
                                return
                        }

                        wins := len(oneUserWin)
                        losses := len(oneUserLoss)
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

        matchToShows := make([]MatchToShow, len(matches))
        for i, m := range matches {
                matchToShows[i] = MatchToShow {
                        Match: m,
                        Expected: m.WinnerRatingBefore >= m.LoserRatingBefore,
                }
        }

        js, err_js := json.Marshal(matchToShows)
        if err_js != nil {
                http.Error(w, err_js.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(js)
}

// Expected score of elo_a in a match against elo_b
func expectedScore(elo_a, elo_b float64) float64{
    return 1 / (1 + math.Pow(10, (elo_b - elo_a) / 400))
}

// Get the new Elo rating.
func newElo(old_elo, expected, score float64) float64 {
    return old_elo + 32.0 * (score - expected)
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
