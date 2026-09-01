<template>
  <div class="relative" ref="menuRoot">
    <button
      @click="toggle"
      class="flex items-center gap-2 bg-white/5 border border-white/20 rounded-lg pl-1.5 pr-2.5 py-1 hover:bg-white/10 transition"
      :class="buttonClass"
    >
      <span class="w-7 h-7 rounded-full bg-primary-600/80 border border-primary-300/40 flex items-center justify-center text-white text-sm font-bold uppercase shrink-0">
        {{ initial }}
      </span>
      <span class="text-sm text-white max-w-[120px] truncate">{{ username }}</span>
      <svg class="w-3.5 h-3.5 text-neutral-400 transition-transform" :class="{ 'rotate-180': open }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7" />
      </svg>
    </button>

    <div
      v-if="open"
      class="absolute right-0 top-full mt-2 w-52 bg-neutral-900/95 backdrop-blur border border-white/15 rounded-xl shadow-2xl overflow-hidden z-[10000] animate-scale-in origin-top-right"
    >
      <div class="px-4 py-3 border-b border-white/10">
        <p class="text-xs text-neutral-500">Signed in as</p>
        <p class="text-sm text-white font-semibold truncate">{{ username }}</p>
      </div>
      <div class="py-1">
        <slot name="items"></slot>
        <button
          @click="doLogout"
          class="w-full text-left px-4 py-2.5 text-sm text-red-400 hover:bg-red-500/20 flex items-center gap-2 transition"
        >
          <span>➜</span> Logout
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';
import api from '../services/api';

const emit = defineEmits(['logout']);
const props = defineProps({
  buttonClass: {
    type: String,
    default: ''
  }
});

const menuRoot = ref(null);
const open = ref(false);

const username = computed(() => api.getUsername() || 'admin');
const initial = computed(() => (username.value ? username.value.charAt(0) : '?'));

function toggle() {
  open.value = !open.value;
}

function doLogout() {
  open.value = false;
  emit('logout');
}

function handleOutsideClick(event) {
  if (open.value && menuRoot.value && !menuRoot.value.contains(event.target)) {
    open.value = false;
  }
}

onMounted(() => {
  document.addEventListener('click', handleOutsideClick);
});

onUnmounted(() => {
  document.removeEventListener('click', handleOutsideClick);
});
</script>
