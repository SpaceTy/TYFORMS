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
  }
}; 