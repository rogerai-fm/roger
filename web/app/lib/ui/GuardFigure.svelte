<script>
  import IsoFigure from "#lib/ui/IsoFigure.svelte";
  // The guard rails, drawn (pricing §3, 2026-11-21; in the isometric hand 2026-11-22): one price
  // scale per million tokens out as a long slab. Its front half is the request's lane: thin
  // plates stand across it at $10 (the ceiling you did not set) and $20 (where you type the
  // price back), a wall closes it at $100 (the broker lists nothing above), and the gate, your
  // ceiling, is the lit plate on the lane that slides when you raise it. Its back half is the
  // shelf where the stations stand as tag blocks at the price each one asked. The four cards
  // under it say the same four things in words; the figure shows what each one does to a
  // request, which runs the lane as a comet to the station that answers.
  //
  // Written the way app/lib/ui figures are (TuningFigure.svelte): runes, props in, no <style>
  // block (its rules are src/styles/guard-figure.css, a sheet the page lists; the hand's tones
  // and lines are the .iso__* rules of components.css), no colour literal, no inline style,
  // every word in HTML over the stage, the drawing in inline SVG with no <text>. The SVG
  // between the iso markers is written by computer-scientist/article/fig_guard_gen.py
  // (--svelte), which also prints the two slides and the label percents in guard-figure.css.
  // Lead class .gf. The step list is <ol data-scrub data-scrub-auto>, so scrub.js plays the
  // four steps itself until the reader touches it, and the picture reads the tuned step
  // through the figure's data-step (scrub.js). data-draw: the figure draws itself in on first view (draw.js). Reduced
  // motion, no JS and print get the still: step 1, the $10 ceiling holding.
  //
  // The scale is schematic (the caption says so): $10, $20 and $100 are spaced to read, not
  // to measure. Every dollar figure here is a policy constant of ours (manual §6), or the
  // reader's own example ceiling; never a price a station controls.
  //
  // GUARDS it meets as written: no <img>; no colour literal; every rule in guard-figure.css
  // led by .gf; no SVG <text>; role="group" named by its caption; nothing interactive under
  // aria-hidden; no heading inside; minmax(0, ...) columns; --d-* and --motion only; no em dash.
  let {
    fig = 2,
    caption = "A schematic scale of the price per million tokens out. Your ceiling starts at $10 without you setting it; above $20 you type the number; the broker lists nothing above $100; and your ceiling rides every request, so a station that drifts past it is routed around, not charged.",
  } = $props();
  const steps = [
    { name: "A ceiling you did not set", note: "Set no limit and every connection still stops at $10. The $14 station is never dialed." },
    { name: "Type the price back", note: "Raise it past $20 and you type the number, not y. Now the $14 station can answer." },
    { name: "A listing ceiling", note: "A station asking $140 is refused at the broker. Nothing lists above $100, public or private." },
    { name: "No price hike, ever", note: "Your ceiling rides every request. The $14 station drifts to $30 and is routed around, not charged." },
  ];
</script>

<IsoFigure lead="gf" {fig} {caption} {steps} stepsClass="gf__steps" stepsLabel="The four guard rails, one at a time">
  <div class="iso__stage gf__stage">
    <span class="gf__unit">per 1M tokens out</span>

    <div class="gf__pic" aria-hidden="true">
      <!-- iso:begin -->
      <svg class="iso__svg" viewBox="-31.2 -6 219.8 81.3" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-6.2 2.8 L167 52.8 L148 69.3 L-25.2 19.3 Z"/>
<g class="gf__floor">
<path class="iso__face iso__face--s1" d="M-17 14.7 L144.1 61.2 L144.1 64.7 L-17 18.2 Z"/>
<path class="iso__face iso__face--s2" d="M161.1 46.5 L144.1 61.2 L144.1 64.7 L161.1 50 Z"/>
<path class="iso__face" d="M0 0 L161.1 46.5 L144.1 61.2 L-17 14.7 Z"/>
<path class="iso__ln iso__ln--faint" d="M-5 8.4 L150.8 53.4"/>
<path class="iso__ln iso__ln--lo" d="M-6.1 13.3 L-10.1 16.7"/>
<path class="iso__ln iso__ln--lo" d="M39 26.3 L35 29.7"/>
<path class="iso__ln iso__ln--lo" d="M73.6 36.3 L69.6 39.7"/>
<path class="iso__ln iso__ln--lo" d="M146.3 57.3 L142.3 60.7"/>
</g>
<g class="iso__m gf__m gf__block gf__block--a">
<path class="iso__halo iso__halo--part" d="M25.3 13.1 L23.1 14.4 L20.3 15.4 L16.9 15.8 L13.5 15.7 L10.3 15.1 L7.6 14 L5.7 12.6 L4.8 10.9 L5 9.2 L6.3 7.6 L8.4 6.2 L11.3 5.3 L14.6 4.9 L18.1 5 L21.3 5.6 L24 6.6 L25.8 8.1 L26.7 9.8 L26.5 11.5 Z"/>
<path class="iso__face iso__face--s1" d="M9 6.9 L17.6 9.4 L17.6 13.7 L9 11.2 Z"/>
<path class="iso__face iso__face--s2" d="M22.6 5.1 L17.6 9.4 L17.6 13.7 L22.6 9.4 Z"/>
<path class="iso__face" d="M14 2.6 L22.6 5.1 L17.6 9.4 L9 6.9 Z"/>
<path class="iso__face iso__face--cav" d="M15.9 4.6 L15.4 4.9 L14.7 4.9 L14.1 4.7 L13.8 4.4 L14 4.1 L14.5 3.9 L15.1 3.8 L15.7 4 L16 4.3 Z"/>
<circle class="iso__lamp" cx="15.9" cy="10.3" r="1.3"/>
</g>
<g class="iso__m gf__m gf__block gf__block--b">
<path class="iso__halo iso__halo--part" d="M70.3 26.1 L68.2 27.4 L65.3 28.4 L62 28.8 L58.5 28.7 L55.3 28.1 L52.6 27 L50.8 25.6 L49.9 23.9 L50.1 22.2 L51.3 20.6 L53.5 19.2 L56.3 18.3 L59.7 17.9 L63.1 18 L66.3 18.6 L69 19.6 L70.9 21.1 L71.8 22.8 L71.6 24.5 Z"/>
<path class="iso__face iso__face--s1" d="M54 19.9 L62.6 22.4 L62.6 26.7 L54 24.2 Z"/>
<path class="iso__face iso__face--s2" d="M67.6 18.1 L62.6 22.4 L62.6 26.7 L67.6 22.4 Z"/>
<path class="iso__face" d="M59 15.6 L67.6 18.1 L62.6 22.4 L54 19.9 Z"/>
<path class="iso__face iso__face--cav" d="M60.9 17.6 L60.4 17.9 L59.7 17.9 L59.1 17.7 L58.8 17.4 L59 17.1 L59.5 16.9 L60.2 16.8 L60.8 17 L61 17.3 Z"/>
<circle class="iso__lamp" cx="60.9" cy="23.3" r="1.3"/>
</g>
<g class="iso__m gf__m gf__block gf__block--c">
<path class="iso__halo iso__halo--part" d="M139.6 46.1 L137.5 47.4 L134.6 48.4 L131.2 48.8 L127.8 48.7 L124.6 48.1 L121.9 47 L120.1 45.6 L119.2 43.9 L119.3 42.2 L120.6 40.6 L122.7 39.2 L125.6 38.3 L129 37.9 L132.4 38 L135.6 38.6 L138.3 39.6 L140.1 41.1 L141 42.8 L140.9 44.5 Z"/>
<path class="iso__face iso__face--s1" d="M123.3 39.9 L131.9 42.4 L131.9 46.7 L123.3 44.2 Z"/>
<path class="iso__face iso__face--s2" d="M136.9 38.1 L131.9 42.4 L131.9 46.7 L136.9 42.4 Z"/>
<path class="iso__face" d="M128.3 35.6 L136.9 38.1 L131.9 42.4 L123.3 39.9 Z"/>
<path class="iso__face iso__face--cav" d="M130.2 37.6 L129.7 37.9 L129 37.9 L128.4 37.7 L128.1 37.4 L128.3 37.1 L128.8 36.9 L129.5 36.8 L130 37 L130.3 37.3 Z"/>
<circle class="iso__lamp" cx="130.2" cy="43.3" r="1.3"/>
</g>
<g class="gf__rule gf__rule--10">
<path class="iso__face iso__face--s1" d="M34.9 22.2 L36 22.5 L36 29.4 L34.9 29.1 Z"/>
<path class="iso__face iso__face--s2" d="M43.5 16 L36 22.5 L36 29.4 L43.5 22.9 Z"/>
<path class="iso__face" d="M42.4 15.7 L43.5 16 L36 22.5 L34.9 22.2 Z"/>
<circle class="iso__lamp" cx="36.2" cy="21.7" r="1.4"/>
</g>
<g class="gf__rule gf__rule--20">
<path class="iso__face iso__face--s1" d="M69.6 32.2 L70.6 32.5 L70.6 39.4 L69.6 39.1 Z"/>
<path class="iso__face iso__face--s2" d="M78.1 26 L70.6 32.5 L70.6 39.4 L78.1 32.9 Z"/>
<path class="iso__face" d="M77.1 25.7 L78.1 26 L70.6 32.5 L69.6 32.2 Z"/>
<circle class="iso__lamp" cx="70.9" cy="31.7" r="1.4"/>
</g>
<g class="gf__rule gf__rule--100">
<path class="iso__face iso__face--s1" d="M142.8 48.2 L144.6 48.7 L144.6 60.8 L142.8 60.3 Z"/>
<path class="iso__face iso__face--s2" d="M160.6 34.8 L144.6 48.7 L144.6 60.8 L160.6 46.9 Z"/>
<path class="iso__face" d="M158.8 34.3 L160.6 34.8 L144.6 48.7 L142.8 48.2 Z"/>
<circle class="iso__lamp" cx="144.7" cy="47.5" r="1.5"/>
</g>
<g class="iso__m gf__m gf__block gf__block--d">
<path class="iso__halo iso__halo--part" d="M181.2 58.1 L179 59.4 L176.1 60.4 L172.8 60.8 L169.4 60.7 L166.2 60.1 L163.5 59 L161.6 57.6 L160.7 55.9 L160.9 54.2 L162.1 52.6 L164.3 51.2 L167.2 50.3 L170.5 49.9 L174 50 L177.2 50.6 L179.8 51.6 L181.7 53.1 L182.6 54.8 L182.4 56.5 Z"/>
<path class="iso__face iso__face--s1" d="M164.8 51.9 L173.5 54.4 L173.5 58.7 L164.8 56.2 Z"/>
<path class="iso__face iso__face--s2" d="M178.5 50.1 L173.5 54.4 L173.5 58.7 L178.5 54.4 Z"/>
<path class="iso__face" d="M169.8 47.6 L178.5 50.1 L173.5 54.4 L164.8 51.9 Z"/>
<path class="iso__face iso__face--cav" d="M171.7 49.6 L171.2 49.9 L170.6 49.9 L170 49.7 L169.7 49.4 L169.8 49.1 L170.3 48.9 L171 48.8 L171.6 49 L171.9 49.3 Z"/>
<circle class="iso__lamp" cx="171.8" cy="55.3" r="1.3"/>
<path class="gf__stop" d="M170.4 48.4 L173 53.2"/>
<path class="gf__stop" d="M176.5 50.2 L166.9 51.5"/>
</g>
<path class="iso__way gf__way gf__way--a" d="M-8.1 14.5 L5.8 18.5 L12.8 12.4"/>
<path class="iso__way gf__way gf__way--b" d="M-8.1 14.5 L50.8 31.5 L57.8 25.4"/>
<path class="comet iso__comet gf__comet gf__comet--a" d="M-8.1 14.5 L5.8 18.5 L12.8 12.4"/>
<path class="comet iso__comet gf__comet gf__comet--b" d="M-8.1 14.5 L50.8 31.5 L57.8 25.4"/>
<g class="iso__m gf__m gf__gate">
<path class="iso__halo iso__halo--part" d="M46.8 28.5 L45 29.6 L42.6 30.4 L39.9 30.7 L37.1 30.7 L34.5 30.2 L32.3 29.3 L30.7 28.1 L30 26.7 L30.2 25.3 L31.2 24 L32.9 22.9 L35.3 22.1 L38 21.8 L40.8 21.9 L43.5 22.4 L45.6 23.2 L47.2 24.4 L47.9 25.8 L47.8 27.2 Z"/>
<path class="iso__face iso__face--s1" d="M34.4 20 L35.6 20.4 L35.6 29.9 L34.4 29.5 Z"/>
<path class="iso__face iso__face--s2" d="M43.1 13.9 L35.6 20.4 L35.6 29.9 L43.1 23.4 Z"/>
<path class="iso__face" d="M41.9 13.5 L43.1 13.9 L35.6 20.4 L34.4 20 Z"/>
<circle class="iso__lamp iso__lamp--on" cx="36" cy="19.3" r="1.6"/>
</g>
      </svg>
      <!-- iso:end -->

      <!-- your ceiling: its name above the gate at rest; the typed number once you raise it -->
      <code class="gf__typed">24</code>
      <span class="iso__label gf__gatelabel">your ceiling</span>

      <!-- the stations' prices, above their blocks on the shelf -->
      <b class="gf__price gf__price--a">$4</b>
      <b class="gf__price gf__price--b">$14</b><b class="gf__price gf__price--b gf__later">$30</b>
      <b class="gf__price gf__price--c">$60</b>
      <b class="gf__price gf__price--d">$140</b>

      <ol class="gf__marks">
        <li class="gf__mark gf__mark--0">$0</li>
        <li class="gf__mark gf__mark--10">$10 <small>the default</small></li>
        <li class="gf__mark gf__mark--20">$20 <small>type it back</small></li>
        <li class="gf__mark gf__mark--100">$100 <small>nothing above</small></li>
      </ol>
    </div>
  </div>

</IsoFigure>
