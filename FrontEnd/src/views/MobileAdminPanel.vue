<template>
  <div class="h-full w-full absolute inset-0 flex items-center justify-center">
    <!-- Login Form -->
    <div v-if="!isAuthenticated" class="login-container mc-container animate-fade-in max-w-md w-full mx-4">
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
          <button type="submit" class="mc-button w-full">
            Login
          </button>
        </div>
      </form>
    </div>
    
    <!-- Admin Dashboard -->
    <div v-else class="absolute inset-0 p-2">
      <div ref="adminContainerRef" class="admin-panel mc-panel w-full h-full flex flex-col">
        <!-- Header -->
        <div class="flex flex-col gap-2 px-4 py-3 bg-minecraft-deepslate/80 rounded-t-md">
          <h2 class="mc-title mb-0">Applications Dashboard</h2>
          
          <div class="flex flex-wrap gap-2">
            <button @click="refreshData" class="mc-button text-sm flex-1">
              <span v-if="isLoading">Loading...</span>
              <span v-else>Refresh</span>
            </button>
            
            <button @click="exportToCsv" class="mc-button text-sm flex-1">
              Export CSV
            </button>
            
            <button @click="logout" class="mc-button text-sm bg-red-500 hover:bg-red-600 border-red-700 flex-1">
              Logout
            </button>
          </div>
        </div>
        
        <div v-if="errorMessage" class="bg-red-500/50 text-white p-3 mx-4 my-2 rounded-md">
          {{ errorMessage }}
        </div>
        
        <!-- Applications List -->
        <div class="flex-grow overflow-auto p-2 bg-black/40">
          <div v-if="applications.length === 0 && !isLoading" class="text-center py-10 text-white flex-grow flex items-center justify-center">
            <div>
              <p class="text-xl font-minecraft">No applications yet</p>
              <p class="mt-2">Applications will appear here once submitted</p>
            </div>
          </div>
          
          <div v-else class="space-y-2">
            <div 
              v-for="(app, index) in applications" 
              :key="app.id"
              class="application-card bg-black/50 border border-minecraft-stone rounded-md p-3"
              :class="{'bg-minecraft-important-red/10': !app.isReviewed}"
            >
              <!-- Header -->
              <div class="flex items-center justify-between mb-2">
                <div class="flex items-center gap-2">
                  <span v-if="!app.isReviewed" class="review-star animate-pulse text-minecraft-important-red">⭐</span>
                  <span class="font-minecraft" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    #{{ app.id }}
                  </span>
                </div>
                <span 
                  :class="app.joinedDiscord ? 'bg-minecraft-green' : 'bg-red-500'" 
                  class="px-2 py-1 rounded-full text-xs"
                >
                  {{ app.joinedDiscord ? 'Yes' : 'No' }}
                </span>
              </div>
              
              <!-- Player Info -->
              <div class="space-y-1 mb-2">
                <div class="flex items-center gap-2">
                  <span class="text-minecraft-gold">DC:</span>
                  <span :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    {{ app.discordUsername }}
                  </span>
                </div>
                <div class="flex items-center gap-2">
                  <span class="text-minecraft-gold">MC:</span>
                  <span 
                    class="cursor-pointer hover:text-minecraft-gold transition-colors duration-150"
                    :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed, 'copied': app.minecraftUsername === lastCopiedUsername}"
                    @click="copyToClipboard(app.minecraftUsername)"
                  >
                    {{ app.minecraftUsername }}
                    <span class="copy-icon ml-1 opacity-0 group-hover:opacity-100">📋</span>
                  </span>
                </div>
                <div class="flex items-center gap-2">
                  <span class="text-minecraft-gold">Age:</span>
                  <span :class="getAgeColor(app.age)" class="font-bold">
                    {{ app.age }}
                  </span>
                </div>
              </div>
              
              <!-- Application Content -->
              <div class="space-y-2">
                <div>
                  <span class="text-minecraft-gold block mb-1">FAM:</span>
                  <p class="text-sm" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    {{ app.favoriteAboutMinecraft }}
                  </p>
                </div>
                <div>
                  <span class="text-minecraft-gold block mb-1">SU:</span>
                  <p class="text-sm" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    {{ app.understandingOfSMP }}
                  </p>
                </div>
              </div>
              
              <!-- Date and Status -->
              <div class="mt-2 pt-2 border-t border-minecraft-stone/40">
                <div class="text-xs" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                  Submitted: {{ formatDate(app.submissionDate) }}
                </div>
                <div v-if="app.isReviewed" class="text-xs text-green-500">
                  Reviewed: {{ formatDate(app.reviewedAt) }}
                </div>
                <div v-if="app.isReviewed && app.acceptanceStatus" class="mt-1">
                  <span 
                    class="text-xs px-2 py-0.5 rounded-full"
                    :class="{
                      'bg-green-600 text-white': app.acceptanceStatus === 'accepted',
                      'bg-red-600 text-white': app.acceptanceStatus === 'rejected',
                      'bg-yellow-600 text-white': app.acceptanceStatus === 'pending'
                    }"
                  >
                    {{ app.acceptanceStatus === 'accepted' ? 'Accepted' : 
                       app.acceptanceStatus === 'rejected' ? 'Rejected' : 'Pending' }}
                  </span>
                </div>
              </div>
              
              <!-- Actions -->
              <div class="flex flex-wrap gap-2 mt-3">
                <button 
                  v-if="!app.isReviewed"
                  @click="openReviewModal(app)"
                  class="mc-button text-xs bg-green-600 hover:bg-green-700 border-green-800 px-2 py-1 review-btn animate-pulse flex-1"
                  :disabled="isProcessing === app.id"
                >
                  <span v-if="isProcessing === app.id">...</span>
                  <span v-else>Review</span>
                </button>
                
                <button 
                  v-if="app.isReviewed"
                  @click="confirmUnreview(app.id)"
                  class="mc-button text-xs bg-minecraft-important-red hover:bg-red-700 border-red-800 px-2 py-1 flex-1"
                  :disabled="isProcessing === app.id"
                >
                  <span v-if="isProcessing === app.id">...</span>
                  <span v-else>Unreview</span>
                </button>
                
                <button 
                  @click="confirmDelete(app.id)"
                  class="mc-button text-xs bg-red-500 hover:bg-red-600 border-red-700 px-2 py-1 flex-1"
                  :disabled="isDeleting === app.id"
                >
                  <span v-if="isDeleting === app.id">...</span>
                  <span v-else>Delete</span>
                </button>
                
                <button 
                  v-if="app.isReviewed && app.reviewNotes"
                  @click="showNotes(app)"
                  class="mc-button text-xs bg-minecraft-deepslate hover:bg-gray-700 border-gray-800 px-2 py-1 flex-1"
                >
                  Notes
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Review Modal -->
    <div v-if="showReviewModal" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
      <div class="mc-container w-full animate-scale-in">
        <div class="flex justify-between items-center mb-4">
          <h3 class="mc-title text-lg">Review Application</h3>
          <button @click="showReviewModal = false" class="text-white hover:text-red-500 transition">✕</button>
        </div>
        
        <div v-if="selectedApplication" class="mb-4">
          <!-- Player Info Panel -->
          <div class="bg-minecraft-deepslate/40 rounded-md p-4 mb-4 border border-minecraft-stone/50">
            <p class="text-white">
              <span class="font-minecraft">Player:</span> 
              {{ selectedApplication.minecraftUsername }} 
              <span class="text-gray-400">({{ selectedApplication.discordUsername }})</span>
            </p>
          </div>

          <div class="flex flex-col gap-4">
            <!-- Notes Panel -->
            <div class="bg-minecraft-deepslate/40 rounded-md p-4 border border-minecraft-stone/50">
              <label class="mc-label block mb-2">Review Notes</label>
              <textarea 
                v-model="reviewNotes" 
                class="mc-input w-full bg-black/30" 
                rows="4"
                placeholder="Enter notes about this application (optional)"
              ></textarea>
            </div>

            <!-- Status Panel -->
            <div class="bg-minecraft-deepslate/40 rounded-md p-4 border border-minecraft-stone/50">
              <label class="mc-label block mb-2">Status</label>
              <div class="flex gap-2">
                <button 
                  v-for="status in ['accepted', 'pending', 'rejected']"
                  :key="status"
                  @click="setAcceptanceStatus(status)"
                  class="flex-1 mc-button"
                  :class="{
                    'bg-green-600 hover:bg-green-700 border-green-800': status === 'accepted',
                    'bg-yellow-600 hover:bg-yellow-700 border-yellow-800': status === 'pending',
                    'bg-red-600 hover:bg-red-700 border-red-800': status === 'rejected'
                  }"
                >
                  {{ status.charAt(0).toUpperCase() + status.slice(1) }}
                </button>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Action Buttons Panel -->
        <div class="flex justify-center gap-2 bg-minecraft-deepslate/40 rounded-md p-4 border border-minecraft-stone/50 mt-2">
          <button @click="showReviewModal = false" class="mc-button bg-gray-600 hover:bg-gray-700 border-gray-800 flex-1">
            Cancel
          </button>
          <button 
            @click="submitReview"
            class="mc-button bg-green-600 hover:bg-green-700 border-green-800 flex-1"
            :disabled="isSubmittingReview"
          >
            <span v-if="isSubmittingReview">Submitting...</span>
            <span v-else>Submit Review</span>
          </button>
        </div>
      </div>
    </div>
    
    <!-- Notes Modal -->
    <div v-if="showNotesModal" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
      <div class="mc-container w-full animate-scale-in">
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
          <button @click="showNotesModal = false" class="mc-button w-full">
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, nextTick, inject } from 'vue';
import { gsap } from 'gsap';
import api from '../services/api';
import { useRouter } from 'vue-router';

const router = useRouter();
const confirmation = inject('confirmation');

// UI state
const isLoading = ref(true);
const errorMessage = ref('');
const applications = ref([]);

// Authentication state
const isAuthenticated = ref(false);
const password = ref('');
const authError = ref('');
const passwordInput = ref(null);
const authenticatedPassword = ref('');
const adminContainerRef = ref(null);
const isDeleting = ref(null);
const isProcessing = ref(null);

// Review modal state
const showReviewModal = ref(false);
const selectedApplication = ref(null);
const reviewNotes = ref('');
const acceptanceStatus = ref('pending');
const isSubmittingReview = ref(false);

// Notes modal state
const showNotesModal = ref(false);

// Clipboard state
const lastCopiedUsername = ref('');

// Authenticate user
async function authenticate() {
  try {
    const response = await api.verifyAdminPassword(password.value);
    
    if (response.success) {
      authenticatedPassword.value = password.value;
      
      gsap.to('.login-container', {
        opacity: 0,
        scale: 0.95,
        duration: 0.3,
        onComplete: () => {
          isAuthenticated.value = true;
          password.value = '';
          authError.value = '';
          
          nextTick(() => {
            if (adminContainerRef.value) {
              gsap.from(adminContainerRef.value, {
                opacity: 0,
                scale: 0.98,
                duration: 0.4,
                ease: 'power2.out'
              });
            }
            
            refreshData();
          });
        }
      });
    } else {
      authError.value = 'Invalid password';
      
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
  acceptanceStatus.value = application.acceptanceStatus || 'pending';
  showReviewModal.value = true;
}

// Show notes modal
function showNotes(application) {
  selectedApplication.value = application;
  showNotesModal.value = true;
}

// Set acceptance status
function setAcceptanceStatus(status) {
  acceptanceStatus.value = status;
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
      reviewNotes.value,
      acceptanceStatus.value
    );
    
    // Validate the response has the required structure
    if (response && 
        typeof response === 'object' && 
        response.success === true && 
        response.data && 
        typeof response.data === 'object' && 
        'id' in response.data && 
        'isReviewed' in response.data && 
        'reviewedAt' in response.data) {
      
      const index = applications.value.findIndex(app => app.id === selectedApplication.value.id);
      if (index !== -1) {
        // Ensure acceptanceStatus is set
        if (!response.data.acceptanceStatus) {
          response.data.acceptanceStatus = acceptanceStatus.value;
        }
        
        applications.value[index] = response.data;
        
        nextTick(() => {
          const cards = document.querySelectorAll('.application-card');
          if (cards[index]) {
            gsap.fromTo(cards[index], 
              { backgroundColor: 'rgba(50, 205, 50, 0.3)' },
              { backgroundColor: '', duration: 1.5, ease: 'power2.out' }
            );
          }
        });
      }
      
      showReviewModal.value = false;
    } else {
      console.error('Invalid response structure:', response);
      throw new Error('Invalid response from server');
    }
  } catch (error) {
    console.error('Error submitting review:', error);
    errorMessage.value = 'Failed to review application. Please try again.';
  } finally {
    isSubmittingReview.value = false;
    isProcessing.value = null;
  }
}

// Confirm unreview
async function confirmUnreview(applicationId) {
  const confirmed = await confirmation.confirm({
    title: 'Remove Review Status',
    message: 'Are you sure you want to remove the reviewed status?',
    confirmText: 'Remove',
    cancelText: 'Keep'
  });
  
  if (confirmed) {
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
      const index = applications.value.findIndex(app => app.id === applicationId);
      if (index !== -1) {
        applications.value[index] = response.data;
        
        nextTick(() => {
          const cards = document.querySelectorAll('.application-card');
          if (cards[index]) {
            gsap.fromTo(cards[index], 
              { backgroundColor: 'rgba(255, 82, 82, 0.3)' },
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
async function confirmDelete(applicationId) {
  const confirmed = await confirmation.confirm({
    title: 'Delete Application',
    message: 'Are you sure you want to delete this application? This action cannot be undone.',
    confirmText: 'Delete',
    cancelText: 'Cancel'
  });
  
  if (confirmed) {
    deleteApplication(applicationId);
  }
}

// Delete application
async function deleteApplication(applicationId) {
  isDeleting.value = applicationId;
  
  try {
    const response = await api.deleteApplication(applicationId, authenticatedPassword.value);
    
    if (response.success) {
      const index = applications.value.findIndex(app => app.id === applicationId);
      if (index !== -1) {
        const cards = document.querySelectorAll('.application-card');
        if (cards[index]) {
          gsap.to(cards[index], {
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
    const response = await api.getApplications(authenticatedPassword.value);
    
    if (response && response.success) {
      const oldLength = applications.value.length;
      // Sort applications in reverse order by ID, handle null response
      applications.value = (response.data || []).sort((a, b) => b.id - a.id);
      
      if (oldLength === 0 && applications.value.length > 0) {
        nextTick(() => {
          const cards = document.querySelectorAll('.application-card');
          if (cards.length > 0) {
            gsap.from(cards, {
              opacity: 0,
              y: 20,
              stagger: 0.05,
              duration: 0.4,
              ease: 'power2.out'
            });
          }
        });
      }
    } else {
      applications.value = [];
    }
  } catch (error) {
    console.error('Error fetching applications:', error);
    errorMessage.value = 'Failed to fetch applications. Please try again.';
    applications.value = [];
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
  gsap.to(adminContainerRef.value, {
    opacity: 0,
    scale: 0.98,
    duration: 0.3,
    onComplete: () => {
      isAuthenticated.value = false;
      authenticatedPassword.value = '';
      applications.value = [];
      
      nextTick(() => {
        gsap.from('.login-container', {
          opacity: 0,
          scale: 0.95,
          duration: 0.4,
          ease: 'power2.out',
          onComplete: () => {
            if (passwordInput.value) {
              passwordInput.value.focus();
            }
          }
        });
      });
    }
  });
}

// Copy to clipboard
async function copyToClipboard(text) {
  try {
    await navigator.clipboard.writeText(text);
    lastCopiedUsername.value = text;
    
    setTimeout(() => {
      lastCopiedUsername.value = '';
    }, 2000);
  } catch (err) {
    console.error('Failed to copy text: ', err);
  }
}

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

// Call refreshData on component mount
onMounted(() => {
  if (isAuthenticated.value) {
    refreshData();
  }
});
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

.copied {
  color: #ffd700 !important;
  animation: copyPulse 0.5s ease-out;
}

.copy-icon {
  transition: opacity 0.2s ease-out;
}

.cursor-pointer:hover .copy-icon {
  opacity: 1 !important;
}

@keyframes copyPulse {
  0% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
  100% {
    transform: scale(1);
  }
}
</style> 