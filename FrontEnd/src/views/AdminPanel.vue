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
            <button @click="handleRefresh" class="mc-button text-sm">
              <span v-if="isLoading">Loading...</span>
              <span v-else>Refresh</span>
            </button>
            
            <button @click="handleExport" class="mc-button text-sm">
              Export CSV
            </button>
            
            <button @click="handleLogout" class="mc-button text-sm bg-red-500 hover:bg-red-600 border-red-700">
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
                  <th class="w-16 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap border-r border-minecraft-stone">#</th>
                  <th class="w-32 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap border-r border-minecraft-stone">DC</th>
                  <th class="w-32 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap border-r border-minecraft-stone">MC</th>
                  <th class="w-14 px-3 py-3 text-center text-sm font-minecraft text-white whitespace-nowrap border-r border-minecraft-stone">Age</th>
                  <th class="w-1/4 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap border-r border-minecraft-stone">FAM</th>
                  <th class="w-1/4 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap border-r border-minecraft-stone">SU</th>
                  <th class="w-20 px-3 py-3 text-center text-sm font-minecraft text-white whitespace-nowrap border-r border-minecraft-stone">S?</th>
                  <th class="w-36 px-3 py-3 text-left text-sm font-minecraft text-white whitespace-nowrap border-r border-minecraft-stone">Date</th>
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
                  <td class="px-3 py-2 text-sm border-r border-minecraft-stone/40" :class="{'text-minecraft-important-red font-bold': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="flex items-center justify-between">
                      <span v-if="!app.isReviewed" class="review-star animate-pulse text-minecraft-important-red w-5 flex-shrink-0">⭐</span>
                      <span class="ml-auto">{{ app.id }}</span>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm border-r border-minecraft-stone/40" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container">
                      <div class="truncate">{{ app.discordUsername }}</div>
                      <div class="tooltip">{{ app.discordUsername }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm border-r border-minecraft-stone/40" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container">
                      <div 
                        class="truncate cursor-pointer hover:text-minecraft-gold transition-colors duration-150"
                        @click="copyToClipboard(app.minecraftUsername)"
                        :class="{'copied': app.minecraftUsername === lastCopiedUsername}"
                      >
                        {{ app.minecraftUsername }}
                        <span class="copy-icon ml-1 opacity-0 group-hover:opacity-100">📋</span>
                      </div>
                      <div class="tooltip">{{ app.minecraftUsername }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm font-bold text-center border-r border-minecraft-stone/40" :class="getAgeColor(app.age)">
                    {{ app.age }}
                  </td>
                  
                  <td class="px-3 py-2 text-sm border-r border-minecraft-stone/40" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container">
                      <div class="truncate">{{ app.favoriteAboutMinecraft }}</div>
                      <div class="tooltip">{{ app.favoriteAboutMinecraft }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm border-r border-minecraft-stone/40" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container">
                      <div class="truncate">{{ app.understandingOfSMP }}</div>
                      <div class="tooltip">{{ app.understandingOfSMP }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm text-white text-center border-r border-minecraft-stone/40">
                    <span 
                      :class="app.joinedDiscord ? 'bg-minecraft-green' : 'bg-red-500'" 
                      class="px-2 py-1 rounded-full text-xs"
                    >
                      {{ app.joinedDiscord ? 'Yes' : 'No' }}
                    </span>
                  </td>
                  
                  <td class="px-3 py-2 text-sm border-r border-minecraft-stone/40" :class="{'text-minecraft-important-red': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container">
                      <div class="truncate">{{ formatDate(app.submissionDate) }}</div>
                      <div class="tooltip">{{ formatDate(app.submissionDate, true) }}</div>
                    </div>
                    <div v-if="app.isReviewed" class="text-xs mt-1 text-green-500">
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
          <!-- Player Info Panel -->
          <div class="bg-minecraft-deepslate/40 rounded-md p-4 mb-4 border border-minecraft-stone/50">
            <p class="text-white">
              <span class="font-minecraft">Player:</span> 
              {{ selectedApplication.minecraftUsername }} 
              <span class="text-gray-400">({{ selectedApplication.discordUsername }})</span>
            </p>
          </div>

          <div class="flex gap-2">
            <!-- Notes Panel -->
            <div class="flex-grow bg-minecraft-deepslate/40 rounded-md p-4 border border-minecraft-stone/50">
              <label class="mc-label block mb-2">Review Notes</label>
              <textarea 
                v-model="reviewNotes" 
                class="mc-input w-full bg-black/30" 
                rows="4"
                placeholder="Enter notes about this application (optional)"
              ></textarea>
            </div>

            <!-- Status Panel -->
            <div class="w-16 bg-minecraft-deepslate/40 rounded-md px-2 py-4 border border-minecraft-stone/50 flex flex-col items-center">
              <div class="relative h-40 w-full flex items-center justify-center">
                <!-- Slider Track Container -->
                <div class="relative h-full w-2 flex items-center justify-center">
                  <!-- Gradient Track -->
                  <div class="w-2 h-full rounded-full relative overflow-hidden">
                    <div class="absolute inset-0 bg-gradient-to-b from-green-500 via-yellow-500 to-red-500"></div>
                    <!-- Track Markers -->
                    <div class="absolute inset-x-[-4px] top-0 h-[2px] bg-white/20"></div>
                    <div class="absolute inset-x-[-4px] top-1/2 h-[2px] bg-white/20 -translate-y-px"></div>
                    <div class="absolute inset-x-[-4px] bottom-0 h-[2px] bg-white/20"></div>
                  </div>

                  <!-- Click Areas -->
                  <div class="absolute inset-x-[-12px] h-1/3 top-0 cursor-pointer" @click="setAcceptanceStatus('accepted')"></div>
                  <div class="absolute inset-x-[-12px] h-1/3 top-1/3 cursor-pointer" @click="setAcceptanceStatus('pending')"></div>
                  <div class="absolute inset-x-[-12px] h-1/3 top-2/3 cursor-pointer" @click="setAcceptanceStatus('rejected')"></div>

                  <!-- Slider Handle -->
                  <div 
                    ref="sliderHandle"
                    class="absolute w-7 h-7 left-[-12px] rounded-full shadow-lg cursor-grab active:cursor-grabbing will-change-transform flex items-center justify-center"
                    :class="{
                      'bg-green-600 border-green-400 shadow-green-500/30': acceptanceStatus === 'accepted',
                      'bg-yellow-600 border-yellow-400 shadow-yellow-500/30': acceptanceStatus === 'pending',
                      'bg-red-600 border-red-400 shadow-red-500/30': acceptanceStatus === 'rejected'
                    }"
                    :style="{
                      top: acceptanceStatus === 'accepted' ? '0%' : 
                           acceptanceStatus === 'pending' ? 'calc(50% - 14px)' : 'calc(100% - 28px)',
                      borderWidth: '2px',
                      transform: `scale(${isDragging ? 1.1 : 1})`,
                      transition: 'transform 0.15s ease-out, background-color 0.2s ease-out, border-color 0.2s ease-out'
                    }"
                    @mousedown="startDragging"
                  >
                    <!-- Handle Content -->
                    <span class="text-white text-sm font-minecraft leading-none select-none">
                      {{ acceptanceStatus === 'accepted' ? '✓' : 
                         acceptanceStatus === 'pending' ? '?' : '✕' }}
                    </span>

                    <!-- Tooltip -->
                    <div 
                      class="absolute left-full ml-2 whitespace-nowrap px-2 py-1 bg-black/90 text-white text-xs rounded pointer-events-none transform origin-left transition-all duration-200 border border-white/10"
                      :class="{
                        'opacity-0 scale-95 translate-x-2': isDragging,
                        'opacity-100 scale-100 translate-x-0': !isDragging
                      }"
                      :style="{ zIndex: 50 }"
                    >
                      {{ acceptanceStatus === 'accepted' ? 'Accept Application' : 
                         acceptanceStatus === 'pending' ? 'Pending Decision' : 'Reject Application' }}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- Action Buttons Panel -->
        <div class="flex justify-center gap-2 bg-minecraft-deepslate/40 rounded-md p-4 border border-minecraft-stone/50 mt-2">
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
import { ref, onMounted, onUnmounted, nextTick, inject } from 'vue';
import { gsap } from 'gsap';
import api from '../services/api';
import { useRouter } from 'vue-router';

const router = useRouter();
const confirmation = inject('confirmation');

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
const acceptanceStatus = ref('pending'); // Default to pending
const isSubmittingReview = ref(false);

// Notes modal state
const showNotesModal = ref(null);

// Add to script setup section, after the existing state declarations:
const sliderHandle = ref(null);
let isDragging = false;
let startY = 0;
let startTop = 0;
const lastCopiedUsername = ref('');

function startDragging(event) {
  isDragging = true;
  startY = event.clientY;
  const handleStyle = window.getComputedStyle(sliderHandle.value);
  startTop = parseInt(handleStyle.top);
  
  // Add event listeners for drag and release
  document.addEventListener('mousemove', handleDrag);
  document.addEventListener('mouseup', stopDragging);
  
  // Prevent text selection while dragging
  event.preventDefault();
}

function handleDrag(event) {
  if (!isDragging) return;
  
  const deltaY = event.clientY - startY;
  const parentHeight = sliderHandle.value.parentElement.offsetHeight - sliderHandle.value.offsetHeight;
  let newTop = Math.max(0, Math.min(parentHeight, startTop + deltaY));
  
  // Use requestAnimationFrame for smoother animations
  requestAnimationFrame(() => {
    sliderHandle.value.style.top = `${newTop}px`;
  });
  
  // Determine which third of the track we're in
  const third = parentHeight / 3;
  const newStatus = newTop < third ? 'accepted' : 
                   newTop < third * 2 ? 'pending' : 'rejected';
  
  if (acceptanceStatus.value !== newStatus) {
    acceptanceStatus.value = newStatus;
  }
}

function stopDragging() {
  isDragging = false;
  document.removeEventListener('mousemove', handleDrag);
  document.removeEventListener('mouseup', stopDragging);
  
  // Snap to position with animation
  setAcceptanceStatus(acceptanceStatus.value);
}

function setAcceptanceStatus(status) {
  const positions = {
    accepted: '0%',
    pending: 'calc(50% - 14px)',
    rejected: 'calc(100% - 28px)'
  };
  
  // Use requestAnimationFrame for smoother animations
  requestAnimationFrame(() => {
    gsap.to(sliderHandle.value, {
      top: positions[status],
      duration: 0.3,
      ease: 'back.out(1.7)',
      clearProps: 'transform',
      onComplete: () => {
        // Ensure final position is exact
        sliderHandle.value.style.top = positions[status];
      }
    });
  });
  
  acceptanceStatus.value = status;
}

// Clean up event listeners when component is destroyed
onUnmounted(() => {
  document.removeEventListener('mousemove', handleDrag);
  document.removeEventListener('mouseup', stopDragging);
});

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
  // Set initial acceptance status from existing data or default to pending
  acceptanceStatus.value = application.acceptanceStatus || 'pending';
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
      reviewNotes.value,
      acceptanceStatus.value
    );
    
    if (response.success) {
      showReviewModal.value = false;
      // Clear the selected application and review data
      selectedApplication.value = null;
      reviewNotes.value = '';
      acceptanceStatus.value = 'pending';
      // Refresh the data
      await refreshData();
    } else {
      errorMessage.value = 'Failed to review application. Please try again.';
    }
  } catch (error) {
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
      await refreshData();
    } else {
      errorMessage.value = 'Failed to unreview application. Please try again.';
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
      await refreshData();
    } else {
      errorMessage.value = 'Failed to delete application. Please try again.';
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
    
    if (response.success) {
      const oldLength = applications.value.length;
      // Sort applications in reverse order by ID
      applications.value = response.data.sort((a, b) => b.id - a.id);
      
      if (oldLength === 0 && response.data.length > 0) {
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

          // Scroll to top after animation
          setTimeout(() => {
            const container = document.querySelector('.overflow-auto');
            if (container) {
              container.scrollTo({ top: 0, behavior: 'smooth' });
            }
          }, 500);
        });
      } else if (response.data.length > oldLength) {
        // New entries, animate only the new ones
        nextTick(() => {
          const rows = document.querySelectorAll('.application-row');
          for (let i = 0; i < response.data.length - oldLength; i++) {
            gsap.from(rows[i], {
              opacity: 0,
              y: 15,
              duration: 0.4,
              ease: 'power2.out',
              delay: i * 0.05
            });
          }

          // Scroll to top after animation
          setTimeout(() => {
            const container = document.querySelector('.overflow-auto');
            if (container) {
              container.scrollTo({ top: 0, behavior: 'smooth' });
            }
          }, 500);
        });
      }
    } else {
      errorMessage.value = 'Failed to load applications. Please try again.';
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

// Add to script setup section, after the existing functions:
async function copyToClipboard(text) {
  try {
    await navigator.clipboard.writeText(text);
    lastCopiedUsername.value = text;
    
    // Reset the copied state after 2 seconds
    setTimeout(() => {
      lastCopiedUsername.value = '';
    }, 2000);
  } catch (err) {
    console.error('Failed to copy text: ', err);
  }
}

// Add this new function after the existing functions
function updateTooltipPosition(event) {
  const tooltip = event.currentTarget.querySelector('.tooltip');
  if (tooltip) {
    const rect = event.currentTarget.getBoundingClientRect();
    const tooltipHeight = tooltip.offsetHeight;
    const tooltipWidth = tooltip.offsetWidth;
    
    // Position tooltip above or below based on available space
    let top = rect.bottom + window.scrollY + 5;
    
    // If tooltip would go off bottom of viewport, show it above instead
    if (top + tooltipHeight > window.innerHeight) {
      top = rect.top + window.scrollY - tooltipHeight - 5;
    }
    
    // Center horizontally relative to the container
    let left = rect.left + (rect.width / 2) - (tooltipWidth / 2);
    
    // Ensure tooltip doesn't go off screen horizontally
    left = Math.max(10, Math.min(left, window.innerWidth - tooltipWidth - 10));
    
    tooltip.style.top = `${top}px`;
    tooltip.style.left = `${left}px`;
  }
}

// Add mounted hook to set up event listeners
onMounted(() => {
  // ... existing mounted code ...
  
  // Add event listeners for tooltip positioning
  const tooltipContainers = document.querySelectorAll('.tooltip-container');
  tooltipContainers.forEach(container => {
    container.addEventListener('mouseenter', updateTooltipPosition);
    container.addEventListener('mousemove', updateTooltipPosition);
  });
});

// Add cleanup in onUnmounted
onUnmounted(() => {
  // ... existing unmounted code ...
  
  // Remove tooltip event listeners
  const tooltipContainers = document.querySelectorAll('.tooltip-container');
  tooltipContainers.forEach(container => {
    container.removeEventListener('mouseenter', updateTooltipPosition);
    container.removeEventListener('mousemove', updateTooltipPosition);
  });
});

const handleLogout = async () => {
  const confirmed = await confirmation.confirm({
    title: 'Logout',
    message: 'Are you sure you want to logout?',
    confirmText: 'Logout',
    cancelText: 'Stay'
  });
  
  if (confirmed) {
    logout();
  }
};

const handleExport = async () => {
  const confirmed = await confirmation.confirm({
    title: 'Export Data',
    message: 'This will export all application data to a CSV file. Continue?',
    confirmText: 'Export',
    cancelText: 'Cancel'
  });
  
  if (confirmed) {
    exportToCsv();
  }
};

const handleRefresh = async () => {
  const confirmed = await confirmation.confirm({
    title: 'Refresh Data',
    message: 'This will refresh all application data. Continue?',
    confirmText: 'Refresh',
    cancelText: 'Cancel'
  });
  
  if (confirmed) {
    refreshData();
  }
};
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
  position: fixed;  /* Changed from absolute to fixed for better positioning */
  background-color: rgba(0, 0, 0, 0.9);
  border: 1px solid #555;
  color: white;
  padding: 0.5rem;
  border-radius: 0.25rem;
  z-index: 9999;  /* Increased z-index to ensure it's above all other elements */
  white-space: normal;
  max-width: 300px;
  font-size: 0.75rem;
  margin-top: 0.25rem;
  pointer-events: none;  /* Prevents tooltip from interfering with other interactions */
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);  /* Added shadow for better visibility */
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

/* Add some additional table styling for better readability */
table {
  border-collapse: separate;
  border-spacing: 0;
}

th, td {
  position: relative;
}

/* Slightly lighter backgrounds for better contrast between rows */
tbody tr:nth-child(even) {
  background-color: rgba(30, 30, 30, 0.4) !important;
}

tbody tr:hover {
  background-color: rgba(60, 100, 140, 0.3) !important;
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