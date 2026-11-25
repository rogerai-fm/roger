<script>
  import IsoFigure from "#lib/ui/IsoFigure.svelte";
  // The life of an earning, drawn (pricing §4, 2026-11-21; in the isometric hand 2026-11-22): a
  // day lane (0, 30, 90) as a long slab, the earning as a coin on it at day 0 (a thick disc, your
  // principal, with a thin wafer on it, the 10% reserve slice), and the bank as a box at the lane's
  // end. The page's own lifecycle list (<ol class="lifecycle" data-scrub data-scrub-auto>, passed
  // in as children so the page keeps its words and its contract) is the figure's step list: the
  // picture reads the tuned stage through the figure's data-step (scrub.js). 1 Accrued: the coin lands at day 0. 2 Held: the
  // wafer lifts off to its own row, both parts float on their halos, the hold spans show (to day
  // 30 in front, to day 90 behind). 3 Payable: each part lands on its day mark, lit. 4 Paid: both
  // slide into the bank and its lamp lights.
  //
  // Written the way app/lib/ui figures are (TuningFigure.svelte): runes, props in, no <style>
  // block (its rules are src/styles/hold-figure.css, a sheet the page lists; the hand's tones and
  // lines are the .iso__* rules of components.css), no colour literal, no inline style, every
  // word in HTML over the stage, the drawing in inline SVG with no <text>. The SVG between the
  // iso markers is written by computer-scientist/article/fig_hold_gen.py (--svelte), which also
  // prints every translate and label percent in hold-figure.css. Lead class .hf. data-draw: the
  // figure draws itself in on first view (draw.js). Reduced motion, no JS and print get the
  // still: stage 1, the earning just credited. The days and the 10% slice are the payout policy
  // (manual §7), never a price a station controls.
  //
  // GUARDS it meets as written: no <img>; no colour literal; every rule in hold-figure.css
  // led by .hf; no SVG <text>; role="group" named by its caption; nothing interactive under
  // aria-hidden; no heading inside; minmax(0, ...) columns; --d-* and --motion only; no em dash.
  let {
    fig = 3,
    caption = "An earning is credited the moment a request settles. The principal is held 30 days and a 10% slice to day 90. Then it is payable, and paid to your bank through Stripe when you ask.",
    children,
  } = $props();
</script>

<IsoFigure lead="hf" {fig} {caption}>
  <div class="iso__stage hf__stage">
    <div class="hf__pic" aria-hidden="true">
      <span class="iso__label hf__who hf__who--main">the principal</span>
      <span class="iso__label hf__who hf__who--slice">a 10% slice</span>

      <!-- iso:begin -->
      <svg class="iso__svg" viewBox="-27.2 -6.2 231.4 87" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-6.2 2.8 L198.2 61.8 L183.2 74.8 L-21.2 15.8 Z"/>
<g class="hf__lane">
<path class="iso__face iso__face--s1" d="M-13 11.3 L179.3 66.8 L179.3 70.2 L-13 14.7 Z"/>
<path class="iso__face iso__face--s2" d="M192.3 55.5 L179.3 66.8 L179.3 70.2 L192.3 59 Z"/>
<path class="iso__face" d="M0 0 L192.3 55.5 L179.3 66.8 L-13 11.3 Z"/>
<path class="iso__ln iso__ln--lo" d="M7.7 3.4 L-3.3 12.9"/>
<path class="iso__ln iso__ln--lo" d="M54.4 16.9 L43.4 26.4"/>
<path class="iso__ln iso__ln--lo" d="M148 43.9 L137 53.4"/>
</g>
<path class="hf__hold hf__hold--main" d="M-0.3 10 L46.4 23.5"/>
<path class="hf__hold hf__hold--slice" d="M4.7 5.7 L145 46.2"/>
<path class="iso__way hf__way hf__way--main" d="M46.4 23.3 L164.2 57.3"/>
<path class="iso__way hf__way hf__way--slice" d="M145 45.9 L169.2 52.9"/>
<path class="comet iso__comet hf__comet hf__comet--main" d="M46.4 23.3 L164.2 57.3"/>
<path class="comet iso__comet hf__comet hf__comet--slice" d="M145 45.9 L169.2 52.9"/>
<g class="iso__m hf__m hf__m--main">
<path class="iso__halo iso__halo--part" d="M9.6 13.2 L7.8 14.4 L5.4 15.3 L2.6 15.8 L-0.3 16 L-3.3 15.8 L-6.1 15.3 L-8.5 14.4 L-10.3 13.2 L-11.4 11.8 L-11.8 10.3 L-11.4 8.8 L-10.3 7.4 L-8.5 6.2 L-6.1 5.3 L-3.3 4.7 L-0.3 4.5 L2.6 4.7 L5.4 5.3 L7.8 6.2 L9.6 7.4 L10.8 8.8 L11.2 10.3 L10.8 11.8 Z"/>
<path class="iso__face iso__face--s1" d="M-6.8 4.2 L-6.8 10.3 L-6.8 10.8 L-6.5 11.3 L-6.1 11.8 L-5.6 12.2 L-4.9 12.6 L-4.2 12.9 L-3.3 13.2 L-2.3 13.4 L-1.4 13.5 L-0.3 13.5 L0.7 13.5 L1.7 13.4 L2.6 13.2 L3.5 12.9 L4.3 12.6 L4.9 12.2 L5.5 11.8 L5.8 11.3 L6.1 10.8 L6.2 10.3 L6.2 4.2 Z"/>
<path class="iso__face" d="M5.3 5.9 L4.3 6.5 L2.9 7 L1.3 7.4 L-0.3 7.5 L-2 7.4 L-3.6 7 L-4.9 6.5 L-6 5.9 L-6.6 5.1 L-6.8 4.2 L-6.6 3.4 L-6 2.6 L-4.9 1.9 L-3.6 1.4 L-2 1.1 L-0.3 1 L1.3 1.1 L2.9 1.4 L4.3 1.9 L5.3 2.6 L5.9 3.4 L6.2 4.2 L5.9 5.1 Z"/>
<circle class="iso__lamp" cx="0.6" cy="6" r="1.3"/>
</g>
<g class="iso__m hf__m hf__m--slice">
<path class="iso__halo iso__halo--part" d="M9.6 13.2 L7.8 14.4 L5.4 15.3 L2.6 15.8 L-0.3 16 L-3.3 15.8 L-6.1 15.3 L-8.5 14.4 L-10.3 13.2 L-11.4 11.8 L-11.8 10.3 L-11.4 8.8 L-10.3 7.4 L-8.5 6.2 L-6.1 5.3 L-3.3 4.7 L-0.3 4.5 L2.6 4.7 L5.4 5.3 L7.8 6.2 L9.6 7.4 L10.8 8.8 L11.2 10.3 L10.8 11.8 Z"/>
<path class="iso__face iso__face--s1" d="M-6.8 3 L-6.8 4.2 L-6.8 4.7 L-6.5 5.2 L-6.1 5.7 L-5.6 6.1 L-4.9 6.5 L-4.2 6.9 L-3.3 7.1 L-2.3 7.3 L-1.4 7.4 L-0.3 7.5 L0.7 7.4 L1.7 7.3 L2.6 7.1 L3.5 6.9 L4.3 6.5 L4.9 6.1 L5.5 5.7 L5.8 5.2 L6.1 4.7 L6.2 4.2 L6.2 3 Z"/>
<path class="iso__face" d="M5.3 4.6 L4.3 5.3 L2.9 5.8 L1.3 6.2 L-0.3 6.3 L-2 6.2 L-3.6 5.8 L-4.9 5.3 L-6 4.6 L-6.6 3.9 L-6.8 3 L-6.6 2.2 L-6 1.4 L-4.9 0.7 L-3.6 0.2 L-2 -0.1 L-0.3 -0.2 L1.3 -0.1 L2.9 0.2 L4.3 0.7 L5.3 1.4 L5.9 2.2 L6.2 3 L5.9 3.9 Z"/>
<circle class="iso__lamp" cx="0.6" cy="4.8" r="1.3"/>
</g>
<g class="hf__bank">
<path class="iso__halo" d="M189.2 62.1 L186.4 64 L182.6 65.4 L178.3 66.3 L173.6 66.6 L169 66.3 L164.6 65.4 L160.9 64 L158 62.1 L156.2 60 L155.6 57.6 L156.2 55.3 L158 53.1 L160.9 51.3 L164.6 49.8 L169 48.9 L173.6 48.6 L178.3 48.9 L182.6 49.8 L186.4 51.3 L189.2 53.1 L191 55.3 L191.6 57.6 L191 60 Z"/>
<path class="iso__face iso__face--s1" d="M157.7 45.5 L178.5 51.5 L178.5 65.4 L157.7 59.4 Z"/>
<path class="iso__face iso__face--s2" d="M189.5 42 L178.5 51.5 L178.5 65.4 L189.5 55.9 Z"/>
<path class="iso__face" d="M168.7 36 L189.5 42 L178.5 51.5 L157.7 45.5 Z"/>
<path class="iso__face iso__face--s1" d="M156.4 44.3 L178.9 50.8 L178.9 52.2 L156.4 45.7 Z"/>
<path class="iso__face iso__face--s2" d="M190.9 40.4 L178.9 50.8 L178.9 52.2 L190.9 41.8 Z"/>
<path class="iso__face" d="M168.4 33.9 L190.9 40.4 L178.9 50.8 L156.4 44.3 Z"/>
<path class="iso__face iso__face--cav" d="M170.3 40.3 L179 42.8 L177.7 43.9 L169.1 41.4 Z"/>
<path class="iso__ln iso__ln--faint" d="M161.2 56.9 L170.7 59.7"/>
<path class="iso__ln iso__ln--faint" d="M161.2 54.3 L170.7 57.1"/>
<path class="iso__ln iso__ln--faint" d="M161.2 51.7 L170.7 54.5"/>
<circle class="iso__lamp" cx="175.1" cy="53.6" r="1.6"/>
</g>
      </svg>
      <!-- iso:end -->

      <!-- the shares, riding above their parts -->
      <b class="hf__chip hf__chip--main">90%</b>
      <b class="hf__chip hf__chip--slice">10%</b>

      <ol class="hf__days">
        <li class="hf__day hf__day--0">day 0</li>
        <li class="hf__day hf__day--30">day 30</li>
        <li class="hf__day hf__day--90">day 90</li>
        <li class="hf__day hf__day--bank">your bank</li>
      </ol>
    </div>
  </div>

  {@render children?.()}
</IsoFigure>
