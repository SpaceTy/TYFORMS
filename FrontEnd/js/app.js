import { initRouter } from "./router.js";
import applicationView from "./views/application.js";
import thankYouView from "./views/thankyou.js";
import adminView from "./views/admin.js";
import mobileAdminView from "./views/mobile-admin.js";
import statsView from "./views/stats.js";
import accountsView from "./views/accounts.js";

const views = {
  "/": applicationView,
  "/thank-you": thankYouView,
  "/admin": adminView,
  "/admin/mobile": mobileAdminView,
  "/admin/stats": statsView,
  "/admin/accounts": accountsView,
};

const app = document.getElementById("app");
app.className = "app";
app.innerHTML = `
  <header class="site-header" id="siteHeader">
    <h1 class="site-title">TYSMP</h1>
    <p class="site-subtitle">Private Vanilla MC SMP Application</p>
  </header>
  <main class="app-main" id="viewRoot"></main>
  <footer class="site-footer" id="siteFooter">
    &copy; ${new Date().getFullYear()} TYSMP · All rights reserved
  </footer>
`;

const header = document.getElementById("siteHeader");
const footer = document.getElementById("siteFooter");
const viewRoot = document.getElementById("viewRoot");

let cleanup = null;
let renderToken = 0;

async function renderRoute() {
  const token = ++renderToken;
  let path = location.pathname;

  if (path.length > 1 && path.endsWith("/")) path = path.slice(0, -1);
  if (path === "/admin" && window.innerWidth <= 768) {
    history.replaceState(null, "", "/admin/mobile");
    path = "/admin/mobile";
  }

  if (cleanup) {
    try {
      cleanup();
    } catch (error) {
      console.error(error);
    }
    cleanup = null;
  }

  const isAdmin = path.startsWith("/admin");
  app.classList.toggle("app--admin", isAdmin);
  header.classList.toggle("hidden", isAdmin);
  footer.classList.toggle("hidden", isAdmin);

  viewRoot.innerHTML = "";
  viewRoot.className = "app-main";

  if (!isAdmin) window.scrollTo(0, 0);

  const view = views[path] || views["/"];
  const result = await view(viewRoot);

  if (token !== renderToken) {
    if (typeof result === "function") result();
    return;
  }

  cleanup = typeof result === "function" ? result : null;
}

initRouter(renderRoute);
renderRoute();
