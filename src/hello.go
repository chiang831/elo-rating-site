package guestbook

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"regexp"
	"sort"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

func init() {
	// Main page
	http.HandleFunc("/", root)
	// Child pages
	http.HandleFunc("/admin", admin)
	http.HandleFunc("/add_user", addUser)
	http.HandleFunc("/tournament", showTournaments)
	http.HandleFunc("/tournament/", showTournamentStats)
	http.HandleFunc("/add_match_result", addMatchResult)
	http.HandleFunc("/add_ffa_match_result", showAddFfaMatchResult)
	http.HandleFunc("/profile", profile)

	// Submit data
	http.HandleFunc("/submit_greeting", submitGreeting)
	http.HandleFunc("/submit_user", submitUser)
	http.HandleFunc("/submit_match_result", submitMatchResult)
	http.HandleFunc("/submit_badge", submitBadge)
	http.HandleFunc("/submit_user_badge", submitUserBadge)
	http.HandleFunc("/submit_tournament", submitTournament)
	http.HandleFunc("/submit_ffa_match_result", submitFfaMatchResult)

	// Requests
	http.HandleFunc("/request_users", requestUsers)
	http.HandleFunc("/request_latest_match", requestLatestMatch)
	http.HandleFunc("/request_user_profiles", requestUserProfiles)
	http.HandleFunc("/request_tournament_stats", requestTournamentStats)
	http.HandleFunc("/request_detail_results", requestDetailMatchResults)
	http.HandleFunc("/request_greetings", requestGreetings)
	http.HandleFunc("/request_recent_matches", requestRecentMatches)
	http.HandleFunc("/request_user_matches", requestUserMatches)
	http.HandleFunc("/request_all_badges", requestAllBadges)
	http.HandleFunc("/request_user_badges", requestUserBadges)
	http.HandleFunc("/request_tournaments", requestTournaments)

	// Admin area
	http.HandleFunc("/delete_match_entry", deleteMatchEntry)
	http.HandleFunc("/switch_match_users", switchMatchUsers)
	http.HandleFunc("/rerun", rerunMatches)
	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
}

// guestbookKey returns the key used for all guestbook entries.
func guestbookKey(c context.Context) *datastore.Key {
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

func existUser(c context.Context, name string) (bool, datastore.Key, UserProfile, error) {
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

func existBadge(c context.Context, name string) (bool, datastore.Key, Badge, error) {
	q := datastore.NewQuery("Badge").Ancestor(guestbookKey(c)).Filter("Name =", name)
	var badges []Badge
	keys, err := q.GetAll(c, &badges)
	if err != nil {
		return false, datastore.Key{}, Badge{}, err
	}
	if len(badges) != 0 {
		return true, *keys[0], badges[0], nil
	}
	return false, datastore.Key{}, Badge{}, nil
}

func getUserBadges(c context.Context, username string) []Badge {
	queryBadge := datastore.NewQuery("UserBadge").Ancestor(guestbookKey(c)).Filter("User =", username)
	var userBadges []UserBadge
	if _, err := queryBadge.GetAll(c, &userBadges); err != nil {
		return []Badge{}
	}

	badges := []Badge{}
	if len(userBadges) != 0 {
		badges = make([]Badge, len(userBadges[0].BadgeNames))
		for i, badgeName := range userBadges[0].BadgeNames {
			exist, _, badge, err := existBadge(c, badgeName)
			if exist && err == nil {
				badges[i] = badge
			} else {
				badges = []Badge{}
				break
			}
		}
	}
	return badges
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
		Name:       name,
		Rating:     startingElo,
		Wins:       0,
		Losses:     0,
		JoinDate:   time.Now(),
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

	js, errJs := json.Marshal(users)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func requestLatestMatch(w http.ResponseWriter, r *http.Request) {
	if !existLatestMatch {
		nilJs, nilErrJs := json.Marshal(nil)
		if nilErrJs != nil {
			http.Error(w, nilErrJs.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(nilJs)
		return
	}

	js, errJs := json.Marshal(latestMatch)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func requestUserProfiles(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// Get users
	queryUser := datastore.NewQuery("UserProfile").Ancestor(guestbookKey(c)).Order("-Rating")
	var users []UserProfile
	if _, err := queryUser.GetAll(c, &users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create public user profile
	userProfileToShows := make([]UserProfileToShow, len(users))
	for i, u := range users {
		// Get badges
		userProfileToShows[i] = UserProfileToShow{
			Name:   u.Name,
			Rating: u.Rating,
			Wins:   u.Wins,
			Losses: u.Losses,
			Badges: getUserBadges(c, u.Name),
		}
	}

	js, errJs := json.Marshal(userProfileToShows)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func requestDetailMatchResults(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	tournamentName := r.FormValue("tournament")
	if tournamentName == "" {
		tournamentName = "Default"
	}

	tournamentKey, err := findExistingTournamentKey(ctx, tournamentName)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user tournament stats
	statsList, err := readAllUserStatsForTournament(ctx, tournamentKey.IntID())
	if err != nil {
		http.Error(w, "Failed to read tournament stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	users := make([]UserProfile, len(statsList))
	for i, stats := range statsList {
		users[i], err = readUserProfile(ctx, stats.UserID)
		if err != nil {
			http.Error(w, "Failed to read user profile: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Get matches
	queryMatch := datastore.NewQuery("Match").Ancestor(guestbookKey(ctx)).Filter("Tournament =", tournamentName)
	var matches []Match
	if _, err := queryMatch.GetAll(ctx, &matches); err != nil {
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
		resultTable[idxW][idxL].Wins++
		resultTable[idxL][idxW].Losses++
	}
	for i := range resultTable {
		for j := range resultTable[i] {
			resultTable[i][j].Color = getColor(users[i], users[j], resultTable[i][j].Wins, resultTable[i][j].Losses)
		}
	}

	// Return JSON
	matchData := MatchData{
		Usernames:   usernames,
		ResultTable: resultTable,
	}

	js, errJs := json.Marshal(matchData)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
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
		newLimit, err := strconv.Atoi(keys[0])
		if err == nil && newLimit > 0 {
			limit = newLimit
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

	js, errJs := json.Marshal(greetings)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
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
	limitParam := r.FormValue("num")
	if limitParam != "" {
		newLimit, err := strconv.Atoi(limitParam)
		if err == nil && newLimit > 0 {
			limit = newLimit
		}
	}

	tournament := r.FormValue("tournament")
	if tournament == "" {
		tournament = "Default"
	}

	matchWithKeys := []MatchWithKey{}
	if limit != -1 {
		queryMatch := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).
			Filter("Tournament = ", tournament).
			Order("-Date").
			Limit(limit)
		var matches []Match
		keyMatches, err := queryMatch.GetAll(c, &matches)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		matchWithKeys = make([]MatchWithKey, len(matches))
		for i, m := range matches {
			matchWithKeys[i] = MatchWithKey{
				Match: m,
				Key:   keyMatches[i].Encode(),
			}
		}
	}

	js, errJs := json.Marshal(matchWithKeys)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func requestUserMatches(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// Get username
	username := ""
	keys, ok := r.URL.Query()["user"]
	if ok && len(keys) == 1 {
		username = keys[0]
	}
	exist, _, _, err := existUser(c, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !exist {
		http.Error(w, "User \""+username+"\" does not exist", http.StatusInternalServerError)
		return
	}
	// Get user matches
	queryMatchW := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Filter("Winner =", username)
	var matchesW []Match
	if _, err := queryMatchW.GetAll(c, &matchesW); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	queryMatchL := datastore.NewQuery("Match").Ancestor(guestbookKey(c)).Filter("Loser =", username)
	var matchesL []Match
	if _, err := queryMatchL.GetAll(c, &matchesL); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Create history
	n := len(matchesW)
	m := len(matchesL)
	allMatches := make([]Match, n+m)
	for i := range matchesW {
		allMatches[i] = matchesW[i]
	}
	for i := range matchesL {
		allMatches[i+n] = matchesL[i]
	}
	// Sort history
	sort.Slice(allMatches, func(i, j int) bool {
		return allMatches[i].Date.Before(allMatches[j].Date)
	})

	js, errJs := json.Marshal(allMatches)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func requestAllBadges(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// Get all badges
	queryBadge := datastore.NewQuery("Badge").Ancestor(guestbookKey(c))
	var badges []Badge
	if _, err := queryBadge.GetAll(c, &badges); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	js, errJs := json.Marshal(badges)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func requestUserBadges(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	// Get username
	username := ""
	keys, ok := r.URL.Query()["user"]
	if ok && len(keys) == 1 {
		username = keys[0]
	}
	exist, _, _, err := existUser(c, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !exist {
		http.Error(w, "User \""+username+"\" does not exist", http.StatusInternalServerError)
		return
	}
	// Get User badges
	badges := getUserBadges(c, username)

	js, errJs := json.Marshal(badges)
	if errJs != nil {
		http.Error(w, errJs.Error(), http.StatusInternalServerError)
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
		ratio := float64(wins-losses) / float64(wins+losses)
		// limegreen: rgb(50, 205, 50)
		// gold: rgb(255, 215, 0)
		r := 50.0*ratio + 255.0*(1.0-ratio)
		g := 205.0*ratio + 215.0*(1.0-ratio)
		b := 50.0*ratio + 0.0*(1.0-ratio)
		return fmt.Sprintf("rgb(%.0f,%.0f,%.0f)", r, g, b)
	} else {
		ratio := float64(2*wins) / float64(wins+losses)
		// gold: rgb(255, 215, 0)
		// tomato: rgb(255, 99, 71)
		r := 255.0*ratio + 255.0*(1.0-ratio)
		g := 215.0*ratio + 99.0*(1.0-ratio)
		b := 0.0*ratio + 71.0*(1.0-ratio)
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
	profilePath := path.Join("static", "profile.html")
	tmpl, tmplErr := template.ParseFiles(profilePath)
	if tmplErr != nil {
		http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
		return
	}
	if err = tmpl.Execute(w, user); err != nil {
		http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
		return
	}
	// http.ServeFile(w, r, path.Join("static", "profile.html"))
}
