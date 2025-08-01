const API_URL = process.env.API_URL || "http://localhost:8080/api";
const BASE_URL = API_URL.replace("/api", "/"); // Remove '/api' to get the base URL

export { API_URL, BASE_URL };
