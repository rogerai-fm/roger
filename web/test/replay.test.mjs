// The recorded timelapse (app/lib/ui/Replay.svelte, replay.css, src/js/replay.js): one reusable figure
// driven by a data file. These hold the contract any page's data file must meet, and what the built
// Nano page shows from the fire record (src/data/fire-replay.json, written by the lab's exporter from
// a saved record of the two on-air models asked at every frame).
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync, existsSync } from "node:fs";
import path from "node:path";
import { DIST } from "./_html.mjs";

const WEB = path.resolve(path.dirname(new URL(import.meta.url).pathname), "..");
const data = JSON.parse(readFileSync(path.join(WEB, "src/data/fire-replay.json"), "utf8"));
const page = readFileSync(path.join(DIST, "research-wildfire.html"), "utf8");
const fig = page.match(/<figure class="figure figure--plate rp fr"[\s\S]*?<\/figure>/)?.[0] || "";
const css = readFileSync(path.join(WEB, "src/styles/replay.css"), "utf8");
const scene = readFileSync(path.join(WEB, "src/styles/fire-replay.css"), "utf8");
const js = readFileSync(path.join(WEB, "src/js/replay.js"), "utf8");
const STATES = data.ladder.map((r) => r.key);

test("timelapse data: every frame of every scenario is complete, in time order, with its picture on disk", () => {
  assert.ok(data.scenarios.length >= 1 && data.scenarios.length <= 3, "one to three separate scenarios");
  for (const s of data.scenarios) {
    assert.ok(s.title && s.date && s.frames.length >= 6, `${s.id}: a title, a date and a run of frames`);
    let x = -1;
    for (const [k, f] of s.frames.entries()) {
      const at = `${s.id} frame ${k} (${f.clock})`;
      assert.match(f.clock, /^\d\d:\d\d$/, `${at}: a clock`);
      assert.ok(f.x > x, `${at}: frames run forward in time`); x = f.x;
      assert.ok(existsSync(path.join(WEB, "src", f.src)), `${at}: its picture exists`);
      assert.match(f.close, /^(yes|no)$/, `${at}: the closer look is yes or no`);
      assert.ok(STATES.includes(f.state), `${at}: state ${f.state} is on the ladder`);
      assert.equal(typeof f.told, "boolean", `${at}: told or not`);
      assert.match(f.reading, /^[+\u2212]?\d+\.\d\d$/, `${at}: a reading to two places`);
      if (f.caption) assert.ok(f.caption.split(/\s+/).length <= 12, `${at}: a caption of at most 12 words`);
      assert.ok(f.ms >= 500 && (!f.caption || f.ms >= 3000), `${at}: a moment with a sentence holds long enough to read it`);
    }
    assert.ok(s.frames[0].caption, `${s.id}: the first frame says where we are`);
    assert.ok(s.ms <= 60000, `${s.id}: under a minute (${s.ms} ms)`);
  }
  const [id, k] = data.still; assert.ok(data.scenarios.find((s) => s.id === id)?.frames[k], "the still names a real frame");
});

test("timelapse data: the fire record says what the film says", () => {
  const fire = data.scenarios.find((s) => s.id === "fire"), calm = data.scenarios.find((s) => s.id === "false_alarm");
  const first = (pred) => fire.frames.find(pred);
  const start = first((f) => f.plus === "start"), seen = first((f) => f.close === "yes"), watch = first((f) => f.state === "watch"), up = first((f) => f.state === "elevated");
  assert.ok(start && seen && watch && up, "a start, a first yes, a watch and an elevated frame");
  assert.ok(fire.frames.slice(0, fire.frames.indexOf(start)).every((f) => f.state === "quiet" && f.close === "no" && !f.over), "the morning before the start is calm");
  assert.ok(start.x < seen.x && seen.x < watch.x && watch.x < up.x, "the order is: start, seen up close, watch, elevated");
  assert.equal(watch.told, true, "a person is told at watch, as the controller's action says");
  assert.ok(calm.frames.every((f) => f.close === "no" && !["elevated", "threat"].includes(f.state) && f.air === 0 && !f.asked), "the false alarm: never a yes, never an alert, clean air, the larger model never asked");
  const threat = first((f) => f.state === "threat");
  assert.ok(threat && threat.x > up.x && threat.by === "camera+chemistry" && threat.air === 2, "threat comes after elevated, decided by camera and chemistry, once the simulated air is a signal");
  assert.ok(fire.frames.every((f) => f.asked === ["elevated", "threat"].includes(f.state)), "the larger model is asked exactly when the state is an alert");
  assert.ok(calm.frames.some((f) => f.state === "watch") && calm.frames.at(-1).state === "quiet", "it rises to watch and ends quiet");
});

test("timelapse data: honest labels", () => {
  const text = JSON.stringify(data);
  assert.equal(data.chips[0], "Recorded timelapse"); assert.equal(data.chips[1], "Not live");
  const micro = data.parts.find((p) => p.id === "micro"), pico = data.parts.find((p) => p.id === "pico");
  assert.equal(micro.tag, "in training", "the larger model is in training"); assert.equal(micro.readout, null, "and has no readout of its own");
  assert.equal(data.chat.tag, "simulated example", "the exchange with it is tagged a simulated example");
  assert.ok(data.chat.q.endsWith("?") && data.chat.a.split(/\s+/).length <= 16, "a short question and a short reply");
  assert.equal(pico.tag, "simulated", "the air readings are simulated, and say so on the part");
  assert.match(data.note, /air readings and the wind are simulated, and the controller was asked with them/, "the note says the air is simulated and that the controller saw it");
  assert.match(data.note, /simulated example of what it will do, not an answer from the model/, "the note says the exchange is not the model's answer");
  assert.match(data.note + data.micro.rule, /never lower a decision/, "site context can add support, never lower a decision");
  for (const s of data.scenarios) for (const f of s.frames) assert.ok(f.chat === 0 || (f.asked && f.told), "the exchange shows only on an alert, after a person is told");
  const fire = data.scenarios.find((s) => s.id === "fire"), q = fire.frames.findIndex((f) => f.chat === 1);
  assert.ok(q > 0 && fire.frames[q].caption && fire.frames[q + 1].chat === 2, "the question arrives with its sentence, the reply on the next frame");
  assert.match(data.unit, /as designed/, "the drawing is a design, not an installed product");
  assert.match(data.note, /through the market/, "the note says where the answers came from");
  for (const p of data.parts) if (p.tone) assert.match(css, new RegExp(`\\[data-tone="${p.tone}"\\]`), `${p.id}: its tone is in the sheet`);
  assert.doesNotMatch(text, /deterministic|guarantee|certain\b/i, "no certainty claim");
  assert.doesNotMatch(text, /fire[\s-]*detect|detects?\s+(wild)?fires?/i, "the approved term is early warning");
  assert.doesNotMatch(text, /\u2014/, "no em dash");
});

test("timelapse on the wildfire page: every frame in the page, a still without script, real controls, a text version", () => {
  assert.ok(fig, "the figure is on the page");
  const n = data.scenarios.reduce((a, s) => a + s.frames.length, 0);
  assert.equal((fig.match(/<img class="rp__frame/g) || []).length, n, "every frame is in the page");
  assert.equal((fig.match(/<img class="rp__frame is-on"/g) || []).length, 1, "exactly one frame is the no-script still");
  for (const img of fig.match(/<img [^>]*>/g) || []) { assert.match(img, / alt="/, "every frame has an alt attribute"); assert.match(img, /loading="lazy"/, "frames load lazily"); }
  assert.equal((fig.match(/<button type="button" class="rp__tab"/g) || []).length, data.scenarios.length, "one real button a scenario");
  assert.match(fig, /<button class="rp__play" type="button"/, "play is a real button");
  assert.match(fig, /<input class="rp__range" type="range"[^>]*aria-label="Time"/, "time is a native range, reachable by keyboard");
  assert.match(fig, /<details class="rp__script">[\s\S]*<ol>(<li>[^<]+<\/li>){3,}<\/ol>/, "each scenario is readable as a list");
  for (const p of data.parts) assert.match(fig, new RegExp(`<p class="rp__part" data-part="${p.id}"`), `${p.id} is labelled on the drawing`);
  assert.match(fig, /<svg class="iso__svg"/, "the unit is drawn");
  assert.match(fig, /<span class="rp__statew rp__statew--feed">/, "the state is labelled on the feed, beside the clock");
  assert.match(fig, /<div class="rp__chat" data-tone="micro"><p class="rp__q">[\s\S]*?<p class="rp__a">[\s\S]*?simulated example/, "the exchange is in the page: a question, a reply tagged simulated");
  assert.doesNotMatch(fig, /<style|style="/, "no inline style in the figure");
});

test("the fire unit's drawing follows the timelapse, and nothing loops", () => {
  for (const a of ["data-smoke", "data-close", "data-state", "data-told", "data-air", "data-asked", "data-chat"]) assert.match(scene, new RegExp(`\\[${a}=`), `the drawing reads ${a}`);
  for (const p of data.parts) assert.match(scene, new RegExp(`\\.rp__part\\[data-part="${p.id}"\\] \\{ left: `), `${p.id}: its label has a generated anchor`);
  assert.match(fig, /class="fr__m fr__m--micro" data-tone="micro"/, "Wave Micro is drawn, in its tone");
  assert.match(css, /\.rp__ghost \{[^}]*stroke-dasharray/, "and drawn dashed: in training");
  for (const sheet of [css, scene]) assert.doesNotMatch(sheet, /infinite|@keyframes/, "no looping animation, no keyframe chains");
  assert.match(js, /prefers-reduced-motion: reduce/, "the clock reads the reduced-motion preference");
  assert.match(js, /IntersectionObserver/, "it plays in view and pauses off screen");
  assert.match(js, /visibilitychange/, "and pauses when the tab is hidden");
});

/* ---- the oil and gas timelapse: a simulated night, real answers, a stated rule ---------------------------------- */
const og = JSON.parse(readFileSync(path.join(WEB, "src/data/oilgas-replay.json"), "utf8"));
const ogPage = readFileSync(path.join(DIST, "research-oil-gas.html"), "utf8");
const ogFig = ogPage.match(/<figure class="figure figure--plate rp og"[\s\S]*?<\/figure>/)?.[0] || "";
const ogScene = readFileSync(path.join(WEB, "src/styles/oilgas-replay.css"), "utf8");

test("oil and gas data: complete frames, pictures on disk, captions short, states on its own ladder", () => {
  const states = og.ladder.map((r) => r.key);
  for (const s of og.scenarios) {
    let x = -1;
    for (const [k, f] of s.frames.entries()) {
      const at = `${s.id} frame ${k} (${f.clock})`;
      assert.ok(f.x > x, `${at}: frames run forward in time`); x = f.x;
      assert.ok(existsSync(path.join(WEB, "src", f.src)), `${at}: its picture exists`);
      assert.ok(states.includes(f.state), `${at}: state ${f.state} is on the ladder`);
      assert.equal(f.alert, f.state === "flag", `${at}: only a flag is an alert`);
      assert.ok(f.view, `${at}: the frame names its camera`);
      if (f.caption) assert.ok(f.caption.split(/\s+/).length <= 12, `${at}: a caption of at most 12 words`);
      for (const v of Object.values(f.pico)) assert.ok(v === null || ["none", "stuck", "dropout", "noisy", "drifting", "railed", "abstain"].includes(v), `${at}: Wave Pico's answer is one of the task's choices`);
    }
    assert.ok(s.traces.length >= 2 && s.traces.every((t) => t.label && t.points.length > 10), `${s.id}: labelled traces`);
    assert.ok(s.ms <= 75000, `${s.id}: at most 75 s (each camera view is held, so this film is slower than the fire one)`);
  }
});

test("oil and gas data: the rule gives the states, and the story is the record's", () => {
  const tap = og.scenarios.find((s) => s.id === "tap"), job = og.scenarios.find((s) => s.id === "job");
  for (const f of tap.frames) {                                  // quiet = no record out of line; a flag needs two kinds at once, then stays open
    if (f.state === "quiet") assert.equal(f.kinds.length, 0, `${f.clock}: quiet means nothing is out of line`);
    if (f.state === "watch") assert.equal(f.kinds.length, 1, `${f.clock}: a watch is one kind of record`);
  }
  const first = tap.frames.findIndex((f) => f.state === "flag");
  assert.ok(first > 0 && tap.frames[first].kinds.length >= 2, "the flag is raised when two kinds of record are out of line at once");
  assert.ok(tap.frames.slice(first).every((f) => f.state === "flag"), "a flag stays open until a person closes it");
  const truck = tap.frames.findIndex((f) => f.attrs.truck === 1), gap = tap.frames.findIndex((f) => f.over), stuck = tap.frames.findIndex((f) => f.pico["up-press"] === "stuck");
  assert.ok(truck > 0 && truck < gap && gap <= stuck, "the order is: the truck, the flow gap, the frozen sensor");
  assert.ok(tap.frames.slice(0, truck).every((f) => !f.alert), "nothing is flagged before the truck");
  assert.ok(tap.frames.some((f) => f.nano === "up-press" && f.pico["up-press"] === "stuck"), "Wave Nano names the sensor Wave Pico calls stuck");
  assert.ok(job.frames.every((f) => f.state === "quiet"), "the planned job: nothing is flagged");
  assert.ok(job.frames.some((f) => f.attrs.truck === 1 && f.attrs.wo === 1), "its truck arrives under an open work order");
});

test("oil and gas data: honest labels", () => {
  const text = JSON.stringify(og);
  assert.equal(og.chips[0], "Simulated site"); assert.equal(og.chips[1], "Not live");
  for (const b of og.boxes) assert.match(b.label, /^Site camera: /, "a box is the site camera system's detection, and says so");
  assert.match(og.note, /boxes on them are the site camera system's own detections, simulated: no Wave model recognises a vehicle/);
  assert.equal(og.feed.label, "illustration", "the pictures are labelled illustrations on the feed");
  assert.match(og.note, /sensor readings and the gate and work-order entries are simulated/);
  assert.match(og.note, /Wave Pico's and Wave Nano's answers are real/);
  assert.match(og.note, /The state is a site rule, not a model/);
  assert.match(og.note, /illustrations of no real site/);
  assert.match(og.note, /never names anyone and never changes a bill/);
  assert.equal(og.parts.find((p) => p.id === "micro").tag, "in training"); assert.equal(og.chat.tag, "simulated example");
  assert.match(og.chat.a, /Nothing here says who/, "the simulated reply names no one");
  assert.doesNotMatch(text, /thief|thieves|criminal|suspect|stolen by|caught/i, "a flag is a question, never an accusation");
  assert.doesNotMatch(text, /deterministic|guarantee|certain\b/i); assert.doesNotMatch(text, /—/, "no em dash");
});

test("oil and gas page: the timelapse is in the theft section, with its drawing, controls and text version", () => {
  assert.ok(ogFig, "the figure is on the page");
  assert.match(ogPage, /id="theft"[\s\S]*?<figure class="figure figure--plate rp og"/, "in the theft section");
  const n = og.scenarios.reduce((a, s) => a + s.frames.length, 0);
  assert.equal((ogFig.match(/<img class="rp__frame/g) || []).length, n, "every frame is in the page");
  assert.equal((ogFig.match(/<img class="rp__frame is-on"/g) || []).length, 1, "one no-script still");
  assert.equal((ogFig.match(/<button type="button" class="rp__tab"/g) || []).length, og.scenarios.length);
  assert.match(ogFig, /<input class="rp__range" type="range"/); assert.match(ogFig, /<details class="rp__script">/);
  for (const p of og.parts) { assert.match(ogFig, new RegExp(`<p class="rp__part" data-part="${p.id}"`)); assert.match(ogScene, new RegExp(`\\.rp__part\\[data-part="${p.id}"\\] \\{ left: `), `${p.id}: a generated anchor`); }
  for (const b of og.boxes) { assert.match(ogFig, new RegExp(`<p class="rp__box" data-box="${b.id}"`)); assert.match(ogScene, new RegExp(`\\.og\\[data-img="${b.id}"\\] \\.rp__box\\[data-box="${b.id}"\\]`), `${b.id}: its box shows on its own frame`); }
  const tapS = og.scenarios.find((s) => s.id === "tap"); let cuts = 0;
  for (let k = 1; k < tapS.frames.length; k++) if (tapS.frames[k].view !== tapS.frames[k - 1].view) { cuts++; assert.ok(tapS.frames[k].caption, `${tapS.frames[k].clock}: the camera changes only at a moment with a sentence`); }
  assert.ok(cuts <= 4, `the camera changes a few times, not every frame (${cuts})`);
  for (const a of ["data-truck", "data-gap", "data-s-up-press", "data-alert", "data-chat"]) assert.match(ogScene, new RegExp(`\\[${a}`), `the drawing reads ${a}`);
  assert.doesNotMatch(ogScene, /infinite|@keyframes/, "nothing in the drawing loops");
  assert.doesNotMatch(ogFig, /<style|style="/, "no inline style in the figure");
});

test("the model pages link to both timelapses with a picture and do not carry the figures themselves", () => {
  for (const page of ["research-wave-pico.html", "research-wave-nano.html"]) {
    const html = readFileSync(new URL(`../dist/${page}`, import.meta.url), "utf8");
    assert.match(html, /class="wm-door" href="\/research-wildfire\.html#timelapse"/, `${page}: the fire door`);
    assert.match(html, /class="wm-door" href="\/research-oil-gas\.html#theft-timelapse"/, `${page}: the oil and gas door`);
    assert.equal((html.match(/class="wm-door__shot"><img /g) || []).length, 2, `${page}: a frame on each door`);
    assert.doesNotMatch(html, /class="rp__frame"/, `${page}: the timelapse itself lives on its use-case page`);
  }
  const fire = readFileSync(new URL("../dist/research-wildfire.html", import.meta.url), "utf8");
  const og = readFileSync(new URL("../dist/research-oil-gas.html", import.meta.url), "utf8");
  assert.match(fire, /id="timelapse"/, "the fire door has somewhere to land");
  assert.match(og, /id="theft-timelapse"/, "the oil and gas door has somewhere to land");
});
