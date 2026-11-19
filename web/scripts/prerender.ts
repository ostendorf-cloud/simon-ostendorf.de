/**
 * Renders every route of every locale to static HTML, so a crawler and a
 * browser's first paint both get real markup instead of an empty <div
 * id="app">. Run through "task prerender", after "task web:build" has
 * written the SPA bundle and its manifest into cmd/website/dist.
 *
 * The Go service embeds the files this script writes into
 * cmd/website/prerendered and serves them directly; the SPA bundle this
 * same page loads then hydrates over the server-rendered markup rather
 * than replacing it, through createSSRApp on both sides.
 */
import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

import { createSSRApp } from 'vue'
import { renderToString } from 'vue/server-renderer'
import { createMemoryHistory } from 'vue-router'
import { useTheme } from '@ostendorf-cloud/govue'

import App from '../src/App.vue'
import { createAppI18n, installI18n, SUPPORTED_LOCALES, type Locale } from '../src/i18n'
import { routeTable, type RouteEntry } from '../src/lib/routeTable'
import {
  canonical,
  SITE_URL,
  fullTitle,
  hreflangAlternates,
  NOT_FOUND,
  ogLocale,
  pageMeta,
  type PageRoute,
} from '../src/lib/seo'
import { THEME_NAME } from '../src/lib/theme'
import { attachLocaleSync, createAppRouter, NOT_FOUND_ROUTE_NAME } from '../src/router'

const here = dirname(fileURLToPath(import.meta.url))
const distDir = join(here, '../../cmd/website/dist')
const outDir = join(here, '../../cmd/website/prerendered')

// The same call main.ts makes, so the theme attribute reaches the
// prerendered <head> without a second copy of the theme name living
// here. There is no document here, so the attribute writing half of
// useTheme does nothing; renderDocument below writes that attribute
// itself.
useTheme({ theme: THEME_NAME })

/** loadPages turns the shared route table into this script's own page
 *  list, pairing each entry's name with the router route and the SEO
 *  route it names (the same string serves both now that PageRoute is a
 *  plain string), plus the one page the table itself does not carry: the
 *  404 page, which answers no route and so is not part of the table
 *  internal/webapp's sitemap and mux registration read either. */
function loadPages(): Array<{ routerName: string; seoRoute: PageRoute; file: string }> {
  const pages = routeTable.map((entry) => ({
    routerName: entry.name,
    seoRoute: entry.name,
    file: entry.file,
  }))
  pages.push({ routerName: NOT_FOUND_ROUTE_NAME, seoRoute: NOT_FOUND, file: '404.html' })
  return pages
}

interface ManifestEntry {
  file: string
  css?: string[]
}

interface Assets {
  script: string
  styles: string[]
}

/** stylesheets is every stylesheet the build wrote, whatever shape the
 *  manifest records it in: as a chunk's own "css" list, or as a
 *  standalone entry when the build does not split CSS at all.
 *
 *  All of them, not the entry's alone. A prerendered page is served as
 *  static HTML and the module graph that would pull a route chunk's own
 *  stylesheet in runs after first paint, or never at all for a visitor
 *  with no JavaScript. Linking only the entry's CSS left a view with a
 *  <style> block of its own painted unstyled until then; linking all of
 *  them is correct for any view added later and for either setting of
 *  build.cssCodeSplit, rather than only for the one this repository
 *  happens to use. */
function stylesheets(manifest: Record<string, ManifestEntry>): string[] {
  const entry = manifest['index.html']
  const found = new Set<string>(entry?.css ?? [])
  for (const record of Object.values(manifest)) {
    for (const file of record.css ?? []) {
      found.add(file)
    }
    if (record.file.endsWith('.css')) {
      found.add(record.file)
    }
  }
  if (found.size === 0) {
    throw new Error('prerender: the vite manifest records no stylesheet, run task web:build first')
  }
  // The entry's own sheet first, so a split build still cascades from the
  // base styles outwards.
  return [...found]
}

async function loadManifest(): Promise<Assets> {
  const raw = await readFile(join(distDir, '.vite/manifest.json'), 'utf-8')
  const manifest = JSON.parse(raw) as Record<string, ManifestEntry>
  // The manifest keys its one HTML entry by the input file vite built,
  // index.html, not by the script tag inside it: that script's own
  // bundle is this entry's "file".
  const entry = manifest['index.html']
  if (!entry) {
    throw new Error('prerender: no index.html entry in the vite manifest, run task web:build first')
  }
  return { script: entry.file, styles: stylesheets(manifest) }
}

/** renderBody server-renders the app at routerName, under locale, with a
 *  fresh router and i18n instance per page: nothing from one page's
 *  render may leak into the next. */
async function renderBody(locale: Locale, routerName: string): Promise<string> {
  const i18n = createAppI18n(locale)
  const router = createAppRouter(createMemoryHistory())
  attachLocaleSync(router, i18n.global.locale)

  const app = createSSRApp(App)
  app.use(router)
  installI18n(app, i18n)

  await router.push({ name: routerName, params: { locale } })
  await router.isReady()

  return renderToString(app)
}

function renderHead(locale: Locale, route: PageRoute, assets: Assets): string {
  const i18n = createAppI18n(locale)
  const t = i18n.global.t
  const meta = pageMeta(route, t)
  const siteName = t('common.siteName')
  const title = fullTitle(route, meta, siteName)
  const url = canonical(locale, route)

  const alternates = hreflangAlternates(route)
    .map((a) => `<link rel="alternate" hreflang="${a.hreflang}" href="${a.href}">`)
    .join('\n    ')
  const styleLinks = assets.styles
    .map((href) => `<link rel="stylesheet" href="/${href}">`)
    .join('\n    ')

  return `<meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>${escapeHtml(title)}</title>
    <meta name="description" content="${escapeHtml(meta.description)}">
    <meta name="robots" content="index, follow">
    <link rel="canonical" href="${url}">
    ${alternates}
    <meta property="og:type" content="website">
    <meta property="og:locale" content="${ogLocale(locale)}">
    <meta property="og:site_name" content="${escapeHtml(siteName)}">
    <meta property="og:title" content="${escapeHtml(title)}">
    <meta property="og:description" content="${escapeHtml(meta.description)}">
    <meta property="og:url" content="${url}">
    ${styleLinks}
    <script type="module" src="/${assets.script}" defer></script>`
}

/** renderDocument wraps one rendered page. data-theme is the one attribute
 *  govue's stylesheets switch the palette on, written here rather than
 *  left to useTheme so the very first paint of a prerendered page already
 *  carries it. There is no data-mode attribute: this site pins no mode of
 *  its own, see src/lib/theme.ts, so the absence here is useTheme's own
 *  documented "follow the operating system" state, exactly the starting
 *  point this site wants. */
function renderDocument(locale: Locale, head: string, body: string): string {
  return `<!doctype html>
<html lang="${locale}" data-theme="${THEME_NAME}">
  <head>
    ${head}
  </head>
  <body>
    <div id="app">${body}</div>
  </body>
</html>
`
}

/** routeUrl is the absolute sitemap URL of a route table entry in locale:
 *  the same rule internal/webapp's routeURL applies, a trailing slash for
 *  every page except a legal one, whose canonical path is govue's
 *  /<locale>/<page>. */
function routeUrl(locale: Locale, entry: RouteEntry): string {
  if (entry.slug === '') {
    return `${SITE_URL}/${locale}/`
  }
  return `${SITE_URL}/${locale}/${entry.slug}${entry.legal ? '' : '/'}`
}

/** sitemapXml renders the sitemap govue-check-legal reads. The server serves
 *  its own sitemap.xml, built from the same route table and SITE_URL;
 *  cmd/website's TestPrerenderedSitemapMatchesServedSitemap fails when the
 *  two differ in a single byte, so this file is a build input for the check
 *  and never a second source of truth. */
function sitemapXml(): string {
  const entries = [...routeTable].sort((a, b) => (a.slug < b.slug ? -1 : a.slug > b.slug ? 1 : 0))
  const lines = [
    '<?xml version="1.0" encoding="UTF-8"?>',
    '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9" xmlns:xhtml="http://www.w3.org/1999/xhtml">',
  ]
  for (const entry of entries) {
    for (const locale of SUPPORTED_LOCALES) {
      lines.push('  <url>', `    <loc>${routeUrl(locale, entry)}</loc>`)
      for (const alt of SUPPORTED_LOCALES) {
        lines.push(
          `    <xhtml:link rel="alternate" hreflang="${alt}" href="${routeUrl(alt, entry)}"/>`,
        )
      }
      lines.push(
        `    <xhtml:link rel="alternate" hreflang="x-default" href="${routeUrl('de', entry)}"/>`,
        '  </url>',
      )
    }
  }
  lines.push('</urlset>', '')
  return lines.join('\n')
}

function escapeHtml(value: string): string {
  return value
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
}

async function main(): Promise<void> {
  const assets = await loadManifest()
  const pages = loadPages()

  for (const locale of SUPPORTED_LOCALES) {
    for (const page of pages) {
      const body = await renderBody(locale, page.routerName)
      const head = renderHead(locale, page.seoRoute, assets)
      const html = renderDocument(locale, head, body)

      const target = join(outDir, locale, page.file)
      await mkdir(dirname(target), { recursive: true })
      await writeFile(target, html, 'utf-8')
      console.log(`wrote ${target}`)
    }
  }

  const sitemap = join(outDir, 'sitemap.xml')
  await writeFile(sitemap, sitemapXml(), 'utf-8')
  console.log(`wrote ${sitemap}`)
}

main().catch((error: unknown) => {
  console.error(error)
  process.exitCode = 1
})
