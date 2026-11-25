<script>
  import TocTuner from "#lib/ui/TocTuner.svelte";
  // The Wave Spectrum - one family, pico to exa
  // Ported from web/src/research-wave-family.html (2026-10-29); the copy is the legacy copy.
  import { siteHead, CSS_SHARED, CSS_ACCOUNT } from "#lib/head.js";
  import { content } from "#lib/content.js";
  import Footer from "#lib/ui/Footer.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  import WaveMark from "#lib/ui/WaveMark.svelte";
  import { LADDER, SETTINGS, smallest } from "#lib/waveJobs.js";
  // §3's summary (2026-11-24): the job finder moved to the manual (Appendix B5); this page shows, per
  // slot, how many jobs it is the smallest slot for and two of them, from the same data.
  const JOBS = SETTINGS.flatMap((g) => g.jobs);
  const PER = LADDER.map((s) => ({ slot: s, jobs: JOBS.filter((j) => smallest(j.slot) === s) }));
  const MAX = Math.max(...PER.map((p) => p.jobs.length));
  const facts = content('research-wave-family.html');
  const head = siteHead({
    title: 'The Wave Spectrum - one family, pico to exa',
    desc: 'The Wave family on one SI scale - Pico to Exa, a board beside the machine to an exascale-class flagship - what each tier is for, how each is built, and the measurement that decides which tier can take which industrial job.',
    og: { type: 'article', title: 'The Wave Spectrum', desc: 'One family of industrial models on a single SI scale - what each tier is for, what it takes to run it, and which job each size is allowed to take.', url: 'https://rogerai.fm/research-wave-family.html', image: 'https://rogerai.fm/og-research.png' },
    css: [...CSS_SHARED, 'research.css', 'tier-table.css', 'wave-family.css'],
    scripts: [{ runtime: true }, { page: "research-wave-family" }],
  });
</script>
<svelte:head>{@html head}</svelte:head>
<!-- The Wave family field guide. Linked from research.html's hero action cluster.
     It lives on its own page on purpose: research.html carries a deliberate
     conciseness budget (main < 19500 bytes, enforced by research-page.test.mjs),
     and the family - the rail, the slot table, forty-seven jobs and the variation
     gates - does not fit inside it without cutting founder-approved content. 
     (the page's doc comment, from the legacy file; its sheets are the css: list above) -->
  <Rail head='RogerAI &middot; Field Guide' rev='Rev. 2026.07 &middot; LABS' />

  <Nav variant='marketing' home='/' current='research' />

  <main id="top">

    <section class="research-hero">
      <div class="wrap research-hero__layout">
        <div>
          <a class="wf-back mono" href="/research.html">&larr; RogerAI Labs</a>
          <span class="hero__eyebrow">
            <OnAir />
            <span class="hero__callsign">WAVE FAMILY &middot; FIELD GUIDE</span>
          </span>
          <!-- FOUNDER 2026-08-17: "remove this from the title: 'one family, pico to
               exa.'" The H1 is the name and nothing else now. The <title> and meta
               description keep the long form - they are doing SEO work and no one reads
               them as a headline - but og:title follows the H1, because that IS the
               headline on a shared card. -->
          <h1>The Wave Spectrum.</h1>
          <!-- Plain first paragraph (SITE-SIMPLIFICATION-AUDIT-2026-11-13, fixes 6 and 9). The
               H1 stays the name alone (founder, 2026-08-17); the lead says what the family is. -->
          <p class="research-hero__lead">
            Wave models read machine data (sensors, logs, telemetry) and run on a computer
            inside a site with no internet. We pretrain them from scratch on industrial machine
            and manufacturing data and observations, and fine-tune them on your own data. RogerAI
            builds these models to hold back rather than hallucinate, and to be safe for
            industrial use. We train different sizes for different types of jobs.
          </p>
        </div>

        <!-- The wave mark, lifted verbatim from the homepage social card: two S-curves
             on one centreline, the red tighter than the grey, with the live beacon exactly
             where they cross. It belongs on this page more than anywhere - the family is
             named for the wave, and the crossing node is the point the whole page is about.
             Geometry is the card's, unchanged, so the two read as one mark. Deliberately
             not the vector Ping mascot, which stays on the homepage and the research hub. -->
        <!-- ALIVE (js/wave-mark.js): the resting frame is this exact mark. Behind
             the wide wave sit fainter harmonics at 2x, 3x, 4x - the ladder as a
             spectrum of wavelengths - and every harmonic of one fundamental has a
             node at the centre, so all of them cross under the beacon, always. The
             waves are standing waves that breathe (never flat); a slow shimmer
             rides the live one; a ripple leaves the beacon every few seconds. Under
             prefers-reduced-motion nothing moves and this is the mark to the pixel. -->
        <WaveMark caption='The smallest model that can do the job, on the hardware you already have.' />
      </div>
    </section>

    <!-- §1 SCALE -->
    <TocTuner sections={[
      { id: "scale", no: "§1", name: "SCALE" },
      { id: "slots", no: "§2", name: "SLOTS" },
      { id: "jobs", no: "§3", name: "JOBS" },
      { id: "variations", no: "§4", name: "VARIATIONS" },
      { id: "infinite", no: "§5", name: "WAVE INFINITE" },
    ]} />

    <section class="section" id="scale">
      <div class="wrap">
        <div class="section__head section__head--left">
          <!-- FOUNDER 2026-08-17: the old head ("The gaps are real" + a paragraph about
               a logarithmic axis carrying information) described the CHART. The point is
               the family: each tier is cut for the job it does and the hardware that job
               happens on.
               FOUNDER 2026-08-17 (round 2): the closing sentence - "the rail puts every
               tier at the size its artifact really is, so the distance between two of
               them is a real distance" - was the chart congratulating itself. Cut. The
               paragraph is about the family, so it ends on the family. -->
          <span class="sectionno">&sect;1 / SCALE</span>
          <h2>Sized for where it has to live.</h2>
          <p>A tier is not a rank. Each one is cut for the job it does and the hardware
             that job happens on: Wave Pico fits a single-board computer bolted next to
             one machine, Wave Exa fills a datacenter.</p>
        </div>

        <figure class="wf-rail" aria-labelledby="wf-rail-title">
          <div class="wf-rail__head">
            <span id="wf-rail-title">ARTIFACT SIZE ON DISK</span>
            <span>quantized weights, log scale</span>
            <!-- FOUNDER 2026-08-17 (round 2): the filled/hollow legend was the first
                 paragraph of the caption. It is furniture, not prose, so it moved into
                 the figure's own head where a legend belongs - beside the axis it
                 describes, read before the marks rather than after them. -->
            <span class="wf-rail__key">
              <span class="wf-rail__keyitem"><i class="is-live"></i>trained artifact</span>
              <span class="wf-rail__keyitem"><i></i>program stage</span>
            </span>
          </div>
          <!-- Four decades, 100 MB to 1 TB, so every tier sits ON the axis with
               room to breathe (founder 2026-08-17: the top three markers crowded on
               the old 1 MB-1 TB span - the two empty decades below Pico were paying
               for nothing). Still a true log axis. Node positions are computed, not
               eyeballed: at% = 25 x (log10(size_MB) - 2), size = params x ~0.6 bytes
               (Q4-class quantized weights, the shipping form - the measured 98M build
               is 65 MB at Q4_K_M). Tier midpoints from the locked ladder:
               Pico 275M -> 165 MB (5.4%), Nano 1.15B -> 690 MB (21%),
               Micro 7.5B -> 4.5 GB (41%), Giga 31B -> 18.6 GB (57%),
               Tera 100B -> 60 GB (69.5%), Peta 175B -> 105 GB (75.5%),
               Exa 284B -> 170 GB (80.8%). On four decades the closest neighbours
               (Peta/Exa) sit 46px apart on the 860px track.
               CORRECTED 2026-08-30: that 46px is the spacing of the DOTS, and the
               conclusion drawn from it - "one row, no stagger" - did not hold, because
               a .wf-node LABEL is 58px wide. 58 > 46, so the labels overlapped even
               though the markers cleared: Peta/Exa by 12.4px and Tera/Peta by 6.4px,
               measured at both desktop and mobile (the track is its 860px min-width at
               both, so this was never a mobile-only fault). .wf-node--up existed to
               drop the middle of three exactly here and had been left applied to
               nothing, so the rule was dead and the collision live. Peta wears it
               again. rail-label-collision.test.mjs now measures the arithmetic so a
               future re-spacing cannot quietly re-break it.
               FOUNDER 2026-08-17 (round 2): the last tick was "1 TB". It is now the
               infinity symbol - 100 GB is the last decade any tier needs, and the end of
               the rail is where the numbers stop rather than another size to read against.
               That is the ONE position on this axis that is explicitly not a number, so
               it is where Wave Infinite goes. -->
          <div class="wf-rail__scroll">
            <div class="wf-rail__track">
              <span class="wf-tick" style="--at:0%"><i></i><b>100 MB</b></span>
              <span class="wf-tick" style="--at:25%"><i></i><b>1 GB</b></span>
              <span class="wf-tick" style="--at:50%"><i></i><b>10 GB</b></span>
              <span class="wf-tick" style="--at:75%"><i></i><b>100 GB</b></span>
              <span class="wf-tick wf-tick--end" style="--at:100%"><i></i><b>&infin;</b></span>

              <span class="wf-node is-live" style="--at:5.4%"><i></i><b>Pico</b></span>
              <span class="wf-node" style="--at:21%"><i></i><b>Nano</b></span>
              <span class="wf-node" style="--at:41%"><i></i><b>Micro</b></span>
              <span class="wf-node" style="--at:57%"><i></i><b>Giga</b></span>
              <span class="wf-node" style="--at:69.5%"><i></i><b>Tera</b></span>
              <span class="wf-node wf-node--up" style="--at:75.5%"><i></i><b>Peta</b></span>
              <span class="wf-node" style="--at:80.8%"><i></i><b>Exa</b></span>

              <!-- WAVE INFINITE, on the rail on the founder's own instruction
                   (2026-08-17): "replace the 1 TB with an infinite sign and add it with
                   it's own dot ... and remove the text where it says it's not there."
                   That reverses the 2026-08-14 removal, and wave-infinite.test.mjs is
                   amended to match with the reason recorded there.
                   What keeps it honest is now structural rather than a disclaimer:
                   it is NOT one of the seven sized markers (its own class, outside that
                   set, so the ladder assertion still reads exactly Pico..Exa in order),
                   it sits at the infinity tick - the one point on this axis that is not
                   a finite size - and its marker is HOLLOW, which under this figure's
                   own legend means a program stage rather than a trained artifact.
                   Pico is still the only filled marker on the rail.
                   It comes after all seven nodes so nothing can walk between a node and
                   its own label. -->
              <span class="wf-beyond" style="--at:100%">
                <i></i><b>Wave Infinite</b>
              </span>
            </div>
          </div>
          <!-- FOUNDER 2026-08-17 (round 2): the caption was four paragraphs of
               provenance - legend, Pico's waypoint, why the top three crowd, Infinite.
               A reader at a SIZE rail wants to know what a size means in practice, so
               that is now the caption's job and the provenance moved into the figure's
               own furniture: the legend is in the head, and these two notes sit under
               the track. The "top three really are that close" line is gone - it was a
               third sentence about the chart, and the 60 GB-and-up band below says the
               same thing usefully. Every other fact is still here, one click of the eye
               from the marker it belongs to.
               The Infinite note stays INSIDE this figure on purpose: the lock in
               wave-infinite.test.mjs reads the whole figure for "not a size", and the
               fact belongs beside the axis it is excluded from. -->
          <div class="wf-rail__notes">
            <p class="wf-rail__note"><b>Why the dots sit where they do.</b> A tier&rsquo;s
              size is set by the machine that has to run it &ndash; a Raspberry Pi at one
              end, a regional cloud at the other &ndash; not by a benchmark score. Each tier
              is a band rather than one fixed count, so a later version can move within its
              stretch of the rail and still fit the hardware it was drawn for.</p>
            <!-- FOUNDER 2026-08-17: "remove the text where it says it's not there."
                 The disclaimer is gone. Then: "make it sound better and more informative
                 about what it will be" - the note was defining Wave Infinite by its own
                 marker, which tells a reader nothing. It now says what the programme is
                 for, in §5's own words, in the DESIGN register: "is designed to", never
                 the present tense, because the loop has not been run. "No fixed size"
                 explains why it sits at the ∞ end; it is not the disclaimer that came
                 off.
                 FOUNDER 2026-08-19: the closing "Hollow, like every program stage ..."
                 came off too. It was glossing the marker for a reader who can see the
                 legend six inches above it, and the honest status it carried is
                 structural rather than prose - the marker IS hollow (CSS), the legend
                 says what hollow means, and wave-infinite.test.mjs asserts both. -->
            <p class="wf-rail__note"><b>Wave Infinite sits at the &infin; end</b> because it
              has no fixed size. It is designed to start from a trained base &ndash; Nano,
              Micro or Giga &ndash; and grow new weights on your own work: real parameters,
              not prompts and not retrieval, with every round of growth having to clear a
              certificate that the base behaviour survived it
              (<a href="#infinite">&sect;5</a>).</p>
          </div>

          <figcaption id="wf-rail-cap" class="wf-rail__means">
            <p class="wf-rail__meanslead">What each size means once you have to run it:</p>
            <ol>
              <li><b>Under 1 GB</b> <span>Pico, Nano</span>
                <small>A Raspberry Pi-class board or a gateway box, no GPU, bolted beside
                  the machine. Targets about 50 ms.</small></li>
              <li><b>About 5 GB</b> <span>Micro</span>
                <small>One consumer GPU, or a server already standing in the plant room.
                  The size a workshop or a single site can own outright.</small></li>
              <li><b>About 20 GB</b> <span>Giga</span>
                <small>A large accelerator in the plant&rsquo;s own datacenter. Still one
                  box, still on your side of the fence.</small></li>
              <li><b>50 GB and up</b> <span>Tera, Peta, Exa</span>
                <small>More than one accelerator, with the model split across them: an
                  enterprise or regional cloud, run on the company&rsquo;s own estate.</small></li>
            </ol>
          </figcaption>
        </figure>
      </div>
    </section>

    <!-- §2 SLOTS -->
    <section class="section research-tone band band--inset" id="slots">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;2 / SLOTS</span>
          <h2>What each size is for.</h2>
          <p>Each Wave model has a purpose, and a set of requirements for how it was built and why.</p>
        </div>

        <div class="tier-scroll">
          <table class="data-table tier-matrix">
            <thead>
              <tr>
                <th scope="col">Slot</th><th scope="col">Class</th><th scope="col">Job</th>
                <th scope="col">Runs on</th><th scope="col">Recipe</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <th scope="row">Wave Pico</th><td class="tier-cell">250&ndash;300M</td>
                <td>The smallest: reads one machine&rsquo;s telemetry and says what is wrong, or that it is not sure</td>
                <td>Raspberry Pi &middot; no GPU &middot; ~50 ms target</td><td class="tier-cell">{@html facts.wave('pico.wave-family.recipe')}</td>
              </tr>
              <tr>
                <th scope="row">Wave Nano</th><td class="tier-cell">0.8&ndash;1.5B</td>
                <td>The fleet gateway: gathers many Pico readings, settles conflicts between them, passes the rest up</td>
                <td>Gateway, concentrator</td><td class="tier-cell">{@html facts.wave('nano.wave-family.recipe')}</td>
              </tr>
              <tr>
                <th scope="row">Wave Micro</th><td class="tier-cell">7&ndash;8B</td>
                <td>The site brain: multi-fleet reasoning across a facility, general-capable and industrial</td>
                <td>On-site server</td><td class="tier-cell">{@html facts.wave('micro.wave-family.recipe')}</td>
              </tr>
              <tr>
                <th scope="row">Wave Giga</th><td class="tier-cell">27&ndash;35B</td>
                <td>The plant: full-plant reasoning, competitive on general benchmarks as well as machines</td>
                <td>Plant datacenter</td><td class="tier-cell">Base + specialize</td>
              </tr>
              <tr>
                <th scope="row">Wave Tera</th><td class="tier-cell">80&ndash;120B</td>
                <td>Cross-site enterprise: correlates faults and trends across many plants at once</td>
                <td>Enterprise cloud</td><td class="tier-cell">Base + specialize</td>
              </tr>
              <tr>
                <th scope="row">Wave Peta</th><td class="tier-cell">150&ndash;200B</td>
                <td>Regional scale: a leaner giant, distilled and pruned down from the frontier</td>
                <td>Regional cloud</td><td class="tier-cell">Expert-pruned</td>
              </tr>
              <tr>
                <th scope="row">Wave Exa</th><td class="tier-cell">&gt;280B</td>
                <td>The flagship: exascale-class frontier capability, and the teacher the whole family learns from</td>
                <td>Frontier datacenter</td><td class="tier-cell">Frontier base</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- MODEL PAGES (2026-10-20): the two tiers with a page of their own, linked from
             their table rows above and here, so a reader at the ladder can open either.
             Giga joins when the lab supplies its facts. -->
        <div class="spectrum-more">
          <a href="/research-wave-pico.html">Wave Pico, the model page</a>
          <span>Live and free on the network; weights not downloadable, use it through the
            network. What it is for, how to run it, what it scores and where it stops.</span>
          <a href="/research-wave-nano.html">Wave Nano, the model page</a>
          <span>Two fire models on the network as specialized models. The gateway
            size, with its vision module.</span>
        </div>

        <!-- Founder 2026-08-18, "very subtly": say WHY every tier is a band and
             not a number - each release gets room to move inside its own band
             without changing what the tier is for. Deliberately no version
             numbers: naming a release here would turn a design note into a
             roadmap promise, which is a different kind of claim. -->
        <!-- a HARD rail (ROGER-EDGE-MCU-FEASIBILITY-2026-08-14): no Wave tier runs on an MCU, and the page says
             so where the "Runs on" column is; it stayed here when the other notes went to the manual (2026-11-24) -->
        <p class="wf-bandnote"><b>The microcontroller is the sensor, not the model.</b> An ESP32 or an Arduino reads the vibration, the temperature or the current and hands it up. It cannot also be the model: that board has a few megabytes of flash, and Wave Pico&rsquo;s 293M release is <b>270&ndash;562 MB</b> as GGUF files. Nothing closes that gap, so no Wave tier runs on one. What does live down on the node is <a href="/research-industry.html#devices">Roger Edge</a>, a classifier of a few hundred kilobytes; the smallest thing that runs a Wave tier is the Pi-class board above it, with no GPU.</p>
        <p class="sec-more mono"><a href="/manual.html#sFamily-built">How the family is built: specialists first, the recipes, the mesh &rarr;</a></p>
      </div>
    </section>

    <!-- §3 JOBS -->
    <section class="section" id="jobs">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;3 / JOBS</span>
          <h2>Which slot for which job.</h2>
          <p>The smallest slot that can carry the work. Small
             slots turn messy reality into structure, larger slots reason over that
             structure. These are the jobs the family is built for.</p>
        </div>

        <!-- THE JOB FINDER. Progressive enhancement, in that order:
             (1) With no JavaScript this is one grouped table, whole and readable, and the
                 toolbar below never appears (it ships `hidden`; js/wave-jobs.js is the only
                 thing that reveals it). Nothing is hidden behind a script.
             (2) With JavaScript it becomes a filter: by setting, by slot, with a live count
                 and a bounded scroll panel, so the section stays one screen tall however
                 many jobs it grows.
             (3) Under 760px every row becomes a card - the old 640px-wide table forced a
                 sideways scroll on a phone for all 27 rows.

             TWO ROW FORMS, on purpose. The 27 PLANT jobs keep the bare, attribute-free
             row form: research-page.test.mjs counts them out of this
             table and asserts they match, one for one, the job arcs drawn on the research
             hub's scope figure. Rows BEYOND the plant set carry `data-scope="beyond-plant"`,
             which keeps that count at the 27 the scope draws while letting this section
             cover the hobbyist bench and the top of the ladder. wave-family.test.mjs reads
             BOTH forms, so every row is still checked for a real slot, no duplicate, and no
             clinical work. -->
        <ol class="wf-jobsum" aria-label="Jobs by the smallest slot that can carry them">
          {#each PER as p}
            <li style="--n:{p.jobs.length / MAX}"><b>{p.slot}</b><i aria-hidden="true"></i><span>{p.jobs.length} jobs</span><small>{p.jobs.slice(0, 2).map((j) => j.name).join(" · ")}</small></li>
          {/each}
        </ol>
        <p class="sec-more mono"><a href="/manual.html#sFamily-jobs">All {JOBS.length} jobs, with the filters, in the manual &rarr;</a></p>
        <p class="wj__play"><b>Want to watch one run?</b> The
          <a href="/playbox.html#factory">factory game</a> in the Playbox puts three machines
          and a cookie line in front of you: keep them inside their bands and the line keeps
          moving. Same idea as these jobs, at speed.</p>
      </div>
    </section>

    <!-- §4 VARIATIONS -->
    <section class="section research-tone band band--inset" id="variations">
      <div class="wrap">
        <!-- FOUNDER 2026-08-17: the head was "Six variations, and the gate each one has
             to pass", over "a tier is a size, a variation is what that size was tuned to
             do" - internal vocabulary explaining internal vocabulary. It became "Four
             kinds of input, one family" plus a three-sentence scene and a four-cell
             strip, and the founder's verdict was "again, too convoluted ... don't gate
             yourself to four ... make it simpler to the point multi modal". So: the word
             MULTI-MODAL is said outright, no count is fixed anywhere (a fifth input must
             not make the heading wrong), the strip is gone because the table already
             carries what it was explaining, and the lead is two sentences. The gate
             survives as a clause rather than a definition - it is what keeps the section
             honest.
             FOUNDER 2026-08-17, again: "remove: 'None of this is available today...'".
             Gone, and not relocated - the founder wanted the sentence off, not moved. The
             not-yet-shipped truth now rests on two things inside the table's own frame:
             the GATE column, which is a list of measurements each capability still has to
             pass and therefore says by itself that none has, and the lead's clause
             "nothing ships until it passes its gate". The third column was "Lives at",
             which can be read as "exists at", so it is "Planned for" now - intent, in the
             header, with no caveat sentence added back. -->
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;4 / VARIATIONS</span>
          <h2>Multi-modal, because machines are.</h2>
          <p>A raw sensor signal, a technician&rsquo;s note, a photo of a gauge, a voice
             over the radio with a compressor running behind it - the family is built to
             read what a plant actually produces. Each tier is tuned per input, and
             nothing ships until it passes its gate.</p>
        </div>

        <div class="tier-scroll">
          <table class="data-table tier-matrix">
            <thead>
              <tr><th scope="col">Variation</th><th scope="col">What it does</th><th scope="col">Planned for</th><th scope="col">Gate</th></tr>
            </thead>
            <tbody>
              <tr><th scope="row">Tool</th><td>Typed function calls, JSON contracts, schema adherence</td><td class="tier-cell">Nano and up</td><td>Exact-JSON pass rate, no regression when quantized</td></tr>
              <tr><th scope="row">Text</th><td>Instruction following, summarizing, explanation</td><td class="tier-cell">Micro and up</td><td>Held-out language budget; explanation claimed at Giga</td></tr>
              <tr><th scope="row">Embed</th><td>Retrieval over manuals, procedures, drawings, work history</td><td class="tier-cell">Nano, Micro</td><td>Retrieval benchmark on industrial text</td></tr>
              <tr><th scope="row">Guard</th><td>Refusal, escalation, and saying it does not know</td><td class="tier-cell">Every slot</td><td>Abstention rate on questions it cannot answer</td></tr>
              <tr><th scope="row">Vision</th><td>Gauge and screen reading, weld and assembly inspection</td><td class="tier-cell">Micro to Giga</td><td>Named public benchmark, published spread</td></tr>
              <tr><th scope="row">Audio</th><td>Gloves-on voice control, radio and intercom transcription</td><td class="tier-cell">Pico to Giga</td><td>Word error rate measured in real plant noise</td></tr>
            </tbody>
          </table>
        </div>

        <div class="research-prose">
          <!-- FOUNDER 2026-08-17: this used to open "Section 2 says small slots will be
               wrong when pushed past extraction". Section 2 was the pruning experiment and
               it came off the page, so the sentence stands on its own now. The claim is
               unchanged and it is still the honest one: a specialist pushed past what it
               was built for is wrong, and Guard is the answer to that. -->
          <p><b>Guard is the one buyers ask about first.</b> A slot built to do one thing
             will be wrong if you push it past that thing. Guard is what turns being wrong
             from dangerous into safe, which is why it belongs at every slot rather
             than arriving late.</p>
          <!-- Adapters (features/web/sensor_positioning.feature, PROPOSED 2026-10-15): the
               status words are cap tokens from src/data/capabilities.json. No size and no
               score until a card is publishable. -->
          <!-- The adapter definition is one sentence, the same on every page that uses the
               word (fix 9): research-wave-pico, research-wave-nano, research-hardware,
               research-industry and here. -->
          <p class="adapter-note"><b>One base model, a small add-on per job.</b> Each tier is
             one base model that answers the way it always has. An adapter is a small add-on
             file, trained for one job, that plugs into a Wave model without changing it. It
             is switched on only for that job, and the base stays as it was for everything
             else. The first, fire early warning on Wave Pico,
             is <span data-cap="fire-warning" data-cap-stage="{facts.t.cap('fire-warning.stage')}"><span class="cap-stage cap-stage--inline">{@html facts.cap('fire-warning.label')}</span></span>.
             Spotting a reading drifting out of its normal band is built into the base, not
             an adapter; that skill, telemetry drift, is <span data-cap="telemetry-drift" data-cap-stage="{facts.t.cap('telemetry-drift.stage')}"><span class="cap-stage cap-stage--inline">{@html facts.cap('telemetry-drift.label')}</span></span>.</p>
        </div>

        <div class="research-actions research-actions--lead">
          <a class="research-button" href="/research.html#models">See model status</a>
          <a class="research-button" href="/broadcasts.html">Read the evidence</a>
        </div>
      </div>
    </section>

    <!-- §5 WAVE INFINITE. Canonical definition: WAVE-INFINITE-PROGRAM-2026-07-31.md,
         which is explicit that this is a prototype MODEL PROGRAMME and that "the
         certified runtime is one layer of it" - an earlier version of this section led
         with "a runtime, not a size" and was wrong.
         The hard public constraint, quoted from that doc's standing constraints: never
         claim "self-training" in public material until CURE's gates pass. Layer 3 is
         therefore described as preregistered and in development, never as a capability
         the model has today. Contract: features/web/wave_infinite.feature -->
    <section class="section" id="infinite">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;5 / WAVE INFINITE</span>
          <h2 class="wf-inf__title">The model that <span class="wf-inf__shimmer">grows</span>.</h2>
          <span class="wf-inf__badge">Prototype &middot; measured foundations, growth loop in development</span>
          <p>Every slot above is a fixed size. <b>Wave Infinite</b> is a prototype model
             program. Start from a <b>Nano</b>, <b>Micro</b>, or <b>Giga</b> base, and the
             model grows new weights: real parameters, trained on your work, not prompts
             and not retrieval. All of it happens inside a proof that the base behavior is
             preserved.</p>
        </div>

        <!-- The loop, drawn. Visuals are not claims (web brief), so the figure gets the
             showcase treatment; the badge above and the states below carry the truth.
             In the isometric hand (2026-11-23): one slab with the lemniscate cut into it as a
             groove, the comet and its echo riding it as before, three base blocks with a feed
             each into the entry, the crossing a raised disc with the one lamp, three sprout
             blocks past GROW that rise and sink, the exact path back as a dashed groove round
             the slab's back. Words are HTML over the stage in the old order (the research copy
             lock holds); the SVG between the iso markers is written by
             computer-scientist/article/fig_orbit_gen.py (--svelte); rules in wave-family.css,
             which maps the hand's paper and ink to the figure's own ground inside the stage. -->
        <figure class="wf-orbit" role="img"
          aria-label="The Wave Infinite loop: a Nano, Micro, or Giga base model enters a continuous cycle of observe, measure, grow, and certify. New capacity sprouts where measured utility says it helps, and every change keeps an exact path back to the original model.">
          <div class="iso__stage wf-orbit__stage" aria-hidden="true">
            <div class="wf-orbit__pic">
              <span class="wf-orbit__word wf-orbit__feedlabel wf-orbit__feedlabel--nano iso__ko">NANO</span>
              <span class="wf-orbit__word wf-orbit__feedlabel wf-orbit__feedlabel--micro iso__ko">MICRO</span>
              <span class="wf-orbit__word wf-orbit__feedlabel wf-orbit__feedlabel--giga iso__ko">GIGA</span>
              <!-- iso:orbit -->
            <svg class="wf-orbit__svg" viewBox="-114.2 -6 361.7 136.6" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-6.2 2.8 L241.5 74.3 L183.5 124.6 L-64.2 53.1 Z"/>
<g class="wf-orbit__ground">
<path class="iso__face iso__face--s1" d="M-56 48.5 L179.6 116.5 L179.6 120 L-56 52 Z"/>
<path class="iso__face iso__face--s2" d="M235.6 68 L179.6 116.5 L179.6 120 L235.6 71.5 Z"/>
<path class="iso__face" d="M0 0 L235.6 68 L179.6 116.5 L-56 48.5 Z"/>
</g>
<g class="wf-orbit__floor">
<path class="wf-orbit__feed wf-orbit__feed--nano" d="M15.1 15.6 L17.5 16.4 L19.5 17.5 L21.1 18.8 L22.4 20.3 L23.5 21.9 L24.3 23.6 L25 25.4 L25.6 27.2 L26 29.1 L26.5 31 L26.9 32.8 L27.4 34.5 L28.1 36.2 L28.8 37.7 L29.8 39 L31 40.2 L32.5 41.2 L34.4 41.9"/>
<path class="wf-orbit__feed wf-orbit__feed--micro" d="M-0.3 31.9 L2.2 32.6 L4.6 33.3 L6.8 34 L9 34.6 L11 35.2 L12.9 35.7 L14.8 36.3 L16.6 36.8 L18.3 37.3 L20.1 37.8 L21.8 38.3 L23.4 38.8 L25.2 39.2 L26.9 39.7 L28.7 40.3 L30.5 40.8 L32.4 41.3 L34.4 41.9"/>
<path class="wf-orbit__feed wf-orbit__feed--giga" d="M-15.7 48.2 L-13 48.8 L-10.3 49.1 L-7.4 49.1 L-4.5 48.9 L-1.5 48.4 L1.5 47.8 L4.6 47.1 L7.6 46.3 L10.6 45.4 L13.6 44.6 L16.6 43.8 L19.5 43 L22.3 42.3 L24.9 41.8 L27.5 41.5 L30 41.3 L32.2 41.5 L34.4 41.9"/>
<path class="wf-orbit__back" d="M169.5 80.9 L192 87.4 L214 68.3 L56.4 22.8 L34.4 41.9"/>
<path class="wf-orbit__path" d="M34.4 41.9 L36.6 40.2 L39.1 38.8 L41.9 37.6 L44.8 36.8 L48 36.1 L51.2 35.7 L54.6 35.6 L58.1 35.7 L61.6 36 L65.2 36.5 L68.7 37.2 L72.3 38.1 L75.7 39.2 L79.1 40.5 L82.4 42 L85.5 43.6 L88.5 45.4 L91.2 47.3 L93.7 49.3 L96 51.5 L98 53.8 L99.6 56.3 L101 58.8 L101.9 61.4 L102.8 64 L104.2 66.6 L105.8 69 L107.8 71.3 L110.1 73.5 L112.6 75.5 L115.3 77.4 L118.3 79.2 L121.4 80.8 L124.7 82.3 L128.1 83.6 L131.5 84.7 L135.1 85.6 L138.6 86.3 L142.2 86.8 L145.7 87.1 L149.2 87.2 L152.6 87.1 L155.8 86.7 L159 86.1 L161.9 85.2 L164.7 84 L167.2 82.6 L169.5 80.9 L171.3 79.1 L172.5 77.3 L173.1 75.5 L173.1 73.8 L172.7 72.1 L171.7 70.5 L170.3 69 L168.4 67.5 L166.1 66.1 L163.4 64.8 L160.4 63.7 L157 62.6 L153.4 61.6 L149.5 60.8 L145.3 60.1 L141 59.6 L136.4 59.2 L131.7 59 L126.9 58.9 L122 59 L117 59.3 L112 59.8 L106.9 60.5 L101.9 61.4 L96.9 62.3 L91.9 63 L86.8 63.5 L81.8 63.8 L76.9 63.9 L72.1 63.8 L67.4 63.6 L62.8 63.2 L58.5 62.7 L54.3 62 L50.4 61.2 L46.8 60.2 L43.4 59.1 L40.4 58 L37.7 56.7 L35.4 55.3 L33.6 53.8 L32.1 52.3 L31.1 50.7 L30.7 49 L30.7 47.3 L31.3 45.5 L32.5 43.7 Z" pathLength="100"/>
</g>
<path class="wf-orbit__comet" d="M34.4 41.9 L36.6 40.2 L39.1 38.8 L41.9 37.6 L44.8 36.8 L48 36.1 L51.2 35.7 L54.6 35.6 L58.1 35.7 L61.6 36 L65.2 36.5 L68.7 37.2 L72.3 38.1 L75.7 39.2 L79.1 40.5 L82.4 42 L85.5 43.6 L88.5 45.4 L91.2 47.3 L93.7 49.3 L96 51.5 L98 53.8 L99.6 56.3 L101 58.8 L101.9 61.4 L102.8 64 L104.2 66.6 L105.8 69 L107.8 71.3 L110.1 73.5 L112.6 75.5 L115.3 77.4 L118.3 79.2 L121.4 80.8 L124.7 82.3 L128.1 83.6 L131.5 84.7 L135.1 85.6 L138.6 86.3 L142.2 86.8 L145.7 87.1 L149.2 87.2 L152.6 87.1 L155.8 86.7 L159 86.1 L161.9 85.2 L164.7 84 L167.2 82.6 L169.5 80.9 L171.3 79.1 L172.5 77.3 L173.1 75.5 L173.1 73.8 L172.7 72.1 L171.7 70.5 L170.3 69 L168.4 67.5 L166.1 66.1 L163.4 64.8 L160.4 63.7 L157 62.6 L153.4 61.6 L149.5 60.8 L145.3 60.1 L141 59.6 L136.4 59.2 L131.7 59 L126.9 58.9 L122 59 L117 59.3 L112 59.8 L106.9 60.5 L101.9 61.4 L96.9 62.3 L91.9 63 L86.8 63.5 L81.8 63.8 L76.9 63.9 L72.1 63.8 L67.4 63.6 L62.8 63.2 L58.5 62.7 L54.3 62 L50.4 61.2 L46.8 60.2 L43.4 59.1 L40.4 58 L37.7 56.7 L35.4 55.3 L33.6 53.8 L32.1 52.3 L31.1 50.7 L30.7 49 L30.7 47.3 L31.3 45.5 L32.5 43.7 Z" pathLength="100"/>
<path class="wf-orbit__comet wf-orbit__comet--echo" d="M34.4 41.9 L36.6 40.2 L39.1 38.8 L41.9 37.6 L44.8 36.8 L48 36.1 L51.2 35.7 L54.6 35.6 L58.1 35.7 L61.6 36 L65.2 36.5 L68.7 37.2 L72.3 38.1 L75.7 39.2 L79.1 40.5 L82.4 42 L85.5 43.6 L88.5 45.4 L91.2 47.3 L93.7 49.3 L96 51.5 L98 53.8 L99.6 56.3 L101 58.8 L101.9 61.4 L102.8 64 L104.2 66.6 L105.8 69 L107.8 71.3 L110.1 73.5 L112.6 75.5 L115.3 77.4 L118.3 79.2 L121.4 80.8 L124.7 82.3 L128.1 83.6 L131.5 84.7 L135.1 85.6 L138.6 86.3 L142.2 86.8 L145.7 87.1 L149.2 87.2 L152.6 87.1 L155.8 86.7 L159 86.1 L161.9 85.2 L164.7 84 L167.2 82.6 L169.5 80.9 L171.3 79.1 L172.5 77.3 L173.1 75.5 L173.1 73.8 L172.7 72.1 L171.7 70.5 L170.3 69 L168.4 67.5 L166.1 66.1 L163.4 64.8 L160.4 63.7 L157 62.6 L153.4 61.6 L149.5 60.8 L145.3 60.1 L141 59.6 L136.4 59.2 L131.7 59 L126.9 58.9 L122 59 L117 59.3 L112 59.8 L106.9 60.5 L101.9 61.4 L96.9 62.3 L91.9 63 L86.8 63.5 L81.8 63.8 L76.9 63.9 L72.1 63.8 L67.4 63.6 L62.8 63.2 L58.5 62.7 L54.3 62 L50.4 61.2 L46.8 60.2 L43.4 59.1 L40.4 58 L37.7 56.7 L35.4 55.3 L33.6 53.8 L32.1 52.3 L31.1 50.7 L30.7 49 L30.7 47.3 L31.3 45.5 L32.5 43.7 Z" pathLength="100"/>
<g class="iso__m wf-orbit__base wf-orbit__base--nano">
<path class="iso__halo iso__halo--part" d="M21.6 17.8 L19.9 18.9 L17.6 19.8 L14.9 20.3 L12.1 20.5 L9.2 20.3 L6.6 19.8 L4.3 18.9 L2.6 17.8 L1.5 16.5 L1.1 15 L1.5 13.6 L2.6 12.3 L4.3 11.1 L6.6 10.3 L9.2 9.7 L12.1 9.5 L14.9 9.7 L17.6 10.3 L19.9 11.1 L21.6 12.3 L22.7 13.6 L23.1 15 L22.7 16.5 Z"/>
<path class="iso__face iso__face--s1" d="M7.3 10.8 L13.4 12.6 L13.4 17.4 L7.3 15.7 Z"/>
<path class="iso__face iso__face--s2" d="M16.9 9.5 L13.4 12.6 L13.4 17.4 L16.9 14.4 Z"/>
<path class="iso__face" d="M10.8 7.8 L16.9 9.5 L13.4 12.6 L7.3 10.8 Z"/>
<path class="iso__ln iso__ln--faint" d="M8.6 12.6 L12.1 13.6"/>
</g>
<g class="iso__m wf-orbit__base wf-orbit__base--micro">
<path class="iso__halo iso__halo--part" d="M7.5 34.5 L5.3 35.9 L2.4 37.1 L-1 37.8 L-4.6 38 L-8.2 37.8 L-11.6 37.1 L-14.5 35.9 L-16.7 34.5 L-18.1 32.8 L-18.6 31 L-18.1 29.2 L-16.7 27.5 L-14.5 26 L-11.6 24.9 L-8.2 24.2 L-4.6 24 L-1 24.2 L2.4 24.9 L5.3 26 L7.5 27.5 L8.9 29.2 L9.4 31 L8.9 32.8 Z"/>
<path class="iso__face iso__face--s1" d="M-11.4 25 L-2.8 27.5 L-2.8 34.4 L-11.4 31.9 Z"/>
<path class="iso__face iso__face--s2" d="M2.2 23.2 L-2.8 27.5 L-2.8 34.4 L2.2 30.1 Z"/>
<path class="iso__face" d="M-6.4 20.7 L2.2 23.2 L-2.8 27.5 L-11.4 25 Z"/>
<path class="iso__ln iso__ln--faint" d="M-10.1 26.7 L-4.1 28.5"/>
</g>
<g class="iso__m wf-orbit__base wf-orbit__base--giga">
<path class="iso__halo iso__halo--part" d="M-6.6 51.2 L-9.3 53 L-12.8 54.3 L-16.9 55.2 L-21.3 55.5 L-25.7 55.2 L-29.8 54.3 L-33.3 53 L-36 51.2 L-37.7 49.2 L-38.3 47 L-37.7 44.8 L-36 42.7 L-33.3 41 L-29.8 39.6 L-25.7 38.8 L-21.3 38.5 L-16.9 38.8 L-12.8 39.6 L-9.3 41 L-6.6 42.7 L-4.9 44.8 L-4.3 47 L-4.9 49.2 Z"/>
<path class="iso__face iso__face--s1" d="M-30.2 39.1 L-18.9 42.4 L-18.9 51.4 L-30.2 48.2 Z"/>
<path class="iso__face iso__face--s2" d="M-12.4 36.8 L-18.9 42.4 L-18.9 51.4 L-12.4 45.8 Z"/>
<path class="iso__face" d="M-23.7 33.5 L-12.4 36.8 L-18.9 42.4 L-30.2 39.1 Z"/>
<path class="iso__ln iso__ln--faint" d="M-28.9 40.9 L-20.2 43.4"/>
</g>
<g class="wf-orbit__cross">
<path class="iso__face iso__face--s1" d="M97.4 60.4 L97.4 61.7 L97.5 62.1 L97.6 62.4 L97.9 62.8 L98.3 63.1 L98.7 63.3 L99.3 63.6 L99.9 63.8 L100.5 63.9 L101.2 64 L101.9 64 L102.6 64 L103.3 63.9 L103.9 63.8 L104.5 63.6 L105.1 63.3 L105.5 63.1 L105.9 62.8 L106.2 62.4 L106.3 62.1 L106.4 61.7 L106.4 60.4 Z"/>
<path class="iso__face" d="M105.8 61.5 L105.1 62 L104.2 62.3 L103.1 62.5 L101.9 62.6 L100.7 62.5 L99.7 62.3 L98.7 62 L98 61.5 L97.6 60.9 L97.4 60.4 L97.6 59.8 L98 59.2 L98.7 58.8 L99.7 58.4 L100.7 58.2 L101.9 58.1 L103.1 58.2 L104.2 58.4 L105.1 58.8 L105.8 59.2 L106.3 59.8 L106.4 60.4 L106.3 60.9 Z"/>
<circle class="iso__lamp wf-orbit__core" cx="101.9" cy="60.4" r="1.9"/>
</g>
<g class="iso__m wf-orbit__sprout wf-orbit__sprout--1">
<path class="iso__face iso__face--s1" d="M183.7 73.2 L188.9 74.7 L188.9 79.4 L183.7 77.9 Z"/>
<path class="iso__face iso__face--s2" d="M191.9 72.1 L188.9 74.7 L188.9 79.4 L191.9 76.8 Z"/>
<path class="iso__face" d="M186.7 70.6 L191.9 72.1 L188.9 74.7 L183.7 73.2 Z"/>
</g>
<g class="iso__m wf-orbit__sprout wf-orbit__sprout--2">
<path class="iso__face iso__face--s1" d="M182.4 83.7 L186.7 84.9 L186.7 88.8 L182.4 87.6 Z"/>
<path class="iso__face iso__face--s2" d="M189.2 82.8 L186.7 84.9 L186.7 88.8 L189.2 86.7 Z"/>
<path class="iso__face" d="M184.9 81.5 L189.2 82.8 L186.7 84.9 L182.4 83.7 Z"/>
</g>
<g class="iso__m wf-orbit__sprout wf-orbit__sprout--3">
<path class="iso__face iso__face--s1" d="M168.1 89.3 L171.6 90.3 L171.6 93.4 L168.1 92.4 Z"/>
<path class="iso__face iso__face--s2" d="M173.6 88.6 L171.6 90.3 L171.6 93.4 L173.6 91.7 Z"/>
<path class="iso__face" d="M170.1 87.6 L173.6 88.6 L171.6 90.3 L168.1 89.3 Z"/>
</g>
            </svg>
            <!-- /iso:orbit -->
              <span class="wf-orbit__word wf-orbit__label wf-orbit__label--observe iso__ko">01 OBSERVE</span>
              <span class="wf-orbit__word wf-orbit__label wf-orbit__label--measure iso__ko">02 MEASURE</span>
              <span class="wf-orbit__word wf-orbit__label wf-orbit__label--grow iso__ko">03 GROW</span>
              <span class="wf-orbit__word wf-orbit__label wf-orbit__label--certify iso__ko">04 CERTIFY</span>
            </div>
          </div>
          <figcaption>One loop: observe &middot; measure &middot; grow &middot; certify.
            Any base enters it; an exact path back always remains.</figcaption>
        </figure>

        <!-- The hook is the measurement, not the architecture. This is the finding that
             makes the safety machinery non-optional rather than decorative. -->
        <figure class="wf-inf">
          <blockquote class="wf-inf__claim">
            Why every change carries a certificate: specialization damage can be
            <b>statistically invisible</b> on the very benchmark you would check it with,
            and far worse once the workload shifts.
            <span class="wf-inf__punch">You cannot monitor your way to safe specialization.</span>
          </blockquote>

          <ol class="wf-inf__loop">
            <li data-layer="reflection" data-state="built">
              <span class="wf-inf__step">01</span><b>Reflection</b>
              <span class="wf-inf__state">Built and self-verifying</span>
              <small>The model&rsquo;s own execution - which experts fire, which could
                fire, what the router compared - is captured as first-class data. The
                instrument that captures it checks itself against the model on every
                token.</small>
            </li>
            <li data-layer="certified" data-state="measured">
              <span class="wf-inf__step">02</span><b>Certified adaptation</b>
              <span class="wf-inf__state">Primitives measured</span>
              <small>Prove which parts a workload can never reach - the question a
                compiler asks about dead code - then keep what your work uses in fast
                memory. A guard watches for input leaving the certified region and falls
                back to the full, unmodified model.</small>
            </li>
            <li data-layer="growth" data-state="unrun">
              <span class="wf-inf__step">03</span><b>Self-directed growth</b>
              <span class="wf-inf__state">In development &middot; preregistered, unrun</span>
              <small>The prototype would grow new capacity only where the model itself
                measures that it helps, running both the original and the expanded path
                and learning from the difference. The expansion starts bit-identical to
                the original and keeps an exact fallback.</small>
            </li>
          </ol>

          <figcaption>Three layers, three states of proof. Each change starts provably
            behavior-preserving and keeps an exact path back.</figcaption>
        </figure>

        <div class="wf-inf__use">
          <span class="research-kicker">TRAINING IT IS USING IT</span>
          <ol>
            <li><b>Start.</b> Pick a base: Nano, Micro, or Giga. Infinite is the loop,
              not another tier.</li>
            <li><b>Work.</b> The model observes its own execution and grows new weights
              only where measured utility says they help.</li>
            <li><b>Certify.</b> Changes go live through the same certified path, with an
              exact rollback always resident. If the work shifts beyond what is certified,
              it reverts to the original path by itself.</li>
          </ol>
        </div>

        <div class="wf-inf__proof">
          <div class="wf-inf__proven">
            <span class="research-kicker">WHAT WE HAVE PROVEN</span>
            <dl>
              <div><dt>The certificate is sound</dt>
                <dd>Removing every expert it marked unreachable left the model&rsquo;s
                  routing <b>bit-identical</b> - not close, unchanged.</dd></div>
              <div><dt>Specialization is safe on stable work</dt>
                <dd>Within a single domain the guard <b>almost never fires</b>, so the
                  specialization holds for a stable workload.</dd></div>
              <div><dt>The guard is necessary</dt>
                <dd>The moment the workload changes domain, the same certificates would
                  miss a <b>large share</b> of decisions. The fallback is what makes
                  aggression safe.</dd></div>
              <div><dt>Standard pruning over-removes</dt>
                <dd>Frequency analysis nominated <b>an order of magnitude more</b> for
                  removal than could be proven safe, across three model architectures.</dd></div>
            </dl>
            <p class="wf-inf__provenance">Measured on interim hardware; figures publish
              after final-hardware re-verification.</p>
          </div>
          <div class="wf-inf__unproven">
            <span class="research-kicker">IN MEASUREMENT</span>
            <p>End-to-end speed of certified specialization, and the growth layer&rsquo;s
               quality gains. The prototype program publishes both with error bars.</p>
          </div>
        </div>

        <p class="wf-inf__requires"><b>What it requires.</b> The certificate is
          reachability certification of <b>mixture-of-experts</b> routing. It applies only
          to a model that has experts and a per-expert selection bias to read. That is a
          base-model choice rather than something the family has today. It does not reach
          the microcontroller-class edge at all: a classifier of a few hundred kilobytes
          is not a language model.</p>

      </div>
    </section>

  </main>

  <Footer anchors='/' />

  <!-- The spectrum mark is the default; js/wave-mark.js is the previous
       "classic" mark, kept and reachable with ?mark=classic for comparison.
       Both animate the SAME resting markup above, so with JS off - or under
       prefers-reduced-motion - the page shows the original static mark. -->
  <!-- §4's job finder. The section is whole without it: the script only reveals the
       toolbar and bounds the panel. -->
