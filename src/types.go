package guestbook

import (
	"time"
)

// Greeting message to show in the home page
type Greeting struct {
	Author  string
	Content string
	Date    time.Time
}

// Match wrapper struct for datastore
type Match struct {
	Tournament         string
	Submitter          string
	Winner             string
	Loser              string
	WinnerRatingBefore float64
	WinnerRatingAfter  float64
	LoserRatingBefore  float64
	LoserRatingAfter   float64
	Expected           bool
	Note               string
	Date               time.Time
}

// MatchWithKey wrapper struct for datastore
type MatchWithKey struct {
	Match Match
	Key   string
}

// UserProfile wrapper for datastore
type UserProfile struct {
	Tournament string
	Name       string
	Rating     float64
	Wins       int
	Losses     int
	JoinDate   time.Time
}

// UserProfileToShow wrapper for datastore
type UserProfileToShow struct {
	Name   string
	Rating float64
	Wins   int
	Losses int
	Badges []Badge
}

// DetailMatchResultEntry wrapper for datastore
type DetailMatchResultEntry struct {
	Wins   int
	Losses int
	Color  string
}

// MatchData wrapper for datastore
type MatchData struct {
	Usernames   []string
	ResultTable [][]DetailMatchResultEntry
}

// Badge wrapper for datastore
type Badge struct {
	Name        string
	Description string
	Author      string
	Path        string
}

// UserBadge wrapper for datastore
type UserBadge struct {
	User       string
	BadgeNames []string
}

// Tournament object in datastore represents a particular tournament
type Tournament struct {
	Name string
}

// UserTournamentStats object in datastore represents an user's performance in a particular tournament
type UserTournamentStats struct {
	UserID       int64
	TournamentID int64
	FFAWins      int
	Wins         int
	Losses       int
	Rating       float64
}
