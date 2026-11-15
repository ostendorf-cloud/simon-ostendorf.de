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
  `robots.txt` and `sitemap.xml`, and the catch-all document handler, behind panic recovery
  and a request id. `pages.go` is that catch-all handler: the locale redirect at `/`,
  trailing slash canonicalisation, and serving the matching prerendered file or a locale's
  own 404 page, all built from `internal/routes`' embedded table. `sitemap.go` serves
  `robots.txt` and `sitemap.xml` from that same table, each `<url>` carrying an hreflang
  alternate for both locales. `AdminHandler` is the second, smaller handler govue's admin
  listener serves: health, readiness and metrics, reached only from inside the cluster, never
  through the chart's own Ingress.
- `internal/depsync` checks that `go.mod` and `web/package.json` pin the same govue version,
  and that `web/pnpm-lock.yaml` still resolves govue's Vue package to its `/web` subdirectory.
- `web/` is the Vue SPA, consuming govue's Vue package as a pnpm git dependency. `web/src/i18n`
  holds the German and English message catalogues (German is the source of truth; English
  falls back to German for any key not yet translated) and `web/src/i18n/locales.test.ts`
  fails the build on a key present in one catalogue and missing from the other.
  `web/src/router/index.ts` builds one vue-router route per `routes.json` entry, plus the
  locale redirect and the 404 route. `web/src/lib/seo.ts` resolves each route's title, meta
  description and hreflang alternates from the same table. `web/src/views` holds this
  milestone's five page shells (home, cv, contact, imprint, privacy), each a heading and
  nothing else: a page's actual content is a separate issue, see "Scope of this milestone"
  below. `web/src/components/AppHeader.vue` and `AppFooter.vue` carry the navigation, the
  locale switch and the theme toggle. `web/scripts/prerender.ts` is the static site
  generation step `task prerender` runs, server-rendering every route of every locale into
  `cmd/website/prerendered`. `web/src/theme/simon-ostendorf-de.css` is this site's theme
  file, filling govue's token contract; see its own comment for why it carries a borrowed,
  placeholder palette rather than a real brand colour yet.
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

## Verify before committing

```sh
task check
```

`task --list` shows the individual targets, including the small ones `task check` aggregates.

## Scope of this milestone

The tooling, CI, the Helm chart, the container build, the bilingual page shell, the one route
table and its prerendering are in place. Each of the five pages the shell routes to
(home/about, cv, contact, imprint, privacy) carries a heading and nothing else; the actual
content is each its own issue and none of it is implemented yet:

- A short page about the person, with two clear ways on: development at ostendorf.dev, event
  technology at ostendorf.tech.
- The CV as a page, rendered from the data the `cv` repository publishes, with the PDF offered
  for download.
- Contact, with a proof of work challenge and SMTP delivery.
- Server side page view counting with Umami.
- Imprint and privacy content. This domain needs its own, unlike ostendorf.tech's.
  `old_simon-ostendorf.de` was this domain's previous site, the event technology business
  before it moved to ostendorf.tech, and carried this domain's previous imprint and privacy
  pages; ostendorf.tech has already ported most of that content. Prefer ostendorf.tech's own
  version over old_simon-ostendorf.de's when the two differ, since it is newer, once this
  milestone is actually implemented. None of that content is ported here now.

## Deviations

- `LICENSE` is proprietary, all rights reserved, not the baseline's Apache-2.0: this is Simon
  Ostendorf's own personal domain, and reusing it would misidentify the source of a service.
  No per-file licence header and no reusable license-header workflow for the same reason; the
  single `LICENSE` carries it. `NOTICE` records what `LICENSE` does not cover.
- `.github/workflows/dco.yaml`, `renovate.yaml` and `checks.yaml`'s `renovate-config` job call
  `ostendorf-cloud/workflows`'s reusable workflows at the exact tag `v0.2.2`, the newest tag
  that exists, not a floating `v0` major tag: `ostendorf-cloud/workflows` has not cut one yet.
  Repoint at `v0` once it exists.
- The Helm chart lives at `charts/simon-ostendorf.de/`, matching the repository's own name, but
  its `Chart.yaml` `name` and every rendered object name use `simon-ostendorf-de` instead: a
  literal dot is invalid in a `Service`'s `metadata.name` (RFC 1035 label), which a chart named
  `simon-ostendorf.de` would otherwise produce by default.
- `helm-unittest` is pinned in `mise.toml` through mise's generic `github:` backend with a
  `postinstall` hook, not a plain registry entry: it is not in the mise registry, and
  `helm plugin install <url>` is forbidden by the baseline regardless of Helm version.
- The chart ships no `existingSecret` block, no `replicaCount` upper bound or
  `strategy.type: Recreate`, and no `config.trustedProxyCidrs` required-when-ingress guard:
  this milestone has no secret-shaped configuration, no in-process session state, and no
  client-IP-based rate limiting yet. Add each guard back only once the application actually
  needs it.
