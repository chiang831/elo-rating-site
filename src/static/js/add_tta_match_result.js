Vue.component('v-select', VueSelect.VueSelect);

var users = null;
var userNames = null;

var tournaments = null;
var tournamentNames = null;
var tournamentSelector = null;

var playerRankingList = null;

var playerOne = null;
var playerOneScore = null;
var playerTwo = null;
var playerTwoScore = null;
var playerThree = null;
var playerThreeScore = null;
var playerFour = null;
var playerFourScore = null;

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
      placeholder: ["required", "required", "optional", "optional"],
      player: ["", "", "", ""],
      score: ["", "", "", ""],
      options: userNames
    },
  })


  playerRankingList = new Vue({
    el: '#ranking',
    data: {
      ranking: []
    }
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
  alert("Not implemented! Currently we push the users and scores we got in Player Ranking to show we got them correctly.");
  playerRankingList.ranking = [];
  playerRankingList.ranking.push(matchTable.player);
  playerRankingList.ranking.push(matchTable.score);
}

function submitRanking() {
  if (!pageInitialized) {
    return;
  }

  var matchResult = {
    // Fields must start with capital letters to fit golang requirement
    Tournament: tournamentSelector.selected,
    Ranking: playerRankingList.ranking
  };

  if (matchResult.Ranking.length < 2) {
    alert("At least 2 players are required in a FFA ranking!");
    return;
  }

  httpPostJsonAsync(
    location.origin + "/submit_ffa_match_result",
    matchResult,
    function(responseText) {
      window.location.href = "/tournament/" + matchResult.Tournament;
    });
}
