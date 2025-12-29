<template>
  <div
    class="min-h-screen flex flex-col relative"
    :class="{
      'py-10 px-4 sm:px-6': !isAdminRoute,
      'p-0': isAdminRoute
    }"
    style="background: #050505 radial-gradient(1200px 600px at 50% -10%, rgba(96,165,250,0.03), transparent), radial-gradient(800px 400px at 80% 20%, rgba(250,204,21,0.02), transparent);"
  >
    <!-- Header is hidden on admin route -->
    <header 
      v-if="!isAdminRoute" 
      class="max-w-4xl mx-auto text-center mb-10 relative z-1"
    >
      <h1 class="text-4xl sm:text-5xl font-pixel text-white drop-shadow mb-3 tracking-wider animate-fade-in">
        TYSMP
      </h1>
      <p class="text-sm sm:text-base text-primary-300 tracking-wide">
        Private Vanilla MC SMP Application
      </p>
    </header>
    
    <main 
      :class="{
        'max-w-4xl mx-auto w-full flex-grow flex relative z-1': !isAdminRoute,
        'flex-grow flex w-full': isAdminRoute
      }"
    >
      <router-view v-slot="{ Component }">
        <transition name="fade" mode="out-in" appear>
          <component :is="Component" class="w-full h-full" :key="$route.path" />
        </transition>
      </router-view>
    </main>
    
    <!-- Footer is hidden on admin route -->
    <footer 
      v-if="!isAdminRoute" 
      class="max-w-4xl mx-auto mt-10 text-center text-neutral-400 text-xs tracking-wide relative z-1"
    >
      &copy; {{ new Date().getFullYear() }} TYSMP · All rights reserved
    </footer>

    <!-- Global Confirmation Dialog -->
    <ConfirmationDialog
      :show="show"
      :title="title"
      :message="message"
      :confirm-text="confirmText"
      :cancel-text="cancelText"
      :prevent-backdrop-close="preventBackdropClose"
      @confirm="handleConfirm"
      @cancel="handleCancel"
      @update:show="show = $event"
    />
  </div>
</template>

<script setup>
import { computed, provide } from 'vue';
import { useRoute } from 'vue-router';
import ConfirmationDialog from './components/ConfirmationDialog.vue';
import { useConfirmation } from './composables/useConfirmation';

const route = useRoute();

const {
  show,
  title,
  message,
  confirmText,
  cancelText,
  preventBackdropClose,
  confirm,
  handleConfirm,
  handleCancel
} = useConfirmation();

// Provide the confirmation functionality to all child components
provide('confirmation', {
  confirm
});

// Check if we're on admin route to adjust layout
const isAdminRoute = computed(() => {
  return route.path === '/admin';
});

</script>

<style>
.text-shadow-lg {
  text-shadow: 0 0 10px rgba(0, 0, 0, 0.8);
}

/* Full height layout support */
html, body, #app {
  height: 100%;
  margin: 0;
  padding: 0;
  background-color: transparent;
  color: white;
}

.min-h-screen {
  min-height: 100%;
  display: flex;
  flex-direction: column;
}

/* Transition animations */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Glass effect for form elements */
.glass-panel {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(12px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 20px;
}
</style>
