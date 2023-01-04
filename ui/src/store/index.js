import Vuex from "vuex";
import Vue from "vue";
import axios from "axios";

//load Vuex
Vue.use(Vuex);

//to handle state
const state = {
  info: {},
  csrfToken: "",
};

//to handle state
const getters = {
  info: (state) => state.info,
  csrfToken: (state) => state.csrfToken,
};

//to handle actions
const actions = {
  getInfo({ commit }) {
    axios
      .get("/meta/info")
      .then((response) => {
        commit("SET_INFO", response.data);

        const token = response.headers['x-csrf-token'];
        if (token != null) {
          commit("SET_CSRF_TOKEN", token);
        }
      })
      .catch((e) => {
        console.log(e);
      });
  },
};

//to handle mutations
const mutations = {
  SET_INFO(state, info) {
    state.info = info;
    if (info.version) {
      state.info.version = "v" + info.version;
    }
    if (info.latestVersion) {
      state.info.latestVersion = "v" + info.latestVersion;
    }
  },
  SET_CSRF_TOKEN(state, token) {
    state.csrfToken = token;
  },
};

//export store module
export default new Vuex.Store({
  state,
  getters,
  actions,
  mutations,
});
