import axios from 'axios';

const api = axios.create({
  baseURL: '',
  headers: {
    'Content-Type': 'application/json',
  },
});

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
  
  async getApplications(adminPassword) {
    try {
      const response = await api.post('/api/application/list', { password: adminPassword });
      return response.data;
    } catch (error) {
      console.error('Error fetching applications:', error);
      throw error;
    }
  },
  
  async exportApplications(adminPassword) {
    try {
      const response = await api.post('/api/application/export', 
        { password: adminPassword },
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
      const response = await api.post('/api/application/delete', { 
        id: applicationId,
        password: adminPassword 
      });
      return { success: true, data: response.data };
    } catch (error) {
      console.error('Error deleting application:', error);
      return { success: false, error };
    }
  },
  
  async reviewApplication(applicationId, adminPassword, notes = '', acceptanceStatus = 'pending') {
    try {
      const response = await api.post('/api/application/review', {
        id: applicationId,
        password: adminPassword,
        notes: notes,
        acceptance_status: acceptanceStatus
      });
      return { success: true, data: response.data };
    } catch (error) {
      console.error('Error reviewing application:', error);
      return { success: false, error };
    }
  },
  
  async unreviewApplication(applicationId, adminPassword) {
    try {
      const response = await api.post('/api/application/unreview', {
        id: applicationId,
        password: adminPassword
      });
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