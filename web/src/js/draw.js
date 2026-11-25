/* =====================================================================
   RogerAI - the drawn-in figure (custom-tuning.css, the [data-draw] rules).

   Auto-initializes every [data-draw] figure. Each SVG stroke in it learns its
   own on-screen length (--len, in CSS pixels: the shape's length times the
   scale its svg is shown at; the figures stroke with vector-effect
   non-scaling-stroke, so dashes are measured on screen, not in the viewBox),
   the figure gets .is-drawing (every line hidden behind a dash as long as
   itself, fills clear), and when it first comes into view .is-drawn: the
   page's CSS then draws each line from nothing to whole and fades the fills
   up, part by part. A few seconds later .is-settled lets the dash rules go,
   so the figure's own state transitions run on clean strokes. Shapes that
   carry a dash pattern of their own (dashed cables, the ways that draw
   themselves by state, the comets) keep it and only receive --len, which the
   comets use for their travelling dash. A number marked data-count counts up
   from zero to its value on the same cue.
   Under prefers-reduced-motion, without IntersectionObserver, or without JS
   nothing here runs and the figure is complete from the first paint.
   First use: the five custom-tuning figures (2026-11-15).
   ===================================================================== */
(function () {
  "use strict";
  var still = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  if (still || typeof IntersectionObserver !== "function") return;
  var figs = document.querySelectorAll("[data-draw]");
  if (!figs.length) return;

  function screenLength(el) {
    if (typeof el.getTotalLength !== "function") return 0;
    var l = 0; try { l = el.getTotalLength(); } catch (e) { return 0; }
    var m = el.getScreenCTM && el.getScreenCTM();
    var s = m ? Math.sqrt(Math.abs(m.a * m.d - m.b * m.c)) : 1;   /* the geometric mean of the two axes' scale */
    return l * s * 1.06;                                           /* a little over, so a stretched axis never leaves a gap */
  }

  function prime(fig) {
    var shapes = fig.querySelectorAll("svg :is(path, line, rect, circle, polyline, ellipse)");
    Array.prototype.forEach.call(shapes, function (el) {
      var len = screenLength(el);
      if (len > 0) el.style.setProperty("--len", len.toFixed(1) + "px");
      var dash = getComputedStyle(el).strokeDasharray;
      if (el.hasAttribute("pathLength") || (dash && dash !== "none" && dash !== "0px")) el.setAttribute("data-draw-skip", "");
    });
    fig.classList.add("is-drawing");
  }

  function count(el) {
    var to = parseFloat(el.getAttribute("data-count"));
    if (isNaN(to)) return;
    var dp = (el.getAttribute("data-count").split(".")[1] || "").length;
    var suffix = el.getAttribute("data-count-suffix") || "";
    var t0 = null, dur = 900;
    function frame(t) {
      if (t0 === null) t0 = t;
      var k = Math.min(1, (t - t0) / dur); k = 1 - Math.pow(1 - k, 3);
      el.textContent = (to * k).toFixed(dp) + suffix;
      if (k < 1) requestAnimationFrame(frame);
    }
    requestAnimationFrame(frame);
  }

  var io = new IntersectionObserver(function (entries) {
    entries.forEach(function (e) {
      if (!e.isIntersecting) return;
      var fig = e.target;
      io.unobserve(fig);
      fig.classList.add("is-drawn");
      Array.prototype.forEach.call(fig.querySelectorAll("[data-count]"), count);
      setTimeout(function () { fig.classList.add("is-settled"); }, 3200);
    });
  }, { threshold: 0.3 });

  Array.prototype.forEach.call(figs, function (fig) { prime(fig); io.observe(fig); });
})();
