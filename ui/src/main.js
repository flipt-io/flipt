import Vue from "vue";
import App from "./App";
import Buefy from "buefy";

import router from "./router";
import store from "./store";

import "./assets/sass/main.scss";

Vue.config.productionTip = false;
Vue.use(require("vue-moment"));
Vue.use(Buefy);

Vue.filter("limit", function (value) {
  if (!value) return "";
  value = value.toString();
  if (value.length > 30) {
    return value.substring(0, 29) + "...";
  } else {
    return value;
  }
});

/* eslint-disable no-new */
new Vue({
  el: "#app",
  router,
  store,
  components: { App },
  template: "<App/>",
});
