<script>
  import IsoFigure from "#lib/ui/IsoFigure.svelte";
  // FIG. 2 of the oil and gas page (2026-11-21; redrawn as an isometric solid 2026-11-22): one
  // pipe, six ways to look at it, and what each one hands the model. A section of ground cut
  // open, the line on the trench floor, and the instruments where they work: a pig inside (seen
  // through a cutaway window), a crawler on the wall, a guided-wave collar, a pulsed eddy current
  // probe standing on the insulated span, a drone above, a fibre clipped along the floor. The
  // model's console sits on the far bank. The scrubbed list tunes a method: it lifts off the
  // paper, its lamp lights, its record tag shows, and a comet carries that record along the bank
  // to the console. The model never reads the raw signal; it reads the record the method's own
  // software writes (a tally, a thickness grid, a trace, a map, frames, an event log).
  //
  // The drawing is generated (computer-scientist/article/fig2_iso_gen.py, between the markers):
  // one camera, solids filled with paper and painted back to front so what is in front hides
  // what is behind, tones by facing, three line weights of the one ink, one red accent. No
  // filter, no script beyond draw.js. Same contract as UseCaseFigure.svelte: runes, no <style>
  // (oil-gas.css, lead .pf), no colour literal, no inline style, no SVG <text>, role="group"
  // named by its caption, data-draw. Still (no JS, print, reduced motion): the pig lit,
  // "an anomaly tally" tagged.
  let {
    fig = 3,
    caption = "Six ways to look at one pipe, from the inside and from outside. We advise on which instrument and which robot fits the line; the model reads the record each one already writes, never the raw signal.",
    methods = [
      { name: "A pig, inside", note: "an anomaly tally, run against the last run" },
      { name: "A crawler or a hand probe", note: "a thickness grid, cell by cell" },
      { name: "A guided-wave collar", note: "a reflection trace, tens of metres a ring" },
      { name: "A probe through insulation", note: "a wall map, no insulation removed" },
      { name: "A drone above", note: "thermal frames, flagged spots to open" },
      { name: "A fibre alongside", note: "an event log: digging, tapping, a leak" },
    ],
  } = $props();
</script>

<IsoFigure lead="pf" {fig} {caption} steps={methods} stepsClass="fig-stops pf__methods" stepsLabel="Six inspection methods, one at a time">
  <div class="iso__stage">
    <!-- iso:begin -->
    <svg class="iso__svg" viewBox="-44.2 -6 283 144.5" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-6.2 35.7 L232.8 104.7 L200.8 132.5 L-38.2 63.5 Z"/>
<g class="pf__ground">
<path class="iso__face" d="M0 0 L226.9 65.5 L216.9 74.2 L-10 8.7 Z"/>
<path class="iso__face iso__face--s1" d="M-10 8.7 L216.9 74.2 L216.9 94.9 L-10 29.4 Z"/>
<path class="iso__ln iso__ln--faint" d="M-10 15.6 L216.9 81.1"/>
<path class="iso__ln iso__ln--faint" d="M-10 22.5 L216.9 88"/>
<path class="iso__face iso__face--s1" d="M-10 29.4 L216.9 94.9 L196.9 112.3 L-30 46.8 Z"/>
<path class="iso__face iso__face--s2" d="M-10 8.7 L-30 26 L-30 46.8 L-10 29.4 Z"/>
<path class="iso__face iso__face--s2" d="M226.9 65.5 L216.9 74.2 L216.9 94.9 L196.9 112.3 L196.9 127.9 L226.9 101.9 Z"/>
<path class="iso__face iso__face--s1" d="M-30 46.8 L196.9 112.3 L196.9 127.9 L-30 62.4 Z"/>
<path class="iso__ln iso__ln--faint" d="M-30 52 L196.9 117.5"/>
<path class="iso__ln iso__ln--faint" d="M-30 57.2 L196.9 122.7"/>
<path class="iso__ln iso__ln--faint" d="M-21.3 49.3 L-21.3 50.7"/>
<path class="iso__ln iso__ln--faint" d="M-12.7 51.8 L-12.7 53.2"/>
<path class="iso__ln iso__ln--faint" d="M-4 54.3 L-4 55.7"/>
<path class="iso__ln iso__ln--faint" d="M4.6 56.8 L4.6 58.2"/>
<path class="iso__ln iso__ln--lo" d="M13.3 59.3 L13.3 62"/>
<path class="iso__ln iso__ln--faint" d="M22 61.8 L22 63.2"/>
<path class="iso__ln iso__ln--faint" d="M30.6 64.3 L30.6 65.7"/>
<path class="iso__ln iso__ln--faint" d="M39.3 66.8 L39.3 68.2"/>
<path class="iso__ln iso__ln--faint" d="M47.9 69.3 L47.9 70.7"/>
<path class="iso__ln iso__ln--lo" d="M56.6 71.8 L56.6 74.5"/>
<path class="iso__ln iso__ln--faint" d="M65.3 74.3 L65.3 75.7"/>
<path class="iso__ln iso__ln--faint" d="M73.9 76.8 L73.9 78.2"/>
<path class="iso__ln iso__ln--faint" d="M82.6 79.3 L82.6 80.7"/>
<path class="iso__ln iso__ln--faint" d="M91.2 81.8 L91.2 83.2"/>
<path class="iso__ln iso__ln--lo" d="M99.9 84.3 L99.9 87"/>
<path class="iso__ln iso__ln--faint" d="M108.6 86.8 L108.6 88.2"/>
<path class="iso__ln iso__ln--faint" d="M117.2 89.3 L117.2 90.7"/>
<path class="iso__ln iso__ln--faint" d="M125.9 91.8 L125.9 93.2"/>
<path class="iso__ln iso__ln--faint" d="M134.5 94.3 L134.5 95.7"/>
<path class="iso__ln iso__ln--lo" d="M143.2 96.8 L143.2 99.5"/>
<path class="iso__ln iso__ln--faint" d="M151.9 99.3 L151.9 100.7"/>
<path class="iso__ln iso__ln--faint" d="M160.5 101.8 L160.5 103.2"/>
<path class="iso__ln iso__ln--faint" d="M169.2 104.3 L169.2 105.7"/>
<path class="iso__ln iso__ln--faint" d="M177.8 106.8 L177.8 108.2"/>
<path class="iso__ln iso__ln--lo" d="M186.5 109.3 L186.5 112"/>
<path class="iso__dot" d="M8.4 7.6 L8 7.8 L7.4 7.8 L7 7.7 L6.7 7.4 L6.8 7.2 L7.3 7 L7.8 7 L8.3 7.1 L8.5 7.4 Z"/>
<path class="iso__dot" d="M-3.4 48.7 L-3.8 48.8 L-4.4 48.9 L-4.9 48.7 L-5.1 48.5 L-5 48.2 L-4.6 48 L-4 48 L-3.5 48.1 L-3.3 48.4 Z"/>
<path class="iso__dot" d="M29.2 13.6 L28.8 13.8 L28.2 13.8 L27.7 13.7 L27.5 13.4 L27.6 13.2 L28 13 L28.6 13 L29.1 13.1 L29.3 13.4 Z"/>
<path class="iso__dot" d="M17.3 54.7 L16.9 54.8 L16.4 54.9 L15.9 54.7 L15.7 54.5 L15.8 54.2 L16.2 54 L16.8 54 L17.2 54.1 L17.5 54.4 Z"/>
<path class="iso__dot" d="M50 19.6 L49.6 19.8 L49 19.8 L48.5 19.7 L48.3 19.4 L48.4 19.2 L48.8 19 L49.4 19 L49.9 19.1 L50.1 19.4 Z"/>
<path class="iso__dot" d="M38.1 60.7 L37.7 60.8 L37.2 60.9 L36.7 60.7 L36.5 60.5 L36.6 60.2 L37 60 L37.5 60 L38 60.1 L38.2 60.4 Z"/>
<path class="iso__dot" d="M70.8 25.6 L70.3 25.8 L69.8 25.8 L69.3 25.7 L69.1 25.4 L69.2 25.2 L69.6 25 L70.2 25 L70.6 25.1 L70.9 25.4 Z"/>
<path class="iso__dot" d="M58.9 66.7 L58.5 66.8 L58 66.9 L57.5 66.7 L57.2 66.5 L57.4 66.2 L57.8 66 L58.3 66 L58.8 66.1 L59 66.4 Z"/>
<path class="iso__dot" d="M91.5 31.6 L91.1 31.8 L90.6 31.8 L90.1 31.7 L89.9 31.4 L90 31.2 L90.4 31 L90.9 31 L91.4 31.1 L91.7 31.4 Z"/>
<path class="iso__dot" d="M79.7 72.7 L79.3 72.8 L78.7 72.9 L78.3 72.7 L78 72.5 L78.1 72.2 L78.6 72 L79.1 72 L79.6 72.1 L79.8 72.4 Z"/>
<path class="iso__dot" d="M112.3 37.6 L111.9 37.8 L111.4 37.8 L110.9 37.7 L110.7 37.4 L110.8 37.2 L111.2 37 L111.7 37 L112.2 37.1 L112.4 37.4 Z"/>
<path class="iso__dot" d="M100.5 78.7 L100.1 78.8 L99.5 78.9 L99 78.7 L98.8 78.5 L98.9 78.2 L99.3 78 L99.9 78 L100.4 78.1 L100.6 78.4 Z"/>
<path class="iso__dot" d="M133.1 43.6 L132.7 43.8 L132.1 43.8 L131.7 43.7 L131.4 43.4 L131.6 43.2 L132 43 L132.5 43 L133 43.1 L133.2 43.4 Z"/>
<path class="iso__dot" d="M121.3 84.7 L120.9 84.8 L120.3 84.9 L119.8 84.7 L119.6 84.5 L119.7 84.2 L120.1 84 L120.7 84 L121.2 84.1 L121.4 84.4 Z"/>
<path class="iso__dot" d="M153.9 49.6 L153.5 49.8 L152.9 49.8 L152.4 49.7 L152.2 49.4 L152.3 49.2 L152.8 49 L153.3 49 L153.8 49.1 L154 49.4 Z"/>
<path class="iso__dot" d="M142.1 90.7 L141.6 90.8 L141.1 90.9 L140.6 90.7 L140.4 90.5 L140.5 90.2 L140.9 90 L141.5 90 L141.9 90.1 L142.2 90.4 Z"/>
<path class="iso__dot" d="M174.7 55.6 L174.3 55.8 L173.7 55.8 L173.2 55.7 L173 55.4 L173.1 55.2 L173.5 55 L174.1 55 L174.6 55.1 L174.8 55.4 Z"/>
<path class="iso__dot" d="M162.8 96.7 L162.4 96.8 L161.9 96.9 L161.4 96.7 L161.2 96.5 L161.3 96.2 L161.7 96 L162.2 96 L162.7 96.1 L163 96.4 Z"/>
<path class="iso__dot" d="M195.5 61.6 L195.1 61.8 L194.5 61.8 L194 61.7 L193.8 61.4 L193.9 61.2 L194.3 61 L194.9 61 L195.4 61.1 L195.6 61.4 Z"/>
<path class="iso__dot" d="M183.6 102.7 L183.2 102.8 L182.7 102.9 L182.2 102.7 L182 102.5 L182.1 102.2 L182.5 102 L183 102 L183.5 102.1 L183.7 102.4 Z"/>
<path class="iso__dot" d="M216.2 67.6 L215.8 67.8 L215.3 67.8 L214.8 67.7 L214.6 67.4 L214.7 67.2 L215.1 67 L215.7 67 L216.1 67.1 L216.4 67.4 Z"/>
<path class="iso__dot" d="M204.4 108.7 L204 108.8 L203.4 108.9 L203 108.7 L202.7 108.5 L202.9 108.2 L203.3 108 L203.8 108 L204.3 108.1 L204.5 108.4 Z"/>
</g>
<g class="pf__dock">
<path class="iso__halo iso__halo--part" d="M219.3 69.1 L216.3 71 L212.3 72.6 L207.8 73.5 L202.8 73.8 L197.9 73.5 L193.3 72.6 L189.4 71 L186.4 69.1 L184.5 66.8 L183.8 64.3 L184.5 61.9 L186.4 59.6 L189.4 57.6 L193.3 56.1 L197.9 55.2 L202.8 54.8 L207.8 55.2 L212.3 56.1 L216.3 57.6 L219.3 59.6 L221.2 61.9 L221.8 64.3 L221.2 66.8 Z"/>
<path class="iso__face iso__face--s1" d="M187.2 48.3 L211.5 55.3 L211.5 70.9 L187.2 63.9 Z"/>
<path class="iso__face iso__face--s2" d="M218.5 49.2 L211.5 55.3 L211.5 70.9 L218.5 64.8 Z"/>
<path class="iso__face" d="M194.2 42.2 L218.5 49.2 L211.5 55.3 L187.2 48.3 Z"/>
<path class="iso__face iso__face--s1" d="M190.8 46.4 L209.9 51.9 L209.9 53.7 L190.8 48.2 Z"/>
<path class="iso__face iso__face--s2" d="M214.9 47.6 L209.9 51.9 L209.9 53.7 L214.9 49.3 Z"/>
<path class="iso__face" d="M195.8 42.1 L214.9 47.6 L209.9 51.9 L190.8 46.4 Z"/>
<path class="iso__face iso__face--cav" d="M191.6 54.7 L207.1 59.2 L207.1 57.8 L191.6 53.3 Z"/>
<path class="iso__face" d="M192.4 54.3 L187.9 59.7 L201.8 63.7 L206.3 58.3 Z"/>
<path class="iso__ln iso__ln--lo" d="M193.1 56 L203.5 59"/>
<path class="iso__ln iso__ln--lo" d="M191.4 58.1 L200.1 60.6"/>
<path class="iso__ln iso__ln--faint" d="M192.4 61.9 L206.3 65.9"/>
<path class="iso__ln iso__ln--faint" d="M192.4 59.3 L206.3 63.3"/>
<path class="iso__ln" d="M211.5 49.5 L211.5 40.8"/>
<circle class="iso__lamp iso__lamp--on" cx="211.5" cy="40.8" r="1.4"/>
<circle class="iso__lamp iso__lamp--on" cx="193.3" cy="52.2" r="1.9"/>
</g>
<path class="iso__face" d="M-15.5 19.1 L69.4 43.6 L68.9 43.5 L68.5 43.5 L68 43.7 L67.5 44 L67 44.3 L66.5 44.8 L66 45.4 L65.5 46 L65 46.7 L64.6 47.5 L64.1 48.4 L63.8 49.3 L63.4 50.2 L63.1 51.2 L62.9 52.1 L62.7 53.1 L62.5 54.1 L62.4 55 L62.4 55.9 L62.4 56.8 L62.5 57.6 L62.6 58.3 L62.8 59 L63 59.6 L63.3 60 L63.6 60.4 L64 60.7 L64.4 60.9 L-20.5 36.4 L-20.9 36.2 L-21.3 35.9 L-21.6 35.5 L-21.9 35.1 L-22.1 34.5 L-22.3 33.8 L-22.4 33.1 L-22.5 32.3 L-22.5 31.4 L-22.5 30.5 L-22.4 29.6 L-22.2 28.6 L-22 27.6 L-21.8 26.7 L-21.5 25.7 L-21.1 24.8 L-20.7 23.9 L-20.3 23 L-19.9 22.2 L-19.4 21.5 L-18.9 20.9 L-18.4 20.3 L-17.9 19.8 L-17.4 19.5 L-16.9 19.2 L-16.4 19 L-16 19 Z"/>
<path class="iso__tone iso__tone--s2" d="M-22.5 31.8 L62.4 56.3 L62.4 57 L62.5 57.7 L62.6 58.3 L62.7 58.9 L62.9 59.4 L63.2 59.8 L63.4 60.2 L63.7 60.5 L64 60.7 L64.4 60.9 L-20.5 36.4 L-20.8 36.2 L-21.2 36 L-21.4 35.7 L-21.7 35.3 L-21.9 34.9 L-22.1 34.4 L-22.3 33.8 L-22.4 33.2 L-22.5 32.5 Z"/>
<path class="iso__ln iso__ln--faint" d="M-18.5 20.4 L66.4 44.9"/>
<clipPath id="pf-win"><path class="" d="M17.7 29.2 L62.8 42.2 L62.2 42.5 L61.7 42.9 L61.2 43.4 L60.6 44 L60.1 44.7 L59.6 45.5 L59.2 46.4 L58.7 47.3 L58.4 48.3 L58 49.3 L57.7 50.4 L57.5 51.4 L12.5 38.4 L12.7 37.4 L13 36.3 L13.3 35.3 L13.7 34.3 L14.1 33.4 L14.6 32.5 L15.1 31.7 L15.6 31 L16.1 30.4 L16.7 29.9 L17.2 29.5 Z"/></clipPath>
<path class="iso__face iso__face--cav" d="M17.7 29.2 L62.8 42.2 L62.2 42.5 L61.7 42.9 L61.2 43.4 L60.6 44 L60.1 44.7 L59.6 45.5 L59.2 46.4 L58.7 47.3 L58.4 48.3 L58 49.3 L57.7 50.4 L57.5 51.4 L12.5 38.4 L12.7 37.4 L13 36.3 L13.3 35.3 L13.7 34.3 L14.1 33.4 L14.6 32.5 L15.1 31.7 L15.6 31 L16.1 30.4 L16.7 29.9 L17.2 29.5 Z"/>
<g clip-path="url(#pf-win)"><g class="iso__m pf__m--1">
<path class="iso__halo iso__halo--part" d="M47 51.1 L45 52.4 L42.4 53.2 L39.3 53.6 L36.2 53.5 L33.3 53 L30.9 52 L29.2 50.7 L28.3 49.2 L28.5 47.6 L29.6 46.1 L31.6 44.9 L34.2 44.1 L37.2 43.7 L40.4 43.7 L43.3 44.3 L45.7 45.3 L47.4 46.6 L48.2 48.1 L48.1 49.7 Z"/>
<path class="iso__face" d="M28.5 35.5 L51 42 L50.8 42 L50.5 42 L50.2 42.1 L49.9 42.3 L49.6 42.5 L49.3 42.8 L49 43.1 L48.7 43.5 L48.4 43.9 L48.2 44.4 L47.9 44.9 L47.7 45.4 L47.5 46 L47.3 46.6 L47.1 47.2 L47 47.8 L46.9 48.3 L46.9 48.9 L46.9 49.4 L46.9 50 L46.9 50.4 L47 50.9 L47.1 51.3 L47.2 51.6 L47.4 51.9 L47.6 52.1 L47.8 52.3 L48.1 52.4 L25.5 45.9 L25.3 45.8 L25.1 45.6 L24.9 45.4 L24.7 45.1 L24.6 44.8 L24.5 44.4 L24.4 43.9 L24.3 43.5 L24.3 42.9 L24.4 42.4 L24.4 41.8 L24.5 41.3 L24.6 40.7 L24.8 40.1 L25 39.5 L25.2 38.9 L25.4 38.4 L25.7 37.9 L25.9 37.4 L26.2 37 L26.5 36.6 L26.8 36.3 L27.1 36 L27.4 35.8 L27.7 35.6 L28 35.5 L28.3 35.5 Z"/>
<path class="iso__face" d="M46.8 49.6 L46.9 48.7 L47 47.8 L47.2 46.9 L47.5 46 L47.8 45.1 L48.2 44.3 L48.6 43.6 L49.1 43 L49.5 42.5 L50 42.2 L50.5 42 L50.9 42 L51.3 42.1 L51.6 42.4 L51.9 42.8 L52.1 43.4 L52.2 44.1 L52.2 44.9 L52.2 45.7 L52.1 46.6 L51.9 47.5 L51.6 48.4 L51.3 49.3 L50.9 50.1 L50.5 50.8 L50 51.4 L49.5 51.9 L49.1 52.2 L48.6 52.4 L48.2 52.4 L47.8 52.3 L47.5 52 L47.2 51.6 L47 51 L46.9 50.3 Z"/>
<path class="iso__face" d="M27.3 44.4 L27.5 42.9 L27.8 41.3 L28.3 39.8 L28.9 38.4 L29.7 37.1 L30.5 36.3 L31.3 35.7 L32.1 35.6 L32.7 35.9 L33.2 36.6 L33.5 37.7 L33.6 39 L33.5 40.5 L33.2 42.1 L32.7 43.6 L32.1 45.1 L31.3 46.3 L30.5 47.2 L29.7 47.7 L28.9 47.8 L28.3 47.5 L27.8 46.8 L27.5 45.8 Z"/>
<path class="iso__face" d="M34.3 46.4 L34.4 44.9 L34.7 43.3 L35.2 41.8 L35.9 40.4 L36.6 39.1 L37.4 38.3 L38.2 37.7 L39 37.6 L39.7 37.9 L40.2 38.6 L40.5 39.7 L40.6 41 L40.5 42.5 L40.2 44.1 L39.7 45.6 L39 47.1 L38.2 48.3 L37.4 49.2 L36.6 49.7 L35.9 49.8 L35.2 49.5 L34.7 48.8 L34.4 47.8 Z"/>
<path class="iso__face" d="M42.9 48.9 L43 47.4 L43.4 45.8 L43.9 44.3 L44.5 42.9 L45.3 41.6 L46.1 40.8 L46.9 40.2 L47.7 40.1 L48.3 40.4 L48.8 41.1 L49.1 42.2 L49.2 43.5 L49.1 45 L48.8 46.6 L48.3 48.1 L47.7 49.6 L46.9 50.8 L46.1 51.7 L45.3 52.2 L44.5 52.3 L43.9 52 L43.4 51.3 L43 50.3 Z"/>
<path class="iso__face" d="M50.5 49.3 L50.6 48.6 L50.8 47.8 L51 47 L51.3 46.3 L51.7 45.6 L52.1 45.2 L52.6 44.9 L52.9 44.9 L53.3 45 L53.5 45.4 L53.7 45.9 L53.7 46.6 L53.7 47.3 L53.5 48.1 L53.3 48.9 L52.9 49.7 L52.6 50.3 L52.1 50.7 L51.7 51 L51.3 51.1 L51 50.9 L50.8 50.5 L50.6 50 Z"/>
<path class="iso__ln iso__ln--lo" d="M49.5 44.4 L52.1 46.8"/>
<path class="iso__ln iso__ln--lo" d="M49.5 50 L52.1 49.2"/>
<circle class="iso__lamp" cx="46.3" cy="44.2" r="1.7"/>
</g></g>
<path class="iso__ln" d="M17.7 29.2 L62.8 42.2 L62.2 42.5 L61.7 42.9 L61.2 43.4 L60.6 44 L60.1 44.7 L59.6 45.5 L59.2 46.4 L58.7 47.3 L58.4 48.3 L58 49.3 L57.7 50.4 L57.5 51.4 L12.5 38.4 L12.7 37.4 L13 36.3 L13.3 35.3 L13.7 34.3 L14.1 33.4 L14.6 32.5 L15.1 31.7 L15.6 31 L16.1 30.4 L16.7 29.9 L17.2 29.5"/>
<g class="iso__m pf__m--3">
<path class="iso__face" d="M70.5 39.7 L76.5 41.5 L75.9 41.4 L75.2 41.4 L74.5 41.7 L73.8 42 L73.1 42.6 L72.4 43.2 L71.6 44.1 L70.9 45 L70.3 46 L69.6 47.2 L69 48.4 L68.4 49.7 L67.9 51.1 L67.5 52.4 L67.1 53.9 L66.8 55.3 L66.6 56.7 L66.5 58 L66.4 59.3 L66.5 60.6 L66.6 61.7 L66.7 62.8 L67 63.7 L67.3 64.6 L67.8 65.3 L68.2 65.8 L68.8 66.2 L69.3 66.5 L63.3 64.7 L62.7 64.5 L62.2 64.1 L61.7 63.5 L61.3 62.8 L60.9 62 L60.7 61 L60.5 60 L60.4 58.8 L60.4 57.6 L60.4 56.3 L60.6 54.9 L60.8 53.5 L61.1 52.1 L61.5 50.7 L61.9 49.3 L62.4 48 L62.9 46.7 L63.5 45.4 L64.2 44.3 L64.9 43.2 L65.6 42.3 L66.3 41.5 L67 40.8 L67.8 40.3 L68.5 39.9 L69.2 39.7 L69.8 39.6 Z"/>
<path class="iso__tone iso__tone--s2" d="M60.4 58.1 L66.4 59.9 L66.5 60.9 L66.6 61.9 L66.8 62.8 L67 63.6 L67.2 64.3 L67.6 65 L68 65.5 L68.4 65.9 L68.8 66.3 L69.3 66.5 L63.3 64.7 L62.8 64.5 L62.3 64.2 L61.9 63.8 L61.5 63.2 L61.2 62.6 L60.9 61.9 L60.7 61 L60.5 60.1 L60.4 59.2 Z"/>
<path class="iso__ln iso__ln--faint" d="M66.1 41.7 L72.2 43.4"/>
<path class="iso__face" d="M66.4 59.6 L66.5 57.6 L66.8 55.4 L67.3 53.2 L68 51 L68.8 49 L69.7 47 L70.7 45.3 L71.8 43.9 L72.9 42.7 L74.1 41.9 L75.2 41.5 L76.2 41.4 L77.1 41.7 L77.9 42.4 L78.6 43.5 L79 44.8 L79.3 46.5 L79.4 48.3 L79.3 50.4 L79 52.5 L78.6 54.7 L77.9 56.9 L77.1 59 L76.2 60.9 L75.2 62.6 L74.1 64.1 L72.9 65.2 L71.8 66 L70.7 66.5 L69.7 66.5 L68.8 66.2 L68 65.5 L67.3 64.5 L66.8 63.1 L66.5 61.5 Z"/>
<path class="iso__face iso__face--s1" d="M68.4 57.9 L68.5 56 L68.9 54.1 L69.4 52.2 L70.1 50.3 L71 48.6 L71.9 47.2 L72.9 46.2 L73.9 45.5 L74.9 45.2 L75.7 45.4 L76.5 46.1 L77 47.1 L77.3 48.4 L77.4 50.1 L77.3 51.9 L77 53.8 L76.5 55.8 L75.7 57.6 L74.9 59.3 L73.9 60.7 L72.9 61.8 L71.9 62.4 L71 62.7 L70.1 62.5 L69.4 61.9 L68.9 60.9 L68.5 59.5 Z"/>
<path class="iso__ln iso__ln--faint" d="M65.9 41.8 L71.9 43.6"/>
<path class="iso__ln iso__ln--faint pf__wave pf__wave--1" d="M83.6 46.1 L82.7 46.1 L81.7 46.4 L80.7 47.1 L79.7 48.2 L78.7 49.6 L77.8 51.2 L77 53 L76.3 55 L75.9 57 L75.6 58.9 L75.5 60.8 L75.6 62.5 L75.9 63.9 L76.4 65.1 L77 65.9 L77.8 66.3"/>
<path class="iso__ln iso__ln--faint pf__wave pf__wave--2" d="M92.6 47.7 L91.6 47.6 L90.5 48 L89.4 48.8 L88.2 49.9 L87.2 51.5 L86.2 53.2 L85.3 55.2 L84.6 57.4 L84.1 59.6 L83.7 61.7 L83.6 63.7 L83.8 65.6 L84.1 67.2 L84.6 68.4 L85.3 69.3 L86.2 69.8"/>
<path class="iso__ln iso__ln--faint pf__wave pf__wave--3" d="M101.5 49.2 L100.4 49.1 L99.2 49.6 L98 50.4 L96.8 51.7 L95.6 53.3 L94.6 55.3 L93.6 57.4 L92.8 59.8 L92.3 62.1 L91.9 64.5 L91.8 66.7 L91.9 68.7 L92.3 70.4 L92.9 71.8 L93.6 72.7 L94.6 73.2"/>
<circle class="iso__lamp" cx="65.9" cy="47.9" r="1.7"/>
</g>
<path class="iso__face" d="M75.4 45.3 L116.1 57.1 L115.7 57 L115.2 57 L114.7 57.2 L114.2 57.5 L113.7 57.8 L113.2 58.3 L112.7 58.9 L112.3 59.5 L111.8 60.2 L111.3 61 L110.9 61.9 L110.5 62.8 L110.2 63.7 L109.9 64.7 L109.6 65.6 L109.4 66.6 L109.3 67.6 L109.2 68.5 L109.1 69.4 L109.2 70.3 L109.2 71.1 L109.4 71.8 L109.5 72.5 L109.8 73.1 L110.1 73.5 L110.4 73.9 L110.7 74.2 L111.2 74.4 L70.4 62.6 L70 62.4 L69.7 62.2 L69.3 61.8 L69.1 61.3 L68.8 60.7 L68.7 60.1 L68.5 59.3 L68.5 58.5 L68.4 57.7 L68.5 56.8 L68.6 55.8 L68.7 54.9 L68.9 53.9 L69.2 52.9 L69.5 51.9 L69.8 51 L70.2 50.1 L70.6 49.3 L71.1 48.5 L71.5 47.7 L72 47.1 L72.5 46.5 L73 46.1 L73.5 45.7 L74 45.5 L74.5 45.3 L75 45.2 Z"/>
<path class="iso__tone iso__tone--s2" d="M68.4 58.1 L109.1 69.8 L109.2 70.5 L109.2 71.2 L109.4 71.8 L109.5 72.4 L109.7 72.9 L109.9 73.3 L110.2 73.7 L110.5 74 L110.8 74.2 L111.2 74.4 L70.4 62.6 L70.1 62.5 L69.8 62.3 L69.5 62 L69.2 61.6 L69 61.2 L68.8 60.6 L68.7 60.1 L68.5 59.5 L68.5 58.8 Z"/>
<path class="iso__ln iso__ln--faint" d="M72.4 46.7 L113.1 58.4"/>
<path class="pf__pit" d="M87.8 53 L87.5 53.7 L86.7 54.2 L85.6 54.3 L84.5 54.2 L83.7 53.7 L83.4 53 L83.7 52.4 L84.5 51.9 L85.6 51.7 L86.7 51.9 L87.5 52.4 Z"/>
<g class="iso__m pf__m--2">
<path class="iso__halo iso__halo--part" d="M94.6 52.4 L92.2 53.9 L89.1 54.9 L85.4 55.4 L81.7 55.3 L78.2 54.6 L75.3 53.4 L73.2 51.9 L72.3 50 L72.5 48.2 L73.8 46.4 L76.2 45 L79.3 43.9 L82.9 43.5 L86.7 43.5 L90.2 44.2 L93.1 45.4 L95.2 47 L96.1 48.8 L95.9 50.7 Z"/>
<path class="iso__face iso__face--s2" d="M74.7 49.6 L74.8 48.6 L75.1 47.6 L75.5 46.8 L76 46.2 L76.5 45.9 L76.9 46 L77.2 46.5 L77.3 47.3 L77.2 48.3 L76.9 49.2 L76.5 50.1 L76 50.7 L75.5 50.9 L75.1 50.8 L74.8 50.3 Z"/>
<path class="iso__face iso__face--s2" d="M85.1 52.6 L85.2 51.6 L85.5 50.6 L85.9 49.8 L86.4 49.2 L86.9 48.9 L87.3 49 L87.6 49.5 L87.7 50.3 L87.6 51.3 L87.3 52.2 L86.9 53.1 L86.4 53.7 L85.9 53.9 L85.5 53.8 L85.2 53.3 Z"/>
<path class="iso__face iso__face--s1" d="M73 42.2 L90.4 47.2 L90.4 52.4 L73 47.4 Z"/>
<path class="iso__face iso__face--s2" d="M95.4 42.8 L90.4 47.2 L90.4 52.4 L95.4 48 Z"/>
<path class="iso__face" d="M78 37.8 L95.4 42.8 L90.4 47.2 L73 42.2 Z"/>
<path class="iso__ln" d="M84.2 42.5 L84.2 33.8"/>
<path class="iso__face iso__face--s1" d="M80.8 31.7 L85.1 32.9 L85.1 35.5 L80.8 34.3 Z"/>
<path class="iso__face iso__face--s2" d="M87.6 30.8 L85.1 32.9 L85.1 35.5 L87.6 33.4 Z"/>
<path class="iso__face" d="M83.3 29.5 L87.6 30.8 L85.1 32.9 L80.8 31.7 Z"/>
<path class="iso__ln iso__ln--lo" d="M88.6 51 L86.6 58.8"/>
<path class="iso__face" d="M85.8 59.5 L85.9 58.7 L86.2 57.9 L86.6 57.4 L87 57.2 L87.3 57.5 L87.4 58.1 L87.3 58.9 L87 59.6 L86.6 60.2 L86.2 60.3 L85.9 60.1 Z"/>
<circle class="iso__lamp" cx="86.9" cy="47.9" r="1.7"/>
</g>
<g class="pf__sleeve">
<path class="iso__face" d="M117.2 53.2 L191.7 74.7 L191.1 74.6 L190.4 74.7 L189.7 74.9 L189 75.3 L188.3 75.8 L187.5 76.5 L186.8 77.3 L186.1 78.2 L185.4 79.3 L184.8 80.4 L184.2 81.7 L183.6 83 L183.1 84.3 L182.7 85.7 L182.3 87.1 L182 88.5 L181.8 89.9 L181.7 91.3 L181.6 92.6 L181.6 93.8 L181.7 95 L181.9 96 L182.2 97 L182.5 97.8 L182.9 98.5 L183.4 99.1 L183.9 99.5 L184.5 99.7 L110 78.2 L109.5 78 L108.9 77.6 L108.5 77 L108.1 76.3 L107.7 75.5 L107.5 74.5 L107.3 73.5 L107.2 72.3 L107.1 71.1 L107.2 69.8 L107.3 68.4 L107.6 67 L107.8 65.6 L108.2 64.2 L108.7 62.8 L109.1 61.5 L109.7 60.2 L110.3 58.9 L111 57.8 L111.6 56.7 L112.3 55.8 L113.1 55 L113.8 54.3 L114.5 53.8 L115.2 53.4 L115.9 53.2 L116.6 53.1 Z"/>
<path class="iso__tone iso__tone--s2" d="M107.1 71.6 L181.6 93.1 L181.7 94.2 L181.8 95.1 L181.9 96 L182.2 96.9 L182.4 97.6 L182.8 98.2 L183.1 98.8 L183.6 99.2 L184 99.5 L184.5 99.7 L110 78.2 L109.5 78 L109.1 77.7 L108.7 77.3 L108.3 76.7 L108 76.1 L107.7 75.4 L107.5 74.5 L107.3 73.6 L107.2 72.7 Z"/>
<path class="iso__ln iso__ln--faint" d="M112.9 55.2 L187.4 76.7"/>
<path class="iso__face" d="M181.6 92.8 L181.7 90.8 L182 88.7 L182.5 86.5 L183.1 84.3 L183.9 82.2 L184.9 80.3 L185.9 78.6 L187 77.1 L188.1 76 L189.2 75.1 L190.3 74.7 L191.4 74.6 L192.3 75 L193.1 75.7 L193.7 76.7 L194.2 78.1 L194.5 79.7 L194.6 81.6 L194.5 83.6 L194.2 85.8 L193.7 88 L193.1 90.1 L192.3 92.2 L191.4 94.1 L190.3 95.9 L189.2 97.3 L188.1 98.5 L187 99.3 L185.9 99.7 L184.9 99.8 L183.9 99.5 L183.1 98.8 L182.5 97.7 L182 96.4 L181.7 94.7 Z"/>
<path class="iso__face iso__face--s1" d="M183.6 91.1 L183.7 89.3 L184.1 87.3 L184.6 85.4 L185.3 83.5 L186.2 81.9 L187.1 80.5 L188.1 79.4 L189.1 78.7 L190.1 78.5 L190.9 78.7 L191.6 79.3 L192.2 80.3 L192.5 81.7 L192.6 83.3 L192.5 85.1 L192.2 87.1 L191.6 89 L190.9 90.9 L190.1 92.5 L189.1 93.9 L188.1 95 L187.1 95.7 L186.2 95.9 L185.3 95.7 L184.6 95.1 L184.1 94.1 L183.7 92.7 Z"/>
<path class="iso__ln iso__ln--faint" d="M127.6 56.2 L126.3 56.2 L124.9 56.8 L123.5 58 L122 59.7 L120.7 61.9 L119.5 64.5 L118.6 67.2 L117.9 70 L117.6 72.8 L117.6 75.3 L117.8 77.5 L118.4 79.3 L119.3 80.6 L120.4 81.2"/>
<path class="iso__ln iso__ln--faint" d="M138 59.2 L136.7 59.2 L135.3 59.8 L133.8 61 L132.4 62.7 L131.1 64.9 L129.9 67.5 L129 70.2 L128.3 73 L128 75.8 L127.9 78.3 L128.2 80.5 L128.8 82.3 L129.7 83.6 L130.8 84.2"/>
<path class="iso__ln iso__ln--faint" d="M148.4 62.2 L147.1 62.2 L145.7 62.8 L144.2 64 L142.8 65.7 L141.5 67.9 L140.3 70.5 L139.4 73.2 L138.7 76 L138.4 78.8 L138.3 81.3 L138.6 83.5 L139.2 85.3 L140.1 86.6 L141.2 87.2"/>
<path class="iso__ln iso__ln--faint" d="M158.8 65.2 L157.5 65.2 L156.1 65.8 L154.6 67 L153.2 68.7 L151.9 70.9 L150.7 73.5 L149.8 76.2 L149.1 79 L148.8 81.8 L148.7 84.3 L149 86.5 L149.6 88.3 L150.5 89.6 L151.6 90.2"/>
<path class="iso__ln iso__ln--faint" d="M169.2 68.2 L167.9 68.2 L166.5 68.8 L165 70 L163.6 71.7 L162.3 73.9 L161.1 76.5 L160.2 79.2 L159.5 82 L159.2 84.8 L159.1 87.3 L159.4 89.5 L160 91.3 L160.9 92.6 L162 93.2"/>
<path class="iso__ln iso__ln--faint" d="M179.6 71.2 L178.3 71.2 L176.9 71.8 L175.4 73 L174 74.7 L172.7 76.9 L171.5 79.5 L170.6 82.2 L169.9 85 L169.5 87.8 L169.5 90.3 L169.8 92.5 L170.4 94.3 L171.3 95.6 L172.4 96.2"/>
</g>
<g class="iso__m pf__m--4">
<path class="iso__halo iso__halo--part" d="M138.8 61.7 L136.8 62.9 L134.2 63.8 L131.1 64.2 L128 64.1 L125.1 63.5 L122.7 62.6 L121 61.2 L120.1 59.7 L120.3 58.2 L121.4 56.7 L123.4 55.5 L126 54.6 L129 54.2 L132.2 54.3 L135.1 54.9 L137.5 55.9 L139.2 57.2 L140 58.7 L139.9 60.2 Z"/>
<path class="iso__face iso__face--s1" d="M122 54 L134.2 57.5 L134.2 62.7 L122 59.2 Z"/>
<path class="iso__face iso__face--s2" d="M138.2 54 L134.2 57.5 L134.2 62.7 L138.2 59.2 Z"/>
<path class="iso__face" d="M126 50.5 L138.2 54 L134.2 57.5 L122 54 Z"/>
<path class="iso__ln" d="M130.1 54 L130.1 46.2"/>
<path class="iso__face" d="M132.2 46.8 L131.4 47.2 L130.4 47.4 L129.4 47.4 L128.5 47.1 L127.9 46.7 L127.7 46.1 L128 45.6 L128.7 45.2 L129.7 45 L130.8 45.1 L131.7 45.3 L132.3 45.8 L132.5 46.3 Z"/>
<path class="iso__ln iso__ln--faint" d="M123.8 57.5 L132.4 60"/>
<circle class="iso__lamp" cx="131.6" cy="58" r="1.7"/>
</g>
<g class="iso__m pf__m--5">
<path class="iso__cone" d="M175.1 72.6 L173.1 74 L170.4 75 L167.2 75.7 L163.9 75.9 L160.5 75.7 L157.4 75 L154.7 74 L152.6 72.6 L151.3 71.1 L150.9 69.4 L151.3 67.7 L152.6 66.1 L154.7 64.8 L157.4 63.8 L160.5 63.1 L163.9 62.9 L167.2 63.1 L170.4 63.8 L173.1 64.8 L175.1 66.1 L176.4 67.7 L176.9 69.4 L176.4 71.1 Z"/>
<path class="iso__cone" d="M163.9 31.7 L150.9 69.4 L176.9 69.4 Z"/>
<path class="iso__ln iso__ln--faint" d="M175.1 72.6 L173.1 74 L170.4 75 L167.2 75.7 L163.9 75.9 L160.5 75.7 L157.4 75 L154.7 74 L152.6 72.6 L151.3 71.1 L150.9 69.4 L151.3 67.7 L152.6 66.1 L154.7 64.8 L157.4 63.8 L160.5 63.1 L163.9 62.9 L167.2 63.1 L170.4 63.8 L173.1 64.8 L175.1 66.1 L176.4 67.7 L176.9 69.4 L176.4 71.1"/>
<path class="iso__ln iso__ln--lo" d="M163.9 30 L160.6 23.8"/>
<path class="iso__face" d="M164.9 24 L163.8 24.7 L162.3 25.1 L160.6 25.3 L158.9 25.1 L157.4 24.7 L156.2 24 L155.6 23.2 L155.6 22.4 L156.2 21.5 L157.4 20.9 L158.9 20.4 L160.6 20.3 L162.3 20.4 L163.8 20.9 L164.9 21.5 L165.5 22.4 L165.5 23.2 Z"/>
<path class="iso__ln iso__ln--lo" d="M160.6 23.8 L160.6 22.8"/>
<path class="iso__ln iso__ln--lo" d="M163.9 30 L176.2 28.3"/>
<path class="iso__face" d="M180.5 28.5 L179.4 29.2 L177.9 29.6 L176.2 29.8 L174.4 29.6 L172.9 29.2 L171.8 28.5 L171.2 27.7 L171.2 26.9 L171.8 26 L172.9 25.4 L174.4 24.9 L176.2 24.8 L177.9 24.9 L179.4 25.4 L180.5 26 L181.1 26.9 L181.1 27.7 Z"/>
<path class="iso__ln iso__ln--lo" d="M176.2 28.3 L176.2 27.3"/>
<path class="iso__ln iso__ln--lo" d="M163.9 30 L151.6 31.6"/>
<path class="iso__face" d="M155.9 31.8 L154.8 32.5 L153.3 32.9 L151.6 33.1 L149.9 32.9 L148.4 32.5 L147.2 31.8 L146.6 31 L146.6 30.2 L147.2 29.3 L148.4 28.7 L149.9 28.2 L151.6 28.1 L153.3 28.2 L154.8 28.7 L155.9 29.3 L156.5 30.2 L156.5 31 Z"/>
<path class="iso__ln iso__ln--lo" d="M151.6 31.6 L151.6 30.6"/>
<path class="iso__ln iso__ln--lo" d="M163.9 30 L167.2 36.1"/>
<path class="iso__face" d="M171.5 36.3 L170.4 37 L168.9 37.4 L167.2 37.6 L165.4 37.4 L163.9 37 L162.8 36.3 L162.2 35.5 L162.2 34.7 L162.8 33.8 L163.9 33.2 L165.4 32.7 L167.2 32.6 L168.9 32.7 L170.4 33.2 L171.5 33.8 L172.1 34.7 L172.1 35.5 Z"/>
<path class="iso__ln iso__ln--lo" d="M167.2 36.1 L167.2 35.1"/>
<path class="iso__face iso__face--s1" d="M157.5 28.7 L166.2 31.2 L166.2 34.7 L157.5 32.2 Z"/>
<path class="iso__face iso__face--s2" d="M170.2 27.8 L166.2 31.2 L166.2 34.7 L170.2 31.2 Z"/>
<path class="iso__face" d="M161.5 25.3 L170.2 27.8 L166.2 31.2 L157.5 28.7 Z"/>
<circle class="iso__lamp" cx="162.7" cy="32.4" r="1.5"/>
</g>
<path class="iso__face" d="M190.6 78.6 L211.4 84.6 L210.9 84.5 L210.5 84.5 L210 84.7 L209.5 85 L209 85.3 L208.5 85.8 L208 86.4 L207.5 87 L207 87.7 L206.6 88.5 L206.2 89.4 L205.8 90.3 L205.4 91.2 L205.1 92.2 L204.9 93.1 L204.7 94.1 L204.5 95.1 L204.4 96 L204.4 96.9 L204.4 97.8 L204.5 98.6 L204.6 99.3 L204.8 100 L205 100.6 L205.3 101 L205.6 101.4 L206 101.7 L206.4 101.9 L185.6 95.9 L185.2 95.7 L184.9 95.4 L184.5 95 L184.2 94.6 L184 94 L183.8 93.3 L183.7 92.6 L183.6 91.8 L183.6 90.9 L183.7 90 L183.8 89.1 L183.9 88.1 L184.1 87.1 L184.4 86.2 L184.7 85.2 L185 84.3 L185.4 83.4 L185.8 82.5 L186.3 81.7 L186.7 81 L187.2 80.4 L187.7 79.8 L188.2 79.3 L188.7 79 L189.2 78.7 L189.7 78.5 L190.2 78.5 Z"/>
<path class="iso__tone iso__tone--s2" d="M183.6 91.3 L204.4 97.3 L204.4 98 L204.5 98.7 L204.6 99.3 L204.8 99.9 L205 100.4 L205.2 100.8 L205.5 101.2 L205.7 101.5 L206.1 101.7 L206.4 101.9 L185.6 95.9 L185.3 95.7 L185 95.5 L184.7 95.2 L184.4 94.8 L184.2 94.4 L184 93.9 L183.8 93.3 L183.7 92.7 L183.6 92 Z"/>
<path class="iso__ln iso__ln--faint" d="M187.6 79.9 L208.4 85.9"/>
<path class="iso__face" d="M204.4 97.1 L204.5 95.7 L204.7 94.2 L205 92.7 L205.5 91.2 L206 89.7 L206.6 88.4 L207.4 87.2 L208.1 86.2 L208.9 85.4 L209.7 84.9 L210.4 84.6 L211.1 84.5 L211.8 84.7 L212.3 85.2 L212.8 85.9 L213.1 86.9 L213.3 88 L213.4 89.3 L213.3 90.7 L213.1 92.2 L212.8 93.7 L212.3 95.2 L211.8 96.7 L211.1 98 L210.4 99.2 L209.7 100.2 L208.9 101 L208.1 101.6 L207.4 101.9 L206.6 101.9 L206 101.7 L205.5 101.2 L205 100.5 L204.7 99.5 L204.5 98.4 Z"/>
<path class="iso__face iso__face--cav" d="M205.2 96.4 L205.3 94.9 L205.6 93.3 L206 91.7 L206.6 90.2 L207.3 88.8 L208.1 87.7 L208.9 86.8 L209.7 86.3 L210.5 86 L211.2 86.2 L211.8 86.7 L212.2 87.5 L212.5 88.7 L212.6 90 L212.5 91.5 L212.2 93.1 L211.8 94.7 L211.2 96.2 L210.5 97.6 L209.7 98.7 L208.9 99.6 L208.1 100.2 L207.3 100.4 L206.6 100.2 L206 99.7 L205.6 98.9 L205.3 97.8 Z"/>
<g class="iso__m pf__m--6">
<path class="iso__ln iso__ln--lo pf__fibre" d="M-26.5 43.2 L200.4 108.7"/>
<path class="iso__ln iso__ln--lo" d="M-8.4 47.5 L-10 48.9"/>
<path class="iso__ln iso__ln--lo" d="M17.6 55 L16 56.4"/>
<path class="iso__ln iso__ln--lo" d="M43.6 62.5 L42 63.9"/>
<path class="iso__ln iso__ln--lo" d="M69.6 70 L68 71.4"/>
<path class="iso__ln iso__ln--lo" d="M95.5 77.5 L93.9 78.9"/>
<path class="iso__ln iso__ln--lo" d="M121.5 85 L119.9 86.4"/>
<path class="iso__ln iso__ln--lo" d="M147.5 92.5 L145.9 93.9"/>
<path class="iso__ln iso__ln--lo" d="M173.5 100 L171.9 101.4"/>
<path class="iso__halo iso__halo--part" d="M106.9 82.2 L105.1 83.3 L102.7 84 L99.9 84.2 L97.2 84 L94.8 83.3 L93 82.2 L92.1 80.9 L92.1 79.5 L93 78.2 L94.8 77.2 L97.2 76.5 L99.9 76.2 L102.7 76.5 L105.1 77.2 L106.9 78.2 L107.8 79.5 L107.8 80.9 Z"/>
<path class="iso__face iso__face--s1" d="M94.1 76.4 L102.8 78.9 L102.8 82.8 L94.1 80.3 Z"/>
<path class="iso__face iso__face--s2" d="M105.8 76.3 L102.8 78.9 L102.8 82.8 L105.8 80.2 Z"/>
<path class="iso__face" d="M97.1 73.8 L105.8 76.3 L102.8 78.9 L94.1 76.4 Z"/>
<circle class="iso__lamp" cx="101" cy="79.7" r="1.7"/>
</g>
<path class="iso__way pf__way--1" d="M46.3 44.2 L46.3 25.5 L61.3 17.7 L184.3 53.2 L211.5 40.8"/>
<path class="iso__way pf__way--2" d="M86.9 47.9 L86.9 38.4 L102.9 29.7 L184.3 53.2 L211.5 40.8"/>
<path class="iso__way pf__way--3" d="M65.9 47.9 L65.9 34 L83.4 24.1 L184.3 53.2 L211.5 40.8"/>
<path class="iso__way pf__way--4" d="M131.6 58 L131.6 50.7 L147.1 42.5 L184.3 53.2 L211.5 40.8"/>
<path class="iso__way pf__way--5" d="M162.7 32.4 L178.2 34.1 L178.2 51.5 L184.3 53.2 L211.5 40.8"/>
<path class="iso__way pf__way--6" d="M101 79.7 L101 51.1 L124.5 36 L184.3 53.2 L211.5 40.8"/>
<path class="comet iso__comet pf__comet--1" d="M46.3 44.2 L46.3 25.5 L61.3 17.7 L184.3 53.2 L211.5 40.8"/>
<path class="comet iso__comet pf__comet--2" d="M86.9 47.9 L86.9 38.4 L102.9 29.7 L184.3 53.2 L211.5 40.8"/>
<path class="comet iso__comet pf__comet--3" d="M65.9 47.9 L65.9 34 L83.4 24.1 L184.3 53.2 L211.5 40.8"/>
<path class="comet iso__comet pf__comet--4" d="M131.6 58 L131.6 50.7 L147.1 42.5 L184.3 53.2 L211.5 40.8"/>
<path class="comet iso__comet pf__comet--5" d="M162.7 32.4 L178.2 34.1 L178.2 51.5 L184.3 53.2 L211.5 40.8"/>
<path class="comet iso__comet pf__comet--6" d="M101 79.7 L101 51.1 L124.5 36 L184.3 53.2 L211.5 40.8"/>
    </svg>
    <!-- iso:end -->

    <!-- the words: what each method hands over, shown for the tuned one -->
    <span class="iso__tags" aria-hidden="true">
      {#each methods as m, i}<span class="iso__tag pf__tag--{i + 1}">{m.note}</span>{/each}
    </span>
    <span class="iso__name pf__name--dock">the model</span>
    <span class="iso__name pf__name--ground">ground</span>
    <span class="iso__name pf__name--ins">insulated span</span>
  </div>

</IsoFigure>
