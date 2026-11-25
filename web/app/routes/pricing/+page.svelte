<script>
  import TocTuner from "#lib/ui/TocTuner.svelte";
  // Pricing - RogerAI
  // Ported from web/src/pricing.html (2026-10-29); the copy is the legacy copy.
  import { siteHead, CSS_SHARED, CSS_ACCOUNT } from "#lib/head.js";
  import Footer from "#lib/ui/Footer.svelte";
  import GuardFigure from "#lib/ui/GuardFigure.svelte";
  import HoldFigure from "#lib/ui/HoldFigure.svelte";
  import MoneyFigure from "#lib/ui/MoneyFigure.svelte";
  import Nav from "#lib/ui/Nav.svelte";
  import OnAir from "#lib/ui/OnAir.svelte";
  import Rail from "#lib/ui/Rail.svelte";
  const head = siteHead({
    title: 'Pricing - RogerAI',
    desc: 'RogerAI has no price list: operators set their own per-token rates and many stations broadcast free. What is fixed is the rules - no subscription, a wallet debited per token, a 90/10 split in the operator\'s favor, a $25 minimum payout after a 30-day hold.',
    og: { type: 'website', title: 'RogerAI pricing - the rules, not a price list', desc: 'Per token, in and out, at a price the operator sets. No subscription, no seat, no minimum. Operators keep 90%. Free stations are normal, not a trial.', url: 'https://rogerai.fm/pricing.html', image: 'https://rogerai.fm/og-home.png' },
    css: [...CSS_SHARED, 'research.css', 'pricing.css', 'money-figure.css', 'guard-figure.css', 'hold-figure.css'],
    scripts: [{ runtime: true }, { page: "pricing" }],
  });
</script>
<svelte:head>{@html head}</svelte:head>
<!-- PRICING: the money, as RULES.
     The site had prices in exactly two places - a live row on the dial and one line of
     the homepage spec sheet - so a first-time reader could not learn that free stations
     exist, that there is no subscription, or what an operator keeps. This page answers
     that without ever quoting a number a station controls.
     The doctrine is the same one homepage-data-honesty.test.mjs enforces: the RULES are
     ours and stable, so they are printed; the PRICES are the operators' and change by the
     hour, so they are pointed at. No model is named anywhere on this page - a model name
     only ever appears next to a rate, and that belongs on /models.html.
     Every figure traces to a committed source: the 90/10 split to tos.html, the ceilings
     and the wallet rules to manual.html §6, the payout policy to manual.html §7, the
     90/5/5 relay settlement to broadcasts-run-a-tower.html. Contract: test/pricing-faq-llms.test.mjs. 
     (the page's doc comment, from the legacy file; its sheets are the css: list above) -->
  <Rail head='RogerAI &middot; Pricing' rev='Rev. 2026.08 &middot; ON AIR' />

  <Nav variant='marketing' home='/' current='pricing' />

  <main id="top" data-lift>
    <section class="research-hero">
      <div class="wrap">
        <span class="hero__eyebrow">
          <OnAir />
          <span class="hero__callsign">ROGERAI</span>
          <span class="hero__sep" aria-hidden="true">&middot;</span>
          pricing
        </span>
        <h1>The people who run the machines set the prices. Many charge nothing.</h1>
        <p class="research-hero__lead">
          Every model runs on someone's machine, and the price per token is set by the
          operator, never by us. What never changes: no subscription, you pay per token from a
          wallet, an unusable answer costs nothing, and an operator keeps 90%.
        </p>
        <div class="research-actions research-actions--lead">
          <a class="research-button" href="/models.html">Live prices on the dial <span aria-hidden="true">&rarr;</span></a>
          <a class="research-button" href="#paths">Ways in <span aria-hidden="true">&darr;</span></a>
          <a class="research-button" href="#earn">Earning <span aria-hidden="true">&darr;</span></a>
        </div>
      </div>
    </section>

    <!-- The page's contents as the homepage's tuning scale (components.css .toc-tuner,
         tuner.js). Each station reads its section's own label, so the tuner adds no
         words. Never pinned. Contract: test/models-lift.test.mjs. -->
    <TocTuner sections={[
      { id: "why", no: "§0", name: "WHAT THE FEE BUYS" },
      { id: "paths", no: "§1", name: "THE WAYS IN" },
      { id: "pay", no: "§2", name: "WHAT YOU PAY" },
      { id: "safety", no: "§3", name: "PRICE SAFETY" },
      { id: "earn", no: "§4", name: "WHAT YOU EARN" },
      { id: "calc", no: "§5", name: "THE ARITHMETIC" },
      { id: "free", no: "§6", name: "WHAT IS NOT BILLED" },
      { id: "ask", no: "§7", name: "STILL UNSURE" },
    ]} />

    <section class="section" id="why">
      <div class="wrap">
        <div class="section__head section__head--left" data-reveal>
          <span class="sectionno">&sect;0 / WHAT THE FEE BUYS</span>
          <h2>The waves carry your words. Never your name.</h2>
        </div>

        <!-- THE RELAY WIRE. The whole argument, drawn instead of written: two payloads
             leave your set - the job (the one red) and your identity (grey chips). The
             relay drops the identity; a callsign and the job are all that reach the
             machine that answers. Pure CSS/SVG, no JS, tokens only, so it holds in both
             themes and under the CSP. prefers-reduced-motion gets the honest end-state.
             Contract: test/why-rogerai.test.mjs. -->
        <figure class="rwire" role="img" aria-label="Diagram: a request leaves you carrying the job and your identity. The relay drops the identity - name, card, IP - and only a station callsign with the job reaches the machine that answers.">
          <div class="scroll-box scroll-box--diagram">
          <svg class="rwire__svg" viewBox="0 0 958 190" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
            <line class="rwire__line" x1="100" y1="100" x2="840" y2="100" />
            <!-- YOU -->
            <circle cx="72" cy="100" r="24" class="rwire__ring" />
            <text x="72" y="105" text-anchor="middle" class="rwire__nodelabel">YOU</text>
            <!-- RELAY -->
            <path class="rwire__tower" d="M470 64 L456 118 M470 64 L484 118 M460 104 L480 104 M463 90 L477 90" />
            <circle cx="470" cy="58" r="4" class="rwire__beacon" />
            <text x="470" y="140" text-anchor="middle" class="rwire__nodelabel">RELAY</text>
            <text x="470" y="160" text-anchor="middle" class="rwire__micro">your identity ends here</text>
            <!-- STATION -->
            <rect x="844" y="78" width="44" height="44" class="rwire__ring" />
            <text x="866" y="140" text-anchor="middle" class="rwire__nodelabel">STATION</text>
            <text x="866" y="160" text-anchor="middle" class="rwire__micro">a callsign + the job</text>
            <!-- the identity chips: travel with the job, dropped at the relay -->
            <g class="rwire__id">
              <rect x="96" y="52" width="46" height="20" rx="3" class="rwire__chip" />
              <text x="119" y="66" text-anchor="middle" class="rwire__chiptext">NAME</text>
              <rect x="146" y="52" width="46" height="20" rx="3" class="rwire__chip" />
              <text x="169" y="66" text-anchor="middle" class="rwire__chiptext">CARD</text>
              <rect x="196" y="52" width="32" height="20" rx="3" class="rwire__chip" />
              <text x="212" y="66" text-anchor="middle" class="rwire__chiptext">IP</text>
            </g>
            <!-- the job: the one red, and the only thing that goes the distance -->
            <g class="rwire__job">
              <circle cx="100" cy="100" r="11" class="rwire__jobglow" />
              <circle cx="100" cy="100" r="5" class="rwire__jobdot" />
            </g>
          </svg>
          </div>
        </figure>

        <ul class="whyrows" data-reveal>
          <li><span class="whyrows__k">SEEN</span>
            <span>The upstream sees a station, never a person. Not your name, not your
              card, not your IP.</span></li>
          <li><span class="whyrows__k">PROOF</span>
            <span>A counter-signed receipt on every request, the split printed on it.</span></li>
          <li><span class="whyrows__k">CEILING</span>
            <span>Your maximum price rides every request. A band above it is never dialed.</span></li>
          <li><span class="whyrows__k">SUPPLY</span>
            <span>Real people&rsquo;s machines. Commercial relays wear the &raquo; mark,
              never blurred together.</span></li>
          <li><span class="whyrows__k">NO SEAT</span>
            <span>No subscription, no seat, no minimum. The balance moves only while an
              answer streams.</span></li>
          <li><span class="whyrows__k">EXIT</span>
            <span>The router is a program. Run your own Tower, free, forever.</span></li>
        </ul>

        <p class="whyfine">The limit, plainly: the words themselves still reach the machine
          that answers. Too much? Use a <a href="/confidential.html">confidential station</a>
          or <a href="/tower.html">your own Tower</a>. Steady volume to one vendor? Going
          direct is cheaper. <a href="/manual.html#sMoney-fee">All fourteen things the fee
          buys, in the manual &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="paths">
      <div class="wrap">
        <div class="section__head section__head--left" data-reveal>
          <span class="sectionno">&sect;1 / THE WAYS IN</span>
          <h2>Every way onto the band.</h2>
          <p>Start where you are. Free is not a trial: the rate is zero and everything else
             works the same.</p>
        </div>

        <div class="paths">
          <article class="tint-panel">
            <span class="paths__cost paths__cost--none">costs nothing</span>
            <h3>Your own station</h3>
            <p><code>roger share</code> puts your machine on the air. Free by default, no
               login, nobody to pay.</p>
            <a class="tlink" href="/broadcasts-share-gpu-earn.html">One command puts it up &rarr;</a>
          </article>
          <article class="tint-panel">
            <span class="paths__cost paths__cost--none">costs nothing</span>
            <h3>A friend&rsquo;s station</h3>
            <p>Someone you trust runs the same command and reads you the callsign. Their
               rate, often zero.</p>
            <a class="tlink" href="/manual.html#quickstart">How to tune in &rarr;</a>
          </article>
          <article class="tint-panel">
            <span class="paths__cost paths__cost--none">costs nothing</span>
            <h3>A free station</h3>
            <p>Bands broadcast at zero all the time. You need <b>no account</b> to browse the
               dial or to open a free endpoint.</p>
            <a class="tlink" href="/models.html">See what is free right now &rarr;</a>
          </article>
          <article class="tint-panel">
            <span class="paths__cost">per token</span>
            <h3>A paid station</h3>
            <p>Log in once, and a wallet is debited per token, input and output, at the rate
               the station published.</p>
            <a class="tlink" href="#pay">What you can be charged &darr;</a>
          </article>
          <article class="tint-panel">
            <span class="paths__cost">list + 10%</span>
            <h3>A curated station</h3>
            <p>A paid API someone already pays for, <b>their own provider contracts</b>, on
               the air with the <b>&raquo;</b> mark. The list price plus the 10% routing fee.
               Our own house bands skip the fee - straight pass-through at the provider's list.</p>
            <a class="tlink" href="#why">What the fee buys &rarr;</a>
          </article>
          <article class="tint-panel">
            <span class="paths__cost paths__cost--none">costs nothing</span>
            <h3>Your own tower</h3>
            <p>A Tower running standalone needs nothing from us: your machines, your people,
               your network. No account, no metering.</p>
            <a class="tlink" href="/tower.html">How a Tower works &rarr;</a>
          </article>
        </div>
      </div>
    </section>

    <section class="section" id="pay">
      <div class="wrap">
        <div class="section__head section__head--left" data-reveal>
          <span class="sectionno">&sect;2 / WHAT YOU PAY</span>
          <h2>Per token, from a wallet, at a price the operator set.</h2>
          <p>Who pays whom, drawn once. The rules below never change when you spend more.</p>
        </div>

        <!-- THE MONEY, DRAWN ONCE (app/lib/ui/MoneyFigure.svelte, money-figure.css): the
             wallet pays per token; 90 to the station and 10 to RogerAI, or 90 / 5 / 5 through
             a Tower; the curated list, free models and custom tuning as cases under it.
             Every number is a rule this page already prints, never a price. -->
        <MoneyFigure fig={1} />

        <dl class="pplate tint-panel" data-reveal>
          <div>
            <dt>Unit</dt>
            <dd>Per <b>token</b>, input and output priced separately, in dollars per million.</dd>
          </div>
          <div>
            <dt>Who sets it</dt>
            <dd>The <b>operator</b>, per model. The live number is on the
                <a href="/models.html">dial</a>.</dd>
          </div>
          <div>
            <dt>Subscription</dt>
            <dd><b>None.</b> No plan, no seat, no minimum, nothing to cancel.</dd>
          </div>
          <div>
            <dt>Wallet</dt>
            <dd>Debited live as tokens settle. It <b>cannot go negative</b>; there is no bill
                at the end of the month.</dd>
          </div>
          <div>
            <dt>To start</dt>
            <dd>A <b>$1 starter credit</b> the first time you log in, one per account, while
                it lasts.</dd>
          </div>
          <div>
            <dt>Top-ups</dt>
            <dd>Through Stripe Checkout, above the <b>$1</b> minimum: <button class="copy-code" type="button" data-copy-target aria-label="Copy roger topup 25"><code>roger topup 25</code></button>.
                Stripe holds every card detail; we never see one.</dd>
          </div>
          <div>
            <dt>An unusable answer</dt>
            <dd>Costs <b>nothing</b>. It is voided: you pay zero, the operator earns zero and
                takes a strike.</dd>
          </div>
          <div>
            <dt>The token count</dt>
            <dd>Is the <b>broker's own re-count</b>, and settlement takes the smaller of the
                two.</dd>
          </div>
          <div>
            <dt>Custom tuning</dt>
            <dd>A Wave model trained on your records, served from our cloud on a private band
                only you can reach. <b>Price on request.</b>
                <a href="/research-custom-tuning.html">How it works</a>.</dd>
          </div>
        </dl>
        <p class="sec-more mono"><a href="/manual.html#sMoney-pay">The billing rules in full, in the manual &rarr;</a></p>
      </div>
    </section>

    <section class="section" id="safety">
      <div class="wrap">
        <div class="section__head section__head--left" data-reveal>
          <span class="sectionno">&sect;3 / PRICE SAFETY</span>
          <h2>Set your own budgets and routing configurations.</h2>
          <p>In an open market, having control over how you spend or sell is essential.</p>
        </div>

        <!-- The four rules, drawn: a price scale, a gate for your ceiling, stations as price
             tags (GuardFigure.svelte, guard-figure.css; it plays its four steps itself). -->
        <GuardFigure fig={2} />

        <div class="guards" data-reveal>
          <article class="tint-panel">
            <h3>A ceiling you did not have to set</h3>
            <p>Set no limit and the client still caps every connection at <b>$10 / 1M</b> out,
               even under <code>--yes</code>.</p>
          </article>
          <article class="tint-panel">
            <h3>Type the price back</h3>
            <p>Above <b>$20 / 1M</b> out you type the number, not press y.</p>
          </article>
          <article class="tint-panel">
            <h3>A listing ceiling</h3>
            <p>The broker refuses a station above <b>$100 / 1M</b> out or <b>$50 / 1M</b> in.
               It binds <b>every band, public or private</b>.</p>
          </article>
          <article class="tint-panel">
            <h3>No price hike, ever</h3>
            <p>Your ceiling rides <b>every request</b> as a header. A band that drifts past it
               is routed around, not charged.</p>
          </article>
        </div>
      </div>
    </section>

    <section class="section" id="earn">
      <div class="wrap">
        <div class="section__head section__head--left" data-reveal>
          <span class="sectionno">&sect;4 / WHAT YOU EARN</span>
          <h2>You set the price. You keep 90 percent.</h2>
          <p>The other side of the same market: serve your own hardware, or resell capacity
             you already pay for. Nothing is negotiated or tiered.</p>
        </div>

        <dl class="pplate tint-panel" data-reveal>
          <div>
            <dt>Your GPU</dt>
            <dd><b>90%</b> of every token you serve, at a price you set. The 10% fee is our
                whole take.</dd>
          </div>
          <div>
            <dt>Your contracts</dt>
            <dd>A <b>curated station</b> in front of a provider account you hold. Consumers
                pay <b>list + 10%</b>; you get the <b>list back whole, plus half the routing fee</b>.
                One flag:
                <button class="copy-code" type="button" data-copy-target aria-label="Copy roger share --curated"><code>roger share --curated</code></button>.</dd>
          </div>
          <div>
            <dt>What the 10% is</dt>
            <dd>The fee is the product: the upstream sees <b>a station, never a person</b>,
                every request gets a receipt, a dropped machine fails over mid-request, and
                your cap holds on every call. <a href="#why">The whole trade, above.</a></dd>
          </div>
          <div>
            <dt>Through a Tower</dt>
            <dd><b>90 / 5 / 5</b>: the serving node, the Tower operator, the platform.
                <a href="/broadcasts-run-a-tower.html">How that path works</a>.</dd>
          </div>
          <div>
            <dt>You control</dt>
            <dd>Rates, free hours, time-of-use windows, concurrency, which models. A rate of
                zero is normal.</dd>
          </div>
          <div>
            <dt>Minimum payout</dt>
            <dd><b>$25.</b></dd>
          </div>
          <div>
            <dt>Schedule</dt>
            <dd>On request, released in a monthly batch.</dd>
          </div>
          <div>
            <dt>Identity</dt>
            <dd>Stripe Connect onboarding before any payout leaves. Stripe holds the
                documents; <b>we never store them</b>.</dd>
          </div>
        </dl>

        <!-- The life of an earning, drawn (HoldFigure.svelte, hold-figure.css): the lifecycle
             list is the figure's step list; data-scrub-auto plays it until the reader touches it. -->
        <HoldFigure fig={3}>
        <ol class="lifecycle" data-scrub data-scrub-auto aria-label="The life of an earning, accrued to paid">
          <li><b>Accrued</b><span>A served request credits your share the moment it settles.</span></li>
          <li><b>Held</b><span>The principal for 30 days; a 10% reserve slice stays back until day 90.</span></li>
          <li><b>Payable</b><span>Hold cleared, no open dispute, Connect onboarding done.</span></li>
          <li><b>Paid</b><span>You request a payout and Stripe transfers it to your bank.</span></li>
        </ol>
        </HoldFigure>
        <p class="sec-more mono"><a href="/manual.html#sMoney-earn">The payout rules in full, in the manual &rarr;</a></p>
      </div>
    </section>

    <!-- The calculator. Every number in it is one of three things: the reader's own
         input, a policy constant of ours read from the Go that enforces it, or live
         market data. Never a third party's price - naming one would commit this page to
         keeping somebody else's rate sheet current, on the one page whose whole argument
         is that we do not publish prices we do not control.
         Contract: test/calculator.test.mjs. -->
    <!-- INK PANEL: the page's one moment, the two sums (components.css .tone-zone). -->
    <div class="tone-zone" data-tone="ink">
    <section class="section" id="calc">
      <div class="wrap">
        <div class="section__head section__head--left" data-reveal>
          <span class="sectionno">&sect;5 / THE ARITHMETIC</span>
          <h2>Two sums, both of which show their working.</h2>
          <p>Neither is a forecast. The numbers you type stay in your browser.</p>
        </div>

        <div class="calc" data-reveal>
          <article class="calc__card tint-panel" data-calc="operator">
            <header class="calc__head">
              <h3>What could a card earn?</h3>
              <span class="calc__tag">estimate</span>
            </header>
            <p class="calc__lead">Capacity at a price you set. Nobody can promise traffic, so
               <b>how busy</b> is yours to choose.</p>

            <div class="calc__field">
              <label for="opRate">Your rate, $ per 1M output tokens</label>
              <input type="number" id="opRate" value="0.30" min="0" max="100" step="0.01" inputmode="decimal" />
            </div>
            <div class="calc__field">
              <label for="opTps">Output tokens per second, aggregate across concurrent requests</label>
              <input type="number" id="opTps" value="300" min="1" max="100000" step="1" inputmode="numeric" />
            </div>
            <div class="calc__field">
              <label for="opHours">Hours on air per day</label>
              <input type="number" id="opHours" value="8" min="0" max="24" step="0.5" inputmode="decimal" data-range-twin />
            </div>
            <div class="calc__field">
              <label for="opBusy">How busy, % of on-air time actually serving</label>
              <input type="number" id="opBusy" value="25" min="0" max="100" step="1" inputmode="numeric" data-range-twin />
            </div>

            <output class="calc__out" id="opOut" for="opRate opTps opHours opBusy" aria-live="polite">
              <span class="calc__big" id="opShare">$17.50</span>
              <span class="calc__unit">a month, your share</span>
            </output>
            <p class="calc__formula" data-calc-formula id="opFormula">
              300 tok/s &times; 8 h &times; 25% &times; 30 days = 64.8M tokens &middot;
              &times; $0.30 / 1M = $19.44 gross &middot; your 90% = $17.50
            </p>
            <p class="calc__fine">Thirty-day month. The platform keeps 10%. A station nobody
               tunes to earns nothing, however fast the card.</p>
          </article>

          <article class="calc__card tint-panel" data-calc="consumer">
            <header class="calc__head">
              <h3>What are you paying now?</h3>
              <span class="calc__tag">your numbers</span>
            </header>
            <p class="calc__lead">Type what you pay today; it is re-priced against the band,
               read live.</p>

            <div class="calc__field">
              <label for="cnPrice">What you pay today, $ per 1M output tokens</label>
              <input type="number" id="cnPrice" value="0.60" min="0" max="1000" step="0.01" inputmode="decimal" />
            </div>
            <div class="calc__field">
              <label for="cnVolume">Output tokens a month, in millions</label>
              <input type="number" id="cnVolume" value="20" min="0" max="100000" step="1" inputmode="numeric" />
            </div>
            <div class="calc__field">
              <label for="cnHere">Rate on the band, $ per 1M out</label>
              <!-- SHIPS EMPTY, deliberately. This one field is live data, not a reader
                   input, and the build has no honest value for it - a plausible default
                   here would be exactly the invented rate the page forbids. Until the
                   market is read (or the reader types one), the sum below says so. -->
              <input type="number" id="cnHere" value="" min="0" max="100" step="any"
                     inputmode="decimal" placeholder="set a rate" />
              <span class="calc__hint" id="cnHereNote">reading the live market&hellip;</span>
            </div>

            <output class="calc__out" id="cnOut" for="cnPrice cnVolume cnHere" aria-live="polite">
              <span class="calc__big" id="cnDelta">set a rate</span>
              <!-- The unit is hidden while the answer is a prompt, so the live region
                   does not announce "set a rate a month, difference". -->
              <span class="calc__unit" id="cnUnit" hidden>a month, difference</span>
            </output>
            <p class="calc__formula" data-calc-formula id="cnFormula">
              20M &times; $0.60 / 1M = $12.00 today &middot;
              put a band rate in to compare
            </p>
            <p class="calc__fine">The band rate is the cheapest <b>paid</b> band's output
               price when the page loaded, named above. Free bands are left out, since a
               zero would price the comparison at nothing.</p>
          </article>
        </div>

        <noscript>
          <p class="calc__fine">These two work out in the browser, so with JavaScript off
             they stay as the worked example above: 300 tok/s at $0.30 for eight hours a
             day, a quarter of it busy, is about $17.50 a month of operator share. Small,
             and real - per-token earnings scale with tokens actually served, not with the
             card. The second sum needs a live band rate and cannot have one without
             JavaScript, so it stays unanswered rather than guessing. The rules both sums
             use are all printed on this page.</p>
        </noscript>
      </div>
    </section>
    </div><!-- /tone-zone -->

    <section class="section" id="free">
      <div class="wrap">
        <div class="section__head section__head--left" data-reveal>
          <span class="sectionno">&sect;6 / WHAT IS NOT BILLED</span>
          <h2>Our support and dedication to our service.</h2>
        </div>
        <ul class="freelist" data-reveal>
          <li><b>Reading the dial.</b> The whole live directory, every station and every price, with no account.</li>
          <li><b>The software.</b> The CLI, the terminal app, the iOS and Mac app, and the source. There is no paid edition.</li>
          <li><b>Serving for free.</b> Running a station at a rate of zero costs nothing and is not rate-limited into uselessness.</li>
          <li><b>A standalone Tower.</b> Your own relay, your own machines, your own network, with no account and nothing reported anywhere.</li>
          <li><b>A private band.</b> Handing your own people a key to your own station. <a href="/manual.html#s5">How grants work</a>.</li>
          <li><b>An answer that did not work.</b> Voided, not discounted.</li>
        </ul>
      </div>
    </section>

    <section class="section research-closing" id="ask">
      <div class="wrap">
        <div class="section__head" data-reveal>
          <span class="sectionno">&sect;7 / STILL UNSURE</span>
          <h2>We take pride in privacy and service, let us know your concerns.</h2>
          <p>Can an operator read your prompts? What if a station drops mid-answer? Do you
             open a port? In short: no, you get refunded, and no, but more in the FAQ.</p>
        </div>
        <div class="research-actions research-actions--lead">
          <a class="research-button" href="/faq.html">Read the FAQ <span aria-hidden="true">&rarr;</span></a>
          <a class="research-button" href="/manual.html#s6">The manual on money</a>
          <a class="research-button" data-lets-talk aria-expanded="false"
             href="mailto:labs@rogerai.fm?subject=Pricing%20question">Let&rsquo;s talk</a>
        </div>
      </div>
    </section>

  </main>

  <Footer anchors='/' />
