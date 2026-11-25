<script>
  import TocTuner from "#lib/ui/TocTuner.svelte";
  // Custom tuning - RogerAI Labs (DRAFT for founder review, 2026-11-13; not linked from main until approved)
  // Every number here comes from the lab results matrix, R.CT1-R.CT10 (2026-11-12/13), run on two
  // public stand-in datasets. No customer has used this; nothing on this page is a customer result or a price.
  // 2026-11-14 (founder): show, don't tell. Each section is a heading, a line, a figure, and a strip of
  // one-line cards. The full account of every step lives in the manual (manual.html#sTune) and each
  // section links to it. The figures' scrubbed lists play on their own while in view (data-scrub-auto).
  // 2026-11-25 (founder): tuning is by engagement only, not self-serve. The page explains the process
  // and sends people to the contact form's tuning topic (data-lets-talk="tuning"); no public price.
  import { siteHead, CSS_SHARED } from "#lib/head.js";
  import Footer from "#lib/ui/Footer.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  import TuningFigure from "#lib/ui/TuningFigure.svelte";
  import PrivateBandFigure from "#lib/ui/PrivateBandFigure.svelte";
  import AddonShelfFigure from "#lib/ui/AddonShelfFigure.svelte";
  import ReportFigure from "#lib/ui/ReportFigure.svelte";
  import HandoffFigure from "#lib/ui/HandoffFigure.svelte";
  const head = siteHead({
    title: 'Train a Wave model on your records - RogerAI',
    desc: 'A research prototype: we train a Wave model on your records, send you a report that says how often it was right on records it never saw, and serve it privately on our cloud.',
    og: { type: 'website', title: 'RogerAI - train a Wave model on your records', desc: 'Your records teach a small add-on. You get a report you can check, then a private band on our cloud.', url: 'https://rogerai.fm/research-custom-tuning.html', image: 'https://rogerai.fm/og-research-custom-tuning.png' },
    css: [...CSS_SHARED, 'research.css', 'custom-tuning.css'],
    scripts: [{ runtime: true }, { page: "research-custom-tuning" }],
  });
  const MANUAL = "/manual.html#sTune";
</script>
<svelte:head>{@html head}</svelte:head>
  <Rail head='RogerAI &middot; Custom tuning' rev='Rev. 2026.11 &middot; LABS' />

  <Nav variant='marketing' home='/' promo={false} />

  <main id="top">
    <section class="research-hero">
      <div class="wrap">
        <div class="research-hero__layout research-hero__layout--film">
          <div>
            <a class="wf-back mono" href="/research.html">&larr; RogerAI Labs</a>
            <span class="hero__eyebrow">
              <OnAir />
              <span class="hero__callsign">CUSTOM TUNING &middot; BY ENGAGEMENT</span>
            </span>
            <h1>Train a Wave model on your records. Use it privately, from anywhere.</h1>
            <p class="research-hero__lead">
              We work with you, one engagement at a time. First we check whether a prompt would do.
              If tuning wins, your records teach a Wave model, you get a report you can check, and
              only you can reach the result. Designed in Orange County, California. Our cloud is
              global.
            </p>
          </div>
          <!-- THE FILM (2026-11-15): Ping plugs the records into the set, drawn in the approved
               sepia set and animated on our own GPUs (article/tune_loop_gen.py, Wan 2.2 from the
               approved still). Ships as a poster; film.js starts the loop on screen. Reduced
               motion and no-JS get the poster. -->
          <figure class="figure figure--plate film" role="group" aria-labelledby="hero-film-cap">
            <div class="film__screen">
              <video data-film data-film-src="assets/tuning/hero-loop.mp4" poster="assets/tuning/hero-loop.webp"
                     width="832" height="432" muted loop playsinline preload="none"
                     aria-label="A short drawn loop: Ping plugs your records into the radio set and a strip comes out."></video>
            </div>
            <figcaption class="fig" id="hero-film-cap">Your records go in. A report and a band of your own come out.</figcaption>
          </figure>
        </div>
        <TuningFigure fig={1} addon="34 MB" minutes="2 to 30" />
      </div>
    </section>

    <TocTuner sections={[
      { id: "how", no: "§1", name: "HOW" },
      { id: "fit", no: "§2", name: "FIT" },
      { id: "files", no: "§3", name: "FILES" },
      { id: "where", no: "§4", name: "PRIVATE" },
      { id: "measured", no: "§5", name: "MEASURED" },
      { id: "not-yet", no: "§6", name: "DATA" },
      { id: "talk", no: "§7", name: "TALK" },
    ]} />

    <section class="section" id="how">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">§1 / HOW AN ENGAGEMENT RUNS</span>
          <h2>Five steps, worked through with you.</h2>
          <p>Not a self-serve upload. We scope each job with you, and quote it before anything
            trains.</p>
        </div>
        <div class="device-contract">
          <p><b>1. You tell us the task.</b> Seven short questions in the contact form. A sample of
            records helps.</p>
          <p><b>2. We check the fit.</b> We read your records and try a prompt first. If a prompt
            is enough, we say so and nothing trains.</p>
          <p><b>3. We train an add-on.</b> A file of 14 to 34 MB. Two to thirty minutes on one GPU.</p>
          <p><b>4. You read the report.</b> The untuned model beside the tuned one, on records it
            never saw, against your bar.</p>
          <p><b>5. You use it, or you don&rsquo;t.</b> A private band on our cloud, or we delete
            your examples.</p>
        </div>
        <AddonShelfFigure fig={2} />
        <p class="sec-more mono"><a href="{MANUAL}-how">Every step, in the manual &rarr;</a></p>
      </div>
    </section>

    <section class="section band band--inset" id="fit">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">§2 / WHEN IT FITS</span>
          <h2>Tuning wins on a narrow job done often.</h2>
          <p>Newer models follow a good prompt well. We only train when the report can show the
            difference.</p>
        </div>
        <div class="device-contract">
          <p><b>A good fit.</b> One narrow task, done many times a day. Labelled examples, enough
            to hold some back. A bar you can check.</p>
          <p><b>A prompt is enough.</b> A format or a tone. Facts that change every week. A handful
            of examples, or no way to tell right from wrong.</p>
          <p><b>Why it is not self-serve.</b> Most of the work is the data and the bar, not the
            training run. That needs a conversation.</p>
        </div>
      </div>
    </section>

    <section class="section" id="files">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">§3 / THREE FILES</span>
          <h2>What you send, and what you get back.</h2>
        </div>
        <div class="device-contract">
          <p><b>Examples.</b> One per line: what a person wrote, and the answer you want. One
            command turns a spreadsheet into it. A file made for OpenAI or Fireworks runs
            unchanged.</p>
          <p><b>Rules.</b> The labels allowed, what to strip, questions that are not the job, and
            your bar.</p>
          <p><b>Report.</b> Your bar first, met or not met. Then every detail. It says how often
            it got a record right, not a training loss.</p>
        </div>
        <ReportFigure fig={3} />
        <p class="sec-more mono"><a href="{MANUAL}-files">The three files, in the manual &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="where">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">§4 / WHERE IT RUNS</span>
          <h2>It runs on our cloud. Only you can reach it.</h2>
          <p>Want Wave and your add-on on your own machines, with no network at all?
            <a href="mailto:labs@rogerai.fm?subject=Wave%20on%20site">Write to us</a>.</p>
        </div>
        <PrivateBandFigure fig={4} />
        <div class="device-contract">
          <p><b>A private band.</b> Hidden from the public list. Answers only to your key, through
            the same API as any model here.</p>
          <p><b>The model under it does not move.</b> Your add-on is tied to one Wave version. We
            never swap it, or its precision, under you.</p>
          <p><b>From anywhere.</b> Designed in Orange County, California. Our cloud is global, so
            the band answers your people wherever they work.</p>
        </div>
        <p class="sec-more mono">On site, or on another open model, with a licence written
          down. <a href="{MANUAL}-where">The details, in the manual &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="measured">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">§5 / WHAT WE MEASURED</span>
          <h2>What it did on two public datasets.</h2>
          <p>Excavator work orders and offshore incident reports, both public. Every number is on
            records the model never saw.</p>
        </div>
        <HandoffFigure fig={5} />
        <div class="device-contract">
          <p><b>Excavator work orders.</b> University of Western Australia, CC BY. 5,485 records,
            four codes each. Every code right: Wave Nano 77.9%, Wave Pico 66.4%, untuned 0%.</p>
          <p><b>Offshore incident reports.</b> US Bureau of Safety and Environmental Enforcement,
            public domain. 3,617 reports, seven yes or no flags. Every flag right: Wave Nano
            82.1%, Wave Pico 79.1%.</p>
          <p><b>Handing questions back.</b> With a few practice examples, 98 to 100% of off-topic
            requests went to a person. At most 1 in 369 real records went by mistake.</p>
        </div>
        <p class="sec-more mono"><a href="{MANUAL}-measured">Open models of the same size, how
          many records you need, and the recipe: in the manual &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="not-yet">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">§6 / YOUR DATA, AND HOW WE WORK</span>
          <h2>Your records, and how we work with them.</h2>
        </div>
        <div class="device-contract">
          <p><b>Your data.</b> Used for one job. Say forget and we delete it and write down what
            was deleted. Not yet: encryption per customer, a written retention period. Strip
            names before you send.</p>
          <p><b>Files from other services.</b> An OpenAI or Fireworks training file runs here
            unchanged. Tool calls and images are refused, and the check says why.</p>
          <p><b>By engagement.</b> No upload page and no public price. We scope the job with you
            and run it for you.</p>
        </div>
        <p class="sec-more mono"><a href="{MANUAL}-data">What happens to your records, in
          the manual &rarr;</a></p>
      </div>
    </section>

    <section class="section research-closing" id="talk">
      <div class="wrap">
        <span class="sectionno">§7 / TALK TO US</span>
        <h2>Tell us the task.</h2>
        <p>Seven short questions. We answer with whether tuning fits, before anything trains.</p>
        <div class="research-actions research-actions--lead">
          <a class="research-button" data-lets-talk="tuning" aria-expanded="false"
             href="mailto:labs@rogerai.fm?subject=Custom%20tuning%20engagement">Start an engagement</a>
          <a class="research-button" href="/research.html">Back to Labs</a>
        </div>
      </div>
    </section>
  </main>

  <Footer anchors='/' />
