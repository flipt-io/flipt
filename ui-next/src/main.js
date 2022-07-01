import Vue from "vue";
import App from "./App.vue";
import Buefy from "buefy";
import VueMoment from "vue-moment";
import router from "./router";
import store from "./store";

//import "./assets/sass/main.scss";

Vue.config.productionTip = false;
Vue.use(VueMoment);
Vue.use(Buefy);

new Vue({
  router,
  store,
  render: (h) => h(App),
  template: "<App/>",
}).$mount("#app");
