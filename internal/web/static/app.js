// Everything else is htmx; this only closes the completion dialog and
// registers the service worker that makes the page installable.
document.addEventListener("click", (e) => {
  if (e.target.closest("[data-close]") || e.target.classList.contains("dialog")) {
    document.getElementById("overlay").replaceChildren();
  }
});

document.addEventListener("keydown", (e) => {
  if (e.key === "Escape") {
    document.getElementById("overlay")?.replaceChildren();
  }
});

// focus the notes as soon as the dialog opens
document.addEventListener("htmx:afterSwap", (e) => {
  if (e.detail.target.id === "overlay") {
    e.detail.target.querySelector("textarea")?.focus();
  }
});

if ("serviceWorker" in navigator) {
  navigator.serviceWorker.register("/sw.js");
}
