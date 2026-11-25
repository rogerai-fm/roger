<script>
  // THE WILDFIRE WATCH, as designed: shared by every Wave model page (legacy: _partials/wave-firewatch.html).
  // Warning and mitigation, never fire protection; the model never actuates equipment.
  // The ladder figure (2026-11-24, founder ask): four rungs and a person in the isometric hand, generated
  // between the markers by computer-scientist/article/fig_ladder_gen.py; the step list plays it (scrub.js,
  // data-scrub-auto) and wave-model.css reads the tuned rung through the figure's data-step (scrub.js). Still (no JS, print, reduced
  // motion): the whole ladder at rest. No word in the drawing; the list says each rung.
  let { no, fig } = $props();
</script>
<section class="section" id="wildfire">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;{no} / WILDFIRE</span>
          <h2>The wildfire watch, as designed.</h2>
          <p>Warning and mitigation, never fire protection. In this design the models warn
             and explain; the model never actuates equipment.</p>
        </div>

        <div class="device-contract">
          <p><b>The design.</b> Wave Pico would sit on the mast, reading particulates,
             carbon monoxide, volatile organics, thermal, the camera module&rsquo;s outputs
             and the sprinkler controller&rsquo;s state. Wave Nano would sit on an edge box
             with a camera smoke detector and read the whole site. A larger tier would
             weigh all of it.</p>
          <p><b>What is measured, and where.</b> Each fire number on this page sits beside
             a note on how it was tested, in the adapter section. False-alarm rates at a real
             site over months are not measured yet, and nothing here is a claim about
             suppression.</p>
          <p><b>Pre-wet is the owner&rsquo;s rule.</b> The owner arms it in advance. The
             controller, not the model, executes that owner-armed rule, with its own abort
             window, and only after the escalation ladder has been tried and no person
             acknowledged. The model&rsquo;s part is to mark the site pre-wet eligible; it
             never opens a valve.</p>
        </div>

        <figure class="figure figure--plate wm-ladder-fig lad" role="group" aria-labelledby="wm-ladder-cap">
          <div class="iso__stage lad__stage">
          <!-- iso:begin -->
          <svg class="iso__svg" viewBox="-59.9 -31.8 196.8 115.7" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-3.9 -4.6 L132.9 34.9 L80.9 79.9 L-55.9 40.4 Z"/>
<g class="iso__m lad__m lad__m--1">
<path class="iso__face iso__face--s1" d="M-13 -1.7 L16.4 6.8 L16.4 19.8 L-13 11.3 Z"/>
<path class="iso__face iso__face--s2" d="M29.4 -4.5 L16.4 6.8 L16.4 19.8 L29.4 8.5 Z"/>
<path class="iso__face" d="M0 -13 L29.4 -4.5 L16.4 6.8 L-13 -1.7 Z"/>
<path class="iso__face iso__face--s1" d="M5.1 -14.7 L10.3 -13.2 L10.3 -1.9 L5.1 -3.4 Z"/>
<path class="iso__face iso__face--s2" d="M11.3 -14.1 L10.3 -13.2 L10.3 -1.9 L11.3 -2.8 Z"/>
<path class="iso__face" d="M6.1 -15.6 L11.3 -14.1 L10.3 -13.2 L5.1 -14.7 Z"/>
<path class="iso__face iso__face--cav" d="M6 -12.7 L9.5 -11.7 L9.5 -4.8 L6 -5.8 Z"/>
<circle class="iso__lamp" cx="8.2" cy="-17" r="1.7"/>
</g>
<g class="iso__m lad__m lad__m--2">
<path class="iso__face iso__face--s1" d="M16.4 -6.2 L45.9 2.3 L45.9 28.3 L16.4 19.8 Z"/>
<path class="iso__face iso__face--s2" d="M58.9 -9 L45.9 2.3 L45.9 28.3 L58.9 17 Z"/>
<path class="iso__face" d="M29.4 -17.5 L58.9 -9 L45.9 2.3 L16.4 -6.2 Z"/>
<path class="iso__face iso__face--s1" d="M24.2 -20.4 L29.4 -18.9 L29.4 -9.4 L24.2 -10.9 Z"/>
<path class="iso__face iso__face--s2" d="M30.4 -19.8 L29.4 -18.9 L29.4 -9.4 L30.4 -10.3 Z"/>
<path class="iso__face" d="M25.2 -21.3 L30.4 -19.8 L29.4 -18.9 L24.2 -20.4 Z"/>
<path class="iso__face iso__face--cav" d="M25 -18.5 L28.5 -17.5 L28.5 -12.3 L25 -13.3 Z"/>
<circle class="iso__lamp" cx="27.3" cy="-22.7" r="1.4"/>
<path class="iso__face iso__face--s1" d="M32 -18.2 L37.2 -16.7 L37.2 -7.2 L32 -8.7 Z"/>
<path class="iso__face iso__face--s2" d="M38.2 -17.6 L37.2 -16.7 L37.2 -7.2 L38.2 -8 Z"/>
<path class="iso__face" d="M33 -19.1 L38.2 -17.6 L37.2 -16.7 L32 -18.2 Z"/>
<path class="iso__face iso__face--cav" d="M32.8 -16.2 L36.3 -15.2 L36.3 -10 L32.8 -11 Z"/>
<circle class="iso__lamp" cx="35.1" cy="-20.5" r="1.4"/>
<path class="iso__face iso__face--s1" d="M39.8 -16 L45 -14.5 L45 -4.9 L39.8 -6.4 Z"/>
<path class="iso__face iso__face--s2" d="M46 -15.3 L45 -14.5 L45 -4.9 L46 -5.8 Z"/>
<path class="iso__face" d="M40.8 -16.8 L46 -15.3 L45 -14.5 L39.8 -16 Z"/>
<path class="iso__face iso__face--cav" d="M40.6 -14 L44.1 -13 L44.1 -7.8 L40.6 -8.8 Z"/>
<circle class="iso__lamp" cx="42.9" cy="-18.2" r="1.4"/>
</g>
<g class="iso__m lad__m lad__m--3">
<path class="iso__face iso__face--s1" d="M45.9 -10.7 L75.3 -2.2 L75.3 36.8 L45.9 28.3 Z"/>
<path class="iso__face iso__face--s2" d="M88.3 -13.5 L75.3 -2.2 L75.3 36.8 L88.3 25.5 Z"/>
<path class="iso__face" d="M58.9 -22 L88.3 -13.5 L75.3 -2.2 L45.9 -10.7 Z"/>
<path class="iso__face iso__face--s1" d="M57.6 -13.8 L57.6 -11.7 L57.7 -11 L58.1 -10.3 L58.6 -9.6 L59.3 -9 L60.2 -8.5 L61.3 -8 L62.5 -7.6 L63.8 -7.4 L65.2 -7.2 L66.6 -7.2 L68 -7.2 L69.4 -7.4 L70.7 -7.6 L71.9 -8 L73 -8.5 L73.9 -9 L74.6 -9.6 L75.2 -10.3 L75.5 -11 L75.6 -11.7 L75.6 -13.8 Z"/>
<path class="iso__face" d="M74.4 -11.6 L73 -10.6 L71.1 -9.9 L68.9 -9.5 L66.6 -9.3 L64.3 -9.5 L62.1 -9.9 L60.2 -10.6 L58.8 -11.6 L57.9 -12.7 L57.6 -13.8 L57.9 -15 L58.8 -16.1 L60.2 -17 L62.1 -17.7 L64.3 -18.2 L66.6 -18.3 L68.9 -18.2 L71.1 -17.7 L73 -17 L74.4 -16.1 L75.3 -15 L75.6 -13.8 L75.3 -12.7 Z"/>
<path class="lad__track" d="M72.5 -12.2 L71.9 -11.8 L71.2 -11.4 L70.3 -11.1 L69.4 -10.8 L68.4 -10.6 L67.3 -10.5 L66.3 -10.5 L65.2 -10.6 L64.2 -10.7 L63.2 -11 L62.3 -11.3 L61.6 -11.6 L60.9 -12.1 L60.4 -12.5 L60 -13 L59.8 -13.6 L59.8 -14.1 L60 -14.6 L60.3 -15.1 L60.7 -15.6 L61.3 -16.1 L62.1 -16.4 L62.9 -16.8 L63.8 -17 L64.9 -17.2 L65.9 -17.3 L67 -17.3 L68 -17.2 L69 -17.1 L70 -16.9 L70.9 -16.6 L71.7 -16.2 L72.3 -15.8 L72.8 -15.3 L73.2 -14.8 L73.4 -14.3 L73.4 -13.7 L73.3 -13.2 L73 -12.7 Z"/>
<path class="lad__sweep" d="M72.5 -12.2 L71.9 -11.8 L71.2 -11.4 L70.3 -11.1 L69.4 -10.8 L68.4 -10.6 L67.3 -10.5 L66.3 -10.5 L65.2 -10.6 L64.2 -10.7 L63.2 -11 L62.3 -11.3 L61.6 -11.6 L60.9 -12.1 L60.4 -12.5 L60 -13 L59.8 -13.6 L59.8 -14.1 L60 -14.6 L60.3 -15.1 L60.7 -15.6 L61.3 -16.1 L62.1 -16.4 L62.9 -16.8 L63.8 -17 L64.9 -17.2 L65.9 -17.3 L67 -17.3 L68 -17.2 L69 -17.1 L70 -16.9 L70.9 -16.6 L71.7 -16.2 L72.3 -15.8 L72.8 -15.3 L73.2 -14.8 L73.4 -14.3 L73.4 -13.7 L73.3 -13.2 L73 -12.7 Z" pathLength="100" data-draw-skip/>
<path class="iso__ln lad__hand" d="M66.6 -13.9 L69.4 -16.3"/>
<circle class="iso__lamp" cx="66.6" cy="-13.9" r="1.3"/>
</g>
<g class="iso__m lad__m lad__m--4">
<path class="iso__face iso__face--s1" d="M75.3 -15.2 L104.8 -6.7 L104.8 45.3 L75.3 36.8 Z"/>
<path class="iso__face iso__face--s2" d="M117.8 -18 L104.8 -6.7 L104.8 45.3 L117.8 34 Z"/>
<path class="iso__face" d="M88.3 -26.5 L117.8 -18 L104.8 -6.7 L75.3 -15.2 Z"/>
<path class="lad__mark" d="M104.7 -13.8 L103.9 -13.2 L103 -12.7 L101.9 -12.2 L100.6 -11.9 L99.3 -11.6 L97.9 -11.4 L96.5 -11.3 L95.1 -11.3 L93.7 -11.5 L92.3 -11.7 L91.1 -12 L89.9 -12.4 L88.8 -12.9 L87.9 -13.4 L87.2 -14 L86.6 -14.7 L86.2 -15.4 L86.1 -16.1 L86.1 -16.8 L86.3 -17.5 L86.8 -18.2 L87.4 -18.8 L88.2 -19.4 L89.2 -19.9 L90.3 -20.4 L91.5 -20.8 L92.8 -21 L94.2 -21.2 L95.6 -21.3 L97 -21.3 L98.4 -21.2 L99.8 -21 L101.1 -20.7 L102.2 -20.3 L103.3 -19.8 L104.2 -19.2 L104.9 -18.6 L105.5 -18 L105.9 -17.3 L106 -16.6 L106 -15.8 L105.8 -15.1 L105.3 -14.5 Z" pathLength="100" data-draw-skip/>
<path class="iso__face iso__face--s1" d="M94.7 -23.9 L94.7 -16.1 L94.7 -16 L94.7 -15.9 L94.8 -15.8 L94.9 -15.7 L95.1 -15.7 L95.2 -15.6 L95.4 -15.5 L95.6 -15.5 L95.8 -15.5 L96.1 -15.4 L96.3 -15.5 L96.5 -15.5 L96.7 -15.5 L96.9 -15.6 L97 -15.7 L97.2 -15.7 L97.3 -15.8 L97.4 -15.9 L97.4 -16 L97.5 -16.1 L97.5 -23.9 Z"/>
<path class="iso__face" d="M97.3 -23.6 L97 -23.4 L96.8 -23.3 L96.4 -23.3 L96.1 -23.2 L95.7 -23.3 L95.4 -23.3 L95.1 -23.4 L94.8 -23.6 L94.7 -23.8 L94.7 -23.9 L94.7 -24.1 L94.8 -24.3 L95.1 -24.4 L95.4 -24.5 L95.7 -24.6 L96.1 -24.6 L96.4 -24.6 L96.8 -24.5 L97 -24.4 L97.3 -24.3 L97.4 -24.1 L97.5 -23.9 L97.4 -23.8 Z"/>
<path class="iso__face iso__face--s1" d="M91.9 -25.7 L91.9 -23.9 L91.9 -23.6 L92.1 -23.3 L92.3 -23 L92.7 -22.7 L93.1 -22.5 L93.6 -22.2 L94.2 -22.1 L94.8 -21.9 L95.4 -21.9 L96.1 -21.8 L96.7 -21.9 L97.4 -21.9 L98 -22.1 L98.5 -22.2 L99 -22.5 L99.5 -22.7 L99.8 -23 L100.1 -23.3 L100.2 -23.6 L100.3 -23.9 L100.3 -25.7 Z"/>
<path class="iso__face" d="M99.7 -24.6 L99 -24.2 L98.2 -23.9 L97.1 -23.6 L96.1 -23.6 L95 -23.6 L94 -23.9 L93.1 -24.2 L92.4 -24.6 L92 -25.1 L91.9 -25.7 L92 -26.2 L92.4 -26.7 L93.1 -27.2 L94 -27.5 L95 -27.7 L96.1 -27.8 L97.1 -27.7 L98.2 -27.5 L99 -27.2 L99.7 -26.7 L100.1 -26.2 L100.3 -25.7 L100.1 -25.1 Z"/>
<circle class="iso__lamp" cx="96.1" cy="-28.3" r="1.7"/>
</g>
<g class="lad__ways">
<path class="lad__way lad__way--1" d="M1.7 2.5 L2.1 3.1 L2.4 3.6 L2.8 4.2 L3.1 4.9 L3.5 5.6 L3.8 6.4 L4.2 7.2 L4.5 8.2 L4.9 9.3 L5.2 10.4 L5.6 11.7 L5.9 13.1 L6.3 14.6 L6.6 16.3 L7 18 L7.3 19.9 L7.7 21.9 L8.1 23.9 L8.4 26.1 L8.8 28.3 L9.1 30.5 L9.5 32.9 L9.8 35.2 L10.2 37.6" pathLength="100" data-draw-skip/>
<path class="lad__way lad__way--2" d="M31.2 -2 L30.3 -1.2 L29.5 -0.4 L28.7 0.4 L27.9 1.2 L27.1 2.1 L26.2 3.1 L25.4 4.2 L24.6 5.3 L23.8 6.6 L23 8 L22.1 9.5 L21.3 11.1 L20.5 12.8 L19.7 14.6 L18.9 16.6 L18 18.6 L17.2 20.8 L16.4 23.1 L15.6 25.4 L14.7 27.8 L13.9 30.3 L13.1 32.8 L12.3 35.4 L11.5 37.9" pathLength="100" data-draw-skip/>
<path class="lad__way lad__way--3" d="M60.6 -6.5 L58.6 -5.5 L56.6 -4.5 L54.6 -3.5 L52.6 -2.5 L50.6 -1.3 L48.6 -0.2 L46.7 1.1 L44.7 2.5 L42.7 3.9 L40.7 5.5 L38.7 7.2 L36.7 9 L34.7 10.9 L32.7 13 L30.7 15.1 L28.7 17.4 L26.7 19.8 L24.7 22.2 L22.7 24.8 L20.7 27.4 L18.7 30.1 L16.7 32.8 L14.8 35.6 L12.8 38.3" pathLength="100" data-draw-skip/>
<path class="lad__way lad__way--4" d="M90.1 -11 L86.9 -9.8 L83.7 -8.6 L80.6 -7.4 L77.4 -6.1 L74.2 -4.8 L71.1 -3.4 L67.9 -2 L64.7 -0.4 L61.6 1.3 L58.4 3 L55.2 4.9 L52.1 6.9 L48.9 9.1 L45.7 11.3 L42.6 13.7 L39.4 16.1 L36.2 18.7 L33.1 21.4 L29.9 24.1 L26.7 27 L23.6 29.8 L20.4 32.8 L17.2 35.7 L14.1 38.7" pathLength="100" data-draw-skip/>
</g>
<g class="iso__m lad__m lad__m--5">
<path class="iso__halo iso__halo--part" d="M17.8 53.5 L16.9 54.2 L15.7 54.7 L14.5 55.2 L13 55.5 L11.5 55.7 L10 55.8 L8.4 55.7 L6.9 55.5 L5.5 55.2 L4.2 54.7 L3.1 54.2 L2.2 53.5 L1.5 52.8 L1.1 52.1 L1 51.3 L1.1 50.5 L1.5 49.7 L2.2 49 L3.1 48.4 L4.2 47.8 L5.5 47.4 L6.9 47.1 L8.4 46.9 L10 46.8 L11.5 46.9 L13 47.1 L14.5 47.4 L15.7 47.8 L16.9 48.4 L17.8 49 L18.4 49.7 L18.8 50.5 L19 51.3 L18.8 52.1 L18.4 52.8 Z"/>
<path class="iso__face iso__face--s1" d="M6 41 L6 51.4 L6 51.7 L6.2 52 L6.4 52.3 L6.7 52.5 L7.1 52.8 L7.6 53 L8.1 53.2 L8.7 53.3 L9.3 53.3 L10 53.4 L10.6 53.3 L11.2 53.3 L11.8 53.2 L12.3 53 L12.8 52.8 L13.2 52.5 L13.5 52.3 L13.8 52 L13.9 51.7 L14 51.4 L14 41 Z"/>
<path class="iso__face" d="M13.4 42 L12.8 42.4 L12 42.7 L11 42.9 L10 43 L8.9 42.9 L8 42.7 L7.1 42.4 L6.5 42 L6.1 41.5 L6 41 L6.1 40.5 L6.5 40 L7.1 39.6 L8 39.2 L8.9 39 L10 39 L11 39 L12 39.2 L12.8 39.6 L13.4 40 L13.8 40.5 L14 41 L13.8 41.5 Z"/>
<circle class="iso__face" cx="10" cy="36.2" r="3.6"/>
</g>
<g class="lad__pings">
<path class="lad__ping lad__ping--1 lad__ping--k1" d="M13.4 -15.5 L12.6 -14.9 L11.6 -14.5 L10.4 -14.2 L9.1 -14 L7.8 -14 L6.5 -14.1 L5.2 -14.4 L4.1 -14.8 L3.3 -15.3 L2.6 -15.9 L2.3 -16.5 L2.2 -17.2 L2.5 -17.9 L3 -18.5 L3.8 -19 L4.8 -19.4 L6 -19.8 L7.3 -19.9 L8.7 -20 L10 -19.8 L11.2 -19.6 L12.3 -19.2 L13.2 -18.7 L13.8 -18.1 L14.2 -17.4 L14.2 -16.7 L14 -16.1 Z"/>
<path class="lad__ping lad__ping--1 lad__ping--k2" d="M13.4 -15.5 L12.6 -14.9 L11.6 -14.5 L10.4 -14.2 L9.1 -14 L7.8 -14 L6.5 -14.1 L5.2 -14.4 L4.1 -14.8 L3.3 -15.3 L2.6 -15.9 L2.3 -16.5 L2.2 -17.2 L2.5 -17.9 L3 -18.5 L3.8 -19 L4.8 -19.4 L6 -19.8 L7.3 -19.9 L8.7 -20 L10 -19.8 L11.2 -19.6 L12.3 -19.2 L13.2 -18.7 L13.8 -18.1 L14.2 -17.4 L14.2 -16.7 L14 -16.1 Z"/>
<path class="lad__ping lad__ping--2 lad__ping--p0" d="M31.6 -21.5 L30.8 -21 L29.8 -20.6 L28.6 -20.3 L27.3 -20.2 L26 -20.3 L24.8 -20.6 L23.7 -21 L22.9 -21.5 L22.4 -22.1 L22.3 -22.7 L22.4 -23.4 L22.9 -24 L23.7 -24.5 L24.8 -24.9 L26 -25.1 L27.3 -25.2 L28.6 -25.1 L29.8 -24.9 L30.8 -24.5 L31.6 -24 L32.1 -23.4 L32.3 -22.7 L32.1 -22.1 Z"/>
<path class="lad__ping lad__ping--2 lad__ping--p1" d="M39.4 -19.2 L38.6 -18.7 L37.6 -18.3 L36.4 -18.1 L35.1 -18 L33.8 -18.1 L32.6 -18.3 L31.5 -18.7 L30.7 -19.2 L30.2 -19.8 L30.1 -20.5 L30.2 -21.1 L30.7 -21.7 L31.5 -22.2 L32.6 -22.6 L33.8 -22.9 L35.1 -23 L36.4 -22.9 L37.6 -22.6 L38.6 -22.2 L39.4 -21.7 L39.9 -21.1 L40.1 -20.5 L39.9 -19.8 Z"/>
<path class="lad__ping lad__ping--2 lad__ping--p2" d="M47.2 -17 L46.4 -16.5 L45.4 -16.1 L44.2 -15.8 L42.9 -15.7 L41.6 -15.8 L40.4 -16.1 L39.3 -16.5 L38.5 -17 L38 -17.6 L37.9 -18.2 L38 -18.9 L38.5 -19.5 L39.3 -20 L40.4 -20.4 L41.6 -20.6 L42.9 -20.7 L44.2 -20.6 L45.4 -20.4 L46.4 -20 L47.2 -19.5 L47.7 -18.9 L47.9 -18.2 L47.7 -17.6 Z"/>
<path class="lad__ping lad__ping--5 lad__ping--k1" d="M17.8 53.4 L16.7 54.2 L15.4 54.8 L13.9 55.2 L12.3 55.5 L10.6 55.7 L8.8 55.7 L7.1 55.5 L5.5 55.1 L4 54.6 L2.8 53.9 L1.9 53.2 L1.3 52.4 L1 51.5 L1 50.6 L1.4 49.8 L2.2 48.9 L3.2 48.2 L4.5 47.6 L6 47.2 L7.6 46.9 L9.4 46.7 L11.1 46.7 L12.9 46.9 L14.5 47.3 L15.9 47.8 L17.1 48.5 L18 49.2 L18.7 50 L18.9 50.9 L18.9 51.8 L18.5 52.6 Z"/>
<path class="lad__ping lad__ping--5 lad__ping--k2" d="M17.8 53.4 L16.7 54.2 L15.4 54.8 L13.9 55.2 L12.3 55.5 L10.6 55.7 L8.8 55.7 L7.1 55.5 L5.5 55.1 L4 54.6 L2.8 53.9 L1.9 53.2 L1.3 52.4 L1 51.5 L1 50.6 L1.4 49.8 L2.2 48.9 L3.2 48.2 L4.5 47.6 L6 47.2 L7.6 46.9 L9.4 46.7 L11.1 46.7 L12.9 46.9 L14.5 47.3 L15.9 47.8 L17.1 48.5 L18 49.2 L18.7 50 L18.9 50.9 L18.9 51.8 L18.5 52.6 Z"/>
</g>
          </svg>
          <!-- iso:end -->
          </div>
          <ol class="steps wm-ladder" data-scrub data-scrub-auto aria-label="The escalation ladder, one rung at a time">
            <li><b>Notify.</b> The owner is told, with the reading that raised it.</li>
            <li><b>Escalate.</b> Every path the owner set up is tried, in order.</li>
            <li><b>Human-ack pending.</b> The watch waits for a person to answer.</li>
            <li><b>Pre-wet eligible.</b> Owner-armed: the controller may execute the owner&rsquo;s rule. The model only marks it.</li>
            <li><b>Handed to a person.</b> An acknowledgement at any step ends here.</li>
          </ol>
          <figcaption class="fig mono" id="wm-ladder-cap">FIG. {fig} - The escalation ladder. The model climbs it with records; a person, or the owner&rsquo;s armed rule, acts.</figcaption>
        </figure>

        <p class="wm-aside">This is not a fire alarm or a sprinkler system, and it is never in
           the control path of one. It does not replace code-required detection or
           suppression, or the fire service.</p>
      </div>
    </section>
