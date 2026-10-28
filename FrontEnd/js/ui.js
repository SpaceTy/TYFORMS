function overlayRoot() {
  return document.getElementById("overlay-root") || document.body;
}

export function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

export function formatDate(dateString, includeTime = false) {
  if (!dateString) return "N/A";
  const date = new Date(dateString);
  if (Number.isNaN(date.getTime())) return "N/A";
  return date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    ...(includeTime ? { hour: "2-digit", minute: "2-digit" } : {}),
  });
}

export function getAgeClass(age) {
  if (age < 13) return "age-under-13";
  if (age < 16) return "age-13-15";
  if (age < 18) return "age-16-17";
  return "age-adult";
}

export function getAgeColorMobile(age) {
  if (age < 13) return "age-under-13";
  if (age < 16) return "age-13-15";
  if (age < 18) return "age-16-17";
  return "age-adult";
}

let toastState = null;

function removeToast() {
  if (!toastState) return;
  if (toastState.timer) clearTimeout(toastState.timer);
  toastState.el.remove();
  toastState = null;
}

function startToastTimer() {
  if (!toastState) return;
  toastState.startedAt = Date.now();
  toastState.timer = setTimeout(() => removeToast(), toastState.remaining);
}

function pauseToastTimer() {
  if (!toastState || !toastState.timer) return;
  clearTimeout(toastState.timer);
  toastState.timer = null;
  toastState.remaining -= Date.now() - toastState.startedAt;
}

function resumeToastTimer() {
  if (!toastState || toastState.timer) return;
  startToastTimer();
}

function showToast(type, message, undoAction = null) {
  removeToast();

  const el = document.createElement("div");
  el.className = `toast ${type}`;
  el.innerHTML = `
    <span class="toast-icon">${type === "success" ? "✓" : "✕"}</span>
    <div class="toast-message">${escapeHtml(message)}</div>
    ${undoAction ? '<button type="button" class="toast-undo">↶ Undo</button>' : ""}
    <button type="button" class="toast-close" aria-label="Close">✕</button>
  `;

  overlayRoot().appendChild(el);
  toastState = { el, timer: null, remaining: 5000, startedAt: Date.now() };

  el.addEventListener("mouseenter", pauseToastTimer);
  el.addEventListener("mouseleave", resumeToastTimer);

  const closeButton = el.querySelector(".toast-close");
  closeButton.addEventListener("click", removeToast);

  const undoButton = el.querySelector(".toast-undo");
  if (undoButton && undoAction) {
    undoButton.addEventListener("click", () => {
      removeToast();
      undoAction();
    });
  }

  startToastTimer();
}

export const toast = {
  success(message, undoAction) {
    showToast("success", message, undoAction);
  },
  error(message) {
    showToast("error", message);
  },
  close: removeToast,
};

export function confirmDialog({
  title = "Confirm Action",
  message = "Are you sure you want to proceed?",
  confirmText = "Confirm",
  cancelText = "Cancel",
  preventBackdropClose = false,
} = {}) {
  return new Promise((resolve) => {
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    const overlay = document.createElement("div");
    overlay.className = "overlay";
    overlay.innerHTML = `
      <div class="modal modal--confirm" role="dialog" aria-modal="true">
        <h3 class="modal-title" style="font-family: var(--font-pixel); font-size: 1rem; margin-bottom: 1rem;">${escapeHtml(title)}</h3>
        <p class="modal-text" style="margin-bottom: 1.5rem;">${escapeHtml(message)}</p>
        <div class="modal-actions" style="margin-top: 0;">
          <button type="button" class="modal-btn cancel">${escapeHtml(cancelText)}</button>
          <button type="button" class="modal-btn confirm">${escapeHtml(confirmText)}</button>
        </div>
      </div>
    `;

    const finish = (value) => {
      document.removeEventListener("keydown", onKeydown);
      overlay.remove();
      document.body.style.overflow = previousOverflow;
      resolve(value);
    };

    const onKeydown = (event) => {
      if (event.key === "Escape") finish(false);
    };

    overlay.querySelector(".modal-btn.cancel").addEventListener("click", () => finish(false));
    overlay.querySelector(".modal-btn.confirm").addEventListener("click", () => finish(true));

    if (!preventBackdropClose) {
      overlay.addEventListener("click", (event) => {
        if (event.target === overlay) finish(false);
      });
    }

    document.addEventListener("keydown", onKeydown);
    overlayRoot().appendChild(overlay);
    overlay.querySelector(".modal-btn.confirm").focus();
  });
}

export function notesEditor({ application, status = "pending" } = {}) {
  return new Promise((resolve) => {
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";

    let selectedStatus = status || "pending";
    const overlay = document.createElement("div");
    overlay.className = "overlay";
    overlay.innerHTML = `
      <div class="modal" role="dialog" aria-modal="true">
        <div class="modal-body">
          <h2 class="modal-title">Review Application</h2>
          <div class="modal-meta">
            <div><strong>MC Name:</strong> ${escapeHtml(application?.minecraftUsername || "")}</div>
            <div><strong>Discord:</strong> ${escapeHtml(application?.discordUsername || "")}</div>
            <div><strong>Age:</strong> ${escapeHtml(application?.age ?? "")}</div>
          </div>
          <div>
            <label class="modal-meta" style="display:block; margin-bottom:0.5rem; font-weight:600;">Status</label>
            <div class="modal-status-row">
              <button type="button" data-status="accepted" class="accept">Accept</button>
              <button type="button" data-status="pending" class="pending">Pending</button>
              <button type="button" data-status="rejected" class="reject">Reject</button>
            </div>
          </div>
          <div>
            <label class="modal-meta" style="display:block; margin-bottom:0.5rem; font-weight:600;">Review Notes</label>
            <textarea placeholder="Optional notes...">${escapeHtml(application?.reviewNotes || "")}</textarea>
          </div>
        </div>
        <div class="modal-footer">
          <button type="button" class="modal-btn cancel">Cancel</button>
          <button type="button" class="modal-btn save">Save</button>
        </div>
      </div>
    `;

    const textarea = overlay.querySelector("textarea");
    const statusButtons = Array.from(overlay.querySelectorAll("[data-status]"));

    const paintStatus = () => {
      statusButtons.forEach((button) => {
        button.classList.toggle("selected", button.dataset.status === selectedStatus);
      });
    };

    const finish = (value) => {
      document.removeEventListener("keydown", onKeydown);
      overlay.remove();
      document.body.style.overflow = previousOverflow;
      resolve(value);
    };

    const save = () => finish({ notes: textarea.value, status: selectedStatus });

    const onKeydown = (event) => {
      if (event.key === "Escape") finish(null);
    };

    statusButtons.forEach((button) => {
      button.addEventListener("click", () => {
        selectedStatus = button.dataset.status;
        paintStatus();
      });
    });

    overlay.querySelector(".modal-btn.cancel").addEventListener("click", () => finish(null));
    overlay.querySelector(".modal-btn.save").addEventListener("click", save);
    overlay.addEventListener("click", (event) => {
      if (event.target === overlay) finish(null);
    });
    textarea.addEventListener("keydown", (event) => {
      if (event.key === "Enter" && !event.shiftKey) {
        event.preventDefault();
        save();
      }
    });

    paintStatus();
    document.addEventListener("keydown", onKeydown);
    overlayRoot().appendChild(overlay);
    textarea.focus();
  });
}

export function notesViewer(application) {
  return new Promise((resolve) => {
    const overlay = document.createElement("div");
    overlay.className = "overlay";
    overlay.innerHTML = `
      <div class="modal" role="dialog" aria-modal="true">
        <div class="modal-body">
          <div style="display:flex; justify-content:space-between; align-items:center;">
            <h3 class="modal-title" style="font-size:1rem;">Review Notes</h3>
            <button type="button" class="modal-btn cancel" aria-label="Close">✕</button>
          </div>
          <p class="modal-meta"><strong>Player:</strong> ${escapeHtml(application?.minecraftUsername || "")} (${escapeHtml(application?.discordUsername || "")})</p>
          <div class="notes-view">${escapeHtml(application?.reviewNotes || "No notes provided.")}</div>
          <p class="modal-meta" style="color: var(--gold);">Reviewed on: ${escapeHtml(formatDate(application?.reviewedAt, true))}</p>
        </div>
        <div class="modal-footer">
          <button type="button" class="modal-btn save">Close</button>
        </div>
      </div>
    `;

    const finish = () => {
      document.removeEventListener("keydown", onKeydown);
      overlay.remove();
      resolve();
    };
    const onKeydown = (event) => {
      if (event.key === "Escape") finish();
    };

    overlay.querySelector(".modal-btn.cancel").addEventListener("click", finish);
    overlay.querySelector(".modal-btn.save").addEventListener("click", finish);
    overlay.addEventListener("click", (event) => {
      if (event.target === overlay) finish();
    });
    document.addEventListener("keydown", onKeydown);
    overlayRoot().appendChild(overlay);
  });
}

export function copyToClipboard(text, onCopied) {
  navigator.clipboard
    .writeText(text)
    .then(() => {
      if (onCopied) onCopied();
    })
    .catch((error) => {
      console.error("Failed to copy text", error);
    });
}

export function debounce(fn, delay) {
  let timer = null;
  return (...args) => {
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  };
}
