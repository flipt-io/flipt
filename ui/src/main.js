import Vue from "vue";
import App from "./App";
import Buefy from "buefy";
import VueMoment from "vue-moment";
import router from "./router";
import store from "./store";

import "./assets/sass/main.scss";

Vue.config.productionTip = false;
Vue.use(VueMoment);
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

new Vue({
  router,
  store,
  render: (h) => h(App),
  template: "<App/>",
}).$mount("#app");
