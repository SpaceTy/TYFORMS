const TOKEN_KEY = "adminToken";
const REFRESH_KEY = "adminRefreshToken";
const EXPIRY_KEY = "adminTokenExpiry";
const USERNAME_KEY = "adminUsername";

const NO_RETRY_PATHS = ["/api/auth/login", "/api/auth/verify", "/api/auth/refresh"];

export function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

export function getRefreshToken() {
  return localStorage.getItem(REFRESH_KEY);
}

export function getTokenExpiry() {
  return Number(localStorage.getItem(EXPIRY_KEY) || 0);
}

export function getUsername() {
  return localStorage.getItem(USERNAME_KEY) || "";
}

function setSession(data) {
  if (data.token) localStorage.setItem(TOKEN_KEY, data.token);
  if (data.refreshToken) localStorage.setItem(REFRESH_KEY, data.refreshToken);
  if (data.expiresAt) {
    localStorage.setItem(EXPIRY_KEY, String(new Date(data.expiresAt).getTime()));
  }
  if (data.admin && data.admin.username) {
    localStorage.setItem(USERNAME_KEY, data.admin.username);
  }
}

export function clearSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_KEY);
  localStorage.removeItem(EXPIRY_KEY);
  localStorage.removeItem(USERNAME_KEY);
}

function isNoRetryPath(path) {
  return NO_RETRY_PATHS.some((item) => String(path || "").includes(item));
}

export function authPayload(extra = {}) {
  const payload = { ...extra };
  const token = getToken();
  if (token) payload.token = token;
  return payload;
}

async function rawPost(path, payload, { blob = false } = {}) {
  const response = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload || {}),
  });

  if (blob) {
    if (!response.ok) {
      const error = new Error(`Request failed (${response.status})`);
      error.status = response.status;
      throw error;
    }
    return response.blob();
  }

  const text = await response.text();
  let data = null;
  if (text) {
    try {
      data = JSON.parse(text);
    } catch (error) {
      data = text;
    }
  }

  if (!response.ok) {
    const error = new Error(
      (data && data.error) || (typeof data === "string" && data.trim()) || `Request failed (${response.status})`
    );
    error.status = response.status;
    error.data = data;
    throw error;
  }

  return data;
}

async function request(path, payload = {}, options = {}) {
  try {
    return await rawPost(path, payload, options);
  } catch (error) {
    const canRetry =
      error.status === 401 &&
      !options._retried &&
      !isNoRetryPath(path) &&
      getRefreshToken();

    if (!canRetry) throw error;

    try {
      await refreshSession();
      return await rawPost(path, { ...payload, token: getToken() }, { ...options, _retried: true });
    } catch (refreshError) {
      if (refreshError.status === 401 || refreshError.status === 400 || refreshError.status === 403) {
        clearSession();
      }
      throw refreshError;
    }
  }
}

let refreshInFlight = null;

export function refreshSession() {
  if (!refreshInFlight) {
    const refreshToken = getRefreshToken();
    if (!refreshToken) {
      return Promise.reject(new Error("No refresh token"));
    }
    refreshInFlight = rawPost("/api/auth/refresh", { refreshToken })
      .then((data) => {
        setSession(data);
        return data;
      })
      .finally(() => {
        refreshInFlight = null;
      });
  }
  return refreshInFlight;
}

let autoRefreshTimer = null;

function ensureAutoRefresh() {
  if (autoRefreshTimer) return;
  autoRefreshTimer = setInterval(async () => {
    if (!getRefreshToken()) return;
    const expiry = getTokenExpiry();
    if (!expiry || Date.now() > expiry - 10 * 60 * 1000) {
      try {
        await refreshSession();
      } catch (error) {
        clearSession();
      }
    }
  }, 60 * 1000);
}

document.addEventListener("visibilitychange", () => {
  if (document.visibilityState === "visible" && getRefreshToken()) {
    const expiry = getTokenExpiry();
    if (!expiry || Date.now() > expiry - 10 * 60 * 1000) {
      refreshSession().catch(() => clearSession());
    }
  }
});

ensureAutoRefresh();

async function submitApplication(applicationData) {
  const payload = { ...applicationData, age: Number(applicationData.age) };
  return request("/api/application", payload);
}

async function login(username, password) {
  const data = await request("/api/auth/login", { username, password });
  if (data && data.success && data.token) setSession(data);
  return data;
}

async function logout() {
  const rememberedUsername = getUsername();
  try {
    await request("/api/auth/logout", {
      token: getToken(),
      refreshToken: getRefreshToken(),
    });
  } catch (error) {
    // ignored on purpose
  }
  clearSession();
  if (rememberedUsername) localStorage.setItem(USERNAME_KEY, rememberedUsername);
}

async function validateToken() {
  try {
    const token = getToken();
    if (!token) return { valid: false };
    return await request("/api/auth/validate", { token });
  } catch (error) {
    return { valid: false };
  }
}

async function getApplications(adminPassword, query = "", fields = [], page = 1, pageSize = 50) {
  const payload = authPayload({ query, fields, page, pageSize });
  if (adminPassword && !payload.token) payload.password = adminPassword;
  return request("/api/application/list", payload);
}

async function getApplicationStats(adminPassword) {
  const payload = authPayload();
  if (adminPassword && !payload.token) payload.password = adminPassword;
  const data = await request("/api/application/stats", payload);
  return data;
}

async function exportApplications(adminPassword) {
  const payload = authPayload();
  if (adminPassword && !payload.token) payload.password = adminPassword;
  const blob = await request("/api/application/export", payload, { blob: true });
  const url = window.URL.createObjectURL(new Blob([blob]));
  const link = document.createElement("a");
  link.href = url;
  link.setAttribute("download", "applications.csv");
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(url);
  return true;
}

async function deleteApplication(applicationId, adminPassword) {
  const payload = authPayload({ id: applicationId });
  if (adminPassword && !payload.token) payload.password = adminPassword;
  const data = await request("/api/application/delete", payload);
  return { success: true, data };
}

async function reviewApplication(applicationId, adminPassword, notes = "", acceptanceStatus = "pending") {
  const payload = authPayload({
    id: applicationId,
    notes,
    acceptance_status: acceptanceStatus,
  });
  if (adminPassword && !payload.token) payload.password = adminPassword;
  const data = await request("/api/application/review", payload);
  return { success: true, data };
}

async function unreviewApplication(applicationId, adminPassword) {
  const payload = authPayload({ id: applicationId });
  if (adminPassword && !payload.token) payload.password = adminPassword;
  const data = await request("/api/application/unreview", payload);
  return { success: true, data };
}

async function updateApplication(applicationId, applicationData) {
  const payload = authPayload({ id: applicationId, ...applicationData });
  const data = await request("/api/application/update", payload);
  return { success: true, data };
}

async function getApplicationHistory(applicationId) {
  return request("/api/application/history", authPayload({ id: applicationId }));
}

async function getRecentChanges(limit = 100) {
  return request("/api/application/changes", authPayload({ limit }));
}

async function getAdmins() {
  return request("/api/auth/admins/list", authPayload());
}

async function createAdmin(username, password) {
  return request("/api/auth/admins/create", authPayload({ username, newPassword: password }));
}

async function deleteAdmin(id) {
  return request("/api/auth/admins/delete", authPayload({ id }));
}

async function changeAdminPassword(id, newPassword) {
  return request("/api/auth/admins/password", authPayload({ id, newPassword }));
}

async function setAdminPermissions(id, canManageAdmins) {
  return request("/api/auth/admins/permissions", authPayload({ id, canManageAdmins }));
}

async function ensureFreshSession() {
  if (!getRefreshToken()) return false;
  const expiry = getTokenExpiry();
  if (!expiry || Date.now() > expiry - 10 * 60 * 1000) {
    try {
      await refreshSession();
    } catch (error) {
      clearSession();
      return false;
    }
  }
  return true;
}

export default {
  submitApplication,
  login,
  logout,
  validateToken,
  getApplications,
  getApplicationStats,
  exportApplications,
  deleteApplication,
  reviewApplication,
  unreviewApplication,
  updateApplication,
  getApplicationHistory,
  getRecentChanges,
  getAdmins,
  createAdmin,
  deleteAdmin,
  changeAdminPassword,
  setAdminPermissions,
  getUsername,
  ensureFreshSession,
  verifyAdminPassword: (password) => login(getUsername(), password),
};
