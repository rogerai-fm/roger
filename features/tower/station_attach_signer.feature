# BUILD STATUS: BUILT. cmd/rogerai-broker/toweredgeattach.go and accountkey.go (nodeBoundOwner),
# internal/towercli/main.go (stageAllows).
# A Station's self-attach is signed by its NODE key, and its owner is the account the node is
# bound to.
#
# Found 2026-11-23 while testing the first house Tower (tw-3081d0565b48999cdb34fbce). A
# `roger share` registered, passed its health checks, and then every POST /tower/edge/attach
# was refused 401 "attaching a node requires a signed-in account" (governance log:
# station_attach_refused actor=unsigned).
#
# Why: the share signs the attach with this machine's node key (node.key), the same key that
# registered the node, because the station<->node join requires exactly that
# (nodeRegisteredTo). Core then resolved the signer with OwnerByPubkey, and owner rows are keyed
# by the LOGIN key (user.key). The two keys are different files, so on any machine where they
# differ no share can join any Tower. Unit tests used one key for both roles and could not see
# it. Same class as the 2026-11-13 renewal bug (client and broker disagreeing about which key
# signs).
#
# The owner is already known. A share's registration is co-signed with the owner's login key,
# and Core records that as the node's owner binding (BindNode, TOFU). So the attach is signed by
# the node key, the join proves the node is this machine's, and the owner is read from that
# binding. Core never takes an owner from the request body.
#
# AMENDED (founder re-approved 2026-11-23): the binding alone is not enough. The node-key link lives
# in the live registration, which is pruned after 7 days offline, while the binding is durable.
# A stranger's key could re-register a lapsed node id as a free node and then attach as the
# bound account (found by the pre-push audit, 2026-11-23). So Core also stamps the account that
# co-signed each registration onto that registration, and the owner must be both: the account
# that co-signed the live registration AND the node's owner binding.
#
# Terms:
#   node key       the key in node.key; registers the node and signs the attach
#   login key      the key in user.key; signed in with `roger login`, an owner row
#   owner binding  the account recorded for a node id at its registration (TOFU)
#   co-signer      the account whose login key co-signed the node's LIVE registration, stamped
#                  by Core on that registration (empty for a registration with no owner signature)
#   account key    the canonical key of an account (accountKeyOf), the payee of its lots

Feature: A Station attaches under the account its node is bound to

  Background:
    Given a live joined Tower that may take Stations
    And the attaching Station sends a body that is valid in every other respect, co-signed by its assertion key

  Rule: The node key signs, and the owner comes from the node's owner binding

    Scenario: A signed-in owner whose node key differs from their login key can attach
      Given an owner signed in with login key L
      And their machine registered node N with node key K, co-signed by L
      And K is not L
      When the Station attaches naming node N, signed by K
      Then the attach is accepted
      And the attachment's owner is the account key of L's account
      And the governance record names the owner's login as the actor

    Scenario: An owner whose node key is their login key attaches as before
      Given an owner whose node key and login key are the same key L
      And their machine registered node N with L
      When the Station attaches naming node N, signed by L
      Then the attach is accepted
      And the attachment's owner is the account key of L's account

    Scenario: The payee is the account, never the node key
      Given an owner with two devices, D1 and D2, signed in to one account
      And D2 registered node N with node key K, co-signed by D2's login key
      When the Station attaches naming node N, signed by K
      Then the attachment's owner is the account's canonical key
      And it is neither K nor D2's login key unless that is the canonical key

    Scenario: A house account's node attaches as a house Station
      Given the owner bound to node N is on the house list
      When the Station attaches naming node N, signed by its node key
      Then placement treats it as a house Station

  Rule: The owner must also be the account that co-signed the live registration

    Scenario: A lapsed node id re-registered by a stranger's key cannot attach as its old owner
      Given node N is bound to a signed-in account A
      And N's registration was pruned after going offline
      And a stranger's key K' registers N again with no owner signature
      When the Station attaches naming node N, signed by K'
      Then the attach is refused with 401 "attaching a node requires a signed-in account - run `roger login`"
      And no attachment, invitation or station row is written

    Scenario: A node id re-registered under another account cannot attach as either
      Given node N is bound to a signed-in account A
      And N's registration was pruned after going offline
      And a key K' registers N again, co-signed by the login key of another account B
      When the Station attaches naming node N, signed by K'
      Then the attach is refused with 403 "node_id is not registered to this key - register with `roger share` before attaching"
      And no attachment, invitation or station row is written

    Scenario: The co-signer is what Core observed, never what the registration says
      Given a registration body for node N that names a co-signer but carries no owner signature
      When Core stores the registration
      Then the stored registration has no co-signer
      And an attach naming N, signed by N's node key, is refused with 401

    Scenario: The co-signer survives a broker restart and every instance sees it
      Given an owner's node N registered with node key K, co-signed by their login key
      When the broker restarts and rehydrates its registrations from the store
      Then the attach naming node N, signed by K, is accepted under the owner's account
      And a second instance that only holds the mirrored registration accepts it too

    Scenario: The co-signer is never shown to consumers
      Given an owner's node N registered with a co-signer
      Then no consumer-facing listing of N names the co-signer

  Rule: Anything that does not resolve to a signed-in account is refused, and nothing is stored

    Scenario: A node registered without an owner is refused
      Given node N was registered free with no owner signature, so it has no owner binding
      When the Station attaches naming node N, signed by N's node key
      Then the attach is refused with 401 "attaching a node requires a signed-in account - run `roger login`"
      And no attachment, invitation or station row is written

    Scenario: A node bound to a deleted account is refused
      Given node N is bound to an account that has since been anonymized
      When the Station attaches naming node N, signed by N's node key
      Then the attach is refused with 401
      And no attachment is written

    Scenario: A node bound to an account that may not enroll is refused
      Given node N is bound to an account that is banned or barred from enrolling
      When the Station attaches naming node N, signed by N's node key
      Then the attach is refused with 403 "this account may not attach a node"

    Scenario: An unsigned attach is refused
      When a Station attaches with no request signature
      Then the attach is refused with 401
      And no store read resolves an owner for it

    Scenario: An attach whose signature does not verify is refused
      When a Station attaches with a signature that does not verify over the request
      Then the attach is refused with 401

    Scenario: A signer that is no account and names no node is refused as not signed in
      Given key K is neither an owner row nor the key of any registered node
      When a Station attaches with no node_id, signed by K
      Then the attach is refused with 401

    Scenario: A signer that is no account, naming a node registered to another key, is refused
      Given node N is registered to node key K1 and bound to a signed-in account
      And key K2 is not K1 and is no owner row
      When a Station attaches naming node N, signed by K2
      Then the attach is refused
      And the refusal does not reveal whether N exists or which account owns it

    Scenario: A node whose registration is no longer live is refused
      Given node N is bound to a signed-in account
      But N has no live registration
      When the Station attaches naming node N, signed by N's former node key
      Then the attach is refused with 403 "node_id is not registered to this key - register with `roger share` before attaching"

  Rule: The theft guards are unchanged

    Scenario: A signed-in owner cannot attach another machine's node
      Given node N is registered to node key K1 and bound to account A
      And an owner of account B signs in with login key L2
      When a Station attaches naming node N, signed by L2
      Then the attach is refused with 403 "node_id is not registered to this key - register with `roger share` before attaching"

    Scenario: A signer that is an account key and a node bound to a different account is refused
      Given key L is the login key of account A
      And node N is registered with node key L but its owner binding is account B
      When the Station attaches naming node N, signed by L
      Then the attach is refused with 403
      And the attachment is not written under either account

    Scenario: The assertion-key co-signature is still required
      Given an owner whose node N is bound to their account
      When the Station attaches naming node N, signed by N's node key, without a valid assertion-key co-signature
      Then the attach is refused exactly as it is today

    Scenario: A private band still never attaches
      Given a Station sharing on a private band
      Then it never calls attach, and Core refuses a private node that does

  Rule: Limits and records follow the account, not the key

    Scenario: The attach rate limit is per account, across node keys
      Given an owner whose account has several machines, each with its own node key
      When those machines attach at once beyond the per-account bucket
      Then the excess is refused with 429 and Retry-After
      And minting more node keys under the same account does not raise the limit

    Scenario: A refusal after the owner is known names the owner
      Given node N is bound to a signed-in account with login "op"
      When an attach naming N, signed by N's node key, is refused for any later reason
      Then the governance record names actor "operator:op"

  Rule: The real client and the real broker agree

    Scenario: A real share attaches to a live Tower end to end
      Given a broker with a live joined Tower that may take Stations
      And an owner signed in through the real login code with login key L
      And the real share registration code registers node N with a separate node key K, co-signed by L
      When the real share attach code attaches node N
      Then the attach is accepted and names the live Tower
      And on the previous broker it fails with 401 "attaching a node requires a signed-in account"

  Rule: A house Tower's status says what it can hold

    Scenario: roger tower status on a house Tower does not say it holds no RogerAI Stations
      Given a joined Tower whose owner is a house account
      When the operator runs roger tower status
      Then the trust line says house Towers take RogerAI Stations without a trust cap
      And it does not say "no RogerAI Stations"
