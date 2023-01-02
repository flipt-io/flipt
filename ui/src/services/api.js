import axios from "axios";
import store from "../store";

const CSRFTokenHeader = 'X-CSRF-Token';

export const Api = axios.create({
  baseURL: "/api/v1/",
});

Api.interceptors.request.use((config) => {
  const token = store.getters.csrfToken;
  if (token != null) {
    config.headers[CSRFTokenHeader] = token;
  }

  return config;
}, (error) => {
  return Promise.reject(error);
});
