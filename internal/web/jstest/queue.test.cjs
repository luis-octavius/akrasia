// The logic of static/queue.js, which has no DOM: node --test internal/web/jstest
const test = require("node:test");
const assert = require("node:assert/strict");
const q = require("../static/queue.js");

// 2026-10-25 23:50 local time
const now = new Date(2026, 9, 25, 23, 50);

test("isoDate is the local date", () => {
  assert.equal(q.isoDate(now), "2026-10-25");
  assert.equal(q.isoDate(new Date(2027, 0, 5)), "2027-01-05");
});

test("a task without a date is due tomorrow, by the phone's clock", () => {
  const item = q.makeItem({ id: "a", path: "/add", params: { name: " Pay rent ", kind: "task", date: "", parent: "" } }, now);
  assert.equal(item.kind, "task");
  assert.equal(item.label, "Pay rent");
  assert.equal(item.params.date, "2026-10-26");
  assert.equal(item.state, "waiting");
});

test("tomorrow crosses month and year ends", () => {
  const dec31 = new Date(2026, 11, 31, 12);
  const item = q.makeItem({ id: "a", path: "/add", params: { name: "x", kind: "task" } }, dec31);
  assert.equal(item.params.date, "2027-01-01");
});

test("a subtask without a date keeps none: it is due with its parent", () => {
  const item = q.makeItem({ id: "a", path: "/add", params: { name: "Bake", kind: "task", parent: "Party", date: "" } }, now);
  assert.equal(item.params.date, "");
});

test("a date chosen by the user is kept", () => {
  const item = q.makeItem({ id: "a", path: "/add", params: { name: "x", kind: "task", date: "2026-12-01" } }, now);
  assert.equal(item.params.date, "2026-12-01");
});

test("a dump entry needs no date", () => {
  const item = q.makeItem({ id: "a", path: "/add", params: { name: "idea", kind: "dump" } }, now);
  assert.equal(item.kind, "dump");
  assert.equal(item.params.date, undefined);
});

test("a completion keeps the day it was done", () => {
  const item = q.makeItem({ id: "a", path: "/complete", params: { path: "Run", tab: "today", notes: "5km" } }, now);
  assert.equal(item.kind, "complete");
  assert.equal(item.label, "Run");
  assert.equal(item.params.done_on, "2026-10-25");
});

test("classify", () => {
  assert.equal(q.classify({ status: 200, result: "ok" }), "ok");
  assert.equal(q.classify({ status: 200, result: "error" }), "error");
  assert.equal(q.classify({ status: 401, result: "" }), "login");
  assert.equal(q.classify({ status: 303, result: "" }), "login");
  assert.equal(q.classify({ type: "opaqueredirect", status: 0, result: "" }), "login");
  assert.equal(q.classify({ status: 502, result: "" }), "network");
  assert.equal(q.classify({ status: 503, result: "" }), "network");
});

function fakeStore(items) {
  const map = new Map(items.map((i) => [i.id, i]));
  return {
    map,
    list: async () => Array.from(map.values()).sort((a, b) => a.createdAt - b.createdAt),
    put: async (i) => void map.set(i.id, i),
    remove: async (id) => void map.delete(id),
  };
}

const item = (id, createdAt, extra = {}) => ({ id, createdAt, kind: "task", state: "waiting", params: {}, label: id, ...extra });

test("replay sends oldest first and removes what was applied", async () => {
  const store = fakeStore([item("b", 2), item("a", 1), item("c", 3)]);
  const order = [];
  const out = await q.replay(store, async (i) => {
    order.push(i.id);
    return { status: 200, result: "ok" };
  });
  assert.deepEqual(order, ["a", "b", "c"]);
  assert.deepEqual(out, { sent: 3, failed: 0, stopped: "" });
  assert.equal(store.map.size, 0);
});

test("replay keeps a refused item as failed, with the reason, and goes on", async () => {
  const store = fakeStore([item("a", 1), item("b", 2)]);
  const out = await q.replay(store, async (i) =>
    i.id === "a" ? { status: 200, result: "error", message: "name taken" } : { status: 200, result: "ok" },
  );
  assert.deepEqual(out, { sent: 1, failed: 1, stopped: "" });
  assert.equal(store.map.get("a").state, "failed");
  assert.equal(store.map.get("a").error, "name taken");
  assert.equal(store.map.has("b"), false);
});

test("replay does not send failed items again", async () => {
  const store = fakeStore([item("a", 1, { state: "failed", error: "x" })]);
  let calls = 0;
  await q.replay(store, async () => void calls++);
  assert.equal(calls, 0);
});

test("replay stops when the server can't be reached, keeping everything", async () => {
  const store = fakeStore([item("a", 1), item("b", 2)]);
  const out = await q.replay(store, async () => {
    throw new TypeError("Failed to fetch");
  });
  assert.deepEqual(out, { sent: 0, failed: 0, stopped: "network" });
  assert.equal(store.map.size, 2);
});

test("replay stops at a login, keeping the rest", async () => {
  const store = fakeStore([item("a", 1), item("b", 2)]);
  const out = await q.replay(store, async (i) => (i.id === "a" ? { status: 200, result: "ok" } : { status: 401, result: "" }));
  assert.deepEqual(out, { sent: 1, failed: 0, stopped: "login" });
  assert.deepEqual(Array.from(store.map.keys()), ["b"]);
});

test("replay stops at an answer it doesn't understand", async () => {
  const store = fakeStore([item("a", 1), item("b", 2)]);
  const out = await q.replay(store, async () => ({ status: 502, result: "" }));
  assert.equal(out.stopped, "network");
  assert.equal(store.map.size, 2);
});

test("status", () => {
  const w = item("w", 1);
  const f = item("f", 2, { state: "failed" });
  const on = { online: true, syncing: false, login: false };
  assert.equal(q.status([], on).key, "");
  assert.equal(q.status([w], on).key, "waiting");
  assert.equal(q.status([f], on).key, "failed");
  assert.equal(q.status([w], { ...on, online: false }).key, "offline");
  assert.equal(q.status([], { ...on, online: false }).key, "offline");
  assert.equal(q.status([w], { ...on, syncing: true }).key, "syncing");
  assert.equal(q.status([w], { ...on, login: true }).key, "login");
  assert.deepEqual(q.status([w, w, f], on), { key: "waiting", waiting: 2, failed: 1 });
});

test("fmt fills the server's strings", () => {
  assert.equal(q.fmt("%d waiting", 3), "3 waiting");
  assert.equal(q.fmt("Notes for %q (optional):", "Pay rent"), 'Notes for "Pay rent" (optional):');
  assert.equal(q.fmt("Complete %s and its %d pending subtask(s)?", "Party", 2), "Complete Party and its 2 pending subtask(s)?");
});
