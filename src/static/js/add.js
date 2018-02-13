Vue.component('v-select', VueSelect.VueSelect);

var v_w = null;
var v_l = null;
var vue_created = false;

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
    message += lm.Match.Winner + " (" + lm.Match.WinnerRatingBefore + " &#x27a8; " + lm.Match.WinnerRatingAfter + ") ";
    message += "<br>";
    if (lm.Expected) {
        message += " &#9876; ";
    } else {
        message += " &#x1F525; ";
    }
    message += "<br>";
    message += lm.Match.Loser + " (" + lm.Match.LoserRatingBefore + " &#x27a8; " + lm.Match.LoserRatingAfter + ") ";
    console.log("message = " + message);

    message_div.innerHTML = message;
    message_div.style.display = "block";

    // Best effort to set the default value.
    // Otherwise, need to wait for vue elements created.
    setTimeout(
      function() {
        if (v_w != null && v_l != null && vue_created == true) {
          v_w.selected = lm.Match.Winner;
          v_l.selected = lm.Match.Loser;
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
