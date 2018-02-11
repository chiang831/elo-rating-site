package guestbook

import (
        "fmt"
        "html/template"
        "path"
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

type RootPageVars struct {
        Greetings []Greeting
        Matches []Match
        UserDataToShows []UserDataToShow
}

func init() {
        http.HandleFunc("/", root)
        http.HandleFunc("/sign", sign)
        http.HandleFunc("/register", registerUser)
        http.HandleFunc("/submit_user", submitUser)
        http.HandleFunc("/add", addMatchResult)
        http.HandleFunc("/submit_match_result", submitMatchResult)
}

// guestbookKey returns the key used for all guestbook entries.
func guestbookKey(c appengine.Context) *datastore.Key {
        // The string "default_guestbook" here could be varied to have multiple guestbooks.
        return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

const addUserForm = `
<html>
  <head>
    <title>Add a user</title>
  </head>
  <body>
    <form action="/submit_user" method="post">
      <div><p>NAME</p><textarea name="name" rows="1" cols="10"></textarea></div>
      <div><input type="submit" value="Add a user"></div>
    </form>
  </body>
</html>
`

// [START add_match_result]
func registerUser(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, addUserForm)
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
        http.ServeFile(w, r, path.Join("static", "add.html"))
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


        http.Redirect(w, r, "/", http.StatusFound)
        // [END if_user]

}

// [START func_root]
func root(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        // Ancestor queries, as shown here, are strongly consistent with the High
        // Replication Datastore. Queries that span entity groups are eventually
        // consistent. If we omitted the .Ancestor from this query there would be
        // a slight chance that Greeting that had just been written would not
        // show up in a query.
        // [START query]
        queryGreeting := datastore.NewQuery("Greeting").Ancestor(guestbookKey(c)).Order("-Date").Limit(20)
        // [END query]
        // [START getall]
        greetings := make([]Greeting, 0, 20)
        if _, err := queryGreeting.GetAll(c, &greetings); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        // [END getall]

        // [START query]
        queryMatch := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Order("-Date").Limit(20)
        // [END query]
        // [START getall]
        matches := make([]Match, 0, 20)
        if _, err := queryMatch.GetAll(c, &matches); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        // [END getall]

        // [START query]
        queryUser := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(c)).Order("-Rating")
        // [END query]
        // [START getall]
        var users []UserProfile
        if _, err := queryUser.GetAll(c, &users); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        // For each user, query number of wins and losses.
        userDataToShows := make([]UserDataToShow, len(users))
        for i, u := range users {
                 queryOneUserWin := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Filter("Winner =", u.Name)
                 var oneUserWin []Match
                 if _, err := queryOneUserWin.GetAll(c, &oneUserWin); err != nil {
                        http.Error(w, err.Error(), http.StatusInternalServerError)
                        return
                 }
                 queryOneUserLoss := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Filter("Loser =", u.Name)
                 var oneUserLoss []Match
                 if _, err := queryOneUserLoss.GetAll(c, &oneUserLoss); err != nil {
                        http.Error(w, err.Error(), http.StatusInternalServerError)
                        return
                 }

                 userDataToShows[i] = UserDataToShow {
                         Name: u.Name,
                         Rating: int(u.Rating),
                         Wins: len(oneUserWin),
                         Losses: len(oneUserLoss),
                 }
        }

        // Fill in template.
        vars := RootPageVars {
                Greetings: greetings,
                Matches: matches,
                UserDataToShows: userDataToShows,
        }

        if err := guestbookTemplate.Execute(w, vars); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
        }
}
// [END func_root]

// Template for root page
var guestbookTemplate = template.Must(template.New("book").Parse(`
<html>
  <head>
    <title>ELO Rating</title>
  <style>
    table, th, td {
        border: 1px solid black;
        border-collapse: collapse;
    }
    th, td {
        padding: 5px;
    }
    th {
        text-align: left;
    }
  </style>
  </head>
  <body>
    <h1>Leaderboard</h1>
    <table style="width:100%">
      <tr>
        <th>Name</th>
        <th>Rating</th>
        <th>Wins</th>
        <th>Losses</th>
      </tr>
      {{range .UserDataToShows}}
        <tr>
          <td>{{.Name}}</td>
          <td>{{.Rating}}</td>
          <td>{{.Wins}}</td>
          <td>{{.Losses}}</td>
        </tr>
      {{end}}
    </table>
    <p>
    <form action="/register">
        <input type="submit" value="Add a user" />
    </form>
    <form action="/add">
        <input type="submit" value="Add a match result" />
    </form>
    </p>
    <h1>Recent Matches</h1>
    {{range .Matches}}
      <p>
      {{.Date}}
      {{with .Submitter}}
        <b>{{.}}</b> submitted:
      {{else}}
        An anonymous person submitted:
      {{end}}
      Win: {{.Winner}} ({{.WinnerRatingBefore}} -> {{.WinnerRatingAfter}}) |
      Loss: {{.Loser}} ({{.LoserRatingBefore}} -> {{.LoserRatingAfter}}) | {{.Note}}
      </p>
    {{end}}
    <h1>Recent Comments</h1>
    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <p>
      <div><input type="submit" value="Add a comment"></div>
      </p>
    </form>
    {{range .Greetings}}
      <p>
      {{.Date}}
      {{with .Author}}
        <b>{{.}}</b> wrote:
      {{else}}
        An anonymous person wrote:
      {{end}}
      {{.Content}}
      </p>
    {{end}}
  </body>
</html>
`))

// [START func_sign]
func sign(w http.ResponseWriter, r *http.Request) {
        // [START new_context]
        c := appengine.NewContext(r)
        // [END new_context]
        g := Greeting{
                Content: r.FormValue("content"),
                Date:    time.Now(),
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

// Expected score of elo_a in a match against elo_b
func expectedScore(elo_a, elo_b float64) float64{
    return 1 / (1 + math.Pow(10, (elo_b - elo_a) / 400))
}

// Get the new Elo rating.
func newElo(old_elo, expected, score float64) float64 {
    return old_elo + 32.0 * (score - expected)
}
