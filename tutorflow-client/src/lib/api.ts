import axios from "axios";
import Cookies from "js-cookie";

export const api = axios.create({
  baseURL: "/api", // Now hitting the proxy.ts
  headers: {
    "Content-Type": "application/json",
  },
});

// Response interceptor to handle token refresh
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = Cookies.get("refreshToken");
        if (refreshToken) {
          const response = await axios.post("/api/auth/refresh", {
            refresh_token: refreshToken,
          });

          const { access_token, refresh_token } = response.data.data.tokens;
          Cookies.set("accessToken", access_token, { expires: 1 / 24 }); // 1 hour
          Cookies.set("refreshToken", refresh_token, { expires: 7 }); // 7 days

          // The proxy will automatically attach the new accessToken to this retry
          return api(originalRequest);
        }
      } catch {
        // Clear tokens and redirect to login
        Cookies.remove("accessToken");
        Cookies.remove("refreshToken");
        if (typeof window !== "undefined") {
          window.location.href = "/login";
        }
      }
    }

    return Promise.reject(error);
  }
);

export default api;
