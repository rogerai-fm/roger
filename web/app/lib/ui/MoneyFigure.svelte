<script>
  import IsoFigure from "#lib/ui/IsoFigure.svelte";
  // The money, drawn once: who pays whom on the RogerAI network (site audit 2026-11-13, fix 4).
  // Written the way app/lib/ui figures are (TuningFigure.svelte): runes, props in, no <style>
  // block (its rules are src/styles/money-figure.css, a sheet the page lists), no colour
  // literal, no inline style, every word in HTML, the drawing in inline SVG with no <text>.
  // Since 2026-11-22 the three party glyphs are in the isometric hand (the .iso__* rules of
  // components.css): the wallet a slab with its lid lifted and a token on its lip, the station
  // the Wave console every other figure draws (its socket lamp the one red: the party that
  // keeps 90%), the tower the tower page's tower, small, its lamp in ink. The RogerAI mark
  // stays flat. Their SVG is written between the iso markers by computer-scientist/article/
  // fig_money_glyphs_gen.py (--svelte); the stage, the fork, the shares and the words are the
  // HTML grid they always were.
  //
  // WHAT IT DRAWS, and the page each number comes from (never a price a station controls):
  //   per token                 pricing §2 "Unit"; the FAQ "What does it cost to use a model?"
  //   90% station / 10% RogerAI pricing §4 "Your GPU" (the 10% platform fee); the FAQ; tos
  //   90 / 5 / 5 through a Tower pricing §4 "Through a Tower"; tower §4; docs/tower.md §9
  //   list + 10% on a curated station pricing §1 "A curated station", §4 "Your contracts"
  //   free models cost nothing  pricing §1 "A free station", §6 "What is not billed"
  //   custom tuning: price on request  founder decision 2026-11-13 (no public number)
  //
  // size="full"  (pricing §2): both paid paths drawn, notes under every box, three cases.
  // size="small" (the homepage, later): the direct path only, no notes, the Tower path
  //              folded into the case strip as "90 / 5 / 5".
  //
  // GUARDS it meets as written (web/test): no <img> (perf-guards); every colour a token and
  // every rule in money-figure.css led by .mf (design-system); no SVG <text>, the one red word
  // is var(--live-text) (contrast); role="group" named by its caption, nothing interactive
  // under aria-hidden, no heading inside (a11y-guards); max-width 100% throughout and a grid
  // of minmax(0, …) columns (layout-overflow); no em dash. Since 2026-11-15 it draws itself in and
  // runs a token (a comet) from the wallet to each recipient; reduced motion and print get the still.
  let {
    fig = 1,
    size = "full",                 // "full" | "small"
    caption = "You pay per token. On a paid model the station keeps 90% and RogerAI 10%; through a Tower it is 90, 5 and 5. A curated station is the provider's list price plus 10%. Free models cost nothing. Custom tuning is priced on request.",
  } = $props();
  const small = $derived(size === "small");
  // the fork: one trunk from the wallet, one branch per row, drawn to the cell
  // (preserveAspectRatio="none"), so every branch lands on the centre of its row
  const fork = (n) => `M0 ${n * 50} H85 ` + Array.from({ length: n }, (_, i) => `M85 ${n * 50} V${(i + 0.5) * 100} H100`).join(" ");
  // one comet per recipient: the trunk and that branch as one line, so a token can be seen leaving the wallet and arriving
  const run = (n, i) => `M0 ${n * 50} H85 V${(i + 0.5) * 100} H100`;
</script>

{#snippet wallet()}
  <div class="mf__from">
    <!-- iso:wallet -->
    <svg class="iso__svg mf__glyph mf__glyph--wallet" viewBox="-17.1 -9.6 52 32.9" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M27.9 13.8 L24.4 16.1 L19.9 17.9 L14.6 19 L8.9 19.3 L3.2 19 L-2.1 17.9 L-6.7 16.1 L-10.2 13.8 L-12.4 11.2 L-13.1 8.3 L-12.4 5.5 L-10.2 2.8 L-6.7 0.6 L-2.1 -1.2 L3.2 -2.3 L8.9 -2.7 L14.6 -2.3 L19.9 -1.2 L24.4 0.6 L27.9 2.8 L30.1 5.5 L30.9 8.3 L30.1 11.2 Z"/>
<path class="iso__face iso__face--s1" d="M-10 3.5 L17.7 11.5 L17.7 16.7 L-10 8.7 Z"/>
<path class="iso__face iso__face--s2" d="M27.7 2.8 L17.7 11.5 L17.7 16.7 L27.7 8 Z"/>
<path class="iso__face" d="M0 -5.2 L27.7 2.8 L17.7 11.5 L-10 3.5 Z"/>
<path class="iso__ln iso__ln--faint" d="M-7.4 7.5 L15.1 14"/>
<path class="iso__face iso__face--s1" d="M0 -5.2 L27.7 2.8 L21.2 2.4 L-6.5 -5.6 Z"/>
<path class="iso__face iso__face--s2" d="M27.7 2.8 L21.2 2.4 L21.2 3.2 L27.7 3.7 Z"/>
<path class="iso__face" d="M16.1 8.7 L15.4 9.1 L14.4 9.3 L13.3 9.4 L12.3 9.2 L11.4 8.8 L10.9 8.3 L10.9 7.8 L11.2 7.3 L11.9 6.9 L12.9 6.6 L14 6.6 L15.1 6.8 L15.9 7.1 L16.4 7.6 L16.4 8.2 Z"/>
<path class="iso__face iso__face--cav" d="M14.7 8.3 L14.1 8.5 L13.4 8.6 L12.8 8.4 L12.5 8 L12.6 7.7 L13.2 7.4 L13.9 7.4 L14.5 7.6 L14.8 7.9 Z"/>
</svg>
    <!-- /iso:wallet -->
    <span class="mf__who">Your wallet</span>
    <span class="mf__note">no subscription; it moves per token</span>
  </div>
{/snippet}

{#snippet station(share)}
  <div class="mf__to">
    <!-- iso:station -->
    <svg class="iso__svg mf__glyph mf__glyph--station" viewBox="-32.3 -25.9 70 55.5" aria-hidden="true" focusable="false">
<path class="mf__tip" d="M-31.3 1.8 h9 M-26.3 -3.2 L-21.3 1.8 L-26.3 6.8"/>
<path class="iso__halo" d="M30.5 14.6 L26.7 17.1 L21.7 19 L15.9 20.2 L9.7 20.6 L3.5 20.2 L-2.3 19 L-7.2 17.1 L-11.1 14.6 L-13.5 11.7 L-14.3 8.6 L-13.5 5.5 L-11.1 2.6 L-7.2 0.1 L-2.3 -1.8 L3.5 -3 L9.7 -3.4 L15.9 -3 L21.7 -1.8 L26.7 0.1 L30.5 2.6 L32.9 5.5 L33.7 8.6 L32.9 11.7 Z"/>
<path class="iso__face iso__face--s1" d="M-10 -6.9 L19.4 1.6 L19.4 17.2 L-10 8.7 Z"/>
<path class="iso__face iso__face--s2" d="M29.4 -7.1 L19.4 1.6 L19.4 17.2 L29.4 8.5 Z"/>
<path class="iso__face" d="M0 -15.6 L29.4 -7.1 L19.4 1.6 L-10 -6.9 Z"/>
<path class="iso__face iso__face--s1" d="M-6.4 -8.8 L17.8 -1.8 L17.8 0 L-6.4 -7 Z"/>
<path class="iso__face iso__face--s2" d="M25.8 -8.7 L17.8 -1.8 L17.8 0 L25.8 -7 Z"/>
<path class="iso__face" d="M1.6 -15.7 L25.8 -8.7 L17.8 -1.8 L-6.4 -8.8 Z"/>
<path class="iso__face iso__face--s2" d="M-6.5 4.5 L3.9 7.5 L3.9 2.7 L-6.5 -0.3 Z"/>
<path class="iso__ln iso__ln--faint" d="M-4.8 4.3 L-4.8 0.9"/>
<path class="iso__ln iso__ln--faint" d="M-3.1 4.8 L-3.1 1.4"/>
<path class="iso__ln iso__ln--faint" d="M-1.3 5.3 L-1.3 1.9"/>
<path class="iso__ln iso__ln--faint" d="M0.4 5.8 L0.4 2.4"/>
<path class="iso__ln iso__ln--faint" d="M2.1 6.3 L2.1 2.9"/>
<path class="iso__face" d="M14.4 8.1 L14.3 7.1 L13.8 6.2 L13 5.3 L12.1 4.7 L11.2 4.5 L10.3 4.5 L9.5 4.9 L9 5.6 L8.9 6.5 L9 7.5 L9.5 8.5 L10.3 9.3 L11.2 9.9 L12.1 10.2 L13 10.1 L13.8 9.7 L14.3 9 Z"/>
<path class="iso__ln" d="M11.7 7.3 L13.4 5.9"/>
<path class="iso__face iso__face--cav" d="M-5.7 -1.3 L9.9 3.2 L9.9 1.1 L-5.7 -3.4 Z"/>
<path class="iso__ln iso__ln--faint" d="M-4.8 7.9 L14.2 13.4"/>
<path class="iso__ln" d="M22.1 -7.5 L22.1 -17"/>
<circle class="iso__lamp" cx="22.1" cy="-17" r="1.4"/>
<circle class="iso__lamp iso__lamp--on" cx="16" cy="2.7" r="1.9"/>
</svg>
    <!-- /iso:station -->
    <span class="mf__share">{share}</span>
    <span class="mf__who">Station</span>
    <span class="mf__note">the owner's machine; they set the price</span>
  </div>
{/snippet}

{#snippet tower(share)}
  <div class="mf__to">
    <!-- iso:tower -->
    <svg class="iso__svg mf__glyph mf__glyph--tower" viewBox="-36.5 -32.3 70 55.5" aria-hidden="true" focusable="false">
<path class="mf__tip" d="M-35.5 -4.5 h9 M-30.5 -9.5 L-25.5 -4.5 L-30.5 0.5"/>
<path class="iso__halo" d="M21.1 14.7 L18.2 16.6 L14.5 18 L10.1 18.9 L5.5 19.2 L0.8 18.9 L-3.5 18 L-7.2 16.6 L-10.1 14.7 L-11.9 12.6 L-12.5 10.2 L-11.9 7.9 L-10.1 5.7 L-7.2 3.9 L-3.5 2.5 L0.8 1.6 L5.5 1.2 L10.1 1.6 L14.5 2.5 L18.2 3.9 L21.1 5.7 L22.9 7.9 L23.5 10.2 L22.9 12.6 Z"/>
<path class="iso__face iso__face--s1" d="M-9.5 9.7 L9.5 15.2 L9.5 17.8 L-9.5 12.3 Z"/>
<path class="iso__face iso__face--s2" d="M20.5 5.6 L9.5 15.2 L9.5 17.8 L20.5 8.2 Z"/>
<path class="iso__face" d="M1.5 0.1 L20.5 5.6 L9.5 15.2 L-9.5 9.7 Z"/>
<path class="iso__face iso__face--s1" d="M-1.3 -14.8 L7.3 -12.3 L7.3 11.1 L-1.3 8.6 Z"/>
<path class="iso__face iso__face--s2" d="M12.3 -16.7 L7.3 -12.3 L7.3 11.1 L12.3 6.7 Z"/>
<path class="iso__face" d="M3.7 -19.2 L12.3 -16.7 L7.3 -12.3 L-1.3 -14.8 Z"/>
<path class="iso__face iso__face--s1" d="M-2.7 -16 L7.7 -13 L7.7 -11.6 L-2.7 -14.6 Z"/>
<path class="iso__face iso__face--s2" d="M13.7 -18.2 L7.7 -13 L7.7 -11.6 L13.7 -16.8 Z"/>
<path class="iso__face" d="M3.3 -21.2 L13.7 -18.2 L7.7 -13 L-2.7 -16 Z"/>
<path class="iso__ln iso__ln--faint" d="M0.4 -3.1 L5.6 -1.6"/>
<path class="iso__ln iso__ln--faint" d="M0.4 -5.7 L5.6 -4.2"/>
<path class="iso__ln iso__ln--faint" d="M0.4 -8.3 L5.6 -6.8"/>
<path class="iso__ln iso__ln--faint" d="M0.4 -10.9 L5.6 -9.4"/>
<path class="iso__face iso__face--cav" d="M1.3 5.8 L4.7 6.8 L4.7 4.7 L1.3 3.7 Z"/>
<path class="iso__ln" d="M6 -17.6 L6 -28.3"/>
<circle class="iso__lamp" cx="6" cy="-28.3" r="1.5"/>
</svg>
    <!-- /iso:tower -->
    <span class="mf__share">{share}</span>
    <span class="mf__who">Tower</span>
    <span class="mf__note">a relay somebody else runs</span>
  </div>
{/snippet}

{#snippet rogerai(share)}
  <div class="mf__to">
    <svg class="mf__glyph" viewBox="0 0 96 72" aria-hidden="true" focusable="false">
      <path class="mf__tip" d="M0 36 H9 M5 31 L10 36 L5 41"/>
      <g class="mf__brand" transform="translate(29 16) scale(1.25)">
        <path d="M4 6 h9 v4.5 H8.5 v11 H13 V26 H4 Z"/>
        <path d="M28 6 h-9 v4.5 H23.5 v11 H19 V26 h9 Z"/>
        <circle class="mf__lamp" cx="16" cy="16" r="5"/>
      </g>
    </svg>
    <span class="mf__share">{share}</span>
    <span class="mf__who">RogerAI</span>
    <span class="mf__note">the fee: routing, the receipt, one wallet</span>
  </div>
{/snippet}

<IsoFigure lead="mf" {fig} {caption} extra={small ? 'mf--small' : 'mf--full'}>

  <!-- a paid model, direct: the wallet pays per token; 90 to the station, 10 to RogerAI -->
  <div class="mf__path">
    <span class="mf__tag">A paid model</span>
    <div class="mf__stage mf__stage--2">
      {@render wallet()}
      <div class="mf__fork" aria-hidden="true">
        <svg class="mf__forksvg" viewBox="0 0 100 200" preserveAspectRatio="none"><path class="mf__wire" d={fork(2)}/><path class="comet mf__comet" d={run(2, 0)}/><path class="comet mf__comet" d={run(2, 1)}/></svg>
        <span class="mf__onwire">per token</span>
      </div>
      {@render station("90%")}
      {@render rogerai("10%")}
    </div>
  </div>

  {#if !small}
  <!-- the same payment through a Tower somebody runs: 90 / 5 / 5 -->
  <div class="mf__path">
    <span class="mf__tag">A paid model, through a Tower</span>
    <div class="mf__stage mf__stage--3">
      {@render wallet()}
      <div class="mf__fork" aria-hidden="true">
        <svg class="mf__forksvg" viewBox="0 0 100 300" preserveAspectRatio="none"><path class="mf__wire" d={fork(3)}/><path class="comet mf__comet" d={run(3, 0)}/><path class="comet mf__comet" d={run(3, 1)}/><path class="comet mf__comet" d={run(3, 2)}/></svg>
        <span class="mf__onwire">per token</span>
      </div>
      {@render station("90%")}
      {@render tower("5%")}
      {@render rogerai("5%")}
    </div>
  </div>
  {/if}

  <!-- the cases the two drawings do not cover -->
  <ul class="mf__cases">
    {#if small}
    <li><b>Through a Tower</b><span>90 / 5 / 5: station, Tower, RogerAI</span></li>
    {/if}
    <li><b>A free model</b><span>costs nothing. No account needed.</span></li>
    <li><b>A curated station</b><span>the provider's list price + 10%</span></li>
    <li><b>Custom tuning</b><span>trained on RogerAI's GPUs, served privately on a band only you can reach. <i class="mf__ask">Price on request.</i></span></li>
  </ul>

</IsoFigure>
