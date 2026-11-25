<script>
  import IsoFigure from "#lib/ui/IsoFigure.svelte";
  // The band, tuned (homepage §5, 2026-11-20; in the isometric hand 2026-11-23): your tools talk
  // to one local address; RogerAI picks the strongest station, meters the tokens and signs the
  // receipt; when a station fades, the next one answers at the same address. The homepage's one
  // drawing of routing, failover and metering (the hero's three labels), beside the two commands.
  //
  // One scene on a ground slab: your laptop at the left with its lid up, one wire on the floor to
  // the relay (RogerAI, a low box with three ports), a trunk and three branches to the stations
  // at the right, stacked in depth: a console in front (the strongest), a console behind it (the
  // next), the Tower at the back. The step list is <ol data-scrub data-scrub-auto>, so scrub.js
  // (the runtime) inserts the range control, plays the three steps itself until the reader
  // touches it, and the picture reads the tuned step through the figure's data-step (scrub.js): 1 the first station's lamp
  // is lit and the request runs the route there and back as the comet, the receipt slip lands
  // beside the laptop; 2 it fades (lamp out, the part dimmed, its branch dashed, a cross on its
  // front, no comet, the slip faint); 3 the second station answers and the address plate flashes
  // once because nothing on your side changed. data-draw: the figure draws itself in the first
  // time it is seen (draw.js). Reduced motion, no JS and print get the still: step 1.
  //
  // Written the way app/lib/ui figures are (HoldFigure.svelte): runes, props in, no <style> block
  // (its rules are src/styles/band-figure.css, a sheet the page lists; the hand's tones and lines
  // are the .iso__* rules of components.css), no colour literal, no inline style, every word in
  // HTML over the stage, the drawing in inline SVG with no <text>. The SVG between the iso
  // markers is written by computer-scientist/article/fig_band_gen.py (--svelte), which also
  // prints every label percent in band-figure.css. Lead class .bf.
  //
  // GUARDS it meets as written: no <img>; no colour literal; every rule in band-figure.css led
  // by .bf; no SVG <text>; role="group" named by its caption; nothing interactive under
  // aria-hidden; no heading inside; minmax(0, ...) columns; --d-* and --motion only; no em dash.
  let {
    fig = 4,
    caption = "Your tools talk to one local address. RogerAI picks the strongest station, meters the tokens and signs the receipt. When a station fades, the next one answers at the same address.",
  } = $props();
  const steps = [
    { name: "Tune in", note: "The strongest station answers. You pay per token and get a signed receipt." },
    { name: "It fades", note: "That station goes off the air in the middle of your work." },
    { name: "Re-tuned", note: "The next one answers. Same address, same key. Nothing in your code changes." },
  ];
</script>

<IsoFigure lead="bf" {fig} {caption} {steps} stepsClass="bf__steps" stepsLabel="What happens on the band, step by step">
  <div class="iso__stage bf__stage">
    <div class="bf__pic" aria-hidden="true">
      <!-- iso:begin -->
      <svg class="iso__svg" viewBox="-56.2 -13.5 358.4 126.9" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-6.2 2.8 L224.2 69.3 L180.2 107.4 L-50.2 40.9 Z"/>
<g class="bf__ground">
<path class="iso__face iso__face--s1" d="M-42 36.4 L176.2 99.4 L176.2 102.8 L-42 39.8 Z"/>
<path class="iso__face iso__face--s2" d="M218.2 63 L176.2 99.4 L176.2 102.8 L218.2 66.5 Z"/>
<path class="iso__face" d="M0 0 L218.2 63 L176.2 99.4 L-42 36.4 Z"/>
</g>
<g class="bf__you">
<path class="iso__halo" d="M20.4 31.3 L16.6 33.8 L11.7 35.7 L5.9 36.9 L-0.3 37.3 L-6.6 36.9 L-12.3 35.7 L-17.3 33.8 L-21.1 31.3 L-23.5 28.4 L-24.3 25.3 L-23.5 22.2 L-21.1 19.3 L-17.3 16.8 L-12.3 14.9 L-6.6 13.7 L-0.3 13.3 L5.9 13.7 L11.7 14.9 L16.6 16.8 L20.4 19.3 L22.8 22.2 L23.7 25.3 L22.8 28.4 Z"/>
<path class="iso__face iso__face--s1" d="M-20.6 23.9 L8.9 32.4 L8.9 34.3 L-20.6 25.8 Z"/>
<path class="iso__face iso__face--s2" d="M19.9 22.9 L8.9 32.4 L8.9 34.3 L19.9 24.8 Z"/>
<path class="iso__face" d="M-9.6 14.4 L19.9 22.9 L8.9 32.4 L-20.6 23.9 Z"/>
<path class="iso__ln iso__ln--faint" d="M-8.7 18.7 L12 24.7"/>
<path class="iso__ln iso__ln--faint" d="M-10.7 20.4 L10 26.4"/>
<path class="iso__ln iso__ln--faint" d="M-12.7 22.1 L8 28.1"/>
<path class="iso__face" d="M-9.6 14.4 L19.9 22.9 L23.4 1 L-6.1 -7.5 Z"/>
<path class="iso__face iso__face--s2" d="M19.9 22.9 L23.4 1 L23.8 1.1 L20.3 23.1 Z"/>
<path class="iso__face iso__face--s1" d="M-7.1 13 L18 20.3 L20.8 2.8 L-4.3 -4.5 Z"/>
<path class="iso__ln iso__ln--lo" d="M-2.2 7.8 L1.6 6.7 L-1.5 3.6"/>
<path class="iso__ln iso__ln--lo" d="M3.1 8.4 L10.1 10.4"/>
</g>
<g class="iso__m bf__receipt">
<path class="iso__face iso__face--s1" d="M12.8 34.1 L23.2 37.1 L23.2 37.9 L12.8 34.9 Z"/>
<path class="iso__face iso__face--s2" d="M27.2 33.6 L23.2 37.1 L23.2 37.9 L27.2 34.4 Z"/>
<path class="iso__face" d="M16.8 30.6 L27.2 33.6 L23.2 37.1 L12.8 34.1 Z"/>
<path class="iso__ln iso__ln--lo" d="M15.8 33.5 L22.7 35.5"/>
<path class="iso__ln iso__ln--lo" d="M16.8 32.7 L23.7 34.7"/>
</g>
<g class="bf__relay">
<path class="iso__halo" d="M95.8 53.1 L91.6 55.7 L86.3 57.8 L80 59.1 L73.3 59.6 L66.5 59.1 L60.3 57.8 L54.9 55.7 L50.7 53.1 L48.1 49.9 L47.3 46.6 L48.1 43.2 L50.7 40.1 L54.9 37.4 L60.3 35.3 L66.5 34 L73.3 33.6 L80 34 L86.3 35.3 L91.6 37.4 L95.8 40.1 L98.4 43.2 L99.3 46.6 L98.4 49.9 Z"/>
<path class="iso__face iso__face--s1" d="M52.7 38.6 L83.9 47.6 L83.9 55.4 L52.7 46.4 Z"/>
<path class="iso__face iso__face--s2" d="M93.9 38.9 L83.9 47.6 L83.9 55.4 L93.9 46.7 Z"/>
<path class="iso__face" d="M62.7 29.9 L93.9 38.9 L83.9 47.6 L52.7 38.6 Z"/>
<path class="iso__face iso__face--s1" d="M55.2 37.2 L82.9 45.2 L82.9 46.4 L55.2 38.4 Z"/>
<path class="iso__face iso__face--s2" d="M91.4 37.9 L82.9 45.2 L82.9 46.4 L91.4 39.1 Z"/>
<path class="iso__face" d="M63.7 29.9 L91.4 37.9 L82.9 45.2 L55.2 37.2 Z"/>
<path class="iso__face iso__face--cav" d="M57 45 L60.1 45.9 L60.1 43.7 L57 42.8 Z"/>
<path class="iso__face iso__face--cav" d="M62.2 46.5 L65.3 47.4 L65.3 45.2 L62.2 44.3 Z"/>
<path class="iso__face iso__face--cav" d="M67.4 48 L70.5 48.9 L70.5 46.7 L67.4 45.8 Z"/>
<path class="iso__ln iso__ln--faint" d="M56.1 41.1 L80.4 48.1"/>
</g>
<g class="bf__fan">
<path class="iso__way bf__way bf__way--wire" d="M14.4 29 L57.7 41.5"/>
<path class="iso__way bf__way bf__way--trunk" d="M88.9 50.5 L107.9 56"/>
<path class="iso__way bf__way bf__way--1" d="M107.9 56 L96.9 65.6 L116 71.1"/>
<path class="iso__way bf__way bf__way--2" d="M107.9 56 L110.9 53.4 L145.5 63.4"/>
<path class="iso__way bf__way bf__way--3" d="M107.9 56 L122.9 43 L173.1 57.5"/>
</g>
<path class="comet iso__comet bf__comet bf__comet--out bf__route-1" d="M14.4 29 L57.7 41.5 L88.9 50.5 L107.9 56 L96.9 65.6 L116 71.1"/>
<path class="comet iso__comet bf__comet bf__comet--back bf__route-1" d="M116 71.1 L96.9 65.6 L107.9 56 L88.9 50.5 L57.7 41.5 L14.4 29"/>
<path class="comet iso__comet bf__comet bf__comet--out bf__route-2" d="M14.4 29 L57.7 41.5 L88.9 50.5 L107.9 56 L110.9 53.4 L145.5 63.4"/>
<path class="comet iso__comet bf__comet bf__comet--back bf__route-2" d="M145.5 63.4 L110.9 53.4 L107.9 56 L88.9 50.5 L57.7 41.5 L14.4 29"/>
<path class="comet iso__comet bf__comet bf__comet--out bf__route-3" d="M14.4 29 L57.7 41.5 L88.9 50.5 L107.9 56 L122.9 43 L173.1 57.5"/>
<path class="comet iso__comet bf__comet bf__comet--back bf__route-3" d="M173.1 57.5 L122.9 43 L107.9 56 L88.9 50.5 L57.7 41.5 L14.4 29"/>
<g class="iso__m bf__station bf__station--3">
<path class="iso__halo" d="M198.2 65.3 L195.4 67.2 L191.7 68.6 L187.3 69.5 L182.7 69.8 L178 69.5 L173.7 68.6 L169.9 67.2 L167.1 65.3 L165.3 63.1 L164.7 60.8 L165.3 58.5 L167.1 56.3 L169.9 54.4 L173.7 53 L178 52.1 L182.7 51.8 L187.3 52.1 L191.7 53 L195.4 54.4 L198.2 56.3 L200 58.5 L200.7 60.8 L200 63.1 Z"/>
<path class="iso__face iso__face--s1" d="M167.6 60.2 L186.7 65.7 L186.7 68.3 L167.6 62.8 Z"/>
<path class="iso__face iso__face--s2" d="M197.7 56.2 L186.7 65.7 L186.7 68.3 L197.7 58.8 Z"/>
<path class="iso__face" d="M178.6 50.7 L197.7 56.2 L186.7 65.7 L167.6 60.2 Z"/>
<path class="iso__face iso__face--s1" d="M175.8 35.7 L184.5 38.2 L184.5 61.6 L175.8 59.1 Z"/>
<path class="iso__face iso__face--s2" d="M189.5 33.9 L184.5 38.2 L184.5 61.6 L189.5 57.3 Z"/>
<path class="iso__face" d="M180.8 31.4 L189.5 33.9 L184.5 38.2 L175.8 35.7 Z"/>
<path class="iso__face iso__face--s1" d="M174.5 34.5 L184.9 37.5 L184.9 38.9 L174.5 35.9 Z"/>
<path class="iso__face iso__face--s2" d="M190.9 32.3 L184.9 37.5 L184.9 38.9 L190.9 33.7 Z"/>
<path class="iso__face" d="M180.5 29.3 L190.9 32.3 L184.9 37.5 L174.5 34.5 Z"/>
<path class="iso__ln iso__ln--faint" d="M177.6 47.5 L182.8 49"/>
<path class="iso__ln iso__ln--faint" d="M177.6 44.9 L182.8 46.4"/>
<path class="iso__ln iso__ln--faint" d="M177.6 42.3 L182.8 43.8"/>
<path class="iso__ln iso__ln--faint" d="M177.6 39.7 L182.8 41.2"/>
<path class="iso__face iso__face--cav" d="M178.4 56.4 L181.9 57.4 L181.9 55.2 L178.4 54.2 Z"/>
<path class="iso__ln" d="M183.2 33 L183.2 22.3"/>
<circle class="iso__lamp bf__lamp" cx="183.2" cy="22.3" r="1.5"/>
</g>
<g class="iso__m bf__station bf__station--2">
<path class="iso__halo" d="M181.1 74.2 L177.2 76.7 L172.3 78.6 L166.5 79.8 L160.3 80.2 L154.1 79.8 L148.3 78.6 L143.3 76.7 L139.5 74.2 L137.1 71.3 L136.3 68.2 L137.1 65.1 L139.5 62.2 L143.3 59.7 L148.3 57.8 L154.1 56.6 L160.3 56.2 L166.5 56.6 L172.3 57.8 L177.2 59.7 L181.1 62.2 L183.4 65.1 L184.3 68.2 L183.4 71.3 Z"/>
<path class="iso__face iso__face--s1" d="M140.5 52.7 L170 61.2 L170 76.8 L140.5 68.3 Z"/>
<path class="iso__face iso__face--s2" d="M180 52.5 L170 61.2 L170 76.8 L180 68.1 Z"/>
<path class="iso__face" d="M150.5 44 L180 52.5 L170 61.2 L140.5 52.7 Z"/>
<path class="iso__face iso__face--s1" d="M144.1 50.8 L168.4 57.8 L168.4 59.6 L144.1 52.6 Z"/>
<path class="iso__face iso__face--s2" d="M176.4 50.9 L168.4 57.8 L168.4 59.6 L176.4 52.7 Z"/>
<path class="iso__face" d="M152.1 43.9 L176.4 50.9 L168.4 57.8 L144.1 50.8 Z"/>
<path class="iso__face iso__face--s2" d="M144 64.1 L154.4 67.1 L154.4 62.3 L144 59.3 Z"/>
<path class="iso__ln iso__ln--faint" d="M145.7 63.9 L145.7 60.5"/>
<path class="iso__ln iso__ln--faint" d="M147.5 64.4 L147.5 61"/>
<path class="iso__ln iso__ln--faint" d="M149.2 64.9 L149.2 61.5"/>
<path class="iso__ln iso__ln--faint" d="M150.9 65.4 L150.9 62"/>
<path class="iso__ln iso__ln--faint" d="M152.7 65.9 L152.7 62.5"/>
<path class="iso__face" d="M165 67.8 L164.8 66.8 L164.3 65.8 L163.6 65 L162.7 64.4 L161.7 64.1 L160.8 64.2 L160.1 64.6 L159.6 65.3 L159.4 66.2 L159.6 67.2 L160.1 68.1 L160.8 69 L161.7 69.5 L162.7 69.8 L163.6 69.8 L164.3 69.4 L164.8 68.7 Z"/>
<path class="iso__ln" d="M162.2 67 L163.9 65.5"/>
<path class="iso__face iso__face--cav" d="M144.9 58.3 L160.5 62.8 L160.5 60.7 L144.9 56.2 Z"/>
<path class="iso__ln iso__ln--faint" d="M145.7 67.5 L164.8 73"/>
<path class="iso__ln" d="M172.7 52.2 L172.7 42.6"/>
<circle class="iso__lamp" cx="172.7" cy="42.6" r="1.4"/>
<circle class="iso__lamp bf__lamp" cx="166.5" cy="62.4" r="1.9"/>
<path class="bf__stop" d="M146.6 67.4 L163.9 62"/>
<path class="bf__stop" d="M146.6 57 L163.9 72.4"/>
</g>
<g class="iso__m bf__station bf__station--1">
<path class="iso__halo" d="M151.5 81.8 L147.6 84.3 L142.7 86.2 L136.9 87.4 L130.7 87.8 L124.5 87.4 L118.7 86.2 L113.7 84.3 L109.9 81.8 L107.5 78.9 L106.7 75.8 L107.5 72.7 L109.9 69.8 L113.7 67.3 L118.7 65.4 L124.5 64.2 L130.7 63.8 L136.9 64.2 L142.7 65.4 L147.6 67.3 L151.5 69.8 L153.9 72.7 L154.7 75.8 L153.9 78.9 Z"/>
<path class="iso__face iso__face--s1" d="M111 60.3 L140.4 68.8 L140.4 84.4 L111 75.9 Z"/>
<path class="iso__face iso__face--s2" d="M150.4 60.2 L140.4 68.8 L140.4 84.4 L150.4 75.7 Z"/>
<path class="iso__face" d="M121 51.7 L150.4 60.2 L140.4 68.8 L111 60.3 Z"/>
<path class="iso__face iso__face--s1" d="M114.6 58.5 L138.8 65.5 L138.8 67.2 L114.6 60.2 Z"/>
<path class="iso__face iso__face--s2" d="M146.8 58.5 L138.8 65.5 L138.8 67.2 L146.8 60.3 Z"/>
<path class="iso__face" d="M122.6 51.5 L146.8 58.5 L138.8 65.5 L114.6 58.5 Z"/>
<path class="iso__face iso__face--s2" d="M114.4 71.7 L124.8 74.7 L124.8 69.9 L114.4 66.9 Z"/>
<path class="iso__ln iso__ln--faint" d="M116.2 71.5 L116.2 68.1"/>
<path class="iso__ln iso__ln--faint" d="M117.9 72 L117.9 68.6"/>
<path class="iso__ln iso__ln--faint" d="M119.6 72.5 L119.6 69.1"/>
<path class="iso__ln iso__ln--faint" d="M121.3 73 L121.3 69.6"/>
<path class="iso__ln iso__ln--faint" d="M123.1 73.5 L123.1 70.1"/>
<path class="iso__face" d="M135.4 75.4 L135.2 74.4 L134.7 73.4 L134 72.6 L133.1 72 L132.1 71.7 L131.2 71.8 L130.5 72.2 L130 72.9 L129.8 73.8 L130 74.8 L130.5 75.7 L131.2 76.6 L132.1 77.2 L133.1 77.4 L134 77.4 L134.7 77 L135.2 76.3 Z"/>
<path class="iso__ln" d="M132.6 74.6 L134.3 73.1"/>
<path class="iso__face iso__face--cav" d="M115.3 65.9 L130.9 70.4 L130.9 68.3 L115.3 63.8 Z"/>
<path class="iso__ln iso__ln--faint" d="M116.2 75.2 L135.2 80.7"/>
<path class="iso__ln" d="M143.1 59.8 L143.1 50.2"/>
<circle class="iso__lamp" cx="143.1" cy="50.2" r="1.4"/>
<circle class="iso__lamp bf__lamp" cx="136.9" cy="70" r="1.9"/>
<path class="bf__stop" d="M117 75.1 L134.3 69.7"/>
<path class="bf__stop" d="M117 64.7 L134.3 80.1"/>
</g>
<g class="bf__leads">
<path class="iso__ln iso__ln--lo bf__lead bf__lead--3" d="M183.2 22.3 L225.2 22.3"/>
<path class="iso__ln iso__ln--lo bf__lead bf__lead--2" d="M172.7 52.2 L225.2 52.2"/>
<path class="iso__ln iso__ln--lo bf__lead bf__lead--1" d="M143.1 59.8 L225.2 79.2"/>
</g>
      </svg>
      <!-- iso:end -->

      <!-- the words over the stage, placed by the generator -->
      <div class="bf__word bf__word--you">
        <span class="bf__who">Your tools</span>
        <code class="bf__addr">127.0.0.1:4141</code>
        <span class="bf__note">one address, one key, unchanged</span>
      </div>
      <span class="bf__word bf__onwire iso__ko">per token</span>
      <div class="bf__word bf__word--relay">
        <span class="bf__who">RogerAI</span>
        <span class="bf__note">routes, meters, signs the receipt</span>
      </div>
      <div class="bf__word bf__word--1">
        <span class="bf__who iso__ko iso__ko--ground">A station</span>
        <span class="bf__note">someone's GPU, the strongest right now</span>
      </div>
      <div class="bf__word bf__word--2">
        <span class="bf__who iso__ko iso__ko--ground">Another station</span>
        <span class="bf__note">someone's GPU, next strongest</span>
      </div>
      <div class="bf__word bf__word--3">
        <span class="bf__who iso__ko iso__ko--ground">A Tower</span>
        <span class="bf__note">a relay somebody runs</span>
      </div>
    </div>
  </div>

</IsoFigure>
