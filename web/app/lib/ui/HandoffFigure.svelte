<script>
  import PingMark from "#lib/ui/PingMark.svelte";
  import IsoFigure from "#lib/ui/IsoFigure.svelte";
  // FIG. 5 of the custom-tuning page (2026-11-14; layers in the isometric hand since 2026-11-22): what the
  // add-on does with a request that is not its job. One slab is the customer's system. A request lies on
  // it at the left; its way runs to the station (the console with the add-on in its slot, isoparts, the
  // same objects as FIG. 1 and 2, Ping beside it); from there the way forks. The codes way runs on along
  // the slab to the same card with its codes filled in, back on the plane. The person way turns, rises and
  // runs at a height to a ticket floating above the slab with the headset mark on it: out of the loop,
  // drawn. The scrubbed list (scrub.js, in the runtime) tunes which request is arriving; the picture lifts
  // and lights the chosen outcome through the figure's data-step (scrub.js) and sits the other down (one CSS translate, printed by
  // computer-scientist/article/fig_handoff_gen.py, which writes the drawing between the markers). Written
  // like TuningFigure.svelte: runes, props in, no <style> (rules in src/styles/custom-tuning.css, lead
  // class .ho; names and the step list reuse the .tf__* rules), no colour literal, no inline style, words
  // in HTML over the stage, one aria-hidden SVG with no <text>; Ping is mascot.svg's geometry (two
  // antennae, ONE eye). The still picture (JS off, print, reduced motion) is the hand-off.
  let {
    fig = 5,
    caption = "A work order comes back with its codes. A request that is not its job goes to a person, with the add-on's reason.",
    requests = [
      { name: "A work order", note: "answered with its codes" },
      { name: "Someone's chat message", note: "handed to a person" },
      { name: "Do something else instead", note: "handed to a person" },
    ],
  } = $props();
</script>

<IsoFigure lead="ho" {fig} {caption} steps={requests} stepsClass="tf__steps ho__requests" stepsLabel="What is arriving" numbered={false} start={1}>
  <div class="iso__stage ho__stage">
    <!-- iso:begin -->
    <svg class="iso__svg" viewBox="-42.2 -6 222.1 92.8" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-6.2 2.8 L173.9 54.8 L143.9 80.8 L-36.2 28.8 Z"/>
<g class="ho__slab">
<path class="iso__face iso__face--s1" d="M-28 24.2 L140 72.7 L140 76.2 L-28 27.7 Z"/>
<path class="iso__face iso__face--s2" d="M168 48.5 L140 72.7 L140 76.2 L168 52 Z"/>
<path class="iso__face" d="M0 0 L168 48.5 L140 72.7 L-28 24.2 Z"/>
</g>
<g class="iso__m ho__m ho__m--request">
<path class="iso__halo iso__halo--part" d="M20.9 23.9 L18.2 25.7 L14.7 27 L10.6 27.9 L6.2 28.2 L1.8 27.9 L-2.3 27 L-5.9 25.7 L-8.6 23.9 L-10.3 21.9 L-10.8 19.7 L-10.3 17.5 L-8.6 15.4 L-5.9 13.7 L-2.3 12.3 L1.8 11.5 L6.2 11.2 L10.6 11.5 L14.7 12.3 L18.2 13.7 L20.9 15.4 L22.6 17.5 L23.2 19.7 L22.6 21.9 Z"/>
<path class="iso__face iso__face--s1" d="M-9.6 18.9 L12.9 25.4 L12.9 26.8 L-9.6 20.3 Z"/>
<path class="iso__face iso__face--s2" d="M21.9 17.6 L12.9 25.4 L12.9 26.8 L21.9 19 Z"/>
<path class="iso__face" d="M-0.6 11.1 L21.9 17.6 L12.9 25.4 L-9.6 18.9 Z"/>
<path class="iso__ln iso__ln--faint" d="M0 13.6 L12.1 17.1"/>
<path class="iso__ln iso__ln--faint" d="M-2 15.4 L13.6 19.9"/>
<path class="iso__ln iso__ln--faint" d="M-4 17.1 L6.4 20.1"/>
</g>
<g class="ho__console">
<path class="iso__halo" d="M81.1 40.7 L77.3 43.2 L72.3 45.1 L66.6 46.3 L60.3 46.7 L54.1 46.3 L48.3 45.1 L43.4 43.2 L39.6 40.7 L37.2 37.8 L36.3 34.7 L37.2 31.6 L39.6 28.7 L43.4 26.3 L48.3 24.3 L54.1 23.1 L60.3 22.7 L66.6 23.1 L72.3 24.3 L77.3 26.3 L81.1 28.7 L83.5 31.6 L84.3 34.7 L83.5 37.8 Z"/>
<path class="iso__face iso__face--s1" d="M40.6 19.2 L70.1 27.7 L70.1 43.3 L40.6 34.8 Z"/>
<path class="iso__face iso__face--s2" d="M80.1 19.1 L70.1 27.7 L70.1 43.3 L80.1 34.7 Z"/>
<path class="iso__face" d="M50.6 10.6 L80.1 19.1 L70.1 27.7 L40.6 19.2 Z"/>
<path class="iso__face iso__face--s1" d="M44.2 17.4 L68.5 24.4 L68.5 26.1 L44.2 19.1 Z"/>
<path class="iso__face iso__face--s2" d="M76.5 17.5 L68.5 24.4 L68.5 26.1 L76.5 19.2 Z"/>
<path class="iso__face" d="M52.2 10.5 L76.5 17.5 L68.5 24.4 L44.2 17.4 Z"/>
<path class="iso__face iso__face--s2" d="M44.1 30.6 L54.5 33.6 L54.5 28.9 L44.1 25.9 Z"/>
<path class="iso__ln iso__ln--faint" d="M45.8 30.4 L45.8 27.1"/>
<path class="iso__ln iso__ln--faint" d="M47.5 30.9 L47.5 27.6"/>
<path class="iso__ln iso__ln--faint" d="M49.3 31.4 L49.3 28.1"/>
<path class="iso__ln iso__ln--faint" d="M51 31.9 L51 28.6"/>
<path class="iso__ln iso__ln--faint" d="M52.7 32.4 L52.7 29.1"/>
<path class="iso__face" d="M65 34.3 L64.9 33.3 L64.4 32.3 L63.7 31.5 L62.8 30.9 L61.8 30.6 L60.9 30.7 L60.1 31.1 L59.7 31.8 L59.5 32.7 L59.7 33.7 L60.1 34.7 L60.9 35.5 L61.8 36.1 L62.8 36.4 L63.7 36.3 L64.4 35.9 L64.9 35.2 Z"/>
<path class="iso__ln" d="M62.3 33.5 L64 32"/>
<path class="iso__face iso__face--cav" d="M45 24.8 L60.5 29.3 L60.5 27.2 L45 22.7 Z"/>
<path class="iso__ln iso__ln--faint" d="M45.8 34.1 L64.9 39.6"/>
<path class="iso__ln" d="M72.7 18.7 L72.7 9.2"/>
<circle class="iso__lamp iso__lamp--on" cx="72.7" cy="9.2" r="1.4"/>
<circle class="iso__lamp iso__lamp--on" cx="66.6" cy="28.9" r="1.9"/>
<path class="iso__face iso__face--s1" d="M44 26.3 L54.4 29.3 L54.4 31.9 L44 28.9 Z"/>
<path class="iso__face iso__face--s2" d="M57.9 26.3 L54.4 29.3 L54.4 31.9 L57.9 28.9 Z"/>
<path class="iso__face" d="M47.5 23.3 L57.9 26.3 L54.4 29.3 L44 26.3 Z"/>
<path class="iso__ln iso__ln--faint" d="M45.3 28.4 L53.1 30.7"/>
<circle class="iso__lamp iso__lamp--on" cx="52.7" cy="29.8" r="1.3"/>
</g>
<PingMark root="tf__ping" prefix="tf__ping-" shape="mark" eyeY={22} waveY={22} transform="translate(19.3 16.4) scale(.3)" />
<g class="iso__m ho__m ho__m--codes">
<path class="iso__halo iso__halo--part" d="M120.5 59.6 L117.8 61.4 L114.3 62.7 L110.2 63.6 L105.8 63.9 L101.4 63.6 L97.3 62.7 L93.8 61.4 L91.1 59.6 L89.4 57.6 L88.8 55.4 L89.4 53.2 L91.1 51.1 L93.8 49.4 L97.3 48 L101.4 47.2 L105.8 46.9 L110.2 47.2 L114.3 48 L117.8 49.4 L120.5 51.1 L122.2 53.2 L122.8 55.4 L122.2 57.6 Z"/>
<path class="iso__face iso__face--s1" d="M90 54.6 L112.6 61.1 L112.6 62.5 L90 56 Z"/>
<path class="iso__face iso__face--s2" d="M121.6 53.3 L112.6 61.1 L112.6 62.5 L121.6 54.7 Z"/>
<path class="iso__face" d="M99 46.8 L121.6 53.3 L112.6 61.1 L90 54.6 Z"/>
<path class="iso__ln iso__ln--faint" d="M99.6 49.3 L111.8 52.8"/>
<path class="iso__ln iso__ln--faint" d="M97.6 51.1 L113.2 55.6"/>
<path class="iso__ln iso__ln--faint" d="M95.6 52.8 L106 55.8"/>
<path class="iso__face iso__face--cav ho__chip" d="M113.1 52.6 L117.5 53.9 L116.3 54.9 L111.9 53.7 Z"/>
<path class="iso__face iso__face--cav ho__chip" d="M96.4 54.1 L100.7 55.4 L99.5 56.4 L95.2 55.2 Z"/>
<path class="iso__face iso__face--cav ho__chip" d="M102.4 55.9 L106.8 57.1 L105.6 58.2 L101.2 56.9 Z"/>
<path class="iso__face iso__face--cav ho__chip" d="M108.5 57.6 L112.8 58.9 L111.6 59.9 L107.3 58.7 Z"/>
<circle class="iso__lamp" cx="110" cy="60.4" r="1.3"/>
</g>
<g class="iso__m ho__m ho__m--person">
<path class="iso__halo iso__halo--part" d="M155.6 51.3 L152.9 53 L149.4 54.4 L145.3 55.2 L140.9 55.5 L136.5 55.2 L132.4 54.4 L128.8 53 L126.1 51.3 L124.4 49.2 L123.9 47 L124.4 44.8 L126.1 42.8 L128.8 41 L132.4 39.7 L136.5 38.8 L140.9 38.5 L145.3 38.8 L149.4 39.7 L152.9 41 L155.6 42.8 L157.3 44.8 L157.9 47 L157.3 49.2 Z"/>
<path class="iso__face iso__face--s1" d="M125.1 38.5 L147.6 45 L147.6 46.4 L125.1 39.9 Z"/>
<path class="iso__face iso__face--s2" d="M156.6 37.2 L147.6 45 L147.6 46.4 L156.6 38.6 Z"/>
<path class="iso__face" d="M134.1 30.7 L156.6 37.2 L147.6 45 L125.1 38.5 Z"/>
<path class="iso__ln iso__ln--faint" d="M134.7 33.2 L146.8 36.7"/>
<path class="iso__ln iso__ln--faint" d="M132.7 34.9 L148.3 39.4"/>
<path class="iso__ln iso__ln--faint" d="M130.7 36.6 L141.1 39.6"/>
<circle class="iso__lamp" cx="145" cy="44.2" r="1.3"/>
<g class="ho__person" transform="translate(131 30.1) scale(.24)">
<circle class="iso__ln" cx="34" cy="19" r="7"/>
<path class="iso__ln" d="M18 46 c0 -10 7 -15 16 -15 s16 5 16 15"/>
<path class="iso__ln" d="M25 19 a9 9 0 0 1 18 0"/><path class="iso__ln" d="M43 19 v5 h-4"/>
<rect class="ho__fill" x="23" y="18" width="3" height="5" rx="1"/><rect class="ho__fill" x="42" y="18" width="3" height="5" rx="1"/>
</g>
</g>
<path class="iso__ln ho__port" d="M73.1 34.7 L74.4 35"/>
<path class="iso__way ho__way ho__way--in" d="M17.4 21.9 L24.3 24.4 L53.8 32.9 L45.8 39.8 L45.8 28.1 L49.3 25"/>
<path class="iso__way ho__way ho__way--codes" d="M73.1 34.7 L80 36.7 L80 42.2 L93.9 46.2 L89.4 50.1 L94.5 51.1"/>
<path class="iso__way ho__way ho__way--person" d="M73.1 34.7 L80 36.7 L91.5 26.7 L91.5 24.3 L129.6 35.3"/>
<path class="comet iso__comet ho__comet ho__comet--in" d="M17.4 21.9 L24.3 24.4 L53.8 32.9 L45.8 39.8 L45.8 28.1 L49.3 25"/>
<path class="comet iso__comet ho__comet ho__comet--codes" d="M73.1 34.7 L80 36.7 L80 42.2 L93.9 46.2 L89.4 50.1 L94.5 51.1"/>
<path class="comet iso__comet ho__comet ho__comet--person" d="M73.1 34.7 L80 36.7 L91.5 26.7 L91.5 24.3 L129.6 35.3"/>
    </svg>
    <!-- iso:end -->
    <span class="ho__tag ho__tag--card"><span class="tf__name">A request</span><span class="tf__note">a record, or anything else</span></span>
    <span class="ho__tag ho__tag--ping"><span class="tf__name">Your add-on</span><span class="tf__note">on your private band</span></span>
    <span class="ho__tag ho__tag--codes"><span class="tf__name">Its codes</span><span class="tf__note">back to your system</span></span>
    <span class="ho__tag ho__tag--person"><span class="tf__name">A person</span><span class="tf__note">decides, with the reason</span></span>
  </div>

</IsoFigure>
