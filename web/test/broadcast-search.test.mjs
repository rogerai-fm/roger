// Broadcast 017 ("How search works on RogerAI, in pictures.") explains the Search API with
// plates and few words. These tests pin what makes it safe to publish and keep true:
//   - it is wired like every broadcast (index order, sitemap, llms.txt, OG card, FAQ schema);
//   - every figure explains itself to a screen reader, and both loops have a still stand-in
//     under reduced motion; the route strip's moving dot is decoration, opt-in only;
//   - every product claim matches docs/search-api.md and docs/search-providers.md: the
//     per-query price, the $50 ceiling, the $10 default cap, the three engine kinds, the
//     no-query-text rule and its one legal exception, and when a caller is not charged;
//   - it never says RogerAI crawls the web or owns an index, never names a hosting
//     provider, and never says "open source" (the licence is PolyForm Perimeter).
import { test, before } from "node:test";
import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";
import { decodeMarkup } from "./_html.mjs";

const WEB = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const PAGE = "broadcasts-search-in-pictures.html";
const read = (p) => readFileSync(path.join(WEB, "dist", p), "utf8");
const visible = (html) => html.replace(/<!--[\s\S]*?-->/g, "");
const mainOf = (html) => (visible(html).match(/<main\b[\s\S]*<\/main>/) || [""])[0];
const text = (html) => html.replace(/<script\b[\s\S]*?<\/script>/g, " ").replace(/<style\b[\s\S]*?<\/style>/g, " ")
  .replace(/<[^>]+>/g, " ").replace(/&nbsp;/g, " ").replace(/&rsquo;/g, "'").replace(/&amp;/g, "&").replace(/\s+/g, " ");
const PLATES = ["search-hero", "search-desks", "search-courier", "search-envelope", "search-coin", "search-own"];

before(() => execFileSync("node", ["build.mjs"], { cwd: WEB }));

/* ---- wired like every other broadcast ---------------------------------- */

test("search: the broadcast builds, is indexable, and is in the sitemap and llms.txt", () => {
  assert.doesNotMatch(read(PAGE), /name=["']robots["'][^>]*noindex/i);
  assert.match(read("sitemap.xml"), new RegExp(`<loc>https://rogerai\\.fm/${PAGE}</loc>`));
  assert.match(read("llms.txt"), new RegExp(`\\(https://rogerai\\.fm/${PAGE}\\)`));
});

test("search: the transmission log carries it first, as 017, above 016", () => {
  const log = read("broadcasts.html");
  const rows = [...log.matchAll(/<a class="bc-row__link" href="\/(broadcasts-[\w-]+\.html)">/g)].map((m) => m[1]);
  assert.equal(rows[0], PAGE, "newest first");
  assert.equal(rows[1], "broadcasts-how-llms-work.html", "017 sits right above 016");
  assert.match(log, /<b class="bc-row__no">BROADCAST 017<\/b>/);
  assert.match(log, /href="\/broadcasts-search-in-pictures\.html">[\s\S]*?datetime="2026-11-26"/, "the index row carries the publish date");
  assert.match(decodeMarkup(read(PAGE)), /BROADCAST 017 · FIELD GUIDE/);
});

test("search: a Quick Answer of 40 to 60 words opens the article", () => {
  const qa = read(PAGE).match(/<blockquote class="bc-answer"[^>]*>([\s\S]*?)<\/blockquote>/);
  assert.ok(qa, "a .bc-answer");
  assert.match(qa[1], /<b>Quick answer\.<\/b>/);
  const words = text(qa[1]).replace(/Quick answer\./, "").trim().split(" ").length;
  assert.ok(words >= 40 && words <= 60, `the Quick Answer is ${words} words`);
});

test("search: every section heading but the FAQ is a question a reader would type", () => {
  const h2s = [...mainOf(read(PAGE)).matchAll(/<section class="bc-sec"[^>]*>[\s\S]*?<h2>([^<]+)<\/h2>/g)].map((m) => m[1]);
  assert.equal(h2s.length, 8);
  for (const h of h2s.slice(0, 7)) assert.match(h, /\?$/, `"${h}" is a question`);
});

test("search: the FAQPage JSON-LD parses and mirrors the visible FAQ", () => {
  const html = read(PAGE);
  const m = html.match(/<script type="application\/ld\+json">(\{"@context":"https:\/\/schema\.org","@type":"FAQPage".*?)<\/script>/s);
  assert.ok(m, "a FAQPage block exists");
  const faq = JSON.parse(m[1]);
  assert.ok(faq.mainEntity.length >= 5, `at least 5 questions (got ${faq.mainEntity.length})`);
  const dts = [...html.matchAll(/<dt><b>([^<]*)<\/b><\/dt>\s*<dd>([\s\S]*?)<\/dd>/g)];
  assert.equal(dts.length, faq.mainEntity.length, "schema and visible FAQ have the same count");
  faq.mainEntity.forEach((q, i) => {
    assert.equal(q.name, decodeMarkup(dts[i][1]), "the questions match, in order");
    assert.equal(q.acceptedAnswer.text, decodeMarkup(text(dts[i][2]).trim()), `the answer to "${q.name}" matches`);
  });
});

test("search: the plates, the loops and the OG card are real shipped assets", () => {
  const html = read(PAGE);
  for (const name of PLATES) {
    assert.match(html, new RegExp(`src="assets/broadcasts/${name}\\.webp(\\?v=[0-9a-f]+)?"`), `${name} is shown as WebP`);
    for (const f of [`${name}.webp`, `${name}-800.webp`, `${name}.png`]) readFileSync(path.join(WEB, "dist", "assets", "broadcasts", f));
  }
  for (const loop of ["search-hero-loop.mp4", "search-own-loop.mp4"]) {
    assert.match(html, new RegExp(`<source src="assets/broadcasts/${loop.replace(".", "\\.")}(\\?v=[0-9a-f]+)?" type="video/mp4"`));
    readFileSync(path.join(WEB, "dist", "assets", "broadcasts", loop));
  }
  assert.match(html, /property="og:image" content="https:\/\/rogerai\.fm\/assets\/broadcasts\/search-og\.png"/);
  const og = readFileSync(path.join(WEB, "src", "assets", "broadcasts", "search-og.png"));
  assert.deepEqual([og.readUInt32BE(16), og.readUInt32BE(20)], [1200, 630], "the social card is 1200x630");
});

/* ---- every figure explains itself, motion is opt-in ---------------------- */

test("search: every figure has a caption, in order, and a text alternative", () => {
  const figs = [...mainOf(read(PAGE)).matchAll(/<figure\b[\s\S]*?<\/figure>/g)].map((m) => m[0]);
  assert.equal(figs.length, 8, "eight figures");
  figs.forEach((fig, i) => {
    assert.match(decodeMarkup(fig), new RegExp(`<figcaption class="fig mono">FIG\\. ${i + 1} · `), `FIG. ${i + 1} is captioned, in order`);
    for (const img of fig.match(/<img\b[^>]*>/g) || []) {
      const alt = img.match(/\salt="([^"]*)"/);
      assert.ok(alt && alt[1].length > 40, `FIG. ${i + 1}: the image says what it shows`);
    }
    for (const v of fig.match(/<video\b[^>]*>/g) || []) {
      assert.match(v, /\bmuted\b/); assert.match(v, /\bplaysinline\b/); assert.match(v, /preload="(none|metadata)"/);
      assert.match(v, /aria-label="[^"]{40,}"/, `FIG. ${i + 1}: the loop is described`);
      assert.match(fig, /<img class="srch-still"/, `FIG. ${i + 1}: a still stands in under reduced motion`);
    }
    if (!/<img\b|<video\b/.test(fig)) assert.match(fig, /role="(img|table|group)" aria-label="[^"]{60,}"|<table\b/, `FIG. ${i + 1}: the diagram is described in words`);
  });
});

test("search: a loop hides under reduced motion and the route dot only moves when motion is welcome", () => {
  const html = read(PAGE);
  const css = readFileSync(path.join(WEB, "src", "styles", "broadcast-search.css"), "utf8");
  assert.match(css, /@media \(prefers-reduced-motion: reduce\)\s*\{[^}]*\.srch-loop\s*\{\s*display:\s*none/);
  const anims = [...css.matchAll(/animation\s*:/g)].length;
  const guarded = (css.match(/@media \(prefers-reduced-motion: no-preference\)\s*\{[\s\S]*?\n\}/g) || []).join("");
  assert.equal([...guarded.matchAll(/animation\s*:/g)].length, anims, "every animation sits inside the no-preference guard");
  assert.doesNotMatch(css, /#[0-9a-f]{3,6}\b/i, "colours are tokens, never literals");
  assert.match(html, /class="srch-route"/);
});

test("search: the diagrams carry no SVG text (their words are page text, so they reflow on a phone)", () => {
  const figs = (mainOf(read(PAGE)).match(/<figure\b[\s\S]*?<\/figure>/g) || []).join("");
  for (const svg of figs.match(/<svg\b[\s\S]*?<\/svg>/g) || []) {
    assert.doesNotMatch(svg, /<text\b/);
    assert.match(svg, /aria-hidden="true"/);
  }
});

/* ---- every claim matches the docs ------------------------------------------ */

test("search: the money facts match docs/search-api.md", () => {
  const t = text(mainOf(read(PAGE)));
  assert.match(t, /\$50 per 1,000/, "the ceiling");
  assert.match(t, /\$10 per 1,000/, "the default cap with no station named");
  assert.match(t, /one query/i);
  assert.match(t, /fails, times out/, "the no-charge cases");
  assert.match(t, /zero results/i, "a station that answered with nothing is still paid");
  assert.match(t.replace(/ ([,.])/g, "$1"), /Without a login, only free stations answer/, "an unsigned caller reaches free stations only");
  assert.doesNotMatch(t, /\$\d+(\.\d+)? per (query|search)\b/, "no per-query dollar figure the docs do not give");
});

test("search: the privacy facts match the docs, with the one legal exception said plainly", () => {
  const t = text(mainOf(read(PAGE)));
  assert.match(t, /keeps no query text/);
  assert.match(t, /pseudonym/);
  assert.match(t, /different for (every|each) station/);
  assert.match(t, /child sexual abuse/, "the CSAM exception is stated, not hidden");
  assert.match(t, /as the law requires/);
});

test("search: the index facts are honest: stations bring engines, RogerAI crawls nothing", () => {
  const t = text(mainOf(read(PAGE)));
  for (const kind of ["brave", "searxng", "native"]) assert.match(t, new RegExp(`\\b${kind}\\b`, "i"), `the ${kind} kind is named`);
  assert.match(t, /search\/1/, "the open contract is named");
  assert.match(t, /does not crawl/);
  assert.doesNotMatch(t, /\b(our|RogerAI's) (own )?(search )?index\b/i, "RogerAI owns no index");
  assert.match(t, /signed-in owner/, "a search station needs a signed-in owner");
  assert.match(t, /operator_status/);
  assert.doesNotMatch(t, /\bmixed\b/i, "\"mixed\" is a per-model /market field and never applies to search");
  assert.doesNotMatch(t, /\bnews search\b/i, "there is no news kind");
});

test("search: the commands are the real ones from docs/search-providers.md", () => {
  const html = decodeMarkup(read(PAGE));
  assert.match(html, /roger share --search searxng --upstream http:\/\/127\.0\.0\.1:8888/);
  assert.match(html, /roger login/);
  assert.match(html, /roger share --search brave --price-per-1k 5/, "a Brave station must set a price (5: cmd/rogerai/share_search.go help and the house station; docs/search-providers.md still shows 6)");
  assert.match(html, /BRAVE_SEARCH_API_KEY/, "where the Brave credential goes");
  assert.match(html, /roger 7\.2\.0 or later/, "sharing search needs 7.2.0");
  assert.match(html, /roger payout onboard/, "a paid station onboards once");
  assert.match(html, /roger bands new-code/, "a lost frequency code is replaced, not shown again");
  assert.match(html, /roger search --kind search/, "how to list search stations");
  assert.match(html, /href="https:\/\/rogerai\.fm\/manual#sSearch"/, "links the manual");
  assert.match(html, /href="https:\/\/github\.com\/rogerai-fm\/roger\/blob\/main\/docs\/search-providers\.md"/, "links the provider guide");
  assert.doesNotMatch(html, /house-or\/roger-search/, "the house station is left out");
  assert.match(html, /add json to search\.formats in settings\.yml/, "SearXNG needs JSON output on");
  assert.match(html, /roger search check --upstream http:\/\/127\.0\.0\.1:9000/, "a native provider checks first");
  assert.match(html, /X-Roger-Freq/, "the frequency header is the documented one");
  assert.match(html, /https:\/\/broker\.rogerai\.fm\/v1\/search/);
});

/* ---- what never ships ------------------------------------------------------- */

test("search: no em dashes, no 'open source', no hosting provider, in the source or the build", () => {
  const route = readFileSync(path.join(WEB, "app", "routes", "broadcasts-search-in-pictures", "+page.svelte"), "utf8");
  for (const [name, s] of [["source", route], ["build", mainOf(read(PAGE))]]) {
    assert.doesNotMatch(s, /—|&mdash;/, `${name}: no em dash`);
    assert.doesNotMatch(text(s), /open[- ]source/i, `${name}: the licence is PolyForm Perimeter`);
    assert.doesNotMatch(s, /\b(dokku|ovh|digitalocean|cloudflare|porkbun|zeptomail|stripe)\b/i, `${name}: provider-neutral`);
  }
});
