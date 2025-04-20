import { ref } from 'vue';

export function useConfirmation() {
  const show = ref(false);
  const title = ref('');
  const message = ref('');
  const confirmText = ref('Confirm');
  const cancelText = ref('Cancel');
  const preventBackdropClose = ref(false);
  const resolvePromise = ref(null);

  const confirm = (options = {}) => {
    return new Promise((resolve) => {
      show.value = true;
      title.value = options.title || 'Confirm Action';
      message.value = options.message || 'Are you sure you want to proceed?';
      confirmText.value = options.confirmText || 'Confirm';
      cancelText.value = options.cancelText || 'Cancel';
      preventBackdropClose.value = options.preventBackdropClose || false;
      resolvePromise.value = resolve;
    });
  };

  const handleConfirm = () => {
    if (resolvePromise.value) {
      resolvePromise.value(true);
      resolvePromise.value = null;
    }
    show.value = false;
  };

  const handleCancel = () => {
    if (resolvePromise.value) {
      resolvePromise.value(false);
      resolvePromise.value = null;
    }
    show.value = false;
  };

  return {
    show,
    title,
    message,
    confirmText,
    cancelText,
    preventBackdropClose,
    confirm,
    handleConfirm,
    handleCancel
  };
} 