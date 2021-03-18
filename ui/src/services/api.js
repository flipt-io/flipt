import axios from "axios";

let host =
  process.env.NODE_ENV === "production" ? window.location.host : "127.0.0.1:8080";

export const Api = axios.create({
  baseURL: "//" + host + "/api/v1/"
});
