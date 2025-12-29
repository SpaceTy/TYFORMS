<template>
  <div class="h-full w-full absolute inset-0 bg-black text-white overflow-hidden" @mousemove.stop>
    <!-- Login Form -->
    <div v-if="!isAuthenticated" class="h-full w-full flex flex-col items-center justify-center p-6 login-container z-10 relative">
      <div class="mc-container w-full max-w-sm bg-black/60 border border-white/10 shadow-2xl p-6 rounded-2xl animate-fade-in">
        <h2 class="mc-title text-center text-2xl mb-6">Admin Access</h2>
        
        <form @submit.prevent="authenticate" class="space-y-5">
          <div class="space-y-2">
            <label for="password" class="text-sm font-medium text-gray-300">Password</label>
            <input 
              id="password" 
              v-model="password" 
              type="password" 
              class="w-full bg-black/70 border border-white/20 rounded-lg px-4 py-3 text-white focus:ring-2 focus:ring-primary-500 focus:border-transparent outline-none transition-all placeholder-gray-500"
              placeholder="Enter admin password"
              required
              ref="passwordInput"
            />
          </div>
          
          <div v-if="authError" class="bg-red-500/20 border border-red-500/50 text-red-200 text-sm p-3 rounded-lg flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 flex-shrink-0" viewBox="0 0 20 20" fill="currentColor">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
            </svg>
            {{ authError }}
          </div>
          
          <button type="submit" class="w-full bg-primary-600 hover:bg-primary-500 text-white font-bold py-3 px-4 rounded-lg shadow-lg transform transition active:scale-95 duration-200">
            Login
          </button>
        </form>
      </div>
    </div>
    
    <!-- Admin Dashboard -->
    <div v-else class="h-full flex flex-col relative z-10" ref="adminContainerRef">
      <!-- Header -->
      <header class="bg-black/90 border-b border-white/10 px-4 py-3 flex items-center justify-between shrink-0 z-30">
        <div class="flex items-center gap-2">
          <h2 class="font-pixel text-xl text-primary-400 m-0 leading-none">Dashboard</h2>
          <span class="text-xs bg-black/30 px-2 py-0.5 rounded text-gray-400 font-mono">{{ totalApplications }}</span>
        </div>
        
        <div class="flex items-center gap-2">
          <button @click="toggleSearch" class="p-2 rounded-full hover:bg-white/10 text-gray-300 active:bg-white/20 transition-colors">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </button>
          
          <div class="relative">
            <button @click="toggleMenu" class="p-2 rounded-full hover:bg-white/10 text-gray-300 active:bg-white/20 transition-colors">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
              </svg>
            </button>
            
            <!-- Dropdown Menu -->
            <div v-if="showMenu" class="absolute right-0 top-full mt-2 w-48 bg-black/90 border border-white/10 rounded-xl shadow-xl overflow-hidden z-50 animate-scale-in origin-top-right">
              <div class="py-1">
                <button @click="refreshData(); showMenu = false" class="w-full text-left px-4 py-3 hover:bg-white/5 flex items-center gap-2 text-sm">
                  <span v-if="isLoading" class="animate-spin">⟳</span>
                  <span v-else>↻</span>
                  Refresh Data
                </button>
                <button @click="exportToCsv(); showMenu = false" class="w-full text-left px-4 py-3 hover:bg-white/5 flex items-center gap-2 text-sm">
                  <span>⬇</span> Export CSV
                </button>
                <div class="h-px bg-black/30 my-1"></div>
                <button @click="logout" class="w-full text-left px-4 py-3 hover:bg-red-500/20 text-red-400 flex items-center gap-2 text-sm">
                  <span>➜</span> Logout
                </button>
              </div>
            </div>
            <!-- Backdrop for menu -->
            <div v-if="showMenu" @click="showMenu = false" class="fixed inset-0 z-40 bg-transparent"></div>
          </div>
        </div>
      </header>
      
      <!-- Search Bar (Expandable) -->
      <div v-if="showSearch" class="bg-black/80 border-b border-white/10 px-4 py-3 animate-slide-down shrink-0">
        <div class="relative">
          <input 
            v-model="searchQuery" 
            ref="searchInput"
            class="w-full bg-black/60 border border-white/20 rounded-lg pl-9 pr-4 py-2 text-sm focus:outline-none focus:border-primary-500/50 focus:ring-1 focus:ring-primary-500/50 transition-all"
            placeholder="Search username, discord..." 
          />
          <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
          <button v-if="searchQuery" @click="searchQuery = ''" class="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-500 hover:text-white">
            ✕
          </button>
        </div>
        <!-- Search filters chips -->
        <div class="flex gap-2 mt-2 overflow-x-auto pb-1 hide-scrollbar">
           <label v-for="(enabled, field) in searchFields" :key="field" 
                  class="flex-shrink-0 text-xs px-2 py-1 rounded-full border cursor-pointer select-none transition-colors"
                  :class="enabled ? 'bg-primary-500/20 border-primary-500/40 text-primary-200' : 'bg-black/20 border-white/10 text-gray-400'">
             <input type="checkbox" v-model="searchFields[field]" class="hidden">
             {{ formatFieldName(field) }}
           </label>
        </div>
      </div>
      
      <!-- Error Message -->
      <div v-if="errorMessage" class="bg-red-900/50 border-l-4 border-red-500 text-white p-3 mx-4 my-2 text-sm rounded shadow-lg shrink-0 animate-fade-in">
        <div class="flex justify-between items-start">
          <span>{{ errorMessage }}</span>
          <button @click="errorMessage = ''" class="text-white/60 hover:text-white">×</button>
        </div>
      </div>
      
      <!-- List Content -->
      <div class="flex-grow overflow-y-auto overflow-x-hidden p-4 space-y-3" @scroll="handleScroll">
        <!-- Loading Initial -->
        <div v-if="isLoading && applications.length === 0" class="flex flex-col items-center justify-center p-12 text-gray-400">
           <div class="animate-spin h-8 w-8 border-2 border-primary-500 border-t-transparent rounded-full mb-3"></div>
           <span class="text-sm">Loading applications...</span>
        </div>
        
        <!-- Empty State -->
        <div v-if="applications.length === 0 && !isLoading" class="flex flex-col items-center justify-center h-64 text-center text-gray-400">
          <div class="bg-black/20 p-4 rounded-full mb-3">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <p class="font-medium">No applications found</p>
          <p class="text-xs mt-1 opacity-60">Try adjusting your search filters</p>
        </div>
        
        <!-- Cards -->
        <div 
          v-for="app in filteredApplications" 
          :key="app.id"
          class="application-card bg-black/60 border border-white/5 rounded-xl p-4 shadow-sm active:bg-black/80 transition-colors"
          :class="{
            'border-l-4 border-l-yellow-500': !app.isReviewed,
            'border-l-4 border-l-green-500': app.isReviewed && app.acceptanceStatus === 'accepted',
            'border-l-4 border-l-red-500': app.isReviewed && app.acceptanceStatus === 'rejected',
            'border-l-4 border-l-orange-500': app.isReviewed && app.acceptanceStatus === 'pending',
          }"
        >
          <!-- Card Header: ID & Status -->
          <div class="flex justify-between items-start mb-2">
            <div class="flex items-center gap-2">
              <span class="text-xs font-mono text-gray-400">#{{ app.id }}</span>
              <span v-if="!app.isReviewed" class="text-[10px] bg-yellow-500/20 text-yellow-300 px-1.5 py-0.5 rounded uppercase tracking-wider font-bold">New</span>
              <span v-else class="text-[10px] px-1.5 py-0.5 rounded uppercase tracking-wider font-bold"
                :class="{
                  'bg-green-500/20 text-green-300': app.acceptanceStatus === 'accepted',
                  'bg-red-500/20 text-red-300': app.acceptanceStatus === 'rejected',
                  'bg-orange-500/20 text-orange-300': app.acceptanceStatus === 'pending'
                }"
              >
                {{ app.acceptanceStatus }}
              </span>
            </div>
            <span class="text-xs text-gray-500">{{ formatDate(app.submissionDate) }}</span>
          </div>
          
          <!-- Players Details -->
          <div class="flex items-center gap-3 mb-3">
             <div class="h-10 w-10 rounded bg-black/60 flex items-center justify-center shrink-0">
               <!-- Minecraft Head Placeholder / Initials -->
               <span class="font-pixel text-lg">{{ app.minecraftUsername.charAt(0).toUpperCase() }}</span>
             </div>
             <div class="overflow-hidden">
               <div class="flex items-center gap-1.5" @click="copyToClipboard(app.minecraftUsername)">
                 <h3 class="font-bold text-white truncate">{{ app.minecraftUsername }}</h3>
                 <svg xmlns="http://www.w3.org/2000/svg" class="h-3 w-3 text-gray-500" viewBox="0 0 20 20" fill="currentColor">
                   <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
                   <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
                 </svg>
                 <!-- Copied tooltip could go here -->
               </div>
               <div class="flex items-center gap-2 text-xs text-gray-400">
                 <span class="flex items-center gap-1">
                   <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24"><path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0 12.64 12.64 0 0 0-.617-1.25.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057 19.9 19.9 0 0 0 5.993 3.03.078.078 0 0 0 .084-.028 14.09 14.09 0 0 0 1.226-1.994.076.076 0 0 0-.041-.106 13.107 13.107 0 0 1-1.872-.892.077.077 0 0 0-.008-.128 10.2 10.2 0 0 0 .372-.292.074.074 0 0 0 .032-.055c3.962 1.82 8.24 1.82 12.16 0a.074.074 0 0 0 .032.055 10.5 10.5 0 0 0 .372.292.077.077 0 0 0-.008.128 12.532 12.532 0 0 1-1.872.892.077.077 0 0 0-.041.106c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028 19.9 19.9 0 0 0 6.002-3.03.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03z"/></svg>
                   {{ app.discordUsername }}
                 </span>
                 <span>•</span>
                 <span :class="getAgeColor(app.age)">{{ app.age }} y/o</span>
               </div>
             </div>
          </div>
          
          <!-- Q&A Preview -->
          <div class="space-y-2 mb-4 bg-black/40 p-2 rounded text-xs text-gray-300">
            <div>
              <span class="text-primary-400 font-bold uppercase text-[10px]">Favorite Thing</span>
              <p class="line-clamp-2 leading-relaxed">{{ app.favoriteAboutMinecraft }}</p>
            </div>
            <div>
              <span class="text-primary-400 font-bold uppercase text-[10px]">SMP Understanding</span>
              <p class="line-clamp-2 leading-relaxed">{{ app.understandingOfSMP }}</p>
            </div>
          </div>
          
          <!-- Actions -->
          <div class="grid grid-cols-2 gap-2">
            <button 
              @click="openReviewModal(app)" 
              class="flex items-center justify-center gap-1.5 py-2 px-3 rounded-lg text-xs font-bold transition-all"
              :class="app.isReviewed ? 'bg-black/70 text-white hover:bg-black/60' : 'bg-green-600 text-white hover:bg-green-500 shadow-lg shadow-green-900/20'"
            >
              <span v-if="app.isReviewed">Edit Review</span>
              <span v-else>Review App</span>
            </button>
            
            <button 
              @click="confirmDelete(app.id)"
              class="flex items-center justify-center gap-1.5 py-2 px-3 rounded-lg bg-red-500/10 text-red-400 border border-red-500/20 hover:bg-red-500/20 text-xs font-bold transition-all"
            >
              Delete
            </button>
          </div>
        </div>
        
        <!-- Loading More -->
        <div v-if="isLoadingMore" class="flex justify-center py-4">
          <div class="h-6 w-6 border-2 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
        </div>
        
         <!-- End of Results -->
        <div v-if="applications.length > 0 && applications.length >= totalApplications" class="text-center py-4 text-xs text-gray-600">
          All applications loaded
        </div>
      </div>
    </div>

    <!-- Review Modal -->
    <Transition name="fade">
      <div v-if="showReviewModal" class="fixed inset-0 z-[100] flex items-end sm:items-center justify-center" role="dialog">
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/90" @click="showReviewModal = false"></div>
        
        <!-- Modal Content -->
        <div class="relative w-full sm:w-[500px] bg-[#0a0a0a] sm:rounded-xl rounded-t-xl border-t sm:border border-white/10 shadow-2xl flex flex-col max-h-[90vh] animate-slide-up">
           <!-- Modal Header -->
           <div class="px-5 py-4 border-b border-white/5 flex justify-between items-center shrink-0">
             <h3 class="text-lg font-bold text-white">Review Application</h3>
             <button @click="showReviewModal = false" class="text-gray-400 hover:text-white p-1">✕</button>
           </div>
           
           <!-- Modal Body (Scrollable) -->
           <div class="p-5 overflow-y-auto custom-scrollbar space-y-5">
             <div v-if="selectedApplication">
               <div class="flex items-center gap-3 p-3 bg-black/50 rounded-lg border border-white/5">
                 <div class="h-10 w-10 bg-primary-900/50 rounded flex items-center justify-center text-primary-300 font-bold">
                   {{ selectedApplication.minecraftUsername.charAt(0).toUpperCase() }}
                 </div>
                 <div>
                   <div class="text-white font-bold">{{ selectedApplication.minecraftUsername }}</div>
                   <div class="text-xs text-gray-400">{{ selectedApplication.discordUsername }}</div>
                 </div>
               </div>
               
               <!-- Full Content View for Context -->
               <div class="space-y-3 mt-4">
                  <div>
                    <span class="text-xs uppercase text-gray-500 font-bold tracking-wider">Favorite Thing About MC</span>
                    <p class="text-sm text-gray-300 mt-1 p-3 bg-black/50 rounded-lg border border-white/5">{{ selectedApplication.favoriteAboutMinecraft }}</p>
                  </div>
                  <div>
                    <span class="text-xs uppercase text-gray-500 font-bold tracking-wider">Understanding of SMP</span>
                    <p class="text-sm text-gray-300 mt-1 p-3 bg-black/50 rounded-lg border border-white/5">{{ selectedApplication.understandingOfSMP }}</p>
                  </div>
               </div>

               <!-- Status Selection -->
               <div class="mt-4">
                 <label class="block text-xs uppercase text-gray-500 font-bold tracking-wider mb-2">Decision</label>
                 <div class="grid grid-cols-3 gap-2">
                    <button @click="setAcceptanceStatus('accepted')" 
                      class="py-2.5 rounded-lg border text-sm font-bold transition-all"
                      :class="acceptanceStatus === 'accepted' ? 'bg-green-600 border-green-500 text-white' : 'bg-transparent border-white/10 text-gray-400 hover:bg-white/5'">
                      Accept
                    </button>
                    <button @click="setAcceptanceStatus('pending')" 
                      class="py-2.5 rounded-lg border text-sm font-bold transition-all"
                      :class="acceptanceStatus === 'pending' ? 'bg-orange-600 border-orange-500 text-white' : 'bg-transparent border-white/10 text-gray-400 hover:bg-white/5'">
                      Pending
                    </button>
                    <button @click="setAcceptanceStatus('rejected')" 
                      class="py-2.5 rounded-lg border text-sm font-bold transition-all"
                      :class="acceptanceStatus === 'rejected' ? 'bg-red-600 border-red-500 text-white' : 'bg-transparent border-white/10 text-gray-400 hover:bg-white/5'">
                      Reject
                    </button>
                 </div>
               </div>

               <!-- Notes -->
               <div class="mt-4">
                 <label class="block text-xs uppercase text-gray-500 font-bold tracking-wider mb-2">Internal Notes</label>
                 <textarea 
                   v-model="reviewNotes"
                   rows="3" 
                   class="w-full bg-black/50 border border-white/10 rounded-lg p-3 text-sm text-white focus:outline-none focus:border-primary-500/50"
                   placeholder="Add notes privately..."
                 ></textarea>
               </div>

               <!-- Unreview Action -->
               <div v-if="selectedApplication.isReviewed" class="mt-4 pt-4 border-t border-white/5">
                 <button 
                   @click="unreviewApplication"
                   class="text-xs text-red-400 hover:text-red-300 flex items-center gap-1"
                   :disabled="isUnreviewing"
                 >
                   <span v-if="isUnreviewing">Processing...</span>
                   <span v-else>Reset Application Status (Unreview)</span>
                 </button>
               </div>
             </div>
           </div>
           
           <!-- Modal Footer -->
           <div class="p-4 border-t border-white/5 bg-black/70 shrink-0 flex gap-3">
             <button @click="showReviewModal = false" class="flex-1 py-3 rounded-lg text-gray-400 hover:bg-white/5 text-sm font-bold">Cancel</button>
             <button 
               @click="submitReview" 
               class="flex-[2] py-3 rounded-lg bg-white text-black hover:bg-gray-200 text-sm font-bold shadow-lg flex justify-center items-center"
               :disabled="isSubmittingReview"
             >
               <span v-if="isSubmittingReview" class="animate-spin mr-2 h-4 w-4 border-2 border-black border-t-transparent rounded-full"></span>
               {{ isSubmittingReview ? 'Saving...' : 'Save Decision' }}
             </button>
           </div>
        </div>
      </div>
    </Transition>

  </div>
</template>

<script setup>
import { ref, nextTick, inject, computed } from 'vue';
import { gsap } from 'gsap';
import api from '../services/api';

const confirmation = inject('confirmation');

// UI state
const isLoading = ref(true);
const isLoadingMore = ref(false);
const errorMessage = ref('');
const applications = ref([]);
const showMenu = ref(false);
const showSearch = ref(false);

// Pagination state
const currentPage = ref(1);
const pageSize = ref(50);
const totalApplications = ref(0);
const totalPages = ref(0);

// Authentication state
const isAuthenticated = ref(false);
const password = ref('');
const authError = ref('');
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
const isUnreviewing = ref(false);

// Search state
const searchQuery = ref('');
const searchInput = ref(null);
const searchFields = ref({
  discordUsername: true,
  minecraftUsername: true,
  favoriteAboutMinecraft: false,
  understandingOfSMP: false,
  id: false
});

// Format field names for display
function formatFieldName(field) {
  const map = {
    discordUsername: 'Discord',
    minecraftUsername: 'IGN',
    favoriteAboutMinecraft: 'Favorite',
    understandingOfSMP: 'Understanding',
    id: 'ID'
  };
  return map[field] || field;
}

const filteredApplications = computed(() => {
  return applications.value;
});

function toggleMenu() {
  showMenu.value = !showMenu.value;
}

function toggleSearch() {
  showSearch.value = !showSearch.value;
  if(showSearch.value) {
    nextTick(() => {
      searchInput.value?.focus();
    });
  } else {
    searchQuery.value = '';
  }
}

// Authenticate user
async function authenticate() {
  try {
    const response = await api.verifyAdminPassword(password.value);
    
    if (response.success) {
      authenticatedPassword.value = password.value;
      
      gsap.to('.login-container', {
        opacity: 0,
        y: -20,
        duration: 0.3,
        onComplete: () => {
          isAuthenticated.value = true;
          password.value = '';
          authError.value = '';
          
          nextTick(() => {
            refreshData();
          });
        }
      });
    } else {
      authError.value = 'Invalid password';
      gsap.fromTo('.login-container', { x: -5 }, { x: 5, duration: 0.1, repeat: 3, yoyo: true });
    }
  } catch (error) {
    authError.value = 'Authentication error. Please try again.';
  }
}

function getAgeColor(age) {
  if (age < 13) return 'text-red-400';
  if (age < 16) return 'text-yellow-400';
  if (age < 18) return 'text-green-400';
  return 'text-white';
}

function formatDate(dateString, includeTime = false) {
  if (!dateString) return '';
  const date = new Date(dateString);
  if (includeTime) {
    return date.toLocaleString();
  }
  return date.toLocaleDateString();
}

async function copyToClipboard(text) {
  try {
    await navigator.clipboard.writeText(text);
    // Could add a toast notification here
  } catch (err) {
    console.error('Failed to copy', err);
  }
}

function openReviewModal(application) {
  selectedApplication.value = application;
  reviewNotes.value = application.reviewNotes || '';
  acceptanceStatus.value = application.acceptanceStatus || 'pending';
  showReviewModal.value = true;
  showMenu.value = false; // close menu if open
}

function setAcceptanceStatus(status) {
  acceptanceStatus.value = status;
}

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
    
    if (response && response.success) {
      const index = applications.value.findIndex(app => app.id === selectedApplication.value.id);
      if (index !== -1) {
        // Optimistic update
        const updatedApp = { ...applications.value[index], ...response.data };
        // Ensure status logic if backend didn't return it perfectly immediately (safety)
        if (!updatedApp.acceptanceStatus) updatedApp.acceptanceStatus = acceptanceStatus.value;
        
        applications.value[index] = updatedApp;
      }
      showReviewModal.value = false;
    }
  } catch (error) {
    console.error('Error submitting review:', error);
    errorMessage.value = 'Failed to submit review.';
  } finally {
    isSubmittingReview.value = false;
    isProcessing.value = null;
  }
}

async function unreviewApplication() {
  if (!selectedApplication.value) return;
  
  const confirmed = await confirmation.confirm({
    title: 'Unreview Application',
    message: 'Are you sure you want to reset the review status? This will mark it as new.',
    confirmText: 'Reset',
    cancelText: 'Cancel'
  });
  
  if (!confirmed) return;

  isUnreviewing.value = true;
  try {
    const response = await api.unreviewApplication(
      selectedApplication.value.id,
      authenticatedPassword.value
    );
    
    if (response.success) {
       const index = applications.value.findIndex(app => app.id === selectedApplication.value.id);
       if (index !== -1) {
         applications.value[index] = response.data;
       }
       showReviewModal.value = false;
    }
  } catch(error) {
    console.error('Error unreviewing', error);
    errorMessage.value = 'Failed to unreview application';
  } finally {
    isUnreviewing.value = false;
  }
}

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

async function deleteApplication(applicationId) {
  isDeleting.value = applicationId;
  try {
    const response = await api.deleteApplication(applicationId, authenticatedPassword.value);
    if (response.success) {
      applications.value = applications.value.filter(app => app.id !== applicationId);
      totalApplications.value--; 
    }
  } catch (error) {
    errorMessage.value = 'Failed to delete application.';
  } finally {
    isDeleting.value = null;
  }
}

async function exportToCsv() {
  try {
    await api.exportApplications(authenticatedPassword.value);
    showMenu.value = false;
  } catch (error) {
    errorMessage.value = 'Failed to export applications.';
  }
}

async function refreshData() {
  isLoading.value = true;
  errorMessage.value = '';
  currentPage.value = 1;

  try {
    const selectedFields = Object.keys(searchFields.value).filter(key => searchFields.value[key]);
    const response = await api.getApplications(
      authenticatedPassword.value,
      searchQuery.value,
      selectedFields,
      currentPage.value,
      pageSize.value
    );

    if (response && response.success) {
      applications.value = response.data || [];
      totalApplications.value = response.total;
      totalPages.value = response.totalPages;
      currentPage.value = response.page;
      pageSize.value = response.pageSize;
    } else {
      applications.value = [];
    }
  } catch (error) {
    errorMessage.value = 'Failed to fetch applications.';
    applications.value = [];
  } finally {
    isLoading.value = false;
  }
}

async function loadMoreApplications() {
  if (isLoadingMore.value || currentPage.value >= totalPages.value) return;
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
      applications.value = [...applications.value, ...(response.data || [])];
      currentPage.value = response.page;
    }
  } catch (error) {
    console.error(error);
  } finally {
    isLoadingMore.value = false;
  }
}

function handleScroll(event) {
  const { scrollTop, scrollHeight, clientHeight } = event.target;
  if (scrollTop + clientHeight >= scrollHeight - 300) {
    loadMoreApplications();
  }
}

async function logout() {
  isAuthenticated.value = false;
  authenticatedPassword.value = '';
  password.value = '';
  applications.value = [];
}
</script>

<style scoped>
.hide-scrollbar::-webkit-scrollbar {
  display: none;
}
.hide-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

/* Animations */
.animate-fade-in {
  animation: fadeIn 0.4s ease-out;
}
.animate-scale-in {
  animation: scaleIn 0.2s ease-out;
}
.animate-slide-down {
  animation: slideDown 0.2s ease-out;
}
.animate-slide-up {
  animation: slideUp 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes scaleIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

@keyframes slideDown {
  from { opacity: 0; transform: translateY(-10px); }
  to { opacity: 1; transform: translateY(0); }
}

@keyframes slideUp {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>