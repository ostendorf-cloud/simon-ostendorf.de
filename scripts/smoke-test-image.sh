#!/usr/bin/env bash
#
# Starts the container image given as $1 and asserts it actually serves
# a request: health, readiness and metrics on the admin listener, and
# the home page on the public one. Run by the docker-build job in
# .github/workflows/checks.yaml right after it builds and loads the
# image, so this never needs a registry.
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <image-ref>" >&2
  exit 1
fi

image="$1"
container="simon-ostendorf-de-smoke-$$"
base_url="http://127.0.0.1:8080"
admin_url="http://127.0.0.1:8081"
ready_timeout=30
fail=0

cleanup() {
  local status=$?
  if [[ $status -ne 0 ]]; then
    echo "---- ${container} logs ----" >&2
    docker logs "$container" 2>&1 || true
    echo "---- end ${container} logs ----" >&2
  fi
  docker rm --force "$container" >/dev/null 2>&1 || true
  exit "$status"
}
trap cleanup EXIT

docker run \
  --detach \
  --name "$container" \
  --publish "127.0.0.1:8080:8080" \
  --publish "127.0.0.1:8081:8081" \
  "$image" >/dev/null

# wait_for_health polls /healthz on the admin listener for up to
# ready_timeout seconds, failing immediately once the container itself
# has exited rather than waiting out the full timeout for a process that
# is already gone.
wait_for_health() {
  local elapsed=0 state exit_code
  while ((elapsed < ready_timeout)); do
    state="$(docker inspect --format '{{.State.Running}}' "$container" 2>/dev/null || echo false)"
    if [[ $state != "true" ]]; then
      exit_code="$(docker inspect --format '{{.State.ExitCode}}' "$container" 2>/dev/null || echo unknown)"
      echo "${container} exited early with code ${exit_code}" >&2
      return 1
    fi
    if curl --fail --silent --show-error --max-time 2 "${admin_url}/healthz" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
    elapsed=$((elapsed + 1))
  done
  echo "timed out after ${ready_timeout}s waiting for /healthz" >&2
  return 1
}

if ! wait_for_health; then
  exit 1
fi

# assert_status records a failure and keeps going, checking every route
# rather than stopping at the first failure. It defaults to base_url, the
# public listener; a third argument points at the admin listener instead.
assert_status() {
  local path="$1" want="$2" url="${3:-$base_url}" got
  got="$(curl --silent --output /dev/null --write-out '%{http_code}' "${url}${path}")" || got=""
  if [[ $got != "$want" ]]; then
    echo "GET ${url}${path}: want status ${want}, got \"${got}\"" >&2
    fail=1
  fi
}

assert_status "/healthz" "200" "$admin_url"
assert_status "/readyz" "200" "$admin_url"
assert_status "/metrics" "200" "$admin_url"

# The public listener never answers /metrics: health and metrics only
# live on the admin listener above.
assert_status "/metrics" "404"

assert_status "/" "302"

# A visitor with no locale yet lands on one by redirect, never on a page
# served directly at /: checking the status alone would also pass a
# redirect to the wrong place, so this checks where / actually sends a
# visitor, not just that it sends them somewhere.
#
# redirect_url is the absolute URL curl resolves Location to, so the match
# is anchored at both ends on the expected host. An end-only match such as
# /(de|en)/$ would also accept /foo/de/ or the same path on another host.
# Do not loosen it.
location="$(curl --silent --output /dev/null --write-out '%{redirect_url}' "${base_url}/")" || location=""
if [[ $location != "${base_url}/de/" && $location != "${base_url}/en/" ]]; then
  echo "GET ${base_url}/: want redirect to a locale path, got Location \"${location}\"" >&2
  fail=1
fi

# Both locales are real pages, not just one with the other silently
# missing: a site that serves German but not English, or the reverse,
# has to fail this script.
assert_status "/de/" "200"
assert_status "/en/" "200"

# There is no top level index.html any more: every prerendered document
# lives under its own locale's path, so this is an unknown path like any
# other and gets that locale's 404 page, same as /de/portfolio/ would.
assert_status "/index.html" "404"

if [[ $fail -ne 0 ]]; then
  echo "smoke test failed" >&2
  exit 1
fi

echo "smoke test passed"
