<svelte:options namespace="svg" />
<script>
  // The brand-mark Ping (app/lib/ping.js MARKS) as one SVG group in its 64-unit space: optional radio waves,
  // the body in currentColor, an optional glow, and the eye (the --live lamp, or a hollow ring off the air).
  // Each place keeps its own class names through `prefix` (ping-mark__ on the homepage, tf__ping- in the custom
  // tuning figures, ping__ on the 404), so the sheets and ping-eye.js that animate them are unchanged.
  // Colours are the sheets' (base.css, custom-tuning.css, notfound.css): no literal here.
  import { MARKS } from "#lib/ping.js";
  /** @type {{ prefix: string, shape?: keyof MARKS, eyeY?: number, waveY?: number|null, glow?: boolean, eye?: "live"|"hollow", root?: string, transform?: string }} */
  let { prefix, shape = "mark", eyeY = 17, waveY = null, glow = true, eye = "live", root, transform } = $props();
</script>

<g class={root} {transform}>
  {#if waveY !== null}
    <path class="{prefix}wave {prefix}wave--a" d="M16 {waveY} A18 18 0 0 1 48 {waveY}" />
    <path class="{prefix}wave {prefix}wave--b" d="M11 {waveY + 2} A24 24 0 0 1 53 {waveY + 2}" />
  {/if}
  <g class="{prefix}body">{#each MARKS[shape] as [el, attrs]}<svelte:element this={el} {...attrs} />{/each}</g>
  {#if glow}<circle class="{prefix}glow" cx="32" cy={eyeY} r="9" />{/if}
  {#if eye === "hollow"}<circle class="{prefix}eye-hollow" cx="32" cy={eyeY} r="6" />{:else}<circle class="{prefix}eye" cx="32" cy={eyeY} r="5" />{/if}
</g>
