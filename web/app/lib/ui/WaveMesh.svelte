<script>
  // HOW THE MESH WORKS: one explainer, on the family's parent page (research-models.html, since 2026-11-24; it was on each Wave model page) (legacy: _partials/wave-mesh.html).
  // no = the section number on the page; fig = its figure number; parent = the model the page is about.
  //
  // The figure (2026-11-23, in the isometric hand): a stage above the three word columns, in their
  // order. The three child nodes at the left, stacked in depth (the pump at the back, the valve, the
  // compressor in front); the broker in the middle, a tall thin slab with the log on its face as three
  // marks, one per record (the valve's dashed: it escalates; the compressor's heavy: the record that
  // decides); the parent console at the right. Three ways on the floor carry the records to the
  // broker, one carries the read to the parent. The one red at rest is the compressor's lamp. The
  // loop (wave-model.css, 10 s, once the figure has drawn itself in): each record runs to the broker
  // and writes its mark, the compressor lights and lifts, the parent reads. data-draw: draw.js (each
  // page's own module) draws the figure in on first view; reduced motion, no JS and print get the
  // still. The stage carries no word: the columns under it name each place, so the pages' copy locks
  // (test/fixtures/research-words.json) do not move. The SVG between the iso markers is written by
  // computer-scientist/article/fig_mesh_gen.py (--svelte). Rules: wave-model.css, led by .wm-mesh;
  // the tones and lines are the .iso__* rules of components.css.
  let { no, fig, parent, inset = true } = $props();   // inset=false where the band next to it is already inset
</script>
<section class="section" class:band={inset} class:band--inset={inset} id="mesh">
      <div class="wrap">
        <div class="section__head section__head--left">
          <span class="sectionno">&sect;{no} / MESH</span>
          <h2>How the mesh works.</h2>
          <p>Every Wave node speaks in typed records, not prose. A parent node reads the
             records and answers a fixed set of questions about the fleet, so a reading can
             climb from one sensor to a person without anyone retyping it.</p>
        </div>

        <figure class="figure figure--plate wm-mesh" data-draw role="group" aria-labelledby="wm-mesh-cap">
          <div class="iso__stage wm-mesh__stage" aria-hidden="true">
            <!-- iso:begin -->
            <svg class="iso__svg" viewBox="-54.2 -6 232.4 102.7" aria-hidden="true" focusable="false">
<path class="iso__halo" d="M-6.2 2.8 L172.2 54.3 L130.2 90.7 L-48.2 39.2 Z"/>
<g class="wm-mesh__ground">
<path class="iso__face iso__face--s1" d="M-40 34.6 L126.3 82.6 L126.3 86.1 L-40 38.1 Z"/>
<path class="iso__face iso__face--s2" d="M166.3 48 L126.3 82.6 L126.3 86.1 L166.3 51.5 Z"/>
<path class="iso__face" d="M0 0 L166.3 48 L126.3 82.6 L-40 34.6 Z"/>
</g>
<g class="wm-mesh__ways">
<path class="iso__way wm-mesh__way wm-mesh__way--pump" d="M54.5 23.3 L76.1 29.5"/>
<path class="iso__way wm-mesh__way wm-mesh__way--valve" d="M19.8 28.3 L63.1 40.8"/>
<path class="iso__way wm-mesh__way wm-mesh__way--comp" d="M-8.8 35.1 L50.1 52.1"/>
<path class="iso__way wm-mesh__way wm-mesh__way--read" d="M64.8 49.4 L102.9 60.4"/>
</g>
<path class="comet iso__comet wm-mesh__run wm-mesh__run--pump" d="M54.5 23.3 L76.1 29.5"/>
<path class="comet iso__comet wm-mesh__run wm-mesh__run--valve" d="M19.8 28.3 L63.1 40.8"/>
<path class="comet iso__comet wm-mesh__run wm-mesh__run--comp" d="M-8.8 35.1 L50.1 52.1"/>
<path class="comet iso__comet wm-mesh__run wm-mesh__run--read" d="M64.8 49.4 L102.9 60.4"/>
<g class="iso__m wm-mesh__child wm-mesh__child--pump">
<path class="iso__halo iso__halo--part" d="M59.7 25.3 L56.8 27.2 L53.1 28.6 L48.8 29.5 L44.1 29.8 L39.4 29.5 L35.1 28.6 L31.4 27.2 L28.5 25.3 L26.7 23.1 L26.1 20.8 L26.7 18.5 L28.5 16.3 L31.4 14.4 L35.1 13 L39.4 12.1 L44.1 11.8 L48.8 12.1 L53.1 13 L56.8 14.4 L59.7 16.3 L61.5 18.5 L62.1 20.8 L61.5 23.1 Z"/>
<path class="iso__face iso__face--s1" d="M33.5 11.8 L33.5 20.1 L33.6 20.7 L33.9 21.3 L34.4 21.9 L35 22.4 L35.8 22.9 L36.8 23.3 L37.9 23.6 L39 23.9 L40.2 24 L41.5 24.1 L42.7 24 L44 23.9 L45.1 23.6 L46.2 23.3 L47.2 22.9 L48 22.4 L48.6 21.9 L49.1 21.3 L49.4 20.7 L49.5 20.1 L49.5 11.8 Z"/>
<path class="iso__face" d="M48.4 13.8 L47.2 14.7 L45.5 15.3 L43.6 15.7 L41.5 15.8 L39.4 15.7 L37.5 15.3 L35.8 14.7 L34.6 13.8 L33.8 12.9 L33.5 11.8 L33.8 10.8 L34.6 9.8 L35.8 9 L37.5 8.4 L39.4 8 L41.5 7.8 L43.6 8 L45.5 8.4 L47.2 9 L48.4 9.8 L49.2 10.8 L49.5 11.8 L49.2 12.9 Z"/>
<path class="iso__face iso__face--s1" d="M37.5 7.9 L37.5 11.8 L37.5 12.1 L37.7 12.5 L37.9 12.7 L38.3 13 L38.7 13.2 L39.1 13.5 L39.7 13.6 L40.3 13.7 L40.9 13.8 L41.5 13.8 L42.1 13.8 L42.7 13.7 L43.3 13.6 L43.8 13.5 L44.3 13.2 L44.7 13 L45.1 12.7 L45.3 12.5 L45.4 12.1 L45.5 11.8 L45.5 7.9 Z"/>
<path class="iso__face" d="M45 8.9 L44.3 9.4 L43.5 9.7 L42.5 9.9 L41.5 9.9 L40.5 9.9 L39.5 9.7 L38.7 9.4 L38 8.9 L37.6 8.5 L37.5 7.9 L37.6 7.4 L38 6.9 L38.7 6.5 L39.5 6.2 L40.5 6 L41.5 5.9 L42.5 6 L43.5 6.2 L44.3 6.5 L45 6.9 L45.4 7.4 L45.5 7.9 L45.4 8.5 Z"/>
<path class="iso__face" d="M48.2 16 L55.2 18 L55 18 L54.9 18 L54.8 18.1 L54.7 18.1 L54.5 18.2 L54.4 18.4 L54.2 18.5 L54.1 18.7 L54 18.9 L53.9 19.1 L53.8 19.3 L53.7 19.6 L53.6 19.8 L53.5 20.1 L53.4 20.3 L53.4 20.6 L53.3 20.8 L53.3 21.1 L53.3 21.3 L53.3 21.6 L53.3 21.8 L53.3 22 L53.4 22.2 L53.5 22.3 L53.5 22.4 L53.6 22.5 L53.7 22.6 L53.8 22.7 L46.9 20.7 L46.8 20.6 L46.7 20.5 L46.6 20.4 L46.5 20.3 L46.5 20.2 L46.4 20 L46.4 19.8 L46.4 19.6 L46.4 19.3 L46.4 19.1 L46.4 18.8 L46.4 18.6 L46.5 18.3 L46.6 18.1 L46.6 17.8 L46.7 17.6 L46.8 17.3 L46.9 17.1 L47.1 16.9 L47.2 16.7 L47.3 16.5 L47.5 16.4 L47.6 16.2 L47.7 16.1 L47.9 16.1 L48 16 L48.1 16 Z"/>
<path class="iso__tone iso__tone--s2" d="M46.4 19.4 L53.3 21.4 L53.3 21.6 L53.3 21.8 L53.3 22 L53.4 22.1 L53.4 22.3 L53.5 22.4 L53.6 22.5 L53.6 22.6 L53.7 22.6 L53.8 22.7 L46.9 20.7 L46.8 20.6 L46.7 20.6 L46.6 20.5 L46.6 20.4 L46.5 20.3 L46.5 20.1 L46.4 20 L46.4 19.8 L46.4 19.6 Z"/>
<path class="iso__ln iso__ln--faint" d="M47.4 16.4 L54.4 18.4"/>
<path class="iso__face" d="M53.3 21.4 L53.3 21 L53.4 20.6 L53.4 20.2 L53.6 19.8 L53.7 19.4 L53.9 19.1 L54.1 18.8 L54.3 18.5 L54.5 18.3 L54.7 18.1 L54.9 18 L55.1 18 L55.3 18.1 L55.4 18.2 L55.5 18.4 L55.6 18.7 L55.7 19 L55.7 19.3 L55.7 19.7 L55.6 20.1 L55.5 20.5 L55.4 20.9 L55.3 21.3 L55.1 21.6 L54.9 21.9 L54.7 22.2 L54.5 22.4 L54.3 22.6 L54.1 22.7 L53.9 22.7 L53.7 22.6 L53.6 22.5 L53.4 22.3 L53.4 22 L53.3 21.7 Z"/>
<circle class="iso__lamp" cx="42.2" cy="9.3" r="1.4"/>
</g>
<g class="iso__m wm-mesh__child wm-mesh__child--valve">
<path class="iso__halo iso__halo--part" d="M27.6 31.1 L24.8 32.9 L21 34.4 L16.7 35.3 L12 35.6 L7.4 35.3 L3 34.4 L-0.7 32.9 L-3.5 31.1 L-5.3 28.9 L-6 26.6 L-5.3 24.2 L-3.5 22.1 L-0.7 20.2 L3 18.8 L7.4 17.9 L12 17.6 L16.7 17.9 L21 18.8 L24.8 20.2 L27.6 22.1 L29.4 24.2 L30 26.6 L29.4 28.9 Z"/>
<path class="iso__face" d="M-0.1 17 L6.8 19 L6.7 19 L6.5 19 L6.3 19 L6.2 19.1 L6 19.2 L5.8 19.4 L5.7 19.6 L5.5 19.8 L5.4 20 L5.2 20.3 L5.1 20.6 L4.9 20.9 L4.8 21.2 L4.7 21.5 L4.6 21.9 L4.6 22.2 L4.5 22.5 L4.5 22.8 L4.5 23.1 L4.5 23.4 L4.5 23.7 L4.6 23.9 L4.6 24.1 L4.7 24.3 L4.8 24.5 L4.9 24.6 L5 24.7 L5.2 24.8 L-1.8 22.8 L-1.9 22.7 L-2 22.6 L-2.1 22.5 L-2.2 22.3 L-2.3 22.1 L-2.4 21.9 L-2.4 21.7 L-2.4 21.4 L-2.4 21.1 L-2.4 20.8 L-2.4 20.5 L-2.4 20.2 L-2.3 19.9 L-2.2 19.5 L-2.1 19.2 L-2 18.9 L-1.9 18.6 L-1.7 18.3 L-1.6 18 L-1.4 17.8 L-1.2 17.6 L-1.1 17.4 L-0.9 17.2 L-0.7 17.1 L-0.6 17 L-0.4 17 L-0.3 17 Z"/>
<path class="iso__tone iso__tone--s2" d="M-2.4 21.2 L4.5 23.2 L4.5 23.5 L4.5 23.7 L4.6 23.9 L4.6 24.1 L4.7 24.3 L4.7 24.4 L4.8 24.5 L4.9 24.6 L5 24.7 L5.2 24.8 L-1.8 22.8 L-1.9 22.7 L-2 22.6 L-2.1 22.5 L-2.2 22.4 L-2.3 22.3 L-2.3 22.1 L-2.4 21.9 L-2.4 21.7 L-2.4 21.5 Z"/>
<path class="iso__ln iso__ln--faint" d="M-1.1 17.4 L5.8 19.4"/>
<path class="iso__face iso__face--s1" d="M3.3 18 L11.1 20.2 L11.1 28.9 L3.3 26.6 Z"/>
<path class="iso__face iso__face--s2" d="M15.6 16.3 L11.1 20.2 L11.1 28.9 L15.6 25 Z"/>
<path class="iso__face" d="M7.8 14.1 L15.6 16.3 L11.1 20.2 L3.3 18 Z"/>
<path class="iso__face" d="M13.7 21 L20.7 23 L20.5 23 L20.4 23 L20.2 23 L20 23.1 L19.9 23.2 L19.7 23.4 L19.5 23.6 L19.4 23.8 L19.2 24 L19.1 24.3 L18.9 24.6 L18.8 24.9 L18.7 25.2 L18.6 25.5 L18.5 25.9 L18.4 26.2 L18.4 26.5 L18.4 26.8 L18.3 27.1 L18.3 27.4 L18.4 27.7 L18.4 27.9 L18.5 28.1 L18.5 28.3 L18.6 28.5 L18.8 28.6 L18.9 28.7 L19 28.8 L12.1 26.8 L11.9 26.7 L11.8 26.6 L11.7 26.5 L11.6 26.3 L11.5 26.1 L11.5 25.9 L11.4 25.7 L11.4 25.4 L11.4 25.1 L11.4 24.8 L11.5 24.5 L11.5 24.2 L11.6 23.9 L11.7 23.5 L11.8 23.2 L11.9 22.9 L12 22.6 L12.1 22.3 L12.3 22 L12.4 21.8 L12.6 21.6 L12.8 21.4 L12.9 21.2 L13.1 21.1 L13.3 21 L13.4 21 L13.6 21 Z"/>
<path class="iso__tone iso__tone--s2" d="M11.4 25.2 L18.3 27.2 L18.3 27.5 L18.4 27.7 L18.4 27.9 L18.5 28.1 L18.5 28.3 L18.6 28.4 L18.7 28.5 L18.8 28.6 L18.9 28.7 L19 28.8 L12.1 26.8 L12 26.7 L11.9 26.6 L11.8 26.5 L11.7 26.4 L11.6 26.3 L11.5 26.1 L11.5 25.9 L11.4 25.7 L11.4 25.5 Z"/>
<path class="iso__ln iso__ln--faint" d="M12.7 21.4 L19.7 23.4"/>
<path class="iso__face" d="M18.3 27.2 L18.4 26.7 L18.4 26.2 L18.5 25.7 L18.7 25.2 L18.9 24.7 L19.1 24.3 L19.3 23.9 L19.6 23.5 L19.8 23.3 L20.1 23.1 L20.4 23 L20.6 23 L20.8 23.1 L21 23.2 L21.1 23.5 L21.2 23.8 L21.3 24.1 L21.3 24.6 L21.3 25 L21.2 25.5 L21.1 26.1 L21 26.6 L20.8 27 L20.6 27.5 L20.4 27.9 L20.1 28.2 L19.8 28.5 L19.6 28.7 L19.3 28.8 L19.1 28.8 L18.9 28.7 L18.7 28.5 L18.5 28.3 L18.4 28 L18.4 27.6 Z"/>
<path class="iso__ln" d="M9.4 17.2 L9.4 13.7"/>
<path class="iso__face" d="M14.2 15.1 L13.3 15.6 L12.2 16.1 L10.9 16.4 L9.4 16.4 L8 16.4 L6.7 16.1 L5.6 15.6 L4.7 15.1 L4.1 14.4 L3.9 13.7 L4.1 13 L4.7 12.3 L5.6 11.8 L6.7 11.3 L8 11 L9.4 10.9 L10.9 11 L12.2 11.3 L13.3 11.8 L14.2 12.3 L14.8 13 L14.9 13.7 L14.8 14.4 Z"/>
<path class="iso__face iso__face--cav" d="M10.7 14.1 L10.2 14.3 L9.4 14.4 L8.7 14.3 L8.1 14.1 L7.9 13.7 L8.1 13.3 L8.7 13 L9.4 12.9 L10.2 13 L10.7 13.3 L10.9 13.7 Z"/>
<circle class="iso__lamp" cx="10.4" cy="19.3" r="1.4"/>
</g>
<g class="iso__m wm-mesh__child wm-mesh__child--comp">
<path class="iso__halo iso__halo--part" d="M-6.2 36.3 L-9 38.2 L-12.7 39.6 L-17.1 40.5 L-21.7 40.8 L-26.4 40.5 L-30.7 39.6 L-34.5 38.2 L-37.3 36.3 L-39.1 34.2 L-39.7 31.8 L-39.1 29.5 L-37.3 27.3 L-34.5 25.5 L-30.7 24 L-26.4 23.1 L-21.7 22.8 L-17.1 23.1 L-12.7 24 L-9 25.5 L-6.2 27.3 L-4.4 29.5 L-3.7 31.8 L-4.4 34.2 Z"/>
<path class="iso__face iso__face--s1" d="M-39.5 29.7 L-11 37.9 L-11 39 L-39.5 30.7 Z"/>
<path class="iso__face iso__face--s2" d="M-4 31.9 L-11 37.9 L-11 39 L-4 32.9 Z"/>
<path class="iso__face" d="M-32.5 23.6 L-4 31.9 L-11 37.9 L-39.5 29.7 Z"/>
<path class="iso__face iso__face--s1" d="M-37.7 21.2 L-19.5 26.4 L-19.5 34.9 L-37.7 29.6 Z"/>
<path class="iso__face iso__face--s2" d="M-13.5 21.2 L-19.5 26.4 L-19.5 34.9 L-13.5 29.7 Z"/>
<path class="iso__face" d="M-31.7 16 L-13.5 21.2 L-19.5 26.4 L-37.7 21.2 Z"/>
<path class="iso__face iso__face--s1" d="M-35 19.7 L-33.8 20 L-33.8 21.6 L-35 21.3 Z"/>
<path class="iso__face iso__face--s2" d="M-29 15.9 L-33.8 20 L-33.8 21.6 L-29 17.4 Z"/>
<path class="iso__face" d="M-30.2 15.5 L-29 15.9 L-33.8 20 L-35 19.7 Z"/>
<path class="iso__face iso__face--s1" d="M-32 20.5 L-30.8 20.9 L-30.8 22.5 L-32 22.1 Z"/>
<path class="iso__face iso__face--s2" d="M-26 16.7 L-30.8 20.9 L-30.8 22.5 L-26 18.3 Z"/>
<path class="iso__face" d="M-27.2 16.4 L-26 16.7 L-30.8 20.9 L-32 20.5 Z"/>
<path class="iso__face iso__face--s1" d="M-29.1 21.4 L-27.9 21.7 L-27.9 23.3 L-29.1 23 Z"/>
<path class="iso__face iso__face--s2" d="M-23.1 17.6 L-27.9 21.7 L-27.9 23.3 L-23.1 19.1 Z"/>
<path class="iso__face" d="M-24.3 17.2 L-23.1 17.6 L-27.9 21.7 L-29.1 21.4 Z"/>
<path class="iso__face iso__face--s1" d="M-26.1 22.2 L-24.9 22.6 L-24.9 24.2 L-26.1 23.8 Z"/>
<path class="iso__face iso__face--s2" d="M-20.1 18.4 L-24.9 22.6 L-24.9 24.2 L-20.1 20 Z"/>
<path class="iso__face" d="M-21.3 18.1 L-20.1 18.4 L-24.9 22.6 L-26.1 22.2 Z"/>
<path class="iso__face iso__face--s1" d="M-23.2 23.1 L-22 23.4 L-22 25 L-23.2 24.7 Z"/>
<path class="iso__face iso__face--s2" d="M-17.2 19.3 L-22 23.4 L-22 25 L-17.2 20.8 Z"/>
<path class="iso__face" d="M-18.4 18.9 L-17.2 19.3 L-22 23.4 L-23.2 23.1 Z"/>
<path class="iso__face" d="M-15.4 24.1 L-7.6 26.3 L-7.8 26.3 L-8 26.3 L-8.2 26.4 L-8.5 26.5 L-8.7 26.7 L-8.9 26.9 L-9.2 27.2 L-9.4 27.5 L-9.6 27.8 L-9.8 28.2 L-10 28.6 L-10.2 29 L-10.4 29.4 L-10.5 29.9 L-10.6 30.3 L-10.7 30.8 L-10.8 31.3 L-10.8 31.7 L-10.9 32.1 L-10.8 32.5 L-10.8 32.9 L-10.7 33.2 L-10.7 33.5 L-10.6 33.8 L-10.4 34 L-10.3 34.2 L-10.1 34.3 L-9.9 34.4 L-17.7 32.2 L-17.9 32.1 L-18.1 32 L-18.2 31.8 L-18.3 31.6 L-18.5 31.3 L-18.5 31 L-18.6 30.6 L-18.6 30.3 L-18.6 29.9 L-18.6 29.4 L-18.6 29 L-18.5 28.6 L-18.4 28.1 L-18.3 27.6 L-18.2 27.2 L-18 26.8 L-17.8 26.3 L-17.6 25.9 L-17.4 25.6 L-17.2 25.2 L-17 24.9 L-16.7 24.7 L-16.5 24.5 L-16.3 24.3 L-16 24.2 L-15.8 24.1 L-15.6 24.1 Z"/>
<path class="iso__tone iso__tone--s2" d="M-18.6 30 L-10.9 32.3 L-10.8 32.6 L-10.8 32.9 L-10.7 33.2 L-10.7 33.5 L-10.6 33.7 L-10.5 33.9 L-10.4 34.1 L-10.2 34.3 L-10.1 34.4 L-9.9 34.4 L-17.7 32.2 L-17.9 32.1 L-18 32 L-18.2 31.9 L-18.3 31.7 L-18.4 31.5 L-18.5 31.3 L-18.5 31 L-18.6 30.7 L-18.6 30.4 Z"/>
<path class="iso__ln iso__ln--faint" d="M-16.8 24.7 L-9 27"/>
<path class="iso__face" d="M-10.9 32.2 L-10.8 31.5 L-10.7 30.8 L-10.6 30.1 L-10.4 29.4 L-10.1 28.8 L-9.8 28.1 L-9.5 27.6 L-9.1 27.1 L-8.8 26.7 L-8.4 26.5 L-8 26.3 L-7.7 26.3 L-7.4 26.4 L-7.1 26.7 L-6.9 27 L-6.8 27.4 L-6.7 28 L-6.7 28.6 L-6.7 29.2 L-6.8 29.9 L-6.9 30.6 L-7.1 31.3 L-7.4 32 L-7.7 32.6 L-8 33.2 L-8.4 33.6 L-8.8 34 L-9.1 34.3 L-9.5 34.4 L-9.8 34.4 L-10.1 34.3 L-10.4 34.1 L-10.6 33.8 L-10.7 33.3 L-10.8 32.8 Z"/>
<circle class="iso__lamp" cx="-21.7" cy="25.8" r="1.4"/>
</g>
<g class="wm-mesh__broker">
<path class="iso__halo" d="M86.5 48.1 L83 50.3 L78.5 52.1 L73.2 53.2 L67.5 53.6 L61.8 53.2 L56.5 52.1 L51.9 50.3 L48.4 48.1 L46.2 45.4 L45.5 42.6 L46.2 39.7 L48.4 37.1 L51.9 34.8 L56.5 33 L61.8 31.9 L67.5 31.6 L73.2 31.9 L78.5 33 L83 34.8 L86.5 37.1 L88.7 39.7 L89.5 42.6 L88.7 45.4 Z"/>
<path class="iso__face iso__face--s1" d="M46.1 33.5 L54.8 36 L54.8 58.5 L46.1 56 Z"/>
<path class="iso__face iso__face--s2" d="M88.8 6.6 L54.8 36 L54.8 58.5 L88.8 29.1 Z"/>
<path class="iso__face" d="M80.1 4.1 L88.8 6.6 L54.8 36 L46.1 33.5 Z"/>
<path class="iso__face iso__face--s1" d="M47.5 32.3 L54.4 34.3 L54.4 35.3 L47.5 33.3 Z"/>
<path class="iso__face iso__face--s2" d="M87.4 5.7 L54.4 34.3 L54.4 35.3 L87.4 6.8 Z"/>
<path class="iso__face" d="M80.5 3.7 L87.4 5.7 L54.4 34.3 L47.5 32.3 Z"/>
<path class="wm-mesh__rec wm-mesh__rec--1 iso__ln" d="M84.8 16.1 L73.8 25.6"/>
<path class="wm-mesh__rec wm-mesh__rec--2 iso__ln" d="M84.8 20 L73.8 29.5"/>
<path class="wm-mesh__rec wm-mesh__rec--3 iso__ln" d="M84.8 23.9 L73.8 33.4"/>
</g>
<g class="iso__m wm-mesh__parent">
<path class="iso__halo" d="M138.4 71.1 L134.6 73.6 L129.6 75.5 L123.8 76.7 L117.6 77.1 L111.4 76.7 L105.6 75.5 L100.7 73.6 L96.8 71.1 L94.4 68.2 L93.6 65.1 L94.4 62 L96.8 59.1 L100.7 56.6 L105.6 54.7 L111.4 53.5 L117.6 53.1 L123.8 53.5 L129.6 54.7 L134.6 56.6 L138.4 59.1 L140.8 62 L141.6 65.1 L140.8 68.2 Z"/>
<path class="iso__face iso__face--s1" d="M97.9 49.6 L127.3 58.1 L127.3 73.7 L97.9 65.2 Z"/>
<path class="iso__face iso__face--s2" d="M137.3 49.5 L127.3 58.1 L127.3 73.7 L137.3 65.1 Z"/>
<path class="iso__face" d="M107.9 41 L137.3 49.5 L127.3 58.1 L97.9 49.6 Z"/>
<path class="iso__face iso__face--s1" d="M101.5 47.8 L125.8 54.8 L125.8 56.5 L101.5 49.5 Z"/>
<path class="iso__face iso__face--s2" d="M133.8 47.8 L125.8 54.8 L125.8 56.5 L133.8 49.6 Z"/>
<path class="iso__face" d="M109.5 40.8 L133.8 47.8 L125.8 54.8 L101.5 47.8 Z"/>
<path class="iso__face iso__face--s2" d="M101.4 61 L111.8 64 L111.8 59.3 L101.4 56.3 Z"/>
<path class="iso__ln iso__ln--faint" d="M103.1 60.8 L103.1 57.4"/>
<path class="iso__ln iso__ln--faint" d="M104.8 61.3 L104.8 57.9"/>
<path class="iso__ln iso__ln--faint" d="M106.6 61.8 L106.6 58.4"/>
<path class="iso__ln iso__ln--faint" d="M108.3 62.3 L108.3 58.9"/>
<path class="iso__ln iso__ln--faint" d="M110 62.8 L110 59.4"/>
<path class="iso__face" d="M122.3 64.7 L122.2 63.7 L121.7 62.7 L120.9 61.9 L120 61.3 L119.1 61 L118.2 61.1 L117.4 61.5 L117 62.2 L116.8 63.1 L117 64.1 L117.4 65.1 L118.2 65.9 L119.1 66.5 L120 66.8 L120.9 66.7 L121.7 66.3 L122.2 65.6 Z"/>
<path class="iso__ln" d="M119.6 63.9 L121.3 62.4"/>
<path class="iso__face iso__face--cav" d="M102.2 55.2 L117.8 59.7 L117.8 57.6 L102.2 53.1 Z"/>
<path class="iso__ln iso__ln--faint" d="M103.1 64.5 L122.2 70"/>
<path class="iso__ln" d="M130 49.1 L130 39.6"/>
<circle class="iso__lamp" cx="130" cy="39.6" r="1.4"/>
<circle class="iso__lamp" cx="123.9" cy="59.3" r="1.9"/>
</g>
            </svg>
            <!-- iso:end -->
          </div>
          <div class="wm-mesh__grid">
            <div class="wm-mesh__col">
              <span class="research-kicker">Child nodes</span>
              <ul class="wm-mesh__nodes">
                <li><b>Pump</b><span>reads its channels</span></li>
                <li><b>Valve</b><span>reads its channels</span></li>
                <li><b>Compressor</b><span>reads its channels</span></li>
              </ul>
            </div>
            <div class="wm-mesh__col">
              <span class="research-kicker">Broker event log</span>
              <ol class="wm-log">
                <li><code>assert</code> <span><b>pump</b> call: drifting &middot; margin: clears its gate &middot; evidence: bearing temperature</span></li>
                <li><code>escalate</code> <span><b>valve</b> call: none &middot; margin: under its gate &middot; evidence: valve position</span></li>
                <li class="is-deciding"><code>assert</code> <span><b>compressor</b> call: stuck &middot; action: notify &middot; margin: clears its gate &middot; evidence: discharge pressure</span></li>
              </ol>
            </div>
            <div class="wm-mesh__col">
              <span class="research-kicker">Parent node &middot; {parent}</span>
              <ol class="wm-q">
                <li class="is-answered"><b>T11</b> Which asset is degraded? <em>The compressor, from the record marked in red.</em></li>
                <li><b>T12</b> Adjudicate a channel from the evidence shown</li>
                <li><b>T13</b> Resolve two children&rsquo;s conflict</li>
                <li><b>T14</b> Mesh health, from the manifest</li>
                <li><b>T15</b> Discovery: who joined, who left</li>
                <li><b>T16</b> Site rollup</li>
                <li><b>T17</b> The fire-site rollup</li>
              </ol>
            </div>
          </div>
          <figcaption class="fig mono" id="wm-mesh-cap">FIG. {fig} - Children write typed records to the log; the parent reads the log and answers. One record decides.</figcaption>
        </figure>

        <div class="device-contract wm-mesh__rules">
          <p><b>Typed assertions.</b> Each record is
             <code>&#123;kind: assert | escalate, call, action, margin, evidence&#125;</code>,
             written to the broker&rsquo;s event log. Nothing downstream has to parse a
             sentence.</p>
          <p><b>Assert or escalate.</b> A node asserts only when its margin clears its own
             measured gate. Below the gate it escalates: a record that says where to look,
             not an answer. That is the right outcome for a read it cannot make.</p>
          <p><b>Six questions, and a seventh for fire.</b> The parent answers T11 to T16
             about the fleet. The fire family adds T17: the site state (quiet, watch,
             elevated, threat, instrument fault, or abstain), the next action, which input
             decided (mast, detector, controller or neighbours), and what is inconsistent
             (camera against thermal, a valve open with zero flow, a stale record).</p>
          <p><b>Persistence, then a person.</b> A threat needs two consecutive elevated
             asserts, or one with the camera above its bar in the same tick. Pre-wet becomes
             eligible only after every path was tried for ten minutes with no
             acknowledgement, and it stays an owner-armed rule. An acknowledged escalation
             hands the site to the person.</p>
        </div>

        <p class="wm-aside">A classical rule set answers every clean log by construction. The
           model earns its place on dirty logs - a missing tick, a duplicated line, records
           out of order - and on explaining what it saw.</p>
      </div>
    </section>
