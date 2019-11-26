import axios from "axios";

let host =
  process.env.NODE_ENV === "production"
    ? window.location.host
    : "localhost:8080";

export const Api = axios.create({
  baseURL: "//" + host + "/api/v1/"
});
