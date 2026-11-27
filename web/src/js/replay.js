/* =====================================================================
   replay.js - the clock of a recorded timelapse (Replay.svelte, replay.css; 2026-11-26)
   A [data-replay] figure holds every frame of every scenario in the page, each with its own
   state as data attributes. This shows a frame: it turns the frame on, writes its clock, reading
   and answers into the labels, sets the figure's attributes (state, alert, over, look, and every key of the frame's data-a as data-*; the scene's sheet reads them) and
   moves the playhead. Showing a frame is a pure function of its number, so play, pause, scrub
   and replay are the same call.
   It plays once when most of the feed is in view, pauses off screen and when the tab is hidden,
   and rests on the last frame. The range scrubs it (and stops it); a scenario button starts its
   own story from the first frame. Under prefers-reduced-motion it never plays by itself: the
   still stays, and the button steps to the next captioned moment.
   No script: the page's own still and the text list are the figure.
   ===================================================================== */
(function () {
  if (typeof document === "undefined") return;
  [].forEach.call(document.querySelectorAll("figure[data-replay]"), function (fig) {
    var btn = fig.querySelector(".rp__play"), range = fig.querySelector(".rp__range"), mark = fig.querySelector(".rp__head-mark");
    var tabs = [].slice.call(fig.querySelectorAll(".rp__tab"));
    if (!btn || !range) return;
    var still = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    var $ = function (sel) { return fig.querySelector(sel); };
    var out = { time: $(".rp__time"), plus: $(".rp__plus"), cap: $(".rp__cap"), axis: $(".rp__axistext"), view: $(".rp__viewname") };
    var states = [].slice.call(fig.querySelectorAll(".rp__statew")), outs = [].slice.call(fig.querySelectorAll(".rp__out"));
    var scn = fig.getAttribute("data-scn"), frames = [], n = -1, raf = 0, t = 0, last = 0, held = false, started = false;

    function load(id) {
      scn = id; frames = [].slice.call(fig.querySelectorAll('.rp__frame[data-scn="' + id + '"]'));
      range.max = String(frames.length - 1);
      fig.setAttribute("data-scn", id);
      tabs.forEach(function (b) { b.setAttribute("aria-pressed", String(b.getAttribute("data-scn") === id)); });
      [].forEach.call(fig.querySelectorAll(".rp__trace"), function (s) { s.classList.toggle("is-on", s.getAttribute("data-scn") === id); });
    }
    // A frame's scene states (data-a) and readouts (data-r) are one JSON object each. They were spread as
    // one attribute per key until 2026-11-26: a spread on an <img> makes the renderer add inline onload and
    // onerror handlers, which the site's content security policy refuses (78 console errors on one page).
    function bag(f, name) { try { return JSON.parse(f.getAttribute(name) || "{}") || {}; } catch (e) { return {}; } }
    function dur(k) { return Number(frames[k].getAttribute("data-ms")) || 800; }
    function total() { var ms = 0; for (var i = 0; i < frames.length; i++) ms += dur(i); return ms; }
    function at(ms) { var k = 0; while (k < frames.length - 1 && ms >= dur(k)) ms -= dur(k++); return k; }
    function startOf(k) { var ms = 0; for (var i = 0; i < k; i++) ms += dur(i); return ms; }
    function caption(k) { for (var i = k; i >= 0; i--) { var c = frames[i].getAttribute("data-say"); if (c) return c; } return ""; }
    function show(k, seek) {                                        // idempotent: every state from k alone
      if (k === n) return; n = k;
      var f = frames[k];
      if (seek) fig.classList.add("is-seek");
      [].forEach.call(fig.querySelectorAll(".rp__frame.is-on"), function (el) { if (el !== f) el.classList.remove("is-on"); });
      f.classList.add("is-on");
      ["state", "alert", "over", "look"].forEach(function (a) { fig.setAttribute("data-" + a, f.getAttribute("data-" + a)); });
      var a = bag(f, "data-a"); Object.keys(a).forEach(function (k) { fig.setAttribute("data-" + k, a[k]); });   // the scene's own states
      if (out.time) out.time.textContent = f.getAttribute("data-clock");
      if (out.plus) out.plus.textContent = f.getAttribute("data-plus");
      if (out.axis) out.axis.textContent = f.getAttribute("data-axis");
      if (out.view) out.view.textContent = f.getAttribute("data-view");
      if (out.cap) out.cap.textContent = caption(k);
      var r = bag(f, "data-r"); outs.forEach(function (el) { var v = r[el.getAttribute("data-read")]; el.textContent = v == null ? "" : v; });   // what each part says at this frame
      var word = fig.getAttribute("data-word-" + f.getAttribute("data-state")) || "";
      states.forEach(function (el) { el.textContent = word; });
      if (mark) mark.style.left = f.getAttribute("data-x") + "%";
      range.value = String(k); range.setAttribute("aria-valuetext", f.getAttribute("data-clock") + " " + caption(k));
      if (seek) { void fig.offsetWidth; fig.classList.remove("is-seek"); }
    }
    function label() { btn.textContent = btn.getAttribute(raf ? "data-pause" : n >= frames.length - 1 ? "data-replay-label" : "data-play"); }
    function tick(now) {
      var T = total(); t = Math.min(T, t + Math.min(now - last, 100)); last = now;   // a throttled tab never jumps ahead
      show(at(t));
      if (t < T) raf = requestAnimationFrame(tick); else { raf = 0; label(); }
    }
    function play() {
      if (raf) return;
      if (n >= frames.length - 1) { t = 0; n = -1; show(0, true); }
      started = true; last = performance.now(); raf = requestAnimationFrame(tick); label();
    }
    function stop() { if (raf) cancelAnimationFrame(raf); raf = 0; label(); }
    function seek(k) { stop(); held = true; t = startOf(k); n = -1; show(k, true); label(); }
    function nextMoment() { for (var i = n + 1; i < frames.length; i++) if (frames[i].getAttribute("data-say")) return i; return 0; }

    btn.addEventListener("click", function () {
      if (still) { seek(nextMoment()); return; }                    // reduced motion: the button steps to the next moment
      if (raf) { held = true; stop(); } else { held = false; play(); }
    });
    range.addEventListener("input", function () { seek(Number(range.value)); });
    tabs.forEach(function (b) { b.addEventListener("click", function () {
      stop(); load(b.getAttribute("data-scn")); t = 0; n = -1; show(0, true);
      if (still) { held = true; label(); } else { held = false; play(); }
    }); });

    var k0 = Number(range.value) || 0;
    load(scn); fig.classList.add("is-ready"); n = -1; show(k0, true); t = startOf(k0); label();
    if (still || typeof IntersectionObserver !== "function") return;
    new IntersectionObserver(function (e) {
      if (e[0].isIntersecting) { if (!held && (!started || n < frames.length - 1)) { if (!started) { t = 0; n = -1; show(0, true); } play(); } }
      else stop();
    }, { threshold: 0.6 }).observe(fig.querySelector(".rp__feed") || fig);   // the feed is what must be in view; the figure is taller than a small screen
    document.addEventListener("visibilitychange", function () { if (document.hidden) stop(); });
  });
})();
