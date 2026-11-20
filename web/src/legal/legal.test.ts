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
    expect(text).toContain(locale === 'de' ? 'nach 30 Tagen' : 'after 30 days')
    expect(text).toContain(locale === 'de' ? 'nach 14 Monaten' : 'after 14 months')
    expect(text).toMatch(
      locale === 'de' ? /Art\. 6\s+Abs\. 1 lit\. f DSGVO/ : /Art\. 6\(1\)\(f\) GDPR/,
    )
    expect(text).not.toMatch(/zu hause|at home|at my home/i)
    expect(text).not.toMatch(/hetzner|strato|ionos|ovh|aws|amazon/i)
  })

  it('states the access log fields and the separate ingress log', () => {
    const section = [...page().querySelectorAll('h3')]
      .find((h) => (h.textContent ?? '').trim().startsWith('2.1'))
      ?.closest('div')
    expect(section).toBeTruthy()
    const norm = (s: string | null) => (s ?? '').replace(/\s+/g, ' ').trim()
    const items = [...(section?.querySelectorAll('ul li') ?? [])].map((li) => norm(li.textContent))
    const paragraphs = [...(section?.querySelectorAll('p') ?? [])].map((p) => norm(p.textContent))
    const ingress = paragraphs.find((p) =>
      p.includes(locale === 'de' ? 'Ingress des Clusters' : "cluster's ingress"),
    )
    expect(ingress).toBeTruthy()

    if (locale === 'de') {
      expect(items).toEqual([
        'den Zeitpunkt der Anfrage',
        'die Anfragemethode',
        'die aufgerufene Seite als Routenmuster, nie den Pfad oder die Parameter',
        'den HTTP-Statuscode und die Menge der übertragenen Daten',
        'die Dauer der Bearbeitung',
        'Ihre IP-Adresse, die ungekürzt protokolliert wird; hinter dem vorgeschalteten Ingress ist es die Adresse, die dieser als Ihre meldet',
        'eine zufällige Anfragekennung',
      ])
      expect(paragraphs[0]).toContain('Zugriffsprotokoll')
      for (const field of [
        'Ihre IP-Adresse',
        'Datum und Uhrzeit',
        'die Anfragezeile mit Pfad und Parametern',
        'den HTTP-Statuscode, die übertragene Datenmenge',
        'den Referrer',
        'den User-Agent',
      ]) {
        expect(ingress).toContain(field)
      }
      expect(items.join(' ')).not.toMatch(/User-Agent|Referrer/)
      expect(paragraphs[1]).toContain(
        'Nicht protokolliert werden der User-Agent, der Referrer und der aufgerufene Pfad',
      )
    } else {
      expect(items).toEqual([
        'the time of the request',
        'the request method',
        'the page as a route pattern, never the path or the parameters',
        'the HTTP status code and the amount of data transferred',
        'the duration of the request',
        'your IP address, logged in full; behind the ingress in front of the application it is the address the ingress reports as yours',
        'a random request identifier',
      ])
      expect(paragraphs[0]).toContain('access log')
      for (const field of [
        'your IP address',
        'the date and time',
        'the request line with path and parameters',
        'the HTTP status code, the amount of data transferred',
        'the referrer',
        'the user agent',
      ]) {
        expect(ingress).toContain(field)
      }
      expect(items.join(' ')).not.toMatch(/user agent|referrer/i)
      expect(paragraphs[1]).toContain(
        'The user agent, the referrer and the requested path with its parameters are not logged',
      )
    }
  })

  it('no longer claims the IP address is neither stored nor logged', () => {
    const text = page().textContent ?? ''
    expect(text).not.toMatch(
      /IP-Adresse nicht|not store or log your IP|kein Zugriffsprotokoll|no access log/i,
    )
  })

  it('states one legitimate interest sentence per purpose and the weighing', () => {
    const text = page().textContent ?? ''
    const sentences =
      locale === 'de'
        ? [
            'Für beide Protokolle ist die Rechtsgrundlage Art. 6 Abs. 1 lit. f DSGVO. Mein berechtigtes Interesse ist der sichere Betrieb der Website, die Analyse von Fehlern und die Abwehr von Missbrauch. Ich habe dieses Interesse gegen Ihre Rechte abgewogen und halte die Daten gering.',
            'Für die Zählung der Seitenaufrufe ist die Rechtsgrundlage Art. 6 Abs. 1 lit. f DSGVO. Mein berechtigtes Interesse ist, zu verstehen, wie die Website genutzt wird. Ich habe dieses Interesse gegen Ihre Rechte abgewogen und halte die Daten gering.',
          ]
        : [
            'For both logs the legal basis is Art. 6(1)(f) GDPR. My legitimate interest is the secure operation of the website, the analysis of errors and the defence against abuse. I weighed this interest against your rights and keep the data minimal.',
            'For counting page views the legal basis is Art. 6(1)(f) GDPR. My legitimate interest is to understand how the website is used. I weighed this interest against your rights and keep the data minimal.',
          ]
    for (const sentence of sentences) {
      expect(text.replace(/\s+/g, ' ')).toContain(sentence)
    }
    expect(text).toContain(
      locale === 'de' ? 'Die Kennung wechselt jeden Monat.' : 'The identifier changes every month.',
    )
    expect(text).not.toMatch(/deshalb jeden Monat|therefore changes/)
  })

  it('uses the pseudonymous wording and has a right to object paragraph', () => {
    const text = page().textContent ?? ''
    expect(text).not.toMatch(/anonym/i)
    expect(text).toContain(locale === 'de' ? 'pseudonyme Kennung' : 'pseudonymous identifier')
    expect(text).not.toMatch(
      /Datenschutzbeauftragte|data protection officer|Empfänger|recipient|Cloudflare|Let's Encrypt|Drittland|third countr/i,
    )
    expect(text).toContain('Art. 21')
    expect(text).toContain(
      locale === 'de'
        ? 'Global Privacy Control sendet; diese Signale beachtet die Website'
        : 'Global Privacy Control; the website honours these signals',
    )
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
