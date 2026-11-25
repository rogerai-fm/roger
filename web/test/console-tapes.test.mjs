// Regression locks for the FIG. 3 console (the "demo replay" tape deck) and the three tapes at
// the end of its preset bank:
//   - "using":   a FILM, drawn in the site's isometric hand (an inline SVG in the page, written by
//                computer-scientist/article/fig_film_gen.py), played by the frame engine in steps
//   - "hosting": the companion film, the same way
//   - "ping":    the real roger --ping screensaver capture (assets/ping-demo.{mp4,gif}), a MEDIA tape
// The two films replaced the sepia cartoon videos on 2026-11-23 (the founder's ask). These pin the
// tape markup, the film layer's contract and the one video's lazy-load contract: the multi-MB poster
// gif must NOT be fetched on first paint (the eager `poster=` is deferred to `data-poster=` the
// player swaps in on open/scroll-in, and the gif <img> fallback is `loading="lazy"`).
// Run: node --test test/console-tapes.test.mjs
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync, existsSync, statSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";
import { source } from "./_html.mjs";

const SRC = path.join(path.dirname(fileURLToPath(import.meta.url)), "../src");
const read = source;  // a page's source, legacy or app route (test/_html.mjs)

const index = read("index.html");
const termJs = read("js/terminal.js");
const homeCss = read("styles/home.css");

test("console: the five original CLI tapes are still present, in order", () => {
  for (const d of ["roger", "tunein", "agent", "share", "payouts"]) {
    assert.match(index, new RegExp(`data-demo="${d}"`), `preset ${d} present`);
  }
});

test("console: using, hosting and ping are appended to the preset bank, in that order", () => {
  for (const [d, label] of [["using", "using"], ["hosting", "hosting"], ["ping", "ping"]]) {
    const re = new RegExp(
      `<button class="term__preset" type="button" data-demo="${d}" aria-pressed="false">\\s*` +
      `<span class="term__preset-led" aria-hidden="true"></span>${label}`,
    );
    assert.match(index, re, `${d} preset button matches the tape markup`);
  }
  assert.ok(
    index.indexOf('data-demo="payouts"') < index.indexOf('data-demo="using"') &&
      index.indexOf('data-demo="using"') < index.indexOf('data-demo="hosting"') &&
      index.indexOf('data-demo="hosting"') < index.indexOf('data-demo="ping"'),
    "order is payouts < using < hosting < ping",
  );
});

test("console: the film layer holds the two drawn tapes, the card and the caption", () => {
  assert.match(index, /<div class="term__film" id="termFilm" hidden aria-hidden="true">/, "#termFilm, hidden until picked, decorative");
  for (const [id, name] of [["termFilmUsing", "using"], ["termFilmHosting", "hosting"]]) {
    const at = index.indexOf(`id="${id}"`);
    assert.ok(at > 0, `${name}: the stage exists`);
    const block = index.slice(at, index.indexOf("</svg>", at));
    assert.match(block, new RegExp(`<!-- iso:film-${name} -->`), `${name}: the generator's markers`);
    assert.match(block, /<svg class="fm__svg" viewBox="/, `${name}: the drawn scene`);
    assert.doesNotMatch(block, /<text\b/, `${name}: no SVG text (the words are the script's, in HTML)`);
    assert.match(block, /class="fm__comet fm__comet--out fm__route-1"[^>]*pathLength="100"/, `${name}: the request route carries its own dash scale`);
  }
  assert.match(index, /<p class="term__film-cap" id="termFilmCap"><\/p>/, "the caption line");
  assert.match(index, /<div class="term__film-card" id="termFilmCard">/, "the intertitle card");
  assert.match(index, /id="termFilmStep"[\s\S]*id="termFilmTitle"[\s\S]*id="termFilmLine"/, "the card's three lines");
  // the two films no longer ship as video: one <video> in the media layer, the ping capture
  const media = index.slice(index.indexOf('id="termMedia"'), index.indexOf("</div>", index.indexOf('id="termMedia"')));
  assert.equal((media.match(/<video\b/g) || []).length, 1, "one video tape left: ping");
  assert.match(media, /id="termPing"/, "and it is ping");
});

test("console: the ping tape stays a muted, looping, inline video with mp4 + gif", () => {
  assert.match(index, /id="termMedia"[^>]*\bhidden\b/, "#termMedia slot, hidden until picked");
  const block = index.slice(index.indexOf('id="termPing"'));
  assert.match(index, /<video[^>]*id="termPing"/, "ping: <video> present");
  assert.match(block.slice(0, 400), /\bmuted\b/, "ping: muted");
  assert.match(block.slice(0, 400), /\bloop\b/, "ping: loops");
  assert.match(block.slice(0, 400), /\bplaysinline\b/, "ping: plays inline");
  assert.match(block, /<source src="assets\/ping-demo\.mp4" type="video\/mp4">/, "ping: mp4 source");
  assert.match(block, /<img[^>]*src="assets\/ping-demo\.gif"/, "ping: gif fallback");
});

test("console: the ping <video> does NOT eagerly fetch its poster gif on first paint", () => {
  const block = index.slice(index.indexOf('id="termPing"'), index.indexOf('id="termPing"') + 400);
  assert.doesNotMatch(block, /(?<!-)\bposter="assets\//, "no eager poster= (would fetch the gif on paint)");
  assert.match(block, /data-poster="assets\/ping-demo\.gif"/, "poster deferred to data-poster");
  assert.match(block, /preload="none"/, "preload=none");
  const at = index.indexOf('class="term__video-gif" src="assets/ping-demo.gif"');
  const imgTag = index.slice(at, index.indexOf(">", at) + 1);
  assert.match(imgTag, /loading="lazy"/, "the gif <img> fallback is loading=\"lazy\"");
  assert.match(termJs, /data-poster|dataset\.poster|getAttribute\(["']data-poster["']\)/, "terminal.js reads the deferred data-poster");
});

test("terminal.js: using + hosting are FILM tapes wired to their stages, ping a MEDIA tape, in order", () => {
  assert.match(termJs, /using:\s*{[^}]*film:\s*true[^}]*el:\s*"termFilmUsing"/, "using -> #termFilmUsing");
  assert.match(termJs, /hosting:\s*{[^}]*film:\s*true[^}]*el:\s*"termFilmHosting"/, "hosting -> #termFilmHosting");
  assert.match(termJs, /ping:\s*{[^}]*media:\s*true[^}]*el:\s*"termPing"/, "ping -> #termPing");
  assert.ok(
    termJs.indexOf('el: "termFilmUsing"') < termJs.indexOf('el: "termFilmHosting"') &&
      termJs.indexOf('el: "termFilmHosting"') < termJs.indexOf('el: "termPing"'),
    "DEMOS order is using < hosting < ping",
  );
  // the films play through the frame engine: steps and cards, the last frame under reduced motion
  assert.match(termJs, /function filmCompile\(/, "the film compiler exists");
  assert.match(termJs, /function filmShow\(/, "a film frame is shown");
  assert.match(termJs, /classList\.toggle\("fm-s" \+ i, i <= k\)/, "the steps are cumulative classes");
  for (const name of ["using", "hosting"]) {
    const body = termJs.slice(termJs.indexOf(`${name}: {`), termJs.indexOf("});\n      }\n    }", termJs.indexOf(`${name}: {`)));
    const scenes = [...body.matchAll(/c\.scene\((\d)/g)].map((m) => Number(m[1]));
    assert.deepEqual(scenes, [1, 2, 3, 4, 5, 6, 7, 8], `${name}: eight steps in order`);
    assert.ok((body.match(/c\.card\(/g) || []).length >= 3, `${name}: intertitle cards`);
    assert.doesNotMatch(body, /\u2014|—/, `${name}: no em dash in its words`);
  }
  assert.match(termJs, /var ORDER = \[[^\]]*"payouts", "using", "hosting"\]/, "the films join the rotation after payouts");
  // the one video tape still plays
  assert.match(termJs, /function selectMedia\(/, "selectMedia branch exists");
  assert.match(termJs, /function playMedia\(/, "playMedia exists");
  assert.match(termJs, /prefers-reduced-motion|REDUCED/, "reduced-motion aware");
});

test("assets: the ping tape's files exist (no broken tile in prod)", () => {
  for (const f of ["ping-demo.mp4", "ping-demo.gif"]) {
    const p = path.join(SRC, "assets", f);
    assert.ok(existsSync(p), `assets/${f} exists`);
    assert.ok(statSync(p).size > 0, `assets/${f} is non-empty`);
  }
});

test("home.css: the media and film slots match the ASCII screen slot", () => {
  assert.match(homeCss, /\.term__media\s*{[^}]*min-height:\s*420px/, "media slot is the same 420px slot");
  assert.match(homeCss, /\.term__film\s*{[^}]*min-height:\s*420px/, "film slot is the same 420px slot");
  assert.match(homeCss, /\.term__video\s*{[^}]*width:\s*100%/, "video sizes full-width like the prior diagram");
  // the film's steps are classes on the layer, cumulative: s7 must come after s4 in the sheet so it wins
  assert.ok(homeCss.indexOf('.term__film[data-tape="using"].fm-s4') < homeCss.indexOf('.term__film[data-tape="using"].fm-s7'), "later steps are later rules");
});
