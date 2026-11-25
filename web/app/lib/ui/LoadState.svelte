<script>
  // The account screens' loading line (2026-11-24, UX review): shown in place of the card while the
  // screen's data loads, so a signed-in page never opens on an empty panel. A polite live region,
  // so a screen reader hears it once. Each screen keeps its own fault line (with a Try again button).
  /** @type {{ status: string, what?: string }} */
  let { status, what = "this page" } = $props();
</script>

{#if status === "loading"}<p class="acct-state tint-panel acct-loading" role="status">Loading {what}&hellip;</p>{/if}
