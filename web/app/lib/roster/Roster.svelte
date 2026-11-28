<script>
  // A live roster of stations on air: the voices (/voices.html) and the search stations (/search.html). One
  // component (2026-11-26); /voices was a 330-line page script (src/js/voices.js) and its rules are kept:
  //   real data only: an empty answer is an honest quiet band, never an invented row;
  //   a transient failure HOLDS the last roster (a 429 carries no rows, and reading it as empty made the
  //   table flicker), for at most HOLD_MAX polls, then says the broker could not be reached;
  //   a 429's Retry-After paces the next poll only (when the broker exposes that header to the page; without
  //   it the poll keeps its 30s); no background polling under reduced motion;
  //   the first read waits until the roster is in view.
  // The search box, the sort and the filters are the Models page's tuner (tuner.css), over the rows in hand.
  // Picking a row writes it into the one command under the table. No node address is in either feed.
  import { onMount } from "svelte";
  import { BROKER } from "#lib/api.js";
  import { KINDS, SORTS, keyed, latency, usd, view } from "./kinds.js";

  let { kind } = $props();
  const K = $derived(KINDS[kind]);
  const POLL_MS = 30000, HOLD_MAX = 12;   // 12 holds at 30s bound a held roster to about six minutes

  let rows = $state([]);
  let mode = $state("loading");   // loading | live | holding | quiet | off
  let q = $state(""), sort = $state("cheapest"), free = $state(false), pick = $state(""), picked = $state("");
  let root, reload = () => {};

  const shown = $derived(view(rows, { q, sort, free, pick }));
  const picks = $derived([...new Set(rows.map((r) => r.pick).filter(Boolean))].sort());
  const filtered = $derived(!!(q.trim() || free || pick));
  const status = $derived(
    mode === "loading" ? "tuning in to the broker…"
    : mode === "holding" ? "holding the last roster · re-reading…"
    : mode === "off" ? "couldn't reach the broker just now"
    : mode === "quiet" ? `the band is quiet. ${K.none}`
    : `${rows.length} ${rows.length === 1 ? K.one : K.many} listed · live from ${K.path}`);
  const ref = $derived(rows.some((r) => r.ref === picked) ? picked : "");   // a picked station that left the air is no longer named

  onMount(() => {
    const reduced = matchMedia("(prefers-reduced-motion: reduce)").matches;
    let timer, visible = false, inflight = false, loaded = false, holds = 0, retryMs = 0;
    const fail = () => {
      if (rows.length && holds < HOLD_MAX) { holds++; mode = "holding"; }
      else { rows = []; mode = "off"; }
    };
    async function load() {
      if (inflight) return;
      inflight = true;
      const ctrl = new AbortController(), to = setTimeout(() => ctrl.abort(), 8000);
      try {
        const r = await fetch(BROKER + K.path, { signal: ctrl.signal, cache: "no-store" });
        loaded = true;
        if (!r.ok) { const s = parseInt(r.headers?.get?.("Retry-After"), 10); retryMs = s > 0 ? s * 1000 : 0; fail(); return; }
        const data = await r.json();
        holds = 0;
        rows = keyed((Array.isArray(data?.[K.key]) ? data[K.key] : []).map(K.row).filter((x) => x.ok));
        mode = rows.length ? "live" : "quiet";
      } catch { fail(); }
      finally { clearTimeout(to); inflight = false; }
    }
    function schedule() {
      if (reduced) return;
      clearTimeout(timer);
      const wait = Math.max(POLL_MS, retryMs); retryMs = 0;
      timer = setTimeout(async () => { if (visible) await load(); schedule(); }, wait);   // the next wait is set once this read has answered, so a 429 paces the very next one
    }
    reload = () => load().then(schedule);   // a refresh by hand restarts the wait too, so a 429 it draws paces the next read
    const io = new IntersectionObserver((es) => es.forEach((e) => {
      visible = e.isIntersecting;
      if (visible && !loaded) load().then(schedule);
    }));   // any part in view counts: a long roster is still being read when less than a sixth of it fits the screen
    io.observe(root);
    return () => { io.disconnect(); clearTimeout(timer); };
  });
  function refresh() { if (rows.length) mode = "holding"; reload(); }
  function clear() { q = ""; free = false; pick = ""; }
</script>
<div class="roster" bind:this={root}>
  <div class="tuner">
    <div class="tuner__search">
      <span class="tuner__prompt mono" aria-hidden="true">/</span>
      <input type="search" id="rosterSearch" class="tuner__input mono" placeholder={K.placeholder} autocomplete="off" spellcheck="false" aria-label={"Search " + K.many} bind:value={q} />
      <button class="tuner__clear" id="rosterClear" type="button" aria-label="Clear the search and filters" hidden={!filtered} onclick={clear}>clear</button>
    </div>
    <div class="tuner__filters" role="group" aria-label="Filters and sort">
      <label class="tuner__sel">
        <span class="tuner__k">SORT</span>
        <select id="rosterSort" class="mono" aria-label={"Sort " + K.many} bind:value={sort}>
          {#each SORTS as [id, label]}<option value={id}>{label}</option>{/each}
        </select>
      </label>
      <label class="tuner__sel">
        <span class="tuner__k">{K.pickLabel}</span>
        <select id="rosterPick" class="mono" aria-label={K.pickAll} bind:value={pick}>
          <option value="">{K.pickAll}</option>
          {#each picks as p}<option value={p}>{p}</option>{/each}
        </select>
      </label>
      <button class="tuner__chip" class:is-on={free} id="rosterFree" type="button" aria-pressed={free} onclick={() => (free = !free)}>free now</button>
    </div>
  </div>

  <div class="bands-panel">
    <div class="bands-panel__bar">
      <span class="bands-panel__status" class:is-live={mode === "live" || mode === "holding"} class:is-quiet={mode === "quiet"} class:is-off={mode === "off"} id="rosterStatus">
        <span class="beacon" aria-hidden="true">((<span class="beacon__dot"></span>))</span>
        <span id="rosterStatusText">{status}</span>
      </span>
      <button class="market__refresh" id="rosterRefresh" type="button" aria-label="Refresh the roster" onclick={refresh}>refresh</button>
    </div>

    <div class="voice-row voice-row--head" aria-hidden="true">
      <span class="voice-cell voice-cell--name">{K.head[0]}</span>
      <span class="voice-cell voice-cell--lang">{K.head[1]}</span>
      <span class="voice-cell voice-cell--lat">{K.head[2]}</span>
      <span class="voice-cell voice-cell--price">$ / 1k {K.unit}</span>
    </div>

    <ul class="voice-list" id="rosterList" aria-live="polite">
      {#if mode === "loading"}
        {#each [0, 1, 2, 3] as _}<li class="voice-skel" aria-hidden="true"></li>{/each}
      {:else}
        {#each shown as r, i (r.key)}
          <li class="voice-row" class:is-picked={r.ref && r.ref === ref} style:--i={i}>
            <span class="voice-cell voice-cell--name">
              <span class="voice-name-line"><span class="voice-dot voice-dot--on" aria-hidden="true">◉</span>
                {#if r.ref}<button type="button" class="voice-name voice-pick" aria-pressed={r.ref === ref} onclick={() => (picked = r.ref === ref ? "" : r.ref)}>{r.name}</button>{:else}<span class="voice-name">{r.name}</span>{/if}</span>
              {#if r.op}<span class="voice-op mono">by @{r.op}</span>{/if}
              {#if r.sub}<span class="voice-sub">{r.sub}</span>{/if}
            </span>
            <span class="voice-cell voice-cell--lang">{#if r.pick}<span class="voice-lang mono">{r.pick}</span>{:else}<span class="voice-unit--idle">-</span>{/if}</span>
            <span class="voice-cell voice-cell--lat">{#if r.latency}<b class="mono voice-lat">{latency(r.latency)}</b>{:else}<span class="voice-unit--idle">-</span>{/if}</span>
            <span class="voice-cell voice-cell--price">{#if r.free}<span class="band-tag band-tag--free">FREE</span>{:else}<b class="mono ember">{usd(r.price)}</b><span class="voice-unit"> /1k</span>{/if}</span>
          </li>
        {/each}
      {/if}
    </ul>

    <div class="bands-panel__quiet" id="rosterQuiet" hidden={!(mode === "quiet" || mode === "off")}>
      <p class="bands-panel__quiet-head mono"><span class="beacon" aria-hidden="true">((<span class="beacon__dot"></span>))</span>
        {mode === "off" ? "Couldn't reach the broker. No live roster to show." : `The band is quiet. ${K.none}`}</p>
      <p class="bands-panel__quiet-sub mono">{K.shareWhat} <code class="inline">{K.share}</code></p>
    </div>
    <p class="bands-panel__empty mono" id="rosterEmpty" hidden={!(rows.length && !shown.length)}>no {K.many} match. Widen the search or clear the filters.</p>

    <p class="market__foot mono" id="rosterFoot">
      <span class="ember">◉</span> on air · <span class="band-tag band-tag--free">FREE</span> no charge right now ·
      {K.meter} · prices in $ / 1k {K.unit} · real data only · refreshes every 30s
    </p>
  </div>

  <div class="voice-tune">
    <span class="fig" aria-hidden="true">FIG. 1 - {K.fig}</span>
    <p class="voice-tune__lead mono">{K.lead}</p>
    <button class="install__box voice-tune__cmd" id="rosterCmd" type="button" data-os-lock="any" aria-label={K.copy}>
      <span class="install__prompt" aria-hidden="true">$</span>
      <code class="install__code" id="rosterCmdCode">{K.cmd(ref)}</code>
      <span class="install__copy" aria-hidden="true">copy</span>
    </button>
    <p class="voice-tune__note mono">{K.note}</p>
  </div>
</div>
