package guestbook

import (
        "fmt"
        "html/template"
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
        Note       string
        Date       time.Time
}
// [END match_struct]

// [START user_profile]
type UserProfile struct {
        Tournament string
        Name       string
        Rating     int
        JoinDate   time.Time
}

type RootPageVars struct {
        Greetings []Greeting
        Matches []Match
}

func init() {
        http.HandleFunc("/", root)
        http.HandleFunc("/sign", sign)
        http.HandleFunc("/register", registerUser)
        http.HandleFunc("/submit_user", submitUser)
        http.HandleFunc("/add", add_match_result)
        http.HandleFunc("/submit_match_result", submit_match_result)
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

        const startingElo = 1200
        g := UserProfile{
                Tournament: "Default",
                Name: name,
                Rating: startingElo,
                JoinDate: time.Now(),
        }

        // [START query]
        q := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(c)).Filter("Name =", g.Name)
        // [END query]
        // [START getall]
        var users []UserProfile
        if _, err := q.GetAll(c, &users); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        if len(users) != 0 {
                http.Error(w, "Already registered", http.StatusBadRequest)
                return
        }

        // [END getall]
        key := datastore.NewIncompleteKey(c, "UserProfile", guestbookKey(c))
        _, err := datastore.Put(c, key, &g)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        http.Redirect(w, r, "/", http.StatusFound)
        // [END if_user]
}

const addMatchForm = `
<html>
  <head>
    <title>Add a match result</title>
  </head>
  <body>
    <form action="/submit_match_result" method="post">
      <div><p>WINNER</p><textarea name="winner" rows="1" cols="10"></textarea></div>
      <div><p>LOSER</p><textarea name="loser" rows="1" cols="10"></textarea></div>
      <div><p>NOTE</p><textarea name="note" rows="3" cols="20"></textarea></div>
      <div><input type="submit" value="Add a match result"></div>
    </form>
  </body>
</html>
`

// [START add_match_result]
func add_match_result(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, addMatchForm)
}

// [START submit_match_result]
func submit_match_result(w http.ResponseWriter, r *http.Request) {
        // [START new_context]
        c := appengine.NewContext(r)
        // [END new_context]
        g := Match{
                Tournament: "Default",
                Winner: r.FormValue("winner"),
                Loser: r.FormValue("loser"),
                Note: r.FormValue("note"),
                Date:    time.Now(),
        }
        // [START if_user]
        if u := user.Current(c); u != nil {
                g.Submitter= u.String()
        }
        // We set the same parent key on every Greeting entity to ensure each Greeting
        // is in the same entity group. Queries across the single entity group
        // will be consistent. However, the write rate to a single entity group
        // should be limited to ~1/second.
        key := datastore.NewIncompleteKey(c, "Match", guestbookKey(c))
        _, err := datastore.Put(c, key, &g)
        if err != nil {
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
        query_greeting := datastore.NewQuery("Greeting").Ancestor(guestbookKey(c)).Order("-Date").Limit(10)
        // [END query]
        // [START getall]
        greetings := make([]Greeting, 0, 10)
        if _, err := query_greeting.GetAll(c, &greetings); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        // [END getall]

        // [START query]
        query_match := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Order("-Date").Limit(10)
        // [END query]
        // [START getall]
        matches := make([]Match, 0, 10)
        if _, err := query_match.GetAll(c, &matches); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        // [END getall]

        vars := RootPageVars {
                Greetings: greetings,
                Matches: matches,
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
  </head>
  <body>
    <form action="/register">
        <input type="submit" value="Add a new user" />
    </form>
    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Add new comment"></div>
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
    <form action="/add">
        <input type="submit" value="Add new match result" />
    </form>
    {{range .Matches}}
      <p>
      {{.Date}}
      {{with .Submitter}}
        <b>{{.}}</b> submitted:
      {{else}}
        An anonymous person submitted:
      {{end}}
      {{.Winner}} W {{.Loser}} L {{.Note}}
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
