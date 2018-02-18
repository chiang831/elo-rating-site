Vue.component('v-select', VueSelect.VueSelect);

var v_w = null;
var v_l = null;
var vue_created = false;
var users;

function createVueElements(r) {
  users = JSON.parse(r);
  console.log("users = " + users);
  var names = users.map(u => u.Name);
  console.log("names = " + names);
  v_w = new Vue({
    el: '#winner_select',
    data: function() {
        return {
          names: names,
          selected: ""
       }
    },

    methods: {
      selectedChanged : function (value) {
        this.selected = value;
        updateCalculator();
      }
    }
  });

  v_l = new Vue({
    el: '#loser_select',
    data: function() {
        return {
          names: names,
          selected: ""
       }
    },

    methods: {
      selectedChanged : function (value) {
        this.selected = value;
        updateCalculator();
      }
    }
  });
  vue_created = true;
}

function httpGetAsync(theUrl, callback)
{
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.onreadystatechange = function() {
      if (xmlHttp.readyState == 4 && xmlHttp.status == 200)
          callback(xmlHttp.responseText);
  }
  xmlHttp.open("GET", theUrl, true); // true for asynchronous
  xmlHttp.send(null);
}

function onLoad() {
  getAlert();
  getUsers();
}

function getUsers () {
  console.log("get users");
  // Get available user data from JSON API.
  httpGetAsync(location.origin + "/users", createVueElements);
}

function getAlert() {
  console.log("get message");
  // Get available message from JSON API.
  httpGetAsync(location.origin + "/latest_match", fillInMessage);
}

function fillInMessage(r) {
  var lm = JSON.parse(r);
  console.log("latest_match = " + lm);
  if (lm != null ) {
    var message_div = document.getElementById("message");
    var message = "Latest match:<br>";
    message += lm.Winner + " (" + Math.round(lm.WinnerRatingBefore) + " &#x27a8; " + Math.round(lm.WinnerRatingAfter) + ") ";
    message += "<br>";
    if (lm.Expected) {
        message += " &#9876; ";
    } else {
        message += " &#x1F525; ";
    }
    message += "<br>";
    message += lm.Loser + " (" + Math.round(lm.LoserRatingBefore) + " &#x27a8; " + Math.round(lm.LoserRatingAfter) + ") ";
    console.log("message = " + message);

    message_div.innerHTML = message;
    message_div.style.display = "block";

    // Best effort to set the default value.
    // Otherwise, need to wait for vue elements created.
    setTimeout(
      function() {
        if (v_w != null && v_l != null && vue_created == true) {
          v_w.selected = lm.Winner;
          v_l.selected = lm.Loser;
        }
      }, 500
    );
  }
}

function switchSelected() {
  var temp = v_w.selected;
  v_w.selected = v_l.selected;
  v_l.selected = temp;
}

function updateCalculator() {
  // Names selected can be null or "". Ignore such cases.
  // Default: "". Cleared: null.
  if (!v_w.selected || !v_l.selected) {
    winner_rating_elem = document.getElementById("winner_rating");
    loser_rating_elem = document.getElementById("loser_rating");
    winner_rating_elem.innerHTML = "";
    loser_rating_elem.innerHTML = "";
    prob_elem = document.getElementById("exp_win_prob");
    prob_elem.innerHTML = "";
    return;
  }

  // Check the names.
  console.log('winner is ' + v_w.selected);
  console.log('loser is ' + v_l.selected);

  // Find element by Name.
  winner = users.find(elem => elem.Name == v_w.selected);
  loser = users.find(elem => elem.Name == v_l.selected);

  console.log('winner rating before matching : ' + winner.Rating);
  console.log('loser rating before matching: ' + loser.Rating);

  // Compute expected win rate.
  expected_w = expectedScore(winner.Rating, loser.Rating);
  console.log('expected winner win rate: ' + expected_w);
  expected_l = 1 - expected_w;
  console.log('expected loser win rate: ' + expected_l);

  // Assume winner wins the game.
  winner_rating_diff = diffElo(expected_w, 1.0);

  // Assume loser loses the game.
  loser_rating_diff = diffElo(expected_l, 0.0);

  console.log('winner rating diff after matching : ' + Math.round(winner_rating_diff));
  console.log('loser rating diff after matching: ' + Math.round(loser_rating_diff));

  winner_rating_elem = document.getElementById("winner_rating");
  loser_rating_elem = document.getElementById("loser_rating");
  winner_rating_elem.innerHTML = winner.Name + " (" + Math.round(winner.Rating) +
                 " <font color=\"green\">" + " + " + Math.round(winner_rating_diff) + " </font> ) ";
  loser_rating_elem.innerHTML = loser.Name + " (" + Math.round(loser.Rating) +
                 " <font color=\"red\">" + " - " + Math.abs(Math.round(loser_rating_diff)) + " </font> ) ";
  prob_elem = document.getElementById("exp_win_prob");
  console.log('prob_elem = ' + prob_elem);
  prob_elem.innerHTML = (expected_w * 100).toFixed(1) + '%';
}

// Expected score of elo_a in a match against elo_b
function expectedScore(elo_a, elo_b) {
    return 1 / (1 + Math.pow(10, (elo_b - elo_a) / 400.0));
}

// Get the diff Elo rating.
function diffElo(expected, score) {
    return 32.0 * (score - expected);
}
