<script>
  // Voices - RogerAI voice directory
  // Ported from web/src/voices.html (2026-10-29); the copy is the legacy copy.
  import { siteHead, CSS_SHARED } from "#lib/head.js";
  import Footer from "#lib/ui/Footer.svelte";
  import Island from "#lib/ui/Island.svelte";
  import Roster from "#lib/roster/Roster.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  const head = siteHead({
    title: 'Voices - RogerAI voice directory',
    desc: 'The live RogerAI voice directory: on-air text-to-speech voices, each one an operator\'s station. Read the roster, pick a voice, pay by the character on the CLI or in the app. Real data only - no station, no voice.',
    canonical: 'https://rogerai.fm/voices.html',
    css: [...CSS_SHARED, 'tuner.css', 'voices.css'],
    scripts: [{ runtime: true }],
  });
</script>
<svelte:head>{@html head}</svelte:head>
  <!-- ===== running head: this page is one bound document ===== -->
  <Rail head='RogerAI · Voice Directory' rev='Rev. 2026.06 · ((•))' orange />

  <!-- ============================ NAV ============================ -->
  <Nav variant='marketing' home='/' current='voices' />

  <main id="top">

  <!-- ============================ HERO ============================ -->
  <section class="bands-hero">
    <div class="wrap bands-hero__inner">

      <span class="hero__eyebrow" data-reveal>
        <OnAir />
        <span class="hero__callsign">THE VOICES</span>
        <span class="hero__sep" aria-hidden="true">·</span>
        the on-air roster
      </span>

      <h1 class="bands-hero__title" data-reveal data-reveal-delay="60">
        Pick a voice.
      </h1>

      <p class="bands-hero__sub" data-reveal data-reveal-delay="120">
        Every voice is an operator's <b>station on air</b> - a home text-to-speech
        rig (Kokoro, Piper, and friends) put on the band. Read the roster, then
        speak with any of them by the character on the CLI or in the app. This
        directory is <b>live, real data only</b> - we list every voice on air,
        cheapest first; operators stay a <b>@callsign</b>, nothing more. No node
        address ever leaves the broker.
      </p>

    </div>
  </section>

  <!-- ============================ DIRECTORY ============================ -->
  <!-- INK PANEL: the roster and how to speak, one moment (components.css .tone-zone). -->
  <div class="tone-zone" data-tone="ink">
  <section class="section bands-dir" id="directory">
    <div class="wrap">
      <div class="section__head section__head--left" data-reveal>
        <span class="sectionno">§1 / THE ROSTER · LIVE</span>
        <h2>Read the roster.</h2>
        <p>The live voice directory, mirroring <code class="inline">roger voices</code> and
          the broker's <code class="inline">/voices</code> picker. Real data only - if no
          voice is on air, the band is quiet. Cheapest first. Search it, sort it, filter it.
          Auto-refreshes every 30s.</p>
        <p class="dir-sibling"><span>Also on the network:</span>
          <a href="/models.html">Models <span aria-hidden="true">&rarr;</span></a>
          <a href="/search.html">Search <span aria-hidden="true">&rarr;</span></a>
          <a href="/tower.html">The Tower <span aria-hidden="true">&rarr;</span></a></p>
      </div>

      <!-- ---- the tuner, the live roster and how to speak: one island (lib/roster/Roster.svelte) ---- -->
      <Island name="voices" component={Roster} props={{ kind: "voices" }} />
    </div>
  </section>
  </div><!-- /tone-zone -->

  </main>

  <!-- ============================ FOOTER ============================ -->
  <Footer anchors='/' />
