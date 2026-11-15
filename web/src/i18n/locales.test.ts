import { describe, expect, it } from 'vitest'

import de from './locales/de.json'
import en from './locales/en.json'

/**
 * A missing translation key is caught here, at test time, rather than as a
 * runtime fallback to German or a raw key string shown to a visitor:
 * createAppI18n's fallbackLocale (see index.ts) hides a missing English
 * message well enough that nothing short of reading every page in both
 * locales would otherwise notice one slipped in.
 */
type Catalogue = { [key: string]: Catalogue | string }

/** flattenEntries pairs every leaf message key of catalogue with its own
 *  string value, dotted so "legal.imprint.title" compares directly against
 *  the other locale's own flattened list regardless of how deep either
 *  catalogue nests it. */
function flattenEntries(catalogue: Catalogue, prefix = ''): [string, string][] {
  const entries: [string, string][] = []
  for (const [key, value] of Object.entries(catalogue)) {
    const path = prefix === '' ? key : `${prefix}.${key}`
    if (typeof value === 'string') {
      entries.push([path, value])
    } else {
      entries.push(...flattenEntries(value, path))
    }
  }
  return entries
}

function flattenKeys(catalogue: Catalogue): string[] {
  return flattenEntries(catalogue)
    .map(([key]) => key)
    .sort()
}

describe('the German and English catalogues', () => {
  const deEntries = flattenEntries(de)
  const enEntries = flattenEntries(en)
  const deKeys = new Set(flattenKeys(de))
  const enKeys = new Set(flattenKeys(en))

  it('carry the same keys', () => {
    const onlyInGerman = [...deKeys].filter((key) => !enKeys.has(key))
    const onlyInEnglish = [...enKeys].filter((key) => !deKeys.has(key))

    expect(onlyInGerman, 'keys missing from en.json').toEqual([])
    expect(onlyInEnglish, 'keys missing from de.json').toEqual([])
  })

  it('both carry at least one key', () => {
    expect(deKeys.size).toBeGreaterThan(0)
    expect(enKeys.size).toBeGreaterThan(0)
  })

  it('carry no empty or whitespace only message', () => {
    const blankInGerman = deEntries.filter(([, value]) => value.trim() === '').map(([key]) => key)
    const blankInEnglish = enEntries.filter(([, value]) => value.trim() === '').map(([key]) => key)

    expect(blankInGerman, 'blank messages in de.json').toEqual([])
    expect(blankInEnglish, 'blank messages in en.json').toEqual([])
  })
})
