import { render } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import type { App, Plugin } from 'vue'

import AppFooter from '../components/AppFooter.vue'
import { createAppI18n, installI18n, type Locale } from '../i18n'
import { routeTable } from '../lib/routeTable'
import ImprintView from '../views/ImprintView.vue'
import PrivacyView from '../views/PrivacyView.vue'
import { OPERATOR } from './operator'

/** plugins collects what installI18n installs, so a component under test
 *  gets exactly the app's own i18n setup. */
function appPlugins(locale: Locale): Plugin[] {
  const plugins: Plugin[] = []
  const fake = {
    use(plugin: Plugin) {
      plugins.push(plugin)
      return fake
    },
  }
  installI18n(fake as unknown as App, createAppI18n(locale))
  return plugins
}

function renderPage(component: typeof ImprintView, locale: Locale): HTMLElement {
  return render(component, { global: { plugins: appPlugins(locale) } }).container as HTMLElement
}

const locales: Locale[] = ['de', 'en']

// The wording a dispute resolution section would use, in both languages. The
// owner decided no imprint carries one.
const disputeResolution = /streit|schlicht|dispute|arbitration|ec\.europa\.eu\/consumers\/odr/i

describe.each(locales)('the %s imprint', (locale) => {
  const page = () => renderPage(ImprintView, locale)

  it('has one h1 and states the operator', () => {
    const root = page()
    expect(root.querySelectorAll('h1')).toHaveLength(1)
    const text = root.textContent ?? ''
    for (const fact of [
      OPERATOR.name,
      OPERATOR.street,
      OPERATOR.postalCode,
      OPERATOR.city,
      OPERATOR.phone,
      OPERATOR.email,
      OPERATOR.vatId,
    ]) {
      expect(text).toContain(fact)
    }
    expect(root.querySelector(`a[href="tel:+4915730971784"]`)).not.toBeNull()
    expect(root.querySelector(`a[href="mailto:${OPERATOR.email}"]`)).not.toBeNull()
  })

  it('has no dispute resolution section', () => {
    expect(page().textContent ?? '').not.toMatch(disputeResolution)
  })
})

describe.each(locales)('the %s privacy page', (locale) => {
  const page = () => renderPage(PrivacyView, locale)

  it('has one h1 and names the controller and the counting', () => {
    const root = page()
    expect(root.querySelectorAll('h1')).toHaveLength(1)
    const text = root.textContent ?? ''
    expect(text).toContain(OPERATOR.name)
    expect(text).toContain(OPERATOR.email)
    expect(text).toContain('Umami')
    for (const field of ['/24', '/48', 'DNT: 1', 'Sec-GPC: 1', 'govue-theme']) {
      expect(text).toContain(field)
    }
  })

  it('has no dispute resolution section', () => {
    expect(page().textContent ?? '').not.toMatch(disputeResolution)
  })

  it('states own hardware, retention and the legal basis, and names no hosting provider', () => {
    const text = page().textContent ?? ''
    expect(text).toContain(locale === 'de' ? 'eigener Hardware' : 'own hardware')
    expect(text).toContain('30')
    expect(text).toContain('14')
    expect(text).toMatch(
      locale === 'de' ? /Art\. 6\s+Abs\. 1 lit\. f DSGVO/ : /Art\. 6\(1\)\(f\) GDPR/,
    )
    expect(text).not.toMatch(/hetzner|strato|ionos|ovh|aws|amazon/i)
  })

  it('no longer hedges the counting on a set up Umami instance', () => {
    const text = page().textContent ?? ''
    expect(text).not.toMatch(
      /nur eingeschaltet|eingerichtet|only on when|set up for this website|Without one/i,
    )
  })
})

describe('the German and English texts', () => {
  it('say the same: one heading per chapter in both languages', () => {
    const headings = (locale: Locale) =>
      renderPage(PrivacyView, locale).querySelectorAll('h2, h3').length
    expect(headings('en')).toBe(headings('de'))
  })

  it('mark the English privacy and imprint pages as a translation of the binding German text', () => {
    expect(renderPage(ImprintView, 'en').textContent).toContain('legally binding')
    expect(renderPage(PrivacyView, 'en').textContent).toContain('legally binding')
  })
})

describe.each(locales)('the %s footer', (locale) => {
  it('links the imprint and the privacy page of its own language at the exact path', () => {
    const { container } = render(AppFooter, { global: { plugins: appPlugins(locale) } })
    const hrefs = [...container.querySelectorAll('a')].map((a) => a.getAttribute('href'))
    expect(hrefs).toContain(`/${locale}/imprint`)
    expect(hrefs).toContain(`/${locale}/privacy`)
  })

  it('has no terms link, the site has no terms', () => {
    const { container } = render(AppFooter, { global: { plugins: appPlugins(locale) } })
    const hrefs = [...container.querySelectorAll('a')].map((a) => a.getAttribute('href'))
    expect(hrefs.some((href) => href?.endsWith('/terms'))).toBe(false)
  })
})

describe('the route table', () => {
  it('marks the imprint and the privacy page legal and has no terms route', () => {
    const legal = routeTable.filter((entry) => entry.legal).map((entry) => entry.slug)
    expect(legal.sort()).toEqual(['imprint', 'privacy'])
    expect(routeTable.some((entry) => entry.slug === 'terms')).toBe(false)
  })
})
