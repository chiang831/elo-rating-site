package guestbook

import (
	"net/http"
	"path"
)

// [START add_tournament]
func addTournament(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("static", "add_tournament.html"))
}
