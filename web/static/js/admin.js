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
  getBadges();
}

function getBadges() {
  httpGetAsync(location.origin + "/request_all_badges", fillInBadges);
}

function fillInBadges(r) {
  var badges = JSON.parse(r);
  var badge_table = document.getElementById("badges");
  var content = "<tr>" +
                "<th>Badge</th>" +
                "<th>Icon</th>" +
                "<th>Description</th>" +
                "<th>Author</th>" +
                "</tr>";
  for (var i in badges) {
    var badge = badges[i];
    var row = "<tr>" +
              "<td>" + badge.Name + "</td>" +
              "<td><img src=\"" + badge.Path + "\" width=32 height=32></img></td>" +
              "<td>" + badge.Description + "</td>" +
              "<td>" + badge.Author + "</td>" +
              "</tr>";
    content += row;
  }
  badge_table.innerHTML = content;
}

