let onRouteChange = null;

export function navigate(path, { replace = false } = {}) {
  if (path === `${location.pathname}${location.search}`) return;
  if (replace) history.replaceState(null, "", path);
  else history.pushState(null, "", path);
  if (onRouteChange) onRouteChange(location.pathname);
}

export function initRouter(handler) {
  onRouteChange = handler;

  document.addEventListener("click", (event) => {
    if (event.defaultPrevented || event.button !== 0) return;
    if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
    const link = event.target.closest("a[data-link]");
    if (!link) return;
    const href = link.getAttribute("href");
    if (!href || href.startsWith("http")) return;
    event.preventDefault();
    navigate(href);
  });

  window.addEventListener("popstate", () => {
    if (onRouteChange) onRouteChange(location.pathname);
  });
}
