// The offline queue of the phone app. When a task, a dump entry or a
// completion can't reach the server, it waits here (in IndexedDB) and is
// sent when the server answers again. Nothing is edited offline.
//
// Every write carries an id (X-Queue-Id) that the server remembers, so a
// request that did arrive but whose answer was lost is not applied twice
// when it is sent again. The server says how a write went in the
// X-Akrasia-Result header.
//
// The first part has no DOM and is tested with node (internal/web/jstest);
// the second one wires it to the page.
(function (root) {
  "use strict";

  const WRITES = ["/add", "/complete"];
  const TABS = ["/today", "/all", "/dump"];
  const RETRY_MS = 30000;

  // ---- logic ---------------------------------------------------------

  const pad = (n) => String(n).padStart(2, "0");
  // the local date of d as yyyy-mm-dd: the phone's day, not UTC's
  const isoDate = (d) => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  const tomorrow = (d) => new Date(d.getFullYear(), d.getMonth(), d.getDate() + 1);

  // makeItem is what is stored for a write that didn't reach the server.
  // It settles what the server would decide when the item is sent, which
  // can be days later: the due date "tomorrow" and the day a task was done.
  function makeItem({ id, path, params }, now) {
    const p = Object.assign({}, params);
    let kind;
    let label;
    if (path === "/complete") {
      kind = "complete";
      label = p.path || "";
      if (!p.done_on) p.done_on = isoDate(now);
    } else if (p.kind === "dump") {
      kind = "dump";
      label = (p.name || "").trim();
    } else {
      kind = "task";
      label = (p.name || "").trim();
      // a subtask without a date is due with its parent
      if (!p.date && !p.parent) p.date = isoDate(tomorrow(now));
    }
    return { id, kind, path, params: p, label, createdAt: now.getTime(), state: "waiting", error: "" };
  }

  // classify says what an answer to a replayed write means.
  function classify(res) {
    if (res.type === "opaqueredirect" || res.status === 401 || res.status === 302 || res.status === 303) return "login";
    if (res.result === "ok") return "ok";
    if (res.result === "error") return "error";
    // anything else (an error page of a proxy, the offline page): try later
    return "network";
  }

  // replay sends the waiting items one by one, oldest first (a task is
  // added before its subtasks). store has list/put/remove, send(item)
  // resolves to {status, type, result, message} or rejects when the server
  // can't be reached. It stops at the first item it can't settle.
  async function replay(store, send) {
    const out = { sent: 0, failed: 0, stopped: "" };
    const items = (await store.list()).filter((i) => i.state === "waiting");

    for (const item of items) {
      let res;
      try {
        res = await send(item);
      } catch (err) {
        out.stopped = "network";
        break;
      }

      const verdict = classify(res);
      if (verdict === "ok") {
        await store.remove(item.id);
        out.sent++;
      } else if (verdict === "error") {
        await store.put(Object.assign({}, item, { state: "failed", error: res.message || "" }));
        out.failed++;
      } else {
        out.stopped = verdict;
        break;
      }
    }
    return out;
  }

  // status is what the bar says: the first of login, offline, syncing, the
  // number waiting, the number failed.
  function status(items, { online, syncing, login }) {
    const waiting = items.filter((i) => i.state === "waiting").length;
    const failed = items.filter((i) => i.state === "failed").length;
    if (login && waiting) return { key: "login", waiting, failed };
    if (!online) return { key: "offline", waiting, failed };
    if (syncing && waiting) return { key: "syncing", waiting, failed };
    if (waiting) return { key: "waiting", waiting, failed };
    if (failed) return { key: "failed", waiting, failed };
    return { key: "", waiting, failed };
  }

  const fmt = (text, ...args) => {
    let i = 0;
    return text.replace(/%[sdvq]/g, (m) => (m === "%q" ? JSON.stringify(String(args[i++])) : String(args[i++])));
  };

  const api = { isoDate, makeItem, classify, replay, status, fmt, WRITES, TABS };
  if (typeof module !== "undefined" && module.exports) module.exports = api;
  else root.AkrasiaQueue = api;
  if (typeof document === "undefined") return;

  // ---- page ----------------------------------------------------------

  const uuid = () => {
    if (crypto.randomUUID) return crypto.randomUUID();
    const b = crypto.getRandomValues(new Uint8Array(16));
    b[6] = (b[6] & 0x0f) | 0x40;
    b[8] = (b[8] & 0x3f) | 0x80;
    const h = Array.from(b, (x) => x.toString(16).padStart(2, "0")).join("");
    return `${h.slice(0, 8)}-${h.slice(8, 12)}-${h.slice(12, 16)}-${h.slice(16, 20)}-${h.slice(20)}`;
  };

  // IndexedDB keeps the queue across closing the app; without it (some
  // private windows) the queue lives until the page is closed.
  function memoryStore() {
    const items = new Map();
    return {
      list: async () => Array.from(items.values()).sort((a, b) => a.createdAt - b.createdAt),
      put: async (i) => void items.set(i.id, i),
      remove: async (id) => void items.delete(id),
    };
  }

  function idbStore() {
    const opened = new Promise((resolve, reject) => {
      const req = indexedDB.open("akrasia-queue", 1);
      req.onupgradeneeded = () => req.result.createObjectStore("items", { keyPath: "id" });
      req.onsuccess = () => resolve(req.result);
      req.onerror = () => reject(req.error);
    });
    const run = (mode, fn) =>
      opened.then(
        (db) =>
          new Promise((resolve, reject) => {
            const tx = db.transaction("items", mode);
            const req = fn(tx.objectStore("items"));
            tx.oncomplete = () => resolve(req.result);
            tx.onerror = () => reject(tx.error);
          }),
      );
    return {
      list: () => run("readonly", (s) => s.getAll()).then((all) => all.sort((a, b) => a.createdAt - b.createdAt)),
      put: (i) => run("readwrite", (s) => s.put(i)),
      remove: (id) => run("readwrite", (s) => s.delete(id)),
    };
  }

  let store;
  try {
    store = "indexedDB" in root ? idbStore() : memoryStore();
  } catch (err) {
    store = memoryStore();
  }

  let items = [];
  let syncing = false;
  let login = false;

  const $ = (sel) => document.querySelector(sel);

  async function reload() {
    try {
      items = await store.list();
    } catch (err) {
      store = memoryStore();
      items = [];
    }
    render();
  }

  // render shows the queue on the page: the bar, the list of what waits,
  // the tasks completed offline as done. It runs after every swap, because
  // htmx replaces the whole app.
  function render() {
    const bar = $("#queue-status");
    if (!bar) return;
    const t = bar.dataset;

    const s = status(items, { online: navigator.onLine, syncing, login });
    const parts = [];
    if (s.key === "login") parts.push(t.login);
    else if (s.key === "offline") parts.push(t.offline);
    else if (s.key === "syncing") parts.push(t.syncing);
    if (s.waiting && s.key !== "login") parts.push(fmt(t.waiting, s.waiting));
    if (s.failed) parts.push(fmt(t.failed, s.failed));
    bar.textContent = parts.join(" · ");
    bar.hidden = parts.length === 0;
    bar.dataset.state = s.key;

    $("#queue-list")?.remove();
    if (items.length) $("#app section.list")?.before(panel(t));

    for (const input of document.querySelectorAll("#app input[type=date]")) input.min = isoDate(new Date());
    for (const item of items) {
      if (item.kind !== "complete" || item.state !== "waiting") continue;
      markDone(item.params.path);
    }
  }

  function panel(t) {
    const section = document.createElement("section");
    section.id = "queue-list";
    section.className = "list";

    const title = document.createElement("h2");
    title.textContent = t.title;
    section.append(title);

    const list = document.createElement("ul");
    for (const item of items) {
      const li = document.createElement("li");
      li.className = "todo queued" + (item.state === "failed" ? " failed" : "");

      const icon = document.createElement("span");
      icon.className = "check";
      icon.textContent = item.kind === "complete" ? "✔" : "○";

      const name = document.createElement("span");
      name.className = "name";
      name.textContent = item.label;
      if (item.state === "failed" && item.error) {
        const why = document.createElement("small");
        why.className = "why";
        why.textContent = item.error;
        name.append(why);
      }

      const badge = document.createElement("span");
      badge.className = "due";
      const kind = t["kind" + item.kind[0].toUpperCase() + item.kind.slice(1)];
      badge.textContent = `${kind} · ${item.state === "failed" ? t.failedbadge : t.pending}`;

      li.append(icon, name, badge);
      if (item.state === "failed") {
        const drop = document.createElement("button");
        drop.className = "act del";
        drop.textContent = "×";
        drop.title = drop.ariaLabel = t.dismiss;
        drop.dataset.drop = item.id;
        li.append(drop);
      }
      list.append(li);
    }
    section.append(list);
    return section;
  }

  // markDone shows a task completed offline as done in the list
  function markDone(path) {
    for (const li of document.querySelectorAll("#app li.todo[data-path]")) {
      if (li.dataset.path !== path || li.classList.contains("done")) continue;
      li.classList.add("done");
      const form = li.querySelector("form");
      if (form) {
        const mark = document.createElement("span");
        mark.className = "check";
        mark.textContent = "✔";
        form.replaceWith(mark);
      }
    }
  }

  // ---- what happens when a request fails -----------------------------

  // every write gets its id before it is sent, and keeps it if it has to wait
  document.addEventListener("htmx:configRequest", (e) => {
    const cfg = e.detail;
    if (cfg.verb === "post" && WRITES.includes(cfg.path)) cfg.headers["X-Queue-Id"] = uuid();
  });

  async function enqueue(id, path, params) {
    await store.put(makeItem({ id, path, params }, new Date()));
    await reload();
    sync();
  }

  function failed(e) {
    const cfg = e.detail.requestConfig;
    if (!cfg) return;
    const path = cfg.path.split("?")[0];

    if (cfg.verb === "post" && WRITES.includes(path) && cfg.headers["X-Queue-Id"]) {
      const params = Object.fromEntries(cfg.formData.entries());
      enqueue(cfg.headers["X-Queue-Id"], path, params);
      // the form is done with: the task is in the queue
      if (path === "/add") cfg.elt.querySelector("[name=name]").value = "";
      else $("#overlay")?.replaceChildren();
    } else if (cfg.verb === "get" && path === "/complete") {
      openComplete(cfg.formData.get("path"), cfg.formData.get("tab"));
    } else if (cfg.verb === "get" && TABS.includes(path)) {
      // no copy of this tab as a fragment: the whole page is cached
      location.assign(path);
    } else {
      // deleting or converting a dump entry, opening its dialog: not offline
      notice($("#queue-status")?.dataset.needscomputer);
    }
  }

  // notice says something in the add bar, where the server's messages go
  function notice(text) {
    const form = $("#app form.add");
    if (!form || !text) return;
    form.querySelector(".msg")?.remove();
    const p = document.createElement("p");
    p.className = "msg err";
    p.setAttribute("role", "alert");
    p.textContent = text;
    form.prepend(p);
  }
  document.addEventListener("htmx:sendError", failed);
  document.addEventListener("htmx:timeout", failed);

  // openComplete is the completion dialog without the server: the notes,
  // and for a parent with pending subtasks whether to complete them too.
  function openComplete(path, tab) {
    const tpl = $("#tpl-offline-complete");
    if (!tpl) return;
    const dialog = tpl.content.cloneNode(true);
    const form = dialog.querySelector("form");
    const li = Array.from(document.querySelectorAll("#app li.todo[data-path]")).find((x) => x.dataset.path === path);
    const pending = li ? Number(li.dataset.pending) || 0 : 0;

    form.elements.path.value = path;
    form.elements.tab.value = tab || "";
    const title = dialog.querySelector(".title");
    if (pending > 0) {
      form.elements.all.disabled = false;
      title.textContent = fmt(form.dataset.all, path, pending);
      title.classList.add("warn");
    } else {
      title.textContent = fmt(form.dataset.notes, path);
    }
    $("#overlay").replaceChildren(dialog);
    $("#overlay textarea")?.focus();
  }

  document.addEventListener("submit", (e) => {
    const form = e.target.closest("form[data-offline-complete]");
    if (!form) return;
    e.preventDefault();
    const params = Object.fromEntries(new FormData(form).entries());
    $("#overlay").replaceChildren();
    enqueue(uuid(), "/complete", params);
  });

  document.addEventListener("click", async (e) => {
    const drop = e.target.closest("[data-drop]");
    if (drop) {
      await store.remove(drop.dataset.drop);
      reload();
    } else if (e.target.closest("#queue-status[data-state=login]")) {
      location.assign("/login");
    }
  });

  // ---- sending -------------------------------------------------------

  async function send(item) {
    const res = await fetch(item.path, {
      method: "POST",
      headers: { "HX-Request": "true", "X-Queue-Id": item.id, "Content-Type": "application/x-www-form-urlencoded" },
      body: new URLSearchParams(item.params),
      credentials: "same-origin",
      redirect: "manual",
    });
    const out = { status: res.status, type: res.type, result: res.headers.get("X-Akrasia-Result") || "", message: "" };
    if (out.result === "error") {
      const doc = new DOMParser().parseFromString(await res.text(), "text/html");
      out.message = doc.querySelector(".msg.err")?.textContent.trim() || "";
    }
    return out;
  }

  async function sync() {
    if (syncing || !navigator.onLine) return;
    items = await store.list().catch(() => []);
    if (!items.some((i) => i.state === "waiting")) return;

    syncing = true;
    render();
    try {
      const out = await replay(store, send);
      login = out.stopped === "login";
      if (out.sent > 0) refresh();
    } finally {
      syncing = false;
      await reload();
    }
  }

  // refresh shows what the server has now, unless the user is typing
  function refresh() {
    const busy = $("#overlay")?.firstChild || document.activeElement?.closest?.("#app form.add");
    if (!busy && root.htmx && location.pathname !== "/login") {
      root.htmx.ajax("GET", location.pathname, { target: "#app", swap: "outerHTML" });
    }
  }

  document.addEventListener("htmx:afterSettle", reload);
  root.addEventListener("online", () => {
    render();
    sync();
  });
  root.addEventListener("offline", render);
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") sync();
  });
  setInterval(sync, RETRY_MS);
  reload().then(sync);
})(globalThis);
