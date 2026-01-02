import { ref } from 'vue';

const show = ref(false);
const type = ref('success');
const message = ref('');
const undoAction = ref(null);
const onClose = ref(null);

export function useNotification() {
  function success(msg, undo) {
    type.value = 'success';
    message.value = msg;
    undoAction.value = undo;
    showNotification();
  }

  function error(msg) {
    type.value = 'error';
    message.value = msg;
    undoAction.value = null;
    showNotification();
  }

  function showNotification(options = {}) {
    if (options.type) type.value = options.type;
    if (options.message) message.value = options.message;
    if (options.undoAction) undoAction.value = options.undoAction;
    if (options.onClose) onClose.value = options.onClose;
    show.value = true;
  }

  function close() {
    show.value = false;
    undoAction.value = null;
  }

  return {
    show,
    type,
    message,
    undoAction,
    onClose,
    success,
    error,
    showNotification,
    close
  };
}
