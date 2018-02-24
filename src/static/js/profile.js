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
  getRatingHistory(username);
}

function getRatingHistory(username) {
  httpGetAsync(location.origin + "/request_rating_history?user=" + username, d3DrawRating)
}

function d3DrawRating(r) {
  var ratingHistories = JSON.parse(r);
  // Setup
  var graph_div = document.getElementById("rating_history");
  var graph_pos = graph_div.getBoundingClientRect();
  var margin = {left: 60, top: 20, right: 10, bottom: 30};  // left, top, right, bottom
  var w = graph_pos.width - margin.left - margin.right;
  var h = graph_pos.height - margin.top - margin.bottom;
  // x-axis and y-axis
  var x = d3.scale.linear().domain([0, ratingHistories.length]).range([0, w]);
  var minRating = d3.min(ratingHistories, function(d){ return d.Rating; });
  var maxRating = d3.max(ratingHistories, function(d){ return d.Rating; });
  var y = d3.scale.linear().domain([minRating - 20, maxRating + 20]).range([h, 0]);
  // Draw axis and line
  var line = d3.svg.line()
                   .x(function(d, i) { return x(i); })
	           .y(function(d) { return y(d.Rating); });
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
       .attr("d", line(ratingHistories));
}
