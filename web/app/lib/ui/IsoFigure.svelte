<script>
  // The shell every drawn figure shares (2026-11-24): the plate, its caption id, the scrubbed step list and
  // the FIG. caption, so a figure component holds only its picture. The picture is the children snippet
  // (the iso stage, the generated SVG between its <!-- iso:begin --> / <!-- iso:end --> markers, the HTML
  // words over it); anything between the picture and the list is the `between` snippet.
  //
  // The step list is the one scrub.js plays (data-scrub, data-scrub-auto); it marks this figure
  // data-step="n", which the figure's own sheet reads. A step is { name, note }: numbered steps render
  // <b><i>n</i>name</b><span>note</span>, unnumbered ones <b>name</b><span>note</span>. No steps, no list.
  // Any other attribute (data-report, data-keyed ...) passes through to the <figure>.
  /**
   * @type {{
   *   lead: string, fig: number|string, caption: string, extra?: string,
   *   steps?: { name: string, note: string }[], stepsClass?: string, stepsLabel?: string,
   *   numbered?: boolean, start?: number, draw?: boolean,
   *   children: import('svelte').Snippet, between?: import('svelte').Snippet, [attr: string]: unknown
   * }}
   */
  let { lead, fig, caption, extra = "", steps = [], stepsClass = "", stepsLabel = "", numbered = true, start,
        draw = true, children, between, ...attrs } = $props();
  const id = $derived(`${lead}-cap-${fig}`);
</script>

<figure class="figure figure--plate {lead}{extra ? ` ${extra}` : ''}" {...attrs} data-draw={draw ? "" : undefined} role="group" aria-labelledby={id}>
  {@render children()}
  {@render between?.()}
  {#if steps.length}
    <ol class={stepsClass} data-scrub data-scrub-auto data-scrub-start={start} aria-label={stepsLabel}>
      {#each steps as s, i}
        <li><b>{#if numbered}<i>{i + 1}</i>{/if}{s.name}</b><span>{s.note}</span></li>
      {/each}
    </ol>
  {/if}
  <figcaption class="fig" id={id}>FIG. {fig} - {caption}</figcaption>
</figure>
