<script>
  // Putting a GPU on air, in three steps (2026-11-24, from the UX review of the signed-in pages): the
  // empty stations screen and the dashboard's earn panel said the same three commands as a bare block.
  // A small Ping (PingMark, the mascot's shape) beside a numbered step list, each command a copy button
  // (components.css .copy-code, the shared copy tick). Words and commands are props so a screen can say
  // its own lead.
  import PingMark from "#lib/ui/PingMark.svelte";
  /** @type {{ steps?: { name: string, cmd: string }[] }} */
  let {
    steps = [
      { name: "Install roger", cmd: "curl -fsSL https://rogerai.fm/install.sh | sh" },
      { name: "Link this account", cmd: "roger login" },
      { name: "Go on air", cmd: "roger share" },
    ],
  } = $props();
</script>

<div class="onair-steps">
  <svg class="onair-steps__ping" viewBox="0 -8 64 80" fill="currentColor" aria-hidden="true" focusable="false">
    <PingMark root="ping-mark" prefix="ping-mark__" waveY={8} />
  </svg>
  <ol class="steps onair-steps__list">
    {#each steps as s}
      <li><b>{s.name}</b><button class="copy-code copy-code--block" type="button" data-copy-target aria-label="Copy {s.cmd}"><code>{s.cmd}</code></button></li>
    {/each}
  </ol>
</div>
