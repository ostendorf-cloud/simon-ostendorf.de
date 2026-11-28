import { render } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import { createMemoryHistory } from 'vue-router'

import { createAppI18n } from '../i18n'
import { createAppRouter } from '../router'
import LocaleSwitch from './LocaleSwitch.vue'

async function switchHref(name: string, from: 'de' | 'en'): Promise<string | null> {
  const router = createAppRouter(createMemoryHistory())
  await router.push({ name, params: { locale: from } })
  await router.isReady()
  const { container } = render(LocaleSwitch, {
    global: { plugins: [router, createAppI18n(from)] },
  })
  return container.querySelector('a')?.getAttribute('href') ?? null
}

describe('LocaleSwitch', () => {
  // A page's canonical path has no trailing slash below the language root.
  // A link with one would be a 301 on every language switch.
  it('links the other language of a page without a trailing slash', async () => {
    expect(await switchHref('privacy', 'de')).toBe('/en/privacy')
    expect(await switchHref('imprint', 'en')).toBe('/de/imprint')
    expect(await switchHref('cv', 'de')).toBe('/en/cv')
  })

  it('links the other language of the home page with the language root slash', async () => {
    expect(await switchHref('home', 'en')).toBe('/de/')
    expect(await switchHref('not-found', 'en')).toBe('/de/')
  })
})
