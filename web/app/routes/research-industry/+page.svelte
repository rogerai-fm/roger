<script>
  // Industrial deployment - RogerAI
  // Ported from web/src/research-industry.html (2026-10-29); the copy is the legacy copy.
  import { siteHead, CSS_SHARED, CSS_ACCOUNT } from "#lib/head.js";
  import AdapterDial from "#lib/ui/AdapterDial.svelte";
  import Footer from "#lib/ui/Footer.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  const head = siteHead({
    title: 'Industrial deployment - RogerAI',
    desc: 'Local inference for oil and gas, power generation, manufacturing, aerospace, mining, water, healthcare, defense sustainment, and wildfire mitigation: where the box sits in the plant network, the standards it speaks, and what it never touches.',
    og: { type: 'website', title: 'RogerAI - local inference where the network stops', desc: 'Where a local model sits in a plant network, the standards it speaks, and the protection paths it never touches.', url: 'https://rogerai.fm/research-industry.html', image: 'https://rogerai.fm/og-research.png' },
    css: [...CSS_SHARED, 'research.css'],
    scripts: [{ runtime: true }],
  });
</script>
<svelte:head>{@html head}</svelte:head>
<!-- Industrial deployment, split out of the research hub when it outgrew the hub's
     conciseness budget. This is the page an OT buyer or a grant reviewer is actually
     sent, so it carries the placement diagram and the standards vocabulary in full. 
     (the page's doc comment, from the legacy file; its sheets are the css: list above) -->
  <Rail head='RogerAI &middot; Industrial' rev='Rev. 2026.07 &middot; LABS' />

  <Nav variant='marketing' home='/' promo={false} />

  <main id="top">
    <section class="research-hero">
      <div class="wrap">
        <a class="wf-back mono" href="/research.html">&larr; RogerAI Labs</a>
        <span class="hero__eyebrow">
          <OnAir />
          <span class="hero__callsign">INDUSTRIAL DEPLOYMENT</span>
        </span>
        <!-- Plain first paragraph (SITE-SIMPLIFICATION-AUDIT-2026-11-13, fix 9): what the page
             is about in words an OT engineer and a stranger both have. The Purdue placement,
             the standards and the boundaries follow unchanged. -->
        <h1>AI that runs inside the plant, because the data cannot leave.</h1>
        <p class="research-hero__lead">
          Factories, power plants, hospitals and remote sites keep their networks closed.
          A cloud model cannot be used there. RogerAI puts a small Wave model on a computer
          inside the site. It reads the sensors and the logs, explains what it sees, and
          never controls a machine. Below: the nine industries, where the box sits in the
          plant network, and what it never touches.
        </p>

        <!-- The nine markets, at a glance, above the fold - a visitor should not
             have to scroll to §1 to see whether their own industry is on the list.
             Each icon jumps straight to its own card in the deployment-grid below
             (id="market-<slug>", matching the asset filename slugs already used for
             the photos), rather than just to the section - a wendy.dev pattern
             (a use-case strip anchored to their detail grid), adapted to this
             page's own monochrome + one-red icon language (.pc-ic in home.css). -->
        <nav class="usecase-strip" aria-label="Jump to an industry">
          <a href="#market-oilgas"><svg class="uc-ic" viewBox="0 0 48 48" aria-hidden="true">
              <rect x="14" y="14" width="20" height="24" rx="2"/><path d="M14 22h20M14 30h20"/>
              <path class="uc-red" d="M24 6v8"/></svg><span>Oil &amp; gas</span></a>
          <a href="#market-power"><svg class="uc-ic" viewBox="0 0 48 48" aria-hidden="true">
              <path d="M26 6 16 26h9l-3 16 17-24h-10z"/><circle class="uc-fill" cx="35" cy="11" r="1.8"/></svg><span>Power</span></a>
          <a href="#market-manufacturing"><svg class="uc-ic" viewBox="0 0 48 48" aria-hidden="true">
              <circle cx="24" cy="24" r="8"/>
              <path d="M24 8v6M24 34v6M8 24h6M34 24h6M12.9 12.9l4.2 4.2M30.9 30.9l4.2 4.2M35.1 12.9l-4.2 4.2M17.1 30.9l-4.2 4.2"/>
              <circle class="uc-fill" cx="24" cy="24" r="2.4"/></svg><span>Manufacturing</span></a>
          <a href="#market-aerospace"><svg class="uc-ic" viewBox="0 0 48 48" aria-hidden="true">
              <path d="M24 6 27 20 40 27v3L27 26l-1 12 5 4v2l-7-2-7 2v-2l5-4-1-12L8 30v-3l13-7Z"/>
              <circle class="uc-fill" cx="24" cy="14" r="1.6"/></svg><span>Aerospace</span></a>
          <a href="#market-mining"><svg class="uc-ic" viewBox="0 0 48 48" aria-hidden="true">
              <path d="M6 36 16 18 22 28 28 14 42 36Z"/><path class="uc-red" d="M22 28 28 14"/></svg><span>Mining</span></a>
          <a href="#market-water"><svg class="uc-ic" viewBox="0 0 48 48" aria-hidden="true">
              <path d="M24 6C24 6 12 22 12 30a12 12 0 0 0 24 0C36 22 24 6 24 6Z"/>
              <path class="uc-red" d="M18 30a6 6 0 0 0 6 6"/></svg><span>Water</span></a>
          <a href="#market-healthcare"><svg class="uc-ic" viewBox="0 0 48 48" aria-hidden="true">
              <rect x="8" y="8" width="32" height="32" rx="6"/><path class="uc-red" d="M24 16v16M16 24h16"/></svg><span>Healthcare</span></a>
          <a href="#market-defense"><svg class="uc-ic" viewBox="0 0 48 48" aria-hidden="true">
              <path d="M24 6 38 12v12c0 10-6 16-14 19-8-3-14-9-14-19V12Z"/>
              <path class="uc-red" d="M18 24l4 4 8-10"/></svg><span>Defense</span></a>
          <a href="#market-wildfire"><svg class="uc-ic" viewBox="0 0 48 48" aria-hidden="true">
              <path d="M24 6c2 8 12 12 12 23a12 12 0 0 1-24 0c0-6 3-10 6-13 1 4 3 6 6 6 0-6-2-10 0-16Z"/>
              <path class="uc-red" d="M24 40a4 4 0 0 1-4-4c0-3 4-7 4-7s4 4 4 7a4 4 0 0 1-4 4Z"/></svg><span>Wildfire</span></a>
        </nav>
      </div>
    </section>

    <section class="section" id="industry">
<div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">§1 / MARKETS</span>
          <h2>Nine industries whose data cannot leave the site.</h2>
          <p>A cloud model is ruled out here. Three things rule it out: segmented
             networks, hazardous areas, and links too thin to stream sensor data. That is
             the constraint we build for. The industrial
             engagements are under NDA and not named here. The one company this page
             names, FireDefense Systems, is named with its agreement. In every
             deployment the model advises a person and stays outside the control loop:
             it never actuates equipment, and it is never part of a protection or
             safety path.</p>
        </div>
        <div class="deployment-grid">
          <article id="market-oilgas" class="tint-panel">
            <figure class="deployment-grid__shot">
              <img src="assets/industry/oilgas.webp" srcset="assets/industry/oilgas-500.webp 500w, assets/industry/oilgas.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1000px) 45vw, 22vw" width="1000" height="750" loading="lazy" decoding="async"
                   alt="A refinery's stacks and towers lit up at dusk.">
              <figcaption class="photo-credit">Photo <a href="https://unsplash.com/photos/u13zBF4r56A" target="_blank" rel="noopener noreferrer">Maksym Kaharlytskyi</a> &middot; Unsplash</figcaption>
            </figure>
            <b>Oil and gas</b><p>Rotating-equipment anomaly explanation and operator copilots
             grounded in the OEM manuals and P&amp;IDs already on site.</p>
            <!-- The one market with a page of its own (2026-11-21): the link covers the whole card
                 (research.css .deployment-grid__more), the photo credit stays clickable above it. -->
            <a class="deployment-grid__more mono" href="/research-oil-gas.html">Oil and gas, in depth &rarr;</a></article>
          <article id="market-power" class="tint-panel">
            <figure class="deployment-grid__shot">
              <img src="assets/industry/power.webp" srcset="assets/industry/power-500.webp 500w, assets/industry/power.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1000px) 45vw, 22vw" width="1000" height="750" loading="lazy" decoding="async"
                   alt="A transformer and switchgear inside a power substation.">
              <figcaption class="photo-credit">Photo <a href="https://unsplash.com/photos/VuR4oHZ3ucc" target="_blank" rel="noopener noreferrer">American Public Power Assoc.</a> &middot; Unsplash</figcaption>
            </figure>
            <b>Power generation</b><p>Historian questions in plain language, alarm-flood triage,
             and work-order drafting for turbines and substation assets.</p></article>
          <article id="market-manufacturing" class="tint-panel">
            <figure class="deployment-grid__shot">
              <img src="assets/industry/manufacturing.webp" srcset="assets/industry/manufacturing-500.webp 500w, assets/industry/manufacturing.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1000px) 45vw, 22vw" width="1000" height="750" loading="lazy" decoding="async"
                   alt="Robotic arms working a factory assembly line.">
              <figcaption class="photo-credit">Photo <a href="https://unsplash.com/photos/8gr6bObQLOI" target="_blank" rel="noopener noreferrer">Simon Kadula</a> &middot; Unsplash</figcaption>
            </figure>
            <b>Manufacturing</b><p>Prompt-defined inspection over the existing classifier, and
             shift notes, maintenance logs, and fault codes read together.</p></article>
          <article id="market-aerospace" class="tint-panel">
            <figure class="deployment-grid__shot">
              <img src="assets/industry/aerospace.webp" srcset="assets/industry/aerospace-500.webp 500w, assets/industry/aerospace.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1000px) 45vw, 22vw" width="1000" height="750" loading="lazy" decoding="async"
                   alt="A jet aircraft parked inside a maintenance hangar.">
              <figcaption class="photo-credit">Photo <a href="https://unsplash.com/photos/vWUOx-3r7SY" target="_blank" rel="noopener noreferrer">Hermeus</a> &middot; Unsplash</figcaption>
            </figure>
            <b>Aerospace</b><p>Document intelligence and assistant workloads on
             disconnected networks, where deployment is the security model.</p></article>
          <article id="market-mining" class="tint-panel">
            <figure class="deployment-grid__shot">
              <img src="assets/industry/mining.webp" srcset="assets/industry/mining-500.webp 500w, assets/industry/mining.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1000px) 45vw, 22vw" width="1000" height="750" loading="lazy" decoding="async"
                   alt="Excavators and haul trucks working an open-pit mine.">
              <figcaption class="photo-credit">Photo <a href="https://unsplash.com/photos/Mk2ls9UBO2E" target="_blank" rel="noopener noreferrer">Dominik Vanyi</a> &middot; Unsplash</figcaption>
            </figure>
            <b>Mining and metals</b><p>Fixed-plant and haul-fleet condition explanation, with
             shift and geology logs read together at sites a link barely reaches.</p></article>
          <article id="market-water" class="tint-panel">
            <figure class="deployment-grid__shot">
              <img src="assets/industry/water.webp" srcset="assets/industry/water-500.webp 500w, assets/industry/water.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1000px) 45vw, 22vw" width="1000" height="750" loading="lazy" decoding="async"
                   alt="Circular treatment tanks at a water treatment plant, seen from above.">
              <figcaption class="photo-credit">Photo <a href="https://unsplash.com/photos/6wSevhW1Dzc" target="_blank" rel="noopener noreferrer">Ivan Bandura</a> &middot; Unsplash</figcaption>
            </figure>
            <b>Water and wastewater</b><p>Treatment-process questions against the local
             historian, and permit and compliance drafting on a network that does not route out.</p></article>
          <article id="market-healthcare" class="tint-panel">
            <figure class="deployment-grid__shot">
              <img src="assets/industry/healthcare.webp" srcset="assets/industry/healthcare-500.webp 500w, assets/industry/healthcare.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1000px) 45vw, 22vw" width="1000" height="750" loading="lazy" decoding="async"
                   alt="An empty hospital imaging suite with a CT scanner.">
              <figcaption class="photo-credit">Photo <a href="https://unsplash.com/photos/QTP8UKW_PgI" target="_blank" rel="noopener noreferrer">Accuray</a> &middot; Unsplash</figcaption>
            </figure>
            <b>Healthcare</b><p>The equipment layer of a hospital: imaging-fleet service logs,
             biomedical work orders, alarm-configuration review, and cold-chain records - on the
             segmented networks that data cannot leave.</p></article>
          <article id="market-defense" class="tint-panel">
            <figure class="deployment-grid__shot">
              <img src="assets/industry/defense.webp" srcset="assets/industry/defense-500.webp 500w, assets/industry/defense.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1000px) 45vw, 22vw" width="1000" height="750" loading="lazy" decoding="async"
                   alt="A military aircraft undergoing maintenance inside a hangar.">
              <figcaption class="photo-credit">Photo <a href="https://unsplash.com/photos/8QZGzUbatJU" target="_blank" rel="noopener noreferrer">Grant Sams</a> &middot; Unsplash</figcaption>
            </figure>
            <b>Defense sustainment</b><p>Maintenance, readiness, and supply on denied,
             degraded, and intermittent links, where there is no reachback to a cloud.</p></article>
          <!-- Wildfire (founder-approved 2026-09-24): the picture is a text-free frame from
               our own FireDefense film (no title card, no caption, no company name burned
               in), not stock photography - and no fire is ever shown on a home. -->
          <article id="market-wildfire" class="tint-panel">
            <figure class="deployment-grid__shot">
              <img src="assets/industry/wildfire.webp" srcset="assets/industry/wildfire-500.webp 500w, assets/industry/wildfire.webp 1000w" sizes="(max-width: 666px) 90vw, (max-width: 1000px) 45vw, 22vw" width="1000" height="750" loading="lazy" decoding="async"
                   alt="A single-story house on a hillside of dense chaparral, a long concrete driveway leading up to it.">
              <figcaption class="photo-credit">Film still &middot; RogerAI</figcaption>
            </figure>
            <b>Wildfire mitigation</b><p>An early-warning watch on the home itself: smoke chemistry
             against a baseline and a camera to confirm, built to run when the grid goes first.</p>
            <a class="deployment-grid__more mono" href="/research-wildfire.html">Wildfire mitigation, in depth &rarr;</a></article>
        </div>

        <!-- The three boundaries a reader must not have to infer. All are the SAME
             no-actuation contract the control-loop paragraph above states; they are
             restated here because healthcare, defense and wildfire are where a reader
             will (correctly) ask what we are claiming. -->
        <div class="device-contract">
            <p><b>Healthcare: equipment and facilities, not patients.</b> What these models
              read is the hospital&rsquo;s equipment and the paperwork around it &ndash;
              imaging-fleet service logs, biomedical work orders, cold-chain records. They do
              not diagnose, do not treat, do not recommend or adjust care, and do not read
              scans or physiologic waveforms. Alarm-configuration review means reading past
              alarm <i>journals</i> to find a threshold set wrong; it is never a decision
              about a live alarm at a bedside. Software that makes clinical decisions is
              regulated as a medical device, and this line is what keeps us on the equipment
              side of it.</p>
            <p><b>Defense: sustainment, not effects.</b> Sustainment is the logistics half of
              defense work &ndash; keeping equipment running and supplied. That means
              maintenance histories, readiness reporting, and parts and supply questions,
              answered on site over links that are denied, degraded or intermittent, with no
              reachback to a cloud. It does not mean the other half: no targeting, no weapons
              employment, and no intelligence exploitation. The same models, the same
              constraint as every other industry here &ndash; they read and they explain, and
              nothing they say moves a machine or a decision on its own.</p>
            <p><b>Wildfire: warning and mitigation, not fire protection.</b> On a home in the
              wildland-urban interface these models watch and explain: gas and fine
              particulate tracked against a rolling baseline, a camera that confirms what the
              traces suggest, and wind and slope read together. The model raises the alert
              and makes the threat call, with a confidence. It is not a listed fire alarm or
              sprinkler system, it is never in the control path of one, and it does not
              replace code-required detection or suppression, an evacuation order, or the
              fire service. A valve opens only on the owner&rsquo;s command, or under a rule
              the owner armed in advance, which the controller &ndash; not the model &ndash;
              executes with its own abort window. The model never opens a valve by itself;
              under an armed rule, its call is one input the owner chose to let the
              controller read.</p>
        </div>

        <!-- The wildfire use case, built with FireDefense Systems (named with their
             agreement). What runs today is stated as it is: the mast is shown in the
             film, Pico is on air, Nano's two fire models are on air and Nano chat is off (web/src/data/wave-status.json). -->
        <div class="device-contract">
          <p><b>Use case: an edge fire watch.</b> Built with FireDefense Systems, a separate
            company that designs and installs exterior wildfire sprinkler systems for homes.
            The watch shown in the FireDefense film is a sealed
            sensor head on a mast above the roofline: a wide camera, a gas head for volatile
            organics, carbon monoxide and fine particulate, and the models on board. A
            wildfire announces itself chemically before it does optically, so the watch
            smells first and looks second: the watch treats one trace leaving baseline as
            weather, three leaving together as combustion.</p>
          <p><b>Where it runs.</b> The design puts the models on the mast, on its own
            battery, with no network required. Wave Pico, on air today as wave-pico-293m,
            is the tier meant to keep the always-on watch. The threat call, which fuses
            gas, image, wind and slope, is Wave Nano work. Nano&rsquo;s fire controller and fire vision are on air behind market guards, as specialized models. In a
            wildfire the towers and the power tend to go before the fire arrives, so a model
            that has to reach a datacenter to think has already lost.</p>
          <p><b>What it never does.</b> It would alert the owner first, and a message that cannot
            leave the property is the case it is built for. It never opens a valve by itself
            and never overrides the sprinkler controller or any alarm. What people should do
            in a fire is for the fire service and official alerts to say, never the watch.
            <a href="/assets/edge/hero-fd.mp4">Watch the FireDefense film &rarr;</a></p>
        </div>

        <!-- Where the box sits - the first question asked in OT, and one prose cannot
             answer because the answer is a position in a hierarchy. Safety is drawn
             untouched rather than described that way. Since 2026-11-22 a building cut open
             in the isometric hand (components.css .iso__*; generated by
             computer-scientist/article/fig_purdue_gen.py between the markers): the six
             levels are floor plates, the Wave console stands on L3 with its riser leaving
             through the DMZ and the roof, Roger Edge stands on the ground floor apart from
             the sensors and wired to nothing, safety on its own ground across a gap. The
             words are HTML: the level names in a column beside the stage, the names and
             tags over it. The packets ride CSS offset-path (research.css). -->
        <figure class="purdue" aria-labelledby="purdue-caption">
          <div class="iso__stage purdue__stage">
            <ol class="purdue__levels">
              <li class="purdue__level purdue__level--l54"><b class="purdue__lv">L5&#8211;L4</b> enterprise &#183; business IT</li>
              <li class="purdue__level purdue__level--dmz"><b class="purdue__lv">L3.5</b> DMZ</li>
              <li class="purdue__level purdue__level--l3"><b class="purdue__lv">L3</b> site operations &#183; historian, MES</li>
              <li class="purdue__level purdue__level--l2"><b class="purdue__lv">L2</b> supervisory &#183; SCADA, HMI</li>
              <li class="purdue__level purdue__level--l1"><b class="purdue__lv">L1</b> basic control &#183; PLC, RTU</li>
              <li class="purdue__level purdue__level--l0"><b class="purdue__lv">L0</b> process &#183; sensors, actuators</li>
            </ol>
            <div class="purdue__pic">
            <!-- iso:begin -->
            <svg class="iso__svg" viewBox="-33.7 -91.7 257.1 180.5" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-3.2 -2.7 L163.1 45.3 L138.6 66.5 L-27.7 18.5 Z"/>
<path class="iso__halo" d="M172.3 50.3 L217.4 63.3 L194.9 82.8 L149.8 69.8 Z"/>
<g class="purdue__plate purdue__plate--l0">
<path class="iso__face iso__face--s1" d="M-20 14.7 L135.9 59.7 L135.9 62.3 L-20 17.3 Z"/>
<path class="iso__face iso__face--s2" d="M155.9 42.4 L135.9 59.7 L135.9 62.3 L155.9 45 Z"/>
<path class="iso__face" d="M0 -2.6 L155.9 42.4 L135.9 59.7 L-20 14.7 Z"/>
</g>
<g class="purdue__m purdue__sensors">
<path class="iso__face iso__face--s1" d="M82.1 36.9 L82.1 42.1 L82.1 42.3 L82.2 42.5 L82.4 42.6 L82.6 42.8 L82.9 43 L83.2 43.1 L83.5 43.2 L83.9 43.3 L84.3 43.3 L84.7 43.4 L85.1 43.3 L85.5 43.3 L85.9 43.2 L86.2 43.1 L86.5 43 L86.8 42.8 L87 42.6 L87.2 42.5 L87.3 42.3 L87.3 42.1 L87.3 36.9 Z"/>
<path class="iso__face" d="M86.9 37.5 L86.5 37.8 L86 38 L85.4 38.1 L84.7 38.2 L84 38.1 L83.4 38 L82.9 37.8 L82.4 37.5 L82.2 37.2 L82.1 36.9 L82.2 36.5 L82.4 36.2 L82.9 35.9 L83.4 35.7 L84 35.6 L84.7 35.6 L85.4 35.6 L86 35.7 L86.5 35.9 L86.9 36.2 L87.2 36.5 L87.3 36.9 L87.2 37.2 Z"/>
<path class="iso__face iso__face--s1" d="M94.2 40.4 L94.2 45.6 L94.2 45.8 L94.3 46 L94.5 46.1 L94.7 46.3 L95 46.5 L95.3 46.6 L95.6 46.7 L96 46.8 L96.4 46.8 L96.8 46.9 L97.2 46.8 L97.6 46.8 L98 46.7 L98.3 46.6 L98.7 46.5 L98.9 46.3 L99.1 46.1 L99.3 46 L99.4 45.8 L99.4 45.6 L99.4 40.4 Z"/>
<path class="iso__face" d="M99.1 41 L98.7 41.3 L98.1 41.5 L97.5 41.6 L96.8 41.7 L96.1 41.6 L95.5 41.5 L95 41.3 L94.6 41 L94.3 40.7 L94.2 40.4 L94.3 40 L94.6 39.7 L95 39.4 L95.5 39.2 L96.1 39.1 L96.8 39.1 L97.5 39.1 L98.1 39.2 L98.7 39.4 L99.1 39.7 L99.3 40 L99.4 40.4 L99.3 40.7 Z"/>
<path class="iso__face iso__face--s1" d="M106.3 43.9 L106.3 49.1 L106.4 49.3 L106.5 49.5 L106.6 49.6 L106.8 49.8 L107.1 50 L107.4 50.1 L107.8 50.2 L108.1 50.3 L108.5 50.3 L108.9 50.4 L109.3 50.3 L109.7 50.3 L110.1 50.2 L110.5 50.1 L110.8 50 L111 49.8 L111.3 49.6 L111.4 49.5 L111.5 49.3 L111.5 49.1 L111.5 43.9 Z"/>
<path class="iso__face" d="M111.2 44.5 L110.8 44.8 L110.2 45 L109.6 45.1 L108.9 45.2 L108.3 45.1 L107.6 45 L107.1 44.8 L106.7 44.5 L106.4 44.2 L106.3 43.9 L106.4 43.5 L106.7 43.2 L107.1 42.9 L107.6 42.7 L108.3 42.6 L108.9 42.6 L109.6 42.6 L110.2 42.7 L110.8 42.9 L111.2 43.2 L111.5 43.5 L111.5 43.9 L111.5 44.2 Z"/>
<path class="iso__face iso__face--s1" d="M115.6 47.2 L127.7 50.7 L127.7 56.8 L115.6 53.3 Z"/>
<path class="iso__face iso__face--s2" d="M131.7 47.3 L127.7 50.7 L127.7 56.8 L131.7 53.3 Z"/>
<path class="iso__face" d="M119.6 43.8 L131.7 47.3 L127.7 50.7 L115.6 47.2 Z"/>
<path class="iso__face iso__face--cav" d="M123.6 52.6 L123.4 51.7 L122.9 50.9 L122.1 50.3 L121.2 50 L120.5 50.2 L119.9 50.7 L119.8 51.5 L119.9 52.3 L120.5 53.2 L121.2 53.7 L122.1 54 L122.9 53.8 L123.4 53.3 Z"/>
</g>
<g class="purdue__edge iso__m purdue__m">
<path class="iso__face iso__face--s1" d="M-6.1 14.4 L4.2 17.4 L4.2 21.7 L-6.1 18.7 Z"/>
<path class="iso__face iso__face--s2" d="M8.7 13.5 L4.2 17.4 L4.2 21.7 L8.7 17.8 Z"/>
<path class="iso__face" d="M-1.6 10.5 L8.7 13.5 L4.2 17.4 L-6.1 14.4 Z"/>
<path class="iso__ln" d="M1.3 13.9 L1.3 7.9"/>
<circle class="iso__lamp iso__lamp--on" cx="1.3" cy="7.9" r="1.4"/>
<path class="iso__face iso__face--s1" d="M9.1 18.9 L9.1 21.9 L9.2 22.1 L9.2 22.2 L9.3 22.3 L9.5 22.5 L9.7 22.6 L9.9 22.7 L10.1 22.7 L10.4 22.8 L10.7 22.8 L10.9 22.8 L11.2 22.8 L11.5 22.8 L11.8 22.7 L12 22.7 L12.2 22.6 L12.4 22.5 L12.5 22.3 L12.7 22.2 L12.7 22.1 L12.7 21.9 L12.7 18.9 Z"/>
<path class="iso__face" d="M12.5 19.3 L12.2 19.5 L11.8 19.7 L11.4 19.8 L10.9 19.8 L10.5 19.8 L10 19.7 L9.7 19.5 L9.4 19.3 L9.2 19.1 L9.1 18.9 L9.2 18.7 L9.4 18.4 L9.7 18.3 L10 18.1 L10.5 18 L10.9 18 L11.4 18 L11.8 18.1 L12.2 18.3 L12.5 18.4 L12.7 18.7 L12.7 18.9 L12.7 19.1 Z"/>
</g>
<g class="purdue__plate purdue__plate--l1">
<path class="iso__face iso__face--s1" d="M-20 -0.9 L135.9 44.1 L135.9 46.7 L-20 1.7 Z"/>
<path class="iso__face iso__face--s2" d="M155.9 26.8 L135.9 44.1 L135.9 46.7 L155.9 29.4 Z"/>
<path class="iso__face" d="M0 -18.2 L155.9 26.8 L135.9 44.1 L-20 -0.9 Z"/>
</g>
<g class="purdue__m purdue__plc">
<path class="iso__face iso__face--s1" d="M64.9 17.6 L90.9 25.1 L90.9 31.1 L64.9 23.6 Z"/>
<path class="iso__face iso__face--s2" d="M94.4 22 L90.9 25.1 L90.9 31.1 L94.4 28.1 Z"/>
<path class="iso__face" d="M68.4 14.5 L94.4 22 L90.9 25.1 L64.9 17.6 Z"/>
<path class="iso__ln iso__ln--faint" d="M68.3 23.6 L68.3 19.6"/>
<path class="iso__ln iso__ln--faint" d="M71.8 24.6 L71.8 20.6"/>
<path class="iso__ln iso__ln--faint" d="M75.3 25.6 L75.3 21.6"/>
<path class="iso__ln iso__ln--faint" d="M78.7 26.6 L78.7 22.6"/>
<path class="iso__ln iso__ln--faint" d="M82.2 27.6 L82.2 23.6"/>
<path class="iso__ln iso__ln--faint" d="M85.7 28.6 L85.7 24.6"/>
<circle class="iso__lamp" cx="66.6" cy="19.3" r="1.1"/>
</g>
<g class="purdue__plate purdue__plate--l2">
<path class="iso__face iso__face--s1" d="M-20 -16.5 L135.9 28.5 L135.9 31.1 L-20 -13.9 Z"/>
<path class="iso__face iso__face--s2" d="M155.9 11.2 L135.9 28.5 L135.9 31.1 L155.9 13.8 Z"/>
<path class="iso__face" d="M0 -33.8 L155.9 11.2 L135.9 28.5 L-20 -16.5 Z"/>
</g>
<g class="purdue__m purdue__scada">
<path class="iso__face iso__face--s1" d="M70.9 8.1 L83.1 11.6 L83.1 13.3 L70.9 9.8 Z"/>
<path class="iso__face iso__face--s2" d="M86.1 9 L83.1 11.6 L83.1 13.3 L86.1 10.7 Z"/>
<path class="iso__face" d="M73.9 5.5 L86.1 9 L83.1 11.6 L70.9 8.1 Z"/>
<path class="iso__face iso__face--s1" d="M69.8 -1.8 L87.2 3.2 L87.2 11 L69.8 6 Z"/>
<path class="iso__face iso__face--s2" d="M88.2 2.4 L87.2 3.2 L87.2 11 L88.2 10.1 Z"/>
<path class="iso__face" d="M70.8 -2.6 L88.2 2.4 L87.2 3.2 L69.8 -1.8 Z"/>
<path class="iso__face iso__face--s2" d="M71.1 5.1 L85.9 9.3 L85.9 4.1 L71.1 -0.2 Z"/>
<path class="iso__ln iso__ln--lo" d="M73.3 2.3 L80.2 4.3"/>
<path class="iso__ln iso__ln--lo" d="M73.3 3.9 L77.6 5.1"/>
</g>
<path class="iso__ln iso__ln--lo purdue__riser purdue__riser--read" d="M79 0.3 L79 -7.5"/>
<g class="purdue__plate purdue__plate--l3">
<path class="iso__face iso__face--s1" d="M-20 -35.5 L135.9 9.5 L135.9 12.1 L-20 -32.9 Z"/>
<path class="iso__face iso__face--s2" d="M155.9 -7.8 L135.9 9.5 L135.9 12.1 L155.9 -5.2 Z"/>
<path class="iso__face" d="M0 -52.8 L155.9 -7.8 L135.9 9.5 L-20 -35.5 Z"/>
</g>
<g class="purdue__m purdue__hist">
<path class="iso__face iso__face--s1" d="M36.2 -29.1 L60.4 -22.1 L60.4 -15.2 L36.2 -22.2 Z"/>
<path class="iso__face iso__face--s2" d="M64.9 -26 L60.4 -22.1 L60.4 -15.2 L64.9 -19.1 Z"/>
<path class="iso__face" d="M40.7 -33 L64.9 -26 L60.4 -22.1 L36.2 -29.1 Z"/>
<path class="iso__ln iso__ln--faint" d="M38.8 -23.2 L57.8 -17.7"/>
<path class="iso__ln iso__ln--faint" d="M38.8 -24.9 L57.8 -19.4"/>
<path class="iso__ln iso__ln--faint" d="M38.8 -26.6 L57.8 -21.1"/>
</g>
<g class="purdue__us iso__m purdue__m">
<path class="iso__face iso__face--s1" d="M66.6 -22.6 L87.4 -16.6 L87.4 -4.5 L66.6 -10.5 Z"/>
<path class="iso__face iso__face--s2" d="M92.4 -21 L87.4 -16.6 L87.4 -4.5 L92.4 -8.8 Z"/>
<path class="iso__face" d="M71.6 -27 L92.4 -21 L87.4 -16.6 L66.6 -22.6 Z"/>
<path class="iso__ln iso__ln--faint" d="M69.2 -12.4 L84.8 -7.9"/>
<path class="iso__ln iso__ln--faint" d="M69.2 -14.5 L84.8 -10"/>
<path class="iso__ln iso__ln--faint" d="M69.2 -16.7 L84.8 -12.2"/>
<path class="iso__face iso__face--cav" d="M70.1 -18.6 L83.9 -14.6 L83.9 -15.9 L70.1 -19.9 Z"/>
<path class="iso__ln" d="M73.6 -24.7 L73.6 -29.9"/>
<circle class="iso__lamp iso__lamp--on" cx="73.6" cy="-29.9" r="1.5"/>
<circle class="iso__lamp iso__lamp--on" cx="84.4" cy="-14.9" r="1.7"/>
</g>
<g class="purdue__plate purdue__plate--dmz purdue__dmz">
<path class="iso__face iso__face--s1" d="M-20 -58.5 L135.9 -13.5 L135.9 -12.2 L-20 -57.2 Z"/>
<path class="iso__face iso__face--s2" d="M155.9 -30.8 L135.9 -13.5 L135.9 -12.2 L155.9 -29.5 Z"/>
<path class="iso__face" d="M0 -75.8 L155.9 -30.8 L135.9 -13.5 L-20 -58.5 Z"/>
</g>
<path class="iso__face iso__face--cav" d="M78.9 -35.1 L84.1 -33.6 L81.1 -31 L75.9 -32.5 Z"/>
<path class="iso__ln purdue__riser purdue__riser--out" d="M80 -22.2 L80 -31.8"/>
<g class="purdue__plate purdue__plate--l54">
<path class="iso__face iso__face--s1" d="M-20 -68.4 L135.9 -23.4 L135.9 -20.8 L-20 -65.8 Z"/>
<path class="iso__face iso__face--s2" d="M155.9 -40.7 L135.9 -23.4 L135.9 -20.8 L155.9 -38.1 Z"/>
<path class="iso__face" d="M0 -85.7 L155.9 -40.7 L135.9 -23.4 L-20 -68.4 Z"/>
</g>
<path class="iso__face iso__face--cav" d="M78.9 -45.1 L84.1 -43.6 L81.1 -41 L75.9 -42.5 Z"/>
<path class="iso__ln purdue__riser purdue__riser--out" d="M80 -33.1 L80 -40.4"/>
<g class="purdue__m purdue__it">
<path class="iso__face iso__face--s1" d="M67.5 -48.4 L86.5 -42.9 L86.5 -37.7 L67.5 -43.2 Z"/>
<path class="iso__face iso__face--s2" d="M92.5 -48.1 L86.5 -42.9 L86.5 -37.7 L92.5 -42.9 Z"/>
<path class="iso__face" d="M73.5 -53.6 L92.5 -48.1 L86.5 -42.9 L67.5 -48.4 Z"/>
<path class="iso__ln iso__ln--faint" d="M70.9 -43.2 L70.9 -46.3"/>
<path class="iso__ln iso__ln--faint" d="M74.4 -42.2 L74.4 -45.3"/>
<path class="iso__ln iso__ln--faint" d="M77.9 -41.2 L77.9 -44.3"/>
<path class="iso__ln iso__ln--faint" d="M81.3 -40.2 L81.3 -43.3"/>
</g>
<path class="iso__ln iso__ln--lo purdue__esc" d="M-4.1 14.8 L-24.1 9.1 L-24.1 -41.2 L-25.6 -39.9 L67.1 -13.1"/>
<g class="purdue__m purdue__sis">
<path class="iso__face iso__face--s1" d="M156.7 65.7 L193 76.2 L193 78.8 L156.7 68.3 Z"/>
<path class="iso__face iso__face--s2" d="M211 60.6 L193 76.2 L193 78.8 L211 63.2 Z"/>
<path class="iso__face" d="M174.7 50.1 L211 60.6 L193 76.2 L156.7 65.7 Z"/>
<path class="iso__face iso__face--s1" d="M166.6 49.5 L189.1 56 L189.1 71.6 L166.6 65.1 Z"/>
<path class="iso__face iso__face--s2" d="M200.1 46.5 L189.1 56 L189.1 71.6 L200.1 62.1 Z"/>
<path class="iso__face" d="M177.6 40 L200.1 46.5 L189.1 56 L166.6 49.5 Z"/>
<path class="iso__ln iso__ln--faint" d="M169.2 62.4 L186.5 67.4"/>
<path class="iso__ln iso__ln--faint" d="M169.2 59.8 L186.5 64.8"/>
<path class="iso__ln iso__ln--faint" d="M169.2 57.2 L186.5 62.2"/>
<circle class="iso__lamp" cx="186.1" cy="57.8" r="1.6"/>
</g>
<g class="purdue__packets">
<circle class="purdue__pkt purdue__pkt--read" r="1.8"/>
<circle class="purdue__pkt purdue__pkt--out" r="2"/>
</g>
            </svg>
            <!-- iso:end -->
            <span class="iso__name purdue__name purdue__name--us iso__ko">ROGERAI</span>
            <span class="iso__sub purdue__sub purdue__sub--us iso__ko">local model &#183; advisory</span>
            <!-- Roger Edge is NOT the L3 box moved downward. It is a different kind of
                 thing: a 100 KB - 2 MB classifier on its own battery-powered sensor node,
                 not a language model at all. It stands beside the instrumentation because
                 that is where its sensor is. It sits at L0 the way the Wave box sits at
                 L3, but it is NOT in the control path: it carries its own sensor and is
                 wired to nothing in the PLC. It emits events and escalates upward
                 whatever it cannot classify, which is the dashed way. -->
            <span class="iso__name purdue__name purdue__name--edge">ROGER EDGE</span>
            <span class="iso__sub purdue__sub purdue__sub--edge">detects &#183; its own sensor node</span>
            <span class="iso__label purdue__tag purdue__tag--out iso__ko">outbound only</span>
            <span class="iso__label purdue__tag purdue__tag--read iso__ko">reads</span>
            <span class="iso__label purdue__tag purdue__tag--esc">escalates</span>
            <span class="iso__name purdue__name purdue__name--sis">SAFETY</span>
            <span class="iso__sub purdue__sub purdue__sub--sis">SIS &#183; interlocks</span>
            <span class="iso__sub purdue__sub purdue__sub--sis purdue__sub--never">never touched</span>
            </div>
          </div>
          <figcaption id="purdue-caption">Two boxes, at two levels, doing two different
            jobs. The Wave model runs at Level 3, beside the historian and MES, reading
            plant data and answering questions about it. It initiates outbound through the
            DMZ and exposes no inbound listener below it. It writes nothing to supervisory
            or basic control, and it performs no closed-loop control. The safety-instrumented system
            is outside its reach entirely, so protection and interlocks stay deterministic.
            It works alongside classical analytics rather than replacing them.
            Roger Edge sits down at the process, because that is where its sensor is. It is
            not a language model: it is a classifier of a few hundred kilobytes on its own
            battery-powered node, wired to nothing in the control system. It detects, emits
            an event with a confidence, and escalates upward whatever it cannot classify.
            Roger Edge detects; Wave reasons.</figcaption>
        </figure>

        <!-- THE ADAPTER SHELF (features/web/sensor_positioning.feature, PROPOSED 2026-10-15,
             founder review). One base, a small add-on per job, every lamp a cap token from
             src/data/capabilities.json (the same dial as the homepage). Design tense, no
             numbers, no company, nothing claimed in the field: the wildfire card, boundary
             and use case above are untouched and still govern every fire word. -->
        <div class="adapter-shelf" id="adapters">
          <h3>Capabilities come as adapters.</h3>
          <!-- The adapter definition is the one sentence shared by every page that uses the word. -->
          <p>A Wave model is one base model. An adapter is a small add-on file, trained for
            one job, that plugs into a Wave model without changing it. Each extra job, such
            as fire early warning, comes as one and is switched on only for that job, so
            everything else runs on the base exactly as before. Each lamp below reads the
            same status file as every other page.</p>
          <AdapterDial />

          <!-- Escalates when unsure. Mechanism only: the gate is drawn with no number,
               and handing off to a person is the good outcome, drawn first and in the
               positive style (playbox_edge_honesty.feature: ESCALATE is a first-class
               good outcome). -->
          <figure class="figure figure--plate gate">
            <div class="gate__fig" role="group" aria-labelledby="gate-caption">
              <div class="gate__scale" aria-hidden="true">
                <span>less sure</span><i class="gate__tick"></i><span>sure enough to vouch for</span>
              </div>
              <div class="gate__ways">
                <div class="gate__way gate__way--good"><span class="tint-panel__tag">short of the gate</span><b>A person decides.</b>
                  The reading goes to a person, with what the model saw. That is the right
                  call, not a fault.</div>
                <div class="gate__way"><span class="tint-panel__tag">past the gate</span><b>A warning, with why.</b>
                  It raises a warning that says what it saw and how sure it is. A person
                  still makes the call.</div>
              </div>
            </div>
            <figcaption class="fig" id="gate-caption">Illustrative: escalates when unsure.
              Either way a person makes the call; the hand-off is what the gate is built
              for.</figcaption>
          </figure>
        </div><!-- /adapter-shelf -->

        <!-- The plant-interface contract. An OT buyer and a grant reviewer check the same
             thing first: does this vendor speak the plant's language, and does it know
             where its box is allowed to sit. -->
        <div class="device-contract">
          <p><b>Signals.</b> OPC UA and Modbus for tags; MQTT Sparkplug B over a unified
             namespace for transport.</p>
          <p><b>Models.</b> ISA-95 for the asset hierarchy and NAMUR NE 107 for device-health
             status, so a verdict lands in the vocabulary operators already use.</p>
          <p><b>Placement.</b> IEC 62443 zones and the Purdue levels, drawn above. The
             model reasons beside classical analytics, never in place of it.</p>
        </div>
        <div class="device-contract" id="devices">
          <p><b>Device evidence contract.</b> A device claim names the exact board,
             runtime, format, quantization, context, peak memory, and measured speed.</p>
          <p><b>Raspberry Pi and NPU.</b> Pi results name model and RAM. NPU results report
             full delegation, partial delegation, and CPU fallback, not merely a successful load.</p>
          <p><b>ESP32.</b> Roger Edge is the sensing and glue layer: wake, VAD, sensing,
             and fixed commands. No Wave tier runs on an ESP32 - the board is the sensor
             and a Raspberry Pi is the model host. The
             <a href="/research-models.html#roger-edge">Roger Edge line</a> carries its own
             status.</p>
        </div>
      </div>
        </section>

    <!-- Why the tiers are what they are. This is the section an OT reviewer can check
         independently, and it is a much stronger claim than "we designed it this way":
         hazardous-area certification and solar power budgets decide the answer before
         anyone picks a model. Standards and physics only - vendor names, prices and
         product matrices are market intelligence carrying their own confidence labels
         and do not belong on a public page. -->
    <section class="section" id="envelope">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">§2 / THE ENVELOPE</span>
          <h2>The tiers are not a preference. They are what certification allows.</h2>
          <p>Before anyone talks architecture, three practical questions settle where a
             model can physically run.</p>
        </div>

        <div class="device-contract">
          <p><b>Heat.</b> Temperature class T4 caps surface temperature at 135&deg;C at
             maximum ambient, and oil and gas sites routinely ask for T3 or T4. In a
             sealed fanless enclosure at 50 to 60&deg;C ambient that is the constraint
             that kills compute, long before anyone runs out of TOPS.</p>
          <p><b>Power in the zone.</b> Intrinsic safety (Ex ia) limits usable power to
             roughly 1.2 to 1.3 watts. That is a transmitter or a microcontroller, not an
             accelerator. <b>We have not found a hazardous-area-certified GPU enclosure on
             the market.</b> The shapes that do work are a safe-area rack, a purged
             cabinet, light inference on a certified low-power gateway, or sensors in the
             zone with the compute outside on fiber.</p>
          <p><b>Power at a remote site.</b> Remote sites size three to seven days of
             battery autonomy. Every continuous watt costs roughly a 10 to 15 watt panel
             and 30 to 60 watt-hours of battery, so a 15 watt device is a 200 to 300 watt
             array and a 2 to 4 kilowatt-hour bank - a skid of its own. Duty-cycled
             inference or microcontroller-class compute is the only thing that fits a
             solar wellhead.</p>
        </div>

      </div>
    </section>

    <section class="section research-closing">
      <div class="wrap">
        <span class="sectionno">§3 / TALK TO US</span>
        <h2>Patterns, not customer logos.</h2>
        <p>Industrial work happens under NDA, so this page shows the deployment shapes we
           build rather than the engagements behind them. When a piece of work can be
           published, it lands in the notebook with its numbers. If one of these shapes
           looks like your plant, write to us.</p>
        <div class="research-actions research-actions--lead">
          <a class="research-button" data-lets-talk aria-expanded="false"
             href="mailto:labs@rogerai.fm?subject=Industrial%20pilot">Contact RogerAI Labs</a>
          <a class="research-button" href="/research.html">Back to Labs</a>
          <a class="research-button" href="/broadcasts.html">Open Air Waves</a>
        </div>
      </div>
    </section>
  </main>

  <Footer anchors='/' />
