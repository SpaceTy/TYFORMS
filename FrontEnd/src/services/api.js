import axios from 'axios';

const api = axios.create({
  baseURL: '',
  headers: {
    'Content-Type': 'application/json',
  },
});

function getToken() {
  return localStorage.getItem('adminToken');
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

  async verifyAdminPassword(password) {
    try {
      const response = await api.post('/api/auth/verify', { password });
      return response.data;
    } catch (error) {
      console.error('Error verifying password:', error);
      throw error;
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
      const payload = { query, fields, page, pageSize };
      const token = getToken();
      if (adminPassword) {
        payload.password = adminPassword;
      } else if (token) {
        payload.token = token;
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
      const payload = {};
      const token = getToken();
      if (adminPassword) {
        payload.password = adminPassword;
      } else if (token) {
        payload.token = token;
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
      const payload = {};
      const token = getToken();
      if (adminPassword) {
        payload.password = adminPassword;
      } else if (token) {
        payload.token = token;
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
      const payload = { id: applicationId };
      const token = getToken();
      if (adminPassword) {
        payload.password = adminPassword;
      } else if (token) {
        payload.token = token;
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
      const payload = {
        id: applicationId,
        notes: notes,
        acceptance_status: acceptanceStatus
      };
      const token = getToken();
      if (adminPassword) {
        payload.password = adminPassword;
      } else if (token) {
        payload.token = token;
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
      const payload = { id: applicationId };
      const token = getToken();
      if (adminPassword) {
        payload.password = adminPassword;
      } else if (token) {
        payload.token = token;
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
      const response = await api.post('/api/application/update', {
        id: applicationId,
        ...applicationData
      });
      return { success: true, data: response.data };
    } catch (error) {
      console.error('Error updating application:', error);
      return { success: false, error };
    }
  }
};
