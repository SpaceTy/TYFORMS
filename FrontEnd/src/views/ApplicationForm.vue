<template>
  <div class="mc-container animate-fade-in">
    <h2 class="mc-title text-center">TYSMP Application Form</h2>
    
    <p class="text-neutral-300 mb-6 text-center">
      Join the 
      <a href="#" @click.prevent="openDiscord" class="font-medium">discord</a>
    </p>
    
    <form @submit.prevent="submitForm" class="space-y-6">
      <!-- Discord Username -->
      <div class="form-group">
        <label for="discordUsername" class="mc-label">
          What is your discord username? <span class="text-red-500">*</span>
        </label>
        <input 
          id="discordUsername" 
          v-model="form.discordUsername" 
          type="text" 
          class="mc-input"
          required
        />
      </div>
      
      <!-- Minecraft Username -->
      <div class="form-group">
        <label for="minecraftUsername" class="mc-label">
          What is your minecraft username? <span class="text-red-500">*</span>
        </label>
        <input 
          id="minecraftUsername" 
          v-model="form.minecraftUsername" 
          type="text" 
          class="mc-input"
          required
        />
      </div>
      
      <!-- Age -->
      <div class="form-group age-selector-container">
        <label class="mc-label">
          How old are you? <span class="text-red-500">*</span>
        </label>
        <div class="flex items-center gap-4">
          <div class="w-full age-slider-container">
            <div class="flex items-center justify-between mb-2">
              <span class="text-white text-xs">13</span>
              <span class="text-white text-xs">99</span>
            </div>
            <div class="relative">
              <input 
                type="range" 
                min="13" 
                max="99" 
                step="1" 
                v-model="form.age" 
                class="age-slider w-full appearance-none bg-minecraft-deepslate/50 h-3 rounded-lg"
                id="ageSlider"
                required
                @input="handleAgeChange"
                @mousedown="showTooltip"
                @touchstart="showTooltip"
              />
              <div 
                class="age-tooltip absolute -top-12 left-0 text-white text-sm py-1 px-2 rounded-md transform -translate-x-1/2 shadow-soft"
                ref="ageTooltip"
                :class="[isTooltipVisible ? 'bg-primary-500/90' : 'opacity-0']"
              >
                {{ form.age }}
              </div>
            </div>
          </div>
          <div class="w-24 flex-shrink-0">
            <input 
              type="number" 
              v-model="form.age" 
              min="13" 
              max="99"
              class="age-input mc-input text-center" 
              required
              @input="handleNumberInput"
              @focus="showTooltip"
            />
          </div>
        </div>
        <div class="text-xs text-primary-300/80 mt-1">
          Drag the slider or type your age (13-99)
        </div>
      </div>
      
      <!-- Favorite About Minecraft -->
      <div class="form-group">
        <label for="favoriteAboutMinecraft" class="mc-label">
          What do you like most about Minecraft? <span class="text-red-500">*</span>
        </label>
        <input 
          id="favoriteAboutMinecraft" 
          v-model="form.favoriteAboutMinecraft" 
          type="text" 
          class="mc-input"
          required
        />
      </div>
      
      <!-- Understanding of SMP -->
      <div class="form-group">
        <label for="understandingOfSMP" class="mc-label">
          How do you understand the concept of a private vanilla SMP? <span class="text-red-500">*</span>
        </label>
        <input 
          id="understandingOfSMP" 
          v-model="form.understandingOfSMP" 
          type="text" 
          class="mc-input"
          required
        />
      </div>
      
      <!-- Joined Discord -->
      <div class="form-group">
        <label class="mc-label">
          Have you joined the discord server? (REQUIRED) link in description of form. <span class="text-red-500">*</span>
        </label>
        <div class="space-y-2">
          <div v-for="option in joinedDiscordOptions" :key="option.value" class="flex items-center">
            <input 
              :id="'joined-discord-' + option.value" 
              v-model="form.joinedDiscord" 
              type="radio" 
              :value="option.value" 
              class="mc-radio mr-2"
              required
            />
            <label :for="'joined-discord-' + option.value" class="text-white cursor-pointer">{{ option.label }}</label>
          </div>
        </div>
      </div>
      
      <!-- Error Message -->
      <div v-if="errorMessage" class="bg-red-600/40 text-white p-3 rounded-lg border border-red-500/30">
        {{ errorMessage }}
      </div>
      
      <!-- Submit Button -->
      <div class="text-center">
        <button 
          type="submit" 
          class="mc-button"
          :disabled="isSubmitting"
        >
          {{ isSubmitting ? 'Submitting...' : 'Submit Application' }}
        </button>
      </div>
    </form>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue';
import { useRouter } from 'vue-router';
import { gsap } from 'gsap';
import api from '../services/api';

const router = useRouter();
const isSubmitting = ref(false);
const errorMessage = ref('');
const ageTooltip = ref(null);
const isTooltipVisible = ref(false);
let tooltipTimer = null;

// Form data
const form = reactive({
  discordUsername: '',
  minecraftUsername: '',
  age: 16,
  favoriteAboutMinecraft: '',
  understandingOfSMP: '',
  joinedDiscord: null
});

// Options for joined discord
const joinedDiscordOptions = [
  { value: true, label: 'yes' },
  { value: false, label: 'no' }
];

onMounted(() => {
  // Initialize without showing tooltip
  updateTooltipPosition(false);
  
  // Add initial animation to the age slider
  gsap.from('.age-slider-container', {
    opacity: 0,
    x: -30,
    duration: 0.5,
    ease: 'power2.out'
  });
  
  gsap.from('.age-input', {
    opacity: 0,
    x: 30,
    duration: 0.5,
    ease: 'power2.out',
    delay: 0.2
  });
  
  // Show tooltip briefly on initial load then hide it
  showTooltip();
});

onBeforeUnmount(() => {
  // Clear any pending timers when component is destroyed
  if (tooltipTimer) clearTimeout(tooltipTimer);
});

// Show tooltip and set timer to hide it
function showTooltip() {
  // Clear any existing timer
  if (tooltipTimer) clearTimeout(tooltipTimer);
  
  // Show the tooltip with animation
  isTooltipVisible.value = true;
  updateTooltipPosition(true);
  
  // Set timer to hide tooltip after 3 seconds
  tooltipTimer = setTimeout(() => {
    hideTooltip();
  }, 3000);
}

// Hide tooltip with animation
function hideTooltip() {
  if (!ageTooltip.value) return;
  
  gsap.to(ageTooltip.value, {
    opacity: 0,
    y: -10,
    duration: 0.3,
    ease: 'power2.in',
    onComplete: () => {
      isTooltipVisible.value = false;
    }
  });
}

// Update tooltip position with optional animation
function updateTooltipPosition(animate = true) {
  if (!ageTooltip.value) return;
  
  // Calculate position based on slider value
  const slider = document.getElementById('ageSlider');
  if (!slider) return;
  
  const percent = (form.age - 13) / (99 - 13);
  const tooltipPos = percent * (slider.offsetWidth - 20) + 10;
  
  if (animate) {
    // Animated positioning
    gsap.to(ageTooltip.value, {
      left: tooltipPos,
      opacity: 1,
      y: -5,
      duration: 0.2,
      ease: 'power2.out',
      onComplete: () => {
        gsap.to(ageTooltip.value, {
          y: 0,
          duration: 0.2,
          ease: 'bounce.out'
        });
      }
    });
  } else {
    // Instant positioning without animation
    gsap.set(ageTooltip.value, {
      left: tooltipPos,
      opacity: isTooltipVisible.value ? 1 : 0,
      y: 0
    });
  }
}

// Handle age slider changes
function handleAgeChange() {
  showTooltip();
  updateTooltipPosition();
}

// Handle numeric input changes
function handleNumberInput() {
  checkAgeRange();
  showTooltip();
  updateTooltipPosition();
}

// Ensure age is within valid range
function checkAgeRange() {
  if (form.age < 13) form.age = 13;
  if (form.age > 99) form.age = 99;
}

// Submit form
async function submitForm() {
  if (isSubmitting.value) return;
  
  try {
    isSubmitting.value = true;
    errorMessage.value = '';
    
    await api.submitApplication(form);
    router.push('/thank-you');
  } catch (error) {
    if (error.response && error.response.status === 409) {
      errorMessage.value = 'A player with this Minecraft username already applied.';
    } else {
      errorMessage.value = 'There was an error submitting your application. Please try again.';
    }
  } finally {
    isSubmitting.value = false;
  }
}

// Open Discord link
function openDiscord() {
  window.open('https://discord.com/invite/gqQFuEK3hk', '_blank');
}
</script>

<style scoped>
/* Override some legacy specifics with the new theme */
.age-slider {
  appearance: none;
  height: 0.5rem;
  border-radius: 0.25rem;
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(4px);
}
.age-slider:focus { outline: none; }
.age-slider::-webkit-slider-thumb {
  appearance: none;
  width: 1.25rem;
  height: 1.25rem;
  border-radius: 9999px;
  background: rgb(59, 130, 246);
  border: 2px solid rgba(255, 255, 255, 0.5);
  cursor: pointer;
  transition: all 200ms;
}
.age-slider::-webkit-slider-thumb:hover { background: rgb(96, 165, 250); }

.age-tooltip {
  backdrop-filter: blur(4px);
  padding: 0.25rem 0.5rem;
  border-radius: 0.375rem;
  color: white;
  font-size: 0.875rem;
  line-height: 1.25rem;
  transition: opacity 200ms;
}

.form-group { margin-bottom: 1rem; }

.animate-fade-in { animation: fadeIn 0.5s ease-out; }
@keyframes fadeIn {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
