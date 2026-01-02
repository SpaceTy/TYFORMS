<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="show"
        @click.self="$emit('close')"
        class="fixed inset-0 z-[99999] flex items-center justify-center bg-black/60 backdrop-blur-sm"
      >
        <div
          class="bg-stone-800 border border-stone-600 rounded-lg shadow-xl max-w-lg w-full mx-4 overflow-hidden"
          @keydown.esc="$emit('close')"
        >
          <div class="p-6">
            <h2 class="text-xl font-bold text-stone-100 mb-4">Review Application</h2>

            <div v-if="application" class="space-y-3 mb-6">
              <div class="text-sm text-stone-400">
                <span class="font-semibold text-stone-200">MC Name:</span> {{ application.minecraftUsername }}
              </div>
              <div class="text-sm text-stone-400">
                <span class="font-semibold text-stone-200">Discord:</span> {{ application.discordName }}
              </div>
              <div class="text-sm text-stone-400">
                <span class="font-semibold text-stone-200">Age:</span> {{ application.age }}
              </div>
            </div>

            <div class="mb-6">
              <label class="block text-sm font-semibold text-stone-200 mb-2">Status</label>
              <div class="flex gap-2">
                <button
                  @click="selectedStatus = 'accepted'"
                  :class="[
                    'flex-1 py-2 rounded-lg text-xs font-bold transition-colors',
                    selectedStatus === 'accepted' ? 'bg-green-600 text-white' : 'bg-green-600/20 text-green-400 hover:bg-green-600/40'
                  ]"
                >
                  Accept
                </button>
                <button
                  @click="selectedStatus = 'pending'"
                  :class="[
                    'flex-1 py-2 rounded-lg text-xs font-bold transition-colors',
                    selectedStatus === 'pending' ? 'bg-yellow-600 text-white' : 'bg-yellow-600/20 text-yellow-400 hover:bg-yellow-600/40'
                  ]"
                >
                  Pending
                </button>
                <button
                  @click="selectedStatus = 'rejected'"
                  :class="[
                    'flex-1 py-2 rounded-lg text-xs font-bold transition-colors',
                    selectedStatus === 'rejected' ? 'bg-red-600 text-white' : 'bg-red-600/20 text-red-400 hover:bg-red-600/40'
                  ]"
                >
                  Reject
                </button>
              </div>
            </div>

            <div class="mb-6">
              <label class="block text-sm font-semibold text-stone-200 mb-2">Review Notes</label>
              <textarea
                v-model="notes"
                placeholder="Optional notes..."
                @keydown.enter.prevent="handleSave"
                class="w-full p-3 bg-stone-900/50 border border-stone-600 rounded-lg text-stone-100 placeholder-stone-500 focus:outline-none focus:border-stone-400 resize-none"
                rows="4"
              />
            </div>
          </div>

          <div class="px-6 py-4 bg-stone-900/50 border-t border-stone-700 flex justify-end gap-3">
            <button
              @click="$emit('close')"
              class="px-4 py-2 rounded-lg text-sm font-bold text-stone-300 hover:text-stone-100 hover:bg-stone-700 transition-colors"
            >
              Cancel
            </button>
            <button
              @click="handleSave"
              class="px-4 py-2 rounded-lg text-sm font-bold bg-stone-600 text-white hover:bg-stone-500 transition-colors"
            >
              Save
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch } from 'vue';

const props = defineProps({
  show: Boolean,
  application: Object,
  initialStatus: String
});

const emit = defineEmits(['close', 'save']);

const selectedStatus = ref(props.initialStatus || 'pending');
const notes = ref('');

watch(() => props.initialStatus, (newVal) => {
  if (newVal) {
    selectedStatus.value = newVal;
  }
});

watch(() => props.application, (newVal) => {
  if (newVal) {
    notes.value = newVal.reviewNotes || '';
  }
});

function handleSave() {
  emit('save', notes.value, selectedStatus.value);
}
</script>
