Vue.component('v-select', VueSelect.VueSelect);

var users = null;
var userNames = null;

var pageInitialized = false;

var currentTournament = getTournamentNameFromURL();

function onLoad() {
  requestUsers();
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

function getTournamentNameFromURL() {
  // Expected URL is "http://..../tournament/<name>/add_ffa_match_result"
  tokens = window.location.href.split("/");
  return tokens[tokens.length - 2];
}

function initializePage() {
  if (users == null) {
    return;
  }

  winnerSelector = new Vue({
    el: '#winner_selector',
    data: function () {
      return {
        options: userNames,
        selected: null
      }
    },
    methods: {
      selectedChanged : function (value) {
        this.selected = value;
      }
    }
  });

  loserSelector = new Vue({
    el: '#loser_selector',
    data: function () {
      return {
        options: userNames,
        selected: null
      }
    },
    methods: {
      selectedChanged : function (value) {
        this.selected = value;
      }
    }
  });

  document.getElementById('container').style.display = 'block';
  pageInitialized = true;
}

function submitResult() {
  if (!pageInitialized) {
    return;
  }

  // Check if a winner is selected
  if (winnerSelector.selected == null) {
    alert("You must select a winner!");
    return;
  }

  // Check if a lower is selected
  if (loserSelector.selected == null) {
    alert("You must select a loser!");
    return;
  }

  // Check if winner == loser
  if (winnerSelector.selected == loserSelector.selected) {
    alert("You must select two different players!");
    return;
  }

  var matchResult = {
    // Fields must start with capital letters to fit golang requirement
    Tournament: currentTournament,
    Players: [winnerSelector.selected, loserSelector.selected],
    Draws: [false]
  };

  console.log("Preparing match result: ");
  console.log(matchResult);

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
