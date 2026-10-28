import api from "../api.js";
import { createProfileMenu } from "../components.js";
import { navigate } from "../router.js";
import {
  confirmDialog,
  copyToClipboard,
  escapeHtml,
  formatDate,
  getAgeClass,
  notesEditor,
  notesViewer,
  toast,
} from "../ui.js";

const UNDO_TIMEOUT = 10000;

const FIELD_LABELS = {
  discordUsername: "Discord",
  minecraftUsername: "IGN",
  favoriteAboutMinecraft: "Favorite",
  understandingOfSMP: "Understanding",
  id: "ID",
};

const SEARCH_ICON = `
  <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
  </svg>`;

const MENU_ICON = `
  <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
  </svg>`;

const FILE_ICON = `
  <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
  </svg>`;

const COPY_ICON = `
  <svg style="width:0.75rem;height:0.75rem;color:var(--muted-3);" viewBox="0 0 20 20" fill="currentColor">
    <path d="M8 3a1 1 0 011-1h2a1 1 0 110 2H9a1 1 0 01-1-1z" />
    <path d="M6 3a2 2 0 00-2 2v11a2 2 0 002 2h8a2 2 0 002-2V5a2 2 0 00-2-2 3 3 0 01-3 3H9a3 3 0 01-3-3z" />
  </svg>`;

export default function render(root) {
  let authenticated = false;
  let localPassword = "";
  let canManageAdmins = false;
  let applications = [];
  let searchQuery = "";
  const searchFields = {
    discordUsername: true,
    minecraftUsername: true,
    favoriteAboutMinecraft: false,
    understandingOfSMP: false,
    id: false,
  };
  let showSearch = false;
  let showMenu = false;
  let currentPage = 1;
  let pageSize = 50;
  let total = 0;
  let totalPages = 1;
  let lastAction = null;
  let profileCleanup = null;
  let searchTimer = null;
  let copiedTimer = null;
  const expanded = new Set();

  function cleanup() {
    if (profileCleanup) profileCleanup();
    profileCleanup = null;
    window.removeEventListener("keydown", onKeydown);
    if (searchTimer) clearTimeout(searchTimer);
    if (copiedTimer) clearTimeout(copiedTimer);
  }

  function showLogin() {
    authenticated = false;
    if (profileCleanup) {
      profileCleanup();
      profileCleanup = null;
    }

    root.innerHTML = `
      <div class="mobile-view">
        <div class="mobile-login">
          <div class="mc-container fade-in">
            <h2 class="mc-title text-center" style="font-size:1.5rem;">Admin Access</h2>
            <form id="mobileLoginForm">
              <div>
                <label class="mc-label" for="mobileUsername">Username</label>
                <input class="mc-input" id="mobileUsername" type="text" placeholder="Username (optional)" autocomplete="username" value="${escapeHtml(api.getUsername())}">
              </div>
              <div>
                <label class="mc-label" for="mobilePassword">Password</label>
                <input class="mc-input" id="mobilePassword" type="password" placeholder="Enter admin password" required autocomplete="current-password">
              </div>
              <div class="form-error hidden" id="mobileAuthError"></div>
              <button type="submit" class="mc-button" style="width:100%;">Login</button>
            </form>
          </div>
        </div>
      </div>
    `;

    const form = root.querySelector("#mobileLoginForm");
    const passwordInput = root.querySelector("#mobilePassword");
    const errorBox = root.querySelector("#mobileAuthError");

    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const username = root.querySelector("#mobileUsername").value.trim();

      try {
        const response = await api.login(username, passwordInput.value);
        if (response.success) {
          canManageAdmins = !!(response.admin && response.admin.canManageAdmins);
          localPassword = "";
          searchQuery = "";
          applications = [];
          currentPage = 1;
          await showDashboard();
        } else {
          errorBox.textContent = "Invalid username or password";
          errorBox.classList.remove("hidden");
        }
      } catch (error) {
        errorBox.textContent =
          error.status === 401 ? "Invalid username or password" : "Authentication error. Please try again.";
        errorBox.classList.remove("hidden");
      }
    });

    passwordInput.focus();
  }

  async function showDashboard() {
    authenticated = true;

    root.innerHTML = `
      <div class="mobile-view">
        <header class="mobile-header">
          <div class="mobile-header-side">
            <span class="mobile-count" id="mobileCount">${total}</span>
          </div>
          <div class="mobile-header-side">
            <button type="button" class="icon-btn" id="searchToggle" aria-label="Search">${SEARCH_ICON}</button>
            <div class="mobile-menu-wrap">
              <button type="button" class="icon-btn" id="menuToggle" aria-label="Menu">${MENU_ICON}</button>
              <div class="mobile-menu hidden" id="mobileMenu">
                <button type="button" data-menu="refresh"><span id="refreshIcon">↻</span> Refresh Data</button>
                <button type="button" data-menu="export"><span>⬇</span> Export CSV</button>
                <button type="button" data-menu="stats"><span>◫</span> Statistics</button>
                ${
                  canManageAdmins
                    ? '<button type="button" data-menu="accounts"><span>⚙</span> Manage Admins</button>'
                    : ""
                }
              </div>
            </div>
            <div id="mobileProfileSlot"></div>
          </div>
        </header>

        <div class="mobile-search hidden" id="mobileSearch">
          <div class="search-input-wrap">
            ${SEARCH_ICON}
            <input id="mobileSearchInput" type="text" placeholder="Search username, discord...">
            <button type="button" class="search-clear hidden" id="searchClear">✕</button>
          </div>
          <div class="search-chips">
            ${Object.keys(searchFields)
              .map(
                (field) => `
                <label class="chip ${searchFields[field] ? "active" : ""}" data-chip="${field}">
                  <input type="checkbox" data-field="${field}" ${searchFields[field] ? "checked" : ""}>
                  ${FIELD_LABELS[field]}
                </label>`
              )
              .join("")}
          </div>
        </div>

        <div class="mobile-error hidden" id="mobileError">
          <span></span>
          <button type="button" id="errorClose">×</button>
        </div>

        <div class="mobile-list" id="mobileList"></div>
      </div>
    `;

    const profile = createProfileMenu(handleLogoutRequest);
    profileCleanup = profile.cleanup;
    root.querySelector("#mobileProfileSlot").appendChild(profile.el);

    wireHeader();
    wireSearch();
    wireList();
    wireBackdrop();
    await loadData();
  }

  function wireHeader() {
    const menuToggle = root.querySelector("#menuToggle");
    const menu = root.querySelector("#mobileMenu");
    const searchToggle = root.querySelector("#searchToggle");

    menuToggle.addEventListener("click", () => {
      showMenu = !showMenu;
      menu.classList.toggle("hidden", !showMenu);
      let backdrop = root.querySelector(".mobile-menu-backdrop");
      if (showMenu) {
        if (!backdrop) {
          backdrop = document.createElement("div");
          backdrop.className = "mobile-menu-backdrop";
          backdrop.addEventListener("click", closeMenu);
          root.querySelector(".mobile-header").appendChild(backdrop);
        }
      } else if (backdrop) {
        backdrop.remove();
      }
    });

    menu.querySelectorAll("[data-menu]").forEach((button) => {
      button.addEventListener("click", async () => {
        const action = button.dataset.menu;
        closeMenu();
        if (action === "refresh") await loadData();
        else if (action === "export") await exportCsv();
        else if (action === "stats") navigate("/admin/stats");
        else if (action === "accounts") navigate("/admin/accounts");
      });
    });

    searchToggle.addEventListener("click", () => {
      showSearch = !showSearch;
      toggleSearch();
    });

    root.querySelector("#errorClose").addEventListener("click", () => {
      root.querySelector("#mobileError").classList.add("hidden");
    });
  }

  function wireBackdrop() {
    const view = root.querySelector(".mobile-view");
    if (!view) return;
    view.addEventListener("click", (event) => {
      if (showMenu && !event.target.closest(".mobile-menu-wrap")) closeMenu();
    });
  }

  function closeMenu() {
    showMenu = false;
    const menu = root.querySelector("#mobileMenu");
    if (menu) menu.classList.add("hidden");
    const backdrop = root.querySelector(".mobile-menu-backdrop");
    if (backdrop) backdrop.remove();
  }

  function wireSearch() {
    const input = root.querySelector("#mobileSearchInput");
    const clear = root.querySelector("#searchClear");

    input.addEventListener("input", () => {
      searchQuery = input.value;
      clear.classList.toggle("hidden", !searchQuery);
      queueSearch();
    });

    clear.addEventListener("click", () => {
      input.value = "";
      searchQuery = "";
      clear.classList.add("hidden");
      queueSearch();
    });

    root.querySelectorAll("[data-chip]").forEach((chip) => {
      const checkbox = chip.querySelector("input[type=checkbox]");
      chip.addEventListener("click", (event) => {
        if (event.target === checkbox) return;
        event.preventDefault();
        checkbox.checked = !checkbox.checked;
        searchFields[chip.dataset.chip] = checkbox.checked;
        chip.classList.toggle("active", checkbox.checked);
        queueSearch();
      });
      checkbox.addEventListener("change", () => {
        searchFields[chip.dataset.chip] = checkbox.checked;
        chip.classList.toggle("active", checkbox.checked);
        queueSearch();
      });
    });
  }

  function toggleSearch() {
    const panel = root.querySelector("#mobileSearch");
    const input = root.querySelector("#mobileSearchInput");
    const clear = root.querySelector("#searchClear");
    panel.classList.toggle("hidden", !showSearch);
    if (showSearch) {
      input.value = searchQuery;
      clear.classList.toggle("hidden", !searchQuery);
      requestAnimationFrame(() => input.focus());
    } else {
      searchQuery = "";
      input.value = "";
      clear.classList.add("hidden");
      queueSearch();
    }
  }

  function queueSearch() {
    if (searchTimer) clearTimeout(searchTimer);
    searchTimer = setTimeout(() => loadData(), 300);
  }

  function selectedFields() {
    return Object.keys(searchFields).filter((key) => searchFields[key]);
  }

  function showError(message) {
    const box = root.querySelector("#mobileError");
    if (!box) return;
    box.querySelector("span").textContent = message;
    box.classList.remove("hidden");
  }

  async function loadData({ append = false } = {}) {
    if (!authenticated) return;
    if (append && currentPage >= totalPages) return;

    const list = root.querySelector("#mobileList");
    const savedScroll = list ? list.scrollTop : 0;
    const errorBox = root.querySelector("#mobileError");
    if (errorBox) errorBox.classList.add("hidden");

    if (append) {
      if (list && !list.querySelector(".list-loading")) {
        const loader = document.createElement("div");
        loader.className = "list-loading";
        loader.textContent = "Loading more...";
        list.appendChild(loader);
      }
    } else {
      currentPage = 1;
      if (list) list.innerHTML = '<div class="mobile-loading"><div class="spinner"></div><span>Loading applications...</span></div>';
    }

    try {
      const page = append ? currentPage + 1 : 1;
      const response = await api.getApplications(localPassword, searchQuery, selectedFields(), page, pageSize);

      if (response && response.success) {
        const incoming = response.data || [];
        applications = append ? applications.concat(incoming) : incoming;
        total = response.total || 0;
        totalPages = response.totalPages || 1;
        currentPage = response.page || page;
        pageSize = response.pageSize || pageSize;
        renderList(savedScroll, append);
        const count = root.querySelector("#mobileCount");
        if (count) count.textContent = String(total);
      } else if (!append) {
        applications = [];
        renderList(0, false);
      }
    } catch (error) {
      if (error.status === 401) {
        showLogin();
        return;
      }
      showError("Failed to fetch applications.");
      if (!append) {
        applications = [];
        renderList(0, false);
      }
    }
  }

  function renderList(savedScroll = 0, append = false) {
    const list = root.querySelector("#mobileList");
    if (!list) return;

    if (applications.length === 0) {
      list.innerHTML = `
        <div class="mobile-empty">
          <div class="mobile-empty-icon">${FILE_ICON}</div>
          <p>No applications found</p>
          <p>Try adjusting your search filters</p>
        </div>
      `;
      return;
    }

    list.innerHTML =
      applications.map(cardHtml).join("") +
      (applications.length >= total
        ? '<div class="list-end">All applications loaded</div>'
        : '<div class="list-end hidden" id="mobileListEnd"></div>');

    requestAnimationFrame(() => {
      list.scrollTop = append ? savedScroll : 0;
    });
  }

  function cardHtml(app) {
    const status = app.acceptanceStatus;
    const borderClass = !app.isReviewed ? "unreviewed" : status === "accepted" ? "accepted" : status === "rejected" ? "rejected" : "pending";
    const badge = !app.isReviewed
      ? '<span class="badge new">New</span>'
      : `<span class="badge ${escapeHtml(status)}">${escapeHtml(status)}</span>`;

    const favoriteKey = `${app.id}-favoriteAboutMinecraft`;
    const understandingKey = `${app.id}-understandingOfSMP`;
    const favoriteExpanded = expanded.has(favoriteKey);
    const understandingExpanded = expanded.has(understandingKey);

    let actions = "";
    if (app.isReviewed && (status === "accepted" || status === "rejected")) {
      actions = `
        <button type="button" class="card-btn delete" data-action="delete" data-id="${app.id}" ${isDeleting === app.id ? "disabled" : ""}>${isDeleting === app.id ? "..." : "Delete"}</button>
        <button type="button" class="card-btn neutral" data-action="unreview" data-id="${app.id}">Unreview</button>`;
    } else if (app.isReviewed && status === "pending") {
      actions = `
        <button type="button" class="card-btn accept" data-action="accept" data-id="${app.id}">Accept</button>
        <button type="button" class="card-btn reject" data-action="reject" data-id="${app.id}">Reject</button>
        <button type="button" class="card-btn delete span-2" data-action="delete" data-id="${app.id}">Delete</button>
        ${app.reviewNotes ? `<button type="button" class="card-btn neutral span-2" data-action="notes" data-id="${app.id}">Notes</button>` : ""}`;
    } else {
      actions = `
        <button type="button" class="card-btn accept" data-action="accept" data-id="${app.id}">Accept</button>
        <button type="button" class="card-btn pending" data-action="pending" data-id="${app.id}">Pending</button>
        <button type="button" class="card-btn reject" data-action="reject" data-id="${app.id}">Reject</button>
        <button type="button" class="card-btn delete" data-action="delete" data-id="${app.id}">Delete</button>`;
    }

    return `
      <div class="app-card ${borderClass}">
        <div class="card-head">
          <div class="card-id-row">
            <span class="card-id">#${app.id}</span>
            ${badge}
          </div>
          <span class="card-date">${escapeHtml(formatDate(app.submissionDate))}</span>
        </div>

        <div class="card-player">
          <div class="player-avatar">${escapeHtml((app.minecraftUsername || "?").charAt(0).toUpperCase())}</div>
          <div class="player-info">
            <div class="player-name-row copyable" data-copy="${escapeHtml(app.minecraftUsername)}">
              <h3 class="player-name">${escapeHtml(app.minecraftUsername)}</h3>
              ${COPY_ICON}
            </div>
            <div class="player-meta">
              <span>${escapeHtml(app.discordUsername)}</span>
              <span>•</span>
              <span class="${getAgeClass(app.age)}">${app.age} y/o</span>
            </div>
          </div>
        </div>

        <div class="card-qa">
          <div>
            <span class="card-qa-label">Favorite Thing</span>
            <p class="card-qa-text ${favoriteExpanded ? "expanded" : ""}" data-expand="${favoriteKey}">${escapeHtml(
              app.favoriteAboutMinecraft
            )}</p>
          </div>
          <div>
            <span class="card-qa-label">SMP Understanding</span>
            <p class="card-qa-text ${understandingExpanded ? "expanded" : ""}" data-expand="${understandingKey}">${escapeHtml(
              app.understandingOfSMP
            )}</p>
          </div>
        </div>

        <div class="card-actions">${actions}</div>
      </div>
    `;
  }

  let isProcessing = null;
  let isDeleting = null;

  function wireList() {
    const list = root.querySelector("#mobileList");

    list.addEventListener("click", async (event) => {
      const expand = event.target.closest("[data-expand]");
      if (expand) {
        const key = expand.dataset.expand;
        if (expanded.has(key)) expanded.delete(key);
        else expanded.add(key);
        expand.classList.toggle("expanded", expanded.has(key));
        return;
      }

      const copyTarget = event.target.closest("[data-copy]");
      if (copyTarget) {
        copyToClipboard(copyTarget.dataset.copy, () => {
          copyTarget.classList.add("copied");
          if (copiedTimer) clearTimeout(copiedTimer);
          copiedTimer = setTimeout(() => copyTarget.classList.remove("copied"), 2000);
        });
        return;
      }

      const button = event.target.closest("[data-action]");
      if (!button) return;

      const id = Number(button.dataset.id);
      const action = button.dataset.action;
      const app = applications.find((item) => item.id === id);
      if (!app) return;

      if (action === "notes") {
        await notesViewer(app);
        return;
      }

      const status = action === "accept" ? "accepted" : action === "reject" ? "rejected" : action === "pending" ? "pending" : null;

      if (status && event.shiftKey) {
        await openNotes(app, status);
        return;
      }

      if (status) {
        await quickReview(app, status);
      } else if (action === "delete") {
        await requestDelete(app.id);
      } else if (action === "unreview") {
        await requestUnreview(app.id);
      }
    });

    list.addEventListener("scroll", () => {
      if (list.scrollTop + list.clientHeight >= list.scrollHeight - 300) {
        loadData({ append: true });
      }
    });
  }

  async function quickReview(app, status) {
    if (isProcessing) return;
    isProcessing = app.id;
    const previousStatus = app.isReviewed ? app.acceptanceStatus : "unreviewed";
    const previousNotes = app.reviewNotes || "";

    try {
      const response = await api.reviewApplication(app.id, localPassword, "", status);
      if (response.success) {
        lastAction = { appId: app.id, previousStatus, previousNotes, timestamp: Date.now() };
        toast.success(`Application marked as ${status}`, () => handleUndo());
        await loadData();
      }
    } catch (error) {
      toast.error(`Failed to mark as ${status}`);
    } finally {
      isProcessing = null;
    }
  }

  async function openNotes(app, status) {
    const result = await notesEditor({ application: app, status });
    if (!result) return;
    if (isProcessing) return;
    isProcessing = app.id;

    const previousStatus = app.isReviewed ? app.acceptanceStatus : "unreviewed";
    const previousNotes = app.reviewNotes || "";

    try {
      const response = await api.reviewApplication(app.id, localPassword, result.notes, result.status);
      if (response.success) {
        lastAction = { appId: app.id, previousStatus, previousNotes, timestamp: Date.now() };
        toast.success(`Application marked as ${result.status}`, () => handleUndo());
        await loadData();
      }
    } catch (error) {
      toast.error("Failed to save review");
    } finally {
      isProcessing = null;
    }
  }

  async function handleUndo() {
    if (!lastAction) return;
    const { appId, previousStatus, previousNotes } = lastAction;
    isProcessing = appId;

    try {
      if (previousStatus === "unreviewed") {
        await api.unreviewApplication(appId, localPassword);
        toast.success("Action undone - application unreviewed");
      } else {
        await api.reviewApplication(appId, localPassword, previousNotes, previousStatus);
        toast.success(`Action undone - reverted to ${previousStatus}`);
      }
      lastAction = null;
      await loadData();
    } catch (error) {
      toast.error("Failed to undo action");
    } finally {
      isProcessing = null;
    }
  }

  async function requestUnreview(id) {
    const confirmed = await confirmDialog({
      title: "Remove Review Status",
      message: "Are you sure you want to remove the reviewed status?",
      confirmText: "Remove",
      cancelText: "Keep",
    });
    if (!confirmed) return;

    isProcessing = id;
    try {
      await api.unreviewApplication(id, localPassword);
      toast.success("Application unreviewed");
      await loadData();
    } catch (error) {
      toast.error("Failed to unreview application");
    } finally {
      isProcessing = null;
    }
  }

  async function requestDelete(id) {
    const confirmed = await confirmDialog({
      title: "Delete Application",
      message: "Are you sure you want to delete this application? This action cannot be undone.",
      confirmText: "Delete",
      cancelText: "Cancel",
    });
    if (!confirmed) return;

    isDeleting = id;
    try {
      await api.deleteApplication(id, localPassword);
      toast.success("Application deleted");
      await loadData();
    } catch (error) {
      toast.error("Failed to delete application");
    } finally {
      isDeleting = null;
    }
  }

  async function exportCsv() {
    try {
      await api.exportApplications(localPassword);
      toast.success("CSV export started");
    } catch (error) {
      toast.error("Failed to export applications");
    }
  }

  async function handleLogoutRequest() {
    const confirmed = await confirmDialog({
      title: "Logout",
      message: "Are you sure you want to logout?",
      confirmText: "Logout",
      cancelText: "Stay",
    });
    if (!confirmed) return;

    await api.logout();
    localPassword = "";
    applications = [];
    lastAction = null;
    showLogin();
  }

  function onKeydown(event) {
    const isCommand = event.metaKey || event.ctrlKey;
    if (isCommand && (event.key === "z" || event.key === "Z")) {
      if (!authenticated || !lastAction) return;
      event.preventDefault();
      if (Date.now() - lastAction.timestamp < UNDO_TIMEOUT) handleUndo();
    }
  }

  window.addEventListener("keydown", onKeydown);

  showLogin();

  api.validateToken().then((result) => {
    if (result && result.valid) {
      canManageAdmins = !!(result.admin && result.admin.canManageAdmins);
      showDashboard();
    }
  });

  return cleanup;
}
