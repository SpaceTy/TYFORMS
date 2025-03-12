import axios from 'axios';

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
});

export default {
  async submitApplication(applicationData) {
    try {
      const response = await api.post('/application', applicationData);
      return response.data;
    } catch (error) {
      console.error('Error submitting application:', error);
      throw error;
    }
  },
  
  async verifyAdminPassword(password) {
    try {
      const response = await api.post('/auth/verify', { password });
      return response.data;
    } catch (error) {
      console.error('Error verifying password:', error);
      throw error;
    }
  },
  
  async getApplications(adminPassword) {
    try {
      const response = await api.post('/application/list', { password: adminPassword });
      return response.data;
    } catch (error) {
      console.error('Error fetching applications:', error);
      throw error;
    }
  },
  
  async exportApplications(adminPassword) {
    try {
      const response = await api.post('/application/export', 
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
      const response = await api.post('/application/delete', { 
        id: applicationId,
        password: adminPassword 
      });
      return response.data;
    } catch (error) {
      console.error('Error deleting application:', error);
      throw error;
    }
  },
  
  async reviewApplication(applicationId, adminPassword, notes = '') {
    try {
      const response = await api.post('/application/review', {
        id: applicationId,
        password: adminPassword,
        notes: notes
      });
      return response.data;
    } catch (error) {
      console.error('Error reviewing application:', error);
      throw error;
    }
  },
  
  async unreviewApplication(applicationId, adminPassword) {
    try {
      const response = await api.post('/application/unreview', {
        id: applicationId,
        password: adminPassword
      });
      return response.data;
    } catch (error) {
      console.error('Error unreview application:', error);
      throw error;
    }
  }
}; 