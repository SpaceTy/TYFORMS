<template>
  <div class="h-full w-full absolute inset-0 flex items-center justify-center">
    <div class="admin-panel mc-panel w-full h-full flex flex-col">
      <!-- Fixed top bar -->
      <div class="flex justify-between items-center px-6 py-4 bg-black/70 sticky top-0 z-20 border-b border-white/10">
        <h2 class="mc-title mb-0">Admin Accounts</h2>

        <div class="flex gap-2 items-center">
          <button @click="goToDashboard" class="mc-button text-sm secondary">
            Dashboard
          </button>
          <ProfileMenu @logout="logout" />
        </div>
      </div>

      <!-- Scrollable content -->
      <div class="flex-grow overflow-auto p-6 bg-black/80">
        <div v-if="errorMessage" class="bg-red-600/40 text-white p-3 mb-4 rounded-lg border border-red-500/30">
          {{ errorMessage }}
        </div>

        <div v-if="notice" class="bg-green-600/30 text-white p-3 mb-4 rounded-lg border border-green-500/40">
          {{ notice }}
        </div>

        <div v-if="isLoading" class="text-neutral-300 py-10 text-center">Loading admin accounts...</div>

        <template v-else>
          <!-- Account list -->
          <div class="max-w-2xl mx-auto space-y-3 mb-8">
            <div
              v-for="admin in admins"
              :key="admin.id"
              class="glass p-4 rounded-xl border border-white/10 flex items-center justify-between gap-4"
            >
              <div class="flex items-center gap-3 min-w-0">
                <span class="w-10 h-10 rounded-full bg-primary-600/80 border border-primary-300/40 flex items-center justify-center text-white text-lg font-bold uppercase shrink-0">
                  {{ admin.username.charAt(0) }}
                </span>
                <div class="min-w-0">
                  <div class="flex items-center gap-2 flex-wrap">
                    <p class="text-white font-semibold truncate">{{ admin.username }}</p>
                    <span
                      v-if="admin.username === SEED_USERNAME"
                      class="text-[10px] uppercase tracking-wide bg-minecraft-gold/20 text-minecraft-gold border border-minecraft-gold/40 px-1.5 py-0.5 rounded"
                    >
                      Seed
                    </span>
                    <span
                      v-if="currentAdmin && admin.id === currentAdmin.id"
                      class="text-[10px] uppercase tracking-wide bg-primary-500/20 text-primary-300 border border-primary-400/40 px-1.5 py-0.5 rounded"
                    >
                      You
                    </span>
                    <span
                      v-if="!admin.isActive"
                      class="text-[10px] uppercase tracking-wide bg-red-500/20 text-red-300 border border-red-400/40 px-1.5 py-0.5 rounded"
                    >
                      Disabled
                    </span>
                  </div>
                  <p class="text-xs text-neutral-400 mt-0.5">Created {{ formatDate(admin.createdAt) }}</p>
                </div>
              </div>

              <div class="flex gap-2 shrink-0">
                <button @click="startPasswordChange(admin)" class="mc-button text-sm secondary">
                  Password
                </button>
                <button
                  @click="removeAdmin(admin)"
                  class="mc-button text-sm danger"
                  :disabled="currentAdmin && admin.id === currentAdmin.id"
                  :title="currentAdmin && admin.id === currentAdmin.id ? 'You cannot delete your own account' : ''"
                >
                  Delete
                </button>
              </div>
            </div>

            <div v-if="admins.length === 0" class="text-center py-8 text-neutral-400">
              No admin accounts found.
            </div>
          </div>

          <!-- Create form -->
          <div class="max-w-2xl mx-auto glass p-6 rounded-xl border border-white/10">
            <h3 class="mc-title text-lg mb-1">Create Admin Account</h3>
            <p class="text-sm text-neutral-400 mb-4">
              New admins can log in with their username and password. All their actions are recorded in the change history.
            </p>

            <form @submit.prevent="createAdmin" class="space-y-4">
              <div class="form-group">
                <label for="new-admin-username" class="mc-label">Username</label>
                <input
                  id="new-admin-username"
                  v-model="newUsername"
                  type="text"
                  class="mc-input"
                  placeholder="3-32 characters (letters, numbers, _ or -)"
                  required
                  autocomplete="off"
                />
              </div>

              <div class="form-group">
                <label for="new-admin-password" class="mc-label">Password</label>
                <input
                  id="new-admin-password"
                  v-model="newPassword"
                  type="password"
                  class="mc-input"
                  placeholder="Minimum 6 characters"
                  required
                  autocomplete="new-password"
                />
              </div>

              <div class="flex justify-end">
                <button type="submit" class="mc-button" :disabled="isSubmitting">
                  <span v-if="isSubmitting">Creating...</span>
                  <span v-else>Create Admin</span>
                </button>
              </div>
            </form>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, inject, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import api from '../services/api';
import ProfileMenu from '../components/ProfileMenu.vue';

const SEED_USERNAME = 'admin';

const router = useRouter();
const confirmation = inject('confirmation');
const notification = inject('notification');

const admins = ref([]);
const currentAdmin = ref(null);
const isLoading = ref(true);
const isSubmitting = ref(false);
const errorMessage = ref('');
const notice = ref('');
const newUsername = ref('');
const newPassword = ref('');

function formatDate(dateString) {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });
}

function flash(message) {
  notice.value = message;
  setTimeout(() => {
    notice.value = '';
  }, 4000);
}

function goToDashboard() {
  router.push('/admin');
}

function logout() {
  api.logout();
  router.push('/admin');
}

async function loadAdmins() {
  errorMessage.value = '';
  try {
    const response = await api.getAdmins();
    admins.value = response.admins || [];
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Failed to load admin accounts.';
  }
}

async function createAdmin() {
  errorMessage.value = '';
  isSubmitting.value = true;

  try {
    const response = await api.createAdmin(newUsername.value.trim(), newPassword.value);
    if (response.success) {
      notification.success(`Admin "${response.admin.username}" created`);
      newUsername.value = '';
      newPassword.value = '';
      await loadAdmins();
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Failed to create admin account.';
  } finally {
    isSubmitting.value = false;
  }
}

async function startPasswordChange(admin) {
  errorMessage.value = '';
  const newPasswordValue = prompt(`New password for "${admin.username}" (minimum 6 characters):`);
  if (newPasswordValue === null) return;
  if (newPasswordValue.length < 6) {
    errorMessage.value = 'Password must be at least 6 characters.';
    return;
  }

  try {
    const response = await api.changeAdminPassword(admin.id, newPasswordValue);
    if (response.success) {
      flash(`Password updated for "${admin.username}".`);
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Failed to update password.';
  }
}

async function removeAdmin(admin) {
  errorMessage.value = '';

  const confirmed = await confirmation.confirm({
    title: 'Delete Admin',
    message: `Delete admin account "${admin.username}"? Their sessions will be revoked immediately. This cannot be undone.`,
    confirmText: 'Delete',
    cancelText: 'Cancel'
  });
  if (!confirmed) return;

  try {
    const response = await api.deleteAdmin(admin.id);
    if (response.success) {
      notification.success(`Admin "${admin.username}" deleted`);
      await loadAdmins();
    }
  } catch (error) {
    errorMessage.value = error?.response?.data?.error || 'Failed to delete admin account.';
  }
}

onMounted(async () => {
  const result = await api.validateToken();
  if (!result.valid) {
    router.push('/admin');
    return;
  }
  currentAdmin.value = result.admin || null;
  await loadAdmins();
  isLoading.value = false;
});
</script>
