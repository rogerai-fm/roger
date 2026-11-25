// Ping's shapes: the iOS rig's wardrobe (roger-ios RogerAI/Features/Avatar/PingRig.swift and PingStyle.swift),
// in the logo's own 32-unit space: brackets are centre lines at their weight, the eye is r 3 at (16, 12), the feet
// stand at y 27, the figure's base is (16, 29.6). Data only, so a look is a set of names (head, body, eye, wear)
// and Ping.svelte draws any combination. Keep in step with the rig when it gains a shape.

export const HEADS = {
  square: { d: "M11 7.5 L6.5 7.5 L6.5 16.5 L11 16.5", w: 3, cap: "butt" },
  angle: { d: "M10.4 6.4 L6.2 12 L10.4 17.6", w: 2.8, cap: "butt" },
  round: { d: "M10.4 6.6 Q3.2 12 10.4 17.4", w: 2.8, cap: "round" },
  curly: { d: "M10.8 7 Q8.4 7 8.4 9.2 Q8.4 12 6 12 Q8.4 12 8.4 14.8 Q8.4 17 10.8 17", w: 2.4, cap: "round" },
  slash: { d: "M10.4 6.8 L6.4 17.2", w: 3, cap: "round" },
  bar: { d: "M7.5 6 L7.5 18", w: 3, cap: "butt" },
};

export const BODIES = {
  classic: "M11 19 L21 19 L21 25 Q21 27 19 27 L13 27 Q11 27 11 25 Z",
  round: "M10.4 23.05 A5.6 4.25 0 1 0 21.6 23.05 A5.6 4.25 0 1 0 10.4 23.05 Z",
  wide: "M12.6 19 L19.4 19 Q19.8 19 20 19.4 L22.8 26.4 Q23 27 22.4 27 L9.6 27 Q9 27 9.2 26.4 L12 19.4 Q12.2 19 12.6 19 Z",
  diamond: "M16 18.3 L22.4 23 L16 27.7 L9.6 23 Z",
  hex: "M13 19 L19 19 L22.2 23 L19 27 L13 27 L9.8 23 Z",
  ghost: "M10.8 27.8 L10.8 23.6 C10.8 17.1 21.2 17.1 21.2 23.6 L21.2 27.8 L19.47 26.3 L17.73 27.8 L16 26.3 L14.27 27.8 L12.53 26.3 L10.8 27.8 Z",
};

/** where the two feet stand, or none (a ghost hovers) */
export const FEET = {
  classic: [13, 16.4, 27], round: [13, 16.4, 27], hex: [13, 16.4, 27], wide: [11.4, 18, 27], diamond: [13.2, 16.2, 27.1],
};

/** the eye colours: --live, or one of the iOS palette tokens (tokens.css --ping-*) */
export const EYES = ["live", "orange", "amber", "green", "teal", "blue", "violet", "pink"];

// The worn things, each a list of shapes: [element, class, attributes]. Classes: ink (filled in ink),
// trim (a lighter band, where the rig uses the eye colour), and the strokes rim, band, stem, trimln, halo.
const r = (x, y, width, height, rx) => ({ x, y, width, height, ...(rx ? { rx } : {}) });
export const WEAR = {
  glasses: [["circle", "rim", { cx: 16, cy: 12, r: 4.5 }], ["path", "rim", { d: "M11.5 12 L7.8 12 M20.5 12 L24.2 12" }]],
  headphones: [["path", "band", { d: "M4.3 9.8 C4.3 0.9 27.7 0.9 27.7 9.8" }],
    ["rect", "trim", r(2.5, 9, 2.6, 6.2, 1.2)], ["rect", "trim", r(26.9, 9, 2.6, 6.2, 1.2)]],
  beanie: [["path", "ink", { d: "M10.4 5.6 C10.4 0.4 21.6 0.4 21.6 5.6 Z" }], ["rect", "trim", r(9.7, 4.7, 12.6, 1.6, 0.8)],
    ["circle", "trim", { cx: 16, cy: 1.1, r: 0.9 }]],
  cowboy: [["path", "ink", { d: "M3.9 4.5 Q9 7.6 16 6.5 Q23 7.6 28.1 4.5 Q22.6 6.2 16 5.3 Q9.4 6.2 3.9 4.5 Z" }],
    ["path", "ink", { d: "M11.2 5.8 L12.2 2 Q13.9 .2 16 2.2 Q18.1 .2 19.8 2 L20.8 5.8 Z" }], ["path", "trimln", { d: "M11.45 4.25 L20.55 4.25" }]],
  antenna: [["path", "stem", { d: "M24.6 6.2 L25.8 1.9" }], ["circle", "trim", { cx: 25.8, cy: 1.9, r: 1 }]],
  wizard: [["path", "ink", { d: "M11.4 5.4 L20.6 5.4 Q17.2 2 19.4 .5 Q14.6 .2 11.4 5.4 Z" }],
    ["ellipse", "ink", { cx: 16, cy: 5.55, rx: 7.6, ry: 0.85 }], ["circle", "trim", { cx: 15.6, cy: 3.5, r: 0.8 }]],
  halo: [["ellipse", "halo", { cx: 16, cy: 2.6, rx: 4.8, ry: 1.2 }]],
  crown: [["path", "ink", { d: "M12.2 5.1 L12.2 2.1 L14.1 3.6 L16 1.2 L17.9 3.6 L19.8 2.1 L19.8 5.1 Z" }],
    ["circle", "trim", { cx: 16, cy: 1.2, r: 0.55 }], ["circle", "trim", { cx: 12.2, cy: 2.1, r: 0.55 }], ["circle", "trim", { cx: 19.8, cy: 2.1, r: 0.55 }]],
  tophat: [["rect", "ink", r(9, 5, 14, 1.1, 0.5)], ["rect", "ink", r(11.6, 0.5, 8.8, 4.8, 0.5)], ["rect", "trim", r(11.6, 3.7, 8.8, 1)]],
  bowtie: [["path", "trim", { d: "M16 20.5 L13.3 19.1 L13.3 21.9 Z M16 20.5 L18.7 19.1 L18.7 21.9 Z" }]],
};

/** the model a Ping can carry: a small cube held in front of the body, its top face the signal */
export const CUBE = [
  ["path", "cube cube--top", { d: "M16 18.2 L19.6 19.9 L16 21.6 L12.4 19.9 Z" }],
  ["path", "cube cube--l", { d: "M12.4 19.9 L16 21.6 L16 25.6 L12.4 23.9 Z" }],
  ["path", "cube cube--r", { d: "M16 21.6 L19.6 19.9 L19.6 23.9 L16 25.6 Z" }],
];

/** the right bracket is the left one mirrored about x = 16 */
export function mirror(d) {
  const t = d.match(/[MLQC]|-?[\d.]+/g);
  const out = [];
  for (let i = 0; i < t.length; ) {
    if (/[MLQC]/.test(t[i])) { out.push(t[i++]); continue; }
    out.push(`${+(32 - t[i]).toFixed(2)} ${t[i + 1]}`); i += 2;
  }
  return out.join(" ");
}

// The brand-mark Ping (src/mascot.svg, a 64-unit space standing at about (32, 69)): the older flat drawing the
// site's own pages use (the homepage mascot, the custom tuning figures, the 404). PingMark.svelte draws it.
// Each shape is a list of [element, attributes], all filled in currentColor.
const MARK_BODY = ["path", { d: "M22 30 h20 v18 a4 4 0 0 1 -4 4 h-12 a4 4 0 0 1 -4 -4 Z" }];
const MARK_FEET = [["rect", { x: 25, y: 64, width: 5, height: 5, rx: 1 }], ["rect", { x: 34, y: 64, width: 5, height: 5, rx: 1 }]];
export const MARKS = {
  /** the mascot: square brackets, a rounded body, two feet */
  mark: [["path", { d: "M10 8 h9 v4.5 H14.5 v9 H19 V26 H10 Z" }], ["path", { d: "M54 8 h-9 v4.5 H49.5 v9 H45 V26 h9 Z" }], MARK_BODY, ...MARK_FEET],
  /** off the air (404): the brackets sag by two */
  dropped: [["path", { d: "M10 10 h9 v4.5 H14.5 v9 H19 V28 H10 Z" }], ["path", { d: "M54 10 h-9 v4.5 H49.5 v9 H45 V28 h9 Z" }], MARK_BODY, ...MARK_FEET],
  /** the small flat form drawn on a bench (custom tuning FIG. 1): thinner brackets, a block body */
  flat: [["path", { d: "M14 10 h7 v4 h-3 v10 h3 v4 h-7z" }], ["path", { d: "M50 10 h-7 v4 h3 v10 h-3 v4 h7z" }],
    ["rect", { x: 24, y: 30, width: 16, height: 14, rx: 3 }], ["rect", { x: 26, y: 46, width: 4, height: 4 }], ["rect", { x: 34, y: 46, width: 4, height: 4 }]],
};
