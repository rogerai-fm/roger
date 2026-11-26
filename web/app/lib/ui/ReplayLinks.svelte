<script>
  // Two doors from a model page to the timelapses that show the models at work (2026-11-26, founder ask): the
  // figures live once, on their use-case pages, and a model page links to them with a picture each instead of
  // repeating them. A frame of each film, its scenario in a few words, and where it leads. No number is
  // printed here, so a model page's rule (every number from its card file) is untouched. The camera frame is
  // shown whole (the shot takes the frame's own shape) and carries its credit, as its permission asks.
  const doors = [
    { href: "/research-wildfire.html#timelapse", img: "assets/fire-replay/a27.webp", w: 800, h: 486,
      kicker: "Wildfire, a recorded timelapse", title: "A fire starts, one frame a minute",
      line: "A real camera sequence, with what each model said at every frame." },
    { href: "/research-oil-gas.html#theft-timelapse", img: "assets/oilgas-replay/gate-arrive.webp", w: 960, h: 560,
      kicker: "Oil and gas, a simulated afternoon", title: "A tap on the line",
      line: "A truck at the gate, a flow that drifts, a sensor that goes flat." },
  ];
</script>

<nav class="wm-doors" aria-label="See the models at work">
  {#each doors as d}
    <a class="wm-door" href={d.href}>
      <span class="wm-door__shot"><img src={d.img} width={d.w} height={d.h} alt="" loading="lazy" decoding="async" /><i aria-hidden="true"></i></span>
      <span class="wm-door__kicker">{d.kicker}</span>
      <b>{d.title}</b>
      <span class="wm-door__line">{d.line}</span>
    </a>
  {/each}
  <p class="wm-credit">The valley frame: data courtesy of HPWREN, University of California San Diego. The gate picture is generated, not a photograph.</p>
</nav>
