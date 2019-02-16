import Vue from "vue";
import App from "./App";
import Buefy from "buefy";

import router from "./router";

require("./assets/sass/main.scss");

Vue.config.productionTip = false;
Vue.use(require("vue-moment"));
Vue.use(Buefy);

/* eslint-disable no-new */
new Vue({
  el: "#app",
  router,
  components: { App },
  template: "<App/>"
});
