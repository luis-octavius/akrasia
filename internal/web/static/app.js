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

// focus the dialog's text as soon as it opens: the notes, or the name of a
// task made from a dump
document.addEventListener("htmx:afterSwap", (e) => {
  if (e.detail.target.id === "overlay") {
    e.detail.target.querySelector("textarea, input[name=name]")?.focus();
  }
});

if ("serviceWorker" in navigator) {
  navigator.serviceWorker.register("/sw.js");
}
