<template>
  <div class="h-full w-full absolute inset-0 flex items-center justify-center">
    <!-- Tooltip Container - Added to ensure tooltips are always on top -->
    <div id="tooltip-root" class="fixed inset-0 pointer-events-none z-[99999]"></div>
    
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
        
        <div v-if="authError" class="bg-red-600/40 text-white p-3 rounded-lg border border-red-500/30">
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
        <!-- Fixed top bar -->
        <div class="flex justify-between items-center px-6 py-4 bg-black/70 sticky top-0 z-20 border-b border-white/10">
          <h2 class="mc-title mb-0">TYFORMS</h2>
          
          <div class="flex gap-2">
            <button @click="handleRefresh" class="mc-button text-sm">
              <span v-if="isLoading">Loading...</span>
              <span v-else>Refresh</span>
            </button>
            
            <button @click="handleExport" class="mc-button text-sm secondary">
              Export CSV
            </button>
            
            <button @click="handleLogout" class="mc-button text-sm danger">
              Logout
            </button>
          </div>
        </div>
        
        <div v-if="errorMessage" class="bg-red-600/40 text-white p-3 mx-6 my-4 rounded-lg border border-red-500/30">
          {{ errorMessage }}
        </div>
        
        <!-- Scrollable content area -->
        <div class="flex-grow overflow-hidden flex flex-col p-6 bg-black/80">
          <div v-if="applications.length === 0 && !isLoading" class="text-center py-10 text-white flex-grow flex items-center justify-center">
            <div class="glass p-6 rounded-xl">
              <p class="text-xl font-medium">No applications yet</p>
              <p class="mt-2 text-neutral-300">Applications will appear here once submitted</p>
            </div>
          </div>
          
          <div v-else class="overflow-auto flex-grow">
            <!-- Rounded wrapper to clip inner table corners -->
            <!-- Use overflow-clip for better Firefox clipping with sticky header -->
            <div class="rounded-lg border border-white/10 bg-black/60" style="overflow: clip;">
              <table class="w-full table-fixed">
                <thead class="sticky top-0 z-10">
                  <tr class="bg-black/70">
                  <th class="w-20 px-3 py-3 text-left text-xs font-medium text-neutral-200 uppercase tracking-wide whitespace-nowrap border-r border-white/10">#</th>
                  <th class="w-32 px-3 py-3 text-left text-xs font-medium text-neutral-200 uppercase tracking-wide whitespace-nowrap border-r border-white/10">DC</th>
                  <th class="w-32 px-3 py-3 text-left text-xs font-medium text-neutral-200 uppercase tracking-wide whitespace-nowrap border-r border-white/10">MC</th>
                  <th class="w-14 px-3 py-3 text-center text-xs font-medium text-neutral-200 uppercase tracking-wide whitespace-nowrap border-r border-white/10">Age</th>
                  <th class="w-1/4 px-3 py-3 text-left text-xs font-medium text-neutral-200 uppercase tracking-wide whitespace-nowrap border-r border-white/10">FAM</th>
                  <th class="w-1/4 px-3 py-3 text-left text-xs font-medium text-neutral-200 uppercase tracking-wide whitespace-nowrap border-r border-white/10">SU</th>
                  <th class="w-20 px-3 py-3 text-center text-xs font-medium text-neutral-200 uppercase tracking-wide whitespace-nowrap border-r border-white/10">S?</th>
                  <th class="w-36 px-3 py-3 text-left text-xs font-medium text-neutral-200 uppercase tracking-wide whitespace-nowrap border-r border-white/10">Date</th>
                  <th class="w-32 px-3 py-3 text-center text-xs font-medium text-neutral-200 uppercase tracking-wide whitespace-nowrap">Actions</th>
                </tr>
              </thead>
              
              <tbody class="divide-y divide-white/5">
                <tr 
                  v-for="(app, index) in filteredApplications" 
                  :key="app.id" 
                  :class="[
                    {'bg-black/10': index % 2 === 0},
                    {'bg-red-500/10': !app.isReviewed}
                  ]"
                  class="hover:bg-primary-400/10 transition-colors duration-150 application-row"
                >
                  <td class="px-3 py-2 text-sm" :class="{'text-minecraft-important-red font-bold': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="flex items-center justify-between">
                      <span v-if="!app.isReviewed" class="review-star animate-pulse text-minecraft-important-red w-5 flex-shrink-0">⭐</span>
                      <span class="ml-auto">{{ app.id }}</span>
                    </div>
                  </td>
                  
                   <td class="px-3 py-2 text-sm" :class="{'text-red-400': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container" :data-tooltip="app.discordUsername">
                      <div class="truncate">{{ app.discordUsername }}</div>
                    </div>
                  </td>
                  
                   <td class="px-3 py-2 text-sm" :class="{'text-red-400': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container" :data-tooltip="app.minecraftUsername">
                      <div
                        class="truncate cursor-pointer hover:text-primary-300 transition-colors duration-150"
                        @click="copyToClipboard(app.minecraftUsername)"
                        :class="{'copied': app.minecraftUsername === lastCopiedUsername}"
                      >
                        {{ app.minecraftUsername }}
                        <span class="copy-icon ml-1 opacity-0 group-hover:opacity-100">📋</span>
                      </div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm font-bold text-center" :class="getAgeColor(app.age)">
                    {{ app.age }}
                  </td>
                  
                   <td class="px-3 py-2 text-sm" :class="{'text-red-400': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container" :data-tooltip="app.favoriteAboutMinecraft">
                      <div class="truncate">{{ app.favoriteAboutMinecraft }}</div>
                    </div>
                  </td>

                  <td class="px-3 py-2 text-sm" :class="{'text-red-400': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container" :data-tooltip="app.understandingOfSMP">
                      <div class="truncate">{{ app.understandingOfSMP }}</div>
                    </div>
                  </td>
                  
                  <td class="px-3 py-2 text-sm text-white text-center">
                    <span 
                      :class="app.joinedDiscord ? 'bg-green-600' : 'bg-red-600'" 
                      class="px-2 py-1 rounded-full text-xs"
                    >
                      {{ app.joinedDiscord ? 'Yes' : 'No' }}
                    </span>
                  </td>
                  
                   <td class="px-3 py-2 text-sm" :class="{'text-red-400': !app.isReviewed, 'text-white': app.isReviewed}">
                    <div class="tooltip-container" :data-tooltip="formatDate(app.submissionDate, true)">
                      <div class="truncate">{{ formatDate(app.submissionDate) }}</div>
                    </div>
                    <div v-if="app.isReviewed" class="text-xs mt-1 text-green-400">
                      Reviewed: {{ formatDate(app.reviewedAt) }}
                    </div>
                    <div v-if="app.isReviewed && app.acceptanceStatus" class="mt-1">
                      <span
                        class="text-xs px-2 py-0.5 rounded-full"
                        :class="{
                          'bg-green-600 text-white': app.acceptanceStatus === 'accepted',
                          'bg-red-600 text-white': app.acceptanceStatus === 'rejected',
                          'bg-yellow-500 text-black': app.acceptanceStatus === 'pending'
                        }"
                      >
                        {{ app.acceptanceStatus === 'accepted' ? 'Accepted' :
                           app.acceptanceStatus === 'rejected' ? 'Rejected' : 'Pending' }}
                      </span>
                    </div>
                  </td>
                  
                   <td class="px-3 py-2 text-sm text-white text-center">
                    <!-- Reviewed with Accepted or Rejected status -->
                    <div v-if="app.isReviewed && (app.acceptanceStatus === 'accepted' || app.acceptanceStatus === 'rejected')" class="grid grid-rows-2 gap-2">
                      <button
                        @click="confirmDelete(app.id)"
                        class="mc-button danger px-2 py-1 border border-white/20"
                        :disabled="isDeleting === app.id"
                        :title="'Delete'"
                      >
                        <span v-if="isDeleting === app.id">...</span>
                        <span v-else class="text-sm">🗑</span>
                      </button>

                      <button
                        @click="confirmUnreview(app.id)"
                        class="mc-button secondary px-2 py-1 border border-white/30"
                        :disabled="isProcessing === app.id"
                        :title="'Unreview'"
                      >
                        <span v-if="isProcessing === app.id">...</span>
                        <span v-else class="text-sm">↶</span>
                      </button>
                    </div>

                    <!-- Reviewed with Pending status -->
                    <div v-else-if="app.isReviewed && app.acceptanceStatus === 'pending'" class="grid grid-cols-2 gap-2">
                      <button
                        @click="quickReview(app, 'accepted', $event)"
                        @click.shift="openNotesModal(app, 'accepted')"
                        class="mc-button px-2 py-1 bg-green-600 border border-green-400/30"
                        :disabled="isProcessing === app.id"
                        :title="'Accept'"
                      >
                        <span v-if="isProcessing === app.id">...</span>
                        <span v-else class="text-sm">✓</span>
                      </button>

                      <button
                        @click="quickReview(app, 'rejected', $event)"
                        @click.shift="openNotesModal(app, 'rejected')"
                        class="mc-button px-2 py-1 bg-red-600 border border-red-400/30"
                        :disabled="isProcessing === app.id"
                        :title="'Reject'"
                      >
                        <span v-if="isProcessing === app.id">...</span>
                        <span v-else class="text-sm">✕</span>
                      </button>

                      <button
                        @click="confirmDelete(app.id)"
                        class="col-span-2 mc-button danger px-2 py-1 border border-white/20"
                        :disabled="isDeleting === app.id"
                        :title="'Delete'"
                      >
                        <span v-if="isDeleting === app.id">...</span>
                        <span v-else class="text-sm">🗑</span>
                      </button>
                    </div>

                    <!-- Unreviewed applications -->
                    <div v-else class="grid grid-cols-2 gap-2">
                      <button
                        @click="quickReview(app, 'accepted', $event)"
                        @click.shift="openNotesModal(app, 'accepted')"
                        class="mc-button px-2 py-1 bg-green-600 border border-green-400/30"
                        :disabled="isProcessing === app.id"
                        :title="'Accept'"
                      >
                        <span v-if="isProcessing === app.id">...</span>
                        <span v-else class="text-sm">✓</span>
                      </button>

                      <button
                        @click="quickReview(app, 'pending', $event)"
                        @click.shift="openNotesModal(app, 'pending')"
                        class="mc-button px-2 py-1 bg-yellow-600 border border-yellow-400/30"
                        :disabled="isProcessing === app.id"
                        :title="'Pending'"
                      >
                        <span v-if="isProcessing === app.id">...</span>
                        <span v-else class="text-sm">•</span>
                      </button>

                      <button
                        @click="quickReview(app, 'rejected', $event)"
                        @click.shift="openNotesModal(app, 'rejected')"
                        class="mc-button px-2 py-1 bg-red-600 border border-red-400/30"
                        :disabled="isProcessing === app.id"
                        :title="'Reject'"
                      >
                        <span v-if="isProcessing === app.id">...</span>
                        <span v-else class="text-sm">✕</span>
                      </button>

                      <button
                        @click="confirmDelete(app.id)"
                        class="mc-button danger px-2 py-1 border border-white/20"
                        :disabled="isDeleting === app.id"
                        :title="'Delete'"
                      >
                        <span v-if="isDeleting === app.id">...</span>
                        <span v-else class="text-sm">🗑</span>
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Infinite Scroll Loading Indicator -->
          <div v-if="isLoadingMore" class="mt-4 text-center py-4">
            <div class="text-sm text-neutral-300">Loading more applications...</div>
          </div>

          <!-- End of results indicator -->
          <div v-else-if="applications.length > 0 && applications.length >= totalApplications" class="mt-4 text-center py-4">
            <div class="text-sm text-neutral-400">
              Showing all {{ totalApplications }} applications
              <span v-if="searchQuery.trim()" class="ml-2 text-primary-400">(filtered)</span>
            </div>
          </div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Floating Search Panel -->
    <Teleport to="body">
      <Transition name="slide-up">
        <div v-if="showSearchPanel" class="fixed inset-x-0 bottom-0 z-[9999] px-4 pb-6 pointer-events-none">
          <div class="mx-auto max-w-3xl pointer-events-auto">
            <div class="mc-container bg-black/80 border border-white/10 rounded-xl shadow-xl">
              <div class="flex items-center justify-between mb-3">
                <h3 class="mc-title text-base mb-0">Search</h3>
                <button @click="closeSearchPanel" class="text-white hover:text-red-400 transition">✕</button>
              </div>
              <div class="flex items-center gap-3">
                <input
                  ref="searchInputRef"
                  v-model="searchQuery"
                  type="text"
                  class="mc-input flex-1"
                  placeholder="Type to search... (Esc to close)"
                />
              </div>
              <div class="flex flex-wrap gap-4 mt-3 text-xs text-neutral-200">
                <label class="flex items-center gap-1">
                  <input type="checkbox" v-model="searchFields.discordUsername" />
                  <span>Discord</span>
                </label>
                <label class="flex items-center gap-1">
                  <input type="checkbox" v-model="searchFields.minecraftUsername" />
                  <span>Minecraft</span>
                </label>
                <label class="flex items-center gap-1">
                  <input type="checkbox" v-model="searchFields.favoriteAboutMinecraft" />
                  <span>FAM</span>
                </label>
                <label class="flex items-center gap-1">
                  <input type="checkbox" v-model="searchFields.understandingOfSMP" />
                  <span>SU</span>
                </label>
                <label class="flex items-center gap-1">
                  <input type="checkbox" v-model="searchFields.id" />
                  <span>#</span>
                </label>
              </div>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    
    <!-- Notes Modal -->
    <Teleport to="body">
      <Transition name="fade">
        <NotesModal
          v-if="showNotesModal"
          :show="showNotesModal"
          :application="selectedApplication"
          :initial-status="notesModalStatus"
          @close="showNotesModal = false"
          @save="handleNotesSave"
        />
      </Transition>
    </Teleport>

    <!-- Notes Display Modal -->
    <div v-if="showNotesModal" class="fixed inset-0 bg-black/80 flex items-center justify-center z-50">
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
import { ref, onMounted, onUnmounted, nextTick, inject, computed, watch } from 'vue';
import { gsap } from 'gsap';
import api from '../services/api';
import { useRouter } from 'vue-router';
import NotesModal from '../components/NotesModal.vue';

const router = useRouter();
const confirmation = inject('confirmation');
const notification = inject('notification');

// UI state
const isLoading = ref(false);
const isLoadingMore = ref(false);
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

// Notes modal state
const showNotesModal = ref(false);
const notesModalApp = ref(null);
const notesModalStatus = ref('');
const notesModalNotes = ref('');
const selectedApplication = ref(null); // For notes display

// Undo state
const lastAction = ref(null);
const UNDO_TIMEOUT = 10000; // 10 seconds

// Add to script setup section, after the existing state declarations:
const sliderHandle = ref(null);
let isDragging = false;
let startY = 0;
let startTop = 0;
const lastCopiedUsername = ref('');

// Search state
const searchQuery = ref('');
const searchFields = ref({
  discordUsername: true,
  minecraftUsername: true,
  favoriteAboutMinecraft: false,
  understandingOfSMP: false,
  id: false
});
const showSearchPanel = ref(false);
const searchInputRef = ref(null);

// Pagination state
const currentPage = ref(1);
const pageSize = ref(50);
const totalApplications = ref(0);
const totalPages = ref(0);

function fuzzyIncludes(haystack, needle) {
  if (!needle) return true;
  const h = String(haystack ?? '').toLowerCase();
  const n = String(needle).toLowerCase();
  let i = 0;
  for (let c of n) {
    i = h.indexOf(c, i);
    if (i === -1) return false;
    i++;
  }
  return true;
}

// Filtered applications - now handled server-side, just return applications
const filteredApplications = computed(() => {
  return applications.value;
});

function openSearchPanel() {
  showSearchPanel.value = true;
  nextTick(() => {
    if (searchInputRef.value) searchInputRef.value.focus();
  });
}

function closeSearchPanel() {
  showSearchPanel.value = false;
}

function handleGlobalKeydown(e) {
  const isCmd = e.metaKey || e.ctrlKey;
  if (isCmd && (e.key === 'f' || e.key === 'F')) {
    e.preventDefault();
    openSearchPanel();
  } else if (e.key === 'Escape' && showSearchPanel.value) {
    e.preventDefault();
    closeSearchPanel();
  }
}

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
  window.removeEventListener('keydown', handleGlobalKeydown);
  detachScrollListener();
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
            
            // Load data and set up tooltips
            refreshData();
            setupTooltipListeners();
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

// Show notes display modal
function showNotes(application) {
  selectedApplication.value = application;
  showNotesModal.value = true;
}

// Quick review (normal click)
async function quickReview(app, status, event) {
  if (event.shiftKey) return; // Let shift handler take over

  isProcessing.value = app.id;
  const previousStatus = app.isReviewed ? app.acceptanceStatus : 'unreviewed';

  try {
    const response = await api.reviewApplication(
      app.id,
      authenticatedPassword.value,
      '', // No notes
      status
    );

    if (response.success) {
      // Store undo action
      lastAction.value = {
        appId: app.id,
        previousStatus,
        newStatus: status,
        timestamp: Date.now()
      };

      // Show success notification with undo
      notification.success(
        `Application marked as ${status}`,
        () => handleUndo()
      );

      await refreshData();
    }
  } catch (error) {
    notification.error(`Failed to mark as ${status}`);
  } finally {
    isProcessing.value = null;
  }
}

// Open notes modal (shift+click)
function openNotesModal(app, status) {
  notesModalApp.value = app;
  notesModalStatus.value = status;
  notesModalNotes.value = app.reviewNotes || '';
  showNotesModal.value = true;
}

// Save from notes modal
async function handleNotesSave(notes, status) {
  if (!notesModalApp.value) return;

  isProcessing.value = notesModalApp.value.id;
  const previousStatus = notesModalApp.value.isReviewed
    ? notesModalApp.value.acceptanceStatus
    : 'unreviewed';

  try {
    const response = await api.reviewApplication(
      notesModalApp.value.id,
      authenticatedPassword.value,
      notes,
      status
    );

    if (response.success) {
      lastAction.value = {
        appId: notesModalApp.value.id,
        previousStatus,
        newStatus: status,
        timestamp: Date.now()
      };

      notification.success(
        `Application marked as ${status}`,
        () => handleUndo()
      );

      showNotesModal.value = false;
      await refreshData();
    }
  } catch (error) {
    notification.error('Failed to save review');
  } finally {
    isProcessing.value = null;
  }
}

// Undo last action
async function handleUndo() {
  if (!lastAction.value) return;

  const { appId, previousStatus } = lastAction.value;
  isProcessing.value = appId;

  try {
    if (previousStatus === 'unreviewed') {
      // Unreview application
      const response = await api.unreviewApplication(
        appId,
        authenticatedPassword.value
      );
      if (response.success) {
        notification.success('Action undone - application unreviewed');
      }
    } else {
      // Revert to previous status
      const app = applications.value.find(a => a.id === appId);
      const response = await api.reviewApplication(
        appId,
        authenticatedPassword.value,
        app?.reviewNotes || '',
        previousStatus
      );
      if (response.success) {
        notification.success(`Action undone - reverted to ${previousStatus}`);
      }
    }

    lastAction.value = null;
    await refreshData();
  } catch (error) {
    notification.error('Failed to undo action');
  } finally {
    isProcessing.value = null;
  }
}

// Global Ctrl+Z listener
function setupUndoListener() {
  window.addEventListener('keydown', (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'z') {
      e.preventDefault();
      if (lastAction.value &&
          Date.now() - lastAction.value.timestamp < UNDO_TIMEOUT) {
        handleUndo();
      }
    }
  });
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
      notification.success('Application unreviewed');
      await refreshData();
    } else {
      errorMessage.value = 'Failed to unreview application. Please try again.';
    }
  } catch (error) {
    notification.error('Failed to unreview application');
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
      notification.success('Application deleted');
      await refreshData();
    } else {
      errorMessage.value = 'Failed to delete application. Please try again.';
    }
  } catch (error) {
    notification.error('Failed to delete application');
  } finally {
    isDeleting.value = null;
  }
}

// Global tooltip element
let globalTooltip = null;

// Function to initialize global tooltip
function initializeGlobalTooltip() {
  // Remove any existing tooltip first
  const existingTooltip = document.getElementById('global-tooltip');
  if (existingTooltip) {
    existingTooltip.remove();
  }
  
  // Create the tooltip element
  globalTooltip = document.createElement('div');
  globalTooltip.id = 'global-tooltip';
  globalTooltip.style.position = 'fixed';
  globalTooltip.style.backgroundColor = 'rgba(0, 0, 0, 0.95)';
  globalTooltip.style.color = 'white';
  globalTooltip.style.padding = '0.5rem 0.75rem';
  globalTooltip.style.borderRadius = '0.25rem';
  globalTooltip.style.border = '1px solid #888';
  globalTooltip.style.boxShadow = '0 4px 12px rgba(0, 0, 0, 0.7), 0 0 0 1px rgba(255, 255, 255, 0.1)';
  globalTooltip.style.zIndex = '99999';
  globalTooltip.style.maxWidth = '300px';
  globalTooltip.style.fontSize = '0.875rem';
  globalTooltip.style.lineHeight = '1.4';
  globalTooltip.style.wordBreak = 'break-word';
  globalTooltip.style.pointerEvents = 'none';
  globalTooltip.style.opacity = '0';
  globalTooltip.style.visibility = 'hidden';
  globalTooltip.style.transition = 'opacity 0.15s ease';
  
  // Append to root element or body
  const tooltipRoot = document.getElementById('tooltip-root');
  if (tooltipRoot) {
    tooltipRoot.appendChild(globalTooltip);
  } else {
    document.body.appendChild(globalTooltip);
  }
}

// Function to show tooltip
function showTooltip(event) {
  if (!globalTooltip) return;
  
  // Get tooltip text from data attribute
  const tooltipText = event.currentTarget.getAttribute('data-tooltip');
  
  if (!tooltipText || tooltipText.trim() === '') return;
  
  // Set content
  globalTooltip.textContent = tooltipText;
  
  // Position tooltip first to get correct dimensions
  positionTooltip(event);
  
  // Make visible
  globalTooltip.style.visibility = 'visible';
  globalTooltip.style.opacity = '1';
}

// Function to hide tooltip
function hideTooltip() {
  if (!globalTooltip) return;
  
  globalTooltip.style.opacity = '0';
  globalTooltip.style.visibility = 'hidden';
}

// Function to position tooltip
function positionTooltip(event) {
  if (!globalTooltip) return;
  
  const rect = event.currentTarget.getBoundingClientRect();
  const tooltipHeight = globalTooltip.offsetHeight;
  const tooltipWidth = globalTooltip.offsetWidth;
  
  // Default position - above element
  let top = rect.top - tooltipHeight - 8;
  
  // If not enough room above, position below
  if (top < 8) {
    top = rect.bottom + 8;
  }
  
  // Center horizontally
  let left = rect.left + (rect.width / 2) - (tooltipWidth / 2);
  
  // Keep tooltip within viewport
  left = Math.max(8, Math.min(left, window.innerWidth - tooltipWidth - 8));
  
  // Position the tooltip - note we're using fixed positioning so no need to add scrollY
  globalTooltip.style.top = `${top}px`;
  globalTooltip.style.left = `${left}px`;
}

// Setup tooltip listeners
function setupTooltipListeners() {
  // Initialize the global tooltip
  nextTick(() => {
    initializeGlobalTooltip();

    // Find all tooltip containers
    const tooltipContainers = document.querySelectorAll('.tooltip-container');

    // Remove existing content and add data attributes
    tooltipContainers.forEach(container => {
      const tooltipElement = container.querySelector('.tooltip');
      if (tooltipElement) {
        // Store tooltip content as data attribute
        container.setAttribute('data-tooltip', tooltipElement.textContent);
        // Remove the tooltip element
        tooltipElement.remove();
      }
    });

    // Find table container for event delegation
    const tableContainer = document.querySelector('.overflow-auto');
    if (!tableContainer) return;

    // Remove existing event listeners (if any)
    tableContainer.removeEventListener('mouseover', handleMouseOver);
    tableContainer.removeEventListener('mouseout', handleMouseOut);
    tableContainer.removeEventListener('mousemove', handleMouseMove);
    tableContainer.removeEventListener('scroll', hideTooltip);

    // Add delegated event listeners
    tableContainer.addEventListener('mouseover', handleMouseOver);
    tableContainer.addEventListener('mouseout', handleMouseOut);
    tableContainer.addEventListener('mousemove', handleMouseMove);
    tableContainer.addEventListener('scroll', hideTooltip);
  });
}

let currentTooltipContainer = null;

function handleMouseOver(event) {
  const container = findTooltipContainer(event.target);
  if (!container) {
    hideTooltip();
    currentTooltipContainer = null;
    return;
  }
  if (container === currentTooltipContainer) return;
  currentTooltipContainer = container;
  // Create a synthetic event with currentTarget set to container
  const syntheticEvent = { currentTarget: container };
  showTooltip(syntheticEvent);
}

function handleMouseOut(event) {
  const relatedTarget = event.relatedTarget;
  const newContainer = findTooltipContainer(relatedTarget);
  // If moving to an element that is not inside the current container, hide tooltip
  if (newContainer !== currentTooltipContainer) {
    hideTooltip();
    currentTooltipContainer = null;
  }
}

function handleMouseMove(event) {
  if (!currentTooltipContainer) return;
  const container = findTooltipContainer(event.target);
  if (container !== currentTooltipContainer) return;
  const syntheticEvent = { currentTarget: container };
  positionTooltip(syntheticEvent);
}

function findTooltipContainer(element) {
  if (!element) return null;
  return element.closest('.tooltip-container');
}

// Update refreshData to hide tooltips
function hideAllTooltips() {
  if (globalTooltip) {
    globalTooltip.style.opacity = '0';
    globalTooltip.style.visibility = 'hidden';
  }
}

// Load initial data or refresh (replaces existing data)
async function refreshData() {
  isLoading.value = true;
  errorMessage.value = '';

  // Reset to page 1
  currentPage.value = 1;

  try {
    // Get selected search fields
    const selectedFields = Object.keys(searchFields.value).filter(key => searchFields.value[key]);

    // Call API with pagination and search params
    const response = await api.getApplications(
      authenticatedPassword.value,
      searchQuery.value,
      selectedFields,
      currentPage.value,
      pageSize.value
    );

    if (response.success) {
      const oldLength = applications.value.length;
      // Replace applications with new data
      applications.value = response.data || [];

      // Update pagination metadata
      totalApplications.value = response.total;
      totalPages.value = response.totalPages;
      currentPage.value = response.page;
      pageSize.value = response.pageSize;

      if (oldLength === 0 && applications.value.length > 0) {
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

          // Set up tooltip listeners after data is rendered
          setupTooltipListeners();
        });
      } else {
        // Scroll to top when refreshing
        nextTick(() => {
          const container = document.querySelector('.overflow-auto');
          if (container) {
            container.scrollTo({ top: 0, behavior: 'smooth' });
          }
          setupTooltipListeners();
        });
      }

      // Attach scroll listener for infinite scroll
      attachScrollListener();
    } else {
      errorMessage.value = 'Failed to load applications. Please try again.';
    }
  } catch (error) {
    errorMessage.value = 'Failed to load applications. Please try again.';
  } finally {
    isLoading.value = false;
  }
}

// Load more applications for infinite scroll (appends to existing data)
async function loadMoreApplications() {
  // Don't load if already loading or if we've loaded all
  if (isLoadingMore.value || currentPage.value >= totalPages.value) {
    return;
  }

  isLoadingMore.value = true;

  try {
    const selectedFields = Object.keys(searchFields.value).filter(key => searchFields.value[key]);
    const nextPage = currentPage.value + 1;

    const response = await api.getApplications(
      authenticatedPassword.value,
      searchQuery.value,
      selectedFields,
      nextPage,
      pageSize.value
    );

    if (response.success) {
      // Append new applications to existing list
      applications.value = [...applications.value, ...(response.data || [])];

      // Update pagination metadata
      totalApplications.value = response.total;
      totalPages.value = response.totalPages;
      currentPage.value = response.page;
      pageSize.value = response.pageSize;

      // Set up tooltip listeners after data is rendered
      nextTick(() => {
        setupTooltipListeners();
      });
    }
  } catch (error) {
    console.error('Error loading more applications:', error);
  } finally {
    isLoadingMore.value = false;
  }
}

// Infinite scroll handler
function handleScroll(event) {
  const container = event.target;
  const scrollTop = container.scrollTop;
  const scrollHeight = container.scrollHeight;
  const clientHeight = container.clientHeight;

  // Load more when user is within 300px of bottom
  const threshold = 300;
  if (scrollHeight - scrollTop - clientHeight < threshold) {
    loadMoreApplications();
  }
}

// Attach scroll listener
let scrollContainer = null;
function attachScrollListener() {
  nextTick(() => {
    scrollContainer = document.querySelector('.overflow-auto');
    if (scrollContainer) {
      scrollContainer.addEventListener('scroll', handleScroll);
    }
  });
}

// Detach scroll listener
function detachScrollListener() {
  if (scrollContainer) {
    scrollContainer.removeEventListener('scroll', handleScroll);
  }
}

// Debounced search
let searchDebounceTimeout = null;
function triggerSearch() {
  // Clear existing timeout
  if (searchDebounceTimeout) {
    clearTimeout(searchDebounceTimeout);
  }

  // Set new timeout for 300ms - refreshData will reset to page 1
  searchDebounceTimeout = setTimeout(() => {
    refreshData();
  }, 300);
}

// Watch for search query and field changes
watch([searchQuery, searchFields], () => {
  triggerSearch();
}, { deep: true });

// Export to CSV
async function exportToCsv() {
  try {
    await api.exportApplications(authenticatedPassword.value);
    notification.success('CSV export started');
  } catch (error) {
    notification.error('Failed to export applications');
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

// Add to mounted hook to set up event listeners
onMounted(() => {
  // Focus password input when mounted
  if (passwordInput.value) {
    passwordInput.value.focus();
  }

  // Set up tooltip event listeners
  setupTooltipListeners();
  // Global search hotkey
  window.addEventListener('keydown', handleGlobalKeydown);
  // Undo listener
  setupUndoListener();
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
  refreshData();
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

/* New tooltip implementation */
.tooltip-container {
  position: relative;
  cursor: default;
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

/* Firefox-friendly corner rounding fallback on first/last rows */
tbody tr:first-child td:first-child { border-top-left-radius: 0.5rem; }
tbody tr:first-child td:last-child { border-top-right-radius: 0.5rem; }
tbody tr:last-child td:first-child { border-bottom-left-radius: 0.5rem; }
tbody tr:last-child td:last-child { border-bottom-right-radius: 0.5rem; }

/* Slightly lighter backgrounds for better contrast between rows */
tbody tr:nth-child(even) {
  background-color: rgba(255, 255, 255, 0.03) !important;
}

tbody tr:hover {
  background-color: rgba(96, 165, 250, 0.12) !important;
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
  box-shadow: 0 0 8px rgba(50, 205, 50, 0.35);
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

/* Slide-up transition for floating search */
.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(16px);
  opacity: 0;
}
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 150ms ease;
}
</style>
