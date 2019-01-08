package guestbook

import (
	"net/http"
	"path"
)

func showAddTtaMatchResult(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, path.Join("static", "add_tta_match_result.html"))
}

