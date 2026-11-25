/* =====================================================================
   RogerAI - the film plate (components.css .film).

   Auto-initializes every <video data-film> on the page. The markup ships the
   film with a poster and no src, so a page costs the poster and nothing more
   until the plate is on screen; then this sets the src from data-film-src,
   plays it muted in a loop, and pauses it again when it scrolls away. Under
   prefers-reduced-motion, without JS, or without IntersectionObserver the
   poster is the picture and nothing moves.
   First use: the custom-tuning hero (research-custom-tuning.html).
   ===================================================================== */
(function () {
  "use strict";
  var still = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
  if (still || typeof IntersectionObserver !== "function") return;
  Array.prototype.forEach.call(document.querySelectorAll("video[data-film]"), function (v) {
    var src = v.getAttribute("data-film-src");
    if (!src) return;
    v.muted = true; v.loop = true; v.setAttribute("playsinline", "");
    var io = new IntersectionObserver(function (entries) {
      entries.forEach(function (e) {
        if (e.isIntersecting) {
          if (!v.getAttribute("src")) v.setAttribute("src", src);
          var p = v.play(); if (p && p.catch) p.catch(function () {});
        } else if (!v.paused) {
          v.pause();
        }
      });
    }, { threshold: 0.25 });
    io.observe(v);
  });
})();
