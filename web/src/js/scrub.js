/* =====================================================================
   RogerAI - the scrubber component (components.css .scrub).

   Auto-initializes every list marked [data-scrub] (an <ol>/<ul>, complete
   and readable without JS): it inserts a native range control above the
   list; drag it, or arrow through it, and the item under the needle is
   tuned (.is-tuned - the page styles what tuned looks like). Pointing at an
   item moves the scrubber to it. It starts on the first item, or on the item
   data-scrub-start names (a zero-based index) when the list is laid out whole;
   a sideways gallery always starts on its first item, which is the one it shows. Every item
   stays fully readable; tuning only marks one, it never fades the others.
   The control's accessible name is the list's aria-label; each position
   reads out the item's <b> text.
   When a page lays the list out as a sideways gallery (it overflows its own
   width, e.g. the homepage spectrum on phones), the two stay in step:
   scrubbing brings the tuned item to the middle, and swiping the gallery
   tunes the item that comes to rest in the middle.
   data-scrub-auto (2026-11-14): the list plays itself while it is in view, one
   item every 2.8 s, so a figure tells its story without a drag; the first
   pointer, key or swipe from the reader stops it for good. A figure with more
   to read on each stop names its own pace, data-scrub-ms (2026-11-25). Off under
   prefers-reduced-motion and without IntersectionObserver.
   Markup contract: web/DESIGN-SYSTEM.md.
   ===================================================================== */
(function () {
  "use strict";
  Array.prototype.forEach.call(document.querySelectorAll("[data-scrub]"), init);

  function init(list) {
    var items = list.querySelectorAll("li");
    if (!items.length) return;

    var range = document.createElement("input");
    range.type = "range";
    range.min = "0";
    range.max = String(items.length - 1);
    range.step = "1";
    range.value = "0";
    range.className = "scrub";
    range.style.setProperty("--scrub-n", String(items.length));
    range.setAttribute("aria-label", list.getAttribute("aria-label") || "");

    // the figure the list plays (if any) carries the tuned step as data-step="1".."n", so its sheet can read
    // the state as [data-step="3"] rather than a :has(li:nth-child(3).is-tuned) chain (2026-11-24). Without
    // JS the attribute is absent: a sheet's still picture is :not([data-step]).
    var host = typeof list.closest === "function" ? list.closest("figure") : null;
    function tune(k) {
      range.value = String(k);
      for (var i = 0; i < items.length; i++) items[i].classList.toggle("is-tuned", i === k);
      if (host) host.setAttribute("data-step", String(k + 1));
      var name = items[k].querySelector("b");
      range.setAttribute("aria-valuetext", name ? name.textContent : String(k));
    }

    // the sideways gallery: only when the list overflows its own width
    function overflows() {
      return typeof list.scrollTo === "function" && list.scrollWidth > list.clientWidth;
    }
    function center(k) {
      if (!overflows()) return;
      var li = items[k];
      var still = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
      list.scrollTo({ left: li.offsetLeft - (list.clientWidth - li.offsetWidth) / 2, behavior: still ? "auto" : "smooth" });
    }
    function nearest() {
      var mid = list.scrollLeft + list.clientWidth / 2, best = 0, gap = Infinity;
      for (var i = 0; i < items.length; i++) {
        var d = Math.abs(items[i].offsetLeft + items[i].offsetWidth / 2 - mid);
        if (d < gap) { gap = d; best = i; }
      }
      return best;
    }

    range.addEventListener("input", function () { tune(Number(range.value)); center(Number(range.value)); });
    Array.prototype.forEach.call(items, function (li, k) {
      li.addEventListener("pointerenter", function () { tune(k); center(k); });
    });
    // a swipe tunes where it comes to rest (after the scroll settles, so a
    // smooth scroll the scrubber started never re-tunes on its way past)
    if (typeof list.addEventListener === "function") {
      var settle = null;
      list.addEventListener("scroll", function () {
        if (!overflows()) return;
        clearTimeout(settle);
        settle = setTimeout(function () { tune(nearest()); }, 120);
      }, { passive: true });
    }
    list.parentNode.insertBefore(range, list);
    // data-scrub-start="k": the item tuned on load (default the first). A sideways
    // gallery ignores it: it shows its first item on load and loading never scrolls
    // it, so the needle starts on that first item too (the needle and the view agree)
    var start = parseInt(list.getAttribute("data-scrub-start"), 10);
    tune(start >= 0 && start < items.length && !overflows() ? start : 0);
    if (list.getAttribute("data-scrub-auto") != null) autoplay();

    // the self-playing figure: advances while in view, stops at the reader's first touch
    function autoplay() {
      var still = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
      if (still || typeof IntersectionObserver !== "function") return;
      var timer = null, stopped = false;
      function stop() { stopped = true; clearInterval(timer); timer = null; }
      function step() { tune((Number(range.value) + 1) % items.length); center(Number(range.value)); }
      var ms = Number(list.getAttribute("data-scrub-ms")) || 2800;
      function play() { if (!stopped && !timer) timer = setInterval(step, ms); }
      function pause() { clearInterval(timer); timer = null; }
      ["pointerdown", "keydown", "touchstart"].forEach(function (ev) {
        range.addEventListener(ev, stop, { passive: true });
        list.addEventListener(ev, stop, { passive: true });
      });
      new IntersectionObserver(function (entries) {
        entries.forEach(function (e) { if (e.isIntersecting) play(); else pause(); });
      }, { threshold: 0.4 }).observe(list);
    }
  }
})();
