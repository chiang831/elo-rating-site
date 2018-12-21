Vue.component('v-select', VueSelect.VueSelect);

var users = null;
var user_names = null;
var user_selector = null;

var tournaments = null;
var tournament_names = null;
var tournament_selector = null;

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
  user_names = users.map(u => u.Name);
  console.log("user names = " + user_names);
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
  tournaments_names = tournaments.map(t => t.Name);
  console.log("tournament names = " + tournaments_names);
  initializePage();
}

function initializePage() {
  if (users == null || tournaments == null) {
    return;
  }

  tournament_selector = new Vue({
    el: '#tournament_selector',
    data: function() {
      return {
        options: tournaments_names,
        selected: ""
      }
    },

    methods: {
      onChange: function(value) {
        console.log("Tournament " + value + "is selected")
      }
    }
  })

  v_w = new Vue({
    el: '#user_selector',
    data: function() {
        return {
          names: user_names,
          selected: ""
       }
    }
  });
}
