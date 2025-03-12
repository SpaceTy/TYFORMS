<template>
  <div class="h-full w-full absolute inset-0 flex items-center justify-center">
    <!-- Login Form -->
    <div v-if="!isAuthenticated" class="login-container mc-container animate-fade-in max-w-md">
      <h2 class="mc-title text-center">Admin Login</h2>
      
      <form @submit.prevent="authenticate" class="space-y-6">
        <div class="form-group">
          <label for="password" class="mc-label">Password</label>
          <input 
            id="password" 
            v-model="password" 
            type="password" 
            class="mc-input"
            placeholder="Enter admin password"
            required
            ref="passwordInput"
          />
        </div>
        
        <div v-if="authError" class="bg-red-500/50 text-white p-3 rounded-md">
          {{ authError }}
        </div>
        
        <div class="text-center">
          <button type="submit" class="mc-button">
            Login
          </button>
        </div>
      </form>
    </div>
    
    <!-- Admin Dashboard -->
    <div v-else class="absolute inset-0 p-4">
      <div ref="adminContainerRef" class="admin-panel mc-panel w-full h-full flex flex-col">
        <div class="flex justify-between items-center px-6 py-4 bg-minecraft-deepslate/80 rounded-t-md">
          <h2 class="mc-title mb-0">Applications Dashboard</h2>
          
          <div class="flex gap-2">
            <button @click="refreshData" class="mc-button text-sm">
              <span v-if="isLoading">Loading...</span>
              <span v-else>Refresh</span>
            </button>
            
            <button @click="exportToCsv" class="mc-button text-sm">
              Export CSV
            </button>
            
            <button @click="logout" class="mc-button text-sm bg-red-500 hover:bg-red-600 border-red-700">
              Logout
            </button>
          </div>
        </div>
        
        <div v-if="errorMessage" class="bg-red-500/50 text-white p-3 mx-6 my-4 rounded-md">
          {{ errorMessage }}
        </div>
        
        <div class="flex-grow overflow-hidden flex flex-col p-6 bg-black/40">
          <div v-if="applications.length === 0 && !isLoading" class="text-center py-10 text-white flex-grow flex items-center justify-center">
            <div>
              <p class="text-xl font-minecraft">No applications yet</p>
              <p class="mt-2">Applications will appear here once submitted</p>
            </div>
          </div>
          
          <div v-else class="overflow-auto flex-grow">
            <table class="w-full table-fixed bg-black/50 border border-minecraft-stone rounded-md">
              <thead class="sticky top-0 z-10">
                <tr class="bg-minecraft-deepslate">
                  <th class="w-12 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap">#</th>
                  <th class="w-32 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap">DC</th>
                  <th class="w-32 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap">MC</th>
                  <th class="w-14 px-3 py-3 text-center text-sm font-minecraft text-white whitespace-nowrap">Age</th>
                  <th class="w-1/4 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap">FAM</th>
                  <th class="w-1/4 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap">SU</th>
                  <th class="w-20 px-3 py-3 text-center text-sm font-minecraft text-white whitespace-nowrap">S?</th>
                  <th class="w-36 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap">Date</th>
                  <th class="w-20 px-3 py-3 text-center text-sm font-minecraft text-white whitespace-nowrap">Actions</th>
                </tr>
              </thead>
              
              <tbody>
                <tr 
                  v-for="(app, index) in applications" 
                  :key="app.id" 
                  :class="{'bg-black/30': index % 2 === 0}"
                  class="border-t border-minecraft-stone hover:bg-minecraft-water/20 transition-colors duration-150"
                >
                  <td class="px-3 py-2 text-sm text-white">
                    {{ app.id }}
                  </td>
                  
                  <td class="px-3 py-2 text-sm text-white">
                    <div class="tooltip-container">
                      <div class="truncate">{{ app.discordUsername }}</div>
                      <div class="tooltip">{{ app.discordUsername }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm text-white">
                    <div class="tooltip-container">
                      <div class="truncate">{{ app.minecraftUsername }}</div>
                      <div class="tooltip">{{ app.minecraftUsername }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm font-bold text-center" :class="getAgeColor(app.age)">
                    {{ app.age }}
                  </td>
                  
                  <td class="px-3 py-2 text-sm text-white">
                    <div class="tooltip-container">
                      <div class="truncate">{{ app.favoriteAboutMinecraft }}</div>
                      <div class="tooltip">{{ app.favoriteAboutMinecraft }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm text-white">
                    <div class="tooltip-container">
                      <div class="truncate">{{ app.understandingOfSMP }}</div>
                      <div class="tooltip">{{ app.understandingOfSMP }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm text-white text-center">
                    <span 
                      :class="app.joinedDiscord ? 'bg-minecraft-green' : 'bg-red-500'" 
                      class="px-2 py-1 rounded-full text-xs"
                    >
                      {{ app.joinedDiscord ? 'Yes' : 'No' }}
                    </span>
                  </td>
                  
                  <td class="px-3 py-2 text-sm text-white">
                    <div class="tooltip-container">
                      <div class="truncate">{{ formatDate(app.submissionDate) }}</div>
                      <div class="tooltip">{{ formatDate(app.submissionDate, true) }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm text-white text-center">
                    <button 
                      @click="confirmDelete(app.id)"
                      class="mc-button text-xs bg-red-500 hover:bg-red-600 border-red-700 px-2 py-1"
                      :disabled="isDeleting === app.id"
                    >
                      <span v-if="isDeleting === app.id">...</span>
                      <span v-else>Delete</span>
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue';
import { gsap } from 'gsap';
import api from '../services/api';

// UI state
const isLoading = ref(false);
const errorMessage = ref('');
const applications = ref([]);

// Authentication state
const isAuthenticated = ref(false);
const password = ref('');
const authError = ref('');
const passwordInput = ref(null);
const authenticatedPassword = ref(''); // Store authenticated password for subsequent requests
const adminContainerRef = ref(null);
const isDeleting = ref(null); // Track which row is being deleted

// Authenticate user
async function authenticate() {
  try {
    const response = await api.verifyAdminPassword(password.value);
    
    if (response.success) {
      // Store the authenticated password for future API requests
      authenticatedPassword.value = password.value;
      
      // Animate login transition
      gsap.to('.login-container', {
        opacity: 0,
        scale: 0.95,
        duration: 0.3,
        onComplete: () => {
          isAuthenticated.value = true;
          password.value = '';
          authError.value = '';
          
          // After view changes, animate the admin panel in
          nextTick(() => {
            // Load applications data after successful login
            refreshData();
            
            // Make sure the admin panel is visible with proper initial state
            if (adminContainerRef.value) {
              // Ensure it's visible before animation
              gsap.set(adminContainerRef.value, { 
                opacity: 0,
                scale: 0.95
              });
              
              // Animate it in
              gsap.to(adminContainerRef.value, {
                opacity: 1,
                scale: 1,
                duration: 0.3,
                delay: 0.1 // Small delay to ensure DOM is ready
              });
            }
          });
        }
      });
    } else {
      handleAuthError();
    }
  } catch (error) {
    console.error('Authentication error:', error);
    handleAuthError();
  }
}

function handleAuthError() {
  authError.value = 'Incorrect password. Please try again.';
  password.value = '';
  
  // Focus back on password input
  nextTick(() => {
    if (passwordInput.value) {
      passwordInput.value.focus();
    }
  });
  
  // Shake animation for error
  gsap.to('.login-container', {
    x: -10,
    duration: 0.1,
    repeat: 5,
    yoyo: true,
    ease: 'power1.inOut'
  });
}

// Logout function
function logout() {
  // Animate out
  gsap.to(adminContainerRef.value, {
    opacity: 0,
    scale: 0.95,
    duration: 0.3,
    onComplete: () => {
      isAuthenticated.value = false;
      authenticatedPassword.value = ''; // Clear stored password
      
      // Reset the UI
      nextTick(() => {
        // Animate login form back in
        gsap.fromTo('.login-container', 
          { opacity: 0, scale: 0.95 },
          { opacity: 1, scale: 1, duration: 0.3 }
        );
        
        if (passwordInput.value) {
          passwordInput.value.focus();
        }
      });
    }
  });
}

// Get age display color based on value
function getAgeColor(age) {
  if (age < 16) {
    return 'text-red-400';
  } else if (age >= 16 && age <= 20) {
    return 'text-green-400';
  } else {
    return 'text-purple-400';
  }
}

// Load data on mount if authenticated
onMounted(() => {
  // Focus on password input
  nextTick(() => {
    if (passwordInput.value) {
      passwordInput.value.focus();
    }
  });
});

// Refresh application data
async function refreshData() {
  if (!isAuthenticated.value) return;
  
  try {
    isLoading.value = true;
    errorMessage.value = '';
    
    const data = await api.getApplications(authenticatedPassword.value);
    applications.value = data;
  } catch (error) {
    console.error('Error fetching applications:', error);
    handleApiError(error);
  } finally {
    isLoading.value = false;
  }
}

// Delete application
async function deleteApplication(id) {
  if (!isAuthenticated.value) return;
  
  try {
    isDeleting.value = id;
    errorMessage.value = '';
    
    await api.deleteApplication(id, authenticatedPassword.value);
    
    // Remove the application from the list
    applications.value = applications.value.filter(app => app.id !== id);
    
    // Show success animation on the row that was deleted
    // This is handled by removing it from the array
  } catch (error) {
    console.error('Error deleting application:', error);
    handleApiError(error);
  } finally {
    isDeleting.value = null;
  }
}

// Confirm delete with user
function confirmDelete(id) {
  if (confirm('Are you sure you want to delete this application? This action cannot be undone.')) {
    deleteApplication(id);
  }
}

// Export to CSV
async function exportToCsv() {
  try {
    await api.exportApplications(authenticatedPassword.value);
  } catch (error) {
    handleApiError(error);
  }
}

// Handle API errors, including authentication failures
function handleApiError(error) {
  if (error.response && error.response.status === 401) {
    errorMessage.value = 'Authentication error. Please log in again.';
    // Force logout on authentication error
    setTimeout(() => {
      logout();
    }, 2000);
  } else {
    errorMessage.value = 'Failed to complete the operation. Please try again.';
  }
}

// Format date
function formatDate(dateString, detailed = false) {
  if (!dateString) return '';
  
  const date = new Date(dateString);
  
  if (detailed) {
    return new Intl.DateTimeFormat('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit',
      timeZoneName: 'short'
    }).format(date);
  }
  
  return new Intl.DateTimeFormat('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date);
}
</script>

<style scoped>
/* Make the component take full height and width */
:deep(body), :deep(html), :deep(#app), :deep(.min-h-screen), :deep(main) {
  height: 100% !important;
  width: 100% !important;
  overflow: hidden !important;
  position: relative !important;
}

/* Fix for mc-title in header */
.mc-title {
  margin-bottom: 0;
}

/* Admin panel styling */
.mc-panel {
  background-color: rgba(74, 74, 74, 0.2);
  border: 2px solid #7f8c8d;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(4px);
}

/* Fixed table header styles */
thead {
  background-color: #2c3e50;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.2);
}

/* Table styles */
table {
  border-collapse: collapse;
  border-spacing: 0;
}

th, td {
  vertical-align: middle;
  padding: 0.5rem 0.75rem;
}

tr {
  margin: 0;
}

/* Tooltip styles */
.tooltip-container {
  position: relative;
  cursor: pointer;
  width: 100%;
}

.tooltip-container .truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tooltip-container:hover .tooltip {
  display: block;
}

.tooltip {
  display: none;
  position: absolute;
  left: 0;
  background-color: #2c3e50;
  color: white;
  padding: 0.75rem;
  border-radius: 6px;
  font-size: 0.875rem;
  line-height: 1.5;
  white-space: normal;
  max-width: 300px;
  word-break: break-word;
  z-index: 50;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.5);
  border: 1px solid #7f8c8d;
  transition: opacity 0.2s ease;
  opacity: 0;
}

/* Make sure tooltips don't overflow horizontally and show properly */
.tooltip-container:hover .tooltip {
  display: block;
  opacity: 1;
  max-width: min(300px, 90vw);
}

/* Tooltip arrow indicators */
.tooltip::after {
  content: '';
  position: absolute;
  width: 0;
  height: 0;
  border-left: 8px solid transparent;
  border-right: 8px solid transparent;
}

/* Tooltip positions based on row position */
tr:nth-child(-n+3) .tooltip {
  top: 130%; /* For first 3 rows, show below */
}

tr:nth-child(n+4) .tooltip {
  bottom: 130%; /* For rest of rows, show above */
}

/* Horizontal tooltip positioning */
td:nth-child(-n+3) .tooltip {
  left: 0; /* For first 3 columns, align left */
}

td:nth-child(n+4) .tooltip {
  right: 0; /* For right columns, align right */
  left: auto;
}

/* Arrow for tooltips that show below */
tr:nth-child(-n+3) .tooltip::after {
  top: -8px;
  left: 15px;
  border-bottom: 8px solid #2c3e50;
}

/* Arrow for tooltips that show above */
tr:nth-child(n+4) .tooltip::after {
  bottom: -8px;
  left: 15px;
  border-top: 8px solid #2c3e50;
}

/* Adjust arrow position for right-aligned tooltips */
td:nth-child(n+4) .tooltip::after {
  left: auto;
  right: 15px;
}

/* Make sure tooltips don't overflow horizontally */
.tooltip-container:hover .tooltip {
  max-width: min(300px, 90vw);
  display: block;
  opacity: 1;
}

/* Make sure the admin panel is visible */
.admin-panel {
  opacity: 1;
  transform: scale(1);
}

/* Make the table row hover effect more noticeable */
tbody tr:hover {
  background-color: rgba(52, 152, 219, 0.2) !important;
}
</style> 