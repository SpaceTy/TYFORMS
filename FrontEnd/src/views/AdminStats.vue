<template>
  <div class="h-full w-full absolute inset-0 flex items-center justify-center">
    <div v-if="!isAuthenticated" class="login-container mc-container animate-fade-in max-w-md">
      <h2 class="mc-title text-center">Admin Statistics</h2>

      <form @submit.prevent="authenticate" class="space-y-6">
        <div class="form-group">
          <label for="username" class="mc-label">Username</label>
          <input
            id="username"
            v-model="username"
            type="text"
            class="mc-input"
            placeholder="Username (leave empty for admin password)"
            autocomplete="username"
          />
        </div>

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
            autocomplete="current-password"
          />
        </div>

        <div v-if="authError" class="bg-red-600/40 text-white p-3 rounded-lg border border-red-500/30">
          {{ authError }}
        </div>

        <div class="text-center">
          <button type="submit" class="mc-button">Login</button>
        </div>
      </form>
    </div>

    <div v-else class="absolute inset-0 p-4">
      <div class="admin-panel mc-panel w-full h-full flex flex-col">
        <div class="flex justify-between items-center px-6 py-4 bg-black/70 sticky top-0 z-20 border-b border-white/10">
          <h2 class="mc-title mb-0">Application Statistics</h2>
          <div class="flex gap-2">
            <button @click="goToDashboard" class="mc-button text-sm secondary">Dashboard</button>
            <button @click="loadStats" class="mc-button text-sm" :disabled="isLoading">
              <span v-if="isLoading">Loading...</span>
              <span v-else>Refresh</span>
            </button>
            <button @click="logout" class="mc-button text-sm danger">Logout</button>
          </div>
        </div>

        <div class="flex-grow overflow-auto p-6 bg-black/80">
          <div v-if="errorMessage" class="bg-red-600/40 text-white p-3 mb-4 rounded-lg border border-red-500/30">
            {{ errorMessage }}
          </div>

          <div v-if="isLoading && !stats" class="text-neutral-300">Loading statistics...</div>

          <template v-else-if="stats">
            <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4 mb-6">
              <div class="glass p-4 rounded-lg border border-white/10">
                <p class="text-xs uppercase text-neutral-400">Total Applications</p>
                <p class="text-3xl font-bold text-white">{{ stats.totalApplications }}</p>
              </div>
              <div class="glass p-4 rounded-lg border border-white/10">
                <p class="text-xs uppercase text-neutral-400">Reviewed</p>
                <p class="text-3xl font-bold text-green-400">{{ stats.reviewedApplications }}</p>
              </div>
              <div class="glass p-4 rounded-lg border border-white/10">
                <p class="text-xs uppercase text-neutral-400">Unreviewed</p>
                <p class="text-3xl font-bold text-yellow-400">{{ stats.unreviewedApplications }}</p>
              </div>
              <div class="glass p-4 rounded-lg border border-white/10">
                <p class="text-xs uppercase text-neutral-400">Average Age</p>
                <p class="text-3xl font-bold text-primary-300">{{ stats.averageAge.toFixed(1) }}</p>
              </div>
            </div>

            <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div class="glass p-4 rounded-lg border border-white/10">
                <h3 class="text-lg font-semibold text-white mb-4">Review Outcome</h3>
                <div class="space-y-3">
                  <div>
                    <div class="flex justify-between text-sm">
                      <span class="text-green-300">Accepted</span>
                      <span>{{ stats.acceptedApplications }}</span>
                    </div>
                    <div class="h-2 bg-white/10 rounded mt-1 overflow-hidden">
                      <div class="h-full bg-green-500" :style="{ width: `${percentage(stats.acceptedApplications)}%` }"></div>
                    </div>
                  </div>
                  <div>
                    <div class="flex justify-between text-sm">
                      <span class="text-yellow-300">Pending</span>
                      <span>{{ stats.pendingApplications }}</span>
                    </div>
                    <div class="h-2 bg-white/10 rounded mt-1 overflow-hidden">
                      <div class="h-full bg-yellow-500" :style="{ width: `${percentage(stats.pendingApplications)}%` }"></div>
                    </div>
                  </div>
                  <div>
                    <div class="flex justify-between text-sm">
                      <span class="text-red-300">Rejected</span>
                      <span>{{ stats.rejectedApplications }}</span>
                    </div>
                    <div class="h-2 bg-white/10 rounded mt-1 overflow-hidden">
                      <div class="h-full bg-red-500" :style="{ width: `${percentage(stats.rejectedApplications)}%` }"></div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="glass p-4 rounded-lg border border-white/10">
                <h3 class="text-lg font-semibold text-white mb-4">Last 7 Days</h3>
                <div class="space-y-2">
                  <div
                    v-for="item in stats.recentSubmissions"
                    :key="item.date"
                    class="grid grid-cols-[70px_1fr_40px] gap-3 items-center text-sm"
                  >
                    <span class="text-neutral-300">{{ formatShortDate(item.date) }}</span>
                    <div class="h-2 bg-white/10 rounded overflow-hidden">
                      <div class="h-full bg-primary-500" :style="{ width: `${recentPercentage(item.count)}%` }"></div>
                    </div>
                    <span class="text-right text-neutral-200">{{ item.count }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="glass p-4 rounded-lg border border-white/10 mt-6">
              <p class="text-sm text-neutral-300">
                Joined Discord:
                <span class="text-white font-semibold">{{ stats.joinedDiscordCount }}</span>
                / {{ stats.totalApplications }}
              </p>
            </div>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import api from '../services/api';

const router = useRouter();
const isAuthenticated = ref(false);
const username = ref(api.getUsername());
const password = ref('');
const authError = ref('');
const errorMessage = ref('');
const isLoading = ref(false);
const stats = ref(null);
const passwordInput = ref(null);

const maxRecentCount = computed(() => {
  if (!stats.value?.recentSubmissions?.length) {
    return 0;
  }
  return Math.max(...stats.value.recentSubmissions.map((item) => item.count));
});

function percentage(value) {
  if (!stats.value?.totalApplications) {
    return 0;
  }
  return Math.round((value / stats.value.totalApplications) * 100);
}

function recentPercentage(value) {
  if (!maxRecentCount.value) {
    return 0;
  }
  return Math.round((value / maxRecentCount.value) * 100);
}

function formatShortDate(dateString) {
  const date = new Date(`${dateString}T00:00:00Z`);
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
}

async function authenticate() {
  try {
    const response = await api.login(username.value.trim(), password.value);
    if (!response.success) {
      authError.value = 'Invalid username or password';
      return;
    }

    isAuthenticated.value = true;
    authError.value = '';
    password.value = '';
    await loadStats();
  } catch (error) {
    if (error?.response?.status === 401) {
      authError.value = 'Invalid username or password';
    } else {
      authError.value = 'Authentication error. Please try again.';
    }
  }
}

async function loadStats() {
  if (!isAuthenticated.value) return;

  isLoading.value = true;
  errorMessage.value = '';

  try {
    const response = await api.getApplicationStats();
    if (response?.success) {
      stats.value = response.data;
      return;
    }
    errorMessage.value = 'Failed to load statistics.';
  } catch (error) {
    if (error?.response?.status === 401) {
      isAuthenticated.value = false;
      localStorage.removeItem('adminToken');
      errorMessage.value = '';
      return;
    }
    errorMessage.value = 'Failed to load statistics.';
  } finally {
    isLoading.value = false;
  }
}

function goToDashboard() {
  router.push('/admin');
}

function logout() {
  isAuthenticated.value = false;
  password.value = '';
  stats.value = null;
  api.logout();
  router.push('/admin');
}

async function tryRestoreSession() {
  const result = await api.validateToken();
  if (result.valid) {
    isAuthenticated.value = true;
    loadStats();
    return true;
  }
  return false;
}

onMounted(async () => {
  const restored = await tryRestoreSession();
  if (!restored) {
    nextTick(() => {
      if (passwordInput.value) {
        passwordInput.value.focus();
      }
    });
  }
});
</script>
