import api from "../api.js";
import { createNavbar } from "../components.js";
import { navigate } from "../router.js";
import { confirmDialog, escapeHtml, formatDate, toast } from "../ui.js";

const SEED_USERNAME = "admin";

export default function render(root) {
  let currentAdmin = null;
  let admins = [];
  let isLoading = true;
  let isSubmitting = false;
  let isTogglingPermission = null;
  let noticeTimer = null;
  let navbarCleanup = null;

  function cleanup() {
    if (navbarCleanup) navbarCleanup();
    navbarCleanup = null;
    if (noticeTimer) clearTimeout(noticeTimer);
  }

  function isSelf(admin) {
    return currentAdmin && admin.id === currentAdmin.id;
  }

  function showError(message) {
    const box = root.querySelector("#accountsError");
    if (!box) return;
    box.textContent = message;
    box.classList.toggle("hidden", !message);
  }

  function flash(message) {
    const box = root.querySelector("#accountsNotice");
    if (!box) return;
    box.textContent = message;
    box.classList.toggle("hidden", !message);
    if (noticeTimer) clearTimeout(noticeTimer);
    if (message) {
      noticeTimer = setTimeout(() => box.classList.add("hidden"), 4000);
    }
  }

  async function showDashboard() {
    root.innerHTML = `
      <div class="admin-dashboard admin-dashboard--flush">
        <div class="admin-panel mc-panel admin-panel--flush">
          <div id="navbarSlot"></div>
          <div class="admin-content admin-content--scroll">
            <div class="admin-error hidden" id="accountsError"></div>
            <div class="admin-notice hidden" id="accountsNotice"></div>
            <div id="accountsBody" class="muted">Loading admin accounts...</div>
          </div>
        </div>
      </div>
    `;

    const navbar = await createNavbar({ onLogout: logout });
    navbarCleanup = navbar.cleanup;
    root.querySelector("#navbarSlot").replaceWith(navbar.el);

    await loadAdmins();
    isLoading = false;
    renderAccounts();
  }

  async function loadAdmins() {
    showError("");
    try {
      const response = await api.getAdmins();
      admins = response.admins || [];
    } catch (error) {
      if (error.status === 401) {
        navigate("/admin", { replace: true });
        return;
      }
      showError((error.data && error.data.error) || "Failed to load admin accounts.");
    }
  }

  function renderAccounts() {
    const body = root.querySelector("#accountsBody");
    if (!body) return;

    if (isLoading) {
      body.textContent = "Loading admin accounts...";
      body.className = "muted";
      return;
    }

    body.className = "";
    body.innerHTML = `
      <div class="accounts-wrap">
        <div id="accountList"></div>
        <div class="create-admin glass">
          <h3 class="mc-title" style="font-size:1rem; margin-bottom:0.25rem;">Create Admin Account</h3>
          <p class="hint">New admins can log in with their username and password. All their actions are recorded in the change history.</p>
          <form id="createAdminForm">
            <div class="form-group">
              <label class="mc-label" for="newAdminUsername">Username</label>
              <input class="mc-input" id="newAdminUsername" type="text" placeholder="3-32 characters (letters, numbers, _ or -)" required autocomplete="off">
            </div>
            <div class="form-group">
              <label class="mc-label" for="newAdminPassword">Password</label>
              <input class="mc-input" id="newAdminPassword" type="password" placeholder="Minimum 6 characters" required autocomplete="new-password">
            </div>
            <div class="actions">
              <button type="submit" class="mc-button" id="createAdminButton">Create Admin</button>
            </div>
          </form>
        </div>
      </div>
    `;

    renderList();
    wireCreateForm();
  }

  function renderList() {
    const container = root.querySelector("#accountList");
    if (!container) return;

    if (admins.length === 0) {
      container.innerHTML = '<div class="text-center muted" style="padding:2rem 0;">No admin accounts found.</div>';
      return;
    }

    container.innerHTML = admins.map(accountHtml).join("");

    container.querySelectorAll("[data-permission]").forEach((checkbox) => {
      const id = Number(checkbox.dataset.permission);
      const admin = admins.find((item) => item.id === id);
      checkbox.addEventListener("change", () => toggleManage(admin, checkbox));
    });

    container.querySelectorAll("[data-password]").forEach((button) => {
      const id = Number(button.dataset.password);
      const admin = admins.find((item) => item.id === id);
      button.addEventListener("click", () => startPasswordChange(admin));
    });

    container.querySelectorAll("[data-delete]").forEach((button) => {
      const id = Number(button.dataset.delete);
      const admin = admins.find((item) => item.id === id);
      button.addEventListener("click", () => removeAdmin(admin));
    });
  }

  function accountHtml(admin) {
    const self = isSelf(admin);
    const waiting = isTogglingPermission === admin.id;
    const disabled = self || waiting;

    return `
      <div class="account-row glass">
        <div class="account-main">
          <span class="account-avatar">${escapeHtml((admin.username || "?").charAt(0))}</span>
          <div style="min-width:0;">
            <div class="account-name-row">
              <p class="account-name">${escapeHtml(admin.username)}</p>
              ${admin.username === SEED_USERNAME ? '<span class="account-badge seed">Seed</span>' : ""}
              ${self ? '<span class="account-badge you">You</span>' : ""}
              ${!admin.isActive ? '<span class="account-badge disabled">Disabled</span>' : ""}
            </div>
            <p class="account-meta">Created ${escapeHtml(formatDate(admin.createdAt))}</p>
          </div>
        </div>

        <div class="account-actions">
          <label class="permission-label ${disabled ? (waiting ? "waiting" : "disabled") : ""}" title="${
            self ? "You cannot change your own admin-management permission" : "Allow this admin to access and manage the admins page"
          }">
            <input type="checkbox" class="mc-radio" data-permission="${admin.id}" ${
              admin.canManageAdmins ? "checked" : ""
            } ${disabled ? "disabled" : ""}>
            <span>Manage Admins</span>
          </label>
          <button type="button" class="mc-button small secondary" data-password="${admin.id}">Password</button>
          <button type="button" class="mc-button small danger" data-delete="${admin.id}" ${
            self ? "disabled" : ""
          } title="${self ? "You cannot delete your own account" : ""}">Delete</button>
        </div>
      </div>
    `;
  }

  async function toggleManage(admin, checkbox) {
    const canManage = checkbox.checked;
    isTogglingPermission = admin.id;
    showError("");

    try {
      await api.setAdminPermissions(admin.id, canManage);
      admin.canManageAdmins = canManage;
      flash(`Admin-management permission ${canManage ? "granted to" : "removed from"} "${admin.username}".`);
    } catch (error) {
      checkbox.checked = admin.canManageAdmins;
      showError((error.data && error.data.error) || "Failed to update permissions.");
    } finally {
      isTogglingPermission = null;
      renderList();
    }
  }

  function wireCreateForm() {
    const form = root.querySelector("#createAdminForm");
    if (!form) return;

    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      if (isSubmitting) return;

      const username = root.querySelector("#newAdminUsername").value.trim();
      const password = root.querySelector("#newAdminPassword").value;
      isSubmitting = true;
      showError("");

      const button = root.querySelector("#createAdminButton");
      if (button) {
        button.disabled = true;
        button.textContent = "Creating...";
      }

      try {
        const response = await api.createAdmin(username, password);
        if (response.success) {
          toast.success(`Admin "${response.admin.username}" created`);
          form.reset();
          await loadAdmins();
          renderList();
        }
      } catch (error) {
        showError((error.data && error.data.error) || "Failed to create admin account.");
      } finally {
        isSubmitting = false;
        const current = root.querySelector("#createAdminButton");
        if (current) {
          current.disabled = false;
          current.textContent = "Create Admin";
        }
      }
    });
  }

  async function startPasswordChange(admin) {
    showError("");
    const newPassword = prompt(`New password for "${admin.username}" (minimum 6 characters):`);
    if (newPassword === null) return;
    if (newPassword.length < 6) {
      showError("Password must be at least 6 characters.");
      return;
    }

    try {
      const response = await api.changeAdminPassword(admin.id, newPassword);
      if (response.success) flash(`Password updated for "${admin.username}".`);
    } catch (error) {
      showError((error.data && error.data.error) || "Failed to update password.");
    }
  }

  async function removeAdmin(admin) {
    showError("");
    const confirmed = await confirmDialog({
      title: "Delete Admin",
      message: `Delete admin account "${admin.username}"? Their sessions will be revoked immediately. This cannot be undone.`,
      confirmText: "Delete",
      cancelText: "Cancel",
    });
    if (!confirmed) return;

    try {
      const response = await api.deleteAdmin(admin.id);
      if (response.success) {
        toast.success(`Admin "${admin.username}" deleted`);
        await loadAdmins();
        renderList();
      }
    } catch (error) {
      showError((error.data && error.data.error) || "Failed to delete admin account.");
    }
  }

  async function logout() {
    await api.logout();
    navigate("/admin", { replace: true });
  }

  api.validateToken().then(async (result) => {
    if (!result || !result.valid) {
      navigate("/admin", { replace: true });
      return;
    }
    currentAdmin = result.admin || null;
    if (!currentAdmin || !currentAdmin.canManageAdmins) {
      navigate("/admin", { replace: true });
      return;
    }
    await showDashboard();
  });

  return cleanup;
}
