import rawTable from '../../../internal/routes/routes.json'

/** RouteEntry is one row of internal/routes/routes.json, the site's one
 *  route table: see internal/routes/routes.go for the Go side's own copy
 *  of this shape and who else reads the same file. */
export interface RouteEntry {
  name: string
  slug: string
  file: string
  /** One of govue's legal pages: canonical path /<locale>/<slug>, no
   *  trailing slash. Mirrors Route.Legal in routes.go. */
  legal?: boolean
}

/** LEGAL_SLUGS mirrors routes.LegalSlugs in routes.go. */
export const LEGAL_SLUGS: readonly string[] = ['imprint', 'privacy', 'terms']

/**
 * pagePath is the canonical path of a page in locale, the one rule the
 * canonical and hreflang URLs, the sitemap and the router share with the
 * server (govue's frontend package serves the sitemap's loc paths as they
 * are): a trailing slash, except for a legal page, whose path is govue's /<locale>/<page>. The home
 * page and the 404 page have an empty slug and give /<locale>/.
 */
export function pagePath(locale: string, slug: string, legal = false): string {
  if (slug === '') {
    return `/${locale}/`
  }
  return legal ? `/${locale}/${slug}` : `/${locale}/${slug}/`
}

/**
 * validateRouteTable enforces the same invariants
 * internal/routes/routes.go's Parse enforces for the Go side, so a
 * malformed table fails here too instead of silently corrupting seo.ts's
 * name-keyed slugs map (the later entry wins) or colliding with
 * vue-router's own duplicate-name handling (which drops the earlier of
 * the two routes outright).
 *
 * This cannot simply call into routes.go: Vite's JSON import above
 * resolves routes.json directly into a JS value at this module's own
 * build time, never through the Go binary, so the frontend has no Go
 * validation to inherit and needs its own. The three rules below are
 * that validator's own copy of routes.go's Parse, kept in step by
 * running both against the fixtures under internal/routes/testdata:
 * routeTable.test.ts reads the exact same files Parse's own test does,
 * so a fixture added to either side is checked by both validators
 * without further code changes.
 *
 *   - name is required and must be unique: it is the key router/index.ts
 *     (the route name) and this file's slugs/metadata lookups both use.
 *   - file is required and must be unique: two entries sharing one
 *     would make web/scripts/prerender.ts's render loop overwrite one
 *     route's prerendered output with the other's, the same "later
 *     entry wins" failure as a duplicate name, just inside the
 *     generated HTML on disk instead of a map.
 *   - slug must be unique, but may be empty: the home page's own
 *     convention. An empty slug is rejected only once a second entry
 *     repeats it, exactly like any other duplicate value.
 */
export function validateRouteTable(table: readonly RouteEntry[]): void {
  const names = new Set<string>()
  const slugs = new Set<string>()
  const files = new Set<string>()

  for (const entry of table) {
    if (entry.name === '') {
      throw new Error(`routeTable: entry with slug "${entry.slug}" has no name`)
    }
    if (entry.file === '') {
      throw new Error(`routeTable: route "${entry.name}" has no file`)
    }
    if (entry.legal && !LEGAL_SLUGS.includes(entry.slug)) {
      throw new Error(`routeTable: legal route "${entry.name}" has slug "${entry.slug}"`)
    }
    if (names.has(entry.name)) {
      throw new Error(`routeTable: duplicate name "${entry.name}"`)
    }
    names.add(entry.name)
    if (slugs.has(entry.slug)) {
      throw new Error(`routeTable: duplicate slug "${entry.slug}"`)
    }
    slugs.add(entry.slug)
    if (files.has(entry.file)) {
      throw new Error(`routeTable: duplicate file "${entry.file}"`)
    }
    files.add(entry.file)
  }
}

/** routeTable is the checked-in route table, read at build time so the
 *  browser bundle, the prerender script, the Vue router and the SEO
 *  metadata all carry the exact same list without restating it.
 *  Validated once here, at import time, so every consumer inherits the
 *  guarantee instead of each having to check it. */
export const routeTable: RouteEntry[] = rawTable
validateRouteTable(routeTable)
