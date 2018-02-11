Vue.component('v-select', VueSelect.VueSelect);

function createVueElements(r) {
  users = JSON.parse(r);
  console.log("users = " + users);
  var v_w = new Vue({
    el: '#winner_select',
    data: function() {
        return {
          users: users,
          selected: {"Name": " "}
       }
    },

    methods: {
    }
  });

  var v_l = new Vue({
    el: '#loser_select',
    data: function() {
        return {
          users: users,
          selected: {"Name": " "}
       }
    },

    methods: {
    }
  });
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

function getUsers () {
  console.log("get users");
  // Get available user data from JSON API.
  httpGetAsync(location.origin + "/users", createVueElements);
}
