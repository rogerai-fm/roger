/* =====================================================================
   RogerAI - the talking crew (company-crew.css .pg__hit / .pg__say; CompanyFigure.svelte).
   Each Ping in the company figure has a button, prerendered hidden so a page without this script
   shows no dead control. This shows them; a tap puts that Ping's line in the bubble (a polite live
   region) and marks the figure data-said=<key>, which the sheet reads to hop that Ping and place
   the bubble over it. A second tap, Escape, a tap elsewhere or eight seconds closes it.
   ===================================================================== */
const fig = document.querySelector(".pg");
const say = fig && fig.querySelector(".pg__say");
if (fig && say) {
  let timer = 0;
  const close = () => { clearTimeout(timer); say.hidden = true; delete fig.dataset.said; };
  const open = (b) => {
    clearTimeout(timer);
    delete fig.dataset.said;
    void fig.offsetWidth;                          // restart the hop when the same Ping is tapped again
    say.querySelector("b").textContent = b.dataset.who;
    say.querySelector("span").textContent = b.dataset.say;
    fig.dataset.said = b.dataset.key;
    say.hidden = false;
    timer = setTimeout(close, 8000);
  };
  fig.querySelectorAll(".pg__hit").forEach((b) => {
    b.hidden = false;
    b.addEventListener("click", (e) => {
      e.stopPropagation();
      if (fig.dataset.said === b.dataset.key && !say.hidden) close(); else open(b);
    });
  });
  document.addEventListener("click", (e) => { if (!say.hidden && !say.contains(e.target)) close(); });
  document.addEventListener("keydown", (e) => { if (e.key === "Escape" && !say.hidden) close(); });
}
