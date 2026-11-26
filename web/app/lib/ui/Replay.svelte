<script>
  // The recorded timelapse (2026-11-26, founder ask: start from the story, show the feed, show how long it
  // took and why). The hero is the sensor's own feed, one recorded frame a minute, with its clock and one
  // sentence at each moment that matters. Under it a time axis: the reading as a trace against its line,
  // a playhead, and the range that scrubs it. Under that, a drawing a third of the size: the page's own
  // scene (the unit as designed), whose parts show what each model said at the playhead. Scenarios are
  // separate and switched by a button each, so two stories are never mixed. Reusable: a page supplies a
  // data file (the schema is src/data/fire-replay.json, written from a saved record by an exporter) and
  // a scene.
  //
  // Every frame carries its own state as data attributes; the script (src/js/replay.js) is a clock that
  // shows a frame: it turns the frame on, writes its clock, its axis note and each part's words into the
  // labels, and sets the figure's attributes: data-state, data-alert, data-over, data-look, and one data-<name>
  // for every entry of the frame's `attrs` (the scene's own vocabulary), which the scene's sheet reads. Nothing loops. With no script, and under reduced motion until a reader moves the range,
  // the still is the frame the data names, and the story is the list under "Read it as text".
  //
  // Contract: runes, no <style> (replay.css, lead .rp), no colour literal, no inline style, every word
  // in HTML, no SVG <text>. A part's colour is its tone (blue, nano, pico, micro: the tier tokens); the
  // alert red is only a reading over its line, an alert state and a person told.
  let { data, fig = 1, kind = "", scene } = $props();
  const d = $derived(data);
  const sc0 = $derived(d.scenarios.find((s) => s.id === d.still[0]));
  const f0 = $derived(sc0.frames[d.still[1]]);
  const cap0 = $derived([...sc0.frames.slice(0, d.still[1] + 1)].reverse().find((f) => f.caption)?.caption || "");
  const id = $derived(`rp-cap-${fig}`);
  const pts = (t) => t.map((p) => `${p[0]},${p[1]}`).join(" ");
</script>

<figure class="figure figure--plate rp{kind ? ` ${kind}` : ''}" data-replay data-scn={sc0.id} data-state={f0.state} data-alert={f0.alert ? 1 : 0} data-over={f0.over ? 1 : 0}
  data-look={f0.look ? 1 : 0} {...Object.fromEntries(Object.entries(f0.attrs || {}).map(([k, v]) => [`data-${k}`, v]))}
  {...Object.fromEntries(Object.entries(d.words.states).map(([k, v]) => [`data-word-${k}`, v]))} role="group" aria-labelledby={id}>
  <div class="rp__head">
    <p class="rp__chips"><span class="rp__chip">{d.chips[0]}</span><span class="rp__chip rp__chip--plain">{d.chips[1]}</span></p>
    <div class="rp__tabs" role="group" aria-label="Scenario">
      {#each d.scenarios as s}<button type="button" class="rp__tab" data-scn={s.id} aria-pressed={s.id === sc0.id}>{s.title}</button>{/each}
    </div>
    <button class="rp__play" type="button" data-play={d.controls.play} data-pause={d.controls.pause} data-replay-label={d.controls.replay}>{d.controls.play}</button>
  </div>

  <div class="rp__feed">
    {#each d.scenarios as s}
      {#each s.frames as f, k}
        <img class="rp__frame" class:is-on={s.id === sc0.id && k === d.still[1]} data-scn={s.id} data-k={k} src={f.src} width={d.feed.size?.[0] || 800} height={d.feed.size?.[1] || 486} loading="lazy" decoding="async"
          alt={f.caption ? `${s.title}, ${f.clock}: ${f.caption}` : ""} data-clock={f.clock} data-plus={f.plus || ""} data-view={f.view || ""} data-axis={f.axis || ""} data-state={f.state}
          data-alert={f.alert ? 1 : 0} data-over={f.over ? 1 : 0} data-look={f.look ? 1 : 0} data-ms={f.ms} data-x={f.x} data-say={f.caption || ""}
          {...Object.fromEntries(Object.entries(f.attrs || {}).map(([k, v]) => [`data-a-${k}`, v]))}
          {...Object.fromEntries(Object.entries(f.says || {}).map(([k, v]) => [`data-r-${k}`, v]))} />
      {/each}
    {/each}
    <p class="rp__clock" aria-hidden="true"><b class="rp__time">{f0.clock}</b><span class="rp__plus">{f0.plus}</span><span class="rp__statew rp__statew--feed">{d.words.states[f0.state]}</span></p>
    {#if d.feed.look}<p class="rp__look" data-tone="nano" aria-hidden="true"><b>{d.feed.look}</b></p>{/if}
    {#each d.boxes || [] as b}<p class="rp__box" data-box={b.id} data-tone="blue" aria-hidden="true"><b>{b.label}</b></p>{/each}
    <p class="rp__feedlabel"><span class="rp__viewname">{f0.view || ""}</span>{d.feed.label}{#if d.feed.every} &middot; {d.feed.every}{/if}</p>
    <p class="rp__caption" aria-hidden="true"><span class="rp__cap">{cap0}</span></p>
  </div>

  <div class="rp__axis">
    {#each d.scenarios as s}
      <svg class="rp__trace" class:is-on={s.id === sc0.id} data-scn={s.id} viewBox="0 -8 100 116" preserveAspectRatio="none" aria-hidden="true" focusable="false">
        {#each s.lines || [] as y}<line class="rp__line" x1="0" x2="100" y1={y} y2={y} />{/each}
        {#each s.traces as t}<polyline class="rp__path" data-trace={t.id} data-tone={t.tone} points={pts(t.points)} />{/each}
        {#each s.events as e}<line class="rp__flag" data-state={e.state} x1={s.frames[e.k].x} x2={s.frames[e.k].x} y1="-8" y2="108" />{/each}
      </svg>
    {/each}
    <i class="rp__head-mark" aria-hidden="true"></i>
    <input class="rp__range" type="range" min="0" max={sc0.frames.length - 1} step="1" value={d.still[1]} aria-label="Time" />
    <p class="rp__axisnote" aria-hidden="true"><span class="rp__axistext">{f0.axis || ""}</span>{#each sc0.traces.filter((t) => t.label) as t}<span class="rp__key" data-tone={t.tone}><i></i>{t.label}</span>{/each}</p>
  </div>

  <div class="rp__strip" class:has-scene={scene}>
    {#if scene}<div class="rp__scene" aria-hidden="true"><div class="rp__draw">{@render scene()}
      {#each d.parts as p}
        <p class="rp__part" data-part={p.id} data-tone={p.tone}><b>{p.name}</b>{#if p.readout}<span class="rp__out" data-read={p.id}>{f0.says[p.id]}</span>{/if}{#if p.decides}<span class="rp__statew">{d.words.states[f0.state]}</span>{/if}{#if p.tag}<i class="rp__tagd">{p.tag}</i>{/if}</p>
      {/each}
      {#if d.chat}
        <div class="rp__chat" data-tone="micro">
          <p class="rp__q"><b>{d.chat.who}</b><span>{d.chat.q}</span></p>
          <p class="rp__a"><b>{d.chat.by} <i class="rp__tagd">{d.chat.tag}</i></b><span>{d.chat.a}</span></p>
        </div>
      {/if}
      </div>
      <p class="rp__unit">{d.unit}</p>
    </div>{/if}
    <ul class="rp__rows">
      {#each d.parts as p}<li data-tone={p.tone}><i aria-hidden="true"></i><b>{p.name}</b> <span>{p.role}{#if p.tag}{p.role ? ", " : ""}{p.tag}{/if}</span>{#if p.readout}<em class="rp__out" data-read={p.id}>{f0.says[p.id]}</em>{/if}</li>{/each}
    </ul>
  </div>

  <details class="rp__script">
    <summary>{d.controls.script}</summary>
    {#each d.scenarios as s}
      <h3>{s.title}, {s.date}</h3>
      <ol>{#each s.frames.filter((f) => f.caption) as f}<li>{f.clock}{f.plus ? ` (${f.plus})` : ""}: {f.caption} {d.words.states[f.state]}.</li>{/each}</ol>
    {/each}
    {#if d.chat}<p>{d.chat.who}: &ldquo;{d.chat.q}&rdquo; {d.chat.by} ({d.chat.tag}): &ldquo;{d.chat.a}&rdquo; {d.micro.rule}</p>{/if}
  </details>
  <p class="rp__note">{d.note}</p>
  {#if d.credit}<p class="wm-credit">{d.credit}</p>{/if}
  <figcaption class="fig" id={id}>FIG. {fig} - {d.caption}</figcaption>
</figure>
