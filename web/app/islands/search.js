// The search roster (/search.html): the page prerendered the tuner and the table; hydrate them.
import { island } from "#lib/island.js";
import Roster from "#lib/roster/Roster.svelte";
island("search", Roster, { kind: "search" });
