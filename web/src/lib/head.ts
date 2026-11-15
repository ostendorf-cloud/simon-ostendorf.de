import { watchEffect, type ComputedRef } from 'vue'
import type { Composer } from 'vue-i18n'

import type { Locale } from '../i18n'
import { canonical, fullTitle, hreflangAlternates, ogLocale, pageMeta, type PageRoute } from './seo'

/**
 * usePageHead keeps the document's title, description, canonical link and
 * hreflang alternates in step with the current route and locale, for a
 * client-side navigation between two already-hydrated pages.
 *
 * It is a no-op outside a browser: the prerender script builds the exact
 * same tags itself, from the same pageMeta, directly into the HTML
 * document it writes, since there is no DOM there to update.
 */
export function usePageHead(
  route: ComputedRef<PageRoute>,
  locale: ComputedRef<Locale>,
  i18n: Composer,
): void {
  watchEffect(() => {
    if (typeof document === 'undefined') {
      return
    }
    const meta = pageMeta(route.value, i18n.t)
    const siteName = i18n.t('common.siteName')
    document.title = fullTitle(route.value, meta, siteName)
    setMeta('description', meta.description)
    setLink('canonical', canonical(locale.value, route.value))
    setProperty('og:title', document.title)
    setProperty('og:description', meta.description)
    setProperty('og:url', canonical(locale.value, route.value))
    setProperty('og:locale', ogLocale(locale.value))
    replaceAlternates(hreflangAlternates(route.value))
  })
}

function setMeta(name: string, content: string): void {
  const el =
    document.querySelector<HTMLMetaElement>(`meta[name="${name}"]`) ?? createMeta('name', name)
  el.content = content
}

function setProperty(property: string, content: string): void {
  const el =
    document.querySelector<HTMLMetaElement>(`meta[property="${property}"]`) ??
    createMeta('property', property)
  el.content = content
}

function createMeta(attribute: 'name' | 'property', value: string): HTMLMetaElement {
  const el = document.createElement('meta')
  el.setAttribute(attribute, value)
  document.head.appendChild(el)
  return el
}

function setLink(rel: string, href: string): void {
  let el = document.querySelector<HTMLLinkElement>(`link[rel="${rel}"]`)
  if (!el) {
    el = document.createElement('link')
    el.rel = rel
    document.head.appendChild(el)
  }
  el.href = href
}

function replaceAlternates(alternates: Array<{ hreflang: string; href: string }>): void {
  document.querySelectorAll('link[rel="alternate"][hreflang]').forEach((el) => el.remove())
  for (const alternate of alternates) {
    const el = document.createElement('link')
    el.rel = 'alternate'
    el.hreflang = alternate.hreflang
    el.href = alternate.href
    document.head.appendChild(el)
  }
}
