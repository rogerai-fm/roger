<script>
  // The page's table of contents: the shared tuner (components.css .toc-tuner, src/js/tuner.js). One station
  // per section, each { id, no, name }: the anchor, the § number and the short name the readout shows. The
  // section's own heading may say more; the tuner keeps the short form.
  /** @type {{ sections: { id: string, no: string, name: string }[] }} */
  let { sections } = $props();
</script>

<nav class="toc-tuner" data-tuner aria-label="Sections">
  <div class="wrap toc-tuner__band">
    <div class="toc-tuner__readout" aria-hidden="true"></div>
    <ol class="toc-tuner__scale">
      {#each sections as s}
        <li><a class="toc-tuner__st" href="#{s.id}"><b>{s.no}</b><span class="toc-tuner__name">{s.name}</span></a></li>
      {/each}
    </ol>
    <span class="toc-tuner__needle" aria-hidden="true"></span>
  </div>
</nav>
