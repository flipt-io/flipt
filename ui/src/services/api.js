import axios from "axios";

export const Api = axios.create({
  baseURL: "//" + window.location.host + "/api/v1/"
});
