<script>
  import TocTuner from "#lib/ui/TocTuner.svelte";
  // Wildfire mitigation (2026-11-24, founder request): the industrial page's wildfire card opens here, the
  // way its oil and gas card opens research-oil-gas.html. The founder named the places (power lines,
  // hillsides, unattended land, homes, a portable unit, a factory catching smoke, fire or gas) and asked for
  // research on the ones missed; the brief is the lab's wildfire use-cases brief (2026-11-24)
  // (sources and dates there; most outside numbers are vendor or news figures, so none is printed here).
  // Show, don't tell: each section is a heading, a line, the four-part contract, and a link.
  //
  // Claims: every number on this page is a model-cards.json token, read through the Nano page's own facts
  // (as manual B6 does), each beside how it was tested. Wording: "early warning", never "detection"; the
  // model warns and explains and never actuates; it is never a fire alarm, a protection system or part of
  // one. FireDefense Systems is named with its agreement (the industrial page's §1 says so).
  import { siteHead, CSS_SHARED } from "#lib/head.js";
  import { content } from "#lib/content.js";
  import Footer from "#lib/ui/Footer.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  import FireScene from "#lib/ui/FireScene.svelte";
  import WaveFirewatch from "#lib/ui/WaveFirewatch.svelte";
  const nano = content('research-wave-nano.html');
  const head = siteHead({
    title: 'Wildfire mitigation: an early-warning watch on site - RogerAI',
    desc: 'A wildfire early-warning watch that smells first and looks second: smoke chemistry, heat and wind against a baseline, a camera frame to confirm, and a warning with its reason. On its own battery, around power lines, homes, hillsides, unattended land, a portable unit and factories. Advisory, never a fire alarm.',
    og: { type: 'website', title: 'RogerAI - wildfire early warning, on site', desc: 'Power lines, homes, hillsides, unattended land, a portable unit, factories. Smell first, look second, tell a person why.', url: 'https://rogerai.fm/research-wildfire.html', image: 'https://rogerai.fm/og-research.png' },
    css: [...CSS_SHARED, 'research.css', 'wave-model.css', 'wildfire.css'],
    scripts: [{ runtime: true }, { page: "research-wildfire" }],
  });
  const TUNING = "/research-custom-tuning.html";
</script>
<svelte:head>{@html head}</svelte:head>
  <Rail head='RogerAI &middot; Wildfire mitigation' rev='Rev. 2026.11 &middot; LABS' />

  <Nav variant='marketing' home='/' promo={false} />

  <main id="top">
    <section class="research-hero">
      <div class="wrap">
        <a class="wf-back mono" href="/research-industry.html">&larr; Industrial deployment</a>
        <span class="hero__eyebrow">
          <OnAir />
          <span class="hero__callsign">WILDFIRE MITIGATION</span>
        </span>
        <h1>Wildfire early warning, on site.</h1>
        <p class="research-hero__lead">
          A watch that smells first and looks second. It reads smoke chemistry, heat and wind
          against its own baseline, confirms with a camera frame, and tells a person why. It runs
          on its own battery, because in a wildfire the towers and the power tend to go before the
          fire arrives.
        </p>
        <p class="research-hero__lead">
          Our Wave models run on the watch itself, with no network needed. It warns and explains;
          it never opens a valve.
          <a href={TUNING}>Tuned on your own site &rarr;</a>
        </p>
        <FireScene fig={1} />
      </div>
    </section>

    <TocTuner sections={[
      { id: "reads", no: "§1", name: "HOW IT READS" },
      { id: "lines", no: "§2", name: "POWER LINES" },
      { id: "homes", no: "§3", name: "HOMES" },
      { id: "land", no: "§4", name: "OPEN LAND" },
      { id: "portable", no: "§5", name: "PORTABLE" },
      { id: "factories", no: "§6", name: "FACTORIES" },
      { id: "more", no: "§7", name: "MORE PLACES" },
      { id: "measured", no: "§8", name: "MEASURED" },
      { id: "wildfire", no: "§9", name: "WILDFIRE" },
    ]} />

    <section class="section" id="reads">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;1 / HOW IT READS</span>
          <h2>Smell first, look second</h2>
          <p>A wildfire announces itself chemically before it does optically. One trace leaving its
            baseline is weather; three leaving together is combustion. A camera frame then confirms
            the smoke, and the wind says which way it is going.</p>
        </div>
        <div class="device-contract">
          <p><b>What goes in.</b> Fine particulate, carbon monoxide and volatile organics, heat, wind
            speed and direction, and a camera frame.</p>
          <p><b>What comes out.</b> The site&rsquo;s state (quiet, watch, elevated, threat, or an
            instrument fault), the reading that decided it, how sure it is, and anything that does not
            add up.</p>
          <p><b>Who acts.</b> A person: the owner, the crew, or whoever holds the land. The model
            marks; it never actuates.</p>
          <p><b>Where it runs.</b> On the mast, on its own battery, with no network required. A link,
            when there is one, carries the warning out.</p>
        </div>
      </div>
    </section>

    <section class="section band band--inset" id="lines">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;2 / POWER LINES</span>
          <h2>Beside the line, on its own battery</h2>
          <p>On high-wind days a line may be shut off, and a camera needs power and a link. A watch at a
            span or a substation keeps reading through the outage, and it reports smoke, not an
            electrical fault.</p>
        </div>
        <div class="device-contract">
          <p><b>What goes in.</b> Chemistry, heat and wind at the span or the substation fence, a camera
            frame along the corridor, and your line-sensor events if you have them.</p>
          <p><b>What comes out.</b> A warning with the reading and the wind direction, for the control
            room or the patrol.</p>
          <p><b>Who acts.</b> The utility&rsquo;s operators. Nothing on the line opens or closes on
            the model&rsquo;s word.</p>
          <p><b>What we ask you to send.</b> A season of readings from a few spans, with the patrol
            and outage records, so the report can say how often the watch would have spoken first.</p>
        </div>
        <p class="sec-more mono"><a href="{TUNING}#files">The report you read before you use it &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="homes">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;3 / HOMES</span>
          <h2>A mast above the roofline</h2>
          <p>Built with FireDefense Systems, which designs and installs exterior wildfire sprinkler
            systems for homes. The watch is a sealed head above the roof: a wide camera, a gas head and
            the models on board. It tells the owner first, with the reading that raised it.</p>
        </div>
        <div class="device-contract">
          <p><b>What goes in.</b> The mast&rsquo;s own air and heat, the wind, the camera, and the
            sprinkler controller&rsquo;s state.</p>
          <p><b>What comes out.</b> A message to the owner with its reason. If no one answers, every
            path the owner set up is tried in order (&sect;9).</p>
          <p><b>Who acts.</b> The owner. Pre-wet is the owner&rsquo;s rule, armed in advance and run
            by the controller, never by the model.</p>
          <p><b>What it is not.</b> A fire alarm or a sprinkler system. It does not replace
            code-required detection or the fire service.</p>
        </div>
      </div>
    </section>

    <section class="section band band--inset" id="land">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;4 / HILLSIDES AND OPEN LAND</span>
          <h2>Where no one is watching</h2>
          <p>A slope of brush under a ridge, a lot no one visits, a parcel between homes. A mast on
            solar keeps the same watch, and the warning goes to whoever holds the land.</p>
        </div>
        <div class="device-contract">
          <p><b>What goes in.</b> Chemistry, heat and wind at the mast, and a camera frame across the
            slope.</p>
          <p><b>What comes out.</b> A warning with its reason and the wind, to the owner, the county or
            the land trust, over whatever link reaches.</p>
          <p><b>Who acts.</b> The landholder and the people they call.</p>
          <p><b>What we ask you to send.</b> A season of the site&rsquo;s own air, with the days that
            were smoky for other reasons marked, so the watch learns your normal.</p>
        </div>
      </div>
    </section>

    <section class="section" id="portable">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;5 / PORTABLE</span>
          <h2>A watch on a trailer</h2>
          <p>Solar, a battery, a telescoping mast and the same models, for a prescribed burn, a crew,
            an event or a site for a few weeks. It reads the air and the camera together and gives its
            reason on the unit.</p>
        </div>
        <div class="device-contract">
          <p><b>What goes in.</b> The unit&rsquo;s chemistry, heat, wind and camera, read where it is
            parked.</p>
          <p><b>What comes out.</b> A warning with its reason, on the unit and to the crew&rsquo;s
            phones when a link is there.</p>
          <p><b>Who acts.</b> The burn boss or the site lead.</p>
          <p><b>Where it runs.</b> Entirely on the trailer. Moving it does not need a new network.</p>
        </div>
      </div>
    </section>

    <section class="section band band--inset" id="factories">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;6 / FACTORIES</span>
          <h2>Smoke, heat and gas inside the fence</h2>
          <p>The same reading inside a plant: a smouldering bin, a warm bearing, a gas that should not
            be there. It sits beside the plant&rsquo;s listed alarm and explains what it sees, never in
            place of it.</p>
        </div>
        <div class="device-contract">
          <p><b>What goes in.</b> Particulate, carbon monoxide, volatile organics and heat at the
            watch, the plant&rsquo;s own gas readings, and a camera where one is allowed.</p>
          <p><b>What comes out.</b> A note to the shift lead with the reading and where it is.</p>
          <p><b>Who acts.</b> The shift and the plant&rsquo;s safety team. The plant&rsquo;s alarm
            and suppression stay as they are.</p>
          <p><b>What we ask you to send.</b> The site&rsquo;s gas and heat history and its alarm
            journal.</p>
        </div>
        <p class="sec-more mono"><a href="/research-industry.html#industry">Other industries &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="more">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;7 / MORE PLACES</span>
          <h2>The same watch fits more places</h2>
          <p>Anywhere smoke or gas should be read on site and a person told why.</p>
        </div>
        <ul class="wl-cases">
          <li><b>Battery storage</b><span>Cells give off gas before they run away. A watch outside the
            container reads the air, the heat and a camera together.</span></li>
          <li><b>Solar farms</b><span>Wide sites, few people, dry ground under the panels.</span></li>
          <li><b>Recycling yards and landfills</b><span>Lithium batteries in the waste stream start fires
            that smoulder before they show.</span></li>
          <li><b>Vineyards</b><span>A record of the smoke the crop was exposed to, even when the fire is
            far away.</span></li>
          <li><b>Construction sites</b><span>Hot work done in the day, read through the night after the
            crew leaves.</span></li>
          <li><b>Rail and road corridors</b><span>Sparks start fires along the right of way, far from
            anyone.</span></li>
          <li><b>Telecom towers</b><span>Remote, on their own power, and needed most during a fire.</span></li>
        </ul>
      </div>
    </section>

    <section class="section band band--inset" id="measured">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;8 / MEASURED</span>
          <h2>What is measured, and how</h2>
          <p>Each number beside how it was tested. False-alarm rates at a real site over months are not
            measured yet, and nothing here is a claim about suppression.</p>
        </div>
        <ul class="wl-measured">
          <li><span>The fire controller, through its market guard</span><b>{@html nano.card('nano.market-controller')}</b>
            <small>1,200 ticks of recorded camera readings from one field site &middot; {@html nano.cardRef('nano.market-controller')}</small></li>
          <li><span>The fire vision model, smoke in a frame, through its market guard</span><b>AUC {@html nano.card('nano.market-vision')}</b>
            <small>1,129 held-out frames from the camera network it learned on &middot; {@html nano.cardRef('nano.market-vision')}</small></li>
          <li><span>A camera detector with the vision model confirming, false alarms</span><b>{@html nano.card('nano.vl-pair')}</b>
            <small>against {@html nano.card('nano.vl-det')} for the detector alone at the same recall, held-out frames &middot; {@html nano.cardRef('nano.vl-pair')}</small></li>
          <li><span>The sensor-reading adapter with chemistry and heat only, threat call / escalation</span><b>{@html nano.card('nano.fire-chem')}</b>
            <small>simulated fire scenarios with known answers &middot; {@html nano.cardRef('nano.fire-chem')}</small></li>
          <li><span>Override requests it should refuse that it complied with</span><b>{@html nano.card('nano.fd-unsafe')}</b>
            <small>simulated fire scenarios with rule-checked answers &middot; {@html nano.cardRef('nano.fd-unsafe')}</small></li>
        </ul>
        <p class="sec-more mono"><a href="/research-wave-nano.html#adapter">Every fire number, on the Wave Nano page &rarr;</a></p>
      </div>
    </section>

    <WaveFirewatch no='9' fig='2' />

    <section class="section research-closing">
      <div class="wrap">
        <h2>Start with one site.</h2>
        <p>A season of your own readings, one watch, and a report you read before you trust it.</p>
        <div class="research-actions research-actions--lead">
          <a class="research-button" href={TUNING}>Custom tuning</a>
          <a class="research-button" href="/research-wave-nano.html">Wave Nano</a>
          <a class="research-button" href="/research-industry.html">Industrial deployment</a>
        </div>
      </div>
    </section>
  </main>

  <Footer anchors='/' />
