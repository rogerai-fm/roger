/* =====================================================================
   RogerAI - the key switch (a figure's two-way state).

   Auto-initializes every [data-keyswitch] group: its buttons carry data-key, the
   figure they belong to (the nearest [data-keyed]) carries the same attribute, and
   the page's CSS draws each state. Without JS the group is hidden (html.js gates it)
   and the figure shows its default state, which is the still picture.
   First use: the private-band figure on the custom-tuning page (PrivateBandFigure.svelte).
   ===================================================================== */
(function () {
  "use strict";
  Array.prototype.forEach.call(document.querySelectorAll("[data-keyswitch]"), function (keys) {
    var fig = keys.closest("[data-keyed]");
    if (!fig) return;
    var buttons = keys.querySelectorAll("[data-key]");
    keys.addEventListener("click", function (e) {
      var b = e.target.closest("[data-key]");
      if (!b || !keys.contains(b)) return;
      fig.setAttribute("data-key", b.getAttribute("data-key"));
      Array.prototype.forEach.call(buttons, function (x) { x.setAttribute("aria-pressed", String(x === b)); });
    });
  });
})();
