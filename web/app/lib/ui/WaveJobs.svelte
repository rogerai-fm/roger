<script>
  // The job finder (2026-11-24): every job the Wave family is built for, grouped by setting, with the
  // smallest slot that can carry it. It lived in §3 of research-wave-family.html until the founder moved
  // the family page's tables to the manual (Appendix B5); the family page keeps a per-slot summary drawn
  // from the same data (#lib/waveJobs.js). Progressive enhancement, as before: the served table is
  // whole and grouped, the toolbar ships hidden, and src/js/wave-jobs.js reveals the filters, counts
  // what is on screen and bounds the panel. Styles: tier-table.css (.wj, .tier-matrix).
  import { LADDER, SETTINGS } from "#lib/waveJobs.js";
</script>

<div class="wj" data-wj>
  <div class="wj__bar" data-wj-bar hidden>
    <div class="wj__set" role="group" aria-label="Filter jobs by setting">
      <span class="wj__legend">Setting</span>
      <button class="wj__chip" type="button" data-wj-setting="all" aria-pressed="true">Everywhere</button>
      {#each SETTINGS as g}<button class="wj__chip" type="button" data-wj-setting={g.key} aria-pressed="false">{g.label}</button>{/each}
    </div>
    <div class="wj__pick">
      <label class="wj__legend" for="wj-slot">Slot</label>
      <select class="wj__select" id="wj-slot" data-wj-slot>
        <option value="all" selected>Any slot</option>
        {#each LADDER as s}<option value={s}>{s}</option>{/each}
      </select>
    </div>
    <p class="wj__count" data-wj-count role="status" aria-live="polite"></p>
  </div>
  <div class="wj__scroll" data-wj-scroll>
    <table class="data-table tier-matrix">
      <thead>
        <tr>
          <th scope="col">Job</th><th scope="col">Smallest slot</th>
          <th scope="col">In and out</th><th scope="col">Why it stays in-house</th>
        </tr>
      </thead>
      {#each SETTINGS as g}
        <tbody data-setting={g.key} data-label={g.label}>
          <tr class="wj__group"><th colspan="4" scope="colgroup"><b>{g.label}</b> <span>{g.sub}</span></th></tr>
          {#each g.jobs as j}
            <tr data-scope={j.scope}><th scope="row">{j.name}</th><td class="tier-cell">{j.slot}</td><td data-label="In and out">{j.io}</td><td data-label="Why it stays in-house">{j.why}</td></tr>
          {/each}
        </tbody>
      {/each}
      <tbody class="wj__nobody" data-wj-empty hidden>
        <tr><td colspan="4">Nothing carries both of those. Widen one of the filters.</td></tr>
      </tbody>
    </table>
  </div>
</div>
