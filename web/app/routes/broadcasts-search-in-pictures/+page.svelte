<script>
  // How search works on RogerAI, in pictures.
  // A broadcast on the one post layout (lib/ui/BroadcastPost.svelte). Few words, many plates.
  // Every claim comes from docs/search-api.md and docs/search-providers.md; the tests in
  // test/broadcast-search.test.mjs hold them.
  import BroadcastPost from "#lib/ui/BroadcastPost.svelte";
  import Faq from "#lib/ui/Faq.svelte";
  import Signoff from "#lib/ui/Signoff.svelte";
  const post = {
    slug: "broadcasts-search-in-pictures",
    number: "017",
    title: "How search works on RogerAI, in pictures.",
    desc: "How the RogerAI Search API works: one query, one search station, one engine. How a search station differs from a model, where the index comes from, what we keep, and what a search costs. Illustrated, in plain words.",
    dek: "One query in. One station answers. You can see whose index it came from.",
    date: "2026-11-26",
    dateLabel: {"text":"NOVEMBER 2026","rev":"2026.11"},
    ogimage: "https://rogerai.fm/assets/broadcasts/search-og.png",
    ogdesc: "One query in, one station answers, and the listing names the engine. How search stations work on RogerAI, where the index comes from, and what we keep. An illustrated field guide.",
    css: ["broadcast-search.css"],
    faq: [
      {"q":"Does RogerAI have its own search index?","a":"No. RogerAI does not crawl the web. Each search station brings its own engine: a Brave account, a SearXNG instance, or an index its operator built. Every listing names the engine."},
      {"q":"Does a search station see who is searching?","a":"No. The station gets a pseudonym that is different for every station. It never gets your account, credentials, email or IP address."},
      {"q":"Does RogerAI keep my searches?","a":"The broker keeps no query text. It records the station, the time taken, the result count and the cost. The one exception is material flagged as child sexual abuse, which is preserved as the law requires."},
      {"q":"How is a search station different from a language model?","a":"A model writes new text and is paid per token. A search station returns links to pages that exist and is paid per query. An agent can use both: search first, then read, then answer."},
      {"q":"What happens if a search fails?","a":"You pay nothing. A request that is refused, fails, times out, or has its results withheld by the content screen is not charged."},
      {"q":"Can I run my own search engine on RogerAI?","a":"Yes. Sign in with roger login, then share your engine with roger share --search. A native station serves your own index over the open search/1 contract."},
    ],
  };
</script>
<BroadcastPost {...post} renderFaq={false} signoff={false}>
<!-- Broadcast 017: the Search API, in pictures. Tests: test/broadcast-search.test.mjs, test/voice.test.mjs. -->
  {#snippet meta()}
        <div class="bc-post__meta mono">
          <span>One request: <b>one query</b></span>
          <span>The index: <b>the station's own</b></span>
          <span>Reading: <b>mostly pictures</b></span>
        </div>
  {/snippet}

      <figure class="figure figure--plate srch-fig" data-reveal>
        <video class="srch-loop" width="832" height="464" autoplay muted loop playsinline preload="metadata"
               poster="assets/broadcasts/search-hero-800.webp"
               aria-label="A switchboard operator in headphones slowly pushes a cord with a red plug into one jack of a tall switchboard, connecting one caller to one station.">
          <source src="assets/broadcasts/search-hero-loop.mp4" type="video/mp4" />
        </video>
        <img class="srch-still" src="assets/broadcasts/search-hero.webp" srcset="assets/broadcasts/search-hero-800.webp 800w, assets/broadcasts/search-hero.webp 1600w" sizes="(max-width: 1000px) 100vw, 960px" width="1600" height="896"
             alt="A vintage operating-manual drawing of a switchboard operator in headphones holding a cord with a red plug, connecting one call to one jack among many." loading="lazy" />
        <figcaption class="fig mono">FIG. 1 &middot; ONE CALL, PATCHED THROUGH TO ONE STATION.</figcaption>
      </figure>

      <nav class="toc-tuner" data-tuner aria-label="Sections">
        <div class="wrap toc-tuner__band">
        <div class="toc-tuner__readout" aria-hidden="true"></div>
        <ol class="toc-tuner__scale">
          <li><a class="toc-tuner__st" href="#s1"><b>&sect;1</b><span class="toc-tuner__name">THE TRIP</span></a></li>
          <li><a class="toc-tuner__st" href="#s2"><b>&sect;2</b><span class="toc-tuner__name">NOT A MODEL</span></a></li>
          <li><a class="toc-tuner__st" href="#s3"><b>&sect;3</b><span class="toc-tuner__name">THE INDEX</span></a></li>
          <li><a class="toc-tuner__st" href="#s4"><b>&sect;4</b><span class="toc-tuner__name">WHAT WE KEEP</span></a></li>
          <li><a class="toc-tuner__st" href="#s5"><b>&sect;5</b><span class="toc-tuner__name">THE PRICE</span></a></li>
          <li><a class="toc-tuner__st" href="#s6"><b>&sect;6</b><span class="toc-tuner__name">READ, DON'T OBEY</span></a></li>
          <li><a class="toc-tuner__st" href="#s7"><b>&sect;7</b><span class="toc-tuner__name">RUN YOUR OWN</span></a></li>
          <li><a class="toc-tuner__st" href="#s8"><b>&sect;8</b><span class="toc-tuner__name">FAQ</span></a></li>
        </ol>
        <span class="toc-tuner__needle" aria-hidden="true"></span>
        </div>
      </nav>

      <blockquote class="bc-answer" data-reveal>
        <b>Quick answer.</b> You send one query to the RogerAI Search API. The broker picks one
        search station. That station asks its own engine and sends results back in one shape.
        You pay that station's price for one query. RogerAI does not crawl the web. Each
        listing names the engine, so you know whose index you read.
      </blockquote>

      <div class="bc-post__body">

        <p class="bc-post__lead" data-reveal>
          Someone asked how our search works, and where the index comes from. Short answer:
          the stations bring it. Here is the rest, mostly in pictures.
        </p>

        <!-- §1 the trip -->
        <section class="bc-sec" id="s1" data-reveal>
          <span class="bc-sec__no">&sect;1 &middot; The trip</span>
          <h2>What happens when an app sends a search?</h2>
          <figure class="figure figure--plate">
            <div class="srch-route" role="img" aria-label="Diagram: a query goes from your app to the RogerAI broker, then to one search station chosen from several, then to that station's engine. The results come back along the same path, and the broker keeps no query text.">
              <ol class="srch-route__steps" aria-hidden="true">
                <li><b>Your app</b><small>one query</small></li>
                <li><b>Broker</b><small>picks one station</small></li>
                <li class="srch-route__live"><b>One station</b><small>sees a pseudonym</small></li>
                <li><b>Its engine</b><small>Brave, SearXNG or its own</small></li>
              </ol>
              <span class="srch-route__dot" aria-hidden="true"></span>
            </div>
            <figcaption class="fig mono">FIG. 2 &middot; OUT TO ONE STATION AND BACK. RESULTS COME HOME IN ONE SHAPE.</figcaption>
          </figure>
          <p>
            Your app calls <code>POST /v1/search</code>. With no station named, the broker
            picks one the same way it picks a chat station. Each result is a title, a link,
            a snippet and a date when there is one.
          </p>
        </section>

        <!-- §2 not a model -->
        <section class="bc-sec" id="s2" data-reveal>
          <span class="bc-sec__no">&sect;2 &middot; Not a model</span>
          <h2>How is a search station different from a model?</h2>
          <figure class="figure figure--plate">
            <img src="assets/broadcasts/search-desks.webp" srcset="assets/broadcasts/search-desks-800.webp 800w, assets/broadcasts/search-desks.webp 1600w" sizes="(max-width: 1000px) 100vw, 960px" width="1600" height="896"
                 alt="Two desks in one room: on the left a man types at a typewriter, writing from memory; on the right a woman pulls one red index card from a tall card catalog." loading="lazy" />
            <figcaption class="fig mono">FIG. 3 &middot; ONE WRITES FROM MEMORY. ONE LOOKS IT UP.</figcaption>
          </figure>
          <figure class="figure figure--plate">
            <div class="scroll-box">
            <table class="data-table bc-table srch-vs">
              <thead><tr><th scope="col"></th><th scope="col">Model station</th><th scope="col">Search station</th></tr></thead>
              <tbody>
                <tr><th scope="row">Answers with</th><td>new text</td><td>links to pages that exist</td></tr>
                <tr><th scope="row">Paid per</th><td>token</td><td>query</td></tr>
                <tr><th scope="row">Listed on</th><td><code>/v1/<wbr>models</code></td><td><code>/search/<wbr>stations</code></td></tr>
              </tbody>
            </table>
            </div>
            <figcaption class="fig mono">FIG. 4 &middot; SAME MARKET, SAME WALLET, DIFFERENT WORK.</figcaption>
          </figure>
          <p>An agent can use both. Search first, read the pages, then answer.</p>
        </section>

        <!-- §3 the index -->
        <section class="bc-sec" id="s3" data-reveal>
          <span class="bc-sec__no">&sect;3 &middot; The index</span>
          <h2>Where does the search index come from?</h2>
          <figure class="figure figure--plate">
            <img src="assets/broadcasts/search-courier.webp" srcset="assets/broadcasts/search-courier-800.webp 800w, assets/broadcasts/search-courier.webp 1600w" sizes="(max-width: 1000px) 100vw, 960px" width="1600" height="896"
                 alt="A courier on a bicycle with a red satchel rides past three library buildings on one street, collecting from each." loading="lazy" />
            <figcaption class="fig mono">FIG. 5 &middot; EVERY STATION BRINGS ITS OWN LIBRARY.</figcaption>
          </figure>
          <p>RogerAI does not crawl the web. Each station brings an engine, of one of three kinds:</p>
          <ul class="bc-dial">
            <li><b>brave</b> Brave's Search API, on the operator's own Brave account.</li>
            <li><b>searxng</b> A SearXNG instance. It asks several engines and merges what they return.</li>
            <li><b>native</b> An index the operator built, served over an open contract called <code>search/1</code>. It can sit in front of Meilisearch, Typesense, OpenSearch, a YaCy peer or a home-made crawler.</li>
          </ul>
          <p>Every listing names its engine. You know whose index you are reading before you send a word.</p>
        </section>

        <!-- §4 what we keep -->
        <section class="bc-sec" id="s4" data-reveal>
          <span class="bc-sec__no">&sect;4 &middot; What we keep</span>
          <h2>Who sees my search?</h2>
          <figure class="figure figure--plate">
            <img src="assets/broadcasts/search-envelope.webp" srcset="assets/broadcasts/search-envelope-800.webp 800w, assets/broadcasts/search-envelope.webp 1600w" sizes="(max-width: 1000px) 100vw, 960px" width="1600" height="896"
                 alt="A clerk takes a sealed envelope with a red wax seal through a small window. He sees the envelope, not the person who sent it." loading="lazy" />
            <figcaption class="fig mono">FIG. 6 &middot; THE STATION GETS THE LETTER, NOT THE NAME.</figcaption>
          </figure>
          <ul class="bc-dial">
            <li><b>The broker</b> keeps no query text. It records the station, the time, the result count and the cost.</li>
            <li><b>The station</b> gets a pseudonym, different for every station. Never your account, your credentials, your email or your IP.</li>
            <li><b>The one exception.</b> Material flagged as child sexual abuse is preserved, as the law requires.</li>
          </ul>
        </section>

        <!-- §5 the price -->
        <section class="bc-sec" id="s5" data-reveal>
          <span class="bc-sec__no">&sect;5 &middot; The price</span>
          <h2>What does a search cost?</h2>
          <figure class="figure figure--plate">
            <img src="assets/broadcasts/search-coin.webp" srcset="assets/broadcasts/search-coin-800.webp 800w, assets/broadcasts/search-coin.webp 1600w" sizes="(max-width: 1000px) 100vw, 960px" width="1600" height="896"
                 alt="A hand drops a single red coin into the slot of a small wooden box on a counter." loading="lazy" />
            <figcaption class="fig mono">FIG. 7 &middot; ONE QUERY, ONE COIN. A FAILED CALL COSTS NOTHING.</figcaption>
          </figure>
          <ul class="bc-dial">
            <li><b>One request is one query</b>, for 1 to 20 results.</li>
            <li><b>Stations set the price</b> per 1,000 queries, up to $50 per 1,000.</li>
            <li><b>With no station named</b>, the broker stays at or under $10 per 1,000. You can raise or lower that.</li>
            <li><b>You pay nothing</b> if the request is refused, the station fails, times out, or its results are withheld.</li>
            <li><b>Zero results</b> still counts. The station did the work.</li>
            <li><b>Without a login</b>, only free stations answer. A paid station needs you signed in.</li>
          </ul>
        </section>

        <!-- §6 read, don't obey -->
        <section class="bc-sec" id="s6" data-reveal>
          <span class="bc-sec__no">&sect;6 &middot; Read, don't obey</span>
          <h2>Can an agent trust what it finds?</h2>
          <p>
            Results are text from the open web. Anyone can title a page anything. An agent
            should read results as information, never as instructions.
          </p>
        </section>

        <!-- §7 run your own -->
        <section class="bc-sec" id="s7" data-reveal>
          <span class="bc-sec__no">&sect;7 &middot; Run your own</span>
          <h2>How do I put my own search engine on the market?</h2>
          <figure class="figure figure--plate srch-fig">
            <video class="srch-loop" width="832" height="464" autoplay muted loop playsinline preload="none"
                   poster="assets/broadcasts/search-own-800.webp"
                   aria-label="A man at a desk at home turns a dial on a small transmitter under a red desk lamp, his own card catalog beside him.">
              <source src="assets/broadcasts/search-own-loop.mp4" type="video/mp4" />
            </video>
            <img class="srch-still" src="assets/broadcasts/search-own.webp" srcset="assets/broadcasts/search-own-800.webp 800w, assets/broadcasts/search-own.webp 1600w" sizes="(max-width: 1000px) 100vw, 960px" width="1600" height="896"
                 alt="A man at a desk at home beside his own small card catalog, with a transmitter and a red desk lamp, cables running between them." loading="lazy" />
            <figcaption class="fig mono">FIG. 8 &middot; YOUR OWN CATALOG, ON THE AIR.</figcaption>
          </figure>
          <div class="code-block"><pre><code># roger 7.2.0 or later
brew upgrade roger      # or: curl -fsSL https://rogerai.fm/install.sh | sh
roger login
roger payout onboard    # once, to be paid

# SearXNG: add json to search.formats in settings.yml first
roger share --search searxng --upstream http://127.0.0.1:8888

# Brave: reads BRAVE_SEARCH_API_KEY, or asks with input hidden
roger share --search brave --price-per-1k 5

# Your own index: check it, then share it
roger search check --upstream http://127.0.0.1:9000
roger share --search native --upstream http://127.0.0.1:9000 --scope "Go module docs"</code></pre></div>
          <p>
            A search station needs a signed-in owner, so every search listing reads
            <code>operator_status: verified</code>. A focused index that finds nothing for a
            test word still passes the check.
          </p>
          <p>
            Add <code>--private</code> and roger prints a frequency code once, on your screen.
            The station is on no listing and answers only to that code, sent in the
            <code>X-Roger-Freq</code> header. If you lose it, <code>roger bands new-code</code>
            gives the band a new one. The free plan allows one private band. A paid private
            station still charges your wallet.
          </p>
          <p>
            To see the search stations on the air, run <code>roger search --kind search</code>.
            Every flag is in the <a href="https://rogerai.fm/manual#sSearch">manual</a> and the
            <a href="https://github.com/rogerai-fm/roger/blob/main/docs/search-providers.md">provider guide</a>.
          </p>
          <p>Calling search from code:</p>
          <div class="code-block"><pre><code>curl https://broker.rogerai.fm/v1/search \
  -H "Authorization: Bearer $ROGER_KEY" \
  -d '&#123;"query": "tallest building in the world", "count": 5&#125;'</code></pre></div>
        </section>

        <!-- §8 faq -->
        <Faq no={8} items={post.faq} title="Questions people actually ask" />

        <Signoff lead={"<p>To try it, install the set, then <code class=\"cmd-words\"><span>roger</span> <span>login</span></code>.</p>"} tail={"<p>If the market is new to you, start with <a href=\"/broadcasts-how-llms-work.html\">how LLMs work, in pictures</a>.</p>"}>curl <span class="tok">-fsSL</span> <span class="tok">https://rogerai.fm/install.sh</span> <span class="tok">| sh</span></Signoff>

      </div>
</BroadcastPost>
