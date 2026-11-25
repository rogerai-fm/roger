<script>
  import IsoFigure from "#lib/ui/IsoFigure.svelte";
  // The private band (custom-tuning page, 2026-11-13; layers in the isometric hand since 2026-11-22): your
  // app, the CLI and the web UI tune to a band on RogerAI's GPUs with your key; anyone else finds no
  // carrier. The dial is the top face of a slab, RogerAI's machine; the private band is a channel standing
  // on it with its lamp; three client cards float above and behind the slab and their lines drop onto the
  // channel; the fourth card sits off the slab's near end and its line meets a wall at the edge. Generated
  // by computer-scientist/article/fig_band_gen.py between the markers (the sweep keyframes too). Written
  // like TuningFigure.svelte: runes, props in, no <style> (rules in src/styles/custom-tuning.css, lead
  // class .pb), no colour literal, no inline style, words in HTML over the stage, one aria-hidden SVG with
  // no <text>. The still default is data-key="yours" (the needle on the band, lifted and lit); the key
  // switch (src/js/keyswitch.js, hidden without JS) flips it to "other": the needle sweeps the slab and
  // never locks, the band sits down dark. Motion scales with --motion.
  let {
    fig = 2,
    caption = "One private band per customer. Your app, the CLI and the web UI tune to it with your key; without it the needle finds nothing.",
  } = $props();
</script>

<IsoFigure lead="pb" {fig} {caption} data-keyed="" data-key="yours">
  <div class="iso__stage pb__stage">
    <!-- iso:begin -->
    <svg class="iso__svg" viewBox="-33.7 -31.3 234.5 109.3" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-3.7 -2.3 L181.6 51.2 L157.6 72 L-27.7 18.5 Z"/>
<g class="pb__machine">
<path class="iso__face iso__face--s1" d="M-20 12.1 L153.2 62.1 L153.2 67.3 L-20 17.3 Z"/>
<path class="iso__face iso__face--s2" d="M173.2 44.8 L153.2 62.1 L153.2 67.3 L173.2 50 Z"/>
<path class="iso__face" d="M0 -5.2 L173.2 44.8 L153.2 62.1 L-20 12.1 Z"/>
<path class="iso__ln iso__ln--lo" d="M-12.3 11.5 L150.5 58.5"/>
<path class="iso__ln iso__ln--lo" d="M-8.8 12.5 L-5.8 9.9"/>
<path class="iso__ln iso__ln--faint" d="M-1 14.7 L0.5 13.4"/>
<path class="iso__ln iso__ln--faint" d="M6.7 17 L8.2 15.7"/>
<path class="iso__ln iso__ln--faint" d="M14.5 19.2 L16 17.9"/>
<path class="iso__ln iso__ln--lo" d="M22.3 21.5 L25.3 18.9"/>
<path class="iso__ln iso__ln--faint" d="M30.1 23.7 L31.6 22.4"/>
<path class="iso__ln iso__ln--faint" d="M37.9 26 L39.4 24.7"/>
<path class="iso__ln iso__ln--faint" d="M45.7 28.2 L47.2 26.9"/>
<path class="iso__ln iso__ln--lo" d="M53.5 30.5 L56.5 27.9"/>
<path class="iso__ln iso__ln--faint" d="M61.3 32.7 L62.8 31.4"/>
<path class="iso__ln iso__ln--faint" d="M69.1 35 L70.6 33.7"/>
<path class="iso__ln iso__ln--faint" d="M76.9 37.2 L78.4 35.9"/>
<path class="iso__ln iso__ln--lo" d="M84.7 39.5 L87.7 36.9"/>
<path class="iso__ln iso__ln--faint" d="M92.5 41.7 L94 40.4"/>
<path class="iso__ln iso__ln--faint" d="M100.3 44 L101.8 42.7"/>
<path class="iso__ln iso__ln--faint" d="M108.1 46.2 L109.6 44.9"/>
<path class="iso__ln iso__ln--lo" d="M115.9 48.5 L118.9 45.9"/>
<path class="iso__ln iso__ln--faint" d="M123.7 50.7 L125.2 49.4"/>
<path class="iso__ln iso__ln--faint" d="M131.5 53 L133 51.7"/>
<path class="iso__ln iso__ln--faint" d="M139.3 55.2 L140.8 53.9"/>
<path class="iso__ln iso__ln--lo" d="M147 57.5 L150 54.9"/>
</g>
<g class="iso__m pb__band">
<path class="iso__halo iso__halo--part" d="M92.4 32.7 L89.4 34.7 L85.5 36.2 L80.9 37.2 L76 37.5 L71.1 37.2 L66.5 36.2 L62.6 34.7 L59.5 32.7 L57.6 30.5 L57 28 L57.6 25.5 L59.5 23.2 L62.6 21.3 L66.5 19.8 L71.1 18.8 L76 18.5 L80.9 18.8 L85.5 19.8 L89.4 21.3 L92.4 23.2 L94.3 25.5 L95 28 L94.3 30.5 Z"/>
<path class="iso__face iso__face--s1" d="M59.7 29.2 L78.8 34.7 L78.8 36.6 L59.7 31.1 Z"/>
<path class="iso__face iso__face--s2" d="M92.3 23 L78.8 34.7 L78.8 36.6 L92.3 24.9 Z"/>
<path class="iso__face" d="M73.2 17.5 L92.3 23 L78.8 34.7 L59.7 29.2 Z"/>
<path class="iso__face iso__face--cav" d="M73.9 18.9 L89.5 23.4 L78 33.3 L62.4 28.8 Z"/>
<circle class="iso__lamp" cx="76.2" cy="34.8" r="1.4"/>
</g>
<path class="pb__needle" d="M87.2 15.7 L64.2 35.6"/>
<g class="iso__m pb__client pb__client--own pb__client--app">
<path class="iso__halo iso__halo--part" d="M48.6 0.7 L46.3 2.2 L43.4 3.3 L40.1 4 L36.4 4.2 L32.8 4 L29.4 3.3 L26.5 2.2 L24.3 0.7 L22.9 -0.9 L22.4 -2.8 L22.9 -4.6 L24.3 -6.3 L26.5 -7.7 L29.4 -8.8 L32.8 -9.5 L36.4 -9.8 L40.1 -9.5 L43.4 -8.8 L46.3 -7.7 L48.6 -6.3 L50 -4.6 L50.4 -2.8 L50 -0.9 Z"/>
<path class="iso__face iso__face--s1" d="M24.3 -19.2 L41.6 -14.2 L41.6 -12.8 L24.3 -17.8 Z"/>
<path class="iso__face iso__face--s2" d="M48.6 -20.3 L41.6 -14.2 L41.6 -12.8 L48.6 -18.9 Z"/>
<path class="iso__face" d="M31.3 -25.3 L48.6 -20.3 L41.6 -14.2 L24.3 -19.2 Z"/>
<g class="pb__mark" transform="translate(30 -26.2) scale(.36)"><rect class="pb__glyph-face" x="10" y="3" width="16" height="30" rx="3"/><line class="pb__glyph" x1="15" y1="29" x2="21" y2="29"/><rect class="pb__glyph" x="13" y="7" width="10" height="17" rx="1"/></g>
</g>
<g class="iso__m pb__client pb__client--own pb__client--cli">
<path class="iso__halo iso__halo--part" d="M104.9 17 L102.6 18.4 L99.7 19.6 L96.4 20.3 L92.7 20.5 L89.1 20.3 L85.7 19.6 L82.8 18.4 L80.6 17 L79.2 15.3 L78.7 13.5 L79.2 11.7 L80.6 10 L82.8 8.5 L85.7 7.4 L89.1 6.7 L92.7 6.5 L96.4 6.7 L99.7 7.4 L102.6 8.5 L104.9 10 L106.3 11.7 L106.7 13.5 L106.3 15.3 Z"/>
<path class="iso__face iso__face--s1" d="M80.6 -3 L97.9 2 L97.9 3.4 L80.6 -1.6 Z"/>
<path class="iso__face iso__face--s2" d="M104.9 -4 L97.9 2 L97.9 3.4 L104.9 -2.6 Z"/>
<path class="iso__face" d="M87.6 -9 L104.9 -4 L97.9 2 L80.6 -3 Z"/>
<g class="pb__mark" transform="translate(86.3 -10) scale(.36)"><rect class="pb__glyph-face" x="3" y="7" width="30" height="22" rx="2"/><path class="pb__glyph" d="M9 14 l4 4 -4 4"/><line class="pb__glyph" x1="16" y1="22" x2="24" y2="22"/></g>
</g>
<g class="iso__m pb__client pb__client--own pb__client--web">
<path class="iso__halo iso__halo--part" d="M161.2 33.2 L158.9 34.7 L156 35.8 L152.7 36.5 L149 36.7 L145.4 36.5 L142 35.8 L139.1 34.7 L136.9 33.2 L135.5 31.6 L135 29.7 L135.5 27.9 L136.9 26.2 L139.1 24.8 L142 23.7 L145.4 23 L149 22.7 L152.7 23 L156 23.7 L158.9 24.8 L161.2 26.2 L162.6 27.9 L163 29.7 L162.6 31.6 Z"/>
<path class="iso__face iso__face--s1" d="M136.9 13.3 L154.2 18.3 L154.2 19.7 L136.9 14.7 Z"/>
<path class="iso__face iso__face--s2" d="M161.2 12.2 L154.2 18.3 L154.2 19.7 L161.2 13.6 Z"/>
<path class="iso__face" d="M143.9 7.2 L161.2 12.2 L154.2 18.3 L136.9 13.3 Z"/>
<g class="pb__mark" transform="translate(142.5 6.3) scale(.36)"><rect class="pb__glyph-face" x="3" y="6" width="30" height="24" rx="2"/><line class="pb__glyph" x1="3" y1="12" x2="33" y2="12"/><circle class="pb__glyph" cx="8" cy="9" r="1"/><line class="pb__glyph" x1="9" y1="19" x2="27" y2="19"/><line class="pb__glyph" x1="9" y1="24" x2="21" y2="24"/></g>
</g>
<g class="iso__m pb__client pb__client--other pb__client--other">
<path class="iso__halo iso__halo--part" d="M192.9 64.9 L190.7 66.4 L187.8 67.5 L184.4 68.2 L180.8 68.4 L177.2 68.2 L173.8 67.5 L170.9 66.4 L168.7 64.9 L167.3 63.2 L166.8 61.4 L167.3 59.6 L168.7 57.9 L170.9 56.5 L173.8 55.4 L177.2 54.7 L180.8 54.4 L184.4 54.7 L187.8 55.4 L190.7 56.5 L192.9 57.9 L194.3 59.6 L194.8 61.4 L194.3 63.2 Z"/>
<path class="iso__face iso__face--s1" d="M168.6 51.9 L186 56.9 L186 58.3 L168.6 53.3 Z"/>
<path class="iso__face iso__face--s2" d="M193 50.9 L186 56.9 L186 58.3 L193 52.2 Z"/>
<path class="iso__face" d="M175.6 45.9 L193 50.9 L186 56.9 L168.6 51.9 Z"/>
<g class="pb__mark" transform="translate(174.3 44.9) scale(.36)"><circle class="pb__glyph" cx="18" cy="12" r="6"/><path class="pb__glyph" d="M6 32 c0 -8 5 -12 12 -12 s12 4 12 12"/><path class="pb__glyph" d="M26 4 l6 6 M32 4 l-6 6"/></g>
</g>
<g class="pb__wall">
<path class="iso__face iso__face--s1" d="M162.2 47.4 L163.5 47.8 L163.5 54.7 L162.2 54.3 Z"/>
<path class="iso__face iso__face--s2" d="M169.5 42.6 L163.5 47.8 L163.5 54.7 L169.5 49.5 Z"/>
<path class="iso__face" d="M168.2 42.2 L169.5 42.6 L163.5 47.8 L162.2 47.4 Z"/>
</g>
<path class="iso__ln pb__line pb__line--own" d="M32.9 -15.3 L30.9 1.1 L81.7 21.1"/>
<path class="iso__ln pb__line pb__line--own" d="M89.2 0.9 L87.2 17.4 L81.7 21.1"/>
<path class="iso__ln pb__line pb__line--own" d="M145.5 17.2 L143.5 33.6 L81.7 21.1"/>
<path class="iso__ln pb__line pb__line--other" d="M172.1 50.3 L168.7 51.9 L166.1 51.1"/>
<path class="comet iso__comet pb__comet pb__comet--own pb__comet--app" d="M32.9 -15.3 L30.9 1.1 L81.7 21.1"/>
<path class="comet iso__comet pb__comet pb__comet--own pb__comet--cli" d="M89.2 0.9 L87.2 17.4 L81.7 21.1"/>
<path class="comet iso__comet pb__comet pb__comet--own pb__comet--web" d="M145.5 17.2 L143.5 33.6 L81.7 21.1"/>
<path class="comet iso__comet pb__comet pb__comet--other pb__comet--other" d="M172.1 50.3 L168.7 51.9 L166.1 51.1"/>
    </svg>
    <!-- iso:end -->
    <ul class="pb__clients">
      <li class="pb__tag pb__tag--app"><b>Your app</b><small>with your key</small></li>
      <li class="pb__tag pb__tag--cli"><b>The CLI</b><small>with your key</small></li>
      <li class="pb__tag pb__tag--web"><b>The web console</b><small>with your key</small></li>
      <li class="pb__tag pb__tag--other"><b>Anyone else</b><small>without it</small></li>
    </ul>
  </div>
  <div class="pb__readout" aria-live="polite">
    <span><span data-when="yours"><b>Tuned</b> · your band · key accepted</span><span data-when="other"><b>No carrier</b> · key refused</span></span>
    <span>one band per customer</span>
  </div>

  <p class="pb__gpu"><b>RogerAI&rsquo;s cloud.</b> Your add-on and its key stay on it, on a band nobody else is given.</p>

  <div class="pb__keys" data-keyswitch role="group" aria-label="Who is tuning">
    <span>Tune with</span>
    <span class="pb__keyset">
      <button type="button" class="pb__key" data-key="yours" aria-pressed="true">Your key</button>
      <button type="button" class="pb__key" data-key="other" aria-pressed="false">Someone else&rsquo;s</button>
    </span>
  </div>
</IsoFigure>
