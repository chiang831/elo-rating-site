function httpGetAsync(theUrl, callback)
{
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.onreadystatechange = function() {
    if (xmlHttp.readyState == 4) {
      if (xmlHttp.status == 200)
        callback(xmlHttp.responseText);
      else if (xmlHttp.status == 401)
        alert("You are not admin QQ");
    }
  }
  xmlHttp.open("GET", theUrl, true); // true for asynchronous
  xmlHttp.send(null);
}

function onLoad() {
  getLeaderboard();
  getDetailMatchResult();
  getRecentMatches();
  getGreetings();
}

function getLeaderboard() {
  httpGetAsync(location.origin + "/request_user_profiles", fillInLeaderboard);
}

function getDetailMatchResult() {
  httpGetAsync(location.origin + "/request_detail_results", fillInDetailMatchResult);
}

function getRecentMatches() {
  var num_matches = document.getElementById("num_matches").value;
  httpGetAsync(location.origin + "/request_recent_matches?num=" + num_matches, fillInRecentMatches);
}

function getGreetings() {
  var num_greeting = document.getElementById("num_greetings").value;
  httpGetAsync(location.origin + "/request_greetings?num=" + num_greeting, fillInGreetings);
}

function fillInLeaderboard(r) {
  var users = JSON.parse(r);
  var leaderboard_table = document.getElementById("leaderboard");
  var content = "<tr>" +
                "<th>Player</th>" +
                "<th><a href=\"https://en.wikipedia.org/wiki/Elo_rating_system\">ELO Rating</a></th>" +
                "<th>Wins</th>" +
                "<th>Losses</th>" +
                "<th>Badges</th>" +
                "</tr>";
  for (var i in users) {
    user = users[i];
    var badge_imgs = "";
    for (var j in user.Badges) {
      badge_imgs += "<img src=\"" + user.Badges[j].Path + "\" " +
                    "title=\"" + user.Badges[j].Description + "\" width=16 height=16></img>";
    }
    var row = "<tr>" +
              "<td><a href=\"/profile?user=" + user.Name + "\">" + user.Name + "</a></td>" +
              "<td>" + Math.round(user.Rating) + "</td>" +
              "<td>" + user.Wins + "</td>" +
              "<td>" + user.Losses + "</td>" +
              "<td>" + badge_imgs + "</td>" +
              "</tr>";
    content += row;
  }
  leaderboard_table.innerHTML = content;
}

function fillInDetailMatchResult(r) {
  var matchData = JSON.parse(r);
  var usernames = matchData.Usernames;
  var resultTable = matchData.ResultTable;
  var detail_result_table = document.getElementById("detail_result");
  // header
  var content = "<tr><td></td>";
  for (var i in usernames) {
    content += ("<td>" + usernames[i] + "</td>");
  }
  content += "</tr>";
  // rows
  for (var i in resultTable) {
    resultRow = resultTable[i];
    var row = "<tr><td>" + usernames[i] + "</td>";
    for (var j in resultRow) {
      resultEntry = resultRow[j];
      row += "<td style=\"background-color:" + resultEntry.Color + "\">" +
             resultEntry.Wins + " / " + resultEntry.Losses + "</td>";
    }
    row += "</tr>";
    content += row;
  }
  detail_result_table.innerHTML = content;
}

function fillInRecentMatches(r) {
  var matchWithKeys = JSON.parse(r);
  if (matches.length == 0) return;
  var matches_div = document.getElementById("matches");
  var content = "";
  for (var i in matchWithKeys) {
    match = matchWithKeys[i].Match;
    key = matchWithKeys[i].Key
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
    var edit_div_str = "<div id=" + key + " style=\"display:none\">" +
                       "<input type=\"button\" value=\"Delete\"" +
                       "onclick=confirmDelete('" + key + "')></input>" +
                       "<input type=\"button\" value=\"Switch\" style=\"margin-left:10px\"" +
                       "onclick=confirmSwitch('" + key + "')></input>" +
                       "</div>";
    var message_div_str = "<div class=\"Match\" onclick=\"show_hide('" + key + "')\">" +
                          result + log +
                          edit_div_str +
                          "</div>";
    var match_div_str = "<div>" + message_div_str + "</div>";
    content += match_div_str;
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

// Callback function for editing matches
function refreshData(ret) {
  console.log(JSON.parse(ret));
  getLeaderboard();
  getDetailMatchResult();
  getRecentMatches();
}

function confirmDelete(key) {
  if (confirm("Are you sure to delete this match?")) {
    httpGetAsync(location.origin + "/delete_match_entry?key=" + key, refreshData);
  }
}

function confirmSwitch(key) {
  if (confirm("Are you sure to switch winner/loser of this match?")) {
    httpGetAsync(location.origin + "/switch_match_users?key=" + key, refreshData);
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
