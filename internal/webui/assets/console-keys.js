// console-keys.js - the ACCOUNT tab's "Account keys", "Grants" and "Grants you added" cards
// (features/console/account_keys.feature, features/grants/grant_surfaces.feature). A new secret is held in this page's memory only,
// shown once, and dropped when the operator is done, leaves the tab or reloads. Everything is
// drawn with textContent, so a name can never become markup. The words for a limit, a spend and
// a state, and the export lines, come from the console, which shares them with the CLI and TUI.
(function (root) {
  "use strict";

  var RESETS = [["none", "no reset"], ["daily", "daily"], ["weekly", "weekly"], ["monthly", "monthly"]];

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

  function input(doc, cls, hint) {
    var i = node(doc, "input", cls + " inp small");
    i.setAttribute("type", "text");
    i.setAttribute("placeholder", hint);
    return i;
  }

  function select(doc, cls, value) {
    var s = node(doc, "select", cls + " inp small");
    RESETS.forEach(function (r) {
      var o = node(doc, "option", null, r[1]);
      o.value = r[0];
      s.appendChild(o);
    });
    s.value = value;
    return s;
  }

  function checkbox(doc, cls, label) {
    var l = node(doc, "label", "chk");
    var box = node(doc, "input", cls);
    box.setAttribute("type", "checkbox");
    l.appendChild(box);
    l.appendChild(node(doc, "span", null, label));
    return { label: l, box: box };
  }

  var NO_LIMIT = "Unlimited. This key can spend your whole wallet.";

  // call makes one console request and resolves to {status, ok, data}. A reply that is not
  // JSON becomes {message: its text}.
  function call(actions, method, path, body) {
    var opts = { method: method, headers: {} };
    if (body !== undefined) {
      opts.headers["Content-Type"] = "application/json";
      opts.body = JSON.stringify(body);
    }
    return actions.fetch(path, opts).then(function (r) {
      return r.text().then(function (t) {
        var data;
        try { data = JSON.parse(t); } catch (_) { data = { message: t.trim() }; }
        return { status: r.status, ok: r.ok, data: data || {} };
      });
    }, function () {
      return { status: 0, ok: false, data: { message: "The console could not be reached." } };
    });
  }

  // reveal draws a shown-once secret with copy and done.
  function reveal(doc, actions, label, shown, done) {
    var box = node(doc, "div", "key-reveal");
    box.appendChild(node(doc, "div", "key-reveal-label", label));
    box.appendChild(node(doc, "code", "code-value", shown.secret));
    if (shown.warn) box.appendChild(node(doc, "div", "key-warn", shown.warn));
    var note = node(doc, "span", "key-note");
    box.appendChild(button(doc, "copy", function () {
      actions.clipboard.writeText(shown.lines).then(
        function () { note.textContent = "copied"; },
        function () { note.textContent = "copy failed - select the key and copy it by hand"; });
    }));
    box.appendChild(button(doc, "done", done));
    box.appendChild(note);
    return box;
  }

  // mount draws both cards and resolves to their controls: load re-reads, leave drops secrets.
  function mount(doc, actions) {
    var hostK = doc.getElementById("card-keys-body"), hostG = doc.getElementById("card-grants-body");
    var hostA = doc.getElementById("card-added-body"); // optional: a page without it has no added card
    var card = { load: load, leave: leave };
    if (!hostK || !hostG) return Promise.resolve(card);

    var keys = [], keysNote = "", keysOff = false, formNote = "", shown = null, confirming = "", rowNotes = {};
    // visit counts the times the operator left. A secret that arrives for an older visit is
    // dropped, never drawn.
    var visit = 0;
    var LATE = "It was made while you were away, so its secret was dropped. Revoke it and make another.";
    // grantsNote is why the list could not be read; grantMsg is what the last action said.
    var grants = [], grantsNote = "", grantMsg = "", grantsHint = false, grantShown = null, grantConfirming = "";
    // grantRepricing is the grant whose Reprice question is open.
    var grantRepricing = "";
    // The added card: its list, why it could not be read, what the last action said, whether
    // this console is logged out, and the added grant whose Remove question is open.
    var added = [], addedNote = "", addedMsg = "", addedLogin = false, removing = "";

    // The mint form is built once, so what the operator typed survives a redraw.
    var name = input(doc, "key-name", "key name");
    var limit = input(doc, "key-limit", "limit in $");
    var reset = select(doc, "key-reset", "none");
    var noLimit = checkbox(doc, "key-nolimit", "unlimited");
    var mintBtn = button(doc, "Mint key", mint);
    noLimit.box.addEventListener("change", function () {
      if (noLimit.box.checked) limit.value = "";
      limit.disabled = noLimit.box.checked;
    });
    var form = node(doc, "div", "card-actions");
    [name, limit, reset, noLimit.label, mintBtn].forEach(function (e) { form.appendChild(e); });

    var grantName = input(doc, "grant-name", "grant name");
    var grantPrice = input(doc, "grant-price", "$ per 1M out, blank is free");
    var grantBtn = button(doc, "Create", createGrant);
    var grantForm = node(doc, "div", "card-actions");
    [grantName, grantPrice, grantBtn].forEach(function (e) { grantForm.appendChild(e); });

    // The add box is a password field: a pasted secret is not left readable on screen.
    var addBox = input(doc, "added-secret", "paste a grant secret");
    addBox.setAttribute("type", "password");
    addBox.setAttribute("autocomplete", "off");
    var addBtn = button(doc, "Add", addGrant);
    var addForm = node(doc, "div", "card-actions");
    [addBox, addBtn].forEach(function (e) { addForm.appendChild(e); });

    function keyRow(k) {
      var li = node(doc, "li", "key-row");
      li.setAttribute("data-name", k.name);
      [k.name, k.hint, k.limit_text, k.used_text, k.state].forEach(function (t) { li.appendChild(node(doc, "span", "key-cell", t)); });
      if (confirming === k.id) {
        li.appendChild(node(doc, "span", "key-ask", "Revoke " + k.name + "? Anything using it stops working."));
        li.appendChild(button(doc, "revoke", function () {
          confirming = "";
          call(actions, "POST", "/api/keys/revoke", { id: k.id }).then(function (r) { changed(k, r); });
        }));
        li.appendChild(button(doc, "keep", function () { confirming = ""; draw(); }));
        return li;
      }
      var rowLimit = input(doc, "row-limit", "limit $");
      var was = k.limit_usd > 0 ? String(k.limit_usd) : "";
      rowLimit.value = was;
      var rowReset = select(doc, "row-reset", k.reset);
      li.appendChild(rowLimit);
      li.appendChild(rowReset);
      li.appendChild(button(doc, "save", function () {
        var fields = {}, text = String(rowLimit.value).trim();
        if (text !== was) fields.limit_usd = text;
        if (rowReset.value !== k.reset) fields.reset = rowReset.value;
        if (Object.keys(fields).length) set(k, fields);
      }));
      li.appendChild(button(doc, k.disabled ? "turn on" : "turn off", function () { set(k, { disabled: !k.disabled }); }));
      li.appendChild(button(doc, "revoke", function () { confirming = k.id; draw(); }));
      if (rowNotes[k.id]) li.appendChild(node(doc, "span", "key-note", rowNotes[k.id]));
      return li;
    }

    function set(k, fields) {
      fields.id = k.id;
      call(actions, "POST", "/api/keys/set", fields).then(function (r) {
        // A limit that was just removed says what that means, as a mint with none does.
        var freed = r.ok && "limit_usd" in fields && r.data.key && !(r.data.key.limit_usd > 0);
        changed(k, r, freed ? NO_LIMIT : "");
      });
    }

    // changed redraws from the broker's list after a change, keeping a refusal or a note on
    // its row.
    function changed(k, r, note) {
      rowNotes = {};
      if (!r.ok) rowNotes[k.id] = r.data.message || "the change failed";
      else if (note) rowNotes[k.id] = note;
      return loadKeys().then(draw);
    }

    function grantRow(g) {
      var li = node(doc, "li", "grant-row");
      li.setAttribute("data-name", g.name);
      [g.name, g.price, g.status].forEach(function (t) { li.appendChild(node(doc, "span", "key-cell", t)); });
      if (grantConfirming === g.name && g.status !== "revoked") {
        li.appendChild(node(doc, "span", "key-ask", "Revoke " + g.name + "? Anyone using it loses access."));
        li.appendChild(button(doc, "revoke", function () {
          grantConfirming = "";
          call(actions, "POST", "/api/grants/revoke", { name: g.name }).then(function (r) {
            grantMsg = r.ok ? "" : (r.data.message || "the revoke failed");
            return loadGrants();
          }).then(draw);
        }));
        li.appendChild(button(doc, "keep", function () { grantConfirming = ""; draw(); }));
      } else if (grantRepricing === g.name && g.status !== "revoked") {
        var price = input(doc, "reprice-price", "new $ per 1M out");
        li.appendChild(price);
        li.appendChild(node(doc, "span", "key-ask", "Reprice " + g.name + "? This makes a new secret. The old secret stops working at once, " +
          "and every account that added it is on the public price until they add the new one."));
        li.appendChild(button(doc, "yes, reprice", function () { reprice(g, String(price.value).trim()); }));
        li.appendChild(button(doc, "keep", function () { grantRepricing = ""; draw(); }));
      } else if (g.status !== "revoked") {
        li.appendChild(button(doc, "Reprice", function () { grantRepricing = g.name; grantConfirming = ""; draw(); }));
        li.appendChild(button(doc, "revoke", function () { grantConfirming = g.name; grantRepricing = ""; draw(); }));
      }
      return li;
    }

    // reprice sends the typed price as typed; the console reads it as `roger grant create
    // --price-out` does. The new secret is shown once, with what the reprice ended.
    function reprice(g, price) {
      var asked = visit;
      grantRepricing = "";
      call(actions, "POST", "/api/grants/reprice", { name: g.name, price_out: price }).then(function (r) {
        if (!r.ok || !r.data.secret) { grantMsg = r.data.message || "the reprice failed"; return loadGrants().then(draw); }
        if (asked !== visit) { grantMsg = "Grant " + g.name + " was repriced. " + LATE; return loadGrants().then(draw); }
        grantMsg = "";
        grantShown = { label: "New secret for " + g.name + ". Store it now. It is not shown again.", secret: r.data.secret, lines: r.data.export, warn: r.data.message };
        return loadGrants().then(draw);
      });
    }

    function addedRow(g) {
      var li = node(doc, "li", "added-row");
      li.setAttribute("data-name", g.name);
      var where = g.stations.length ? "stations: " + g.stations.join(", ") : "stations: none yet";
      var what = g.models.length ? "models: " + g.models.join(", ") : "models: any";
      [g.name, g.price, where, what, g.state, g.added ? "added " + g.added : ""].forEach(function (t) {
        if (t) li.appendChild(node(doc, "span", "key-cell", t));
      });
      g.notes.forEach(function (n) { li.appendChild(node(doc, "span", "key-note", n)); });
      if (removing === g.id) {
        li.appendChild(node(doc, "span", "key-ask", "Remove " + g.name + " from your account? Its stations go back to the public price for you."));
        li.appendChild(button(doc, "yes, remove", function () {
          removing = "";
          call(actions, "POST", "/api/grants/remove", { id: g.id }).then(function (r) {
            addedMsg = r.ok ? "" : (r.data.message || "the remove failed");
            return loadAdded();
          }).then(draw);
        }));
        li.appendChild(button(doc, "keep", function () { removing = ""; draw(); }));
      } else {
        li.appendChild(button(doc, "Remove", function () { removing = g.id; draw(); }));
      }
      return li;
    }

    // addGrant sends the pasted secret once and clears the box whatever the answer: the secret
    // is not kept in the page. The reply shows the grant, never the secret.
    function addGrant() {
      var secret = String(addBox.value).trim();
      addBox.value = "";
      addedMsg = "";
      if (!secret) { addedMsg = "Paste the grant secret."; draw(); return; }
      addBtn.disabled = true;
      call(actions, "POST", "/api/grants/add", { secret: secret }).then(function (r) {
        addBtn.disabled = false;
        if (!r.ok) { addedMsg = r.data.message || "the grant was not added"; addedLogin = r.status === 401; return draw(); }
        addedMsg = "Added " + r.data.grant.name + " at " + r.data.grant.price + ".";
        return loadAdded().then(draw);
      });
    }

    function list(rows, row) {
      if (!rows.length) return node(doc, "div", "muted", "none yet");
      var ul = node(doc, "ul", "key-list mono small");
      rows.forEach(function (r) { ul.appendChild(row(r)); });
      return ul;
    }

    function draw() {
      var k = [node(doc, "div", "card-desc", "For you or your own bot. Works on any public station. Pays from your wallet.")];
      if (shown) k.push(reveal(doc, actions, "New key. Store it now. It is not shown again.", shown, function () { shown = null; draw(); }));
      if (keysNote) k.push(node(doc, "div", "key-note", keysNote));
      if (!keysOff) k.push(list(keys, keyRow));
      if (formNote) k.push(node(doc, "div", "key-note", formNote));
      form.hidden = keysOff;
      k.push(form);
      hostK.replaceChildren.apply(hostK, k);

      var g = [node(doc, "div", "card-desc", "For someone else. Works on your stations only. Free by default.")];
      if (grantShown) g.push(reveal(doc, actions, grantShown.label || "New grant. Store it now. It is not shown again.", grantShown, function () { grantShown = null; draw(); }));
      if (grantsNote || grantMsg) g.push(node(doc, "div", "key-note", grantsNote || grantMsg));
      // Only when keys do work here: a console that is logged out or offline has no form above.
      if (grantsHint && !keysOff) g.push(node(doc, "div", "key-note", "For your own bot, mint an account key above."));
      if (!grantsNote) g.push(list(grants, grantRow));
      g.push(grantForm);
      hostG.replaceChildren.apply(hostG, g);
      if (hostA) drawAdded();
    }

    function drawAdded() {
      var a = [node(doc, "div", "card-desc", "Grants someone gave you. Their stations bill you at the grant price.")];
      if (addedLogin) {
        a.push(node(doc, "div", "key-note", "Log in to add grants."));
        a.push(button(doc, "Log in", function () { if (actions.login) actions.login(); }));
      } else {
        if (addedNote || addedMsg) a.push(node(doc, "div", "key-note", addedNote || addedMsg));
        if (!addedNote) a.push(list(added, addedRow));
      }
      addForm.hidden = addedLogin || !!addedNote;
      a.push(addForm);
      hostA.replaceChildren.apply(hostA, a);
    }

    // A limit is sent as the operator typed it. The console reads it with the same code as
    // `roger keys --limit`, so this page has no rules of its own for what an amount is.
    function mint() {
      var amount = noLimit.box.checked ? "none" : String(limit.value).trim(), label = String(name.value).trim();
      formNote = "";
      if (!label) { formNote = "give the key a name"; draw(); return; }
      var asked = visit;
      mintBtn.disabled = true; // one click, one key
      call(actions, "POST", "/api/keys/mint", { name: label, limit_usd: amount, reset: reset.value }).then(function (r) {
        mintBtn.disabled = false;
        if (!r.ok || !r.data.secret) { formNote = r.data.message || "the mint failed"; draw(); return; }
        if (asked !== visit) { formNote = "Key " + label + " was minted. " + LATE; return loadKeys().then(draw); }
        shown = { secret: r.data.secret, lines: r.data.export, warn: r.data.key && r.data.key.limit_usd > 0 ? "" : NO_LIMIT };
        name.value = ""; limit.value = ""; limit.disabled = false; noLimit.box.checked = false; reset.value = "none";
        return loadKeys().then(draw);
      });
    }

    function createGrant() {
      var label = String(grantName.value).trim(), price = String(grantPrice.value).trim();
      grantMsg = "";
      if (!label) { grantMsg = "Name the grant."; draw(); return; }
      var asked = visit;
      grantBtn.disabled = true; // one click, one grant
      var body = { name: label };
      if (price) body.price_out = price; // as typed: the console reads it as the CLI does
      call(actions, "POST", "/api/grants", body).then(function (r) {
        grantBtn.disabled = false;
        if (!r.ok || !r.data.secret) { grantMsg = r.data.message || "the grant was not created"; grantsHint = true; draw(); return; }
        if (asked !== visit) { grantMsg = "Grant " + label + " was created. " + LATE; return loadGrants().then(draw); }
        grantShown = { secret: r.data.secret, lines: r.data.export };
        grantName.value = ""; grantPrice.value = "";
        return loadGrants().then(draw);
      });
    }

    // loadKeys reads the key list. It resolves to false when this console has no account to
    // ask about at all, so the grants are not asked for either.
    function loadKeys() {
      return call(actions, "GET", "/api/keys").then(function (r) {
        keys = r.ok && Array.isArray(r.data.keys) ? r.data.keys : [];
        keysOff = !r.ok;
        keysNote = r.ok ? "" : r.status === 401 ? "Log in to manage keys." : (r.data.message || "the keys could not be read");
        return r.status !== 409;
      });
    }

    function loadGrants() {
      return call(actions, "GET", "/api/grants").then(function (r) {
        grants = r.ok && Array.isArray(r.data) ? r.data : [];
        grantsNote = r.ok ? "" : (r.data.message || "the grants could not be read");
        grantsHint = grantsHint || !r.ok;
      });
    }

    // loadAdded reads the added list. A 401 means logged out: the card offers the login button.
    function loadAdded() {
      if (!hostA) return Promise.resolve();
      return call(actions, "GET", "/api/grants/added").then(function (r) {
        added = r.ok && Array.isArray(r.data.grants) ? r.data.grants : [];
        addedLogin = r.status === 401;
        addedNote = r.ok || addedLogin ? "" : (r.data.message || "the added grants could not be read");
      });
    }

    function load() {
      grantMsg = ""; grantsHint = false; addedMsg = "";
      return loadKeys().then(function (account) {
        if (account) return Promise.all([loadGrants(), loadAdded()]);
        grants = []; grantsNote = keysNote; added = []; addedNote = keysNote; addedLogin = false;
      }).then(draw);
    }

    // leave drops every shown secret and half-made choice: the operator went elsewhere.
    function leave() {
      visit++;
      shown = null; grantShown = null; confirming = ""; grantConfirming = ""; formNote = ""; grantMsg = ""; rowNotes = {};
      grantRepricing = ""; removing = ""; addedMsg = ""; addBox.value = "";
      draw();
    }

    // Nothing is read until the operator opens the tab: load() is the console's to call.
    draw();
    return Promise.resolve(card);
  }

  var api = { mount: mount };
  if (typeof module === "object" && module.exports) {
    module.exports = api;
  } else {
    root.RogerConsoleKeys = api;
  }
})(this);
