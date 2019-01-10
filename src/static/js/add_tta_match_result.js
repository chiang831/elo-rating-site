Vue.component('v-select', VueSelect.VueSelect);

var userNames = null;
var userSelector = null;

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

function initializePage() {
<<<<<<< HEAD
  if (users == null || tournaments == null) {
    return;
  }

  tournamentSelector = new Vue({
    el: '#tournament_selector',
    data: function() {
      return {
        options: tournamentsNames,
        selected: null
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


  userSelector = new Vue({
    el: '#user_selector',
    data: function() {
      return {
        options: userNames,
        selected: null
      }
=======
    if (users == null || tournaments == null) {
        return;
    }

    tournamentSelector = new Vue({
        el: '#tournament_selector',
        data: function() {
            return {
                options: tournamentsNames,
                selected: null
            }
        }
    })

    playerOne = new Vue({
        el: '#player_one',
        data: function() {
            return {
                value: '',
                options: userNames
            }
        }
    })

    playerOneScore = new Vue({
        el: '#player_one_score',
        data: function() {
            return {
                value: null
            }
        }
    })

    playerTwo = new Vue({
        el: '#player_two',
        data: function() {
            return {
                value: '',
                options: userNames
            }
        }
    })

    playerTwoScore = new Vue({
        el: '#player_two_score',
        data: function() {
            return {
                value: null
            }
        }
    })

    playerThree = new Vue({
        el: '#player_three',
        data: function() {
            return {
                value: '',
                options: userNames
            }
        }
    })

    playerThreeScore = new Vue({
        el: '#player_three_score',
        data: function() {
            return {
                value: null
            }
        }
    })

    playerFour = new Vue({
        el: '#player_four',
        data: function() {
            return {
                value: '',
                options: userNames
            }
        }
    })

    playerFourScore = new Vue({
        el: '#player_four_score',
        data: function() {
            return {
                value: null
            }
        }
    })

    userSelector = new Vue({
        el: '#user_selector',
        data: function() {
            return {
                options: userNames,
                selected: null
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

    // Check if tournament is selected
    if (tournamentSelector.selected == null) {
        alert("You must select a tournament!");
        return;
>>>>>>> 3b364bb... Adds a table capturing user and score in a match.
    }

    // Check if player is selected
    if (userSelector.selected == null) {
        alert("You must select a player!");
        return;
    }

    // Check if the player is already added into the ranking
    if (playerRankingList.ranking.includes(userSelector.selected)) {
        alert("Player " + userSelector.selected + " is already in ranking list!");
        return;
    }

    playerRankingList.ranking.push(userSelector.selected);
}

function preview() {
    alert("Not implemented! Currently we push the users and scores we got in Player Ranking to show we got them correctly.");
    playerRankingList.ranking = [];
    playerRankingList.ranking.push(playerOne.value);
    playerRankingList.ranking.push(playerTwo.value);
    playerRankingList.ranking.push(playerThree.value);
    playerRankingList.ranking.push(playerFour.value);
    playerRankingList.ranking.push(playerOneScore.value);
    playerRankingList.ranking.push(playerTwoScore.value);
    playerRankingList.ranking.push(playerThreeScore.value);
    playerRankingList.ranking.push(playerFourScore.value);
}

function preview() {
  alert("Not implemented! Currently we push the users and scores we got in Player Ranking to show we got them correctly.");
  playerRankingList.ranking = [];
  playerRankingList.ranking.push(matchTable.player);
  playerRankingList.ranking.push(matchTable.score);
}

function submitRanking() {
<<<<<<< HEAD
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
=======
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
>>>>>>> 3b364bb... Adds a table capturing user and score in a match.
