<template>
  <div 
    class="min-h-screen flex flex-col"
    :class="{
      'py-10 px-4 sm:px-6': !isAdminRoute,
      'p-0': isAdminRoute
    }"
  >
    <!-- Header is hidden on admin route -->
    <header 
      v-if="!isAdminRoute" 
      class="max-w-4xl mx-auto text-center mb-10"
    >
      <h1 class="text-3xl sm:text-4xl font-pixel text-white text-shadow-lg mb-4 tracking-wider">
        TYSMP
      </h1>
      <p class="text-sm sm:text-base text-minecraft-water font-minecraft tracking-wide">
        Private Vanilla MC SMP Application
      </p>
    </header>
    
    <main 
      :class="{
        'max-w-4xl mx-auto w-full flex-grow flex': !isAdminRoute,
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
      class="max-w-4xl mx-auto mt-10 text-center text-minecraft-stone font-minecraft text-xs tracking-wide"
    >
      &copy; {{ new Date().getFullYear() }} TYSMP. All rights reserved.
    </footer>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();

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
</style> 