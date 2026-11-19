# simon-ostendorf.de

The personal hub at [simon-ostendorf.de](https://simon-ostendorf.de): the way to development
or event technology, plus the CV.

## What it does

A Go service serves a Vue 3 single page application built on the organisation's own `govue`
framework, prerendered to static HTML in German and English from one shared route table. The
page shell carries the navigation, the footer and the theme switch; none of the five pages'
own content exists yet. See `AGENTS.md`'s "Scope of this milestone" for what is planned and
not yet built.

## Install

```sh
task setup
```

Installs the pinned toolchain (Go, Node, pnpm, and the lint and release tooling) through
`mise`.

## Use

```sh
task web:build   # builds the SPA into cmd/website/dist
task prerender    # renders every route and locale into cmd/website/prerendered
task build        # compiles the website binary into bin/website
bin/website
```

The server reads its configuration from the environment; see the Configuration table below.
There is no secret-shaped variable yet. `/healthz`, `/readyz` and `/metrics` answer on a
second, administrative listener (`ADMIN_ADDR`), never on the public one.

## Configuration

| Variable | Purpose | Default | Required |
| --- | --- | --- | --- |
| `ADDR` | Address the public HTTP server listens on | `:8080` | no |
| `ADMIN_ADDR` | Address govue's admin listener listens on: `/healthz`, `/readyz` and `/metrics` | `:8081` | no |
| `SITE_URL` | This site's own public origin, the base `/robots.txt` and `/sitemap.xml` build every absolute URL from | `https://simon-ostendorf.de` | no |
| `LOG_LEVEL` | `debug`, `info`, `warn` or `error` | `info` | no |
| `UMAMI_URL` | Base URL of the Umami instance. Counting is on only when this and `UMAMI_WEBSITE_ID` are both set; one alone is a startup error. Not a secret | unset, counting off | no |
| `UMAMI_WEBSITE_ID` | The website's UUID in Umami. Not a secret | unset, counting off | no |
| `UMAMI_TIMEOUT` | Bound of one post to Umami, 1s to 30s | `5s` | no |
| `UMAMI_QUEUE_SIZE` | Views waiting to be sent, 1 to 10000. A full queue drops the view | `256` | no |
| `TRUSTED_PROXIES` | Comma separated CIDRs of the reverse proxies whose `X-Forwarded-For` is believed. Needed behind an ingress for counting and the access log to see visitors | empty, trust nothing | no |

## Page view counting

Page views are counted on the server with [Umami](https://umami.is), through govue's
`analytics` package. No script runs in the browser and nothing reaches it.

What is counted, and what is not. This is the factual basis for the privacy page
(`web/src/legal/PrivacyText*.vue`), which states the same in both languages. Change both
together:

- A view is counted after a page was served: a `GET` that answered 200 with an HTML page.
- The URL reported is the route pattern, for example `/de/cv/`. The query string, an identifier
  and the raw request path are never sent.
- No cookie is set and nothing is read from the visitor's device.
- The visitor's address is masked to its /24 (IPv4) or /48 (IPv6) before it is sent, and is
  never stored by Umami. The application's own access log is separate, see below.
- Also sent: the user agent, the first language tag of `Accept-Language`, and the origin
  (scheme, host, port, no path) of another site that linked here.
- A visitor who sends `DNT: 1` or `Sec-GPC: 1` is not counted at all.
- A prefetch, a request that is not a document load, a bot, a client on a private address, a
  redirect, a 404, an asset, `robots.txt`, `sitemap.xml` and the health and metrics routes are
  not counted.

Counting is off unless `UMAMI_URL` and `UMAMI_WEBSITE_ID` are both set, and the site logs once
at startup which it is. Behind a reverse proxy set `TRUSTED_PROXIES`, or every request looks
like it came from the proxy and none is counted. A slow or unreachable Umami never delays a
response: views go through a bounded queue and are dropped when it is full, counted in the
`analytics_views_dropped_total` metric.

## Access log

The application writes one structured line per request through govue's `middleware.AccessLog`:
method, route pattern (never the raw path or query), status, bytes, duration, the client address
and the request id. The client address is resolved through `TRUSTED_PROXIES`, so behind an
ingress it is the visitor and not the proxy. No user agent and no referrer are logged. The
health routes of the admin listener are not on this handler. The privacy page (2.1) states
these fields, and separately the cluster ingress log, which is outside this repository. Both
logs are kept 30 days, operated by the owner. Change the fields and the text together.

## Container

```sh
docker compose up --build
```

Builds the multi-stage `Dockerfile` (frontend, static Go binary, then a minimal non-root
runtime stage with no shell) and serves the site at `http://localhost:8080`.
`docker-compose.yaml` is for local development only; see `charts/simon-ostendorf.de/README.md`
for a cluster installation.

## Development

```sh
task setup
task check
```

`task check` runs every lint, build and test target one after another; `task --list` shows
them individually. `task web:build` and `task prerender` both have to run before `task build`
or `task test:go` can compile `cmd/website`, since it embeds both outputs; `task build` and
`task test:go` already declare that dependency, so running either one through Task is enough.

Releases are driven by [release-me](https://github.com/ostendorf-cloud/release-me):
Conventional Commits decide the version, and merging the release pull request it maintains
cuts the release. `CHANGELOG.md` is generated, never hand edited.

## License

Proprietary, all rights reserved. See `LICENSE` and `NOTICE`.
