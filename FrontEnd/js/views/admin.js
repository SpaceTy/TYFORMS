import api from "../api.js";
import { createNavbar } from "../components.js";
import {
  confirmDialog,
  copyToClipboard,
  escapeHtml,
  formatDate,
  getAgeClass,
  notesEditor,
  toast,
} from "../ui.js";

const UNDO_TIMEOUT = 10000;

const COLUMNS = [
  { label: "#", width: "5rem" },
  { label: "DC", width: "8rem" },
  { label: "MC", width: "8rem" },
  { label: "Age", width: "3.5rem", center: true },
  { label: "FAM", width: "23%" },
  { label: "SU", width: "23%" },
  { label: "S?", width: "5rem", center: true },
  { label: "Date", width: "9rem" },
  { label: "Actions", width: "8rem", center: true },
];

export default function render(root) {
  let authenticated = false;
  let localPassword = "";
  let isProcessing = null;
  let isDeleting = null;
  let isLoading = false;
  let isLoadingMore = false;
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
  let currentPage = 1;
  let pageSize = 50;
  let total = 0;
  let totalPages = 1;
  let lastAction = null;
  let navbarCleanup = null;
  let tooltipEl = null;
  let tooltipContainer = null;
  let scrollEl = null;
  let searchTimer = null;
  let copiedTimer = null;
  let newPollTimer = null;
  let lastServerTime = null;
  let firstPoll = true;

  function cleanup() {
    if (navbarCleanup) navbarCleanup();
    navbarCleanup = null;
    window.removeEventListener("keydown", onKeydown);
    if (searchTimer) clearTimeout(searchTimer);
    if (copiedTimer) clearTimeout(copiedTimer);
    if (newPollTimer) clearInterval(newPollTimer);
    detachTooltip();
    detachScroll();
  }

  function showLogin() {
    authenticated = false;
    if (navbarCleanup) {
      navbarCleanup();
      navbarCleanup = null;
    }
    detachTooltip();
    detachScroll();

    root.innerHTML = `
      <div class="admin-view">
        <div class="admin-login mc-container fade-in" id="loginBox">
          <h2 class="mc-title text-center">Admin Login</h2>
          <form id="loginForm">
            <div class="form-group">
              <label class="mc-label" for="adminUsername">Username</label>
              <input class="mc-input" id="adminUsername" type="text" placeholder="Username (leave empty for admin password)" autocomplete="username" value="${escapeHtml(api.getUsername())}">
            </div>
            <div class="form-group">
              <label class="mc-label" for="adminPassword">Password</label>
              <input class="mc-input" id="adminPassword" type="password" placeholder="Enter admin password" required autocomplete="current-password">
            </div>
            <div class="form-error hidden" id="authError"></div>
            <div class="text-center">
              <button type="submit" class="mc-button">Login</button>
            </div>
          </form>
        </div>
      </div>
    `;

    const form = root.querySelector("#loginForm");
    const passwordInput = root.querySelector("#adminPassword");
    const errorBox = root.querySelector("#authError");
    const loginBox = root.querySelector("#loginBox");

    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const username = root.querySelector("#adminUsername").value.trim();
      const password = passwordInput.value;

      try {
        const response = await api.login(username, password);
        if (response.success) {
          localPassword = "";
          searchQuery = "";
          applications = [];
          currentPage = 1;
          await showDashboard();
        } else {
          errorBox.textContent = "Invalid username or password";
          errorBox.classList.remove("hidden");
          loginBox.classList.remove("login-shake");
          void loginBox.offsetWidth;
          loginBox.classList.add("login-shake");
        }
      } catch (error) {
        errorBox.textContent =
          error.status === 401 ? "Invalid username or password" : "Authentication error. Please try again.";
        errorBox.classList.remove("hidden");
        loginBox.classList.remove("login-shake");
        void loginBox.offsetWidth;
        loginBox.classList.add("login-shake");
      }
    });

    passwordInput.focus();
  }

  async function showDashboard() {
    authenticated = true;

    root.innerHTML = `
      <div class="admin-dashboard">
        <div class="admin-panel mc-panel">
          <div id="navbarSlot"></div>
          <div class="admin-content">
            <div class="admin-error hidden" id="adminError"></div>
            <div class="empty-state hidden" id="emptyState">
              <div class="empty-state-inner glass">
                <p>No applications yet</p>
                <p>Applications will appear here once submitted</p>
              </div>
            </div>
            <div class="table-scroll hidden" id="tableScroll">
              <div class="table-frame">
                <table class="app-table">
                  <colgroup>
                    ${COLUMNS.map((column) => `<col style="width:${column.width}">`).join("")}
                  </colgroup>
                  <thead>
                    <tr>
                      ${COLUMNS.map(
                        (column) => `<th class="${column.center ? "center" : ""}">${column.label}</th>`
                      ).join("")}
                    </tr>
                  </thead>
                  <tbody id="tableBody"></tbody>
                </table>
              </div>
              <div class="list-loading hidden" id="loadingMore">Loading more applications...</div>
              <div class="list-end hidden" id="listEnd"></div>
            </div>
          </div>
        </div>
      </div>

      <div class="search-panel" id="searchPanel">
        <div class="search-panel-inner">
          <div class="search-head">
            <h3 class="mc-title" style="font-size:1rem; margin:0;">Search</h3>
            <button type="button" class="search-close" id="searchClose">✕</button>
          </div>
          <input class="mc-input" id="searchInput" type="text" placeholder="Type to search... (Esc to close)">
          <div class="search-fields">
            <label><input type="checkbox" data-field="discordUsername" checked> <span>Discord</span></label>
            <label><input type="checkbox" data-field="minecraftUsername" checked> <span>Minecraft</span></label>
            <label><input type="checkbox" data-field="favoriteAboutMinecraft"> <span>FAM</span></label>
            <label><input type="checkbox" data-field="understandingOfSMP"> <span>SU</span></label>
            <label><input type="checkbox" data-field="id"> <span>#</span></label>
          </div>
        </div>
      </div>
    `;

    const navbar = await createNavbar({ onLogout: handleLogoutRequest });
    navbarCleanup = navbar.cleanup;
    root.querySelector("#navbarSlot").replaceWith(navbar.el);

    const refreshButton = document.createElement("button");
    refreshButton.type = "button";
    refreshButton.className = "mc-button refresh-button";
    refreshButton.innerHTML = '<svg class="refresh-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M18.36 18.36a9 9 0 1 1 0-12.72L21 9"/><path d="M21 3v6h-6"/></svg>';
    refreshButton.title = "Refresh";
    refreshButton.setAttribute("aria-label", "Refresh");
    refreshButton.addEventListener("click", () => loadData({ animate: true }));

    const exportButton = document.createElement("button");
    exportButton.type = "button";
    exportButton.className = "mc-button small secondary";
    exportButton.textContent = "Export CSV";
    exportButton.addEventListener("click", handleExport);

    navbar.actions.appendChild(refreshButton);
    navbar.actions.appendChild(exportButton);

    refreshButton.dataset.role = "refresh";

    attachTooltip();
    attachScroll();
    wireSearchPanel();
    wireTable();

    await loadData();
    startNewApplicationPolling();
  }

  function startNewApplicationPolling() {
    if (newPollTimer) clearInterval(newPollTimer);
    firstPoll = true;
    newPollTimer = setInterval(checkForNewApplications, 5000);
  }

  async function checkForNewApplications() {
    if (!authenticated) return;
    try {
      const result = await api.hasNewApplications(lastServerTime);
      if (result && typeof result.now === "string") lastServerTime = result.now;
      if (firstPoll) {
        firstPoll = false;
        return;
      }
      if (result && result.new) {
        toast.success("New application received");
        if (!searchQuery.trim()) {
          await loadData({ animate: false });
        }
      }
    } catch (error) {
      if (error.status === 401) showLogin();
    }
  }

  function wireSearchPanel() {
    const panel = root.querySelector("#searchPanel");
    const input = root.querySelector("#searchInput");
    if (!panel || !input) return;

    input.value = searchQuery;
    panel.querySelectorAll("[data-field]").forEach((checkbox) => {
      checkbox.checked = !!searchFields[checkbox.dataset.field];
      checkbox.addEventListener("change", () => {
        searchFields[checkbox.dataset.field] = checkbox.checked;
        queueSearch();
      });
    });

    input.addEventListener("input", () => {
      searchQuery = input.value;
      queueSearch();
    });

    panel.querySelector("#searchClose").addEventListener("click", closeSearch);
  }

  function queueSearch() {
    if (searchTimer) clearTimeout(searchTimer);
    searchTimer = setTimeout(() => loadData(), 300);
  }

  function openSearch() {
    if (!authenticated) return;
    showSearch = true;
    const panel = root.querySelector("#searchPanel");
    if (!panel) return;
    panel.classList.add("open");
    const input = panel.querySelector("#searchInput");
    if (input) {
      input.focus();
      input.select();
    }
  }

  function closeSearch() {
    showSearch = false;
    const panel = root.querySelector("#searchPanel");
    if (panel) panel.classList.remove("open");
  }

  function onKeydown(event) {
    const isCommand = event.metaKey || event.ctrlKey;
    if (isCommand && (event.key === "f" || event.key === "F")) {
      if (!authenticated) return;
      event.preventDefault();
      openSearch();
      return;
    }
    if (event.key === "Escape" && showSearch) {
      event.preventDefault();
      closeSearch();
      return;
    }
    if (isCommand && (event.key === "z" || event.key === "Z")) {
      if (!authenticated || !lastAction) return;
      event.preventDefault();
      if (Date.now() - lastAction.timestamp < UNDO_TIMEOUT) handleUndo();
    }
  }

  window.addEventListener("keydown", onKeydown);

  function selectedFields() {
    return Object.keys(searchFields).filter((key) => searchFields[key]);
  }

  function setRefreshLoading(loading) {
    const button = root.querySelector('[data-role="refresh"]');
    if (!button) return;
    button.disabled = loading;
    button.setAttribute("aria-busy", String(loading));
    const icon = button.querySelector(".refresh-icon");
    if (icon) icon.classList.toggle("spin", loading);
  }

  async function loadData({ append = false, animate = false } = {}) {
    if (!authenticated) return;
    if (append && (isLoadingMore || currentPage >= totalPages)) return;

    const tableScroll = root.querySelector("#tableScroll");
    const savedScroll = tableScroll ? tableScroll.scrollTop : 0;
    const errorBox = root.querySelector("#adminError");

    if (errorBox) errorBox.classList.add("hidden");

    if (append) {
      isLoadingMore = true;
      const loadingMore = root.querySelector("#loadingMore");
      if (loadingMore) loadingMore.classList.remove("hidden");
    } else {
      if (isLoading) return;
      isLoading = true;
      setRefreshLoading(true);
      currentPage = 1;
    }

    const refreshStarted = performance.now();
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
        renderTable(savedScroll, append);
      } else if (!append) {
        applications = [];
        total = 0;
        totalPages = 1;
        renderTable(0, false);
        showError("Failed to load applications. Please try again.");
      }
    } catch (error) {
      if (error.status === 401) {
        showLogin();
        return;
      }
      if (!append) showError("Failed to load applications. Please try again.");
    } finally {
      if (animate) {
        await new Promise((resolve) => setTimeout(resolve, Math.max(0, 700 - (performance.now() - refreshStarted))));
      }
      isLoading = false;
      isLoadingMore = false;
      setRefreshLoading(false);
      const loadingMore = root.querySelector("#loadingMore");
      if (loadingMore) loadingMore.classList.add("hidden");
    }
  }

  function showError(message) {
    const errorBox = root.querySelector("#adminError");
    if (!errorBox) return;
    errorBox.textContent = message;
    errorBox.classList.remove("hidden");
  }

  function renderTable(savedScroll = 0, append = false) {
    const tbody = root.querySelector("#tableBody");
    const tableScroll = root.querySelector("#tableScroll");
    const emptyState = root.querySelector("#emptyState");
    const listEnd = root.querySelector("#listEnd");
    if (!tbody || !tableScroll || !emptyState) return;

    tbody.innerHTML = applications.map(rowHtml).join("");

    if (applications.length === 0) {
      tableScroll.classList.add("hidden");
      emptyState.classList.remove("hidden");
    } else {
      emptyState.classList.add("hidden");
      tableScroll.classList.remove("hidden");
    }

    if (listEnd) {
      if (applications.length > 0 && applications.length >= total) {
        listEnd.textContent = `Showing all ${total} applications${searchQuery.trim() ? " (filtered)" : ""}`;
        listEnd.classList.remove("hidden");
      } else {
        listEnd.classList.add("hidden");
      }
    }

    requestAnimationFrame(() => {
      if (tableScroll) tableScroll.scrollTop = append ? savedScroll : 0;
    });
  }

  function rowHtml(app) {
    const reviewed = app.isReviewed;
    const status = app.acceptanceStatus;
    const rowClasses = reviewed ? "" : "row-unreviewed";
    const textClass = reviewed ? "" : "cell-unreviewed";
    const notes = escapeHtml(app.favoriteAboutMinecraft || "");
    const understanding = escapeHtml(app.understandingOfSMP || "");
    const discord = escapeHtml(app.discordUsername || "");
    const minecraft = escapeHtml(app.minecraftUsername || "");

    let actions = "";
    if (reviewed && (status === "accepted" || status === "rejected")) {
      actions = `
        <div class="actions-grid single-col">
          <button type="button" class="action-btn remove" data-action="delete" data-id="${app.id}" title="Delete">🗑</button>
          <button type="button" class="action-btn neutral" data-action="unreview" data-id="${app.id}" title="Unreview">↶</button>
        </div>`;
    } else if (reviewed && status === "pending") {
      actions = `
        <div class="actions-grid">
          <button type="button" class="action-btn accept" data-action="accept" data-id="${app.id}" title="Accept">✓</button>
          <button type="button" class="action-btn reject" data-action="reject" data-id="${app.id}" title="Reject">✕</button>
          <button type="button" class="action-btn remove span-2" data-action="delete" data-id="${app.id}" title="Delete">🗑</button>
        </div>`;
    } else {
      actions = `
        <div class="actions-grid">
          <button type="button" class="action-btn accept" data-action="accept" data-id="${app.id}" title="Accept">✓</button>
          <button type="button" class="action-btn pending" data-action="pending" data-id="${app.id}" title="Pending">•</button>
          <button type="button" class="action-btn reject" data-action="reject" data-id="${app.id}" title="Reject">✕</button>
          <button type="button" class="action-btn remove" data-action="delete" data-id="${app.id}" title="Delete">🗑</button>
        </div>`;
    }

    const statusPill =
      reviewed && status
        ? `<span class="status-pill ${escapeHtml(status)}">${
            status === "accepted" ? "Accepted" : status === "rejected" ? "Rejected" : "Pending"
          }</span>`
        : "";

    return `
      <tr class="${rowClasses}">
        <td class="${textClass}">
          <div class="cell-id">
            ${reviewed ? "" : '<span class="review-star">⭐</span>'}
            <span style="margin-left:auto;">${app.id}</span>
          </div>
        </td>
        <td class="${textClass}">
          <div class="tooltip-container" data-tooltip="${discord}"><div class="truncate">${discord}</div></div>
        </td>
        <td class="${textClass}">
          <div class="tooltip-container" data-tooltip="${minecraft}">
            <div class="truncate copyable" data-copy="${minecraft}">${minecraft}</div>
          </div>
        </td>
        <td class="center ${getAgeClass(app.age)}" style="font-weight:700;">${app.age}</td>
        <td class="${textClass}">
          <div class="tooltip-container" data-tooltip="${notes}"><div class="truncate">${notes}</div></div>
        </td>
        <td class="${textClass}">
          <div class="tooltip-container" data-tooltip="${understanding}"><div class="truncate">${understanding}</div></div>
        </td>
        <td class="center">
          <span class="status-pill" style="background:${app.joinedDiscord ? "var(--green-600)" : "var(--red-600)"}; color:#fff;">${
            app.joinedDiscord ? "Yes" : "No"
          }</span>
        </td>
        <td class="${textClass} date-cell">
          <div class="tooltip-container" data-tooltip="${escapeHtml(formatDate(app.submissionDate, true))}">
            <div class="truncate">${escapeHtml(formatDate(app.submissionDate))}</div>
          </div>
          ${reviewed ? `<div class="reviewed">Reviewed: ${escapeHtml(formatDate(app.reviewedAt))}</div>` : ""}
          ${statusPill}
        </td>
        <td class="center">${actions}</td>
      </tr>
    `;
  }

  function wireTable() {
    const tbody = root.querySelector("#tableBody");
    if (!tbody) return;

    tbody.addEventListener("click", async (event) => {
      const copyTarget = event.target.closest("[data-copy]");
      if (copyTarget) {
        const target = copyTarget;
        copyToClipboard(target.dataset.copy, () => {
          target.classList.add("copied");
          if (copiedTimer) clearTimeout(copiedTimer);
          copiedTimer = setTimeout(() => target.classList.remove("copied"), 2000);
        });
        return;
      }

      const button = event.target.closest("[data-action]");
      if (!button) return;

      const id = Number(button.dataset.id);
      const action = button.dataset.action;
      const app = applications.find((item) => item.id === id);
      if (!app) return;

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
  }

  function markBusy(id) {
    root.querySelectorAll(`[data-id="${id}"]`).forEach((button) => {
      button.disabled = true;
    });
  }

  async function quickReview(app, status) {
    if (isProcessing) return;
    isProcessing = app.id;
    markBusy(app.id);

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

  async function handleExport() {
    const confirmed = await confirmDialog({
      title: "Export Data",
      message: "This will export all application data to a CSV file. Continue?",
      confirmText: "Export",
      cancelText: "Cancel",
    });
    if (!confirmed) return;

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

  function attachScroll() {
    scrollEl = root.querySelector("#tableScroll");
    if (!scrollEl) return;
    scrollEl.addEventListener("scroll", onScroll);
  }

  function detachScroll() {
    if (scrollEl) {
      scrollEl.removeEventListener("scroll", onScroll);
      scrollEl = null;
    }
  }

  function onScroll() {
    if (!scrollEl) return;
    const remaining = scrollEl.scrollHeight - scrollEl.scrollTop - scrollEl.clientHeight;
    if (remaining < 300) loadData({ append: true });
  }

  function attachTooltip() {
    tooltipEl = document.getElementById("global-tooltip");
    if (tooltipEl) tooltipEl.remove();

    tooltipEl = document.createElement("div");
    tooltipEl.id = "global-tooltip";
    tooltipEl.className = "tooltip";
    (document.getElementById("tooltip-root") || document.body).appendChild(tooltipEl);

    const scroll = root.querySelector("#tableScroll");
    if (!scroll) return;
    scroll.addEventListener("mouseover", onTooltipOver);
    scroll.addEventListener("mouseout", onTooltipOut);
    scroll.addEventListener("mousemove", onTooltipMove);
    scroll.addEventListener("scroll", hideTooltip);
  }

  function detachTooltip() {
    const scroll = root.querySelector("#tableScroll");
    if (scroll) {
      scroll.removeEventListener("mouseover", onTooltipOver);
      scroll.removeEventListener("mouseout", onTooltipOut);
      scroll.removeEventListener("mousemove", onTooltipMove);
      scroll.removeEventListener("scroll", hideTooltip);
    }
    if (tooltipEl) {
      tooltipEl.remove();
      tooltipEl = null;
    }
    tooltipContainer = null;
  }

  function onTooltipOver(event) {
    const container = event.target.closest(".tooltip-container");
    if (!container) {
      hideTooltip();
      tooltipContainer = null;
      return;
    }
    if (container === tooltipContainer) return;
    tooltipContainer = container;
    showTooltip(container);
  }

  function onTooltipOut(event) {
    const next = event.relatedTarget ? event.relatedTarget.closest(".tooltip-container") : null;
    if (next !== tooltipContainer) {
      hideTooltip();
      tooltipContainer = null;
    }
  }

  function onTooltipMove() {
    if (!tooltipContainer) return;
    positionTooltip(tooltipContainer);
  }

  function showTooltip(container) {
    if (!tooltipEl) return;
    const text = container.getAttribute("data-tooltip");
    if (!text || !text.trim()) return;
    tooltipEl.textContent = text;
    positionTooltip(container);
    tooltipEl.classList.add("visible");
  }

  function positionTooltip(container) {
    if (!tooltipEl) return;
    const rect = container.getBoundingClientRect();
    const tooltipHeight = tooltipEl.offsetHeight;
    const tooltipWidth = tooltipEl.offsetWidth;

    let top = rect.top - tooltipHeight - 8;
    if (top < 8) top = rect.bottom + 8;

    let left = rect.left + rect.width / 2 - tooltipWidth / 2;
    left = Math.max(8, Math.min(left, window.innerWidth - tooltipWidth - 8));

    tooltipEl.style.top = `${top}px`;
    tooltipEl.style.left = `${left}px`;
  }

  function hideTooltip() {
    if (tooltipEl) tooltipEl.classList.remove("visible");
  }

  showLogin();

  api.validateToken().then((result) => {
    if (result && result.valid) showDashboard();
  });

  return cleanup;
}
