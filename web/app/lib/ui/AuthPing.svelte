<script>
  // Ping on the sign-in cards (login, device approval, session link; 2026-11-25). The mascot's shape
  // (PingMark) at the top of the plate: waves and a lit eye while the page is listening or waiting on
  // the broker, a hollow eye with no waves when nothing more will happen (denied, signed out, a dead
  // link). Decoration only: the card's words carry the state. Motion is base.css .ping-mark (bob,
  // pulse, blink, waves), off under reduced motion.
  import PingMark from "#lib/ui/PingMark.svelte";
  /** @type {{ on?: boolean }} */
  let { on = true } = $props();
</script>

<svg class="auth-ping" viewBox="0 -8 64 80" fill="currentColor" aria-hidden="true" focusable="false">
  <PingMark root="ping-mark" prefix="ping-mark__" waveY={on ? 8 : null} glow={on} eye={on ? "live" : "hollow"} />
</svg>
