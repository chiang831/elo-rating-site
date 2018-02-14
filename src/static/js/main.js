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
  // Get available matches  from JSON API.
  httpGetAsync(location.origin + "/request_match_data", fillInMatchData);
}


function getMatches() {
  console.log("get matches")
  // Get available matches  from JSON API.
  httpGetAsync(location.origin + "/request_matches", fillInMatches);
}

function getGreetings() {
  console.log("get greetings")
  // Get available matches  from JSON API.
  httpGetAsync(location.origin + "/request_greetings", fillInGreetings);
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
  // userData is originally sorted by name
  userData.sort(function(a, b){ return b.Rating - a.Rating; })
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
  var matchToShows = JSON.parse(r);
  var matches_div = document.getElementById("matches");
  var content = "";
  for (var i in matchToShows) {
    expected = matchToShows[i].Expected;
    match = matchToShows[i].Match;
    var result = "<h3>" +
                 match.Winner + " (" + match.WinnerRatingBefore +
                 " <font color=\"green\">&#x27a8;</font> " + 
                 match.WinnerRatingAfter + ") " +
                 (expected? " &#9876; " : " &#x1F525; ") +
                 match.Loser + " (" + match.LoserRatingBefore +
                 " <font color=\"red\">&#x27a8;</font> " +
                 match.LoserRatingAfter + ") " +
                 match.Note + "</h3>";
    var log = "( Submitted by " + match.Submitter + " @ " + match.Date + " )";
    content += "<div><div class=\"Match\">" + result + log + "</div></div>";
  }
  matches_div.innerHTML = content;
}

function fillInGreetings(r) {
  var greetings = JSON.parse(r);
  var greetings_div = document.getElementById("greetings");
  var content = "";
  for (var i in greetings) {
    greeting = greetings[i];
    var message = "<b>" + greeting.Author + "</b> wrote:" +
                  "<h3>" + greeting.Content + "</h3>" +
                  "( timestamp: " + greeting.Date + " )";
    content += "<div><div class=\"Greeting\">" + message + "</div><div>";
  }
  greetings_div.innerHTML = content;
}
