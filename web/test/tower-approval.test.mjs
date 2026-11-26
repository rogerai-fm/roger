// Executable lock for features/tower/tower_operator_guidance.feature, "The Tower page and the
// manual describe approval": both say that RogerAI approves every joined Tower, again after a
// suspension or a lapse, and that a standalone Tower needs none. Read from the source, so the
// check needs no build.
import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import path from "node:path";

const WEB = path.join(path.dirname(fileURLToPath(import.meta.url)), "..");
const visible = (s) =>
  s.replace(/<!--[\s\S]*?-->/g, " ").replace(/<script[\s\S]*?<\/script>/g, " ").replace(/<[^>]+>/g, " ").replace(/\s+/g, " ");

for (const page of ["app/routes/tower/+page.svelte", "app/routes/manual/+page.svelte"]) {
  test(`${page} says every joined Tower is approved by RogerAI`, () => {
    const text = visible(readFileSync(path.join(WEB, page), "utf8"));
    assert.match(text, /Every joined Tower is reviewed and approved by RogerAI before it carries public traffic/);
    assert.match(text, /again after a suspension or a lapse/);
    assert.match(text, /A standalone Tower needs no approval/);
    assert.doesNotMatch(text, /—/, "no em dashes");
  });
}

test("the Tower page says where to ask about a Tower, with a mailto link", () => {
  const src = readFileSync(path.join(WEB, "app/routes/tower/+page.svelte"), "utf8");
  assert.match(visible(src),
    /Questions about your Tower: email labs@rogerai\.fm and quote the tower id that roger tower status shows\./);
  assert.match(src, /<a href="mailto:labs@rogerai\.fm">labs@rogerai\.fm<\/a>/);
  assert.doesNotMatch(visible(src), /\u2014/, "no em dashes");
});
