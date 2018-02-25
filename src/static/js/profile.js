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

function onLoad(username) {
  getUserMatches(username);
}

function getUserMatches(username) {
  httpGetAsync(location.origin + "/request_user_matches?user=" + username, processUserMatches(username));
}

function processUserMatches(username) {
  return function(r) {
    var userMatches = JSON.parse(r);
    if (userMatches.length == 0) {
      return;
    }
    var ratings = [getUserRatingBefore(username, userMatches[0])];
    for (var i in userMatches) {
      ratings.push(getUserRatingAfter(username, userMatches[i]));
    }
    d3DrawRating(ratings);
    fillInUserMatches(username, userMatches);
  }
}

function getUserRatingBefore(username, match) {
  if (match.Winner == username) {
    return Math.round(match.WinnerRatingBefore);
  } else if (match.Loser == username) {
    return Math.round(match.LoserRatingBefore);
  }
  // should not get here
  console.log("User does not play this natch.");
  return -1;
}

function getUserRatingAfter(username, match) {
  if (match.Winner == username) {
    return Math.round(match.WinnerRatingAfter);
  } else if (match.Loser == username) {
    return Math.round(match.LoserRatingAfter);
  }
  // should not get here
  console.log("User does not play this natch.");
  return -1;
}

function fillInUserMatches(username, matches) {
  var matches_div = document.getElementById("user_matches");
  var content = "";
  for (var i = matches.length - 1; i >= 0; --i) {
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
    var log = "( Timestamp: " + getTime(match.Date) + " )";
    var message_div_str = "<div class=\"Match\" style=\"background-color:" + getColor(username, match) + "\">" + result + log + "</div>";
    var match_div_str = "<div>" + message_div_str + "</div>";
    content += match_div_str;
  }
  matches_div.innerHTML = content;
}

// Decide win/lose display color
function getColor(username, match) {
  if (match.Winner == username) {
    return "honeydew";
  } else if (match.Loser == username) {
    return "seashell";
  }
  // should not get here
  console.log("User does not play this natch.");
  return "white";
}

function d3DrawRating(ratings) {
  // Setup
  var graph_div = document.getElementById("rating_history");
  var graph_pos = graph_div.getBoundingClientRect();
  var margin = {left: 50, top: 20, right: 40, bottom: 30};  // left, top, right, bottom
  var w = graph_pos.width - margin.left - margin.right;
  var h = graph_pos.height - margin.top - margin.bottom;
  // x-axis and y-axis
  var x = d3.scale.linear().domain([0, ratings.length]).range([0, w]);
  var minRating = d3.min(ratings);
  var maxRating = d3.max(ratings);
  var y = d3.scale.linear().domain([minRating - 20, maxRating + 20]).range([h, 0]);
  // Draw axis and line
  var line = d3.svg.line()
                   .x(function(d, i) { return x(i); })
	           .y(function(d) { return y(d); });
  var graph = d3.select("#rating_history")
                .append("svg:svg")
		.attr("width", graph_pos.width)
		.attr("height", graph_pos.height)
		.append("svg:g")
                .attr("transform", "translate(" + margin.left + "," + margin.top + ")");
  var xAxis = d3.svg.axis().scale(x).ticks(5).tickSize(2).orient("bottom");
  graph.append("svg:g")
       .attr("class", "x axis")
       .attr("transform", "translate(0, " + h + ")")
       .call(xAxis);
  var yAxis = d3.svg.axis().scale(y).ticks(5).tickSize(2).orient("left");
  graph.append("svg:g")
       .attr("class", "y axis")
       .attr("transform", "translate(0, 0)")
       .call(yAxis);
  graph.append("svg:path")
       .attr("fill", "none")
       .attr("stroke", "navy")
       .attr("stroke-linejoin", "round")
       .attr("stroke-linecap", "round")
       .attr("stroke-width", 2)
       .attr("d", line(ratings));
}

// Transform to local time
function getTime(dateString) {
  var date = new Date(dateString)
  return date.toLocaleString();
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

