<script>
  import TocTuner from "#lib/ui/TocTuner.svelte";
  // Oil and gas - five jobs for a Wave model (2026-11-21, founder request). The industrial page's
  // oil and gas card opens here. Ten use cases were drawn up from three research briefs
  // (the lab's oil and gas briefs, 2026-11-21) and narrowed to the five a
  // small on-site model could plausibly do from records a site already keeps; the list and the
  // reasons are in OIL-GAS-USE-CASES-2026-11-21.md. Show, don't tell: each section is a heading,
  // a line, a strip of one-line cards, a figure where a picture says it faster, and a link.
  //
  // Claims: the hero and the section heads are the founder's words of 2026-11-24 (the hero no longer
  // says none ships today; custom tuning is offered where the general models are not enough). No
  // number on this page: the research doc holds the numbers with their sources. The advisory
  // boundary is stated in the theft contract and in §7 Limits; the hero's line went with the
  // founder's rewrite. The engagements the founder's own history covers are not named.
  // §1 (2026-11-24, founder ask; first by the founder's order of the same day: field, machines,
  // pipelines, theft, sites, manuals, with the founder's headings and leads): the macro view, one field from above, and the use case that only
  // reads at that scale: exception-based surveillance, the morning list (the five jobs, field-wide).
  import { siteHead, CSS_SHARED } from "#lib/head.js";
  import Footer from "#lib/ui/Footer.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  import UseCaseFigure from "#lib/ui/UseCaseFigure.svelte";
  import PipeFigure from "#lib/ui/PipeFigure.svelte";
  import CrossCheckFigure from "#lib/ui/CrossCheckFigure.svelte";
  import FieldFigure from "#lib/ui/FieldFigure.svelte";
  const head = siteHead({
    title: 'Oil and gas: specialized Wave models on site - RogerAI',
    desc: 'Specialized Wave models inside an oil and gas site, reading records the site already keeps: the whole field read together, machines and sensors explained, pipeline inspections compared, theft flagged, sites compared, manuals answered. Advisory, on site, never in the control loop.',
    og: { type: 'website', title: 'RogerAI - specialized Wave models on oil and gas sites', desc: 'The field, machines, pipelines, theft, sites, manuals. Read from the records a site already keeps, answered on site, never in the control loop.', url: 'https://rogerai.fm/research-oil-gas.html', image: 'https://rogerai.fm/og-research.png' },
    css: [...CSS_SHARED, 'research.css', 'oil-gas.css'],
    scripts: [{ runtime: true }, { page: "research-oil-gas" }],
  });
  const TUNING = "/research-custom-tuning.html";
</script>
<svelte:head>{@html head}</svelte:head>
  <Rail head='RogerAI &middot; Oil and gas' rev='Rev. 2026.11 &middot; LABS' />

  <Nav variant='marketing' home='/' promo={false} />

  <main id="top">
    <section class="research-hero">
      <div class="wrap">
        <a class="wf-back mono" href="/research-industry.html">&larr; Industrial deployment</a>
        <span class="hero__eyebrow">
          <OnAir />
          <span class="hero__callsign">OIL AND GAS</span>
        </span>
        <h1>RogerAI specialized Wave models on oil and gas sites.</h1>
        <p class="research-hero__lead">
          Our industrialized models and general workflows operate specifically on oil and gas
          machinery. Compressors, pumps, pipelines, custody meters and rigs already write records
          every day. A Wave model on a computer inside the site reads those records and hands a
          person things to act on. See our solutions in action and how they can transform your
          current fields today, on your own data and machines.
        </p>
        <p class="research-hero__lead">
          If our general industrialized models are not enough, we provide a custom tuning job on
          your own records.
          <a href={TUNING}>How custom tuning works &rarr;</a>
        </p>
        <UseCaseFigure fig={1} />
      </div>
    </section>

    <TocTuner sections={[
      { id: "field", no: "§1", name: "FIELD" },
      { id: "machines", no: "§2", name: "MACHINES" },
      { id: "pipelines", no: "§3", name: "PIPELINES" },
      { id: "theft", no: "§4", name: "THEFT" },
      { id: "sites", no: "§5", name: "SITES" },
      { id: "manuals", no: "§6", name: "MANUALS" },
      { id: "limits", no: "§7", name: "LIMITS" },
    ]} />

    <section class="section" id="field">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;1 / THE FIELD</span>
          <h2>Field management and observability</h2>
          <p>A view of your whole field: every site you instrument with our models is read
            together. You can manage each site independently on the edge and identify issues or
            optimizations from how different sites perform. Our models can list work and repair
            orders based on your own rules.</p>
        </div>
        <FieldFigure fig={2} />
        <div class="device-contract">
          <p><b>What goes in.</b> Pump cards and run times, tank gauges and haul tickets, each
            station&rsquo;s alarm journal, and the platforms&rsquo; daily reports.</p>
          <p><b>What comes out.</b> A short list of stops, worst first, each with its reason and
            the records behind it. The jobs below, read across the whole field at once.</p>
          <p><b>Who acts.</b> The lease operator plans the day&rsquo;s route from it, and the
            foreman moves a crew. Nothing on the field changes on its own.</p>
          <p><b>What we ask you to send.</b> A season of records from every site, with the visits
            that found something marked, so the report can say how often the list would have put
            them first.</p>
        </div>
        <p class="sec-more mono"><a href="{TUNING}#files">The report you read before you use it &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="machines">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;2 / MACHINES</span>
          <h2>Our models read your machines and sensors</h2>
          <p>You keep your detectors as they are. We plug into them and provide other visual
            capabilities or sensors as needed. Our models know OEM manuals and trends and can be
            fine-tuned to your specific needs.</p>
        </div>
        <div class="device-contract">
          <p><b>What goes in.</b> Pressure, vibration, temperature and flow trends from the
            historian. The alarm journal. The OEM manual and the maintenance history.</p>
          <p><b>What comes out.</b> A flag with its reason and the manual page behind it. "Not
            sure" when the records do not say. Quiet through startups and maintenance: it takes
            the machine&rsquo;s mode as an input.</p>
          <p><b>Who acts.</b> The maintenance engineer, on the next shift, with the page open.
            Nothing changes on the machine.</p>
          <p><b>What we ask you to send.</b> A few months of trends, the alarm export, and the
            manuals as the PDFs you already have.</p>
        </div>
        <p class="sec-more mono"><a href="{TUNING}#how">Trained on your records, in minutes &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="pipelines">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;3 / PIPELINES</span>
          <h2>Compare previous inspections to current ones</h2>
          <p>Every inspection is a record. Our models identify progressive changes between
            inspections to find possible maintenance drift or risk. Today we work with PIG
            (pipeline inspection gauge) data, lidar, sonar, and drone images and video to rank
            possible problems.</p>
        </div>
        <PipeFigure fig={3} />
        <div class="device-contract">
          <p><b>What goes in.</b> The anomaly tally from two inspection runs. Thickness grids
            from outside the pipe. The thermal frames a drone flagged.</p>
          <p><b>What comes out.</b> A ranked list of what grew, worst first, with the reason and
            the spot on the line.</p>
          <p><b>What we advise on.</b> Which instrument fits the line, inside or outside the
            pipe, and which record to send. We do not sell the robot.</p>
          <p><b>Who acts.</b> The integrity engineer decides what to dig, and the vendor's
            analyst still signs the inspection.</p>
        </div>
        <p class="sec-more mono"><a href="/research-hardware.html">The computers it runs on &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="theft">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;4 / THEFT AND SECURITY</span>
          <h2>Theft and security posture</h2>
          <p>You know your numbers. With your metrics and visual security, our models can read
            small differences and alert you to possible theft or tampering.</p>
        </div>
        <CrossCheckFigure fig={4} />
        <div class="device-contract">
          <p><b>What goes in.</b> Head and tail flow, tank levels, the custody reconciliation,
            the leak system&rsquo;s event log, what a camera flagged, and the timestamps.</p>
          <p><b>What comes out.</b> A ranked flag with the rule it broke and where on the line.
            Repeat spots, shifts and nearby digging named.</p>
          <p><b>Who acts.</b> A person goes and looks. A flag is a question, never an accusation,
            and never a change to a bill.</p>
          <p><b>What we ask you to send.</b> A season of the four records, with the taps you
            already found marked, so the report can say how many it would have caught.</p>
        </div>
        <p class="sec-more mono"><a href="{TUNING}#files">The report you read before you use it &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="sites">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;5 / SITES</span>
          <h2>Site comparisons</h2>
          <p>Two platforms or rigs with the same equipment perform differently. Without
            replicating an entire digital twin, our models can read both sites and give important
            insights.</p>
        </div>
        <div class="device-contract">
          <p><b>What goes in.</b> Daily drilling or production reports from both sites, a few
            live channels, and the historian&rsquo;s roll-up per platform.</p>
          <p><b>What comes out.</b> Downtime and flat time rig against rig, an uptime table with
            the normalisation written down, and the configuration lines that differ.</p>
          <p><b>Who acts.</b> The drilling superintendent or the production lead, who knows
            which difference is a choice and which is a problem.</p>
          <p><b>What we ask you to send.</b> The same months from both sites. The reports are
            the main thing; the live channels sharpen the connection times.</p>
        </div>
        <p class="sec-more mono"><a href="/research-wave-family.html">The Wave family, tier by tier &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="manuals">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;6 / MANUALS AND P&amp;IDS</span>
          <h2>Every industrial manual, including your own</h2>
          <p>A site keeps thousands of pages it cannot send to a cloud: OEM manuals, P&amp;IDs,
            procedures, work orders. The model answers from them, on site, and shows where.</p>
        </div>
        <div class="device-contract">
          <p><b>What goes in.</b> The documents as they are: PDFs, drawings, procedures, the
            last years of work orders.</p>
          <p><b>What comes out.</b> An answer with the page and the drawing it came from.
            "Not in the documents" when it is not.</p>
          <p><b>Who acts.</b> The operator or the technician who asked. The documents never leave
            the site.</p>
          <p><b>What we ask you to send.</b> Nothing leaves: this job runs on a machine inside
            the site, on your own copy of the documents.</p>
        </div>
        <p class="sec-more mono"><a href="/research-industry.html#envelope">Where the box sits in the plant network &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="limits">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;7 / LIMITS</span>
          <h2>What it never does.</h2>
        </div>
        <div class="device-contract">
          <p><b>Never in the control loop.</b> It reads and it explains. Nothing it writes moves a
            valve, a pump or a machine, and it is never part of a protection or safety path.</p>
          <p><b>Never settles a bill, never names a person.</b> A custody flag is a line for an
            engineer to check. An unvalidated output never changes a billed quantity.</p>
          <p><b>Never claims what it did not measure.</b> Every job comes with a report that says
            how often it was right on records it never saw, how early, and how often it cried
            wolf, before you rely on it.</p>
        </div>
        <p class="research-hero__lead">
          Have a site and a record set in mind?
          <a href="mailto:labs@rogerai.fm?subject=Oil%20and%20gas">Write to us</a>. A few lines
          is plenty. Designed in Orange County, California. Served on our global decentralized
          cloud, or on a machine inside your site.
        </p>
      </div>
    </section>
  </main>

  <Footer anchors='/' />
