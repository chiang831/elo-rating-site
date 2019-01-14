Vue.component('v-select', VueSelect.VueSelect);

var users = null;
var userNames = null;

var tournaments = null;
var tournamentNames = null;
var tournamentSelector = null;

var pageInitialized = false;

function onLoad() {
  requestUsers();
  requestTournaments();
}

function requestUsers() {
  console.log("get users");
  // Get available user data from JSON API.
  httpGetAsync(location.origin + "/request_users", handleUsersResponse);
}

function handleUsersResponse(responseText) {
  users = JSON.parse(responseText);
  console.log("users = " + users);
  userNames = users.map(u => u.Name);
  console.log("user names = " + userNames);
  initializePage();
}

function requestTournaments() {
  console.log("get tournaments");
  // Get available user data from JSON API.
  httpGetAsync(location.origin + "/request_tournaments", handleTournamentsResponse);
}

function handleTournamentsResponse(responseText) {
  tournaments = JSON.parse(responseText);
  console.log("tournaments = " + tournaments);
  tournamentsNames = tournaments.map(t => t.Name);
  console.log("tournament names = " + tournamentsNames);
  initializePage();
}

function getTournamentNameFromURL() {
  // Expected URL is "http://..../tournament/<name>/add_tta_match_result"
  tokens = window.location.href.split("/");
  return tokens[tokens.length - 2];
}

function initializePage() {
  if (users == null || tournaments == null) {
    return;
  }

  var currentTournament = getTournamentNameFromURL();

  tournamentSelector = new Vue({
    el: '#tournament_selector',
    data: function() {
      return {
        options: tournamentsNames,
        selected: currentTournament
      }
    }
  })

  matchTable = new Vue({
    el: "#match_table",
    data: {
      header: ["Player 1", "Player 2", "Player 3", "Player 4"],
      player: ["", "", "", ""],
      score: ["", "", "", ""],
      options: userNames
    },
  })
  rankingTable = new Vue({
    el: "#ranking_table",
    data: {
      player: [],
      ranking: [],
      order: [],
      score: [],
    },
  })


  document.getElementById('container').style.display = 'block';
  pageInitialized = true;
}

function addUser() {
  if (!pageInitialized) {
    return;
  }
  // Check if tournament is selected
  if (tournamentSelector.selected == null) {
    alert("You must select a tournament!");
    return;
  }
}

function preview() {
  var i;
  for (i = 0; i < 2; i++) {
    if (matchTable.player[i] == "" ||
      matchTable.score[i] == ""
    ) {
      alert('Player 1 and Player 2 must be set.');
      return;
    }

  }
  var num_players = 2;
  for (i = 2; i < matchTable.player.length; i++) {
    if (matchTable.player[i] != "" &&
      matchTable.score[i] != "") {
      if (i >= num_players + 1) {

        alert('Player ' + Number(num_players + 1) + ' is not specified but Player ' + Number(i + 1) + ' is. This is an error.!');
        return;

      }
      num_players = i + 1;

    }
    if (
      matchTable.player[i] == "" != matchTable.score[i] == ""
    ) {
      alert('When Player ' + Number(num_players + 1) + ' is set, you must set both User and Score!');
      return;
    }

  }


  const sorted = matchTable.score.slice(0, num_players).sort(
    function(a, b) {
      return b - a
    }
  )
  var rank = matchTable.score.slice(0, num_players).map(x => sorted.indexOf(x) + 1)
  rankingTable.ranking = rank;
  rankingTable.player = matchTable.player.slice(0, num_players);
  rankingTable.score = matchTable.score.slice(0, num_players);
  rankingTable.order = []
  for (i = 0; i < num_players; i++) {
    rankingTable.order.push(i + 1);
  }
}

function submitRanking() {
  if (!pageInitialized) {
    return;
  }
  alert("Not implemented! Your result is not submitted.");
}
