package guestbook

import (
	"net/http"
	"path"
)

// [START add_tournament]
func showAddFfaMatchResult(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("static", "add_ffa_match_result.html"))
}
