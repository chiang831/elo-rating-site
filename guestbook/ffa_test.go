package guestbook

import "testing"

func TestGenerate1v1MatchResults(t *testing.T) {
	// wanted match results generated for 4 player game
	//  * 2nd > 3rd (index 1, 2)
	//  * 1st > 3rd (index 0, 2)
	//  * 2nd > 4th (index 1, 3)
	//  * 1st > 2nd (index 0, 1)
	//  * 3rd > 4th (index 2, 3)
	wantedResults := []MathcResult1v1{
		New1v1MatchResult(1, 2, 1.5),
		New1v1MatchResult(0, 2, 1.5),
		New1v1MatchResult(1, 3, 1.5),
		New1v1MatchResult(0, 1, 1.5),
		New1v1MatchResult(2, 3, 1.5),
	}

	results := Generate1v1MatchResults(4)

	if len(results) != len(wantedResults) {
		t.Errorf("Wanted number of matches: %d, got number of matches: %d", len(results), len(wantedResults))
		return
	}

	for i, wantedResult := range wantedResults {
		gotResult := &results[i]

		if gotResult.winner != wantedResult.winner || gotResult.loser != wantedResult.loser {
			t.Errorf("Wanted result on index %d is {%d, %d}, but got {%d, %d}",
				i,
				wantedResult.winner, wantedResult.loser,
				gotResult.winner, gotResult.loser)
		}
	}
}
