Vue.component('v-select', VueSelect.VueSelect);

var users = null;
var userNames = null;
var userSelector = null;

var tournaments = null;
var tournamentNames = null;
var tournamentSelector = null;

var playerRankingList = null;

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
  // Expected URL is "http://..../tournament/<name>/add_ffa_match_result"
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
    data: function () {
      return {
        options: tournamentsNames,
        selected: currentTournament
      }
    }
  })

  userSelector = new Vue({
    el: '#user_selector',
    data: function () {
      return {
        options: userNames,
        selected: null
      }
    }
  });

  playerRankingList = new Vue({
    el: '#players',
    data: {
      players: [],
      draws: []
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

  // Check if player is selected
  if (userSelector.selected == null) {
    alert("You must select a player!");
    return;
  }

  // Check if the player is already added into the ranking
  if (playerRankingList.players.includes(userSelector.selected)) {
    alert("Player " + userSelector.selected + " is already in ranking list!");
    return;
  }

  playerRankingList.players.push(userSelector.selected);
}

function submitResult() {
  if (!pageInitialized) {
    return;
  }

  var matchResult = {
    // Fields must start with capital letters to fit golang requirement
    Tournament: tournamentSelector.selected,
    Players: playerRankingList.players,
    Draws: playerRankingList.draws
  };

  if (matchResult.Players.length < 2) {
    alert("At least 2 players are required in a FFA ranking!");
    return;
  }

  var confirmMsg = "Submitting result: ";
  for (const [i, player] of matchResult.Players.entries()) {
    confirmMsg += player;
    if (i < matchResult.Draws.length) {
      if (matchResult.Draws[i]) {
        confirmMsg += " = ";
      } else {
        confirmMsg += " > ";
      }
    }
  }

  if (window.confirm(confirmMsg) == false) {
    return;
  }

  httpPostJsonAsync(
    location.origin + "/submit_ffa_match_result",
    matchResult,
    function (responseText) {
      // redirect to tournament stats page
      window.location.href = "/tournament/" + matchResult.Tournament;
    });
}