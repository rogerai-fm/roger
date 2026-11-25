/* =====================================================================
   RogerAI - the bar slider (the report figure's one live row).

   Auto-initializes every [data-barslider] group (the figure's own bar row,
   which carries [data-bar-row] too): its range control, laid over the row's
   track with its thumb as the tick, sets the customer's bar: the row's
   .rf__bar tick follows (the still picture's tick; hidden under the thumb
   with JS), the row's data-met flips by comparing data-value (what the run
   read) with the bar, the bar number and any <output> say the bar, and the
   control's aria-valuetext says it in words. Without JS the control is
   hidden (html.js gates it) and the row shows the stated bar.
   First use: the report figure on the custom-tuning page (ReportFigure.svelte).
   ===================================================================== */
(function () {
  "use strict";
  Array.prototype.forEach.call(document.querySelectorAll("[data-barslider]"), function (set) {
    var fig = set.closest("[data-report]");
    var range = set.querySelector("input[type=range]");
    var out = set.querySelector("output");
    var row = fig && fig.querySelector("[data-bar-row]");
    if (!fig || !range || !row) return;
    var tick = row.querySelector(".rf__bar");
    var num = row.querySelector(".rf__barnum");
    var value = Number(row.getAttribute("data-value"));
    function set_(v) {
      v = Number(v);
      if (tick) { tick.setAttribute("x1", String(v)); tick.setAttribute("x2", String(v)); }
      row.setAttribute("data-met", value >= v ? "yes" : "no");
      if (num) num.textContent = v + "%";
      if (out) out.textContent = v + "%";
      range.setAttribute("aria-valuetext", v + " percent, " + (value >= v ? "met" : "not met"));
    }
    range.addEventListener("input", function () { set_(range.value); });
    set_(range.value);
  });
})();
