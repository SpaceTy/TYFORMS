<template>
  <Teleport to="body">
    <div class="fixed inset-0 bg-black/85 flex items-center justify-center z-[9999] p-4" @click.self="$emit('close')">
      <div class="bg-neutral-900 border border-white/15 rounded-xl shadow-2xl w-full max-w-md max-h-[85vh] flex flex-col animate-scale-in">
        <div class="flex justify-between items-center px-5 py-4 border-b border-white/10 shrink-0">
          <h3 class="text-lg font-bold text-white">Admin Accounts</h3>
          <button @click="$emit('close')" class="text-neutral-400 hover:text-red-400 transition text-xl leading-none">&times;</button>
        </div>

        <div class="overflow-y-auto px-5 py-4 grow">
          <p v-if="error" class="bg-red-500/20 border border-red-500/40 text-red-200 text-sm rounded-lg p-3 mb-4">{{ error }}</p>
          <p v-if="notice" class="bg-green-500/15 border border-green-500/40 text-green-200 text-sm rounded-lg p-3 mb-4">{{ notice }}</p>

          <div v-if="isLoading" class="text-neutral-400 text-sm py-6 text-center">Loading admins...</div>

          <template v-else>
            <div class="space-y-2 mb-6">
              <div
                v-for="admin in admins"
                :key="admin.id"
                class="flex items-center justify-between bg-black/40 border border-white/10 rounded-lg px-4 py-3"
              >
                <div>
                  <p class="text-white text-sm font-medium">{{ admin.username }}</p>
                  <p class="text-xs text-neutral-500">Created {{ formatDate(admin.createdAt) }}</p>
                </div>
                <div class="flex gap-2">
                  <button
                    @click="startPasswordChange(admin)"
                    class="text-xs px-3 py-1.5 rounded-md border border-white/20 text-neutral-300 hover:bg-white/10 transition"
                  >
                    Password
                  </button>
                  <button
                    @click="removeAdmin(admin)"
                    class="text-xs px-3 py-1.5 rounded-md border border-red-500/40 text-red-300 hover:bg-red-500/20 transition"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>

            <div class="border-t border-white/10 pt-4">
              <h4 class="text-sm font-semibold text-neutral-300 mb-3">Create new admin</h4>
              <form @submit.prevent="createAdmin" class="space-y-3">
                <input
                  v-model="newUsername"
                  type="text"
                  placeholder="Username (3-32 chars)"
                  class="w-full bg-black/60 border border-white/20 rounded-lg px-4 py-2.5 text-white text-sm outline-none focus:ring-2 focus:ring-primary-500 transition placeholder-neutral-600"
                  required
                />
                <input
                  v-model="newPassword"
                  type="password"
                  placeholder="Password (min 6 chars)"
                  class="w-full bg-black/60 border border-white/20 rounded-lg px-4 py-2.5 text-white text-sm outline-none focus:ring-2 focus:ring-primary-500 transition placeholder-neutral-600"
                  required
                />
                <button
                  type="submit"
                  :disabled="isSubmitting"
                  class="w-full bg-primary-600 hover:bg-primary-500 disabled:opacity-50 text-white text-sm font-semibold py-2.5 rounded-lg transition"
                >
                  <span v-if="isSubmitting">Working...</span>
                  <span v-else>Create Admin</span>
                </button>
              </form>
            </div>
          </template>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import api from '../services/api';

defineEmits(['close']);

const admins = ref([]);
const isLoading = ref(true);
const isSubmitting = ref(false);
const error = ref('');
const notice = ref('');
const newUsername = ref('');
const newPassword = ref('');

function formatDate(dateString) {
  if (!dateString) return '';
  return new Date(dateString).toLocaleDateString();
}

function flash(message) {
  notice.value = message;
  setTimeout(() => {
    notice.value = '';
  }, 3000);
}

async function loadAdmins() {
  isLoading.value = true;
  error.value = '';
  try {
    const response = await api.getAdmins();
    admins.value = response.admins || [];
  } catch (err) {
    error.value = err?.response?.data?.error || 'Failed to load admin accounts.';
  } finally {
    isLoading.value = false;
  }
}

async function createAdmin() {
  error.value = '';
  isSubmitting.value = true;
  try {
    const response = await api.createAdmin(newUsername.value.trim(), newPassword.value);
    if (response.success) {
      newUsername.value = '';
      newPassword.value = '';
      flash(`Admin "${response.admin.username}" created.`);
      await loadAdmins();
    }
  } catch (err) {
    error.value = err?.response?.data?.error || 'Failed to create admin.';
  } finally {
    isSubmitting.value = false;
  }
}

async function startPasswordChange(admin) {
  error.value = '';
  const newPasswordValue = prompt(`New password for "${admin.username}" (min 6 chars):`);
  if (newPasswordValue === null) return;
  if (newPasswordValue.length < 6) {
    error.value = 'Password must be at least 6 characters.';
    return;
  }
  try {
    const response = await api.changeAdminPassword(admin.id, newPasswordValue);
    if (response.success) {
      flash(`Password updated for "${admin.username}".`);
    }
  } catch (err) {
    error.value = err?.response?.data?.error || 'Failed to update password.';
  }
}

async function removeAdmin(admin) {
  error.value = '';
  if (!confirm(`Delete admin account "${admin.username}"? This also revokes their sessions.`)) return;
  try {
    const response = await api.deleteAdmin(admin.id);
    if (response.success) {
      flash(`Admin "${admin.username}" deleted.`);
      await loadAdmins();
    }
  } catch (err) {
    error.value = err?.response?.data?.error || 'Failed to delete admin.';
  }
}

onMounted(loadAdmins);
</script>
