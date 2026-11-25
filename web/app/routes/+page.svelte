<script>
  import TocTuner from "#lib/ui/TocTuner.svelte";
  // RogerAI - local AI models and inference infrastructure
  // Ported from web/src/index.html (2026-10-29); the copy is the legacy copy.
  import { siteHead, CSS_SHARED, CSS_ACCOUNT } from "#lib/head.js";
  import { content } from "#lib/content.js";
  import { ROGERAI_IS } from "#lib/copy.js";
  import AdapterDial from "#lib/ui/AdapterDial.svelte";
  import BandFigure from "#lib/ui/BandFigure.svelte";
  import Footer from "#lib/ui/Footer.svelte";
  import InstallBox from "#lib/ui/InstallBox.svelte";
  import MoneyFigure from "#lib/ui/MoneyFigure.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  import TubePing from "#lib/ui/TubePing.svelte";
  const facts = content('index.html');
  const head = siteHead({
    title: 'RogerAI - local AI models and inference infrastructure',
    desc: 'RogerAI is an American AI research and infrastructure company building open models for constrained hardware alongside an OpenAI-compatible network.',
    theme: "inline",
    og: "home",
    css: [...CSS_SHARED, 'home.css', 'money-figure.css', 'band-figure.css'],
    scripts: [{ runtime: true }, { page: "index" }],
  });
</script>
<svelte:head>{@html head}</svelte:head>
  <!-- ===== running head: this page is one bound document ===== -->
  <Rail head='RogerAI · Operating Manual' rev='Rev. 2026.06 · ((•))' orange />

  <!-- ============================ NAV ============================ -->
  <Nav variant='marketing' home='#top' />

  <!-- ============================ HERO ============================ -->
  <main id="top" data-anchor-hold data-lift>
  <!-- INK PANEL: the cover. -->
  <div class="tone-zone" data-tone="ink">
  <section class="hero">
    <div class="wrap hero__inner">

      <!-- .hero__head (eyebrow + headline) spans the full content width;
           under it, .hero__copy (the narrative + the Labs proof) sits beside
           .hero__tools (FIG.1 the install command, then Ping) once there's
           room (home.css). Narrow viewports stack all three in this order. -->
      <div class="hero__grid">
      <div class="hero__head">
      <span class="hero__eyebrow">
        <OnAir />
        <span class="hero__callsign">ROGERAI</span>
        <span class="hero__sep" aria-hidden="true">·</span>
        American AI research + infrastructure
      </span>

      <!-- Each line is a block span (was <br>) so the lines can stagger in
           (home.css, lift-line). The newline between spans keeps the word
           boundary in the accessible name and in copied text.
           The operative word is the shared cycle word (DESIGN-SYSTEM.md, "Cycle word"):
           it tunes decentralized -> specialized -> industrialized on the carrier's beat,
           each load starting on the next. The HTML ships "decentralized" (SEO, no JS,
           print, the hero text fixture); the slot reserves that word, the widest, and
           "Models." closes up on a shorter word by transform (hero__tail), so the line
           never reflows. FOR FOUNDER REVIEW. -->
      <!-- Third line: the founder's sentence of 2026-11-21 ("Sovereign inference ..."), in place of
           "On your hardware or theirs. Pay per token, often nothing." (2026-11-13). The two
           cycle-word lines are the founder's (ruled 2026-10-26) and stay. -->
      <h1 class="hero__title">
        <span class="hero__line"><span class="carrier">Tune in</span> to the world&rsquo;s</span>
        <span class="hero__line"><span class="hero__op"><span class="carrier cycle-word cycle-word--tune" data-cycle="decentralized|specialized|industrialized" data-cycle-reserve="decentralized" data-cycle-store="hero-word" data-cycle-fit=".hero__line">decentralized</span> <span class="hero__tail">Models</span></span><span class="hero__tail">.</span></span>
        <span class="hero__line hero__fast">Sovereign inference that runs on your hardware or someone else&rsquo;s through a privatized network.</span>
      </h1>
      </div>

      <div class="hero__copy">

      <!-- Show, don't tell (founder 2026-11-14): one short definition, then the product
           promise as three labels, not a paragraph. The words the positioning guard holds
           (one OpenAI-compatible local endpoint; routing, failover, metering; community and
           private hardware; signed receipts) are the labels. The long form is the manual. -->
      <p class="hero__sub">
        <span class="hero__sub-lead">{ROGERAI_IS}</span>
      </p>
      <ul class="hero__specs mono" aria-label="What you get">
        <li>One OpenAI-compatible local endpoint, on your own machine</li>
        <li>Routing, failover and metering across community and private hardware</li>
        <li>Signed receipts for what you paid <a href="/pricing.html#why">Read the receipts &rarr;</a></li>
      </ul>

      <!-- LABS PLATE (founder ruling 2026-10-16): the plate names the whole Wave
           Spectrum and the live stage of its two leading tiers. The stage words are
           not typed here: each wave-stage token renders that tier's `stage` from
           src/data/wave-status.json (stageWords), so when a tier moves in the data
           the plate moves with it. -->
      <a class="hero__proof" href="/research.html">
        <span>ROGERAI LABS</span>
        <b>THE WAVE SPECTRUM</b>
        <span>PICO {@html facts.waveStage('pico')} · NANO {@html facts.waveStage('nano')}</span>
      </a>
      </div>

      <!-- FIG.1 the install command (command-as-hero), then Ping (always-on
           concierge) right under it. Ping's mascot is ALWAYS on, with a
           scrolling LED banner to its right (idle it rolls flavor phrases;
           on a reply it scrolls the reply, like a station's now-playing
           display). Clicking the mascot OR the banner opens a small
           record-player chat (Ping concierge) for typing.
           edge-story.js drives #pingTag / [data-onair]. -->
      <div class="hero__tools">
      <!-- FIG.1 - THE hero object: the install command (command-as-hero). -->
      <div class="install" data-frame="panel">
        <span class="fig" aria-hidden="true">FIG. 1 - TUNE IN</span>
        <!-- What the command actually downloads. The page used to hand a stranger
             a curl line with no noun attached to it: "tune in to local models" says
             what happens, never what the binary IS. One line, stated plainly, and
             the same claim the closing section makes so the two read as one. -->
        <p class="install__lead">
          <code class="inline">roger</code> is one small program. It finds the models already
          on your machine and <b>connects you to everyone else&rsquo;s</b>.
        </p>
        <InstallBox id='installCmd'>curl <span class="tok">-fsSL</span> https://rogerai.fm/install.sh <span class="tok">| sh</span></InstallBox>
        <div class="install__meta">
          <span id="installNote">macOS · Linux - no account needed to browse</span>
          <a class="install__alt" href="https://github.com/rogerai-fm/roger/releases">or download a binary &rarr;</a>
        </div>
        <div class="install__meta">
          <span id="installNoteWin" data-alt-note>On Windows (PowerShell): <code class="inline">irm https://rogerai.fm/install.ps1 | iex</code></span>
        </div>
      </div>

      <span class="hero__ping" id="pingDock">
        <button class="hero__ping-trigger" id="pingTrigger" type="button"
                aria-haspopup="dialog" aria-expanded="false" aria-controls="pingDeck"
                aria-label="Talk to Ping, the RogerAI on-air concierge">
          <TubePing />
          <span class="hero__ping-tag mono" id="pingTag">standing by</span>
        </button>

        <!-- ALWAYS-ON scrolling LED banner, beside the mascot. Part of the mascot,
             not the popup. Clicking it opens the chat. Idle: flavor phrases roll;
             on a reply: the reply scrolls across. #pingTag (above) is the single
             status; this banner carries flavor + responses only. -->
        <button class="pingband" id="pingBanner" type="button"
                title="Open Ping concierge to type a message"
                aria-controls="pingDeck">
          <span class="pingband__grille" aria-hidden="true"></span>
          <span class="pingband__window">
            <span class="pingband__text mono" id="pingBannerText">tune in / share / earn / signed receipts only</span>
          </span>
        </button>

        <!-- The record-player / tape-deck chat surface (TYPING only now). Starts
             CLOSED + collapsed so it never covers the hero. Draggable by its header.
             Lazily wired on first open by ping-chat.js. -->
        <div class="pingdeck" id="pingDeck" role="dialog" aria-modal="false" aria-label="Ping, the RogerAI concierge" hidden>
          <div class="pingdeck__head" id="pingHead">
            <span class="pingdeck__onair mono"><i class="pingdeck__led" id="pingDeckLight" aria-hidden="true"></i>PING / CONCIERGE</span>
            <span class="pingdeck__grip mono" aria-hidden="true">::::</span>
            <button class="pingdeck__close" id="pingClose" type="button" aria-label="Close Ping">
              <svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"/></svg>
            </button>
          </div>

          <!-- the deck: twin reels (spin while transmitting), purely decorative -->
          <div class="pingdeck__deck" aria-hidden="true">
            <svg class="pingdeck__reel pingdeck__reel--l" viewBox="0 0 48 48" focusable="false">
              <circle cx="24" cy="24" r="21" fill="none" stroke="currentColor" stroke-width="1.2" opacity="0.5"/>
              <circle cx="24" cy="24" r="6" fill="none" stroke="currentColor" stroke-width="1.4"/>
              <g class="pingdeck__spokes" stroke="currentColor" stroke-width="1.4" stroke-linecap="round">
                <path d="M24 6 V13"/><path d="M24 35 V42"/><path d="M6 24 H13"/><path d="M35 24 H42"/>
                <path d="M11.5 11.5 L16 16"/><path d="M32 32 L36.5 36.5"/><path d="M36.5 11.5 L32 16"/><path d="M16 32 L11.5 36.5"/>
              </g>
            </svg>
            <span class="pingdeck__deckline mono">NOW TRANSMITTING / replies roll on the banner</span>
            <svg class="pingdeck__reel pingdeck__reel--r" viewBox="0 0 48 48" focusable="false">
              <circle cx="24" cy="24" r="21" fill="none" stroke="currentColor" stroke-width="1.2" opacity="0.5"/>
              <circle cx="24" cy="24" r="6" fill="none" stroke="currentColor" stroke-width="1.4"/>
              <g class="pingdeck__spokes" stroke="currentColor" stroke-width="1.4" stroke-linecap="round">
                <path d="M24 6 V13"/><path d="M24 35 V42"/><path d="M6 24 H13"/><path d="M35 24 H42"/>
                <path d="M11.5 11.5 L16 16"/><path d="M32 32 L36.5 36.5"/><path d="M36.5 11.5 L32 16"/><path d="M16 32 L11.5 36.5"/>
              </g>
            </svg>
          </div>

          <div class="pingdeck__log" id="pingLog" aria-live="polite" aria-atomic="false"></div>

          <form class="pingdeck__form" id="pingForm" autocomplete="off">
            <input class="pingdeck__input mono" id="pingInput" type="text" maxlength="500"
                   placeholder="ask Ping about tuning in..." aria-label="Message to Ping" />
            <button class="pingdeck__send" id="pingSend" type="submit" aria-label="Send">
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h13M13 6l6 6-6 6"/></svg>
            </button>
          </form>
        </div>
      </span>
      </div><!-- /hero__tools -->
      </div><!-- /hero__grid -->

      <!-- FIG.2 - the Roger Edge story reel. Four films (Story A, Story B, the
           FireDefense use case and The Network), muted + playing back to back forever
           in random order; edge-story.js swaps the <source> on `ended`, never
           repeating the one that just played, and randomizes which opens on each
           page load. The power-on/
           off (the picture unfolding from / collapsing to a bright center line,
           an old CRT set) is keyed to scroll via IntersectionObserver: it plays
           once when the reel enters the viewport and once when it leaves, either
           direction. Lazy: no video bytes fetch until the reel nears the
           viewport. edge-story.js also owns #pingTag / [data-onair], inherited
           from the FIG.2 this replaced. -->
      <figure class="reel" id="edgeReel" data-reveal data-reveal-delay="240" data-lift-skip>
        <span class="fig" aria-hidden="true">FIG. 2 - ROGER EDGE &middot; ON THE AIR</span>
        <p class="reel__bridge">
          <a href="/research-models.html#roger-edge">More on the Edge line &rarr;</a>
        </p>
      </figure>

    </div>

    <!-- EXPERIMENT: the screen itself breaks all the way out to the full
         content width (spine to viewport edge) - a plain sibling of .wrap,
         not a 100vw/-50vw bleed. That trick overshoots on this page because
         body is inset by the fixed spine rather than centered on the raw
         viewport (see .dialwrap in models.css, which hit the same bug and
         fixed it the same way: don't fight the offset, just don't nest
         inside a max-width container at all). -->
    <div class="reel-band" id="edgeBand">
      <div class="reel__screen" id="edgeScreen" role="button" tabindex="0"
           aria-pressed="true" title="Turn the Roger Edge set off">
        <video class="reel__video" id="edgeVideo" muted playsinline preload="none"
               poster="assets/edge/poster-a.webp"
               aria-label="Roger Edge brand films, Story A, Story B, the FireDefense use case and The Network, looping in random order">
          <!-- mp4/H.264 FIRST: a <video> picks the first <source> whose type it
               can play, in DOM order - not "best" by any quality/size metric.
               H.264 hardware-decodes almost everywhere; VP9 hardware decode is
               inconsistent across GPUs/drivers, and falls back to a notably
               heavier software decode when it's missing or the GPU is already
               busy (edge-story.js:tick-free by design, so a decode-side stutter
               here is coming from the codec choice, not from this file). -->
          <source type="video/mp4">
          <source type="video/webm">
        </video>
        <span class="reel__tag mono" id="edgeTag" aria-hidden="true">STORY A</span>
      </div>
      <!-- a DOM sibling of .reel__screen, not a child - .reel__screen already
           has role="button" for the power toggle, and a real <button> nested
           inside it would fire both handlers on every click (event bubbling).
           Positioned as an overlay on the screen's corner via CSS instead. -->
      <button class="reel__mute" id="edgeMute" type="button" aria-pressed="false" title="Turn the film's sound on">
        <svg class="reel__icon--muted" viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M4 9v6h4l5 5V4L8 9H4z"></path><line x1="17" y1="4" x2="23" y2="20"></line><line x1="23" y1="4" x2="17" y2="20"></line>
        </svg>
        <svg class="reel__icon--sound" viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
          <path d="M4 9v6h4l5 5V4L8 9H4z"></path><path d="M16.5 8.5a5 5 0 0 1 0 7"></path><path d="M19 6a9 9 0 0 1 0 12"></path>
        </svg>
        <span id="edgeMuteLabel">sound off</span>
      </button>
    </div>
  </section>
  </div><!-- /tone-zone -->



  <!-- The three doors are the founder's lines of 2026-11-21. The third door used to open the
       broadcasts ("Open Air Waves"); it now opens the industrial deployment page, where the
       use cases are. Under the doors, "why industrial": four reasons, each one a claim another
       page already makes (research-industry, research-wave-family, §4 below), in one line. The
       numbers are mono index marks; the row has no link of its own, the door above it is the link.
       Draw-in: each reason rises on first view through the shared scroll lift (main[data-lift]),
       staggered on the scroll position. -->
  <aside class="institution-strip" aria-label="RogerAI institutions">
    <div class="wrap institution-strip__grid">
      <a class="institution-strip__item" href="/company.html">
        <span class="institution-strip__label">RogerAI Network</span>
        <span class="institution-strip__copy">Use a model on our marketplace, or share yours and create specialized models.</span>
      </a>
      <a class="institution-strip__item" href="/research.html">
        <span class="institution-strip__label">RogerAI Labs</span>
        <span class="institution-strip__copy">Wave, small models for machines in industrial settings, built and measured here. Trainable on your own records.</span>
      </a>
      <a class="institution-strip__item" href="/research-industry.html">
        <span class="institution-strip__label">Made for manufacturing</span>
        <span class="institution-strip__copy">Usable anywhere. See our use cases in oil and gas, fire early warning and more.</span>
      </a>
    </div>
    <div class="wrap institution-strip__why" data-reveal data-reveal-delay="120">
      <span class="institution-strip__why-label">Why industrial</span>
      <ol class="institution-strip__reasons">
        <li><b>Resilient</b><span>Runs on site, offline if you want, on links that are denied, degraded or intermittent.</span></li>
        <li><b>Robust</b><span>Advisory beside the control system, never in it. It says when it is not sure and hands the rest to a person.</span></li>
        <li><b>Pretrained on machines</b><span>Wave models are trained from scratch on machine data: sensors, logs and telemetry.</span></li>
        <li><b>Yours to extend</b><span>Every use case is an adapter on the same base. We train one on your records today; building your own is next.</span></li>
      </ol>
    </div>
  </aside>

  <!-- THE TUNER: the page's table of contents as a tuning scale, in the page
       (nothing pinned). Point at or focus a station and the red needle swings
       to it and its name shows in the readout (CSS); tuner.js adds arrow-key
       roving and rests the needle on the section in view. Plain links
       without CSS or JS. -->
  <TocTuner sections={[
    { id: "demo", no: "§1", name: "OPERATING PROCEDURE" },
    { id: "market", no: "§2", name: "THE BAND" },
    { id: "company", no: "§3", name: "COMPANY" },
    { id: "sensor", no: "§4", name: "ON SITE" },
    { id: "what", no: "§5", name: "THE TUNE" },
    { id: "spec", no: "§6", name: "SPECIFICATIONS" },
    { id: "books", no: "§7", name: "THE INDUSTRIAL SERIES" },
    { id: "monetize", no: "§8", name: "MONETIZE" },
    { id: "operator", no: "§9", name: "PRIVACY-FIRST AIRWAVES" },
    { id: "go", no: "§10", name: "GO" },
  ]} />

  <!-- INK PANEL: the console + the band. -->
  <div class="tone-zone" data-tone="ink">
  <!-- ===================== TERMINAL DEMO (multi-demo player) ===================== -->
  <section class="demo" id="demo">
    <div class="wrap">
      <div class="demo__head" data-reveal>
        <span class="sectionno">§1 / OPERATING PROCEDURE</span>
        <h2>Use a model. Share a model.</h2>
        <p>One program does both. Use a model running on another machine, or put your own
          on the air at a price you set, free included.</p>
        <p class="demo__fine">Your tools get a local address and a key at
          <code class="inline">127.0.0.1</code>. Anything built for the OpenAI API works
          with it unchanged.</p>
      </div>


      <span class="fig" aria-hidden="true">FIG. 3 - THE CONSOLE · DEMO REPLAY</span>
      <div class="term" id="term" data-reveal data-reveal-delay="100" role="group"
           aria-label="An interactive replay of the roger command-line radio. Nine demos, in order: boot the dial and read the band; tune in and open a local OpenAI-compatible channel, with a live wallet debit and failover when a station drops; point your own tools at that endpoint with a base URL, a key and a model, from the OpenAI SDK or an editor like opencode, Cursor or Cline; let the built-in agent run a multi-step job on its own; share your model on air with a price and schedule; cash out your earnings. Three video tapes follow: using, hosting, and the Ping World screensaver.">
        <div class="term__bar">
          <span class="term__dots" aria-hidden="true"><i></i><i></i><i></i></span>
          <span class="term__title" id="termTitle">roger - the dial</span>
        </div>

        <!-- station-preset bar: switchable demos, current one lit red. Follows the roger
             arc: boot -> borrow (tune in) -> integrate (your tools) -> automate (agent)
             -> lend (share) -> get paid (payouts). -->
        <div class="term__presets" id="termPresets" role="group" aria-label="Choose a demo">
          <button class="term__preset is-on" type="button" data-demo="roger" aria-pressed="true">
            <span class="term__preset-led" aria-hidden="true"></span>roger
          </button>
          <button class="term__preset" type="button" data-demo="tunein" aria-pressed="false">
            <span class="term__preset-led" aria-hidden="true"></span>tune in
          </button>
          <button class="term__preset" type="button" data-demo="yourtools" aria-pressed="false">
            <span class="term__preset-led" aria-hidden="true"></span>your tools
          </button>
          <button class="term__preset" type="button" data-demo="agent" aria-pressed="false">
            <span class="term__preset-led" aria-hidden="true"></span>agent
          </button>
          <button class="term__preset" type="button" data-demo="share" aria-pressed="false">
            <span class="term__preset-led" aria-hidden="true"></span>share
          </button>
          <button class="term__preset" type="button" data-demo="payouts" aria-pressed="false">
            <span class="term__preset-led" aria-hidden="true"></span>payouts
          </button>
          <button class="term__preset" type="button" data-demo="using" aria-pressed="false">
            <span class="term__preset-led" aria-hidden="true"></span>using
          </button>
          <button class="term__preset" type="button" data-demo="hosting" aria-pressed="false">
            <span class="term__preset-led" aria-hidden="true"></span>hosting
          </button>
          <button class="term__preset" type="button" data-demo="ping" aria-pressed="false">
            <span class="term__preset-led" aria-hidden="true"></span>ping
          </button>
        </div>

        <pre class="term__screen" id="termScreen" aria-hidden="true"></pre>

        <!-- MEDIA tape: the same slot as the ASCII screen, shown when the "ping" (real
             roger --ping screensaver capture) preset is picked. muted + loop +
             plays-inline; the gif is the poster and the no-<video> fallback. Hidden
             until selected (terminal.js).
             LAZY: no eager poster= (browsers fetch a poster gif even under preload=none,
             and each is 2.6-3.7MB); the gif path is deferred to data-poster and terminal.js
             swaps it onto .poster on scroll-in / tape-open. The <img> fallbacks are
             loading="lazy" too, so first paint costs zero tape bytes. -->
        <div class="term__media" id="termMedia" hidden>
          <video class="term__video" id="termPing" muted loop playsinline preload="none"
                 data-poster="assets/ping-demo.gif" aria-hidden="true" hidden>
            <source src="assets/ping-demo.mp4" type="video/mp4">
            <img class="term__video-gif" src="assets/ping-demo.gif" alt="" loading="lazy">
          </video>
        </div>

        <!-- FILM tapes (2026-11-23): "using" and "hosting" are two scenes drawn in the site's
             isometric hand (the SVG between the iso markers is written by
             computer-scientist/article/fig_film_gen.py), played by the same transport as the
             ASCII tapes: each frame is an intertitle card or a step of the scene with a caption
             (terminal.js sets the step classes s1, s2, ... on the layer and the words of the card
             and the caption, the way the ASCII tapes' words live in the script). Rules:
             home.css .term__film. Hidden until a film is picked; the ping tape above stays a video. -->
        <div class="term__film" id="termFilm" hidden aria-hidden="true">
          <div class="term__film-stage" id="termFilmUsing" hidden>
            <div class="iso__stage fm__stage"><div class="fm__pic">
<!-- iso:film-using -->
<svg class="fm__svg" viewBox="-56.2 -13.5 286.4 126.9" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-6.2 2.8 L224.2 69.3 L180.2 107.4 L-50.2 40.9 Z"/>
<g class="fm__ground">
<path class="iso__face iso__face--s1" d="M-42 36.4 L176.2 99.4 L176.2 102.8 L-42 39.8 Z"/>
<path class="iso__face iso__face--s2" d="M218.2 63 L176.2 99.4 L176.2 102.8 L218.2 66.5 Z"/>
<path class="iso__face" d="M0 0 L218.2 63 L176.2 99.4 L-42 36.4 Z"/>
</g>
<g class="iso__m fm__you">
<path class="iso__halo" d="M20.4 31.3 L16.6 33.8 L11.7 35.7 L5.9 36.9 L-0.3 37.3 L-6.6 36.9 L-12.3 35.7 L-17.3 33.8 L-21.1 31.3 L-23.5 28.4 L-24.3 25.3 L-23.5 22.2 L-21.1 19.3 L-17.3 16.8 L-12.3 14.9 L-6.6 13.7 L-0.3 13.3 L5.9 13.7 L11.7 14.9 L16.6 16.8 L20.4 19.3 L22.8 22.2 L23.7 25.3 L22.8 28.4 Z"/>
<path class="iso__face iso__face--s1" d="M-20.6 23.9 L8.9 32.4 L8.9 34.3 L-20.6 25.8 Z"/>
<path class="iso__face iso__face--s2" d="M19.9 22.9 L8.9 32.4 L8.9 34.3 L19.9 24.8 Z"/>
<path class="iso__face" d="M-9.6 14.4 L19.9 22.9 L8.9 32.4 L-20.6 23.9 Z"/>
<path class="iso__ln iso__ln--faint" d="M-8.7 18.7 L12 24.7"/>
<path class="iso__ln iso__ln--faint" d="M-10.7 20.4 L10 26.4"/>
<path class="iso__ln iso__ln--faint" d="M-12.7 22.1 L8 28.1"/>
<path class="iso__face" d="M-9.6 14.4 L19.9 22.9 L23.4 1 L-6.1 -7.5 Z"/>
<path class="iso__face iso__face--s2" d="M19.9 22.9 L23.4 1 L23.8 1.1 L20.3 23.1 Z"/>
<path class="iso__face iso__face--s1" d="M-7.1 13 L18 20.3 L20.8 2.8 L-4.3 -4.5 Z"/>
<path class="iso__ln iso__ln--lo" d="M-2.2 7.8 L1.6 6.7 L-1.5 3.6"/>
<path class="iso__ln iso__ln--lo" d="M3.1 8.4 L10.1 10.4"/>
</g>
<g class="iso__m fm__prog">
<path class="iso__face iso__face--s1" d="M-30.5 27.5 L-20.1 30.5 L-20.1 33.4 L-30.5 30.4 Z"/>
<path class="iso__face iso__face--s2" d="M-16.1 27 L-20.1 30.5 L-20.1 33.4 L-16.1 30 Z"/>
<path class="iso__face" d="M-26.5 24 L-16.1 27 L-20.1 30.5 L-30.5 27.5 Z"/>
<path class="iso__ln iso__ln--faint" d="M-28.8 28.9 L-21.9 30.9"/>
<circle class="iso__lamp" cx="-21.3" cy="29" r="1.1"/>
</g>
<g class="iso__m fm__slip">
<path class="iso__face iso__face--s1" d="M12.8 34.1 L23.2 37.1 L23.2 37.9 L12.8 34.9 Z"/>
<path class="iso__face iso__face--s2" d="M27.2 33.6 L23.2 37.1 L23.2 37.9 L27.2 34.4 Z"/>
<path class="iso__face" d="M16.8 30.6 L27.2 33.6 L23.2 37.1 L12.8 34.1 Z"/>
<path class="iso__ln iso__ln--lo" d="M15.8 33.5 L22.7 35.5"/>
<path class="iso__ln iso__ln--lo" d="M16.8 32.7 L23.7 34.7"/>
</g>
<g class="iso__m fm__wallet">
<path class="iso__face iso__face--s1" d="M-17.5 35.7 L-0.1 40.7 L-0.1 44.1 L-17.5 39.1 Z"/>
<path class="iso__face iso__face--s2" d="M6.1 35.2 L-0.1 40.7 L-0.1 44.1 L6.1 38.7 Z"/>
<path class="iso__face" d="M-11.2 30.2 L6.1 35.2 L-0.1 40.7 L-17.5 35.7 Z"/>
<path class="iso__ln iso__ln--faint" d="M-15.8 38.2 L-1.8 42.2"/>
<path class="iso__face iso__face--s1" d="M-11.2 30.2 L6.1 35.2 L2 35 L-15.3 30 Z"/>
<path class="iso__face iso__face--s2" d="M6.1 35.2 L2 35 L2 35.5 L6.1 35.8 Z"/>
<path class="iso__face" d="M-1.2 38.9 L-1.6 39.2 L-2.2 39.3 L-2.9 39.4 L-3.6 39.2 L-4.1 39 L-4.4 38.7 L-4.4 38.4 L-4.2 38 L-3.7 37.8 L-3.1 37.6 L-2.5 37.6 L-1.8 37.7 L-1.3 38 L-1 38.3 L-0.9 38.6 Z"/>
<path class="iso__face iso__face--cav" d="M-2 38.7 L-2.4 38.8 L-2.8 38.9 L-3.2 38.7 L-3.4 38.5 L-3.3 38.3 L-3 38.1 L-2.5 38.1 L-2.1 38.2 L-1.9 38.4 Z"/>
</g>
<g class="iso__m fm__relay">
<path class="iso__halo" d="M95.8 53.1 L91.6 55.7 L86.3 57.8 L80 59.1 L73.3 59.6 L66.5 59.1 L60.3 57.8 L54.9 55.7 L50.7 53.1 L48.1 49.9 L47.3 46.6 L48.1 43.2 L50.7 40.1 L54.9 37.4 L60.3 35.3 L66.5 34 L73.3 33.6 L80 34 L86.3 35.3 L91.6 37.4 L95.8 40.1 L98.4 43.2 L99.3 46.6 L98.4 49.9 Z"/>
<path class="iso__face iso__face--s1" d="M52.7 38.6 L83.9 47.6 L83.9 55.4 L52.7 46.4 Z"/>
<path class="iso__face iso__face--s2" d="M93.9 38.9 L83.9 47.6 L83.9 55.4 L93.9 46.7 Z"/>
<path class="iso__face" d="M62.7 29.9 L93.9 38.9 L83.9 47.6 L52.7 38.6 Z"/>
<path class="iso__face iso__face--s1" d="M55.2 37.2 L82.9 45.2 L82.9 46.4 L55.2 38.4 Z"/>
<path class="iso__face iso__face--s2" d="M91.4 37.9 L82.9 45.2 L82.9 46.4 L91.4 39.1 Z"/>
<path class="iso__face" d="M63.7 29.9 L91.4 37.9 L82.9 45.2 L55.2 37.2 Z"/>
<path class="iso__face iso__face--cav" d="M57 45 L60.1 45.9 L60.1 43.7 L57 42.8 Z"/>
<path class="iso__face iso__face--cav" d="M62.2 46.5 L65.3 47.4 L65.3 45.2 L62.2 44.3 Z"/>
<path class="iso__face iso__face--cav" d="M67.4 48 L70.5 48.9 L70.5 46.7 L67.4 45.8 Z"/>
<path class="iso__ln iso__ln--faint" d="M56.1 41.1 L80.4 48.1"/>
</g>
<g class="fm__fan">
<path class="iso__way fm__way fm__way--wire" d="M14.4 29 L57.7 41.5"/>
<path class="iso__way fm__way fm__way--trunk" d="M88.9 50.5 L107.9 56"/>
<path class="iso__way fm__way fm__way--1" d="M107.9 56 L96.9 65.6 L116 71.1"/>
<path class="iso__way fm__way fm__way--2" d="M107.9 56 L110.9 53.4 L145.5 63.4"/>
<path class="iso__way fm__way fm__way--3" d="M107.9 56 L122.9 43 L173.1 57.5"/>
</g>
<path class="fm__comet fm__comet--out fm__route-1" d="M14.4 29 L57.7 41.5 L88.9 50.5 L107.9 56 L96.9 65.6 L116 71.1" pathLength="100"/>
<path class="fm__comet fm__comet--back fm__route-1" d="M116 71.1 L96.9 65.6 L107.9 56 L88.9 50.5 L57.7 41.5 L14.4 29" pathLength="100"/>
<path class="fm__comet fm__comet--out fm__route-2" d="M14.4 29 L57.7 41.5 L88.9 50.5 L107.9 56 L110.9 53.4 L145.5 63.4" pathLength="100"/>
<path class="fm__comet fm__comet--back fm__route-2" d="M145.5 63.4 L110.9 53.4 L107.9 56 L88.9 50.5 L57.7 41.5 L14.4 29" pathLength="100"/>
<path class="fm__comet fm__comet--out fm__route-3" d="M14.4 29 L57.7 41.5 L88.9 50.5 L107.9 56 L122.9 43 L173.1 57.5" pathLength="100"/>
<path class="fm__comet fm__comet--back fm__route-3" d="M173.1 57.5 L122.9 43 L107.9 56 L88.9 50.5 L57.7 41.5 L14.4 29" pathLength="100"/>
<g class="iso__m fm__station fm__station--3">
<path class="iso__halo" d="M198.2 65.3 L195.4 67.2 L191.7 68.6 L187.3 69.5 L182.7 69.8 L178 69.5 L173.7 68.6 L169.9 67.2 L167.1 65.3 L165.3 63.1 L164.7 60.8 L165.3 58.5 L167.1 56.3 L169.9 54.4 L173.7 53 L178 52.1 L182.7 51.8 L187.3 52.1 L191.7 53 L195.4 54.4 L198.2 56.3 L200 58.5 L200.7 60.8 L200 63.1 Z"/>
<path class="iso__face iso__face--s1" d="M167.6 60.2 L186.7 65.7 L186.7 68.3 L167.6 62.8 Z"/>
<path class="iso__face iso__face--s2" d="M197.7 56.2 L186.7 65.7 L186.7 68.3 L197.7 58.8 Z"/>
<path class="iso__face" d="M178.6 50.7 L197.7 56.2 L186.7 65.7 L167.6 60.2 Z"/>
<path class="iso__face iso__face--s1" d="M175.8 35.7 L184.5 38.2 L184.5 61.6 L175.8 59.1 Z"/>
<path class="iso__face iso__face--s2" d="M189.5 33.9 L184.5 38.2 L184.5 61.6 L189.5 57.3 Z"/>
<path class="iso__face" d="M180.8 31.4 L189.5 33.9 L184.5 38.2 L175.8 35.7 Z"/>
<path class="iso__face iso__face--s1" d="M174.5 34.5 L184.9 37.5 L184.9 38.9 L174.5 35.9 Z"/>
<path class="iso__face iso__face--s2" d="M190.9 32.3 L184.9 37.5 L184.9 38.9 L190.9 33.7 Z"/>
<path class="iso__face" d="M180.5 29.3 L190.9 32.3 L184.9 37.5 L174.5 34.5 Z"/>
<path class="iso__ln iso__ln--faint" d="M177.6 47.5 L182.8 49"/>
<path class="iso__ln iso__ln--faint" d="M177.6 44.9 L182.8 46.4"/>
<path class="iso__ln iso__ln--faint" d="M177.6 42.3 L182.8 43.8"/>
<path class="iso__ln iso__ln--faint" d="M177.6 39.7 L182.8 41.2"/>
<path class="iso__face iso__face--cav" d="M178.4 56.4 L181.9 57.4 L181.9 55.2 L178.4 54.2 Z"/>
<path class="iso__ln" d="M183.2 33 L183.2 22.3"/>
<circle class="iso__lamp fm__lamp" cx="183.2" cy="22.3" r="1.5"/>
</g>
<g class="iso__m fm__station fm__station--2">
<path class="iso__halo" d="M181.1 74.2 L177.2 76.7 L172.3 78.6 L166.5 79.8 L160.3 80.2 L154.1 79.8 L148.3 78.6 L143.3 76.7 L139.5 74.2 L137.1 71.3 L136.3 68.2 L137.1 65.1 L139.5 62.2 L143.3 59.7 L148.3 57.8 L154.1 56.6 L160.3 56.2 L166.5 56.6 L172.3 57.8 L177.2 59.7 L181.1 62.2 L183.4 65.1 L184.3 68.2 L183.4 71.3 Z"/>
<path class="iso__face iso__face--s1" d="M140.5 52.7 L170 61.2 L170 76.8 L140.5 68.3 Z"/>
<path class="iso__face iso__face--s2" d="M180 52.5 L170 61.2 L170 76.8 L180 68.1 Z"/>
<path class="iso__face" d="M150.5 44 L180 52.5 L170 61.2 L140.5 52.7 Z"/>
<path class="iso__face iso__face--s1" d="M144.1 50.8 L168.4 57.8 L168.4 59.6 L144.1 52.6 Z"/>
<path class="iso__face iso__face--s2" d="M176.4 50.9 L168.4 57.8 L168.4 59.6 L176.4 52.7 Z"/>
<path class="iso__face" d="M152.1 43.9 L176.4 50.9 L168.4 57.8 L144.1 50.8 Z"/>
<path class="iso__face iso__face--s2" d="M144 64.1 L154.4 67.1 L154.4 62.3 L144 59.3 Z"/>
<path class="iso__ln iso__ln--faint" d="M145.7 63.9 L145.7 60.5"/>
<path class="iso__ln iso__ln--faint" d="M147.5 64.4 L147.5 61"/>
<path class="iso__ln iso__ln--faint" d="M149.2 64.9 L149.2 61.5"/>
<path class="iso__ln iso__ln--faint" d="M150.9 65.4 L150.9 62"/>
<path class="iso__ln iso__ln--faint" d="M152.7 65.9 L152.7 62.5"/>
<path class="iso__face" d="M165 67.8 L164.8 66.8 L164.3 65.8 L163.6 65 L162.7 64.4 L161.7 64.1 L160.8 64.2 L160.1 64.6 L159.6 65.3 L159.4 66.2 L159.6 67.2 L160.1 68.1 L160.8 69 L161.7 69.5 L162.7 69.8 L163.6 69.8 L164.3 69.4 L164.8 68.7 Z"/>
<path class="iso__ln" d="M162.2 67 L163.9 65.5"/>
<path class="iso__face iso__face--cav" d="M144.9 58.3 L160.5 62.8 L160.5 60.7 L144.9 56.2 Z"/>
<path class="iso__ln iso__ln--faint" d="M145.7 67.5 L164.8 73"/>
<path class="iso__ln" d="M172.7 52.2 L172.7 42.6"/>
<circle class="iso__lamp" cx="172.7" cy="42.6" r="1.4"/>
<circle class="iso__lamp fm__lamp" cx="166.5" cy="62.4" r="1.9"/>
<path class="fm__stop" d="M146.6 67.4 L163.9 62"/>
<path class="fm__stop" d="M146.6 57 L163.9 72.4"/>
</g>
<g class="iso__m fm__station fm__station--1">
<path class="iso__halo" d="M151.5 81.8 L147.6 84.3 L142.7 86.2 L136.9 87.4 L130.7 87.8 L124.5 87.4 L118.7 86.2 L113.7 84.3 L109.9 81.8 L107.5 78.9 L106.7 75.8 L107.5 72.7 L109.9 69.8 L113.7 67.3 L118.7 65.4 L124.5 64.2 L130.7 63.8 L136.9 64.2 L142.7 65.4 L147.6 67.3 L151.5 69.8 L153.9 72.7 L154.7 75.8 L153.9 78.9 Z"/>
<path class="iso__face iso__face--s1" d="M111 60.3 L140.4 68.8 L140.4 84.4 L111 75.9 Z"/>
<path class="iso__face iso__face--s2" d="M150.4 60.2 L140.4 68.8 L140.4 84.4 L150.4 75.7 Z"/>
<path class="iso__face" d="M121 51.7 L150.4 60.2 L140.4 68.8 L111 60.3 Z"/>
<path class="iso__face iso__face--s1" d="M114.6 58.5 L138.8 65.5 L138.8 67.2 L114.6 60.2 Z"/>
<path class="iso__face iso__face--s2" d="M146.8 58.5 L138.8 65.5 L138.8 67.2 L146.8 60.3 Z"/>
<path class="iso__face" d="M122.6 51.5 L146.8 58.5 L138.8 65.5 L114.6 58.5 Z"/>
<path class="iso__face iso__face--s2" d="M114.4 71.7 L124.8 74.7 L124.8 69.9 L114.4 66.9 Z"/>
<path class="iso__ln iso__ln--faint" d="M116.2 71.5 L116.2 68.1"/>
<path class="iso__ln iso__ln--faint" d="M117.9 72 L117.9 68.6"/>
<path class="iso__ln iso__ln--faint" d="M119.6 72.5 L119.6 69.1"/>
<path class="iso__ln iso__ln--faint" d="M121.3 73 L121.3 69.6"/>
<path class="iso__ln iso__ln--faint" d="M123.1 73.5 L123.1 70.1"/>
<path class="iso__face" d="M135.4 75.4 L135.2 74.4 L134.7 73.4 L134 72.6 L133.1 72 L132.1 71.7 L131.2 71.8 L130.5 72.2 L130 72.9 L129.8 73.8 L130 74.8 L130.5 75.7 L131.2 76.6 L132.1 77.2 L133.1 77.4 L134 77.4 L134.7 77 L135.2 76.3 Z"/>
<path class="iso__ln" d="M132.6 74.6 L134.3 73.1"/>
<path class="iso__face iso__face--cav" d="M115.3 65.9 L130.9 70.4 L130.9 68.3 L115.3 63.8 Z"/>
<path class="iso__ln iso__ln--faint" d="M116.2 75.2 L135.2 80.7"/>
<path class="iso__ln" d="M143.1 59.8 L143.1 50.2"/>
<circle class="iso__lamp" cx="143.1" cy="50.2" r="1.4"/>
<circle class="iso__lamp fm__lamp" cx="136.9" cy="70" r="1.9"/>
<path class="fm__stop" d="M117 75.1 L134.3 69.7"/>
<path class="fm__stop" d="M117 64.7 L134.3 80.1"/>
</g>
</svg>
<!-- /iso:film-using -->
            </div></div>
          </div>
          <div class="term__film-stage" id="termFilmHosting" hidden>
            <div class="iso__stage fm__stage"><div class="fm__pic">
<!-- iso:film-hosting -->
<svg class="fm__svg" viewBox="-56.2 -6 251.7 109.4" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-6.2 2.8 L189.5 59.3 L145.5 97.4 L-50.2 40.9 Z"/>
<g class="fm__ground">
<path class="iso__face iso__face--s1" d="M-42 36.4 L141.6 89.4 L141.6 92.8 L-42 39.8 Z"/>
<path class="iso__face iso__face--s2" d="M183.6 53 L141.6 89.4 L141.6 92.8 L183.6 56.5 Z"/>
<path class="iso__face" d="M0 0 L183.6 53 L141.6 89.4 L-42 36.4 Z"/>
</g>
<g class="iso__m fm__host">
<path class="iso__halo" d="M51.6 40.3 L47.8 42.8 L42.8 44.7 L37 45.9 L30.8 46.3 L24.6 45.9 L18.8 44.7 L13.9 42.8 L10 40.3 L7.6 37.4 L6.8 34.3 L7.6 31.2 L10 28.3 L13.9 25.8 L18.8 23.9 L24.6 22.7 L30.8 22.3 L37 22.7 L42.8 23.9 L47.8 25.8 L51.6 28.3 L54 31.2 L54.8 34.3 L54 37.4 Z"/>
<path class="iso__face iso__face--s1" d="M11.1 18.8 L40.5 27.3 L40.5 42.9 L11.1 34.4 Z"/>
<path class="iso__face iso__face--s2" d="M50.5 18.6 L40.5 27.3 L40.5 42.9 L50.5 34.2 Z"/>
<path class="iso__face" d="M21.1 10.1 L50.5 18.6 L40.5 27.3 L11.1 18.8 Z"/>
<path class="iso__face iso__face--s1" d="M14.7 16.9 L39 23.9 L39 25.7 L14.7 18.7 Z"/>
<path class="iso__face iso__face--s2" d="M47 17 L39 23.9 L39 25.7 L47 18.7 Z"/>
<path class="iso__face" d="M22.7 10 L47 17 L39 23.9 L14.7 16.9 Z"/>
<path class="iso__face iso__face--s2" d="M14.6 30.2 L25 33.2 L25 28.4 L14.6 25.4 Z"/>
<path class="iso__ln iso__ln--faint" d="M16.3 30 L16.3 26.6"/>
<path class="iso__ln iso__ln--faint" d="M18 30.5 L18 27.1"/>
<path class="iso__ln iso__ln--faint" d="M19.8 31 L19.8 27.6"/>
<path class="iso__ln iso__ln--faint" d="M21.5 31.5 L21.5 28.1"/>
<path class="iso__ln iso__ln--faint" d="M23.2 32 L23.2 28.6"/>
<path class="iso__face" d="M35.5 33.9 L35.4 32.9 L34.9 31.9 L34.1 31.1 L33.2 30.5 L32.3 30.2 L31.4 30.3 L30.6 30.7 L30.2 31.4 L30 32.3 L30.2 33.3 L30.6 34.2 L31.4 35.1 L32.3 35.6 L33.2 35.9 L34.1 35.9 L34.9 35.4 L35.4 34.8 Z"/>
<path class="iso__face iso__face--cav" d="M15.4 24.4 L31 28.9 L31 26.8 L15.4 22.3 Z"/>
<path class="iso__ln iso__ln--faint" d="M16.3 33.6 L35.4 39.1"/>
<path class="iso__ln" d="M43.2 18.2 L43.2 8.7"/>
<circle class="iso__lamp fm__air" cx="43.2" cy="8.7" r="1.4"/>
<circle class="iso__lamp fm__lamp" cx="37.1" cy="28.5" r="1.9"/>
<g class="fm__needle">
<path class="iso__ln" d="M32.8 33.1 L34.5 31.6"/>
</g>
<g class="fm__cell fm__cell--1">
<path class="iso__face iso__face--s1" d="M18.4 15 L20.3 15.6 L20.3 16.3 L18.4 15.7 Z"/>
<path class="iso__face iso__face--s2" d="M21.8 14.3 L20.3 15.6 L20.3 16.3 L21.8 15 Z"/>
<path class="iso__face" d="M19.9 13.7 L21.8 14.3 L20.3 15.6 L18.4 15 Z"/>
</g>
<g class="fm__cell fm__cell--2">
<path class="iso__face iso__face--s1" d="M21 15.8 L22.9 16.3 L22.9 17 L21 16.5 Z"/>
<path class="iso__face iso__face--s2" d="M24.4 15 L22.9 16.3 L22.9 17 L24.4 15.7 Z"/>
<path class="iso__face" d="M22.5 14.5 L24.4 15 L22.9 16.3 L21 15.8 Z"/>
</g>
<g class="fm__cell fm__cell--3">
<path class="iso__face iso__face--s1" d="M23.6 16.5 L25.5 17.1 L25.5 17.8 L23.6 17.2 Z"/>
<path class="iso__face iso__face--s2" d="M27 15.8 L25.5 17.1 L25.5 17.8 L27 16.5 Z"/>
<path class="iso__face" d="M25.1 15.2 L27 15.8 L25.5 17.1 L23.6 16.5 Z"/>
</g>
<g class="fm__cell fm__cell--4">
<path class="iso__face iso__face--s1" d="M26.2 17.3 L28.1 17.8 L28.1 18.5 L26.2 18 Z"/>
<path class="iso__face iso__face--s2" d="M29.6 16.5 L28.1 17.8 L28.1 18.5 L29.6 17.2 Z"/>
<path class="iso__face" d="M27.7 16 L29.6 16.5 L28.1 17.8 L26.2 17.3 Z"/>
</g>
<g class="fm__cell fm__cell--5">
<path class="iso__face iso__face--s1" d="M28.8 18 L30.7 18.6 L30.7 19.3 L28.8 18.7 Z"/>
<path class="iso__face iso__face--s2" d="M32.2 17.3 L30.7 18.6 L30.7 19.3 L32.2 18 Z"/>
<path class="iso__face" d="M30.3 16.7 L32.2 17.3 L30.7 18.6 L28.8 18 Z"/>
</g>
<g class="fm__cell fm__cell--6">
<path class="iso__face iso__face--s1" d="M31.4 18.8 L33.3 19.3 L33.3 20 L31.4 19.5 Z"/>
<path class="iso__face iso__face--s2" d="M34.8 18 L33.3 19.3 L33.3 20 L34.8 18.7 Z"/>
<path class="iso__face" d="M32.9 17.5 L34.8 18 L33.3 19.3 L31.4 18.8 Z"/>
</g>
<g class="fm__cell fm__cell--7">
<path class="iso__face iso__face--s1" d="M34 19.5 L35.9 20.1 L35.9 20.8 L34 20.2 Z"/>
<path class="iso__face iso__face--s2" d="M37.4 18.8 L35.9 20.1 L35.9 20.8 L37.4 19.5 Z"/>
<path class="iso__face" d="M35.5 18.2 L37.4 18.8 L35.9 20.1 L34 19.5 Z"/>
</g>
<g class="fm__cell fm__cell--8">
<path class="iso__face iso__face--s1" d="M36.6 20.3 L38.5 20.8 L38.5 21.5 L36.6 21 Z"/>
<path class="iso__face iso__face--s2" d="M40 19.5 L38.5 20.8 L38.5 21.5 L40 20.2 Z"/>
<path class="iso__face" d="M38.1 19 L40 19.5 L38.5 20.8 L36.6 20.3 Z"/>
</g>
</g>
<g class="iso__m fm__model">
<path class="iso__face iso__face--s1" d="M53 37.9 L63.4 40.9 L63.4 43.1 L53 40.1 Z"/>
<path class="iso__face iso__face--s2" d="M66.9 37.8 L63.4 40.9 L63.4 43.1 L66.9 40.1 Z"/>
<path class="iso__face" d="M56.5 34.8 L66.9 37.8 L63.4 40.9 L53 37.9 Z"/>
<path class="iso__ln iso__ln--faint" d="M54.3 39.6 L62.1 41.9"/>
<circle class="iso__lamp" cx="61.6" cy="41.3" r="1.3"/>
</g>
<g class="iso__m fm__prog">
<path class="iso__face iso__face--s1" d="M-5.5 27.8 L4.9 30.8 L4.9 33.7 L-5.5 30.7 Z"/>
<path class="iso__face iso__face--s2" d="M8.9 27.3 L4.9 30.8 L4.9 33.7 L8.9 30.3 Z"/>
<path class="iso__face" d="M-1.5 24.3 L8.9 27.3 L4.9 30.8 L-5.5 27.8 Z"/>
<path class="iso__ln iso__ln--faint" d="M-3.8 29.2 L3.2 31.2"/>
<circle class="iso__lamp" cx="3.7" cy="29.3" r="1.1"/>
</g>
<g class="iso__m fm__slip fm__slip--1">
<path class="iso__face iso__face--s1" d="M1.4 40 L11.8 43 L11.8 43.8 L1.4 40.8 Z"/>
<path class="iso__face iso__face--s2" d="M15.8 39.6 L11.8 43 L11.8 43.8 L15.8 40.3 Z"/>
<path class="iso__face" d="M5.4 36.6 L15.8 39.6 L11.8 43 L1.4 40 Z"/>
<path class="iso__ln iso__ln--lo" d="M4.4 39.4 L11.3 41.4"/>
<path class="iso__ln iso__ln--lo" d="M5.4 38.6 L12.3 40.6"/>
</g>
<g class="iso__m fm__slip fm__slip--2">
<path class="iso__face iso__face--s1" d="M4.6 40.3 L15 43.3 L15 44 L4.6 41 Z"/>
<path class="iso__face iso__face--s2" d="M19 39.8 L15 43.3 L15 44 L19 40.6 Z"/>
<path class="iso__face" d="M8.6 36.8 L19 39.8 L15 43.3 L4.6 40.3 Z"/>
<path class="iso__ln iso__ln--lo" d="M7.6 39.7 L14.5 41.7"/>
<path class="iso__ln iso__ln--lo" d="M8.6 38.8 L15.5 40.8"/>
</g>
<g class="iso__m fm__slip fm__slip--3">
<path class="iso__face iso__face--s1" d="M7.8 40.5 L18.2 43.5 L18.2 44.3 L7.8 41.3 Z"/>
<path class="iso__face iso__face--s2" d="M22.2 40 L18.2 43.5 L18.2 44.3 L22.2 40.8 Z"/>
<path class="iso__face" d="M11.8 37 L22.2 40 L18.2 43.5 L7.8 40.5 Z"/>
<path class="iso__ln iso__ln--lo" d="M10.8 39.9 L17.7 41.9"/>
<path class="iso__ln iso__ln--lo" d="M11.8 39 L18.7 41"/>
</g>
<g class="iso__m fm__wallet">
<path class="iso__face iso__face--s1" d="M14.2 44.2 L31.5 49.2 L31.5 52.7 L14.2 47.7 Z"/>
<path class="iso__face iso__face--s2" d="M37.8 43.8 L31.5 49.2 L31.5 52.7 L37.8 47.3 Z"/>
<path class="iso__face" d="M20.5 38.8 L37.8 43.8 L31.5 49.2 L14.2 44.2 Z"/>
<path class="iso__ln iso__ln--faint" d="M15.8 46.8 L29.9 50.8"/>
<path class="iso__face iso__face--s1" d="M20.5 38.8 L37.8 43.8 L33.7 43.5 L16.4 38.5 Z"/>
<path class="iso__face iso__face--s2" d="M37.8 43.8 L33.7 43.5 L33.7 44.1 L37.8 44.4 Z"/>
<path class="iso__face" d="M30.5 47.5 L30.1 47.7 L29.4 47.9 L28.8 47.9 L28.1 47.8 L27.6 47.6 L27.3 47.3 L27.3 46.9 L27.5 46.6 L27.9 46.4 L28.5 46.2 L29.2 46.2 L29.9 46.3 L30.4 46.5 L30.7 46.8 L30.7 47.2 Z"/>
<path class="iso__face iso__face--cav" d="M29.6 47.2 L29.3 47.4 L28.8 47.4 L28.4 47.3 L28.2 47.1 L28.3 46.9 L28.7 46.7 L29.1 46.7 L29.6 46.8 L29.7 47 Z"/>
</g>
<g class="iso__m fm__relay">
<path class="iso__halo" d="M146 67.6 L141.9 70.2 L136.5 72.3 L130.2 73.6 L123.5 74.1 L116.8 73.6 L110.5 72.3 L105.1 70.2 L101 67.6 L98.4 64.4 L97.5 61.1 L98.4 57.7 L101 54.6 L105.1 51.9 L110.5 49.8 L116.8 48.5 L123.5 48.1 L130.2 48.5 L136.5 49.8 L141.9 51.9 L146 54.6 L148.6 57.7 L149.5 61.1 L148.6 64.4 Z"/>
<path class="iso__face iso__face--s1" d="M102.9 53.1 L134.1 62.1 L134.1 69.9 L102.9 60.9 Z"/>
<path class="iso__face iso__face--s2" d="M144.1 53.4 L134.1 62.1 L134.1 69.9 L144.1 61.2 Z"/>
<path class="iso__face" d="M112.9 44.4 L144.1 53.4 L134.1 62.1 L102.9 53.1 Z"/>
<path class="iso__face iso__face--s1" d="M105.4 51.7 L133.1 59.7 L133.1 60.9 L105.4 52.9 Z"/>
<path class="iso__face iso__face--s2" d="M141.6 52.4 L133.1 59.7 L133.1 60.9 L141.6 53.6 Z"/>
<path class="iso__face" d="M113.9 44.4 L141.6 52.4 L133.1 59.7 L105.4 51.7 Z"/>
<path class="iso__face iso__face--cav" d="M107.2 59.5 L110.4 60.4 L110.4 58.2 L107.2 57.3 Z"/>
<path class="iso__face iso__face--cav" d="M112.4 61 L115.5 61.9 L115.5 59.7 L112.4 58.8 Z"/>
<path class="iso__face iso__face--cav" d="M117.6 62.5 L120.7 63.4 L120.7 61.2 L117.6 60.3 Z"/>
<path class="iso__ln iso__ln--faint" d="M106.4 55.6 L130.6 62.6"/>
</g>
<g class="fm__fan">
<path class="iso__way fm__way fm__way--1" d="M107.9 56 L45.5 38"/>
</g>
<path class="fm__comet fm__comet--out fm__route-1" d="M107.9 56 L45.5 38" pathLength="100"/>
<path class="fm__comet fm__comet--back fm__route-1" d="M45.5 38 L107.9 56" pathLength="100"/>
</svg>
<!-- /iso:film-hosting -->
            </div></div>
          </div>
          <p class="term__film-cap" id="termFilmCap"></p>
          <div class="term__film-card" id="termFilmCard">
            <span class="fig" id="termFilmStep"></span>
            <b class="term__film-title" id="termFilmTitle"></b>
            <span class="term__film-line" id="termFilmLine"></span>
          </div>
        </div>

        <!-- transport: play / pause / replay + a tuning-bar progress readout -->
        <div class="term__transport">
          <button class="term__btn" id="termPlay" type="button" aria-label="Pause demo" aria-pressed="true">❚❚</button>
          <button class="term__btn term__btn--replay" id="termReplay" type="button" aria-label="Replay demo">↻</button>
          <span class="term__bar-track" aria-hidden="true"><span class="term__bar-fill" id="termBarFill"></span></span>
        </div>
      </div>
    </div>
  </section>

  <!-- ===================== LIVE SIGNAL FIELD / MARKET ===================== -->
  <section class="market" id="market">
    <div class="wrap">
      <div class="market__head" data-reveal>
        <span class="sectionno">§2 / THE BAND</span>
        <h2>Live models, prices, and speeds.</h2>
        <p>Every model on the network right now, with how many machines serve it, how fast,
          how good, and what it costs. The machine owners set the prices; this list shows
          live data only.</p>
      </div>

      <div class="market__panel" data-reveal data-reveal-delay="80">
        <!-- top status line: the reactive on-air beacon + live broker status -->
        <div class="market__bar">
          <span class="market__status" id="marketStatus">
            <span class="beacon" aria-hidden="true">((<span class="beacon__dot"></span>))</span>
            <span id="marketStatusText">Live market · awaiting JavaScript</span>
          </span>
          <span class="market__meta">
            <button class="market__refresh" id="marketRefresh" type="button"
                    aria-label="Refresh the live band">refresh</button>
          </span>
        </div>

        <!-- column header -->
        <div class="mkt-row mkt-row--head" aria-hidden="true">
          <span class="mkt-cell mkt-cell--model">band · stations</span>
          <span class="mkt-cell mkt-cell--signal">signal</span>
          <span class="mkt-cell mkt-cell--speed">speed</span>
          <span class="mkt-cell mkt-cell--quality">quality</span>
          <span class="mkt-cell mkt-cell--price">price · out</span>
        </div>

        <!-- Live rows replace this honest static fallback after a successful read.
             The fallback row is the no-JS state only (hidden once html.js is set,
             see home.css) so it never stacks with the JS-on "No live data loaded
             yet" foot line below. -->
        <ul class="mkt-list" id="marketList" aria-live="polite">
          <li class="mkt-quiet mkt-quiet--nojs">
            <span class="mkt-quiet__txt">
              Live data loads when JavaScript is available.
              <a href="/models.html">Open the Models directory &rarr;</a>
            </span>
          </li>
        </ul>

        <p class="market__foot mono" id="marketFoot">
          No live data loaded yet · prices appear only after a broker response
        </p>
        <p class="market__tune mono">
          <a href="/models.html">tune the dial - browse every model &amp; station &rarr;</a>
        </p>
      </div>

      <!-- the money meter line (replaces the 3 cards). RATE + REPLY hydrate from the
           SAME live /market read as the band above (market.js) - real figures only.
           The static copy below is the no-JS / broker-down fallback: neutral on
           purpose, never made-up numbers. -->
      <div class="meter" aria-label="What it costs">
        <span id="meterRate"><span class="meter__k">RATE</span><b>live market</b>, set by operators</span>
        <span id="meterReply"><span class="meter__k">REPLY</span><b>priced per token</b>, in and out</span>
        <span><span class="meter__k">WALLET</span><b>debited per request</b>, per token</span>
      </div>
    </div>
  </section>
  </div><!-- /tone-zone -->

  <!-- ===================== COMPANY ===================== -->
  <section class="section company" id="company">
    <div class="wrap">
      <div class="section__head section__head--left" data-reveal>
        <span class="sectionno">§3 / COMPANY</span>
        <h2>What RogerAI is.</h2>
        <!-- The hero already says what RogerAI is (ROGERAI_IS); this lead says where, and what we build (founder 2026-11-21). -->
        <p>RogerAI is an AI company in Orange County, California. We design decentralized networks
          and the marketplace on them, and we pretrain our own models from scratch for machines and
          manufacturing. Our cloud is global. Our footprint is local.</p>
      </div>

      <div class="company__grid" data-reveal data-reveal-delay="60">
        <article class="company__card tint-panel">
          <span class="company__label">ROGER</span>
          <h3>The network</h3>
          <p>One OpenAI-compatible endpoint for developers and teams, answered by machines people
            share. Sharing a model never shares your files (<a href="/faq.html#share">what a requester can
            reach &rarr;</a>).</p>
          <p>Open model and runtime work ships under its model license. The network and
            broker are source-visible under PolyForm Perimeter.</p>
          <!-- The marketplace, in the card (founder 2026-11-21): who sets the price, who keeps
               what, and what the 10% is for, in the pricing page's own words (§4 "What the 10% is"). -->
          <p><b>The marketplace.</b> Anyone can put a machine on the air and set a price per
            token, zero included. On a paid band the station keeps 90%. The 10% is what keeps
            it private and spread out: the upstream sees a station, never a person, every
            request gets a signed receipt, and a dropped machine fails over to the next.
            <a href="/pricing.html#earn">What you earn &rarr;</a></p>
          <!-- Industrialized (founder 2026-11-21): why the same network runs machines. Every claim here is
               shipped and in the manual §E (Edge: certificates, no Tower, no market; the EDGE screen in
               the terminal, the console and the app; the eight standing alert rules and their sinks) or in
               the Wave catalog (the input formats the models read). Nothing exports metrics; the models read them.
               Simplified 2026-11-21 at the founder's word: plain words, no certificates or root, no list names. -->
          <p><b>Industrialized.</b> Your own machines can form an Edge: a private group on your
            own network, with no Tower and no market in the way. One screen, in the terminal, the
            console and the app, shows every machine and its health. Alerts for heat, disk, power
            and a machine gone quiet reach you by webhook, Discord or Telegram. Wave models read
            what a plant already emits: Modbus, Sparkplug B, Prometheus.
            <a href="/manual.html#sEdge">The Edge, in the manual &rarr;</a></p>
          <p class="company__links">
            <a href="/company.html">About the company</a>
            <a href="/models.html">Browse models</a>
            <a href="/manual.html">Read the manual</a>
            <a href="#top">Install and evaluate</a>
          </p>
        </article>

        <article class="company__card company__card--labs tint-panel">
          <TubePing />
          <span class="company__label">ROGERAI LABS · THE WAVE SPECTRUM</span>
          <h3>Wave: small models for machines</h3>
          <p><span class="status">{@html facts.wave('pico.index.chip')}</span>
            <!-- On air per src/data/wave-status.json, and read live on broker.rogerai.fm/v1/models 2026-11-24:
                 wave-pico-293m, wave-nano-fire-controller, wave-nano-fire-vision. -->
            Wave Pico runs free on the network today: 293M parameters, trained from scratch,
            small enough for a Raspberry Pi. Wave Nano, the next size up, is on the network too,
            with two fire adapters: a fire controller and fire vision, as specialized models. Nano
            is sized for an NVIDIA Jetson, a Mac mini or a Mac Studio.</p>
          <p>Small models for factories and other industrial sites. Our lab works on
            enabling open models to run on hardware from a Raspberry Pi up: NVIDIA Jetson, AMD
            and Apple machines too. Each job gets an adapter, a small add-on trained on your
            records and switched on per request.
            <a href="/research-custom-tuning.html">Custom tuning &rarr;</a></p>
          <!-- Licence detail (which network services fall under which terms) lives
               with the catalogue on research-models.html#licence; the card keeps
               the two promises a visitor needs and links out for the rest. -->
          <p>When it releases, the license is intended to be permissive: local use will not
            require RogerAI or its broker, and optional network services use separate network
            terms. <a href="/research-models.html#licence">License details &rarr;</a></p>
          <p class="company__links">
            <a class="company__primary" href="/research.html">Follow the program</a>
            <a href="/broadcasts.html">Open Air Waves</a>
          </p>
        </article>
      </div>

      <!-- THE WHOLE LADDER, not a sample of it. This strip used to name five of
           the seven tiers and elide the rest behind an ellipsis, which read as
           if the family stopped at Giga. All seven are here now, each with the
           one thing a reader actually wants - how big, and what it watches.
           The bar over each tier is the wave itself: a repeating gradient whose
           stripes TIGHTEN as the model grows, so the spectrum is legible before
           a word is read. It is drawn in CSS and costs nothing to ship - the
           engraved tier plates are richer but weigh ~640KB together, which is
           not a homepage's job. -->
      <ol class="home-spectrum" data-scrub data-scrub-auto aria-label="The Wave Spectrum, pico to exa">
        <li><i class="wave-bar" aria-hidden="true"></i><span>{@html facts.wave('pico.index.spectrum')}</span><b>Wave Pico</b><small>one machine&rsquo;s telemetry · Raspberry Pi</small></li>
        <li><i class="wave-bar" aria-hidden="true"></i><span>0.8–1.5B</span><b>Wave Nano</b><small>fleet rollup + routing · a gateway</small></li>
        <li><i class="wave-bar" aria-hidden="true"></i><span>7–8B</span><b>Wave Micro</b><small>site brain · a facility</small></li>
        <li><i class="wave-bar" aria-hidden="true"></i><span>27–35B</span><b>Wave Giga</b><small>full-plant reasoning · a plant</small></li>
        <li><i class="wave-bar" aria-hidden="true"></i><span>80–120B</span><b>Wave Tera</b><small>faults across many plants</small></li>
        <li><i class="wave-bar" aria-hidden="true"></i><span>150–200B</span><b>Wave Peta</b><small>regional scale · pruned from the flagship</small></li>
        <li><i class="wave-bar" aria-hidden="true"></i><span>&gt;280B</span><b>Wave Exa</b><small>the frontier flagship, the family&rsquo;s teacher</small></li>
      </ol>
      <p class="home-spectrum__foot">
        Every tier ships with Wave Mesh baked in and understands the output of every
        tier beneath it. <a href="/research-wave-family.html">The Wave Spectrum &rarr;</a>
        <a href="/research.html">Full program status &rarr;</a>
      </p>
    </div>
  </section>

  <!-- ===================== ON SITE: YOUR DEVICE, NOW A SENSOR =====================
       features/web/sensor_positioning.feature (PROPOSED 2026-10-15, founder review).
       Design tense and number-free on purpose: nothing here is claimed in the field, and
       every status (the step 3 lamp words, the dial) is a cap token from
       src/data/capabilities.json, so this section can never run ahead of the file. -->
  <section class="section sensor" id="sensor">
    <div class="wrap">
      <div class="section__head section__head--left" data-reveal>
        <span class="sectionno">§4 / ON SITE</span>
        <h2>Your device, now a sensor.</h2>
        <p>No box to buy. A computer the site already has, a Wave model, and a small adapter
          for one job that plugs in without changing the model.</p>
      </div>

      <ol class="steps sensor__steps" data-reveal data-reveal-delay="60">
        <li><b>Install RogerAI.</b> One command on a computer you already have: a laptop, a
          desktop, or the industrial PC on the floor.</li>
        <li><b>Pick a Wave model.</b> Sized to the machine. Wave Pico, the smallest, is built
          for one machine&rsquo;s telemetry.</li>
        <li><b>Pick an adapter.</b> A small add-on for one job. The first, fire early
          warning, is <span data-cap="fire-warning" data-cap-stage="{facts.t.cap('fire-warning.stage')}"><span class="cap-stage cap-stage--inline">{@html facts.cap('fire-warning.label')}</span></span>.
          Telemetry drift needs no adapter: it is built into the base.</li>
      </ol>

      <p class="sensor__get" data-reveal>Or we train the adapter on your records
        (<a href="/research-custom-tuning.html">custom tuning</a>), served from our cloud on a
        private band only you can reach, or built to run on site, offline if you want. It hands
        anything it cannot vouch for to a person. It is advisory, beside the control system and
        never in it: it never moves a machine, and a person makes every call.</p>

      <div class="sensor__dial" data-reveal>
        <AdapterDial />
      </div>
      <p class="home-spectrum__foot">Every lamp reads the same status file as every other page. A
        lamp turns red only when you can install that adapter on your own machine.
        <a href="/research-industry.html#adapters">See where it sits &rarr;</a>
        <a href="/research-hardware.html#devices-row">The devices &rarr;</a></p>
    </div>
  </section>

  <!-- ===================== THE TUNE ===================== -->
  <section class="section band band--inset" id="what">
    <div class="wrap">
      <div class="section__head" data-reveal>
        <span class="sectionno">§5 / THE TUNE</span>
        <h2>Your choice: stay local, or reach a network.</h2>
      </div>

      <!-- THE BAND, TUNED (app/lib/ui/BandFigure.svelte, band-figure.css, 2026-11-20): the one
           drawing of routing, failover and metering. It plays its three steps itself. -->
      <div data-reveal>
        <BandFigure fig={4} />
      </div>

      <div class="twoway" data-draw data-reveal data-reveal-delay="80">
        <article class="twoway__side">
          <div class="twoway__tag twoway__tag--air">go on air</div>
          <h3>Share a model</h3>
          <p>Point RogerAI at your local backend (vLLM, llama.cpp, Ollama, LM Studio, Osaurus). Set a call-sign,
            a rate, and your hours. Flip the switch. You are on air.</p>
          <code class="twoway__cmd">roger share <span class="tok">qwen3.8-27b</span> <span class="tok">--rate</span> 0.30</code>
        </article>

        <div class="twoway__link" aria-hidden="true">
          <span class="twoway__beam"></span>
        </div>

        <article class="twoway__side">
          <div class="twoway__tag twoway__tag--tune">tune in</div>
          <h3>Use a model</h3>
          <p>Pick a band, set your margins (max price, min speed). RogerAI locks the strongest
            station and hands you a local OpenAI endpoint + key. When one fades, it re-tunes.</p>
          <code class="twoway__cmd">roger use <span class="tok">qwen3.8-27b</span> <span class="tok">--max-price</span> 0.30 <span class="tok">--min-tps</span> 40</code>
        </article>
      </div>
    </div>
  </section>

  <!-- ===================== SPECIFICATIONS ===================== -->
  <!-- Was eleven full-width rows - most of a screen to say eleven short things. It is the
       nameplate riveted to the back of the set now: an etched label, four rivets, and the
       whole instrument in two dense rows. Nothing was cut to get there;
       test/homepage-plate-and-books.test.mjs counts the eleven and checks their values, so
       a "shorter" plate that quietly dropped a row fails rather than ships. -->
  <section class="section section--plate" id="spec">
    <div class="wrap">
      <div class="plate" data-reveal>
        <!-- The etched label IS the section heading: collapsing the standing
             head into the plate is what bought the space, but a section with no
             heading leaves a hole in the document outline, so it stays an h2 and
             is styled down rather than replaced by a span. -->
        <h2 class="plate__etch">
          <span class="sectionno">§6 / SPECIFICATIONS</span>
          <span class="plate__sub">the whole instrument, one plate</span>
        </h2>
        <dl class="plate__grid" aria-label="RogerAI specifications">
          <div><dt>Protocol</dt><dd>OpenAI-compatible <i>drop-in</i></dd></div>
          <div><dt>Transport</dt><dd>Outbound long-poll <i>no inbound ports</i></dd></div>
          <div><dt>Endpoint</dt><dd><span class="mono">127.0.0.1</span> <i>local, key-scoped</i></dd></div>
          <div><dt>Pricing</dt><dd>Per token <i>operator-set, from free, $/1M</i></dd></div>
          <div><dt>Settlement</dt><dd>Wallet balance <i>debited live</i></dd></div>
          <div><dt>Lineage</dt><dd>ed25519 co-signed <span class="cs">&#9670;</span> <i>weights, shard, token</i></dd></div>
          <div><dt>Split</dt><dd><span class="mono">90 / 10</span> <i>operator's favor</i></dd></div>
          <div><dt>Curated</dt><dd><span class="mono">list + 10%</span> <i>&raquo; marked, split on the receipt</i></dd></div>
          <div><dt>Relay split</dt><dd><span class="mono">90 / 5 / 5</span> <i>station / tower / platform</i></dd></div>
          <div><dt>Interfaces</dt><dd>Terminal + browser <i>one store, two windows</i></dd></div>
          <div><dt>Reach</dt><dd>Only the model port <i>no shell, no files, no context</i></dd></div>
          <div><dt>Platforms</dt><dd><span class="mono">macOS &middot; Linux &middot; Windows</span> <i>Apple Silicon: MLX and GGUF</i></dd></div>
          <div><dt>Backends</dt><dd><span class="mono">vLLM &middot; llama.cpp &middot; Ollama &middot; LM Studio &middot; Osaurus &middot; LiteLLM &middot; any OpenAI-compatible</span></dd></div>
        </dl>
      </div>
    </div>
  </section>

  <!-- ===================== THE INDUSTRIAL SERIES ===================== -->
  <!-- The o'ailly books. Edge Impulse leans on "we wrote the book" and so can we, with the
       difference that ours are free to read and say on the cover who wrote them.
       HONESTY RULE, same as the reserved App Store slot: only a title a stranger can open
       today may appear here. The third title in the series is still a draft and is
       deliberately absent - the test fails if it shows up before it is published.
       Covers are the press's own artwork, shipped local rather than hotlinked. -->
  <section class="section band band--inset" id="books">
    <div class="wrap shelf">
      <div class="shelf__copy" data-reveal>
        <span class="sectionno">§7 / THE INDUSTRIAL SERIES</span>
        <h2>We wrote the book on this. Then we wrote the next one.</h2>
        <p>Two field books for people who own machines: what a small local model can
           honestly do on a plant floor, and the physics of running one on hardware you
           already paid for. Written by machines, put through the press's three-pass
           review, verified by Roger AI, and free to read in full.</p>
        <p class="shelf__colophon">
          <img class="shelf__mark" src="assets/books/oailly.svg" width="640" height="220"
               alt="o'ailly press wordmark: books by AI, for AI, human readable" loading="lazy" />
          <span>Published by <a href="https://oailly.com/" target="_blank" rel="noopener noreferrer">o&rsquo;ailly press</a>,
                a machine-first imprint. The whole catalogue ships with its review trail attached.</span>
        </p>
      </div>

      <div class="shelf__stack" data-reveal data-reveal-delay="100">
        <a class="shelf__book" href="https://oailly.com/read/rogerai-labs--local-llms-for-manufacturing/"
           target="_blank" rel="noopener noreferrer">
          <img class="shelf__cover" src="assets/books/cover-manufacturing.svg" width="600" height="780"
               alt="Cover of Local LLMs for Manufacturing, o'ailly Industrial Series number 1, with a line-drawn beetle"
               loading="lazy" />
          <span class="shelf__no"><b>N&ordm; 1</b> Local LLMs for Manufacturing<em>read it free &rarr;</em></span>
        </a>
        <a class="shelf__book" href="https://oailly.com/read/rogerai-labs--inference-on-the-edge/"
           target="_blank" rel="noopener noreferrer">
          <img class="shelf__cover" src="assets/books/cover-inference.svg" width="600" height="780"
               alt="Cover of Inference on the Edge, o'ailly Industrial Series number 2, with a line-drawn dragonfly"
               loading="lazy" />
          <span class="shelf__no"><b>N&ordm; 2</b> Inference on the Edge<em>read it free &rarr;</em></span>
        </a>
      </div>
    </div>
  </section>

  <!-- INK PANEL: monetize, privacy, go. -->
  <div class="tone-zone" data-tone="ink">
  <!-- ===================== MONETIZE ===================== -->
  <section class="section band" id="monetize">
    <div class="wrap monetize">
      <div class="monetize__copy" data-reveal>
        <span class="sectionno">§8 / MONETIZE</span>
        <h2>Two kinds of transmitter. One dial.</h2>
        <p><b>Your GPU.</b> Set a price and keep 90% of every token, or run it free.</p>
        <p><b>Your contracts.</b> Put an API contract you already pay for on the air as a
          <b>curated station</b>: list price plus 10%, your bill back in full plus half the
          fee. <a href="/pricing.html#earn">The policy &rarr;</a></p>
      </div>

      <aside class="earn"
             data-rate="0.30" data-volume-min="10000000" data-volume-max="30000000"
             data-uptime="0.50" data-share="0.90"
             data-reveal data-reveal-delay="100" aria-label="Illustrative operator earnings panel">
        <span class="fig" aria-hidden="true">FIG. 5 - OPERATOR PANEL</span>
        <div class="earn__head">
          <span class="earn__tag">@you <span class="cs">&#9670;</span></span>
          <span class="earn__mode">ILLUSTRATIVE</span>
        </div>
        <div class="earn__row"><span>model</span><b class="mono">qwen3.8-27b</b></div>
        <div class="earn__row"><span>rate</span><b class="mono">$0.30 / 1M out</b></div>
        <div class="earn__row"><span>volume</span><b class="mono">10M - 30M out / day</b></div>
        <div class="earn__row"><span>uptime</span><b class="mono">50%</b></div>
        <div class="earn__row"><span>your share</span><b class="mono">90%</b></div>
        <div class="earn__spark">
          <span class="earn__label">illustrative / day</span>
          <span class="earn__bars" aria-hidden="true"></span>
          <b class="mono earn__amt">$1.35 - $4.05</b>
        </div>
        <div class="earn__formula mono">rate × volume × uptime × share</div>
        <div class="earn__foot mono">example assumptions · demand and earnings are not guaranteed</div>
      </aside>
    </div>
    <!-- THE MONEY, DRAWN ONCE (SITE-SIMPLIFICATION-AUDIT fix 4): the same figure the pricing
         page carries in full, here in its small size, so the split is seen before it is read. -->
    <div class="wrap" data-reveal>
      <MoneyFigure fig={6} size="small" />
    </div>
  </section>

  <!-- ===================== PRIVACY-FIRST AIRWAVES =====================
       §8-§10 merged (founder, 2026-09-02): operator unlinking, the Tower, and the
       go-direct concession as six pictogram cards - claims, not paragraphs. The
       icons are native SVG in the site system (mono stroke + the one red), so
       they hold in both themes under the strict CSP. -->
  <section class="section band" id="operator">
    <div class="wrap">
      <div class="section__head section__head--left" data-reveal>
        <span class="sectionno">§9 / PRIVACY-FIRST AIRWAVES</span>
        <h2>The provider hears a station. Never you.</h2>
      </div>
      <div class="pcards" data-draw data-reveal>
        <article>
          <svg class="pc-ic" viewBox="0 0 48 48" aria-hidden="true">
            <rect x="7" y="12" width="34" height="24" rx="2"/>
            <circle cx="17" cy="22" r="4"/>
            <path d="M12 32c1.5-3.5 8.5-3.5 10 0"/>
            <path d="M27 20h10M27 25h10M27 30h6"/>
            <path class="pc-red" d="M5 40 L43 8"/>
          </svg>
          <h3>A callsign, not a name</h3>
          <p>Whoever answers hears a callsign and a job. Not your name, not your card,
             not your IP.</p>
        </article>
        <article>
          <svg class="pc-ic" viewBox="0 0 48 48" aria-hidden="true">
            <path d="M8 24 L24 10 L40 24"/>
            <path d="M12 22v16h24V22"/>
            <path d="M22 38v-8h4v8"/>
            <path class="pc-red" d="M24 18v-8M21 13l3-3 3 3" fill="none"/>
          </svg>
          <h3>Nothing ever knocks</h3>
          <p>Stations dial out. No open port, no inbound path, no door to knock on.</p>
        </article>
        <article>
          <svg class="pc-ic" viewBox="0 0 48 48" aria-hidden="true">
            <path d="M6 30c4-2 6-8 10-8s5 6 9 6 5-8 9-8 4 6 8 6"/>
            <path d="M10 38h28"/>
            <circle class="pc-red pc-fill" cx="40" cy="12" r="3"/>
          </svg>
          <h3>The honest limit</h3>
          <p>Your words still reach the machine that answers. Too much? A
             <a href="/confidential.html">confidential station</a> or
             <a href="/tower.html">your own Tower</a>.</p>
        </article>
        <article>
          <svg class="pc-ic" viewBox="0 0 48 48" aria-hidden="true">
            <path d="M24 8 L16 40 M24 8 L32 40 M19 28h10 M21 20h6"/>
            <path d="M10 40h28"/>
            <circle class="pc-red pc-fill" cx="24" cy="6" r="2.5"/>
          </svg>
          <h3>Carry, don&rsquo;t compute</h3>
          <p>No GPU? A Tower relays sealed work and keeps <b>5%</b> of everything it
             carries. <a href="/tower.html">Raise a Tower &rarr;</a></p>
        </article>
        <article>
          <svg class="pc-ic" viewBox="0 0 48 48" aria-hidden="true">
            <path d="M28 8h-16v32h16"/>
            <path d="M20 24h20M35 19l5 5-5 5"/>
            <circle class="pc-red pc-fill" cx="25" cy="24" r="2"/>
          </svg>
          <h3>The exit is built in</h3>
          <p>Run it <b>standalone</b>: your own private network, free, and it never
             phones home. A network you can leave has to earn your staying.</p>
        </article>
        <article>
          <svg class="pc-ic" viewBox="0 0 48 48" aria-hidden="true">
            <path d="M8 16h20a8 8 0 0 1 0 16H8"/>
            <path d="M8 24h32M35 19l5 5-5 5"/>
            <circle class="pc-red pc-fill" cx="22" cy="24" r="2"/>
          </svg>
          <h3>Sometimes, go direct</h3>
          <p>Steady volume to one vendor is cheaper direct - do it. What you&rsquo;d
             leave behind sits <a href="/pricing.html#why">next to the prices &rarr;</a></p>
        </article>
      </div>
    </div>
  </section>

  <!-- ===================== CTA STRIP ===================== -->
  <section class="cta" id="go">
    <div class="wrap cta__inner" data-reveal>
      <span class="sectionno">§10 / GO</span>
      <h2>Start with one command.</h2>
      <p><code class="inline">roger</code> is one small program. It finds the models on your
        machine and <b>connects you to everyone else&rsquo;s</b>.</p>
      <InstallBox id='installCmd2' size='lg'>curl <span class="tok">-fsSL</span> https://rogerai.fm/install.sh <span class="tok">| sh</span></InstallBox>
      <div class="install__meta">
        <span id="installNoteWin2" data-alt-note>On Windows (PowerShell): <code class="inline">irm https://rogerai.fm/install.ps1 | iex</code></span>
      </div>

      <div class="surfaces" data-reveal data-reveal-delay="60">
        <p class="surfaces__lead">Then pick your window. Same station, same wallet, same
          settings.</p>
        <div class="surfaces__grid">
          <article class="surfaces__card">
            <h3 class="surfaces__h">The dial</h3>
            <code class="surfaces__cmd">roger</code>
            <p>The terminal: browse the band, tune a station, chat, put your models on air.</p>
          </article>
          <article class="surfaces__card">
            <h3 class="surfaces__h">The console</h3>
            <code class="surfaces__cmd">roger webui</code>
            <p>The same thing in a browser tab, with room for tables.</p>
          </article>
        </div>
      </div>
    </div>
  </section>
  </div><!-- /tone-zone -->
  </main>

  <!-- ============================ FOOTER ============================ -->
  <Footer anchors='' />

<!-- canonical $ renderer; before market.js (meter REPLY) -->
