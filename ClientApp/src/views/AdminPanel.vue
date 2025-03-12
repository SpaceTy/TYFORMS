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
                  <th class="w-32 px-3 py-3 text-center text-sm font-minecraft text-white whitespace-nowrap">Actions</th>
                </tr>
              </thead>
              
              <tbody>
                <tr 
                  v-for="(app, index) in applications" 
                  :key="app.id" 
                  :class="[
                    {'bg-black/30': index % 2 === 0},
                    {'bg-minecraft-important-red/10': !app.isReviewed}
                  ]"
                  class="border-t border-minecraft-stone hover:bg-minecraft-water/20 transition-colors duration-150 application-row"
                >
                  <td class="px-3 py-2 text-sm" :class="{'text-minecraft-important-red font-bold': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="flex items-center">
                      <span v-if="!app.isReviewed" class="mr-1 review-star animate-pulse text-minecraft-important-red">⭐</span>
                      {{ app.id }}
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container">
                      <div class="truncate">{{ app.discordUsername }}</div>
                      <div class="tooltip">{{ app.discordUsername }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container">
                      <div class="truncate">{{ app.minecraftUsername }}</div>
                      <div class="tooltip">{{ app.minecraftUsername }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm font-bold text-center" :class="getAgeColor(app.age)">
                    {{ app.age }}
                  </td>
                  
                  <td class="px-3 py-2 text-sm" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container">
                      <div class="truncate">{{ app.favoriteAboutMinecraft }}</div>
                      <div class="tooltip">{{ app.favoriteAboutMinecraft }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
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
                  
                  <td class="px-3 py-2 text-sm" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container">
                      <div class="truncate">{{ formatDate(app.submissionDate) }}</div>
                      <div class="tooltip">{{ formatDate(app.submissionDate, true) }}</div>
                    </div>
                    <div v-if="app.isReviewed" class="text-xs mt-1 text-green-500">
                      Reviewed: {{ formatDate(app.reviewedAt) }}
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm text-white text-center">
                    <div class="flex flex-col gap-2">
                      <button 
                        v-if="!app.isReviewed"
                        @click="openReviewModal(app)"
                        class="mc-button text-xs bg-green-600 hover:bg-green-700 border-green-800 px-2 py-1 review-btn animate-pulse"
                        :disabled="isProcessing === app.id"
                      >
                        <span v-if="isProcessing === app.id">...</span>
                        <span v-else>Review</span>
                      </button>
                      
                      <button 
                        v-if="app.isReviewed"
                        @click="confirmUnreview(app.id)"
                        class="mc-button text-xs bg-minecraft-important-red hover:bg-red-700 border-red-800 px-2 py-1"
                        :disabled="isProcessing === app.id"
                      >
                        <span v-if="isProcessing === app.id">...</span>
                        <span v-else>Unreview</span>
                      </button>
                      
                      <button 
                        @click="confirmDelete(app.id)"
                        class="mc-button text-xs bg-red-500 hover:bg-red-600 border-red-700 px-2 py-1"
                        :disabled="isDeleting === app.id"
                      >
                        <span v-if="isDeleting === app.id">...</span>
                        <span v-else>Delete</span>
                      </button>
                      
                      <button 
                        v-if="app.isReviewed && app.reviewNotes"
                        @click="showNotes(app)"
                        class="mc-button text-xs bg-minecraft-deepslate hover:bg-gray-700 border-gray-800 px-2 py-1"
                      >
                        Notes
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Review Modal -->
    <div v-if="showReviewModal" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="mc-container max-w-lg w-full animate-scale-in">
        <div class="flex justify-between items-center mb-4">
          <h3 class="mc-title text-lg">Review Application</h3>
          <button @click="showReviewModal = false" class="text-white hover:text-red-500 transition">✕</button>
        </div>
        
        <div v-if="selectedApplication" class="mb-4">
          <p class="text-white mb-2">
            <span class="font-minecraft">Player:</span> 
            {{ selectedApplication.minecraftUsername }} ({{ selectedApplication.discordUsername }})
          </p>
          
          <div class="mb-4">
            <label class="mc-label block mb-1">Review Notes</label>
            <textarea 
              v-model="reviewNotes" 
              class="mc-input w-full" 
              rows="4"
              placeholder="Enter notes about this application (optional)"
            ></textarea>
          </div>
        </div>
        
        <div class="flex justify-end gap-2">
          <button @click="showReviewModal = false" class="mc-button bg-gray-600 hover:bg-gray-700 border-gray-800">
            Cancel
          </button>
          <button 
            @click="submitReview"
            class="mc-button bg-green-600 hover:bg-green-700 border-green-800"
            :disabled="isSubmittingReview"
          >
            <span v-if="isSubmittingReview">Submitting...</span>
            <span v-else>Submit Review</span>
          </button>
        </div>
      </div>
    </div>
    
    <!-- Notes Modal -->
    <div v-if="showNotesModal" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="mc-container max-w-lg w-full animate-scale-in">
        <div class="flex justify-between items-center mb-4">
          <h3 class="mc-title text-lg">Review Notes</h3>
          <button @click="showNotesModal = false" class="text-white hover:text-red-500 transition">✕</button>
        </div>
        
        <div v-if="selectedApplication" class="mb-4">
          <p class="text-white mb-2">
            <span class="font-minecraft">Player:</span> 
            {{ selectedApplication.minecraftUsername }} ({{ selectedApplication.discordUsername }})
          </p>
          
          <div class="bg-black/50 p-4 rounded-md border border-minecraft-stone">
            <p class="text-white whitespace-pre-line">{{ selectedApplication.reviewNotes || 'No notes provided.' }}</p>
          </div>
          
          <p class="text-xs text-minecraft-gold mt-2">
            Reviewed on: {{ formatDate(selectedApplication.reviewedAt, true) }}
          </p>
        </div>
        
        <div class="flex justify-end">
          <button @click="showNotesModal = false" class="mc-button">
            Close
          </button>
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
const isProcessing = ref(null); // Track which application is being processed (review/unreview)

// Review modal state
const showReviewModal = ref(false);
const selectedApplication = ref(null);
const reviewNotes = ref('');
const isSubmittingReview = ref(false);

// Notes modal state
const showNotesModal = ref(false);

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
            if (adminContainerRef.value) {
              gsap.from(adminContainerRef.value, {
                opacity: 0,
                scale: 0.98,
                duration: 0.4,
                ease: 'power2.out'
              });
            }
            
            // Load data
            refreshData();
          });
        }
      });
    } else {
      authError.value = 'Invalid password';
      
      // Shake animation for incorrect password
      gsap.to('.login-container', {
        x: [-10, 10, -10, 10, -5, 5, -2, 2, 0],
        duration: 0.5,
        ease: 'power2.out'
      });
    }
  } catch (error) {
    authError.value = 'Authentication error. Please try again.';
  }
}

// Get age display color
function getAgeColor(age) {
  if (age < 13) return 'text-red-500';
  if (age < 16) return 'text-yellow-400';
  if (age < 18) return 'text-green-400';
  return 'text-white';
}

// Open review modal
function openReviewModal(application) {
  selectedApplication.value = application;
  reviewNotes.value = '';
  showReviewModal.value = true;
}

// Show notes modal
function showNotes(application) {
  selectedApplication.value = application;
  showNotesModal.value = true;
}

// Submit review
async function submitReview() {
  if (!selectedApplication.value) return;
  
  isSubmittingReview.value = true;
  isProcessing.value = selectedApplication.value.id;
  
  try {
    const response = await api.reviewApplication(
      selectedApplication.value.id,
      authenticatedPassword.value,
      reviewNotes.value
    );
    
    if (response.success) {
      // Update the application in the list
      const index = applications.value.findIndex(app => app.id === selectedApplication.value.id);
      if (index !== -1) {
        applications.value[index] = response.application;
        
        // Add animation to the updated row
        nextTick(() => {
          const rows = document.querySelectorAll('.application-row');
          if (rows[index]) {
            gsap.fromTo(rows[index], 
              { backgroundColor: 'rgba(50, 205, 50, 0.3)' },  // Green for "reviewed" transition
              { backgroundColor: '', duration: 1.5, ease: 'power2.out' }
            );
          }
        });
      }
      
      showReviewModal.value = false;
    }
  } catch (error) {
    errorMessage.value = 'Failed to review application. Please try again.';
  } finally {
    isSubmittingReview.value = false;
    isProcessing.value = null;
  }
}

// Confirm unreview
function confirmUnreview(applicationId) {
  if (confirm('Are you sure you want to remove the reviewed status?')) {
    unreviewApplication(applicationId);
  }
}

// Unreview application
async function unreviewApplication(applicationId) {
  isProcessing.value = applicationId;
  
  try {
    const response = await api.unreviewApplication(
      applicationId,
      authenticatedPassword.value
    );
    
    if (response.success) {
      // Update the application in the list
      const index = applications.value.findIndex(app => app.id === applicationId);
      if (index !== -1) {
        applications.value[index] = response.application;
        
        // Add animation to the updated row - now highlighting with red since unreviewed
        nextTick(() => {
          const rows = document.querySelectorAll('.application-row');
          if (rows[index]) {
            gsap.fromTo(rows[index], 
              { backgroundColor: 'rgba(255, 82, 82, 0.3)' },  // Red for "unreviewed" transition
              { backgroundColor: '', duration: 1, ease: 'power2.out' }
            );
          }
        });
      }
    }
  } catch (error) {
    errorMessage.value = 'Failed to unreview application. Please try again.';
  } finally {
    isProcessing.value = null;
  }
}

// Confirm delete
function confirmDelete(applicationId) {
  if (confirm('Are you sure you want to delete this application? This action cannot be undone.')) {
    deleteApplication(applicationId);
  }
}

// Delete application
async function deleteApplication(applicationId) {
  isDeleting.value = applicationId;
  
  try {
    const response = await api.deleteApplication(applicationId, authenticatedPassword.value);
    
    if (response.success) {
      // Remove the deleted application from the list with animation
      const index = applications.value.findIndex(app => app.id === applicationId);
      if (index !== -1) {
        // Animate the row before removing it
        const rows = document.querySelectorAll('.application-row');
        if (rows[index]) {
          gsap.to(rows[index], {
            opacity: 0,
            height: 0,
            paddingTop: 0,
            paddingBottom: 0,
            duration: 0.3,
            onComplete: () => {
              applications.value = applications.value.filter(app => app.id !== applicationId);
            }
          });
        } else {
          applications.value = applications.value.filter(app => app.id !== applicationId);
        }
      }
    }
  } catch (error) {
    errorMessage.value = 'Failed to delete application. Please try again.';
  } finally {
    isDeleting.value = null;
  }
}

// Fetch applications
async function refreshData() {
  isLoading.value = true;
  errorMessage.value = '';
  
  try {
    const data = await api.getApplications(authenticatedPassword.value);
    
    // Animate the data update
    const oldLength = applications.value.length;
    applications.value = data;
    
    if (oldLength === 0 && data.length > 0) {
      // First data load, animate the table appearance
      nextTick(() => {
        gsap.from('table', {
          opacity: 0,
          y: 20,
          duration: 0.5,
          ease: 'power2.out'
        });
        
        gsap.from('.application-row', {
          opacity: 0,
          y: 15,
          stagger: 0.05,
          duration: 0.4,
          ease: 'power2.out'
        });
      });
    } else if (data.length > oldLength) {
      // New entries, animate only the new ones
      nextTick(() => {
        const rows = document.querySelectorAll('.application-row');
        for (let i = oldLength; i < rows.length; i++) {
          gsap.from(rows[i], {
            opacity: 0,
            y: 15,
            duration: 0.4,
            ease: 'power2.out',
            delay: i * 0.05
          });
        }
      });
    }
  } catch (error) {
    errorMessage.value = 'Failed to load applications. Please try again.';
  } finally {
    isLoading.value = false;
  }
}

// Export to CSV
async function exportToCsv() {
  try {
    await api.exportApplications(authenticatedPassword.value);
  } catch (error) {
    errorMessage.value = 'Failed to export applications. Please try again.';
  }
}

// Logout
function logout() {
  // Animate logout transition
  gsap.to(adminContainerRef.value, {
    opacity: 0,
    scale: 0.98,
    duration: 0.3,
    onComplete: () => {
      isAuthenticated.value = false;
      authenticatedPassword.value = '';
      applications.value = [];
      
      // After view changes, animate the login form in
      nextTick(() => {
        gsap.from('.login-container', {
          opacity: 0,
          scale: 0.95,
          duration: 0.4,
          ease: 'power2.out',
          onComplete: () => {
            // Focus password input
            if (passwordInput.value) {
              passwordInput.value.focus();
            }
          }
        });
      });
    }
  });
}

// Setup on component mount
onMounted(() => {
  // Focus password input when mounted
  if (passwordInput.value) {
    passwordInput.value.focus();
  }
});

// Format date for display
function formatDate(dateString, includeTime = false) {
  if (!dateString) return 'N/A';
  
  const date = new Date(dateString);
  const options = { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric',
    ...(includeTime ? { hour: '2-digit', minute: '2-digit' } : {})
  };
  
  return date.toLocaleDateString('en-US', options);
}
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.5s ease-out;
}

.animate-scale-in {
  animation: scaleIn 0.3s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes scaleIn {
  from { 
    opacity: 0; 
    transform: scale(0.95);
  }
  to { 
    opacity: 1;
    transform: scale(1);
  }
}

.tooltip-container {
  position: relative;
  cursor: default;
}

.tooltip {
  display: none;
  position: absolute;
  background-color: rgba(0, 0, 0, 0.9);
  border: 1px solid #555;
  color: white;
  padding: 0.5rem;
  border-radius: 0.25rem;
  z-index: 20;
  top: 100%;
  left: 0;
  white-space: normal;
  max-width: 300px;
  font-size: 0.75rem;
  margin-top: 0.25rem;
}

.tooltip-container:hover .tooltip {
  display: block;
}

.review-star {
  display: inline-block;
  animation: pulse 1.5s infinite;
  color: #ff5252;
  text-shadow: 0 0 5px rgba(255, 82, 82, 0.7);
}

@keyframes pulse {
  0% {
    opacity: 0.7;
    transform: scale(0.95);
  }
  50% {
    opacity: 1;
    transform: scale(1.1);
    text-shadow: 0 0 10px rgba(255, 82, 82, 0.9);
  }
  100% {
    opacity: 0.7;
    transform: scale(0.95);
  }
}

.review-btn {
  position: relative;
  overflow: hidden;
  box-shadow: 0 0 8px rgba(50, 205, 50, 0.5);
}

.review-btn::after {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(
    90deg,
    transparent,
    rgba(255, 255, 255, 0.3),
    transparent
  );
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% {
    left: -100%;
  }
  100% {
    left: 100%;
  }
}
</style> 