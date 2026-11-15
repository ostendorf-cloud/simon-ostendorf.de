import { describe, expect, it } from 'vitest'

import { createAppI18n, detectPreferredLocale, isLocale } from './index'

describe('createAppI18n', () => {
  it('falls back to German when an English key is missing', () => {
    const i18n = createAppI18n('en')
    // A key this app's real messages do not carry in English yet, added
    // only to prove the fallback rather than to assert anything about the
    // real content files, which may always be fully translated.
    i18n.global.mergeLocaleMessage('de', { onlyInGerman: { greeting: 'Hallo' } })

    expect(i18n.global.t('onlyInGerman.greeting')).toBe('Hallo')
  })

  it('uses the English message directly when one exists', () => {
    const i18n = createAppI18n('en')
    i18n.global.mergeLocaleMessage('de', { both: { greeting: 'Hallo' } })
    i18n.global.mergeLocaleMessage('en', { both: { greeting: 'Hello' } })

    expect(i18n.global.t('both.greeting')).toBe('Hello')
  })

  it('never shows the raw key for a value German also carries', () => {
    const i18n = createAppI18n('en')
    i18n.global.mergeLocaleMessage('de', { onlyInGerman: { greeting: 'Hallo' } })

    expect(i18n.global.t('onlyInGerman.greeting')).not.toBe('onlyInGerman.greeting')
  })

  it('defaults to German', () => {
    const i18n = createAppI18n()
    expect(i18n.global.locale.value).toBe('de')
  })
})

describe('isLocale', () => {
  it('accepts de and en', () => {
    expect(isLocale('de')).toBe(true)
    expect(isLocale('en')).toBe(true)
  })

  it('rejects anything else', () => {
    expect(isLocale('fr')).toBe(false)
    expect(isLocale('')).toBe(false)
  })
})

describe('detectPreferredLocale', () => {
  it('picks English when it is the highest ranked language', () => {
    expect(detectPreferredLocale(['en-US', 'de'])).toBe('en')
  })

  it('falls back to German for an unsupported language', () => {
    expect(detectPreferredLocale(['fr-FR'])).toBe('de')
  })

  it('falls back to German when there is no language at all', () => {
    expect(detectPreferredLocale([])).toBe('de')
  })
})
