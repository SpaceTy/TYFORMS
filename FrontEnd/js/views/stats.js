import api from "../api.js";
import { createNavbar } from "../components.js";
import { navigate } from "../router.js";
import { escapeHtml } from "../ui.js";

export default function render(root) {
  let authenticated = false;
  let localPassword = "";
  let navbarCleanup = null;
  let isLoading = false;
  let stats = null;

  function cleanup() {
    if (navbarCleanup) navbarCleanup();
    navbarCleanup = null;
  }

  function showLogin() {
    authenticated = false;
    if (navbarCleanup) {
      navbarCleanup();
      navbarCleanup = null;
    }

    root.innerHTML = `
      <div class="admin-view">
        <div class="admin-login mc-container fade-in">
          <h2 class="mc-title text-center">Admin Statistics</h2>
          <form id="statsLoginForm">
            <div class="form-group">
              <label class="mc-label" for="statsUsername">Username</label>
              <input class="mc-input" id="statsUsername" type="text" placeholder="Username (leave empty for admin password)" autocomplete="username" value="${escapeHtml(
                api.getUsername()
              )}">
            </div>
            <div class="form-group">
              <label class="mc-label" for="statsPassword">Password</label>
              <input class="mc-input" id="statsPassword" type="password" placeholder="Enter admin password" required autocomplete="current-password">
            </div>
            <div class="form-error hidden" id="statsAuthError"></div>
            <div class="text-center">
              <button type="submit" class="mc-button">Login</button>
            </div>
          </form>
        </div>
      </div>
    `;

    const form = root.querySelector("#statsLoginForm");
    const passwordInput = root.querySelector("#statsPassword");
    const errorBox = root.querySelector("#statsAuthError");

    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      const username = root.querySelector("#statsUsername").value.trim();

      try {
        const response = await api.login(username, passwordInput.value);
        if (response.success) {
          localPassword = "";
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
      <div class="admin-dashboard admin-dashboard--flush">
        <div class="admin-panel mc-panel admin-panel--flush">
          <div id="navbarSlot"></div>
          <div class="admin-content admin-content--scroll">
            <div class="admin-error hidden" id="statsError"></div>
            <div id="statsBody" class="muted">Loading statistics...</div>
          </div>
        </div>
      </div>
    `;

    const navbar = await createNavbar({ onLogout: logout });
    navbarCleanup = navbar.cleanup;
    root.querySelector("#navbarSlot").replaceWith(navbar.el);

    const refreshButton = document.createElement("button");
    refreshButton.type = "button";
    refreshButton.className = "mc-button refresh-button";
    refreshButton.innerHTML = '<svg class="refresh-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.75" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M18.36 18.36a9 9 0 1 1 0-12.72L21 9"/><path d="M21 3v6h-6"/></svg>';
    refreshButton.title = "Refresh";
    refreshButton.setAttribute("aria-label", "Refresh");
    refreshButton.addEventListener("click", () => loadStats({ animate: true }));
    navbar.actions.appendChild(refreshButton);
    refreshButton.dataset.role = "refresh";

    await loadStats();
  }

  async function loadStats({ animate = false } = {}) {
    if (!authenticated) return;
    if (isLoading) return;
    isLoading = true;
    const refreshStarted = performance.now();

    const button = root.querySelector('[data-role="refresh"]');
    if (button) {
      button.disabled = true;
      button.setAttribute("aria-busy", "true");
      const icon = button.querySelector(".refresh-icon");
      if (icon) icon.classList.add("spin");
    }
    const errorBox = root.querySelector("#statsError");
    if (errorBox) errorBox.classList.add("hidden");

    try {
      const response = await api.getApplicationStats(localPassword);
      if (response && response.success) {
        stats = response.data;
        renderStats();
      } else {
        showError("Failed to load statistics.");
      }
    } catch (error) {
      if (error.status === 401) {
        showLogin();
        return;
      }
      showError("Failed to load statistics.");
    } finally {
      if (animate) {
        await new Promise((resolve) => setTimeout(resolve, Math.max(0, 700 - (performance.now() - refreshStarted))));
      }
      isLoading = false;
      const current = root.querySelector('[data-role="refresh"]');
      if (current) {
        current.disabled = false;
        current.setAttribute("aria-busy", "false");
        const icon = current.querySelector(".refresh-icon");
        if (icon) icon.classList.remove("spin");
      }
    }
  }

  function showError(message) {
    const errorBox = root.querySelector("#statsError");
    if (!errorBox) return;
    errorBox.textContent = message;
    errorBox.classList.remove("hidden");
  }

  function percentage(value) {
    if (!stats || !stats.totalApplications) return 0;
    return Math.round((value / stats.totalApplications) * 100);
  }

  function recentPercentage(value) {
    const items = stats && stats.recentSubmissions ? stats.recentSubmissions : [];
    const max = items.length ? Math.max(...items.map((item) => item.count)) : 0;
    if (!max) return 0;
    return Math.round((value / max) * 100);
  }

  function formatShortDate(dateString) {
    const date = new Date(`${dateString}T00:00:00Z`);
    return date.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  }

  function renderStats() {
    const body = root.querySelector("#statsBody");
    if (!body || !stats) return;

    const recent = stats.recentSubmissions || [];
    const averageAge = Number(stats.averageAge || 0).toFixed(1);

    body.innerHTML = `
      <div class="stats-grid">
        <div class="stat-card glass">
          <p>Total Applications</p>
          <p>${stats.totalApplications}</p>
        </div>
        <div class="stat-card glass green">
          <p>Reviewed</p>
          <p>${stats.reviewedApplications}</p>
        </div>
        <div class="stat-card glass yellow">
          <p>Unreviewed</p>
          <p>${stats.unreviewedApplications}</p>
        </div>
        <div class="stat-card glass primary">
          <p>Average Age</p>
          <p>${averageAge}</p>
        </div>
      </div>

      <div class="stats-columns">
        <div class="stats-box glass">
          <h3>Review Outcome</h3>
          <div class="bar-row">
            <div class="bar-label"><span style="color:#86efac;">Accepted</span><span>${stats.acceptedApplications}</span></div>
            <div class="bar-track"><div class="bar-fill green" style="width:${percentage(stats.acceptedApplications)}%"></div></div>
          </div>
          <div class="bar-row">
            <div class="bar-label"><span style="color:#fde047;">Pending</span><span>${stats.pendingApplications}</span></div>
            <div class="bar-track"><div class="bar-fill yellow" style="width:${percentage(stats.pendingApplications)}%"></div></div>
          </div>
          <div class="bar-row">
            <div class="bar-label"><span style="color:#fca5a5;">Rejected</span><span>${stats.rejectedApplications}</span></div>
            <div class="bar-track"><div class="bar-fill red" style="width:${percentage(stats.rejectedApplications)}%"></div></div>
          </div>
        </div>

        <div class="stats-box glass">
          <h3>Last 7 Days</h3>
          ${recent
            .map(
              (item) => `
              <div class="recent-row">
                <span class="date">${formatShortDate(item.date)}</span>
                <div class="bar-track"><div class="bar-fill primary" style="width:${recentPercentage(item.count)}%"></div></div>
                <span class="count">${item.count}</span>
              </div>`
            )
            .join("")}
        </div>
      </div>

      <div class="stats-footer glass">
        Joined Discord: <strong style="color:#fff;">${stats.joinedDiscordCount}</strong> / ${stats.totalApplications}
      </div>
    `;
  }

  async function logout() {
    await api.logout();
    localPassword = "";
    stats = null;
    showLogin();
    navigate("/admin", { replace: true });
  }

  showLogin();

  api.validateToken().then((result) => {
    if (result && result.valid) showDashboard();
  });

  return cleanup;
}
