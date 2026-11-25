<script>
  import IsoFigure from "#lib/ui/IsoFigure.svelte";
  // FIG. 3 of the custom-tuning page (2026-11-14): the top of one real report, drawn. Five measures, each a
  // hairline track filled to what this run read, with a red tick where the customer's bar sits and the
  // verdict beside it. The first row's bar is the customer's to set: a range control lies over that row's
  // track (JS only, hidden without it; src/js/barslider.js) and its thumb IS the red tick, so dragging the
  // tick moves the bar and flips the verdict, and a reader sees that the report opens with THEIR bar, met
  // or not met. Rows and numbers are the page's own (RESULTS-MATRIX R.CT1, Wave
  // Nano, work orders, 547 held-out records). Written like TuningFigure.svelte: runes, props in, no <style>
  // (rules in src/styles/custom-tuning.css, lead class .rf), no colour literal, no inline style (the bar's
  // position is an SVG attribute), words in HTML, no SVG <text>.
  let {
    fig = 3,
    model = "Wave Nano",
    job = "work orders",
    records = "547",
    rows = [
      { name: "every field right", bar: 70, value: 77.9, met: true, yours: true },
      { name: "hands off other people's chat requests", bar: 95, value: 99.7, met: true },
      { name: "hands off your off-task questions", bar: 80, value: 91.7, met: true },
      { name: "hands off attempts to make it do something else", bar: 80, value: 80.0, met: true, note: "within sampling error; could flip on a re-run" },
      { name: "in-scope requests handed off by mistake", bar: 1, value: 0.0, met: true, atMost: true },
    ],
    caption = "The report opens with your bar, met or not met, on records the model never saw.",
    hint = "The first row's bar is yours to set: drag the red tick, or focus it and use the arrow keys. The verdict follows.",
  } = $props();
  const fmt = (v) => v.toFixed(1);
</script>

<IsoFigure lead="rf" {fig} {caption} data-report="">
  <p class="rf__head mono"><b>The top of one real report</b> &middot; {model} &middot; {job} &middot; {records} unseen records</p>
  <div class="rf__cols mono" aria-hidden="true"><span>measure</span><span>your bar</span><span></span><span>this run</span><span>met</span></div>
  <ol class="rf__rows">
    {#each rows as r}
      <li class="rf__row" class:rf__row--yours={r.yours} data-met={r.met ? "yes" : "no"} data-bar-row={r.yours ? "" : undefined} data-value={r.yours ? r.value : undefined} data-barslider={r.yours ? "" : undefined}>
        <span class="rf__name">{r.name}{#if r.atMost}<small>&nbsp;(at most)</small>{/if}</span>
        <span class="rf__barnum mono">{r.bar}%</span>
        <span class="rf__track">
          <svg viewBox="0 0 100 10" preserveAspectRatio="none" aria-hidden="true" focusable="false">
            <line class="rf__rail" x1="0" y1="5" x2="100" y2="5"/>
            <rect class="rf__fill" x="0" y="3" width={r.value} height="4"/>
            <line class="rf__bar" x1={r.bar} y1="0" x2={r.bar} y2="10"/>
          </svg>
          {#if r.yours}
            <!-- the control lies over the track; its thumb is the tick (0 to 100 so the thumb and the svg share one scale) -->
            <input class="rf__range" id="rf-bar-{fig}" type="range" min="0" max="100" step="1" value={r.bar} aria-label="Your bar for {r.name}, in percent">
          {/if}
        </span>
        <span class="rf__num mono" data-count={fmt(r.value)} data-count-suffix="%">{fmt(r.value)}%</span>
        <span class="rf__verdict"><span data-when="yes">met</span><span data-when="no">not met</span>{#if r.note} <small>{r.note}</small>{/if}</span>
      </li>
    {/each}
  </ol>
  <p class="rf__hint mono">{hint}</p>
</IsoFigure>
