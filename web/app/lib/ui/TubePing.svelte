<script>
  import PingMark from "#lib/ui/PingMark.svelte";
</script>

<!-- The WEB mascot: high-fidelity vector Ping (legacy: _partials/tube-ping.html). The ASCII
     Tube Ping in internal/tui/tube_ping.go is the TERMINAL form of the same character; keep
     each in its own medium. Geometry is mascot.svg (app/lib/ping.js MARKS.mark, drawn by PingMark.svelte), inlined so it inherits currentColor; ids
     are classes because the homepage renders it twice; the animations live in base.css. -->
<figure class="tube-ping" role="img" aria-label="Ping, the RogerAI on-air mascot">
  <svg class="tube-ping__mark" viewBox="0 -8 64 80" fill="currentColor" aria-hidden="true" focusable="false">
    <PingMark root="ping-mark" prefix="ping-mark__" waveY={8} />
  </svg>
  <figcaption><b>PING</b><span>RogerAI Labs · right-sized intelligence</span></figcaption>
</figure>
