<script>
  // Search - RogerAI search directory (2026-11-26). The voices page's sibling: the same live roster
  // (lib/roster/Roster.svelte) over the broker's public /search/stations, with the Models page's tuner.
  // Facts: the manual's §S (roger 7.2.0, POST /v1/search, a price per 1,000 queries) and the broker's
  // search.go (a request names one station in "model"; with none, the broker picks one).
  import { siteHead, CSS_SHARED } from "#lib/head.js";
  import Footer from "#lib/ui/Footer.svelte";
  import Island from "#lib/ui/Island.svelte";
  import Roster from "#lib/roster/Roster.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  const head = siteHead({
    title: 'Search - RogerAI search directory',
    desc: 'The live RogerAI search directory: search stations on air, each one an operator\'s own search. Read the roster, pick a station, pay by the query. Real data only.',
    canonical: 'https://rogerai.fm/search.html',
    css: [...CSS_SHARED, 'tuner.css', 'voices.css'],
    scripts: [{ runtime: true }],
  });
</script>
<svelte:head>{@html head}</svelte:head>
  <Rail head='RogerAI · Search Directory' rev='Rev. 2026.11 · ((•))' orange />
  <Nav variant='marketing' home='/' current='search' />

  <main id="top">

  <section class="bands-hero">
    <div class="wrap bands-hero__inner">
      <span class="hero__eyebrow" data-reveal>
        <OnAir />
        <span class="hero__callsign">SEARCH</span>
        <span class="hero__sep" aria-hidden="true">·</span>
        the on-air roster
      </span>
      <h1 class="bands-hero__title" data-reveal data-reveal-delay="60">
        Pick a search.
      </h1>
      <p class="bands-hero__sub" data-reveal data-reveal-delay="120">
        Every search here is an operator's <b>station on air</b>: a search engine they hold
        a key for, one they run themselves, or an index of their own, such as a shop's
        products or a set of docs. Your app sends one query and pays that station's price
        for one query. This directory is <b>live, real data only</b>, and lists operators'
        stations; RogerAI's own is left out. Operators stay a
        <b>@callsign</b>, nothing more, and no station address ever leaves the broker.
      </p>
    </div>
  </section>

  <!-- INK PANEL: the roster and how to search, one moment (components.css .tone-zone). -->
  <div class="tone-zone" data-tone="ink">
  <section class="section bands-dir" id="directory">
    <div class="wrap">
      <div class="section__head section__head--left" data-reveal>
        <span class="sectionno">§1 / THE ROSTER · LIVE</span>
        <h2>Read the roster.</h2>
        <p>The live search directory, read from the broker's <code class="inline">/search/stations</code>,
          the list <code class="inline">roger search --kind search</code> prints, less RogerAI's own
          station. Real data only: if no
          operator's station is listed, the band is quiet. Cheapest first. Search it, sort it, filter it.
          Auto-refreshes every 30s.</p>
        <p class="dir-sibling"><span>Also on the network:</span>
          <a href="/models.html">Models <span aria-hidden="true">&rarr;</span></a>
          <a href="/voices.html">Voices <span aria-hidden="true">&rarr;</span></a>
          <a href="/manual.html#sSearch">Share a search of your own <span aria-hidden="true">&rarr;</span></a></p>
      </div>

      <Island name="search" component={Roster} props={{ kind: "search" }} />

      <p class="voice-tune__note mono roster-key" data-reveal>The Bearer is a grant key from <a href="/keys.html">API keys</a>.
        A request can set <code class="inline">max_price_per_1k</code>, in dollars per 1,000 queries, as the most it will pay.</p>
    </div>
  </section>
  </div><!-- /tone-zone -->

  </main>

  <Footer anchors='/' />
