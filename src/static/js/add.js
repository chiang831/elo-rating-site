Vue.component('v-select', VueSelect.VueSelect);

new Vue({
  el: '#winner_select',
  data: function() {
      return {
        users: [
            {"Name": "moja"},
            {"Name": "jimmy"},
        ],
        selected: {"Name": " "}
     }
  },

  methods: {
  }
});

new Vue({
  el: '#loser_select',
  data: function() {
      return {
        users: [
            {"Name": "moja"},
            {"Name": "jimmy"},
        ],
        selected: {"Name": " "}
     }
  },

  methods: {
  }
});
