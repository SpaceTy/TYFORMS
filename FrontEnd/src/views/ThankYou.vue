<template>
  <div class="mc-container">
    <div class="text-center thank-you-content">
      <h2 class="mc-title">Thank You for Applying!</h2>
      
      <div class="py-6 mb-6">
        <div class="mx-auto w-16 h-16 mb-4 text-4xl">
          <span class="animate-bounce inline-block text-green-400 drop-shadow">✅</span>
        </div>
        
        <p class="text-primary-300 text-xl mb-2 font-medium">
          Your application has been submitted successfully!
        </p>
        
        <p class="text-neutral-200 mb-2">
          Our admins will review your application and contact you via Discord if you are accepted.
        </p>
        
        <p class="text-neutral-300">
          Make sure you've joined our Discord server!
        </p>
      </div>
      
      <div class="flex flex-col sm:flex-row justify-center gap-3">
        <button @click="goToDiscord" class="mc-button">
          Join Discord
        </button>
        
        <button @click="goHome" class="mc-button secondary">
          Back to Form
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { gsap } from 'gsap';

const router = useRouter();

onMounted(() => {
  // Make sure the container is visible first
  gsap.set('.mc-container', { opacity: 1, display: 'block', visibility: 'visible' });
  
  // Improved entrance animation with a timeline
  const tl = gsap.timeline();
  
  // Animate the container first
  tl.from('.mc-container', {
    opacity: 0,
    duration: 0.5,
    ease: 'power2.out'
  });
  
  // Then animate the content
  tl.from('.thank-you-content', {
    opacity: 0,
    y: 20,
    duration: 0.8,
    ease: 'power2.out'
  }, '-=0.3'); // Slight overlap for smoother transition
});

function goToDiscord() {
  window.open('https://discord.com/invite/gqQFuEK3hk', '_blank');
}

function goHome() {
  router.push('/');
}
</script>

<style scoped>
/* Ensure elements remain visible after animation */
.thank-you-content {
  opacity: 1;
  transform: translateY(0);
  position: relative;
  z-index: 10;
}

.mc-container {
  display: block;
  opacity: 1;
  visibility: visible;
  position: relative;
  z-index: 5;
}
</style>
