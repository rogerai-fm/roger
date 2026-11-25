<script>
  import TocTuner from "#lib/ui/TocTuner.svelte";
  // Hardware - what runs where, from an ESP32 to a DGX Spark
  // Ported from web/src/research-hardware.html (2026-10-29); the copy is the legacy copy.
  import { siteHead, CSS_SHARED, CSS_ACCOUNT } from "#lib/head.js";
  import { content } from "#lib/content.js";
  import Footer from "#lib/ui/Footer.svelte";
  import InstallBox from "#lib/ui/InstallBox.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  const facts = content('research-hardware.html');
  const head = siteHead({
    title: 'Hardware - what runs where, from an ESP32 to a DGX Spark',
    desc: 'The Wave family and Roger Edge mapped onto real boards: ESP32, Arduino, Raspberry Pi Pico and Pi 5, Jetson, Mac Studio and DGX Spark. Which tier each board is a target for, what stage that tier is at, and what you can actually run today.',
    og: { type: 'website', title: 'RogerAI hardware - the board you own, and what it is meant to carry', desc: 'From a sensor microcontroller to a desktop datacenter: the model tier each class of board is a target for, with the honest stage of each one.', url: 'https://rogerai.fm/research-hardware.html', image: 'https://rogerai.fm/og-research.png' },
    css: [...CSS_SHARED, 'research.css', 'research-hardware.css'],
    scripts: [{ runtime: true }, { page: "research-hardware" }],
  });
</script>
<svelte:head>{@html head}</svelte:head>
<!-- The hardware page: the board a reader already owns, and the tier it is a target for.
     This page is deliberately the most constrained on the site. It puts photographs of real,
     recognisable equipment beside RogerAI's own model tiers, which is precisely the shape of
     page that drifts into implying a shipped product - so:
       - every tier and every stage string here is COPIED from research-models.html and
         locked to it by test/research-hardware.test.mjs, which fails if the two drift;
       - the words that would imply a deployment ("deployed", "in production", "running
         Wave") are banned inside the board cards by that same test;
       - the ONLY thing presented as available today is the one artifact that actually is:
         rogerai-fyi/DeepSeek-V4-Flash-MTP-GGUF, plus `roger share`, which ships.
     Photographs are other people's work under Creative Commons; the author, licence and
     source page travel with each one and are listed again in §5. 
     (the page's doc comment, from the legacy file; its sheets are the css: list above) -->
  <Rail head='RogerAI &middot; Hardware' rev='Rev. 2026.08 &middot; LABS' />

  <Nav variant='marketing' home='/' promo={false} />

  <main id="top">
    <section class="research-hero">
      <div class="wrap">
        <a class="wf-back mono" href="/research.html">&larr; RogerAI Labs</a>
        <span class="hero__eyebrow">
          <OnAir />
          <span class="hero__callsign">HARDWARE</span>
        </span>
        <!-- Plain first paragraph (SITE-SIMPLIFICATION-AUDIT-2026-11-13, fixes 6 and 9). The
             card stages below stay the status file's own words; the lead gives the three
             plain ones: live / in testing / planned. -->
        <h1>Which machine runs which Wave model.</h1>
        <p class="research-hero__lead">
          Each Wave size was designed for a class of machine. This page lists ten, from a
          sensor chip to a rack, and says how far along each one is. Wave Pico is live and
          free on the network; its weights are not downloadable, you use it through the
          network. Wave Nano&rsquo;s two fire models are on the network as specialized models.
          Wave Micro is training, and the sizes above it are not built yet, so read the stage on each
          card before the tier.
        </p>
      </div>
    </section>

    <!-- ---------- §1 THE SPECTRUM ---------- -->
    <TocTuner sections={[
      { id: "ladder", no: "§1", name: "THE SPECTRUM" },
      { id: "boards", no: "§2", name: "THE BOARDS" },
      { id: "experiments", no: "§3", name: "EXPERIMENTS" },
    ]} />

    <section class="section" id="ladder">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;1 / THE SPECTRUM</span>
          <h2>One band, eight stops, five orders of magnitude.</h2>
          <p>Eight classes of machine, from a tenth of a watt to ten kilowatts, and a wider
             span than that in parameters. Each stop is a class of machine first and a model
             second: the tier was specified to fit the box, which is why the boxes are named
             here at all.</p>
        </div>

        <figure class="hw-figure">
          <!-- The ladder in the isometric hand (2026-11-23): the SVG between the iso markers is written
               by computer-scientist/article/fig_ladder_gen.py (--svelte); the words are the spans over the
               stage; src/js/hardware.js keeps its hooks (.hw-stop[data-stop], data-tier, .is-lit, .is-linked,
               .hw-ladder.is-sweeping). The stage keeps the diagram scroll box: on a phone the band scrolls. -->
          <div class="scroll-box scroll-box--diagram">
          <div class="iso__stage hw-stage">
          <svg class="hw-ladder iso__svg" viewBox="-36.2 -13.8 342.6 131" role="img"
               aria-labelledby="hw-ladder-title hw-ladder-desc">
            <title id="hw-ladder-title">The Wave spectrum across eight classes of hardware</title>
            <desc id="hw-ladder-desc">A horizontal band with eight stops, each a class of
              machine and the model tier specified against it. From the left: Roger Edge, a
              few million parameters on an ESP32-class microcontroller at about a tenth of a
              watt. Wave Pico, 250 to 300 million on a Raspberry Pi with no GPU. Wave Nano,
              0.8 to 1.5 billion on a Jetson-class gateway. Wave Micro, 7 to 8 billion on an
              on-site server. Wave Giga, 27 to 35 billion on a plant datacenter box. Wave
              Tera, 80 to 120 billion on a single RTX 6000-class card. Wave Peta, 150 to 200
              billion across four such cards at around 2.4 kilowatts. And Wave Exa, over 280
              billion, on a multi-node rack drawing ten kilowatts or more. A tuning needle
              sweeps across the band from left to right.</desc>
            <!-- iso:begin -->
<path class="iso__halo" d="M-6.2 2 L300.4 90.5 L276.4 111.2 L-30.2 22.7 Z"/>
<g class="hw-ground">
<path class="iso__face iso__face--s1" d="M-22 19.1 L272.4 104.1 L272.4 106.7 L-22 21.7 Z"/>
<path class="iso__face iso__face--s2" d="M294.4 85 L272.4 104.1 L272.4 106.7 L294.4 87.6 Z"/>
<path class="iso__face" d="M0 0 L294.4 85 L272.4 104.1 L-22 19.1 Z"/>
<path class="iso__ln iso__ln--lo" d="M-16.5 18.3 L271 101.3"/>
<path class="iso__ln iso__ln--lo" d="M0.6 21.5 L-2.4 24.1"/>
<path class="iso__ln iso__ln--lo" d="M36.9 32 L33.9 34.6"/>
<path class="iso__ln iso__ln--lo" d="M73.3 42.5 L70.3 45.1"/>
<path class="iso__ln iso__ln--lo" d="M109.7 53 L106.7 55.6"/>
<path class="iso__ln iso__ln--lo" d="M146 63.5 L143 66.1"/>
<path class="iso__ln iso__ln--lo" d="M182.4 74 L179.4 76.6"/>
<path class="iso__ln iso__ln--lo" d="M218.8 84.5 L215.8 87.1"/>
<path class="iso__ln iso__ln--lo" d="M255.2 95 L252.2 97.6"/>
</g>
<g class="hw-stop iso__m hw-stop--edge" data-stop="edge" data-tier="Roger Edge" data-params="1e6" data-watts="0.1">
<path class="iso__halo iso__halo--part" d="M15.8 17.3 L14.4 18.2 L12.6 18.9 L10.4 19.4 L8.1 19.5 L5.7 19.4 L3.6 18.9 L1.7 18.2 L0.3 17.3 L-0.6 16.2 L-0.9 15 L-0.6 13.9 L0.3 12.8 L1.7 11.8 L3.6 11.1 L5.7 10.7 L8.1 10.5 L10.4 10.7 L12.6 11.1 L14.4 11.8 L15.8 12.8 L16.7 13.9 L17.1 15 L16.7 16.2 Z"/>
<path class="iso__face iso__face--s1" d="M0.4 15 L10.7 18 L10.7 18.7 L0.4 15.7 Z"/>
<path class="iso__face iso__face--s2" d="M15.7 13.7 L10.7 18 L10.7 18.7 L15.7 14.4 Z"/>
<path class="iso__face" d="M5.4 10.7 L15.7 13.7 L10.7 18 L0.4 15 Z"/>
<path class="iso__face iso__face--s1" d="M4.6 13.8 L9 15 L9 16 L4.6 14.8 Z"/>
<path class="iso__face iso__face--s2" d="M11.5 12.8 L9 15 L9 16 L11.5 13.9 Z"/>
<path class="iso__face" d="M7.1 11.6 L11.5 12.8 L9 15 L4.6 13.8 Z"/>
<path class="iso__ln iso__ln--faint" d="M11 14.3 L12.7 14.8"/>
<path class="iso__ln iso__ln--faint" d="M10.2 15 L11.9 15.5"/>
<path class="iso__ln iso__ln--faint" d="M9.5 15.6 L11.2 16.1"/>
<circle class="iso__lamp" cx="1.7" cy="15.7" r="1.3"/>
</g>
<g class="hw-stop iso__m hw-stop--pico" data-stop="pico" data-tier="Wave Pico" data-params="2.5e8" data-watts="5">
<path class="iso__halo iso__halo--part" d="M53.1 28 L51.5 29.1 L49.4 29.9 L47 30.4 L44.4 30.5 L41.8 30.4 L39.4 29.9 L37.4 29.1 L35.8 28 L34.8 26.8 L34.4 25.5 L34.8 24.2 L35.8 23 L37.4 22 L39.4 21.2 L41.8 20.7 L44.4 20.5 L47 20.7 L49.4 21.2 L51.5 22 L53.1 23 L54.1 24.2 L54.4 25.5 L54.1 26.8 Z"/>
<path class="iso__face iso__face--s1" d="M35.9 25.1 L48 28.6 L48 29.4 L35.9 25.9 Z"/>
<path class="iso__face iso__face--s2" d="M53 24.2 L48 28.6 L48 29.4 L53 25.1 Z"/>
<path class="iso__face" d="M40.9 20.7 L53 24.2 L48 28.6 L35.9 25.1 Z"/>
<path class="iso__face iso__face--s1" d="M47.5 22.9 L50.1 23.6 L50.1 25.7 L47.5 25 Z"/>
<path class="iso__face iso__face--s2" d="M51.6 22.3 L50.1 23.6 L50.1 25.7 L51.6 24.4 Z"/>
<path class="iso__face" d="M49 21.6 L51.6 22.3 L50.1 23.6 L47.5 22.9 Z"/>
<path class="iso__face iso__face--s1" d="M45 25.1 L47.6 25.8 L47.6 27.9 L45 27.1 Z"/>
<path class="iso__face iso__face--s2" d="M49.1 24.5 L47.6 25.8 L47.6 27.9 L49.1 26.6 Z"/>
<path class="iso__face" d="M46.5 23.8 L49.1 24.5 L47.6 25.8 L45 25.1 Z"/>
<path class="iso__face iso__face--s1" d="M39.1 23.6 L42.6 24.6 L42.6 25.3 L39.1 24.3 Z"/>
<path class="iso__face iso__face--s2" d="M44.6 22.9 L42.6 24.6 L42.6 25.3 L44.6 23.5 Z"/>
<path class="iso__face" d="M41.1 21.9 L44.6 22.9 L42.6 24.6 L39.1 23.6 Z"/>
<circle class="iso__lamp" cx="37.2" cy="25.9" r="1.3"/>
</g>
<g class="hw-stop iso__m hw-stop--nano" data-stop="nano" data-tier="Wave Nano" data-params="8e8" data-watts="15">
<path class="iso__halo iso__halo--part" d="M90.3 38.8 L88.6 39.9 L86.3 40.8 L83.6 41.3 L80.8 41.5 L78 41.3 L75.3 40.8 L73 39.9 L71.3 38.8 L70.2 37.4 L69.8 36 L70.2 34.6 L71.3 33.3 L73 32.1 L75.3 31.3 L78 30.7 L80.8 30.5 L83.6 30.7 L86.3 31.3 L88.6 32.1 L90.3 33.3 L91.4 34.6 L91.8 36 L91.4 37.4 Z"/>
<path class="iso__face iso__face--s1" d="M73.1 31.5 L83.5 34.5 L83.5 39.7 L73.1 36.7 Z"/>
<path class="iso__face iso__face--s2" d="M88.5 30.2 L83.5 34.5 L83.5 39.7 L88.5 35.4 Z"/>
<path class="iso__face" d="M78.1 27.2 L88.5 30.2 L83.5 34.5 L73.1 31.5 Z"/>
<path class="iso__face iso__face--s1" d="M74.5 29.6 L75.5 29.9 L75.5 31.6 L74.5 31.3 Z"/>
<path class="iso__face iso__face--s2" d="M79.5 26.4 L75.5 29.9 L75.5 31.6 L79.5 28.1 Z"/>
<path class="iso__face" d="M78.5 26.1 L79.5 26.4 L75.5 29.9 L74.5 29.6 Z"/>
<path class="iso__face iso__face--s1" d="M77.1 30.3 L78.1 30.6 L78.1 32.4 L77.1 32.1 Z"/>
<path class="iso__face iso__face--s2" d="M82.1 27.2 L78.1 30.6 L78.1 32.4 L82.1 28.9 Z"/>
<path class="iso__face" d="M81.1 26.9 L82.1 27.2 L78.1 30.6 L77.1 30.3 Z"/>
<path class="iso__face iso__face--s1" d="M79.7 31.1 L80.7 31.4 L80.7 33.1 L79.7 32.8 Z"/>
<path class="iso__face iso__face--s2" d="M84.7 27.9 L80.7 31.4 L80.7 33.1 L84.7 29.6 Z"/>
<path class="iso__face" d="M83.7 27.6 L84.7 27.9 L80.7 31.4 L79.7 31.1 Z"/>
<path class="iso__face iso__face--s1" d="M82.3 31.8 L83.3 32.1 L83.3 33.9 L82.3 33.6 Z"/>
<path class="iso__face iso__face--s2" d="M87.3 28.7 L83.3 32.1 L83.3 33.9 L87.3 30.4 Z"/>
<path class="iso__face" d="M86.3 28.4 L87.3 28.7 L83.3 32.1 L82.3 31.8 Z"/>
<circle class="iso__lamp" cx="82.2" cy="35.4" r="1.3"/>
</g>
<g class="hw-stop iso__m hw-stop--micro" data-stop="micro" data-tier="Wave Micro" data-params="7e9" data-watts="90">
<path class="iso__halo iso__halo--part" d="M127.6 49.5 L125.7 50.8 L123.2 51.7 L120.3 52.3 L117.2 52.5 L114.1 52.3 L111.2 51.7 L108.7 50.8 L106.8 49.5 L105.6 48.1 L105.2 46.5 L105.6 45 L106.8 43.5 L108.7 42.3 L111.2 41.3 L114.1 40.7 L117.2 40.5 L120.3 40.7 L123.2 41.3 L125.7 42.3 L127.6 43.5 L128.8 45 L129.2 46.5 L128.8 48.1 Z"/>
<path class="iso__face iso__face--s1" d="M110.3 35.3 L119 37.8 L119 49.9 L110.3 47.4 Z"/>
<path class="iso__face iso__face--s2" d="M124 33.5 L119 37.8 L119 49.9 L124 45.6 Z"/>
<path class="iso__face" d="M115.3 31 L124 33.5 L119 37.8 L110.3 35.3 Z"/>
<path class="iso__ln iso__ln--faint" d="M112.1 40.1 L117.3 41.6"/>
<path class="iso__ln iso__ln--faint" d="M112.1 38.4 L117.3 39.9"/>
<path class="iso__face iso__face--cav" d="M112.1 45.3 L117.3 46.8 L117.3 45.5 L112.1 44 Z"/>
<circle class="iso__lamp" cx="117.7" cy="38.7" r="1.3"/>
</g>
<g class="hw-stop iso__m hw-stop--giga" data-stop="giga" data-tier="Wave Giga" data-params="2.7e10" data-watts="240">
<path class="iso__halo iso__halo--part" d="M165.7 60.5 L163.4 62 L160.5 63.1 L157.2 63.8 L153.5 64 L149.9 63.8 L146.5 63.1 L143.6 62 L141.4 60.5 L140 58.8 L139.5 57 L140 55.2 L141.4 53.5 L143.6 52.1 L146.5 51 L149.9 50.3 L153.5 50 L157.2 50.3 L160.5 51 L163.4 52.1 L165.7 53.5 L167.1 55.2 L167.5 57 L167.1 58.8 Z"/>
<path class="iso__face iso__face--s1" d="M141.9 52.8 L159.2 57.8 L159.2 62.1 L141.9 57.1 Z"/>
<path class="iso__face iso__face--s2" d="M165.2 52.6 L159.2 57.8 L159.2 62.1 L165.2 56.9 Z"/>
<path class="iso__face" d="M147.9 47.6 L165.2 52.6 L159.2 57.8 L141.9 52.8 Z"/>
<path class="iso__ln iso__ln--faint" d="M143.6 56.8 L143.6 54.2"/>
<path class="iso__ln iso__ln--faint" d="M145.3 57.3 L145.3 54.7"/>
<path class="iso__ln iso__ln--faint" d="M147.1 57.8 L147.1 55.2"/>
<path class="iso__ln iso__ln--faint" d="M148.8 58.3 L148.8 55.7"/>
<path class="iso__ln iso__ln--faint" d="M150.5 58.8 L150.5 56.2"/>
<path class="iso__ln iso__ln--faint" d="M152.3 59.3 L152.3 56.7"/>
<circle class="iso__lamp" cx="157.5" cy="59.5" r="1.3"/>
</g>
<g class="hw-stop iso__m hw-stop--tera" data-stop="tera" data-tier="Wave Tera" data-params="8e10" data-watts="600">
<path class="iso__halo iso__halo--part" d="M202 71 L199.8 72.5 L196.9 73.6 L193.5 74.3 L189.9 74.5 L186.3 74.3 L182.9 73.6 L180 72.5 L177.8 71 L176.4 69.3 L175.9 67.5 L176.4 65.7 L177.8 64 L180 62.6 L182.9 61.5 L186.3 60.8 L189.9 60.5 L193.5 60.8 L196.9 61.5 L199.8 62.6 L202 64 L203.4 65.7 L203.9 67.5 L203.4 69.3 Z"/>
<path class="iso__face iso__face--s1" d="M180.9 65.3 L193 68.8 L193 71.9 L180.9 68.4 Z"/>
<path class="iso__face iso__face--s2" d="M199 63.6 L193 68.8 L193 71.9 L199 66.7 Z"/>
<path class="iso__face" d="M186.9 60.1 L199 63.6 L193 68.8 L180.9 65.3 Z"/>
<path class="iso__face iso__face--s1" d="M187.4 59.6 L188.4 59.9 L188.4 66.4 L187.4 66.1 Z"/>
<path class="iso__face iso__face--s2" d="M192.7 56.2 L188.4 59.9 L188.4 66.4 L192.7 62.7 Z"/>
<path class="iso__face" d="M191.6 55.9 L192.7 56.2 L188.4 59.9 L187.4 59.6 Z"/>
<path class="iso__ln iso__ln--faint" d="M187.9 64.9 L187.9 60.6"/>
<circle class="iso__lamp" cx="191.2" cy="69.6" r="1.3"/>
</g>
<g class="hw-stop iso__m hw-stop--peta" data-stop="peta" data-tier="Wave Peta" data-params="1.5e11" data-watts="2400">
<path class="iso__halo iso__halo--part" d="M240.1 82 L237.6 83.7 L234.3 85 L230.4 85.8 L226.3 86 L222.1 85.8 L218.3 85 L215 83.7 L212.4 82 L210.8 80.1 L210.3 78 L210.8 76 L212.4 74 L215 72.4 L218.3 71.1 L222.1 70.3 L226.3 70 L230.4 70.3 L234.3 71.1 L237.6 72.4 L240.1 74 L241.7 76 L242.3 78 L241.7 80.1 Z"/>
<path class="iso__face iso__face--s1" d="M214.1 75.5 L231.5 80.5 L231.5 83.6 L214.1 78.6 Z"/>
<path class="iso__face iso__face--s2" d="M238.5 74.5 L231.5 80.5 L231.5 83.6 L238.5 77.5 Z"/>
<path class="iso__face" d="M221.1 69.5 L238.5 74.5 L231.5 80.5 L214.1 75.5 Z"/>
<path class="iso__face iso__face--s1" d="M216.8 67.8 L217.8 68.1 L217.8 75.4 L216.8 75.1 Z"/>
<path class="iso__face iso__face--s2" d="M223.1 63.5 L217.8 68.1 L217.8 75.4 L223.1 70.9 Z"/>
<path class="iso__face" d="M222 63.2 L223.1 63.5 L217.8 68.1 L216.8 67.8 Z"/>
<path class="iso__face iso__face--s1" d="M220.7 68.9 L221.7 69.2 L221.7 76.6 L220.7 76.3 Z"/>
<path class="iso__face iso__face--s2" d="M227 64.7 L221.7 69.2 L221.7 76.6 L227 72 Z"/>
<path class="iso__face" d="M225.9 64.4 L227 64.7 L221.7 69.2 L220.7 68.9 Z"/>
<path class="iso__face iso__face--s1" d="M224.6 70 L225.6 70.3 L225.6 77.7 L224.6 77.4 Z"/>
<path class="iso__face iso__face--s2" d="M230.9 65.8 L225.6 70.3 L225.6 77.7 L230.9 73.1 Z"/>
<path class="iso__face" d="M229.8 65.5 L230.9 65.8 L225.6 70.3 L224.6 70 Z"/>
<path class="iso__face iso__face--s1" d="M228.5 71.1 L229.5 71.4 L229.5 78.8 L228.5 78.5 Z"/>
<path class="iso__face iso__face--s2" d="M234.8 66.9 L229.5 71.4 L229.5 78.8 L234.8 74.3 Z"/>
<path class="iso__face" d="M233.7 66.6 L234.8 66.9 L229.5 71.4 L228.5 71.1 Z"/>
<circle class="iso__lamp" cx="229.7" cy="81.3" r="1.3"/>
</g>
<g class="hw-stop iso__m hw-stop--exa" data-stop="exa" data-tier="Wave Exa" data-params="2.8e11" data-watts="10000">
<path class="iso__halo iso__halo--part" d="M276.5 92.5 L274 94.2 L270.7 95.5 L266.8 96.3 L262.7 96.5 L258.5 96.3 L254.7 95.5 L251.4 94.2 L248.8 92.5 L247.2 90.6 L246.7 88.5 L247.2 86.5 L248.8 84.5 L251.4 82.9 L254.7 81.6 L258.5 80.8 L262.7 80.5 L266.8 80.8 L270.7 81.6 L274 82.9 L276.5 84.5 L278.1 86.5 L278.7 88.5 L278.1 90.6 Z"/>
<path class="iso__face iso__face--s1" d="M253.1 60.4 L265.2 63.9 L265.2 93.3 L253.1 89.8 Z"/>
<path class="iso__face iso__face--s2" d="M272.2 57.8 L265.2 63.9 L265.2 93.3 L272.2 87.2 Z"/>
<path class="iso__face" d="M260.1 54.3 L272.2 57.8 L265.2 63.9 L253.1 60.4 Z"/>
<path class="iso__ln iso__ln--faint" d="M254.4 87.6 L263.9 90.3"/>
<path class="iso__ln iso__ln--faint" d="M254.4 85 L263.9 87.7"/>
<path class="iso__ln iso__ln--faint" d="M254.4 82.4 L263.9 85.1"/>
<path class="iso__ln iso__ln--faint" d="M254.4 79.8 L263.9 82.5"/>
<path class="iso__ln iso__ln--faint" d="M254.4 77.2 L263.9 79.9"/>
<path class="iso__ln iso__ln--faint" d="M254.4 74.6 L263.9 77.3"/>
<path class="iso__ln iso__ln--faint" d="M254.4 72 L263.9 74.7"/>
<path class="iso__ln iso__ln--faint" d="M254.4 69.4 L263.9 72.1"/>
<path class="iso__ln iso__ln--faint" d="M254.4 66.8 L263.9 69.5"/>
<path class="iso__ln iso__ln--faint" d="M254.4 64.2 L263.9 67"/>
<path class="iso__face iso__face--s2" d="M254.4 83.3 L263.9 86 L263.9 85 L254.4 82.2 Z"/>
<path class="iso__face iso__face--s2" d="M254.4 75.5 L263.9 78.2 L263.9 77.2 L254.4 74.4 Z"/>
<path class="iso__face iso__face--s2" d="M254.4 67.7 L263.9 70.4 L263.9 69.4 L254.4 66.6 Z"/>
<circle class="iso__lamp" cx="263.5" cy="65.1" r="1.3"/>
</g>
<g class="hw-needle" aria-hidden="true">
<path class="hw-needle__line" d="M0 0 L0 -7.8"/>
<circle class="hw-needle__head" cx="0" cy="-7.8" r="1.5"/>
</g>
            <!-- iso:end -->
          </svg>
          <div class="hw-words" aria-hidden="true">
              <span class="hw-word hw-word--top hw-word--edge"><b class="hw-stop__tier">Roger Edge</b><i class="hw-stop__size">~1&ndash;10M</i></span>
              <span class="hw-word hw-word--foot hw-word--edge"><span class="hw-stop__box">ESP32 class</span><i class="hw-stop__watt">~0.1 W</i></span>
              <span class="hw-word hw-word--top hw-word--pico"><b class="hw-stop__tier">Wave Pico</b><i class="hw-stop__size">250&ndash;300M</i></span>
              <span class="hw-word hw-word--foot hw-word--pico"><span class="hw-stop__box">Raspberry Pi</span><i class="hw-stop__watt">~5 W</i></span>
              <span class="hw-word hw-word--top hw-word--nano"><b class="hw-stop__tier">Wave Nano</b><i class="hw-stop__size">0.8&ndash;1.5B</i></span>
              <span class="hw-word hw-word--foot hw-word--nano"><span class="hw-stop__box">Jetson gateway</span><i class="hw-stop__watt">~15 W</i></span>
              <span class="hw-word hw-word--top hw-word--micro"><b class="hw-stop__tier">Wave Micro</b><i class="hw-stop__size">7&ndash;8B</i></span>
              <span class="hw-word hw-word--foot hw-word--micro"><span class="hw-stop__box">on-site server</span><i class="hw-stop__watt">~90 W</i></span>
              <span class="hw-word hw-word--top hw-word--giga"><b class="hw-stop__tier">Wave Giga</b><i class="hw-stop__size">27&ndash;35B</i></span>
              <span class="hw-word hw-word--foot hw-word--giga"><span class="hw-stop__box">plant datacenter</span><i class="hw-stop__watt">~240 W</i></span>
              <span class="hw-word hw-word--top hw-word--tera"><b class="hw-stop__tier">Wave Tera</b><i class="hw-stop__size">80&ndash;120B</i></span>
              <span class="hw-word hw-word--foot hw-word--tera"><span class="hw-stop__box">1&times; RTX 6000</span><i class="hw-stop__watt">~600 W</i></span>
              <span class="hw-word hw-word--top hw-word--peta"><b class="hw-stop__tier">Wave Peta</b><i class="hw-stop__size">150&ndash;200B</i></span>
              <span class="hw-word hw-word--foot hw-word--peta"><span class="hw-stop__box">4&times; RTX 6000</span><i class="hw-stop__watt">~2.4 kW</i></span>
              <span class="hw-word hw-word--top hw-word--exa"><b class="hw-stop__tier">Wave Exa</b><i class="hw-stop__size">&gt;280B</i></span>
              <span class="hw-word hw-word--foot hw-word--exa"><span class="hw-stop__box">rack, multi-node</span><i class="hw-stop__watt">10 kW+</i></span>
          </div>
          <div class="hw-feet mono" aria-hidden="true"><span>sensor</span><span>frontier datacenter</span></div>
          </div>
          </div>
          <p class="hw-figure__swipe mono" aria-hidden="true">swipe the band &rarr;</p>
          <figcaption class="fig">FIG. 1 - the spectrum is a hardware ladder before it is a
            model ladder. Hover a board below to find its stop.</figcaption>
        </figure>

        <!-- DEVICES A SITE ALREADY HAS (features/web/sensor_positioning.feature, PROPOSED
             2026-10-15, founder review): three devices, one install. Every name, stage and
             line is a cap token from src/data/capabilities.json; a planned device (the iPad
             today) is drawn as a dashed ghost by its data-cap-stage alone. No speed is printed:
             none is measured, and the device evidence contract on the industrial page says
             what a speed claim must name. Not .hw-card on purpose: the board cards are
             counted and held to the model catalogue. -->
        <div class="devices" id="devices-row">
          <h3>Devices a site already has.</h3>
          <!-- The adapter definition is the one sentence shared by every page that uses the word. -->
          <p class="devices__lead">Or start with a computer that is already there. One install,
            then a Wave model and an adapter for the job. An adapter is a small add-on file,
            trained for one job, that plugs into a Wave model without changing it. Each lamp
            reads the same status file as every other page.</p>
          <div class="install devices__install">
            <InstallBox id='installDevices'>curl <span class="tok">-fsSL</span> <span class="tok">https://rogerai.fm/install.sh</span> <span class="tok">| sh</span></InstallBox>
          </div>
          <ol class="devices__list">
            <li class="device tint-panel" data-cap="laptop" data-cap-stage="{facts.t.cap('laptop.stage')}">
              <svg class="device__art" viewBox="0 0 120 72" aria-hidden="true"><rect x="22" y="6" width="76" height="48" rx="3"/><rect x="28" y="12" width="64" height="36" rx="1"/><path d="M8 62h104l-6 6H14Z"/><path d="M22 54 8 62M98 54l14 8"/></svg>
              <b>{@html facts.cap('laptop.name')}</b>
              <span class="device__status">Wave model + adapter <i class="cap-lamp" aria-hidden="true"></i> <span class="cap-stage cap-stage--inline">{@html facts.cap('laptop.label')}</span></span>
              <p>{@html facts.cap('laptop.description')}</p>
            </li>
            <li class="device tint-panel" data-cap="industrial-pc" data-cap-stage="{facts.t.cap('industrial-pc.stage')}">
              <svg class="device__art" viewBox="0 0 120 72" aria-hidden="true"><rect x="16" y="14" width="88" height="46" rx="3"/><path d="M28 14V6M40 14V6M52 14V6M64 14V6M76 14V6M88 14V6"/><rect x="26" y="40" width="12" height="8" rx="1"/><rect x="44" y="40" width="12" height="8" rx="1"/><circle cx="90" cy="44" r="3"/></svg>
              <b>{@html facts.cap('industrial-pc.name')}</b>
              <span class="device__status">Wave model + adapter <i class="cap-lamp" aria-hidden="true"></i> <span class="cap-stage cap-stage--inline">{@html facts.cap('industrial-pc.label')}</span></span>
              <p>{@html facts.cap('industrial-pc.description')}</p>
            </li>
            <li class="device tint-panel" data-cap="ipad" data-cap-stage="{facts.t.cap('ipad.stage')}">
              <svg class="device__art" viewBox="0 0 120 72" aria-hidden="true"><rect x="30" y="4" width="60" height="64" rx="6"/><rect x="36" y="10" width="48" height="50" rx="1"/><circle cx="60" cy="64" r="1.6"/></svg>
              <b>{@html facts.cap('ipad.name')}</b>
              <span class="device__status">Wave model + adapter <i class="cap-lamp" aria-hidden="true"></i> <span class="cap-stage cap-stage--inline">{@html facts.cap('ipad.label')}</span></span>
              <p>{@html facts.cap('ipad.description')}</p>
            </li>
          </ol>
          <p class="fig devices__cap">Three devices, one install. A dashed device is planned, not
            here yet.</p>
        </div><!-- /devices -->
      </div>
    </section>

    <!-- ---------- §2 THE BOARDS ---------- -->
    <section class="section band band--inset" id="boards">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;2 / THE BOARDS</span>
          <h2>Ten machines, and the tier each one is a target for.</h2>
          <p>Every card names the tier the board is specified against and that tier&rsquo;s
             stage, copied from
             <a href="/research-models.html">the model catalogue</a> and checked against it
             by a test.</p>
        </div>

        <div class="hw-grid">

          <article class="hw-card" data-board="esp32" data-tier="Roger Edge"
                   data-stage="{facts.t.wave('edge.research-hardware.card')}" data-today="sensor" data-stop="edge">
            <figure class="hw-card__shot">
              <img src="assets/hardware/esp32.webp" srcset="assets/hardware/esp32-500.webp 500w, assets/hardware/esp32.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1010px) 45vw, 311px" width="1000" height="750" loading="lazy" decoding="async"
                   alt="Close macro of an ESP32-DOWD module soldered to a dark circuit board, surrounded by passive components.">
              <figcaption class="photo-credit">Photo <a href="https://commons.wikimedia.org/wiki/File:ESP32_on_Lolin32_Lite_clone_board_cropped.jpg" target="_blank" rel="noopener noreferrer">Ubahnverleih</a> &middot; CC0 &middot; edited</figcaption>
            </figure>
            <div class="hw-card__body">
              <h3>ESP32</h3>
              <p class="hw-card__spec mono">240 MHz &middot; 520 KB SRAM &middot; no MMU</p>
              <p>The sensor itself: a few hundred kilobytes and tenths of a watt. Roger Edge, a fixed-label classifier, is specified for this class.</p>
              <dl class="hw-card__meta">
                <div><dt>Target tier</dt><dd>Roger Edge</dd></div>
                <div><dt>Stage</dt><dd class="hw-stage">{@html facts.wave('edge.research-hardware.card')}</dd></div>
                <div><dt>A station?</dt><dd>No - too small to serve an API</dd></div>
              </dl>
            </div>
          </article>

          <article class="hw-card" data-board="arduino" data-tier="Roger Edge"
                   data-stage="{facts.t.wave('edge.research-hardware.card')}" data-today="sensor" data-stop="edge">
            <figure class="hw-card__shot">
              <img src="assets/hardware/arduino.webp" srcset="assets/hardware/arduino-500.webp 500w, assets/hardware/arduino.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1010px) 45vw, 311px" width="1000" height="750" loading="lazy" decoding="async"
                   alt="An Arduino Uno board seen from above, its blue PCB carrying the USB-B socket, barrel jack and header strips.">
              <figcaption class="photo-credit">Photo <a href="https://commons.wikimedia.org/wiki/File:Arduino_Uno_board.jpg" target="_blank" rel="noopener noreferrer">JrawX</a> &middot; CC0 &middot; edited</figcaption>
            </figure>
            <div class="hw-card__body">
              <h3>Arduino Uno</h3>
              <p class="hw-card__spec mono">16 MHz &middot; 2 KB SRAM &middot; 32 KB flash</p>
              <p>Below the edge tier, listed for that reason: the sensor and actuator side of a fleet, which the gateway above it talks to.</p>
              <dl class="hw-card__meta">
                <div><dt>Target tier</dt><dd>Roger Edge <span class="hw-note">(host side only)</span></dd></div>
                <div><dt>Stage</dt><dd class="hw-stage">{@html facts.wave('edge.research-hardware.card')}</dd></div>
                <div><dt>A station?</dt><dd>No - sensor and actuator side</dd></div>
              </dl>
            </div>
          </article>

          <article class="hw-card" data-board="pico" data-tier="Roger Edge"
                   data-stage="{facts.t.wave('edge.research-hardware.card')}" data-today="sensor" data-stop="edge">
            <figure class="hw-card__shot">
              <img src="assets/hardware/pico.webp" srcset="assets/hardware/pico-500.webp 500w, assets/hardware/pico.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1010px) 45vw, 311px" width="1000" height="750" loading="lazy" decoding="async"
                   alt="A Raspberry Pi Pico board at an oblique angle, its green PCB showing the RP2040 chip and castellated edge pads.">
              <figcaption class="photo-credit">Photo <a href="https://commons.wikimedia.org/wiki/File:Raspberry_Pi_Pico_oblique.jpg" target="_blank" rel="noopener noreferrer">Phiarc</a> &middot; CC BY-SA 4.0 &middot; edited</figcaption>
            </figure>
            <div class="hw-card__body">
              <h3>Raspberry Pi Pico</h3>
              <p class="hw-card__spec mono">RP2040 &middot; 133 MHz &middot; 264 KB SRAM</p>
              <p>A microcontroller, despite the name. It belongs at the Roger Edge stop, not with the Pi 5.</p>
              <dl class="hw-card__meta">
                <div><dt>Target tier</dt><dd>Roger Edge</dd></div>
                <div><dt>Stage</dt><dd class="hw-stage">{@html facts.wave('edge.research-hardware.card')}</dd></div>
                <div><dt>A station?</dt><dd>No - no OS, no server</dd></div>
              </dl>
            </div>
          </article>

          <article class="hw-card" data-board="pi5" data-tier="Wave Pico"
                   data-stage="{facts.t.wave('pico.research-hardware.card')}" data-today="station" data-stop="pico">
            <figure class="hw-card__shot">
              <img src="assets/hardware/pi5.webp" srcset="assets/hardware/pi5-500.webp 500w, assets/hardware/pi5.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1010px) 45vw, 311px" width="1000" height="750" loading="lazy" decoding="async"
                   alt="A Raspberry Pi 5 board from above, showing its dual HDMI sockets, USB ports and the Raspberry Pi logo silkscreen.">
              <figcaption class="photo-credit">Photo <a href="https://commons.wikimedia.org/wiki/File:Raspberry_Pi_5.jpg" target="_blank" rel="noopener noreferrer">SimonWaldherr</a> &middot; CC BY 4.0 &middot; edited</figcaption>
            </figure>
            <div class="hw-card__body">
              <h3>Raspberry Pi 5</h3>
              <p class="hw-card__spec mono">4&ndash;16 GB &middot; no GPU &middot; ~5 W</p>
              <p>The first stop with a real operating system, and the box Wave Pico is specified for: a few hundred million parameters, no accelerator, a target near 50 ms.</p>
              <dl class="hw-card__meta">
                <div><dt>Target tier</dt><dd>Wave Pico</dd></div>
                <div><dt>Stage</dt><dd class="hw-stage">{@html facts.wave('pico.research-hardware.card')}</dd></div>
                <div><dt>A station?</dt><dd>Yes - small models today</dd></div>
              </dl>
            </div>
          </article>

          <article class="hw-card" data-board="jetson" data-tier="Wave Nano"
                   data-stage="{facts.t.wave('nano.research-hardware.card')}" data-today="station" data-stop="nano">
            <figure class="hw-card__shot">
              <img src="assets/hardware/jetson.webp" srcset="assets/hardware/jetson-500.webp 500w, assets/hardware/jetson.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1010px) 45vw, 311px" width="1000" height="750" loading="lazy" decoding="async"
                   alt="An NVIDIA Jetson development kit carrier board with a large black finned heatsink over the module and USB ports along one edge.">
              <figcaption class="photo-credit">Photo <a href="https://commons.wikimedia.org/wiki/File:Nvidia_Jetson_Nano_2_Development_Kit.jpeg" target="_blank" rel="noopener noreferrer">Ubahnverleih</a> &middot; CC0 &middot; edited</figcaption>
            </figure>
            <div class="hw-card__body">
              <h3>NVIDIA Jetson</h3>
              <p class="hw-card__spec mono">8&ndash;64 GB unified &middot; CUDA &middot; 15&ndash;60 W</p>
              <p>The fleet gateway: accelerated inference in a cabinet&rsquo;s power budget. Wave Nano is specified here, to roll up many children.</p>
              <dl class="hw-card__meta">
                <div><dt>Target tier</dt><dd>Wave Nano</dd></div>
                <div><dt>Stage</dt><dd class="hw-stage">{@html facts.wave('nano.research-hardware.card')}</dd></div>
                <div><dt>A station?</dt><dd>Yes - a good one</dd></div>
              </dl>
            </div>
          </article>

          <article class="hw-card" data-board="macstudio" data-tier="Wave Micro"
                   data-stage="{facts.t.wave('micro.research-hardware.card')}" data-today="station" data-stop="micro">
            <figure class="hw-card__shot">
              <img src="assets/hardware/macstudio.webp" srcset="assets/hardware/macstudio-500.webp 500w, assets/hardware/macstudio.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1010px) 45vw, 311px" width="1000" height="750" loading="lazy" decoding="async"
                   alt="An Apple Mac Studio photographed from the front, a compact square aluminium box with a slot vent along its base.">
              <figcaption class="photo-credit">Photo <a href="https://commons.wikimedia.org/wiki/File:Mac_Studio_(2022)_front.jpg" target="_blank" rel="noopener noreferrer">Yasu</a> &middot; CC BY-SA 3.0 &middot; edited</figcaption>
            </figure>
            <div class="hw-card__body">
              <h3>Mac Studio</h3>
              <p class="hw-card__spec mono">32&ndash;512 GB unified &middot; ~90 W under load</p>
              <p>The site brain: unified memory for a 7&ndash;8B model at a desk&rsquo;s power draw. Wave Micro is specified for this class.</p>
              <dl class="hw-card__meta">
                <div><dt>Target tier</dt><dd>Wave Micro</dd></div>
                <div><dt>Stage</dt><dd class="hw-stage">{@html facts.wave('micro.research-hardware.card')}</dd></div>
                <div><dt>A station?</dt><dd>Yes - see <a href="/broadcasts-what-a-million-tokens-costs.html">the arithmetic</a></dd></div>
              </dl>
            </div>
          </article>

          <article class="hw-card" data-board="dgxspark" data-tier="Wave Giga"
                   data-stage="{facts.t.wave('giga.research-hardware.card')}" data-today="station" data-stop="giga">
            <figure class="hw-card__shot">
              <img src="assets/hardware/dgxspark.webp" srcset="assets/hardware/dgxspark-500.webp 500w, assets/hardware/dgxspark.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1010px) 45vw, 311px" width="1000" height="750" loading="lazy" decoding="async"
                   alt="An NVIDIA DGX Spark on a desk, a small gold-coloured box with a dense metal-foam lattice across its front panel.">
              <figcaption class="photo-credit">Photo <a href="https://commons.wikimedia.org/wiki/File:Nvidia_DGX_Spark_oblique_view_dllu.jpg" target="_blank" rel="noopener noreferrer">Daniel Lu</a> &middot; CC BY-SA 4.0 &middot; edited</figcaption>
            </figure>
            <div class="hw-card__body">
              <h3>NVIDIA DGX Spark</h3>
              <p class="hw-card__spec mono">128 GB coherent &middot; ~240 W &middot; desk-side</p>
              <p>A plant datacenter under a desk: memory for a 27&ndash;35B model on one box. Wave Giga is drawn here.</p>
              <dl class="hw-card__meta">
                <div><dt>Target tier</dt><dd>Wave Giga</dd></div>
                <div><dt>Stage</dt><dd class="hw-stage">{@html facts.wave('giga.research-hardware.card')}</dd></div>
                <div><dt>A station?</dt><dd>Yes - the largest here</dd></div>
              </dl>
            </div>
          </article>

          <article class="hw-card" data-board="rtx6000" data-tier="Wave Tera"
                   data-stage="{facts.t.wave('tera.research-hardware.card')}" data-today="station" data-stop="tera">
            <figure class="hw-card__shot">
              <img src="assets/hardware/rtx6000.webp" srcset="assets/hardware/rtx6000-500.webp 500w, assets/hardware/rtx6000.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1010px) 45vw, 311px" width="1000" height="750" loading="lazy" decoding="async"
                   alt="A professional NVIDIA workstation graphics card in side profile, its black shroud and gold-ringed blower fan lit against a dark background.">
              <figcaption class="photo-credit">Photo <a href="https://commons.wikimedia.org/wiki/File:4%C3%97_NVIDIA_RTX%E2%84%A2_6000_der_Ada-Generation_(%E6%9E%81%E5%AE%A2%E6%B9%BEGeekerwan)_02.png" target="_blank" rel="noopener noreferrer">Geekerwan</a> &middot; CC BY 3.0 &middot; edited</figcaption>
            </figure>
            <div class="hw-card__body">
              <h3>RTX PRO 6000 Blackwell</h3>
              <p class="hw-card__spec mono">96 GB GDDR7 &middot; 600 W Workstation / 300 W Max-Q</p>
              <p>The largest model on one card: 96 GB holds 80&ndash;120B at 4-bit. Wave Tera is drawn here, to correlate faults across plants.</p>
              <dl class="hw-card__meta">
                <div><dt>Target tier</dt><dd>Wave Tera</dd></div>
                <div><dt>Stage</dt><dd class="hw-stage">{@html facts.wave('tera.research-hardware.card')}</dd></div>
                <div><dt>A station?</dt><dd>Yes - a large one</dd></div>
              </dl>
            </div>
          </article>

          <article class="hw-card" data-board="rtx6000x4" data-tier="Wave Peta"
                   data-stage="{facts.t.wave('peta.research-hardware.card')}" data-today="station" data-stop="peta">
            <figure class="hw-card__shot">
              <img src="assets/hardware/rtx6000x4.webp" srcset="assets/hardware/rtx6000x4-500.webp 500w, assets/hardware/rtx6000x4.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1010px) 45vw, 311px" width="1000" height="750" loading="lazy" decoding="async"
                   alt="Four professional NVIDIA workstation graphics cards installed side by side in a chassis, with liquid-cooling tubing looping over the CPU block beside them.">
              <figcaption class="photo-credit">Photo <a href="https://commons.wikimedia.org/wiki/File:4%C3%97_NVIDIA_RTX%E2%84%A2_6000_der_Ada-Generation_(%E6%9E%81%E5%AE%A2%E6%B9%BEGeekerwan)_01.png" target="_blank" rel="noopener noreferrer">Geekerwan</a> &middot; CC BY 3.0 &middot; edited</figcaption>
            </figure>
            <div class="hw-card__body">
              <h3>Four of them, in one chassis</h3>
              <p class="hw-card__spec mono">384 GB across 4 cards &middot; ~2.4 kW, or ~1.2 kW Max-Q</p>
              <p>Past one card, the interconnect is the wall. Four cards carry 150&ndash;200B: the Wave Peta band, pruned and distilled from the flagship.</p>
              <dl class="hw-card__meta">
                <div><dt>Target tier</dt><dd>Wave Peta</dd></div>
                <div><dt>Stage</dt><dd class="hw-stage">{@html facts.wave('peta.research-hardware.card')}</dd></div>
                <div><dt>A station?</dt><dd>Yes - one node, four cards</dd></div>
              </dl>
            </div>
          </article>

          <article class="hw-card hw-card--wide" data-board="datacenter" data-tier="Wave Exa"
                   data-stage="{facts.t.wave('exa.research-hardware.card')}" data-today="station" data-stop="exa">
            <figure class="hw-card__shot">
              <img src="assets/hardware/datacenter.webp" srcset="assets/hardware/datacenter-500.webp 500w, assets/hardware/datacenter.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 899px) 45vw, (max-width: 1119px) 38vw, 407px" width="1000" height="750" loading="lazy" decoding="async"
                   alt="An aisle between two rows of white perforated server cabinets in a data centre, lit from a run of ceiling strip lights down the centre.">
              <figcaption class="photo-credit">Photo <a href="https://commons.wikimedia.org/wiki/File:Racks_Amravati_Data_Center.jpg" target="_blank" rel="noopener noreferrer">PiDatacenters</a> &middot; CC BY-SA 4.0 &middot; edited</figcaption>
            </figure>
            <div class="hw-card__body">
              <h3>A rack, and more than one node</h3>
              <p class="hw-card__spec mono">multi-node &middot; 10 kW+ &middot; frontier class</p>
              <p>The top of the ladder, the Wave Exa class: the teacher the family learns from. <b>DeepSeek-V4-Flash MTP</b>, 284B with 13B active per token, is the one row here you can download today.</p>
              <dl class="hw-card__meta">
                <div><dt>Target tier</dt><dd>Wave Exa</dd></div>
                <div><dt>Stage</dt><dd class="hw-stage">{@html facts.wave('exa.research-hardware.card')}</dd></div>
                <div><dt>Download</dt><dd><a href="https://huggingface.co/rogerai-fyi/DeepSeek-V4-Flash-MTP-GGUF" target="_blank" rel="noopener noreferrer">the GGUF build &rarr;</a></dd></div>
              </dl>
            </div>
          </article>

        </div>
        <div class="man-note hw-rent">
          <span class="man-note__tag mono">DON'T OWN ONE?</span>
          <p>Every machine above can be rented by the hour instead. An interruptible
             instance and one paste into its ONSTART box installs the client, brings the
             model up and goes on air:
             <code class="cmd-words"><span>curl</span> <span>-fsSL</span> <span>https://rogerai.fm/vast-onstart.sh</span> <span>|</span> <span>bash</span></code> - free by
             default, so nothing of yours has to sit on hardware you do not own. The
             <a href="/manual.html#rented">manual has the knobs</a>, and
             <a href="/broadcasts-what-a-million-tokens-costs.html">the arithmetic</a> says
             when renting pays for itself.</p>
        </div>

      </div>
    </section>

    <!-- ---------- §3 EXPERIMENTS ---------- -->
    <section class="section band band--inset" id="experiments">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;3 / EXPERIMENTS</span>
          <h2>The ladder is a result, not a plan.</h2>
          <p>Every stop above was placed by trying to break it: pushing a frontier model
             down onto hardware that should not hold it, and pushing a tiny one up until it
             is useful. These are the runs that decide where the lines actually fall, with
             the stage each is at and what has come out so far.</p>
        </div>

        <div class="hw-exps">
          <article class="hw-exp" data-exp="kimi" data-stage="Current gate &middot; 16K saliency calibration">
            <span class="hw-exp__tag mono">FROM THE TOP</span>
            <h3>Can 2.8 trillion parameters fit on one workstation?</h3>
            <p>Kimi-K3 is 2.8T total with 104B active per token. The run tests whether
               pruning, low-precision execution and conversion together can put a
               frontier-scale model on a single high-memory box - which, if it works, moves
               the right-hand end of this ladder several stops to the left.</p>
            <p class="hw-exp__stage mono">Current gate &middot; 16K saliency calibration</p>
          </article>

          <article class="hw-exp" data-exp="mtp" data-stage="Field note">
            <span class="hw-exp__tag mono">SHIPPED</span>
            <h3>Why did a 284B model decode at 2 tokens a second?</h3>
            <p>The DeepSeek-V4-Flash work started as a hardware-fit question and turned into
               a runtime one: expert precision mattered more than any flag, and the format
               was missing multi-token prediction entirely, so we built MTP into the
               converter. The result is the one artifact on this page you can download.</p>
            <p class="hw-exp__stage mono">Field note &middot; <a href="/broadcasts-deepseek-mtp-gguf.html">the whole build, including what failed</a></p>
          </article>

          <article class="hw-exp" data-exp="concurrency"
                   data-rig="4x RTX PRO 4500" data-model="gpt-oss-120b"
                   data-single="160" data-batched="310" data-users="8">
            <span class="hw-exp__tag mono">MEASURED</span>
            <h3>How many people does one card actually serve?</h3>
            <p>A stop on this ladder is only real if the box serves more than one person at
               a time, so throughput is measured in aggregate under load rather than
               single-stream. This is the one machine on this page we measure on rather than
               specify against - our own bench:</p>
            <table class="data-table hw-bench">
              <caption class="mono">4&times; RTX PRO 4500 &middot; gpt-oss-120b &middot; users swept 1 &rarr; 8</caption>
              <tbody>
                <tr><th>Aggregate</th><td class="mono">~160 tok/s</td><td class="mono">&rarr;</td><td class="mono hw-bench__win">~310 tok/s</td></tr>
                <tr><th>Per user</th><td class="mono">~171 tok/s</td><td class="mono">&rarr;</td><td class="mono">more than halved</td></tr>
              </tbody>
            </table>
            <p>Nearly double the work out of the same hardware, and the reason the
               single-stream number quoted everywhere badly understates a box.</p>
            <p class="hw-exp__stage mono">Published &middot; <a href="/broadcasts-one-gpu-many-users.html">the sweep</a> and <a href="/broadcasts-what-a-million-tokens-costs.html">what it costs to serve</a></p>
          </article>

          <article class="hw-exp" data-exp="pico" data-stage="{facts.t.wave('pico.research-hardware.exp')}">
            <span class="hw-exp__tag mono">FROM THE BOTTOM</span>
            <h3>How small can a model get and still be worth asking?</h3>
            <p>The other direction. Wave Pico targets 250&ndash;300M on a Raspberry Pi with
               no accelerator and a response near 50 ms - small enough to sit where the
               network stops, large enough to answer rather than classify. The 293M release
               is live on the network as wave-pico-293m; the weights are not downloadable, you
               use it through the network.</p>
            <p class="hw-exp__stage mono">{@html facts.wave('pico.research-hardware.exp')}</p>
          </article>
        </div>

        <div class="man-note">
          <span class="man-note__tag mono">HOW WE PUBLISH</span>
          <p>Measurements name the hardware, the artifact, the runtime and the settings, and
             negative results stay up: the 2 tok/s mystery above is on the site because the
             answer was not where anyone looked first, and the file that finally worked
             turned out to be one we already had. Everything lands in
             <a href="/broadcasts.html">Open Air Waves</a> as it happens.</p>
        </div>
      </div>
    </section>

  </main>

  <Footer anchors='/' />
