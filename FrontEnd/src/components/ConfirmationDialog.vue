<template>
  <Teleport to="body">
    <Transition name="dialog">
      <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center">
        <!-- Backdrop -->
        <div 
          class="absolute inset-0 bg-black/70 "
          @click="handleBackdropClick"
        ></div>
        
        <!-- Dialog -->
        <div 
          class="relative bg-minecraft-dark p-6 rounded-lg shadow-xl max-w-md w-full mx-4 transform transition-all"
          :class="{
            'scale-95': !show,
            'scale-100': show
          }"
        >
          <!-- Title -->
          <h3 class="text-xl font-minecraft text-white mb-4">
            {{ title }}
          </h3>
          
          <!-- Message -->
          <p class="text-minecraft-stone mb-6">
            {{ message }}
          </p>
          
          <!-- Actions -->
          <div class="flex justify-end space-x-4">
            <button
              class="px-4 py-2 text-minecraft-stone hover:text-white transition-colors"
              @click="handleCancel"
            >
              {{ cancelText }}
            </button>
            <button
              class="px-4 py-2 bg-minecraft-green text-white rounded hover:bg-minecraft-green/90 transition-colors"
              @click="handleConfirm"
            >
              {{ confirmText }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue';

const props = defineProps({
  show: {
    type: Boolean,
    required: true
  },
  title: {
    type: String,
    default: 'Confirm Action'
  },
  message: {
    type: String,
    required: true
  },
  confirmText: {
    type: String,
    default: 'Confirm'
  },
  cancelText: {
    type: String,
    default: 'Cancel'
  },
  preventBackdropClose: {
    type: Boolean,
    default: false
  }
});

const emit = defineEmits(['confirm', 'cancel', 'update:show']);

const handleConfirm = () => {
  emit('confirm');
  emit('update:show', false);
};

const handleCancel = () => {
  emit('cancel');
  emit('update:show', false);
};

const handleBackdropClick = () => {
  if (!props.preventBackdropClose) {
    handleCancel();
  }
};

// Prevent body scroll when dialog is open
onMounted(() => {
  if (props.show) {
    document.body.style.overflow = 'hidden';
  }
});

onUnmounted(() => {
  document.body.style.overflow = '';
});
</script>

<style scoped>
.dialog-enter-active,
.dialog-leave-active {
  transition: opacity 0.3s ease;
}

.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}

.dialog-enter-active .dialog-content,
.dialog-leave-active .dialog-content {
  transition: transform 0.3s ease;
}

.dialog-enter-from .dialog-content,
.dialog-leave-to .dialog-content {
  transform: scale(0.95);
}
</style> 