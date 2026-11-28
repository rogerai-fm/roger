// What differs between the live rosters (the voices on air, the search stations on air). Everything else,
// the reading, the hold on a transient failure, the filters and the table, is Roster.svelte.
//   row(raw)  one broker row -> { ref, name, op, pick, sub, latency, price, free }
//   cmd(ref)  the one command that uses a station; `ref` is the picked row's id, or none yet
const n0 = (v) => (typeof v === "number" && isFinite(v) && v > 0 ? v : 0);
const bare = (s) => String(s || "").replace(/^@/, "");
// The house's own search station is not shown on the search page (founder decision, 2026-11-26): the page is
// the operators' directory. The broker still lists it and still answers a query that names no station, so the
// words for an empty table say "listed", never "on air".
// The broker knows the house by account and does not say so in the feed, so the page can only match the
// callsign: the exact ones the house runs, never a prefix (an operator may be called house-anything). If the
// house takes another callsign it goes here; a `house` flag in the feed would retire this list.
const HOUSE = ["house-or"];
const house = (op) => HOUSE.includes(bare(op).toLowerCase());
// A station names itself, and its id ends up in a command the visitor copies into a shell. Only an id made of
// plain id characters is ever written into one; any other id (a quote, a space, a backslash, a $) is left out,
// so nothing an operator typed can break out of it. A row with such an id shows as a plain name, with no button to pick it.
export const safe = (ref) => (/^[A-Za-z0-9@/._:-]+$/.test(ref || "") ? ref : "");

export const KINDS = {
  voices: {
    path: "/voices", key: "voices", one: "voice", many: "voices", unit: "chars", meter: "metered by the character",
    head: ["voice · operator", "language", "latency"], pickLabel: "LANGUAGE", pickAll: "All languages",
    none: "No voices on air right now.", placeholder: "search a voice", share: "roger share", shareWhat: "Put a text-to-speech rig on air with",
    row: (v) => ({ ref: safe(v.namespaced_id), name: v.name || v.id || "voice", op: bare(v.operator), pick: v.language || "", sub: "",
      latency: n0(v.latency_ms), price: n0(v.price_per_1k_chars), free: !!v.free || !n0(v.price_per_1k_chars), ok: !!(v.id || v.name) }),
    fig: "HOW TO SPEAK", lead: "Pick a voice above, then speak with it from the CLI:", copy: "Copy the say command to clipboard",
    cmd: (ref) => `roger say --voice ${safe(ref) || "@<station>/<voice>"} "roger that"`,
    note: "The broker proxies every voice. A node's address, host and IP never leave it.",
  },
  search: {
    path: "/search/stations", key: "stations", one: "search station", many: "search stations", unit: "queries", meter: "metered by the query",
    head: ["station · operator", "engine", "latency"], pickLabel: "ENGINE", pickAll: "All engines",
    none: "No operator search stations are listed right now.", placeholder: "search a station, an engine or what it covers", share: "roger share --search", shareWhat: "Put a search on air with",
    row: (s) => ({ ref: safe(s.id), name: s.model || s.id || "search", op: bare(s.station), pick: s.engine || "", sub: s.scope || "",
      latency: n0(s.latency_ms), price: n0(s.price_per_1k_queries), free: !!s.free || !n0(s.price_per_1k_queries), ok: !!(s.id || s.model) && !house(s.station) }),
    fig: "HOW TO SEARCH", lead: "Send one query. Pick a station above to name it, or leave it out and the broker picks one:", copy: "Copy the search request to clipboard",
    cmd: (ref) => `curl https://broker.rogerai.fm/v1/search -H "Authorization: Bearer $ROGER_KEY" -d '{"query": "red shoes", "count": 5${safe(ref) ? `, "model": "${safe(ref)}"` : ""}}'`,
    note: "The broker relays every query. A station's address, host and IP never leave it.",
  },
};

export const SORTS = [
  ["cheapest", "cheapest", (a, b) => (a.free ? 0 : a.price) - (b.free ? 0 : b.price)],   // a free window keeps its listed price; free still sorts first
  ["fastest", "lowest latency", (a, b) => (a.latency || Infinity) - (b.latency || Infinity)],   // unmeasured last
  ["name", "name, A to Z", (a, b) => a.name.localeCompare(b.name)],
];

// A probe time, compact; "-" when unmeasured.
export function latency(ms) {
  if (!(ms > 0)) return "-";
  return ms >= 1000 ? (ms / 1000).toFixed(ms >= 10000 ? 0 : 1) + "s" : Math.round(ms) + "ms";
}
// A rate the broker computed, sub-cent aware so the smallest price never reads "$0.00" (fmt.js when the page has it).
export function usd(v) {
  if (typeof window !== "undefined" && window.RogerFmt && typeof window.RogerFmt.usd === "function") return window.RogerFmt.usd(v);
  return v >= 0.01 ? "$" + v.toFixed(2) : "$" + Number(v.toPrecision(3));
}
// Each row gets a key of its own. The broker sends one row per node, so one operator running the same search
// on two machines sends the same id twice, and two unnamed voices can share a name; a keyed list throws on a
// repeated key, in production too. The first keeps the plain id, the rest are numbered, and a number is
// skipped when another row already holds that very key (an id may itself end in "#1").
export function keyed(rows) {
  const used = new Set();
  return rows.map((r) => {
    const id = r.ref || r.name;
    let key = id;
    for (let n = 1; used.has(key); n++) key = `${id}#${n}`;
    used.add(key);
    return { ...r, key };
  });
}
// What the filters leave, in the chosen order (a stable sort, so equal rows keep the broker's order).
export function view(rows, { q = "", sort = "cheapest", free = false, pick = "" } = {}) {
  const words = q.trim().toLowerCase().split(/\s+/).filter(Boolean);
  const by = (SORTS.find((s) => s[0] === sort) || SORTS[0])[2];
  return rows.filter((r) => (!free || r.free) && (!pick || r.pick === pick)
    && words.every((w) => [r.name, r.op, r.pick, r.sub, r.ref].some((f) => f.toLowerCase().includes(w)))).sort(by);
}
