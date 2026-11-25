<svelte:options namespace="svg" />
<script>
  // Ping, the mascot, as one SVG group in its 32-unit space, any look the iOS app allows (app/lib/ping.js holds
  // the shapes). Place it with a transform on a wrapper: its base (16, 29.6) is where it stands. Drawn twice for
  // depth (a side copy offset behind, then the face), on a 2:1 ground shadow, the eye in --live or a --ping-*
  // token. Styles: the .pg__* rules in company-crew.css (move them to components.css when a second figure uses
  // Ping). No colour literal, no inline style.
  import { HEADS, BODIES, FEET, WEAR, CUBE, mirror } from "#lib/ping.js";
  /** @type {{ name: string, head?: keyof HEADS, body?: keyof BODIES, eye?: string, wear?: string[], carry?: boolean, look?: number }} */
  let { name, head = "square", body = "classic", eye = "live", wear = [], carry = false, look = 0 } = $props();
  const h = $derived(HEADS[head]);
  const feet = $derived(FEET[body]);
  const brackets = $derived([h.d, mirror(h.d)]);
</script>

{#snippet figure(cls)}
  {#each brackets as d}<path class="{cls} pg__br" stroke-width={h.w} stroke-linecap={h.cap} {d} />{/each}
  {#if feet}{#each [feet[0], feet[1]] as x}<rect class={cls} {x} y={feet[2]} width="2.6" height="2.6" rx=".6" />{/each}{/if}
  <path class={cls} d={BODIES[body]} />
{/snippet}

{#snippet shapes(list)}
  {#each list as [el, cls, attrs]}<svelte:element this={el} class={cls.split(" ").map((c) => "pg__" + c).join(" ")} {...attrs} />{/each}
{/snippet}

<g class="pg__ping pg__ping--{name} pg__eye--{eye}">
  <ellipse class="pg__shadow" cx="16" cy="29.8" rx="5.4" ry="2.7" />
  <g class="pg__depth" transform="translate(.8 .35)">{@render figure("pg__side")}</g>
  {@render figure("pg__ink")}
  <g class="pg__look" transform="translate({look * 1.5} 0)">
    <circle class="pg__eye pg__iris" cx="16" cy="12" r="3" />
    <circle class="pg__spec" cx="15" cy="11" r=".8" />
  </g>
  {#each wear as w}{@render shapes(WEAR[w])}{/each}
  {#if carry}<g class="pg__model" transform="translate(16 22) scale(1.6) translate(-16 -22)">{@render shapes(CUBE)}</g>{/if}
  <circle class="pg__ring" cx="16" cy="12" r="4" />
</g>
