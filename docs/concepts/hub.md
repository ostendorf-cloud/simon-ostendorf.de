# Personal hub: about, CV, contact, imprint, privacy

## Problem

- `www.simon-ostendorf.de` has no content yet, only the govue scaffold and a placeholder page.
- The domain used to carry the event technology business, now at ostendorf.tech, and is
  deliberately not redirected: ostendorf.dev (development) also exists, and neither sibling
  brand should swallow the person's own name.
- A visitor landing on this domain today has no way to find the person, the CV, a contact
  channel, or the two sibling sites.
- This domain needs its own imprint and privacy content: it is not the business entity
  ostendorf.tech is, and is not redirected to it.

## Goals

- A short landing page about the person, linking out to ostendorf.dev and ostendorf.tech.
- A CV page rendered from the cv repository's public data, with the PDF offered for download.
- A contact page.
- An imprint and a privacy page that describe this domain's own, actual processing.
- Every page passes the same `task check` this repository already runs.

## Non goals

- No project, skill or technology portfolio content: that lives on ostendorf.dev.
- No event technology service content, and no second copy of that business's own imprint,
  privacy or AGB: that lives on ostendorf.tech.
- No authentication, session, or form backend beyond what contact needs.
- No change to the cv repository's own data model, API or release pipeline.
- No ADR in this change: no decision below is settled enough yet to record as one.

## Approach

- Five pages: about/hub, CV, contact, imprint, privacy. About is a short landing page with
  two links out, not a duplicate of either sibling's own content.
- **CV data: build time, not runtime.** This site's CI pulls the cv repository's published
  JSON release artefact as a build input, the same shape `internal/depsync` already uses to
  pin and check a govue version. A new cv release reaches this site only once this site
  rebuilds against it.
  - Reason: this is a low traffic personal page, not a business-critical service. Adding a
    runtime dependency on the cv API service's uptime (`GET /api/v1/cv`) buys content that
    updates without a rebuild, at the cost of a second in-cluster dependency, a client with
    its own cache and timeout handling (the shape ostendorf.tech's `internal/feedclient`
    needs for the Instagram feed), and a network policy. That cost is not worth it for a page
    that changes only when the CV itself changes.
    - ostendorf.tech's feed pattern degrades by rendering nothing when the feed service is
      unreachable. On that site, the Impressionen section is decorative, so an empty
      section is an acceptable failure mode. The CV page's whole purpose is the CV, so the
      same degrade-to-empty strategy would make the page's only content disappear instead.
  - Build time also fits the prerendering decision below: a page rendered to static HTML at
    build time wants its data at build time, not a page load.
- **PDF delivery: pull the PDF, do not render it.** The PDF offered for download is the cv
  repository's own release artefact, pulled alongside the JSON at build time and served as a
  static file. Rendering a PDF per request is avoided: it buys always-current output at a
  render cost this traffic level does not need, and the result would still go stale the
  moment the JSON does, since both come from the same cv release. Tied directly to the
  build-time choice above: whichever cadence applies to the JSON applies to the PDF.
- **Visibility: filtered at the source, not in this site.** The public CV page and the
  downloadable PDF only ever carry entries the cv data model marks public. An entry marked
  non-public, such as a postal address or an NDA-bound client project, is never sent to this
  site at all. Filtering happens in the cv repository's own release pipeline, before this
  site ever reads the artefact, not as a client-side rule in this site's frontend: a browser's
  developer tools can always see what the frontend received, so a field that must never reach
  a visitor must never be sent in the first place.
- **Prerendering: yes, following ostendorf.tech's pattern.** A build step renders every route
  to static HTML (`task prerender`-equivalent), embedded into the binary and served ahead of
  the SPA, with a route table and a test proving every table entry has a matching prerendered
  file. Reason: the CV page is exactly the kind of content worth indexing, prerendering is
  already proven in this organisation's own govue scaffold family, and it is a natural fit
  once CV data is a build-time input rather than a page-load fetch.
- **Imprint and privacy: this domain's own texts, on govue's legal routes.** Both pages are
  written for what this site does and state only what the code and the repository documents
  back. The imprint names the operator (name, address, phone, e-mail, VAT ID) and has no
  dispute resolution section. The privacy page covers the controller, what the site itself
  stores (no cookie, the theme choice in local storage), the server side page view counting
  from the README's section "Page view counting", contact by e-mail or phone, the rights of
  data subjects and the right to complain. Hosting, processors, retention and the legal basis
  are not stated: no repository document backs them yet. The Instagram section does not
  apply. Served at `/de/imprint`, `/en/imprint`, `/de/privacy` and `/en/privacy` with govue's
  redirects from `/imprint`, `/privacy` and the German aliases. No terms page: this site
  sells nothing.
- Runs on the existing govue scaffold: the public `Handler` serves the five pages and the
  contact endpoint, govue's admin listener keeps health, readiness and metrics internal, and
  the Helm chart already in this repository ships it.

## Alternatives

- **Runtime fetch from `GET /api/v1/cv` on every page load.** Dropped: makes this site's
  uptime depend on another service's uptime for no real benefit at this traffic level, and
  does not fit prerendering without also caching at build time anyway, which is the chosen
  approach already.
- **Runtime fetch with a feedclient-style in-memory cache and periodic sync.** Dropped for the
  same reason as the plain runtime fetch, plus the extra code a sync loop, cache and network
  policy need, for content that changes on the order of cv releases, not minutes.
- **Render the PDF on demand per request.** Dropped: adds render cost and a new failure mode
  for a document that is only ever as current as the last cv release regardless, since the
  JSON and the PDF share one source.
- **Client-side filtering of a full, unfiltered cv payload.** Dropped: relies on the frontend
  never rendering a field it was never supposed to show, which a browser's own developer
  tools defeat trivially; the cv data model's visibility marker exists specifically to avoid
  this.
- **Redirect this domain to ostendorf.tech or ostendorf.dev.** Dropped before this document:
  the domain carries the person's own name, and neither sibling brand should absorb it.

## Open questions

- How does a cv release actually trigger a rebuild here: a manual bump, a Renovate-style
  version bump pull request, or a repository dispatch from the cv repository. Owner: decide
  once the cv repository's release automation is confirmed.
- Does the cv repository's published JSON release artefact already exclude non-public entries,
  or does only its live API filter them. Owner: confirm with the cv repository's maintainer
  before this site's build step is written.
- Does the cv repository's release actually publish a PDF artefact, or only render one on
  demand through its own API. Owner: confirm with the cv repository's maintainer.
- One locale or two. ostendorf.tech runs German and English through a shared route table;
  this site currently has no router or i18n setup at all. Owner: decide before the route
  table and prerender step are built.
- What address the imprint and privacy pages disclose as the responsible party's address.
  A private individual's residential address raises different privacy considerations than a
  business premises, which is itself why the cv data model keeps the postal address entry
  non-public. Owner: the site's operator, before the imprint page is written.
- Exact contact channel and address shown on the contact page. Owner: the site's operator.
