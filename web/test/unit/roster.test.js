// The live rosters, /voices.html and /search.html (lib/roster/Roster.svelte over lib/roster/kinds.js): the rules
// the voices page script kept (real data only, hold on a transient failure, honest quiet and unreachable states),
// the tuner over the rows in hand, and the command that names the picked station.
import { describe, it, expect, beforeEach } from "vitest";
import { render, waitFor, fireEvent } from "@testing-library/svelte";
import Roster from "#lib/roster/Roster.svelte";
import { KINDS, keyed, safe, view, latency } from "#lib/roster/kinds.js";
import { broker } from "./_broker.js";

const el = (id) => document.getElementById(id);
const text = (id) => el(id).textContent.replace(/\s+/g, " ").trim();
const names = () => [...document.querySelectorAll("#rosterList .voice-row .voice-name")].map((n) => n.textContent);
const STATIONS = { stations: [
  { id: "@otter/docs", model: "docs", station: "@otter", engine: "native", scope: "Go module docs", price_per_1k_queries: 0, free: true, latency_ms: 420 },
  { id: "@lynx/web", model: "web", station: "@lynx", engine: "brave", scope: "", price_per_1k_queries: 5, free: false, latency_ms: null },
  { id: "@fox/meta", model: "meta", station: "@fox", engine: "searxng", scope: "", price_per_1k_queries: 2, free: false, latency_ms: 90 },
] };
const VOICES = { voices: [{ id: "v", namespaced_id: "@house/op", operator: "house", name: "Operator", price_per_1k_chars: 0.001, free: false, language: "en-US" }] };

beforeEach(() => {
  // the roster reads when it comes into view: here it is in view at once
  globalThis.IntersectionObserver = class { constructor(cb) { this.cb = cb; } observe() { this.cb([{ isIntersecting: true }]); } disconnect() {} };
  window.matchMedia = () => ({ matches: true });   // reduced motion: no background polling in a test
});
const live = () => waitFor(() => expect(el("rosterStatus").classList.contains("is-live")).toBe(true));

describe("the search roster", () => {
  it("lists what the broker lists, cheapest first, with engine, scope, latency and price", async () => {
    const calls = broker({ "/search/stations": [200, STATIONS] });
    render(Roster, { kind: "search" });
    await live();
    expect(calls.map((c) => c.path)).toEqual(["/search/stations"]);
    expect(calls[0].opts.credentials).toBeUndefined();   // a public feed: no cookie is sent
    expect(text("rosterStatusText")).toBe("3 search stations listed · live from /search/stations");
    expect(names()).toEqual(["docs", "meta", "web"]);
    const first = document.querySelector("#rosterList .voice-row").textContent.replace(/\s+/g, " ");
    expect(first).toContain("by @otter");
    expect(first).toContain("Go module docs");
    expect(first).toContain("native");
    expect(first).toContain("420ms");
    expect(first).toContain("FREE");
    expect(document.querySelectorAll("#rosterList .voice-row")[2].textContent).toContain("$5.00");
  });
  it("the tuner: words, engine, free now and sort work over the rows in hand, and clear resets them", async () => {
    broker({ "/search/stations": [200, STATIONS] });
    render(Roster, { kind: "search" });
    await live();
    expect([...el("rosterPick").options].map((o) => o.value)).toEqual(["", "brave", "native", "searxng"]);
    await fireEvent.input(el("rosterSearch"), { target: { value: "go docs" } });
    expect(names()).toEqual(["docs"]);
    await fireEvent.input(el("rosterSearch"), { target: { value: "nothing like this" } });
    expect(names()).toEqual([]);
    expect(el("rosterEmpty").hidden).toBe(false);
    await fireEvent.click(el("rosterClear"));
    expect(names().length).toBe(3);
    await fireEvent.click(el("rosterFree"));
    expect(names()).toEqual(["docs"]);
    expect(el("rosterFree").getAttribute("aria-pressed")).toBe("true");
    await fireEvent.click(el("rosterFree"));
    await fireEvent.change(el("rosterPick"), { target: { value: "brave" } });
    expect(names()).toEqual(["web"]);
    await fireEvent.click(el("rosterClear"));
    await fireEvent.change(el("rosterSort"), { target: { value: "fastest" } });
    expect(names()).toEqual(["meta", "docs", "web"]);   // an unmeasured station sorts last
  });
  it("the command names no station until one is picked; picking again lets go", async () => {
    broker({ "/search/stations": [200, STATIONS] });
    render(Roster, { kind: "search" });
    await live();
    expect(text("rosterCmdCode")).toBe(`curl https://broker.rogerai.fm/v1/search -H "Authorization: Bearer $ROGER_KEY" -d '{"query": "red shoes", "count": 5}'`);
    const pick = document.querySelectorAll(".voice-pick")[2];
    await fireEvent.click(pick);
    expect(text("rosterCmdCode")).toContain(`"count": 5, "model": "@lynx/web"}'`);
    expect(pick.getAttribute("aria-pressed")).toBe("true");
    await fireEvent.click(pick);
    expect(text("rosterCmdCode")).not.toContain("model");
  });
  it("the house's own station is not shown, and a table left empty by that says 'listed', never 'on air'", async () => {
    broker({ "/search/stations": [200, { stations: [{ id: "@house-or/roger-search", model: "roger-search", station: "@house-or", engine: "brave", price_per_1k_queries: 5 }] }] });
    render(Roster, { kind: "search" });
    await waitFor(() => expect(el("rosterQuiet").hidden).toBe(false));
    expect(names()).toEqual([]);
    expect(text("rosterQuiet")).toContain("No operator search stations are listed right now.");
    expect(document.body.textContent).not.toMatch(/on air (yet|right now)\b.*search|no search stations on air/i);
    expect(KINDS.voices.row({ id: "v", operator: "house-or", name: "Operator" }).ok).toBe(true);   // the voices page still lists the house voices
  });
  it("one operator on two machines sends the same id twice: both rows show and nothing throws", async () => {
    const twin = { id: "@otter/docs", model: "docs", station: "@otter", engine: "native", price_per_1k_queries: 1 };
    broker({ "/search/stations": [200, { stations: [twin, { ...twin, latency_ms: 80 }, twin] }] });
    render(Roster, { kind: "search" });
    await live();
    expect(names()).toEqual(["docs", "docs", "docs"]);
    await fireEvent.change(el("rosterSort"), { target: { value: "fastest" } });   // a reorder is where a repeated key breaks
    expect(names().length).toBe(3);
  });
  it("an operator whose callsign only starts with house is listed", async () => {
    broker({ "/search/stations": [200, { stations: [{ id: "@house-of-cards/s", model: "s", station: "@house-of-cards", engine: "native" }, { id: "@House-OR/x", model: "x", station: "@House-OR", engine: "brave" }] }] });
    render(Roster, { kind: "search" });
    await live();
    expect(names()).toEqual(["s"]);
  });
  it("an empty answer is the quiet band, never an invented row", async () => {
    broker({ "/search/stations": [200, { stations: [] }] });
    render(Roster, { kind: "search" });
    await waitFor(() => expect(el("rosterQuiet").hidden).toBe(false));
    expect(names()).toEqual([]);
    expect(text("rosterStatusText")).toBe("the band is quiet. No operator search stations are listed right now.");
    expect(text("rosterQuiet")).toContain("roger share --search");
  });
  for (const [what, routes] of [["a refusal", { "/search/stations": [429, {}] }], ["no answer", null]]) {
    it(`${what} with nothing in hand says the broker could not be reached`, async () => {
      if (routes) broker(routes); else globalThis.fetch = async () => { throw new Error("offline"); };
      render(Roster, { kind: "search" });
      await waitFor(() => expect(el("rosterStatus").classList.contains("is-off")).toBe(true));
      expect(text("rosterQuiet")).toContain("Couldn't reach the broker");
      expect(names()).toEqual([]);
    });
  }
  it("a failure after a good read holds the last roster (a refusal carries no rows; it is not an empty band)", async () => {
    let answer = [200, STATIONS];
    broker({ "/search/stations": () => answer });
    render(Roster, { kind: "search" });
    await live();
    answer = [429, {}];
    await fireEvent.click(el("rosterRefresh"));
    await waitFor(() => expect(text("rosterStatusText")).toBe("holding the last roster · re-reading…"));
    expect(names().length).toBe(3);
    expect(el("rosterQuiet").hidden).toBe(true);
  });
});

describe("the voices roster", () => {
  it("reads /voices and writes the picked voice into the say command", async () => {
    const calls = broker({ "/voices": [200, VOICES] });
    render(Roster, { kind: "voices" });
    await live();
    expect(calls[0].path).toBe("/voices");
    expect(text("rosterStatusText")).toBe("1 voice listed · live from /voices");
    expect(text("rosterCmdCode")).toBe('roger say --voice @<station>/<voice> "roger that"');
    await fireEvent.click(document.querySelector(".voice-pick"));
    expect(text("rosterCmdCode")).toBe('roger say --voice @house/op "roger that"');
    expect(document.querySelector("#rosterList .voice-row").textContent).toContain("en-US");
  });
  it("two unnamed-station voices with one name both show", async () => {
    broker({ "/voices": [200, { voices: [{ id: "a", name: "Kokoro", language: "en-US" }, { id: "b", name: "Kokoro", language: "en-GB" }] }] });
    render(Roster, { kind: "voices" });
    await live();
    expect(names()).toEqual(["Kokoro", "Kokoro"]);
    expect(document.querySelectorAll(".voice-pick").length).toBe(0);   // with no station id there is nothing to name in the command
  });
});

describe("kinds.js", () => {
  it("a row with no id and no name is not a row; a zero price is free", () => {
    expect(KINDS.search.row({}).ok).toBe(false);
    expect(KINDS.search.row({ id: "@a/b", price_per_1k_queries: 0 }).free).toBe(true);
    expect(KINDS.voices.row({ id: "x", price_per_1k_chars: 0.5 }).free).toBe(false);
  });
  it("view filters by every word and keeps the broker's order among equals", () => {
    const rows = STATIONS.stations.map(KINDS.search.row);
    expect(view(rows, { q: "FOX" }).map((r) => r.name)).toEqual(["meta"]);
    expect(view(rows, { sort: "name" }).map((r) => r.name)).toEqual(["docs", "meta", "web"]);
  });
  it("cheapest puts a free row first even when it keeps a listed price", () => {
    const rows = [{ name: "paid", price: 1, free: false }, { name: "window", price: 3, free: true }].map((r) => ({ op: "", pick: "", sub: "", ref: "", latency: 0, ...r }));
    expect(view(rows, { sort: "cheapest" }).map((r) => r.name)).toEqual(["window", "paid"]);
  });
  it("keyed gives a repeated id a key of its own and leaves the first as it was", () => {
    expect(keyed([{ ref: "a", name: "x" }, { ref: "a", name: "x" }, { ref: "", name: "n" }, { ref: "", name: "n" }]).map((r) => r.key)).toEqual(["a", "a#1", "n", "n#1"]);
  });
  it("keyed never repeats a key, even when an id already ends in a number", () => {
    const keys = keyed([{ ref: "a" }, { ref: "a" }, { ref: "a#1" }, { ref: "a" }]).map((r) => r.key);
    expect(keys).toEqual(["a", "a#1", "a#1#1", "a#2"]);
    expect(new Set(keys).size).toBe(keys.length);
  });
  it("an id that is not plain id characters is never written into a command", () => {
    for (const bad of [`@x/a'; rm -rf ~ #`, `@x/a"b`, "@x/a\\b", "@x/$(id)", "@x/a b", "@x/a`id`"]) {
      expect(safe(bad)).toBe("");
      expect(KINDS.search.cmd(bad)).toBe(KINDS.search.cmd(""));
      expect(KINDS.voices.cmd(bad)).toBe(KINDS.voices.cmd(""));
    }
    expect(KINDS.search.row({ id: `@x/a'b`, model: "a" })).toMatchObject({ ref: "", name: "a", ok: true });   // listed, not pickable
    expect(KINDS.voices.row({ id: "v", namespaced_id: "@x/a b" }).ref).toBe("");
    expect(KINDS.search.cmd("@otter/docs-v1.2:a_b")).toContain(`"model": "@otter/docs-v1.2:a_b"`);
    expect(KINDS.voices.cmd("@house/op")).toContain("--voice @house/op ");
  });
  it("latency reads compact, and unmeasured is a dash", () => {
    expect([latency(0), latency(90), latency(1500), latency(12000)]).toEqual(["-", "90ms", "1.5s", "12s"]);
  });
});

describe("pacing", () => {
  it("a 429's Retry-After sets the wait before the next read; without one the wait is the 30s cadence", async () => {
    window.matchMedia = () => ({ matches: false });   // motion allowed: the roster polls
    const waits = [], real = globalThis.setTimeout;
    globalThis.setTimeout = (fn, ms, ...a) => { if (ms >= 30000) { waits.push(ms); return 0; } return real(fn, ms, ...a); };
    try {
      globalThis.fetch = async () => ({ ok: false, status: 429, headers: { get: (h) => (h === "Retry-After" ? "90" : null) }, json: async () => ({}) });
      const a = render(Roster, { kind: "search" });
      await waitFor(() => expect(waits).toEqual([90000]));
      a.unmount();
      waits.length = 0;
      broker({ "/search/stations": [200, STATIONS] });
      render(Roster, { kind: "search" });
      await waitFor(() => expect(waits).toEqual([30000]));
    } finally { globalThis.setTimeout = real; }
  });
  it("a 429 drawn by the refresh button paces the next read too", async () => {
    window.matchMedia = () => ({ matches: false });
    const waits = [], real = globalThis.setTimeout;
    globalThis.setTimeout = (fn, ms, ...a) => { if (ms >= 30000) { waits.push(ms); return 0; } return real(fn, ms, ...a); };
    try {
      broker({ "/search/stations": [200, STATIONS] });
      render(Roster, { kind: "search" });
      await waitFor(() => expect(waits).toEqual([30000]));
      globalThis.fetch = async () => ({ ok: false, status: 429, headers: { get: (h) => (h === "Retry-After" ? "120" : null) }, json: async () => ({}) });
      await fireEvent.click(el("rosterRefresh"));
      await waitFor(() => expect(waits).toEqual([30000, 120000]));
    } finally { globalThis.setTimeout = real; }
  });
});
