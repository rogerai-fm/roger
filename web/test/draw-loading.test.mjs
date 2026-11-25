// A figure marked data-draw draws itself in only if the page loads src/js/draw.js, which a page opts
// into through its own module (app/pages/<page>.js). Without it the figure still shows, complete, so a
// page that forgets the import fails silently: it never draws in. This keeps every such page honest
// (2026-11-24, from the UI audit). data-draw-skip on a stroke is not a drawn figure.
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync, readdirSync, existsSync } from "node:fs";
import path from "node:path";
import { DIST, APP, pageBundle } from "./_html.mjs";

test("every page with a drawn figure loads draw.js", () => {
  const missing = [];
  for (const f of readdirSync(DIST).filter((n) => n.endsWith(".html"))) {
    const html = readFileSync(path.join(DIST, f), "utf8");
    if (!/\sdata-draw(?=[\s>=])(?!-)/.test(html)) continue;
    const mod = html.match(/js\/pages\/([\w-]+)\.js/)?.[1];
    if (!mod || !existsSync(path.join(APP, "pages", `${mod}.js`)) || !pageBundle(mod).includes("draw")) missing.push(f);
  }
  assert.deepEqual(missing, [], "these pages draw a figure but never load draw.js");
});
