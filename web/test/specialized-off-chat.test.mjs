// The @site example of features/cli/market_listing_kinds.feature (K9, approved 2026-11-25):
// the Playbox chat shelf never offers a one-job (interaction "single_turn") model. The test
// runs the shelf's own bandsFromOffers, cut from src/js/playbox.js with its helpers.
// Run: node --test test/specialized-off-chat.test.mjs
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";

const dir = path.dirname(fileURLToPath(import.meta.url));
const js = readFileSync(path.join(dir, "../src/js/playbox.js"), "utf8");

// cut returns the source of one top-level function of playbox.js (two-space indent).
function cut(name) {
  const start = js.indexOf("  function " + name + "(");
  assert.ok(start >= 0, `playbox.js has no function ${name}`);
  const oneLine = js.indexOf("\n", start);
  if (js.slice(start, oneLine).trim().endsWith("}")) return js.slice(start, oneLine);
  const end = js.indexOf("\n  }\n", start);
  return js.slice(start, end + 4);
}

function shelf(offers) {
  const src = ["isOnline", "callsign", "capsOf", "looksEmbeddingOnly", "looksVision", "isSearch", "isOneJob", "bandsFromOffers"].map(cut).join("\n");
  const fn = new Function(src + "\nreturn bandsFromOffers;")();
  const bands = fn(offers);
  return (Array.isArray(bands) ? bands : Object.values(bands)).map((b) => b.model);
}

test("site Playbox chat shelf: a one-job model is not offered, a chat model is", () => {
  const offers = [
    { model: "llama3.2", modality: "chat", online: true, node_id: "a" },
    { model: "roger-operator", modality: "tts", online: true, node_id: "b" },
    { model: "web-search", modality: "search", online: true, node_id: "c" },
    { model: "wave-nano-fire-controller", modality: "chat", interaction: "single_turn", online: true, node_id: "d" },
    { model: "wave-pico-293m", modality: "chat", interaction: "single_turn", online: true, node_id: "e" },
  ];
  const got = shelf(offers);
  assert.ok(got.includes("llama3.2"), `llama3.2 missing: ${got}`);
  assert.ok(!got.includes("wave-nano-fire-controller"), `one-job model on the shelf: ${got}`);
  assert.ok(!got.includes("wave-pico-293m"), `one-job model on the shelf: ${got}`);
});

test("site Playbox chat shelf: a model with no interaction stays", () => {
  assert.deepEqual(shelf([{ model: "mystery-7b", online: true, node_id: "a" }]), ["mystery-7b"]);
});
