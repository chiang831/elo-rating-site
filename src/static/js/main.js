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
  getMatchData();
  getMatches();
  getGreetings();
}

function getMatchData() {
  console.log("get match data")
  // Get available match data from JSON API.
  httpGetAsync(location.origin + "/request_match_data", fillInMatchData);
}


function getMatches() {
  console.log("get matches")
  // Get available matches from JSON API.
  var num_matches = document.getElementsByName("num_matches")[0].value;
  httpGetAsync(location.origin + "/request_matches?num=" + num_matches, fillInMatches);
}

function getGreetings() {
  console.log("get greetings")
  // Get available greetings from JSON API.
  var num_greeting = document.getElementsByName("num_greetings")[0].value;
  httpGetAsync(location.origin + "/request_greetings?num=" + num_greeting, fillInGreetings);
}

function fillInMatchData(r) {
  var matchData = JSON.parse(r);
  fillInLeaderboard(matchData.UserDataToShows);
  fillInDetailMatchResult(matchData.DetailMatchResults);
}

function fillInLeaderboard(userData) {
  var leaderboard_table = document.getElementById("leaderboard");
  var content = "<tr>" +
                "<th>Player</th>" +
                "<th><a href=\"https://en.wikipedia.org/wiki/Elo_rating_system\">ELO Rating</a></th>" +
                "<th>Wins</th>" +
                "<th>Losses</th>" +
                "</tr>";
  for (var i in userData) {
    user = userData[i];
    var row = "<tr>" +
              "<td>" + user.Name + "</td>" +
              "<td>" + user.Rating + "</td>" +
              "<td>" + user.Wins + "</td>" +
              "<td>" + user.Losses + "</td>" +
              "</tr>";
    content += row;
  }
  leaderboard_table.innerHTML = content;
}

function fillInDetailMatchResult(results) {
  var detail_result_table = document.getElementById("detail_result");
  // header
  var content = "<tr><td></td>";
  for (var i in results) {
    content += ("<td>" + results[i].Name + "</td>");
  }
  content += "</tr>";
  // rows
  for (var i in results) {
    resultRow = results[i].Results;
    var row = "<tr><td>" + results[i].Name + "</td>";
    for (var j in resultRow) {
      resultEntry = resultRow[j];
      row += "<td bgcolor=" + resultEntry.Color + ">" +
             resultEntry.Wins + " / " + resultEntry.Losses + "</td>";
    }
    row += "</tr>";
    content += row;
  }
  detail_result_table.innerHTML = content;
}

function fillInMatches(r) {
  var matches = JSON.parse(r);
  if (matches.length == 0) return;
  var matches_div = document.getElementById("matches");
  var content = "";
  for (var i in matches) {
    match = matches[i];
    var result = "<h3>" +
                 match.Winner + " (" + Math.round(match.WinnerRatingBefore) +
                 " <font color=\"green\">&#x27a8;</font> " + 
                 Math.round(match.WinnerRatingAfter) + ") " +
                 (match.Expected? " &#9876; " : " &#x1F525; ") +
                 match.Loser + " (" + Math.round(match.LoserRatingBefore) +
                 " <font color=\"red\">&#x27a8;</font> " +
                 Math.round(match.LoserRatingAfter) + ") " +
                 match.Note + "</h3>";
    var log = "( Submitted by " + getName(match.Submitter) + " @ " + getTime(match.Date) + " )";
    content += "<div><div class=\"Match\">" + result + log + "</div></div>";
  }
  matches_div.innerHTML = content;
}

function fillInGreetings(r) {
  var greetings = JSON.parse(r);
  if (greetings.length == 0) return;
  var greetings_div = document.getElementById("greetings");
  var content = "";
  for (var i in greetings) {
    greeting = greetings[i];
    var message = "<b>" + getName(greeting.Author) + "</b> wrote:" +
                  "<h3>" + greeting.Content + "</h3>" +
                  "( Timestamp: " + getTime(greeting.Date) + " )";
    content += "<div><div class=\"Greeting\">" + message + "</div><div>";
  }
  greetings_div.innerHTML = content;
}

function show_hide(id) {
  var target = document.getElementById(id);
  if (target) {
    if (target.style.display == "block") {
      target.style.display = "none";
    } else {
      target.style.display = "block";
    }
  }
}

// Remove everything after '@'
function getName(name) {
  var idx = name.indexOf("@");
  if (idx != -1) {
    return name.substr(0, idx);
  }
  return name;
}

// Transform to local time
function getTime(dateString) {
  var date = new Date(dateString)
  return date.toLocaleString();
}
