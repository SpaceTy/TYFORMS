<template>
  <Teleport to="body">
    <Transition name="toast">
      <div
        v-if="show"
        @mouseenter="pauseTimer"
        @mouseleave="resumeTimer"
        class="fixed bottom-4 right-4 z-[99998] min-w-[320px] max-w-md p-4 rounded-lg shadow-lg backdrop-blur-sm flex items-start gap-3"
        :class="{
          'bg-green-900/90 border border-green-500/30 text-green-100': type === 'success',
          'bg-red-900/90 border border-red-500/30 text-red-100': type === 'error'
        }"
      >
        <span class="text-xl mt-0.5">
          {{ type === 'success' ? '✓' : '✕' }}
        </span>
        <div class="flex-1">
          <p class="font-medium">{{ message }}</p>
        </div>
        <button
          v-if="undoAction"
          @click="handleUndo"
          class="flex items-center gap-1 px-2 py-1 rounded text-xs font-bold bg-white/10 hover:bg-white/20 transition-colors"
        >
          ↶ Undo
        </button>
        <button
          @click="close"
          class="ml-2 text-current/60 hover:text-current transition-colors"
        >
          ✕
        </button>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch, onUnmounted } from 'vue';

const props = defineProps({
  show: Boolean,
  type: String,
  message: String,
  undoAction: Function,
  onClose: Function
});

const emit = defineEmits(['undo']);

let timer = null;
const AUTO_DISMISS = 5000;

watch(() => props.show, (newVal) => {
  if (newVal) {
    startTimer();
  } else {
    clearTimer();
  }
});

function startTimer() {
  clearTimer();
  timer = setTimeout(() => {
    props.onClose?.();
  }, AUTO_DISMISS);
}

function pauseTimer() {
  clearTimer();
}

function resumeTimer() {
  startTimer();
}

function clearTimer() {
  if (timer) {
    clearTimeout(timer);
    timer = null;
  }
}

function handleUndo() {
  if (props.undoAction) {
    props.undoAction();
  }
  emit('undo');
}

function close() {
  clearTimer();
  props.onClose?.();
}

onUnmounted(() => {
  clearTimer();
});
</script>
