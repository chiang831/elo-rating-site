package guestbook

import (
        "time"
)

type Greeting struct {
        Author  string
        Content string
        Date    time.Time
}

type Match struct {
        Tournament string
        Submitter  string
        Winner     string
        Loser      string
        WinnerRatingBefore float64
        WinnerRatingAfter float64
        LoserRatingBefore float64
        LoserRatingAfter float64
        Expected   bool
        Note       string
        Date       time.Time
}

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
