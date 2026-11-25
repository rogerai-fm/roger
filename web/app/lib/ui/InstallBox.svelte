<script>
  // The install pill (legacy: _partials/install-box.html): a command with the copy tick
  // site.js wires. The command is the children markup, one <span class="tok"> per token,
  // so a page's source reads as HTML and a facts call can sit inside it.
  // The button's accessible name is the command itself (its visible text; the $ prompt and the
  // copy glyph are aria-hidden), so the name matches what a sighted reader sees (WCAG 2.5.3,
  // axe label-content-name-mismatch, 2026-11-14). What pressing it does is the title.
  let { id, size = "", oslock = "", label = "", children } = $props();
</script>
<button class="install__box{size === 'lg' ? ' install__box--lg' : ''}" {id} type="button" data-os-lock={oslock || undefined}
                title={label || "Copy to clipboard"}>
          <span class="install__prompt" aria-hidden="true">$</span>
          <code class="install__code">{@render children()}</code>
          <span class="install__copy" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round">
              <rect x="9" y="9" width="11" height="11" rx="2"></rect>
              <path d="M5 15V5a2 2 0 0 1 2-2h10"></path>
            </svg>
          </span>
        </button>
