#!/usr/bin/env bash
# ci-valkey.sh up|down: a throwaway Valkey for the coverage job, on GitHub's runner (docker) or on our own
# machine (podman). `up` starts it on a host port the runtime picks, waits for PING, and writes
# ROGERAI_TEST_REDIS_URL and CI_VALKEY_CT to $GITHUB_ENV. `down` removes it. The port is never pinned: on a
# machine people work on, 6379 may be somebody's own server, and the suite would read and write it.
set -euo pipefail
# CONTAINER_RUNTIME names the runtime (CI sets it); otherwise podman, then docker, as scripts/cover-gate.sh.
RT="$(command -v "${CONTAINER_RUNTIME:-podman}" || { [ -z "${CONTAINER_RUNTIME:-}" ] && command -v docker; } || true)"
[ -n "$RT" ] || { echo "[ci-valkey] no docker or podman on this runner" >&2; exit 2; }
case "${1:-}" in
  up)
    # A runner killed mid-job leaves its container and port behind: remove any of ours older than three
    # hours (a job is capped at 75 minutes), never a younger one, which may be another job's.
    now="$(date +%s)"
    "$RT" ps -a --filter "name=^ci-valkey-" --format "{{.Names}} {{.CreatedAt}}" 2>/dev/null | while read -r n c; do
      t="$(date -d "$(echo "$c" | cut -d" " -f1-2 | cut -d. -f1)" +%s 2>/dev/null || echo "$now")"
      [ $((now - t)) -gt 10800 ] && "$RT" rm -f "$n" >/dev/null 2>&1 || true
    done
    CT="ci-valkey-${GITHUB_RUN_ID:-local}-${GITHUB_RUN_ATTEMPT:-0}-$$"
    "$RT" run -d --name "$CT" -p 127.0.0.1::6379 docker.io/valkey/valkey:8 >/dev/null
    PORT="$("$RT" port "$CT" 6379/tcp | head -1 | sed 's/.*://')"
    [ -n "$PORT" ] || { echo "[ci-valkey] could not read the host port" >&2; "$RT" rm -f "$CT" >/dev/null 2>&1; exit 2; }
    ok=0
    for _ in $(seq 1 60); do
      if [ "$("$RT" exec "$CT" valkey-cli ping 2>/dev/null)" = "PONG" ]; then ok=1; break; fi
      sleep 1
    done
    [ "$ok" = 1 ] || { echo "[ci-valkey] Valkey never answered PING" >&2; "$RT" rm -f "$CT" >/dev/null 2>&1; exit 2; }
    echo "[ci-valkey] $CT on 127.0.0.1:$PORT ($(basename "$RT"))" >&2
    if [ -n "${GITHUB_ENV:-}" ]; then
      { echo "ROGERAI_TEST_REDIS_URL=redis://127.0.0.1:$PORT"; echo "CI_VALKEY_CT=$CT"; } >> "$GITHUB_ENV"
    else
      echo "ROGERAI_TEST_REDIS_URL=redis://127.0.0.1:$PORT"; echo "CI_VALKEY_CT=$CT"
    fi
    ;;
  down)
    [ -n "${CI_VALKEY_CT:-}" ] && "$RT" rm -f "$CI_VALKEY_CT" >/dev/null 2>&1 || true
    ;;
  *) echo "usage: ci-valkey.sh up|down" >&2; exit 2 ;;
esac
