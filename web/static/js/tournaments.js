function loadTournaments() {
    httpGetAsync(location.origin + "/request_tournaments", fillTournaments);
}

function fillTournaments(responseText) {
    var tournaments = JSON.parse(responseText);
    var container = document.getElementById("container");
    for (var i in tournaments) {
        var t = tournaments[i];
        var tournamentDiv = document.createElement('h3');
        var anchor = document.createElement('a');
        anchor.href = "tournament/" + t.Name;
        anchor.textContent = t.Name;
        tournamentDiv.appendChild(anchor);
        container.appendChild(tournamentDiv);
    }
}