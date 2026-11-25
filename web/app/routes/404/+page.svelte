<script>
  import PingMark from "#lib/ui/PingMark.svelte";
  // ...static - RogerAI
  // Ported from web/src/404.html (2026-10-29); the copy is the legacy copy.
  import { siteHead, CSS_SHARED, CSS_ACCOUNT } from "#lib/head.js";
  import Footer from "#lib/ui/Footer.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  const head = siteHead({
    title: '...static - RogerAI',
    desc: '...static. the signal dropped. That channel is not on the band. Tune back in to RogerAI.',
    robots: "noindex",
    canonical: 'https://rogerai.fm/404.html',
    css: [...CSS_SHARED, 'notfound.css'],
    scripts: [{ runtime: true }, { page: "404" }],
  });
</script>
<svelte:head>{@html head}</svelte:head>
  <!-- running head: this page is one bound document -->
  <Rail head='RogerAI · Operating Manual' rev='Rev. 2026.06 · ((•))' orange />

  <!-- ============================ NAV ============================ -->
  <Nav variant='marketing' home='/' />

  <!-- ============================ LOST SIGNAL ============================ -->
  <main id="top">
  <section class="lost">
    <div class="wrap">
    <div class="lost__inner tint-panel">
      <span class="fig" aria-hidden="true">FIG. 404 - NO CARRIER</span>

      <!-- Ping in its dropped / "...static" pose: arms sag, the on-air eye goes
           hollow (a dim ring, no live-red fill). Holds still - the signal is
           gone, so nothing transmits. -->
      <span class="lost__ping" aria-hidden="true">
        <svg class="ping ping--dropped" viewBox="0 0 64 72" fill="currentColor" focusable="false">
          <PingMark prefix="ping__" shape="dropped" eyeY={22} glow={false} eye="hollow" />
        </svg>
      </span>

      <h1 class="lost__title">...static. the signal dropped.</h1>
      <p class="lost__sub">
        That channel is not on the band. Either it went off air, or this path
        never broadcast in the first place.
      </p>

      <p class="lost__path mono" aria-label="The page you requested">
        <span class="lost__path-label" aria-hidden="true">no carrier at</span>
        <span class="lost__path-code" id="lostPath">/</span>
      </p>

      <div class="research-actions research-actions--lead lost__actions">
        <a class="research-button" href="/">
          <OnAir />
          tune back in
        </a>
        <a class="research-button" href="/#market">read the band →</a>
      </div>
    </div>
    </div>
  </section>
  </main>

  <!-- ============================ FOOTER ============================ -->
  <Footer anchors='/' />
