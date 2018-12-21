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

function initializePage() {
  if (users == null || tournaments == null) {
    return;
  }

  tournamentSelector = new Vue({
    el: '#tournament_selector',
    data: function () {
      return {
        options: tournamentsNames,
        selected: ""
      }
    },

    methods: {
      onChange: function (value) {
        console.log("Tournament " + value + "is selected")
      }
    }
  })

  userSelector = new Vue({
    el: '#user_selector',
    data: function () {
      return {
        options: userNames,
        selected: ""
      }
    }
  });

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

  playerRankingList.ranking.push(userSelector.selected);
}