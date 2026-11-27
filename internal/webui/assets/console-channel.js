// console-channel.js - the console's endpoint plate (features/console/endpoint_plate.feature):
// the BASE URL / API KEY / MODEL plate `roger use` and the terminal end on, for the model in the
// CHAT picker. The key is drawn masked and is never kept here: copy asks the console for it,
// writes the clipboard, and lets it go.
(function (root) {
  "use strict";

  // How often an open page re-reads the plate, so a channel closed in another tab goes away here.
  var POLL_MS = 1500;

  function mask(hint) { return "••••••••" + (hint || ""); }

  function exportBlock(base, key, model) {
    return "export OPENAI_BASE_URL=" + base + "\nexport OPENAI_API_KEY=" + key + "\nexport OPENAI_MODEL=" + model;
  }

  function node(doc, tag, cls, text) {
    var e = doc.createElement(tag);
    if (cls) e.className = cls;
    if (text) e.textContent = text;
    return e;
  }

  function button(doc, label, fn) {
    var b = node(doc, "button", "btn small", label);
    b.setAttribute("type", "button");
    b.addEventListener("click", fn);
    return b;
  }

  function json(actions, method, path, body) {
    var opts = { method: method, headers: {} };
    if (body !== undefined) {
      opts.headers["Content-Type"] = "application/json";
      opts.body = JSON.stringify(body);
    }
    return actions.fetch(path, opts).then(function (r) {
      return r.json().catch(function () { return { ok: false, message: "the console answered " + r.status }; });
    });
  }

  function money(n) { return "$" + Number(n || 0).toFixed(2); }

  // mount draws the plate into #channel-plate and resolves to its controls.
  function mount(doc, actions) {
    var host = doc.getElementById("channel-plate");
    var plate = { refresh: refresh, open: open, copy: copy, close: close };
    if (!host) return Promise.resolve(plate);

    function drawOpen(st) {
      var rows = node(doc, "div", "plate-rows mono");
      function row(label, value, name) {
        var r = node(doc, "div", "plate-row");
        r.appendChild(node(doc, "span", "plate-label", label));
        var v = node(doc, "span", "plate-value", value);
        if (name) v.setAttribute("aria-label", name);
        r.appendChild(v);
        rows.appendChild(r);
      }
      row("BASE URL", st.base_url);
      row("API KEY", mask(st.key_hint), "API key ending in " + st.key_hint);
      row("MODEL", st.model);
      var acts = node(doc, "div", "plate-actions");
      acts.appendChild(button(doc, "copy", copy));
      acts.appendChild(button(doc, "close", close));
      host.replaceChildren(
        node(doc, "p", "plate-head", "channel open" + (st.station ? " via @" + st.station : "")),
        rows,
        node(doc, "p", "plate-line", "drop-in, OpenAI-compatible - point any OpenAI tool here."),
        acts,
        node(doc, "p", "plate-note muted", "Works on this machine only, while roger is running."),
        node(doc, "p", "plate-note muted", "A bot on a server needs an account key instead: ACCOUNT tab, Account keys")
      );
    }

    function drawClosed(quote, refusal) {
      var model = actions.model();
      var kids = [];
      if (refusal && refusal.message) kids.push(node(doc, "p", "plate-refusal", refusal.message));
      if (refusal && refusal.login_needed) {
        kids.push(button(doc, "go to ACCOUNT", function () { actions.tab("account"); }));
      }
      if (!model) {
        kids.push(node(doc, "p", "plate-note muted", "endpoint closed - pick an open market model to open an endpoint"));
        host.replaceChildren.apply(host, kids);
        return;
      }
      if (quote && quote.on_air) {
        var price = quote.free ? "free now" : money(quote.price_out) + " / 1M out now";
        kids.push(node(doc, "p", "plate-note muted", "endpoint closed - " + price + " - your max " + money(quote.cap_out) + " / 1M out"));
      } else {
        kids.push(node(doc, "p", "plate-note muted", "endpoint closed"));
      }
      if (refusal && refusal.confirm_needed) {
        var typed = node(doc, "input", "inp small");
        typed.setAttribute("type", "text");
        typed.setAttribute("aria-label", "type the out price to confirm");
        kids.push(typed);
        kids.push(button(doc, "confirm " + money(refusal.price_out) + " and open", function () { open(Number(typed.value)); }));
      } else {
        kids.push(button(doc, "open an endpoint for " + model, function () { open(); }));
      }
      host.replaceChildren.apply(host, kids);
    }

    // What the plate is showing, so a re-read that finds nothing new leaves it alone: a
    // redraw would take a refusal, a half-typed price or the keyboard focus away.
    var shown = "";

    function showOpen(st) {
      shown = ["open", st.base_url, st.model, st.station, st.key_hint].join("|");
      drawOpen(st);
    }

    // The quote asks the broker, so it is read when the plate closes, the model changes or
    // an open is refused. Never on the timer.
    function showClosed(refusal) {
      var model = actions.model();
      shown = "closed|" + model;
      if (!model) { drawClosed(null, refusal); return Promise.resolve(); }
      return json(actions, "GET", "/api/channel/quote?model=" + encodeURIComponent(model))
        .then(function (q) { drawClosed(q, refusal); })
        .catch(function () { drawClosed(null, refusal); });
    }

    function refresh() {
      return json(actions, "GET", "/api/channel").then(function (st) {
        if (st && st.open) {
          if (shown !== ["open", st.base_url, st.model, st.station, st.key_hint].join("|")) showOpen(st);
          return;
        }
        if (shown !== "closed|" + actions.model()) return showClosed(null);
      }).catch(function () {});
    }

    function open(confirmPrice) {
      var body = { model: actions.model() };
      if (confirmPrice !== undefined) body.confirm_price_out = confirmPrice;
      return json(actions, "POST", "/api/channel/open", body).then(function (r) {
        if (r && r.ok) { showOpen(r); return; } // only the hint is drawn; the key is dropped here
        return showClosed(r || { message: "the channel did not open" });
      });
    }

    // One reply carries the key with the endpoint it belongs to.
    function copy() {
      return json(actions, "POST", "/api/channel/key").then(function (k) {
        if (!k || !k.key) { actions.toast("copy failed - no channel open", "warn"); return refresh(); }
        return actions.clipboard.writeText(exportBlock(k.base_url, k.key, k.model))
          .then(function () { actions.toast("copied the endpoint, key and model", "ok"); },
                function () { actions.toast("copy failed", "warn"); });
      });
    }

    function close() {
      return json(actions, "POST", "/api/channel/close").then(refresh);
    }

    // Re-read on a timer, so a channel closed in another tab goes away here too.
    actions.every(refresh, POLL_MS);
    return refresh().then(function () { return plate; });
  }

  var api = { mount: mount, POLL_MS: POLL_MS };
  if (typeof module === "object" && module.exports) {
    module.exports = api;
  } else {
    root.RogerConsoleChannel = api;
  }
})(this);
