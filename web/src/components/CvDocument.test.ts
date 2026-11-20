import { render } from '@testing-library/vue'
import { describe, expect, it } from 'vitest'
import type { App, Plugin } from 'vue'

import { createAppI18n, installI18n, type Locale } from '../i18n'
import type { CvDocument as Doc } from '../lib/cv'
import CvDocument from './CvDocument.vue'

const full: Doc = {
  language: 'en',
  experience: [
    {
      id: 'exp-1',
      employer: 'Example Employer',
      title: 'Engineer',
      start: '2020-01',
      end: null,
      summary: 'Summary one.',
      location: null,
      employment_type: null,
    },
  ],
  education: [
    {
      id: 'edu-1',
      institution: 'Example University',
      degree: 'Degree',
      start: '2015-10',
      end: '2019-06',
      summary: null,
      grade: null,
    },
  ],
  certificates: [
    {
      id: 'c1',
      name: 'Old Certificate',
      issuer: 'Body',
      issued: '2020-01',
      valid_until: '2023-01',
    },
    {
      id: 'c2',
      name: 'Current Certificate',
      issuer: 'Body',
      issued: '2025-01',
      valid_until: '2027-01',
    },
    {
      id: 'c3',
      name: 'Lifetime Certificate',
      issuer: 'Body',
      issued: '2021-05',
      valid_until: null,
    },
  ],
  skills: [{ id: 's1', name: 'Kubernetes', level: 'expert' }],
  engagements: [
    {
      id: 'eng-1',
      experience: 'exp-1',
      title: 'Example Engagement',
      industry: 'Example Industry',
      start: '2020-03',
      end: '2021',
      summary: 'Engagement text.',
    },
  ],
  languages: [{ id: 'l1', name: 'English', level: 'proficient' }],
}

const empty: Doc = {
  language: 'en',
  experience: [],
  education: [],
  certificates: [],
  skills: [],
  engagements: [],
  languages: [],
}

const pdfs = { de: '/assets/cv.de-abc.pdf', en: '/assets/cv.en-def.pdf' }
const now = new Date('2026-11-20T12:00:00Z')

function renderCv(locale: Locale, document: Doc): HTMLElement {
  const plugins: Plugin[] = []
  const fake = {
    use(plugin: Plugin) {
      plugins.push(plugin)
      return fake
    },
  }
  installI18n(fake as unknown as App, createAppI18n(locale))
  return render(CvDocument, { props: { document, pdfs, now }, global: { plugins } })
    .container as HTMLElement
}

describe.each<Locale>(['de', 'en'])('the %s CV', (locale) => {
  it('renders every section with its own heading', () => {
    const root = renderCv(locale, full)
    for (const id of [
      'cv-pdf',
      'cv-experience',
      'cv-education',
      'cv-certificates',
      'cv-skills',
      'cv-engagements',
      'cv-languages',
    ]) {
      expect(root.querySelector(`[data-testid="${id}"] h2`), id).not.toBeNull()
    }
    expect(root.textContent).toContain('Example Engagement')
    expect(root.textContent).toContain('Example Employer')
  })

  it('omits an empty section instead of rendering an empty heading', () => {
    const root = renderCv(locale, empty)
    for (const id of [
      'cv-experience',
      'cv-education',
      'cv-certificates',
      'cv-skills',
      'cv-engagements',
      'cv-languages',
    ]) {
      expect(root.querySelector(`[data-testid="${id}"]`), id).toBeNull()
    }
    // Only the download card remains, with its heading.
    expect(root.querySelectorAll('h2')).toHaveLength(1)
  })

  it('marks exactly the certificate past its validity date as expired', () => {
    const root = renderCv(locale, full)
    const items = Array.from(root.querySelectorAll('[data-testid="cv-certificates"] li'))
    const marked = items.filter((li) => li.querySelector('[data-testid="expired"]'))
    expect(marked).toHaveLength(1)
    expect(marked[0]?.textContent).toContain('Old Certificate')
  })

  it('offers the PDF of both languages and says which language each is', () => {
    const root = renderCv(locale, full)
    const links = Array.from(root.querySelectorAll<HTMLAnchorElement>('[data-testid="cv-pdf"] a'))
    expect(links.map((a) => a.getAttribute('href'))).toEqual(
      locale === 'de' ? [pdfs.de, pdfs.en] : [pdfs.en, pdfs.de],
    )
    expect(links.map((a) => a.getAttribute('hreflang'))).toEqual(
      locale === 'de' ? ['de', 'en'] : ['en', 'de'],
    )
    const text = links.map((a) => a.textContent?.trim() ?? '')
    expect(text.some((s) => /Deutsch|German/.test(s))).toBe(true)
    expect(text.some((s) => /Englisch|English/.test(s))).toBe(true)
    for (const a of links) {
      expect(a.getAttribute('download')).toMatch(/^simon-ostendorf-cv\.(de|en)\.pdf$/)
    }
  })

  it('loads nothing from a third party', () => {
    const root = renderCv(locale, full)
    expect(root.querySelectorAll('img, iframe, script, link, source, video, audio')).toHaveLength(0)
  })
})
