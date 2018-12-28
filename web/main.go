package main

// [START import]
import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/chiang831/elo-rating-site/guestbook"
)

// [END import]
// [START main_func]

func main() {
	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	// Main page
	http.HandleFunc("/", guestbook.HandleRoot)
	// Child pages
	http.HandleFunc("/admin", guestbook.HandleAdmin)
	http.HandleFunc("/add_user", guestbook.HandleAddUser)
	http.HandleFunc("/tournament", guestbook.HandleTournaments)
	http.HandleFunc("/tournament/", guestbook.HandleTournamentStats)
	http.HandleFunc("/add_match_result", guestbook.HandleAddMatchResult)
	http.HandleFunc("/add_ffa_match_result", guestbook.HandleAddFfaMatchResult)
	http.HandleFunc("/profile", guestbook.HandleProfile)

	// Submit data
	http.HandleFunc("/submit_greeting", guestbook.HandleSubmitGreeting)
	http.HandleFunc("/submit_user", guestbook.HandleSubmitUser)
	http.HandleFunc("/submit_match_result", guestbook.HandleSubmitMatchResult)
	http.HandleFunc("/submit_badge", guestbook.HandleSubmitBadge)
	http.HandleFunc("/submit_user_badge", guestbook.HandleSubmitUserBadge)
	http.HandleFunc("/submit_tournament", guestbook.HandleSubmitTournament)
	http.HandleFunc("/submit_ffa_match_result", guestbook.HandleSubmitFfaMatchResult)

	// Requests
	http.HandleFunc("/request_users", guestbook.HandleRequestUsers)
	http.HandleFunc("/request_latest_match", guestbook.HandleRequestLatestMatch)
	http.HandleFunc("/request_user_profiles", guestbook.HandleRequestUserProfiles)
	http.HandleFunc("/request_tournament_stats", guestbook.HandleRequestTournamentStats)
	http.HandleFunc("/request_detail_results", guestbook.HandleRequestDetailMatchResults)
	http.HandleFunc("/request_greetings", guestbook.HandleRequestGreetings)
	http.HandleFunc("/request_recent_matches", guestbook.HandleRequestRecentMatches)
	http.HandleFunc("/request_user_matches", guestbook.HandleRequestUserMatches)
	http.HandleFunc("/request_all_badges", guestbook.HandleRequestAllBadges)
	http.HandleFunc("/request_user_badges", guestbook.HandleRequestUserBadges)
	http.HandleFunc("/request_tournaments", guestbook.HandleRequestTournaments)

	// Admin area
	http.HandleFunc("/delete_match_entry", guestbook.HandleDeleteMatchEntry)
	http.HandleFunc("/switch_match_users", guestbook.HandleSwitchMatchUsers)
	http.HandleFunc("/rerun", guestbook.HandleRerunMatches)
	// Static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	// [END setting_port]
}

// [END main_func]
