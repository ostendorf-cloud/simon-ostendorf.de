/**
 * The CV page's data. `task content:build` (cmd/content) fetches the pinned cv
 * release, verifies it and writes web/src/generated/cv.json plus one PDF per
 * language. Those files are build inputs: nothing here fetches anything at
 * run time, and a build without them fails instead of rendering an empty CV.
 */
import type { Locale } from '../i18n'

export interface Experience {
  id: string
  employer: string
  title: string
  start: string
  end: string | null
  summary: string | null
  location: string | null
  employment_type: string | null
}

export interface Education {
  id: string
  institution: string
  degree: string
  start: string
  end: string | null
  summary: string | null
  grade: string | null
}

export interface Certificate {
  id: string
  name: string
  issuer: string
  issued: string
  valid_until: string | null
}

export interface Skill {
  id: string
  name: string
  level: string
}

export interface Engagement {
  id: string
  experience: string
  title: string
  industry: string | null
  start: string | null
  end: string | null
  summary: string
}

export interface Proficiency {
  id: string
  name: string
  level: string
}

export interface CvDocument {
  language: string
  experience: Experience[]
  education: Education[]
  certificates: Certificate[]
  skills: Skill[]
  engagements: Engagement[]
  languages: Proficiency[]
}

export interface CvDataset {
  release: string
  generated_at: string
  documents: Record<Locale, CvDocument>
}

export interface CvData {
  dataset: CvDataset
  /** The URL of each language's PDF, a content hashed file under /assets/. */
  pdfs: Record<Locale, string>
}

const datasets = import.meta.glob<CvDataset>('../generated/cv.json', {
  eager: true,
  import: 'default',
})
// ?no-inline keeps a small file a file: vite would otherwise inline it as a
// data URI, and a download link needs a real URL.
const pdfFiles = import.meta.glob<string>('../generated/cv.*.pdf', {
  eager: true,
  query: '?no-inline',
  import: 'default',
})

/** loadCv assembles the build inputs. It throws when one is missing, so the
 *  prerender, and with it the build, fails loudly rather than shipping a CV
 *  page without a CV. */
export function loadCv(
  sets: Record<string, CvDataset> = datasets,
  pdfs: Record<string, string> = pdfFiles,
): CvData {
  const dataset = sets['../generated/cv.json']
  if (!dataset) {
    throw new Error('cv: web/src/generated/cv.json is missing, run task content:build')
  }
  const out = {} as Record<Locale, string>
  for (const locale of ['de', 'en'] as const) {
    const url = pdfs[`../generated/cv.${locale}.pdf`]
    if (!url) {
      throw new Error(`cv: web/src/generated/cv.${locale}.pdf is missing, run task content:build`)
    }
    if (!dataset.documents?.[locale]) {
      throw new Error(`cv: the dataset has no ${locale} document, run task content:build`)
    }
    out[locale] = url
  }
  return { dataset, pdfs: out }
}

/** currentMonth is the YYYY-MM of date in UTC. */
function currentMonth(date: Date): string {
  return date.toISOString().slice(0, 7)
}

/** isExpired reports whether a certificate whose validity ends in validUntil
 *  (YYYY-MM, valid through the end of that month) is past it at now. A
 *  certificate without an end date never expires. */
export function isExpired(validUntil: string | null, now: Date): boolean {
  if (!validUntil) {
    return false
  }
  return currentMonth(now) > validUntil.slice(0, 7)
}

/** formatDate renders YYYY-MM or YYYY in locale. Anything else is shown as is. */
export function formatDate(value: string, locale: Locale): string {
  const month = /^(\d{4})-(\d{2})$/.exec(value)
  if (month) {
    const date = new Date(Date.UTC(Number(month[1]), Number(month[2]) - 1, 1))
    return new Intl.DateTimeFormat(locale, {
      month: 'long',
      year: 'numeric',
      timeZone: 'UTC',
    }).format(date)
  }
  return value
}
