// Search stations ride the shared feeds (/discover, /market) tagged modality "search",
// the same way voice stations do (founder ruling 2026-11-23). A search station is priced
// per query and answers on /v1/search, so no chat surface on the site may show it as a
// model: not the Playbox shelf, the homepage panel, the band dial, or the calculator.
// Run: node --test test/search-stations-off-chat.test.mjs
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";

const dir = path.dirname(fileURLToPath(import.meta.url));
const read = (f) => readFileSync(path.join(dir, "../src/js", f), "utf8");

// market.js exports its pure parts when `document` is undefined (see market-meter.test.mjs).
function loadMarket() {
  const mod = { exports: {} };
  new Function("module", "window", read("market.js"))(mod, undefined);
  return mod.exports;
}

test("homepage panel: /market rows drop search models", () => {
  const R = loadMarket();
  const rows = [
    { model: "llama3.2", modality: "chat", providers: 1, signal: 50 },
    { model: "web-search", modality: "search", providers: 1, signal: 90 },
    { model: "old-broker-row", providers: 1, signal: 10 },
  ];
  assert.deepEqual(R.fromMarket(rows).map((c) => c.model).sort(), ["llama3.2", "old-broker-row"]);
});

test("homepage panel: /discover offers drop search stations", () => {
  const R = loadMarket();
  const offers = [
    { model: "llama3.2", modality: "chat", node_id: "n1", online: true },
    { model: "web-search", modality: "search", node_id: "n2", online: true },
  ];
  assert.deepEqual(R.aggregate(offers).map((c) => c.model), ["llama3.2"]);
});

test("homepage panel: a feed of only search stations reads as nothing on air", () => {
  const R = loadMarket();
  assert.deepEqual(R.fromMarket([{ model: "web-search", modality: "search", providers: 2 }]), []);
});

for (const f of ["playbox.js", "bands.js", "calc.js"]) {
  test(`${f}: search stations are dropped where the feed is read`, () => {
    const js = read(f);
    assert.ok(/function isSearch\(o\) \{ return !!o && o\.modality === "search"; \}/.test(js),
      `${f} must name the search check once`);
    const uses = js.split("isSearch(").length - 2; // minus the definition
    const want = f === "bands.js" ? 2 : 1;          // bands.js reads /market and /discover
    assert.ok(uses >= want, `${f} must apply the check on every feed it reads (found ${uses})`);
  });
}
