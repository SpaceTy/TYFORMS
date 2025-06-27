<template>
  <div v-if="show" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
    <div class="mc-container max-w-6xl w-full animate-scale-in">
      <!-- Header -->
      <div class="flex justify-between items-center mb-4">
        <h3 class="mc-title text-lg">Edit Application</h3>
        <div class="flex items-center gap-2">
          <button 
            v-for="shortcut in shortcuts" 
            :key="shortcut.key"
            @click="shortcut.action"
            class="mc-button text-xs bg-minecraft-deepslate hover:bg-minecraft-stone"
            :title="shortcut.description"
          >
            {{ shortcut.label }}
          </button>
          <button @click="$emit('close')" class="text-white hover:text-red-500 transition">✕</button>
        </div>
      </div>
      
      <div v-if="application" class="flex flex-col lg:flex-row gap-4">
        <!-- Left Panel - Edit Form -->
        <div class="flex-1 bg-minecraft-deepslate/40 rounded-md p-4 border border-minecraft-stone/50 min-w-0">
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
            <!-- Basic Info Block -->
            <div class="bg-black/30 p-3 rounded border border-minecraft-stone/30 lg:col-span-1">
              <h4 class="text-minecraft-water font-minecraft mb-2">Basic Info</h4>
              <div class="space-y-3">
                <div>
                  <label class="mc-label text-xs">Minecraft Username</label>
                  <input 
                    v-model="editableApplication.minecraftUsername" 
                    class="mc-input text-sm"
                    @input="updatePreview"
                  />
                </div>
                <div>
                  <label class="mc-label text-xs">Discord Username</label>
                  <input 
                    v-model="editableApplication.discordUsername" 
                    class="mc-input text-sm"
                    @input="updatePreview"
                  />
                </div>
                <div>
                  <label class="mc-label text-xs">Age</label>
                  <input 
                    v-model="editableApplication.age" 
                    type="number"
                    class="mc-input text-sm"
                    @input="updatePreview"
                  />
                </div>
              </div>
            </div>

            <!-- Status Block -->
            <div class="bg-black/30 p-3 rounded border border-minecraft-stone/30 lg:col-span-1">
              <h4 class="text-minecraft-water font-minecraft mb-2">Status</h4>
              <div class="space-y-3">
                <div class="flex gap-2">
                  <button 
                    v-for="statusObj in [
                      { value: 'accepted', emoji: '✅', title: 'Accepted' },
                      { value: 'pending', emoji: '⏳', title: 'Pending' },
                      { value: 'rejected', emoji: '❌', title: 'Rejected' },
                    ]"
                    :key="statusObj.value"
                    @click="setStatus(statusObj.value)"
                    class="flex-1 mc-button text-lg py-1"
                    :class="{
                      'bg-green-600 hover:bg-green-700 border-green-800': statusObj.value === 'accepted',
                      'bg-yellow-600 hover:bg-yellow-700 border-yellow-800': statusObj.value === 'pending',
                      'bg-red-600 hover:bg-red-700 border-red-800': statusObj.value === 'rejected',
                      'ring-2 ring-white/80': editableApplication.acceptanceStatus === statusObj.value
                    }"
                    :title="statusObj.title"
                  >
                    {{ statusObj.emoji }}
                  </button>
                </div>
                <div>
                  <label class="mc-label text-xs">Review Notes</label>
                  <textarea 
                    v-model="editableApplication.reviewNotes" 
                    class="mc-input text-sm h-20"
                    @input="updatePreview"
                  ></textarea>
                </div>
              </div>
            </div>

            <!-- Application Details Block -->
            <div class="bg-black/30 p-3 rounded border border-minecraft-stone/30 lg:col-span-2">
              <h4 class="text-minecraft-water font-minecraft mb-2">Application Details</h4>
              <div class="grid grid-cols-1 md:grid-cols-2 gap-x-4 gap-y-3">
                <div v-for="([fieldName, value]) in editableFields" 
                     :key="fieldName"
                     class="space-y-1"
                >
                  <label class="mc-label text-xs">{{ formatLabel(fieldName) }}</label>
                  
                  <!-- Standard Text Input -->
                  <input 
                    v-if="typeof value === 'string'"
                    v-model="editableApplication[fieldName]"
                    class="mc-input text-sm"
                    @input="updatePreview"
                  />
                  
                  <!-- Checkbox for Booleans (e.g., joinedDiscord) -->
                  <div v-else-if="typeof value === 'boolean'" class="flex items-center h-10"> <!-- Adjusted height to match input -->
                    <input 
                      type="checkbox" 
                      v-model="editableApplication[fieldName]"
                      :id="`field-${fieldName}`"
                      class="mc-checkbox mr-2"
                      @change="updatePreview"
                    />
                    <label :for="`field-${fieldName}`" class="text-sm text-white">{{ value ? 'Yes' : 'No' }}</label>
                  </div>

                  <!-- Handle other types like numbers if necessary -->
                  <input
                    v-else-if="typeof value === 'number'"
                    v-model.number="editableApplication[fieldName]"
                    type="number"
                    class="mc-input text-sm"
                    @input="updatePreview"
                  />
                  
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Right Panel - Preview -->
        <div class="flex-1 lg:flex-none lg:w-1/2 lg:max-w-[50%] bg-minecraft-deepslate/40 rounded-md p-4 border border-minecraft-stone/50">
          <h4 class="text-minecraft-water font-minecraft mb-2">Preview</h4>
          <div class="bg-black/30 p-4 rounded border border-minecraft-stone/30">
            <div class="space-y-4">
              <div class="flex items-center gap-2">
                <div class="w-8 h-8 bg-minecraft-stone rounded flex-shrink-0"></div>
                <div class="min-w-0">
                  <h5 class="text-white font-minecraft truncate">{{ preview.minecraftUsername }}</h5>
                  <p class="text-gray-400 text-sm truncate">{{ preview.discordUsername }}</p>
                </div>
              </div>
              
              <div class="grid grid-cols-2 gap-2">
                <div class="bg-minecraft-deepslate/50 p-2 rounded">
                  <p class="text-xs text-gray-400">Age</p>
                  <p class="text-white" :class="getAgeColor(preview.age)">{{ preview.age }}</p>
                </div>
                <div class="p-2 rounded" :class="getStatusBackgroundClass(preview.acceptanceStatus)">
                  <p class="text-xs text-gray-400 opacity-80">Status</p>
                  <p class="font-bold" :class="getStatusTextColor(preview.acceptanceStatus)">
                    {{ preview.acceptanceStatus?.charAt(0).toUpperCase() + preview.acceptanceStatus?.slice(1) || 'N/A' }}
                  </p>
                </div>
              </div>

              <!-- Added FAM -->
              <div class="bg-minecraft-deepslate/50 p-2 rounded">
                <p class="text-xs text-gray-400">Favorite About Minecraft</p>
                <p class="text-white break-words text-sm">{{ preview.favoriteAboutMinecraft || '-' }}</p>
              </div>

              <!-- Added UOSMP -->
              <div class="bg-minecraft-deepslate/50 p-2 rounded">
                <p class="text-xs text-gray-400">Understanding of SMP</p>
                <p class="text-white break-words text-sm">{{ preview.understandingOfSMP || '-' }}</p>
              </div>

              <!-- Added Submission Date -->
              <div class="bg-minecraft-deepslate/50 p-2 rounded">
                <p class="text-xs text-gray-400">Submitted On</p>
                <p class="text-white text-sm">{{ formatDate(preview.submissionDate, true) }}</p>
              </div>

              <div v-if="preview.reviewNotes" class="bg-minecraft-deepslate/50 p-2 rounded">
                <p class="text-xs text-gray-400">Review Notes</p>
                <p class="text-white whitespace-pre-line break-words max-h-32 overflow-y-auto text-sm mc-scrollbar">{{ preview.reviewNotes }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex justify-end gap-2 mt-4">
        <button @click="$emit('close')" class="mc-button bg-gray-600 hover:bg-gray-700 border-gray-800">
          Cancel
        </button>
        <button 
          @click="saveChanges"
          class="mc-button bg-green-600 hover:bg-green-700 border-green-800"
          :disabled="isSaving"
        >
          <span v-if="isSaving">Saving...</span>
          <span v-else>Save Changes</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';

const props = defineProps({
  show: {
    type: Boolean,
    required: true
  },
  application: {
    type: Object,
    required: true
  }
});

const emit = defineEmits(['close', 'save']);

const editableApplication = ref({});
const preview = ref({});
const isSaving = ref(false);

// Keyboard shortcuts
const shortcuts = ref([
  { key: 'S', label: '⌘S', description: 'Save Changes', action: () => saveChanges() },
  { key: 'Esc', label: 'Esc', description: 'Close Modal', action: () => emit('close') },
  { key: 'Z', label: '⌘Z', description: 'Undo Changes', action: () => undoChanges() }
]);

const editableFields = computed(() => {
  const excludedKeys = [
    'id', 'isReviewed', 'reviewedAt', 'createdAt', 'reviewNotes', // Meta/Status fields
    'minecraftUsername', 'discordUsername', 'age', // Basic Info fields
    'acceptanceStatus' // Explicitly exclude status here too
  ];
  return Object.entries(editableApplication.value)
    .filter(([key]) => !excludedKeys.includes(key)); // Removed boolean filter
});

// Initialize editable application
watch(() => props.application, (newApp) => {
  if (newApp) {
    editableApplication.value = JSON.parse(JSON.stringify(newApp));
    preview.value = JSON.parse(JSON.stringify(newApp));
  }
}, { immediate: true });

// Format field labels
function formatLabel(key) {
  // Handle specific known acronyms first
  if (key === 'understandingOfSMP') return 'Understanding of SMP';
  
  return key
    // Insert space before capital letters, but not if preceded by another capital or start of string
    .replace(/([a-z])([A-Z])/g, '$1 $2') 
    // Insert space before capital letters that follow another capital, unless it's the last char
    .replace(/([A-Z])([A-Z])(?=[a-z])/g, '$1 $2') 
    // Uppercase the first letter
    .replace(/^./, str => str.toUpperCase());
}

// Update preview
function updatePreview() {
  preview.value = JSON.parse(JSON.stringify(editableApplication.value));
}

// Set status
function setStatus(status) {
  editableApplication.value.acceptanceStatus = status;
  updatePreview();
}

// Get age color
function getAgeColor(age) {
  if (age < 13) return 'text-red-500';
  if (age < 16) return 'text-yellow-400';
  if (age < 18) return 'text-green-400';
  return 'text-white';
}

// Get status text color
function getStatusTextColor(status) {
  switch (status) {
    case 'accepted': return 'text-green-400';
    case 'rejected': return 'text-red-400';
    case 'pending': return 'text-yellow-400';
    default: return 'text-gray-400';
  }
}

// Get status background class
function getStatusBackgroundClass(status) {
  switch (status) {
    case 'accepted': return 'bg-green-800/60 border border-green-600/80';
    case 'rejected': return 'bg-red-800/60 border border-red-600/80';
    case 'pending': return 'bg-yellow-800/60 border border-yellow-600/80';
    default: return 'bg-minecraft-deepslate/50 border border-minecraft-stone/50';
  }
}

// Format date for display
function formatDate(dateString, includeTime = false) {
  if (!dateString) return 'N/A';
  
  const date = new Date(dateString);
  const options = { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric',
    ...(includeTime ? { hour: '2-digit', minute: '2-digit' } : {})
  };
  
  return date.toLocaleDateString('en-US', options);
}

// Save changes
async function saveChanges() {
  isSaving.value = true;
  try {
    await emit('save', editableApplication.value);
    emit('close');
  } catch (error) {
    console.error('Error saving changes:', error);
  } finally {
    isSaving.value = false;
  }
}

// Undo changes
function undoChanges() {
  editableApplication.value = JSON.parse(JSON.stringify(props.application));
  updatePreview();
}

// Handle keyboard shortcuts
function handleKeyDown(e) {
  if (!props.show) return;
  
  const isMac = /^Mac/.test(navigator.platform);
  const cmdOrCtrl = isMac ? e.metaKey : e.ctrlKey;
  
  if (e.key === 'Escape') {
    emit('close');
  } else if (cmdOrCtrl && e.key.toLowerCase() === 's') {
    e.preventDefault();
    saveChanges();
  } else if (cmdOrCtrl && e.key.toLowerCase() === 'z') {
    e.preventDefault();
    undoChanges();
  }
}

// Add keyboard event listener
onMounted(() => {
  window.addEventListener('keydown', handleKeyDown);
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown);
});
</script>

<style scoped>
.animate-scale-in {
  animation: scaleIn 0.3s ease-out;
}

@keyframes scaleIn {
  from {
    opacity: 0;
    transform: scale(0.95);
  }
  to {
    opacity: 1;
    transform: scale(1);
  }
}
</style> 