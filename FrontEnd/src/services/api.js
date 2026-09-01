import axios from 'axios';

const api = axios.create({
  baseURL: '',
  headers: {
    'Content-Type': 'application/json',
  },
});

const TOKEN_KEY = 'adminToken';
const REFRESH_KEY = 'adminRefreshToken';
const EXPIRY_KEY = 'adminTokenExpiry';
const USERNAME_KEY = 'adminUsername';

// Endpoints that must never trigger the refresh-retry flow
const NO_RETRY_PATHS = ['/api/auth/login', '/api/auth/verify', '/api/auth/refresh'];

function getToken() {
  return localStorage.getItem(TOKEN_KEY);
}

function getRefreshToken() {
  return localStorage.getItem(REFRESH_KEY);
}

function getTokenExpiry() {
  return Number(localStorage.getItem(EXPIRY_KEY) || 0);
}

function getUsername() {
  return localStorage.getItem(USERNAME_KEY) || '';
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

function clearSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(REFRESH_KEY);
  localStorage.removeItem(EXPIRY_KEY);
  localStorage.removeItem(USERNAME_KEY);
}

function isNoRetryUrl(url) {
  return NO_RETRY_PATHS.some((path) => String(url || '').includes(path));
}

// Single in-flight refresh so concurrent 401s share one rotation
let refreshInFlight = null;

function refreshSession() {
  if (!refreshInFlight) {
    const refreshToken = getRefreshToken();
    if (!refreshToken) {
      return Promise.reject(new Error('No refresh token'));
    }
    refreshInFlight = axios
      .post('/api/auth/refresh', { refreshToken })
      .then((response) => {
        setSession(response.data);
        return response.data;
      })
      .finally(() => {
        refreshInFlight = null;
      });
  }
  return refreshInFlight;
}

// On 401: refresh the session once, swap the fresh token into the body, retry
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const config = error.config || {};
    const isUnauthorized = error.response && error.response.status === 401;

    if (
      isUnauthorized &&
      !config._retried &&
      !isNoRetryUrl(config.url) &&
      getRefreshToken()
    ) {
      config._retried = true;
      try {
        await refreshSession();
        let data = {};
        try {
          data = JSON.parse(config.data || '{}');
        } catch (e) {
          data = {};
        }
        data.token = getToken();
        config.data = JSON.stringify(data);
        return api.request(config);
      } catch (refreshError) {
        clearSession();
      }
    }

    throw error;
  }
);

// Proactively refresh shortly before the access token expires
let autoRefreshTimer = null;

function ensureAutoRefresh() {
  if (autoRefreshTimer) return;
  autoRefreshTimer = setInterval(async () => {
    if (!getRefreshToken()) return;
    const expiry = getTokenExpiry();
    if (!expiry || Date.now() > expiry - 10 * 60 * 1000) {
      try {
        await refreshSession();
      } catch (e) {
        clearSession();
      }
    }
  }, 60 * 1000);
}

ensureAutoRefresh();

// Refresh immediately when the tab becomes visible again after being away
document.addEventListener('visibilitychange', () => {
  if (document.visibilityState === 'visible' && getRefreshToken()) {
    const expiry = getTokenExpiry();
    if (!expiry || Date.now() > expiry - 10 * 60 * 1000) {
      refreshSession().catch(() => clearSession());
    }
  }
});

function authPayload(extra = {}) {
  const payload = { ...extra };
  const token = getToken();
  if (token) {
    payload.token = token;
  }
  return payload;
}

export default {
  async submitApplication(applicationData) {
    try {
      const data = {
        ...applicationData,
        age: Number(applicationData.age)
      };
      const response = await api.post('/api/application', data);
      return response.data;
    } catch (error) {
      console.error('Error submitting application:', error);
      throw error;
    }
  },

  async login(username, password) {
    const response = await api.post('/api/auth/login', { username, password });
    if (response.data.success && response.data.token) {
      setSession(response.data);
    }
    return response.data;
  },

  // Legacy alias kept for compatibility
  async verifyAdminPassword(password) {
    return this.login(getUsername(), password);
  },

  async logout() {
    const rememberedUsername = getUsername();
    try {
      await api.post('/api/auth/logout', {
        token: getToken(),
        refreshToken: getRefreshToken(),
      });
    } catch (error) {
      // Session is cleared locally regardless
    }
    clearSession();
    // Keep the username around to prefill the next login
    if (rememberedUsername) {
      localStorage.setItem(USERNAME_KEY, rememberedUsername);
    }
  },

  async validateToken() {
    try {
      const token = getToken();
      if (!token) return { valid: false };
      const response = await api.post('/api/auth/validate', { token });
      return response.data;
    } catch (error) {
      console.error('Error validating token:', error);
      return { valid: false };
    }
  },

  async getApplications(adminPassword, query = '', fields = [], page = 1, pageSize = 50) {
    try {
      const payload = authPayload({ query, fields, page, pageSize });
      if (adminPassword && !payload.token) {
        payload.password = adminPassword;
      }
      const response = await api.post('/api/application/list', payload);
      return response.data;
    } catch (error) {
      console.error('Error fetching applications:', error);
      throw error;
    }
  },

  async getApplicationStats(adminPassword) {
    try {
      const payload = authPayload();
      if (adminPassword && !payload.token) {
        payload.password = adminPassword;
      }
      const response = await api.post('/api/application/stats', payload);
      return response.data;
    } catch (error) {
      console.error('Error fetching application statistics:', error);
      throw error;
    }
  },

  async exportApplications(adminPassword) {
    try {
      const payload = authPayload();
      if (adminPassword && !payload.token) {
        payload.password = adminPassword;
      }
      const response = await api.post('/api/application/export',
        payload,
        { responseType: 'blob' }
      );

      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'applications.csv');
      document.body.appendChild(link);
      link.click();

      return true;
    } catch (error) {
      console.error('Error exporting applications:', error);
      throw error;
    }
  },

  async deleteApplication(applicationId, adminPassword) {
    try {
      const payload = authPayload({ id: applicationId });
      if (adminPassword && !payload.token) {
        payload.password = adminPassword;
      }
      const response = await api.post('/api/application/delete', payload);
      return { success: true, data: response.data };
    } catch (error) {
      console.error('Error deleting application:', error);
      return { success: false, error };
    }
  },

  async reviewApplication(applicationId, adminPassword, notes = '', acceptanceStatus = 'pending') {
    try {
      const payload = authPayload({
        id: applicationId,
        notes: notes,
        acceptance_status: acceptanceStatus
      });
      if (adminPassword && !payload.token) {
        payload.password = adminPassword;
      }
      const response = await api.post('/api/application/review', payload);
      return { success: true, data: response.data };
    } catch (error) {
      console.error('Error reviewing application:', error);
      return { success: false, error };
    }
  },

  async unreviewApplication(applicationId, adminPassword) {
    try {
      const payload = authPayload({ id: applicationId });
      if (adminPassword && !payload.token) {
        payload.password = adminPassword;
      }
      const response = await api.post('/api/application/unreview', payload);
      return { success: true, data: response.data };
    } catch (error) {
      console.error('Error unreview application:', error);
      return { success: false, error };
    }
  },

  async updateApplication(applicationId, applicationData) {
    try {
      const payload = authPayload({ id: applicationId, ...applicationData });
      const response = await api.post('/api/application/update', payload);
      return { success: true, data: response.data };
    } catch (error) {
      console.error('Error updating application:', error);
      return { success: false, error };
    }
  },

  async getApplicationHistory(applicationId) {
    const response = await api.post('/api/application/history', authPayload({ id: applicationId }));
    return response.data;
  },

  async getRecentChanges(limit = 100) {
    const response = await api.post('/api/application/changes', authPayload({ limit }));
    return response.data;
  },

  async getAdmins() {
    const response = await api.post('/api/auth/admins/list', authPayload());
    return response.data;
  },

  async createAdmin(username, password) {
    const response = await api.post('/api/auth/admins/create', authPayload({ username, newPassword: password }));
    return response.data;
  },

  async deleteAdmin(id) {
    const response = await api.post('/api/auth/admins/delete', authPayload({ id }));
    return response.data;
  },

  async changeAdminPassword(id, newPassword) {
    const response = await api.post('/api/auth/admins/password', authPayload({ id, newPassword }));
    return response.data;
  },

  getUsername,
  ensureFreshSession: async function () {
    if (!getRefreshToken()) return false;
    const expiry = getTokenExpiry();
    if (!expiry || Date.now() > expiry - 10 * 60 * 1000) {
      try {
        await refreshSession();
      } catch (e) {
        clearSession();
        return false;
      }
    }
    return true;
  }
};
