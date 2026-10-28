import api from "./api.js";
import { escapeHtml } from "./ui.js";

export function createProfileMenu(onLogout) {
  const username = api.getUsername() || "admin";
  const initial = username ? username.charAt(0) : "?";

  const wrap = document.createElement("div");
  wrap.className = "profile";
  wrap.innerHTML = `
    <button type="button" class="profile-button">
      <span class="profile-avatar">${escapeHtml(initial)}</span>
      <span class="profile-name">${escapeHtml(username)}</span>
      <svg class="profile-chevron" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 9l-7 7-7-7" />
      </svg>
    </button>
    <div class="profile-menu hidden">
      <div class="profile-menu-head">
        <small>Signed in as</small>
        <strong>${escapeHtml(username)}</strong>
      </div>
      <div class="profile-menu-body">
        <button type="button" class="profile-menu-item danger"><span>➜</span> Logout</button>
      </div>
    </div>
  `;

  const button = wrap.querySelector(".profile-button");
  const menu = wrap.querySelector(".profile-menu");
  const logoutButton = wrap.querySelector(".profile-menu-item.danger");

  const close = () => {
    menu.classList.add("hidden");
    wrap.classList.remove("open");
  };

  const toggle = () => {
    menu.classList.toggle("hidden");
    wrap.classList.toggle("open", !menu.classList.contains("hidden"));
  };

  const onButtonClick = (event) => {
    event.stopPropagation();
    toggle();
  };

  const onDocumentClick = (event) => {
    if (!wrap.contains(event.target)) close();
  };

  button.addEventListener("click", onButtonClick);
  logoutButton.addEventListener("click", () => {
    close();
    if (onLogout) onLogout();
  });
  document.addEventListener("click", onDocumentClick);

  return {
    el: wrap,
    cleanup() {
      document.removeEventListener("click", onDocumentClick);
    },
  };
}

export async function createNavbar({ onLogout } = {}) {
  let canManageAdmins = false;
  try {
    const result = await api.validateToken();
    canManageAdmins = !!(result && result.admin && result.admin.canManageAdmins);
  } catch (error) {
    canManageAdmins = false;
  }

  const links = [
    { label: "Dashboard", href: "/admin" },
    { label: "Statistics", href: "/admin/stats" },
  ];
  if (canManageAdmins) links.push({ label: "Admins", href: "/admin/accounts" });

  const el = document.createElement("div");
  el.className = "admin-navbar";
  el.innerHTML = `
    <nav class="navbar-nav">
      ${links
        .map(
          (link) => `<a href="${link.href}" data-link class="navbar-link ${
            location.pathname === link.href ? "active" : ""
          }">${link.label}</a>`
        )
        .join("")}
    </nav>
    <div class="navbar-right">
      <div class="navbar-actions"></div>
    </div>
  `;

  const profile = createProfileMenu(onLogout);
  el.querySelector(".navbar-right").appendChild(profile.el);

  return {
    el,
    actions: el.querySelector(".navbar-actions"),
    cleanup: profile.cleanup,
  };
}
