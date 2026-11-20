import { describe, expect, it } from 'vitest'

import { formatDate, isExpired, loadCv, type CvDataset, type CvDocument } from './cv'

const emptyDoc = (language: string): CvDocument => ({
  language,
  experience: [],
  education: [],
  certificates: [],
  skills: [],
  engagements: [],
  languages: [],
})

const dataset: CvDataset = {
  release: 'v0.1.0',
  generated_at: '2026-11-20T12:00:00Z',
  documents: { de: emptyDoc('de'), en: emptyDoc('en') },
}

describe('isExpired', () => {
  const now = new Date('2026-11-20T12:00:00Z')

  it('keeps a certificate valid through the end of its last month', () => {
    expect(isExpired('2026-11', now)).toBe(false)
    expect(isExpired('2026-12', now)).toBe(false)
  })

  it('marks a certificate past its last month as expired', () => {
    expect(isExpired('2026-10', now)).toBe(true)
    expect(isExpired('2023-01', now)).toBe(true)
  })

  it('never expires a certificate without an end date', () => {
    expect(isExpired(null, now)).toBe(false)
  })
})

describe('formatDate', () => {
  it('renders a month in the locale', () => {
    expect(formatDate('2024-05', 'de')).toBe('Mai 2024')
    expect(formatDate('2024-05', 'en')).toBe('May 2024')
  })

  it('leaves a bare year alone', () => {
    expect(formatDate('2021', 'en')).toBe('2021')
  })
})

describe('loadCv', () => {
  const pdfs = {
    '../generated/cv.de.pdf': '/assets/de.pdf',
    '../generated/cv.en.pdf': '/assets/en.pdf',
  }

  it('fails when the dataset is missing, so a build without the CV cannot succeed', () => {
    expect(() => loadCv({}, pdfs)).toThrow(/task content:build/)
  })

  it('fails when a PDF is missing', () => {
    expect(() =>
      loadCv({ '../generated/cv.json': dataset }, { '../generated/cv.de.pdf': '/a.pdf' }),
    ).toThrow(/cv\.en\.pdf is missing/)
  })

  it('fails when a language has no document', () => {
    const partial = { ...dataset, documents: { de: emptyDoc('de') } } as unknown as CvDataset
    expect(() => loadCv({ '../generated/cv.json': partial }, pdfs)).toThrow(/no en document/)
  })

  it('returns the dataset and one PDF URL per locale', () => {
    const cv = loadCv({ '../generated/cv.json': dataset }, pdfs)
    expect(cv.pdfs).toEqual({ de: '/assets/de.pdf', en: '/assets/en.pdf' })
    expect(cv.dataset.release).toBe('v0.1.0')
  })
})
