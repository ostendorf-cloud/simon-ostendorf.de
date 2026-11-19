import { render } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import type { App, Plugin } from 'vue'

import { createAppI18n, installI18n, type Locale } from '../i18n'
import HomeView from './HomeView.vue'

function renderHome(locale: Locale): HTMLElement {
  const plugins: Plugin[] = []
  const fake = {
    use(plugin: Plugin) {
      plugins.push(plugin)
      return fake
    },
  }
  installI18n(fake as unknown as App, createAppI18n(locale))
  return render(HomeView, { global: { plugins } }).container as HTMLElement
}

describe.each<Locale>(['de', 'en'])('the %s home page', (locale) => {
  it('has one h1', () => {
    expect(renderHome(locale).querySelectorAll('h1')).toHaveLength(1)
  })

  it('offers exactly the two ways on as plain external links', () => {
    const links = Array.from(renderHome(locale).querySelectorAll('a'))
    expect(links.map((a) => a.getAttribute('href'))).toEqual([
      'https://ostendorf.dev',
      'https://ostendorf.tech',
    ])
    for (const a of links) {
      expect(a.getAttribute('rel')).toContain('noopener')
    }
  })

  it('loads nothing from a third party', () => {
    const root = renderHome(locale)
    expect(root.querySelectorAll('img, iframe, script, link, source, video, audio')).toHaveLength(0)
  })
})
