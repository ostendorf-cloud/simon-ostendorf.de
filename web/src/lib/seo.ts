import type { Locale } from '../i18n'
import { SUPPORTED_LOCALES } from '../i18n'
import { pagePath, routeTable } from './routeTable'

/** The site's canonical origin, fixed at build time: the pages, sitemap.xml
 *  and robots.txt carry it, so one frontend image serves one production
 *  origin. Overridable through VITE_SITE_URL for a staging build. */
export const SITE_URL = import.meta.env.VITE_SITE_URL ?? 'https://simon-ostendorf.de'

/** NOT_FOUND is the 404 page's own pseudo route: it answers no path of
 *  its own, so it is deliberately not part of internal/routes' table
 *  (see internal/routes/routes.go) and is added here, and in
 *  router/index.ts as NOT_FOUND_ROUTE_NAME, by hand instead of being
 *  derived from it. */
export const NOT_FOUND = 'notFound' as const

/** PageRoute is every route this milestone prerenders: every name
 *  routes.json carries, plus the 404 page. A plain string rather than a
 *  literal union: routes.json is read at build time, not known to the
 *  type checker, so the real safety net is seo.test.ts, which checks
 *  every name routeTable actually carries against titleKeys and
 *  descriptionKeys below, in both directions. */
export type PageRoute = string

/** slugs maps a route to its path segment, relative to one locale's own
 *  root, derived from routeTable so a page added there needs no edit
 *  here. The 404 page has no segment of its own, the same empty value
 *  the home page's own slug happens to share. */
const slugs: Record<string, string> = Object.fromEntries(
  routeTable.map((entry): [string, string] => [entry.name, entry.slug]),
)
slugs[NOT_FOUND] = ''

const legalRoutes = new Set(routeTable.filter((entry) => entry.legal).map((entry) => entry.name))

/** path is the absolute path of route in locale, the exact path the server
 *  answers with 200: see pagePath. */
export function path(locale: Locale, route: PageRoute): string {
  return pagePath(locale, slugs[route], legalRoutes.has(route))
}

/** canonical is the absolute URL of route in locale. */
export function canonical(locale: Locale, route: PageRoute): string {
  return SITE_URL + path(locale, route)
}

/** ogLocale is the Open Graph og:locale value for locale. */
export function ogLocale(locale: Locale): string {
  return locale === 'de' ? 'de_DE' : 'en_US'
}

/**
 * Translate is a vue-i18n t function bound to one locale, for both a
 * mounted component's own useI18n().t and the prerender script's
 * standalone instance.
 */
export interface Translate {
  (key: string): string
}

export interface PageMeta {
  title: string
  description: string
}

/** titleKeys and descriptionKeys name the i18n keys each route's title
 *  and meta description come from. Kept separate from routes.json
 *  deliberately: routes.json is embedded into the Go binary via
 *  go:embed and read by Go code that has no use for an i18n key;
 *  coupling the Go-owned route table's schema to frontend-only i18n key
 *  strings would leak a frontend concern into a shared, Go-consumed
 *  resource. Exported so seo.test.ts can check they cover exactly
 *  routeTable's names plus NOT_FOUND. */
export const titleKeys: Record<string, string> = {
  home: 'home.meta.title',
  cv: 'cv.meta.title',
  contact: 'contact.meta.title',
  imprint: 'legal.imprint.title',
  privacy: 'legal.privacy.title',
  notFound: 'notFound.title',
}

export const descriptionKeys: Record<string, string> = {
  home: 'home.meta.description',
  cv: 'cv.meta.description',
  contact: 'contact.meta.description',
  imprint: 'legal.imprint.metaDescription',
  privacy: 'legal.privacy.metaDescription',
  notFound: 'notFound.lead',
}

/** pageMeta resolves the title and meta description of route through t. */
export function pageMeta(route: PageRoute, t: Translate): PageMeta {
  return { title: t(titleKeys[route]), description: t(descriptionKeys[route]) }
}

/** fullTitle is the browser tab title: the page's own title for the home
 *  page, and "<title> – <site name>" for every other page. */
export function fullTitle(route: PageRoute, meta: PageMeta, siteName: string): string {
  return route === 'home' ? meta.title : `${meta.title} – ${siteName}`
}

/** hreflangAlternates lists every locale's canonical URL for route, plus
 *  the x-default entry pointing at German, this site's own default when a
 *  visitor's language is not one this site has a dedicated version for. */
export function hreflangAlternates(route: PageRoute): Array<{ hreflang: string; href: string }> {
  const alternates: Array<{ hreflang: string; href: string }> = SUPPORTED_LOCALES.map((locale) => ({
    hreflang: locale,
    href: canonical(locale, route),
  }))
  alternates.push({ hreflang: 'x-default', href: canonical('de', route) })
  return alternates
}
