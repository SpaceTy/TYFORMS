<template>
  <div class="flex flex-wrap justify-between items-center gap-2 md:gap-3 px-4 py-3 md:px-6 md:py-4 bg-black/70 sticky top-0 z-20 border-b border-white/10">
    <nav class="flex items-center gap-1">
      <router-link
        v-for="link in links"
        :key="link.to"
        :to="link.to"
        class="px-2 py-1 text-xs md:px-3 md:py-1.5 md:text-sm rounded-lg transition-colors border"
        :class="isActive(link.to)
          ? 'bg-primary-500/20 text-primary-200 border-primary-400/40'
          : 'text-neutral-300 hover:text-white hover:bg-white/5 border-transparent'"
      >
        {{ link.label }}
      </router-link>
    </nav>

    <div class="flex gap-2 items-center">
      <slot name="actions"></slot>
      <ProfileMenu @logout="$emit('logout')" />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import ProfileMenu from './ProfileMenu.vue';
import api from '../services/api';

defineEmits(['logout']);

const route = useRoute();

const canManageAdmins = ref(false);

const links = computed(() => {
  const base = [
    { label: 'Dashboard', to: '/admin' },
    { label: 'Statistics', to: '/admin/stats' }
  ];
  if (canManageAdmins.value) {
    base.push({ label: 'Admins', to: '/admin/accounts' });
  }
  return base;
});

function isActive(to) {
  return route.path === to;
}

onMounted(async () => {
  const result = await api.validateToken();
  canManageAdmins.value = !!result.admin?.canManageAdmins;
});
</script>
