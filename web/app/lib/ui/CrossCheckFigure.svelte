<script>
  import IsoFigure from "#lib/ui/IsoFigure.svelte";
  // FIG. 3 of the oil and gas page (2026-11-21): the cross-check. Four rows a pipeline already
  // records, on one day ruler: flow in at the head, flow out at the tail, the tank level, and
  // the clocks the three were stamped with. The scrubbed list plays a day: 1 the four agree;
  // 2 flow out falls while flow in holds, the mark says where; 3 one clock drifts from the
  // others, the timestamps stop lining up; 4 the flag, one line with the rule it broke, for a
  // person to take to the site. The model ranks and explains; it never settles a bill and never
  // names a person, and the page says so.
  //
  // Same contract as UseCaseFigure.svelte: runes, no <style> (oil-gas.css, lead .xf), no colour
  // literal, no inline style, no SVG <text>, role="group" named by its caption, data-draw.
  // Still (no JS, print, reduced motion): step 1, a day where the four agree.
  let {
    fig = 4,
    caption = "Four records a line already keeps, on one day. When flow out falls while flow in holds, or one clock drifts from the others, the model writes one line that says which rule broke and where. A person decides what it means.",
    steps = [
      { name: "A normal day", note: "head, tail, tank and clocks agree" },
      { name: "A tap", note: "flow out falls while flow in holds" },
      { name: "A skew", note: "one clock drifts from the other two" },
      { name: "The flag", note: "one line, the rule it broke, where" },
    ],
  } = $props();
</script>

<IsoFigure lead="xf" {fig} {caption} steps={steps} stepsClass="fig-stops xf__steps" stepsLabel="One day on the line, in four steps">
  <div class="xf__stage">
    <ul class="xf__rows" aria-hidden="true">
      <li class="og__name xf__row">Flow in</li>
      <li class="og__name xf__row">Flow out</li>
      <li class="og__name xf__row">Tank level</li>
      <li class="og__name xf__row">Clocks</li>
    </ul>

    <svg class="xf__svg" viewBox="0 0 300 100" preserveAspectRatio="none" aria-hidden="true" focusable="false">
      <!-- the day ruler -->
      <path class="xf__ruler" d="M4 94 H296 M4 90 V98 M150 90 V98 M296 90 V98"/>
      <!-- the row baselines -->
      <path class="xf__base" d="M4 18 H296 M4 40 H296 M4 62 H296"/>

      <!-- flow in: steady all day -->
      <polyline class="xf__trace xf__trace--in" points="4,14 60,15 120,13 180,14 240,13 296,14"/>
      <!-- flow out: steady, then (step 2) it falls at the mark and stays low -->
      <polyline class="xf__trace xf__trace--out" points="4,36 60,37 120,35 180,36 240,35 296,36"/>
      <polyline class="xf__trace xf__trace--out-tap" points="4,36 60,37 120,35 168,36 176,46 240,45 296,46"/>
      <!-- tank level: a gentle fill, then (step 2) it fills slower than the head says it should -->
      <polyline class="xf__trace xf__trace--tank" points="4,66 80,63 160,60 240,57 296,55"/>
      <polyline class="xf__trace xf__trace--tank-tap" points="4,66 80,63 160,60 176,60 240,59 296,58"/>
      <!-- the clocks: three ticks a stamp, in step (left) and (step 3) one drifting -->
      <g class="xf__clock xf__clock--ok">
        <path class="xf__tick" d="M60 76 V84 M150 76 V84 M240 76 V84"/>
        <path class="xf__tick" d="M60 78 V86 M150 78 V86 M240 78 V86"/>
      </g>
      <g class="xf__clock xf__clock--skew">
        <path class="xf__tick" d="M60 76 V84 M150 76 V84 M240 76 V84"/>
        <path class="xf__tick xf__tick--off" d="M68 78 V86 M164 78 V86 M262 78 V86"/>
      </g>

      <!-- the marks: where the tap shows, and the skew -->
      <path class="xf__mark xf__mark--tap" d="M172 30 V50"/>
      <circle class="xf__lamp xf__lamp--tap" cx="172" cy="46" r="2.6"/>
      <path class="xf__mark xf__mark--skew" d="M164 72 V88"/>
      <circle class="xf__lamp xf__lamp--skew" cx="164" cy="82" r="2.6"/>
    </svg>

    <ol class="xf__days" aria-hidden="true">
      <li>0h</li><li>12h</li><li>24h</li>
    </ol>

    <!-- the flag: one line, the rule, the place -->
    <div class="xf__flag" aria-hidden="true">
      <i class="xf__dot"></i>
      <b>Flag</b>
      <span>out fell, in held, from 13:40 at km 42</span>
      <small>rule: in and out must agree within the line&rsquo;s pack</small>
    </div>
  </div>

</IsoFigure>
