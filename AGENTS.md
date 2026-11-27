# simon-ostendorf.de

Simon Ostendorf's personal domain: the way to development or event technology, plus the CV.

The global baseline already applies here and is not repeated below.

## Layout

- `cmd/website` wires the service: config, construction, shutdown. `embed.go` embeds
  `dist/` (the built SPA, written by `task web:build`) and `prerendered/` (the static HTML
  per route and locale, written by `task prerender`). Both directories sit under `cmd/`
  rather than `internal/` because `go:embed` resolves relative to the file holding the
  directive. `routes_test.go`'s `TestPrerenderedOutputMatchesRouteTable` walks the real
  `prerendered/` output and checks it against `internal/routes`' own table, catching both a
  route with no prerendered file and a prerendered file that answers no route; `task
  test:routes` runs it on its own.
- `internal/config` loads and validates the website's own environment once, aggregated into
  one error. `SiteURL` is the base `internal/webapp`'s sitemap and `robots.txt` build every
  absolute URL from. There is no secret-shaped variable: no session, no OIDC, no CSRF and no
  form exist in this milestone. Both `simon-ostendorf.de` and `www.simon-ostendorf.de` exist in
  production, with `www` redirecting to the apex; `SiteURL`'s own default, `.env.example` and
  the frontend's `web/src/lib/seo.ts` all name the apex as the canonical host, while the
  chart's `ingress.hosts` default serves both, which is why the two appear in different places.
- `internal/routes` is the website's one route table: `routes.json`, checked in and embedded
  through `//go:embed`, lists every page as a name, a slug and the prerendered file that
  answers it. `internal/webapp`, `web/scripts/prerender.ts`, and the frontend's own
  `web/src/router/index.ts` and `web/src/lib/seo.ts`, through the shared
  `web/src/lib/routeTable.ts`, each read this one file to build their own list instead of
  restating it, so a page added to one side and not the other can no longer happen.
  `routes.go`'s `Parse` rejects a table with a missing name or file, or a repeated name, slug
  or file, before any of those five consumers ever sees it; `routeTable.ts`'s own
  `validateRouteTable` enforces the same three rules for the frontend, which reads
  `routes.json` through Vite's JSON import rather than through Go. `internal/routes/testdata`
  holds the fixtures both `routes_test.go` and `routeTable.test.ts` read, so a fixture either
  side adds is checked by both validators without touching the other test file.
- `internal/webapp` builds two HTTP handlers from govue's packages. `Handler` is the public
  one: the SPA's static assets, the prerendered page for every known route and locale,
  `robots.txt` and `sitemap.xml`, and the catch-all document handler, behind panic recovery,
  a request id and govue's access log (`middleware.AccessLog`: one line per request with the
  route pattern, the client address resolved through `TRUSTED_PROXIES`, no user agent; the
  privacy text 2.1 names exactly these fields, change both together). The chain is `Recover`,
  `RequestID`, the access log, then govue's `analytics` middleware, which counts page views
  on the server (see the README's "Page view counting"). Analytics stays innermost, directly
  around the mux: it and the access log read the route pattern the mux sets, and `RequestID`
  replaces the request, so nothing may sit between analytics and the mux. `pages.go` is that
  catch-all
  handler: the locale redirect at `/`, trailing slash canonicalisation, and serving the
  matching prerendered file or a locale's own 404 page, all built from `internal/routes`'
  embedded table. `sitemap.go` serves `robots.txt` and `sitemap.xml` from that same table,
  each `<url>` carrying an hreflang alternate for both locales. `AdminHandler` is the
  second, smaller handler govue's admin listener serves: health, readiness and metrics,
  reached only from inside the cluster, never through the chart's own Ingress.
- The legal pages follow govue's scheme. `routes.json` marks the imprint and the privacy page
  `"legal": true`: their canonical path is `/<locale>/<page>` with no trailing slash (every
  other page keeps it), served by an exact pattern, with the slash form redirecting to it. The
  sitemap lists the same path. `language.HandleLegal` (outside any wrapper, `Terms` off)
  registers `/imprint`, `/privacy` and the German aliases `/impressum`, `/datenschutz`,
  `/datenschutzerklaerung` as 302 redirects, which analytics never counts. This site has no
  terms, so `/terms` and `/agb` do not exist: add a terms page, `Terms: true` and
  `--require-terms` together if that changes. `web/src/legal` holds the site's own texts, one
  component per page and language, and `operator.ts`, the one place the operator facts live.
  The footer is `GovueFooter` with `legal`. `task prerender` also writes `sitemap.xml` into
  `prerendered/` and ends with `govue-check-legal`; `cmd/website`'s
  `TestPrerenderedSitemapMatchesServedSitemap` pins that file to the sitemap the server
  serves. A change to what the site stores or sends (a form, a cookie, a third party) must
  change the privacy texts in both languages in the same pull request.
- `internal/depsync` checks that `go.mod` and `web/package.json` pin the same govue version,
  and that `web/pnpm-lock.yaml` still resolves govue's Vue package to its `/web` subdirectory.
  It also checks the Dockerfile's pnpm pin against `mise.toml` and the Renovate rules that
  group the pnpm pins and hold `typescript` below 7.
- `web/` is the Vue SPA, consuming govue's Vue package as a pnpm git dependency. `web/src/i18n`
  holds the German and English message catalogues (German is the source of truth; English
  falls back to German for any key not yet translated) and `web/src/i18n/locales.test.ts`
  fails the build on a key present in one catalogue and missing from the other.
  `web/src/router/index.ts` builds one vue-router route per `routes.json` entry, plus the
  locale redirect and the 404 route. `web/src/lib/seo.ts` resolves each route's title, meta
  description and hreflang alternates from the same table. `web/src/views` holds this
  milestone's five pages (home, cv, contact, imprint, privacy); the home page is the landing
  page with two plain external links (ostendorf.dev, ostendorf.tech), cv and contact are
  a heading and nothing else, their content is a separate issue, see "Scope of this
  milestone" below. `web/src/components/AppHeader.vue` and `AppFooter.vue` carry the navigation, the
  locale switch and the theme toggle. `web/scripts/prerender.ts` is the static site
  generation step `task prerender` runs, server-rendering every route of every locale into
  `cmd/website/prerendered`. `web/src/theme/simon-ostendorf-de.css` is this site's theme
  file, filling govue's token contract. It is generated from the two seed colours in
  `web/src/theme/seeds.json` (accent, neutral) by govue's theme generator: change a seed, then
  run `pnpm run theme:generate` in `web/` and commit both files. `contrast.test.ts` fails a
  text pair below WCAG 2.2 AA in either mode.
- `charts/simon-ostendorf.de/` is the Helm chart. `charts/simon-ostendorf.de/README.md` covers
  installing and configuring it.

## A build needs the frontend first

`cmd/website` cannot compile against an empty `dist/` or `prerendered/` directory, because
`go:embed` fails when its pattern matches no file at all. `dist/` and `prerendered/` are both
gitignored, generated by `task web:build` and `task prerender` respectively.

`task build` and `task test:go` both declare that dependency, so running either one through
Task always populates the frontend first. A bare `go build ./cmd/website` or
`go test ./cmd/website/...` outside Task fails on a fresh checkout for that reason; run
`task build` or `task test:go` instead, or `task web:build && task prerender` once by hand
before reaching for the Go toolchain directly.

## govue's two halves move together

govue releases a Go module and a Vue package under one tag. `go.mod`'s
`require github.com/ostendorf-cloud/govue` and `web/package.json`'s `@ostendorf-cloud/govue`
pnpm git dependency must always name the same tag: a mismatch means the repository takes half
of a govue release and not the other half, silently. `task check:govue-version`
(`internal/depsync`) enforces this; see that package for the detail, including a real pnpm bug
(pnpm/pnpm#12304) it also guards against.

Renovate proposes both halves as one pull request. `renovate.json` has a regex manager for
the pnpm git dependency, which Renovate's npm manager skips. One rule groups both halves as
`govue`. The shared `govue-lockfile` preset updates `web/pnpm-lock.yaml`. The rule's
`constraints.pnpm` must equal the pnpm pin in `mise.toml`, and `internal/depsync` tests it.
A second rule groups the three pnpm pins as `pnpm`, so they move in one pull request: the
`mise.toml` pin, the Dockerfile's `ARG PNPM_VERSION` (a regex manager reads it) and the govue
rule's `constraints.pnpm`. `TestDockerfilePnpmMatchesMise` fails when the Dockerfile and
`mise.toml` differ.

## Verify before committing

```sh
task check
```

`task --list` shows the individual targets, including the small ones `task check` aggregates.

## Scope of this milestone

The tooling, CI, the Helm chart, the container build, the bilingual page shell, the one route
table and its prerendering are in place, and so are the imprint and privacy pages. Each of
the two other pages the shell routes to (cv, contact) carries a heading and
nothing else; the actual content is each its own issue and none of it is implemented yet:

- Done: the landing page, a short page about the person with two clear ways on: development
  at ostendorf.dev, event technology at ostendorf.tech.
- The CV as a page, rendered from the data the `cv` repository publishes, with the PDF offered
  for download.
- Contact, with a proof of work challenge and SMTP delivery.
- Done: the imprint and privacy texts, this domain's own and written for what this site does
  today. The contact page will add a form and mail delivery: the privacy texts must grow with
  it, they now say there is no form.

## Deviations

- `LICENSE` is proprietary, all rights reserved, not the baseline's Apache-2.0: this is Simon
  Ostendorf's own personal domain, and reusing it would misidentify the source of a service.
  No per-file licence header and no reusable license-header workflow for the same reason; the
  single `LICENSE` carries it. `NOTICE` records what `LICENSE` does not cover.
- The Helm chart lives at `charts/simon-ostendorf.de/`, matching the repository's own name, but
  its `Chart.yaml` `name` and every rendered object name use `simon-ostendorf-de` instead: a
  literal dot is invalid in a `Service`'s `metadata.name` (RFC 1035 label), which a chart named
  `simon-ostendorf.de` would otherwise produce by default.
- `helm-unittest` is pinned in `mise.toml` through mise's generic `github:` backend with a
  `postinstall` hook, not a plain registry entry: it is not in the mise registry, and
  `helm plugin install <url>` is forbidden by the baseline regardless of Helm version.
- The chart ships no `existingSecret` block, no `replicaCount` upper bound or
  `strategy.type: Recreate`: this milestone has no secret-shaped configuration and no
  in-process session state. Add each guard back only once the application actually needs it.
- `config.trustedProxies` (`TRUSTED_PROXIES`) exists because page view counting and the
  access log resolve the visitor address, but the chart has no required-when-ingress guard for
  it: an empty value only means no visitor is counted and the access log shows the proxy's
  address behind an ingress.
